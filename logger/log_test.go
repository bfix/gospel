package logger

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
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

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

const (
	NUM_TASKS = 20
)

var (
	wg sync.WaitGroup
)

func task(id int, delay int, ch chan bool) {
	defer wg.Done()
	for {
		select {
		case <-ch:
			Printf(INFO, "[%d] Task started (delayed %d ms)\n", id, delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
			Printf(INFO, "[%d] Task ended\n", id)
			return
		default:
		}
	}
}

func newTask(id int, delay int) chan bool {
	ch := make(chan bool, 0)
	wg.Add(1)
	go task(id, delay, ch)
	return ch
}

func TestLogger(t *testing.T) {
	list := make([]chan bool, NUM_TASKS)
	Println(INFO, "Test run started...")
	for i := 0; i < NUM_TASKS; i++ {
		list[i] = newTask(i, 500+int(rand.Int31n(2500)))
	}
	for _, ch := range list {
		ch <- true
	}
	wg.Wait()
	Println(INFO, "Test run Finished...")
}
