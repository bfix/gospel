//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

package network

import (
	"sync"
	"time"

	"github.com/bfix/gospel/logger"
)

//----------------------------------------------------------------------
// Rate limiter
//----------------------------------------------------------------------

// RateLimiter computes rate limit-compliant delays for requests
type RateLimiter struct {
	rates       []int      // rates [sec, min, hr, day, week]
	lock        sync.Mutex // one request at a time
	last, first *entry     // reference to last and first entry
	intern      bool       // nested calls don't block
}

// NewRateLimiter creates a newly initialitzed rate limiter.
func NewRateLimiter(rate ...int) *RateLimiter {
	lim := new(RateLimiter)
	lim.rates = make([]int, 5)
	copy(lim.rates, rate)
	lim.last = newEntry()
	lim.first = lim.last
	lim.intern = false
	return lim
}

// Stats returns current statistics for the rate limiter
func (lim *RateLimiter) Stats() (stats *RateStats) {
	// only one request at a time
	if !lim.intern {
		lim.lock.Lock()
		defer lim.lock.Unlock()
	}

	// assemble statistics
	stats = &RateStats{
		ts:    time.Now().Unix(),
		rSec:  lim.rates[0],
		rMin:  lim.rates[1],
		rHr:   lim.rates[2],
		rDay:  lim.rates[3],
		rWeek: lim.rates[4],
		xLast: lim.last,
	}
	var e, next *entry
loop:
	for e, next = lim.last, nil; e != nil; next, e = e, e.prev {
		tDiff := stats.ts - e.ts
		switch {
		case tDiff > 3600*24*7:
			// cut off tail
			if next != nil {
				next.prev = nil
			}
			lim.first = next
			// drop out-of-range entries (one week)
			for e != nil {
				e = e.drop()
			}
			// we are done
			break loop
		// Count events in time-slot
		case tDiff > 3600*24:
			stats.xOldest = e
			stats.pWeek++
		case tDiff > 3600:
			stats.xWeek = e
			stats.pDay++
		case tDiff > 60:
			stats.xDay = e
			stats.pHr++
		case tDiff > 0:
			stats.xHr = e
			stats.pMin++
		case tDiff == 0:
			stats.pSec++
		}
	}
	// correct accumulation
	stats.pMin += stats.pSec
	stats.pHr += stats.pMin
	stats.pDay += stats.pHr
	stats.pWeek += stats.pDay

	// adjust boundary links
	if stats.xHr == nil {
		stats.xHr = lim.first
	}
	if stats.xDay == nil {
		stats.xDay = stats.xHr
	}
	if stats.xWeek == nil {
		stats.xWeek = stats.xDay
	}
	if stats.xOldest == nil {
		stats.xOldest = stats.xWeek
	}
	return
}

// Pass waits for a rate limit-compliant delay before passing a new request
func (lim *RateLimiter) Pass() {
	// only one request at a time
	lim.lock.Lock()
	lim.intern = true
	defer func() {
		lim.intern = false
		lim.lock.Unlock()
	}()

	// get current rate statistics
	stats := lim.Stats()
	delay := stats.Wait()
	// delay for given time
	if delay > 0 {
		logger.Printf(logger.DBG, "RateLimit: Delaying for %d seconds", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}
	// prepend new request at beginning of list
	e := newEntry()
	e.prev = lim.last
	lim.last = e
}

//----------------------------------------------------------------------
// Rate statistics
//----------------------------------------------------------------------

// RateStats contains rate statistics
type RateStats struct {
	ts                               int64  // timestamp
	rSec, rMin, rHr, rDay, rWeek     int    // rate limits
	pSec, pMin, pHr, pDay, pWeek     int    // actual rates
	xLast, xHr, xDay, xWeek, xOldest *entry // pointer to border elements
}

// Wait returns the delay (wait time) to be rate-limit compliant
func (rs *RateStats) Wait() int {
	delay := 0
	eval := func(r, p, d int) {
		if r > 0 && p+1 > r {
			if d < 1 {
				d = 1
			}
			if d > delay {
				delay = d
			}
		}
	}
	eval(rs.rSec, rs.pSec, 1)
	eval(rs.rMin, rs.pMin, 61-int(rs.ts-rs.xHr.ts))
	eval(rs.rHr, rs.pHr, 3601-int(rs.ts-rs.xDay.ts))
	eval(rs.rDay, rs.pDay, 86401-int(rs.ts-rs.xWeek.ts))
	eval(rs.rWeek, rs.pWeek, 604801-int(rs.ts-rs.xOldest.ts))
	return delay
}

//----------------------------------------------------------------------
// Helper types
//----------------------------------------------------------------------

// Entry in a single-linked list
type entry struct {
	ts   int64
	prev *entry
}

// Drop the entry (cut link)
// Returns the linked entry.
func (e *entry) drop() *entry {
	p := e.prev
	e.prev = nil
	return p
}

// Return a new request entry
func newEntry() *entry {
	return &entry{
		ts:   time.Now().Unix(),
		prev: nil,
	}
}
