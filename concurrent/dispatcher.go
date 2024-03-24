//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2024 Bernd Fix  >Y<
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

package concurrent

import (
	"context"
	"sync"
	"sync/atomic"
)

// Dispatchable interface
type Dispatchable[T, R any] interface {

	// Worker using channels to read task and write results.
	Worker(ctx context.Context, n int, taskCh chan T, resCh chan R)

	// Eval receives results from workers
	Eval(result R) bool
}

// Dispatcher managing worker go-routines
type Dispatcher[T, R any] struct {
	taskCh  chan T
	resCh   chan R
	ctrl    chan int
	running atomic.Bool
}

// NewDispatcher runs a new dispatcher with given number of workers and
// a Dispatchanle implementation.
func NewDispatcher[T, R any](ctx context.Context, numWorker int, disp Dispatchable[T, R]) *Dispatcher[T, R] {
	d := new(Dispatcher[T, R])
	d.taskCh = make(chan T)
	d.resCh = make(chan R)
	d.ctrl = make(chan int)

	// start worker go-routines
	wg := new(sync.WaitGroup)
	for n := 0; n < numWorker; n++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			disp.Worker(ctx, num, d.taskCh, d.resCh)
		}(n)
	}

	// run dispatcher loop
	d.running.Store(true)
	go func() {
		// clean-up on exit
		defer func() {
			d.running.Store(false)
			wg.Wait()
			close(d.taskCh)
			close(d.resCh)
		}()

		ctxD, cancel := context.WithCancel(ctx)
		for {
			select {
			// handle termination
			case <-ctxD.Done():
				cancel()
				return
			case <-d.ctrl:
				cancel()
				return

			// handle result
			case x := <-d.resCh:
				if disp.Eval(x) {
					cancel()
					return
				}
			}
		}
	}()
	return d
}

// Process a task. Returns false if the dispatcher is closed.
func (d *Dispatcher[T, R]) Process(task T) bool {
	if !d.running.Load() {
		return false
	}
	d.taskCh <- task
	return true
}

// Quit dispatcher run
func (d *Dispatcher[T, R]) Quit() {
	d.ctrl <- 0
}
