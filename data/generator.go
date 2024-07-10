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

package data

import "sync/atomic"

// GeneratorChannel is used as a link between the generator boilerplate
// and a generator function.
type GeneratorChannel[T any] chan T

// Yield a value (computed in a generator function) for a consumer.
// If the return value is false, the consumer indicates no more values
// are expected - so the generator function can terminate immediately.
func (g GeneratorChannel[T]) Yield(val T) (ok bool) {
	g <- val
	_, ok = <-g
	return
}

// Done with this generator - no more values can be generated.
func (g GeneratorChannel[T]) Done() {
	close(g)
}

// GeneratorFunction generates objects of given type and "yields"
// them to a channel. If the generator reaches end-of-output, it
// closes the channel. If yield returns false, the consumer has
// closed the channel to indicate it doesn't want any new objects.
//
// Example:
//
//		   func gen(out GeneratorChannel[string]) {
//			   for i := 0; i < 100; i++ {
//	               val := fmt.Sprintf("%d", i+1)
//			       if !out.Yield(val) {
//			           return
//		           }
//		       }
//		       out.Done()
//		   }
type GeneratorFunction[T any] func(GeneratorChannel[T])

// Generator for elements of given type.
// A generator object only handles the boilerplate on behalf of a
// generator function.
//
// Example:
//
//		   g := NewGenerator(gen)
//		   for s := range g.Run() {
//	             :
//			     if ??? {
//			         g.Stop()
//			         break
//			     }
//		   }
type Generator[T any] struct {
	out    chan T
	ctrl   chan int
	active atomic.Bool
}

// NewGenerator runs the generator function in a go-routine.
// Calling the Run() method on an instance returns a channel of given
// type where values can be retrieved. To terminate a running generator,
// call the Stop() method. After a generator has finished (either by
// itself or because it was stopped), it cannot be restarted.
func NewGenerator[T any](gen GeneratorFunction[T]) *Generator[T] {
	g := new(Generator[T])
	g.out = make(chan T)
	g.ctrl = make(chan int)
	g.active.Store(true)
	var null T
	go func() {
		ch := make(GeneratorChannel[T])
		go gen(ch)
	loop:
		for {
			select {
			case x, ok := <-ch:
				if !ok {
					break loop
				}
				if !g.active.Load() {
					ch.Done()
					break loop
				}
				g.out <- x
				ch <- null
			case <-g.ctrl:
				g.active.Store(false)
			}
		}
		close(g.out)
		g.active.Store(false)
	}()
	return g
}

// Run returns a (read) channel from the generator.
func (g *Generator[T]) Run() <-chan T {
	if !g.active.Load() {
		panic("inactive generator can't be run")
	}
	return g.out
}

// Stop the generator (no more values expected).
func (g *Generator[T]) Stop() {
	if !g.active.Load() {
		return
	}

	var done atomic.Bool
	done.Store(false)
	go func() {
		g.ctrl <- 0
		done.Store(true)
	}()
	for !done.Load() {
		<-g.out
	}
}
