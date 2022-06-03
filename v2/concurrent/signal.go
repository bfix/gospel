package concurrent

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2022 Bernd Fix
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
	"errors"
	"time"
)

// Error codes
var (
	ErrSigInactive   = errors.New("signaller inactive")
	ErrSigNoListener = errors.New("no signal listener")
)

//----------------------------------------------------------------------

// Signal can be any object (intrinsic or custom); it is the responsibility of
// the sender and the receiver of signals to handle them accordingly.
//
// A signal in this context is a stand-alone unit of information. Its "meaning"
// does neither depend on other signals nor on their sequence. Its only purpose
// is to communicate state changes to listeners instead of sharing the state
// globally via memory reference ("Do not communicate by sharing memory; instead,
// share memory by communicating." -- https://go.dev/doc/effective_go.html)
//
type Signal[S any] struct {
	value S
}

// NewSignal instantiates a signal,
func NewSignal[S any](value S) Signal[S] {
	return Signal[S]{
		value: value,
	}
}

// Value returns the value associated with the signal.
func (s *Signal[S]) Value() S {
	return s.value
}

//----------------------------------------------------------------------

// Listener for signals managed by Signaller
type Listener[S any] struct {
	ch     chan Signal[S] // channel to receive on
	refs   int            // number of pending dispatches
	mngr   *Signaller[S]  // back-ref to signaller instance
	active bool           // is listener active?
}

// Signal returns the channel from which to read the signal. If the
// returned signal is nil, no further signals will be received on this
// listener and the select-loop MUST terminate.
func (l *Listener[S]) Signal() <-chan Signal[S] {
	return l.ch
}

// Close listener: This more an announcement than an operation as the
// channel is not closed immediately. It is possible for the listener
// to receive some more signals before it actually closes.
func (l *Listener[S]) Close() {
	l.mngr.drop(l)
}

//----------------------------------------------------------------------
// Signaller (signal dispatcher to listeners)
//----------------------------------------------------------------------

// Signaller dispatches signals to multiple concurrent listeners. The sequence
// in which listeners are served is stochastic.
//
// In highly concurrent environments with a lot of signals the sequence of
// signals seen by a listener can vary. This is due to the fact that a signal
// gets dispatched in a Go routine, so the next signal can be dispatched
// before a listener got the first one if the second Go routine handles the
// listener earlier. It is therefore mandatory that received signals from
// a listener get handled in a Go routine as well to keep latency low. If
// a listener violates that promise, it got removed from the list.
//
type Signaller[S any] struct {
	inCh   chan Signal[S]        // channel for incoming signals
	outChs map[*Listener[S]]bool // channels for out-going signals

	cmdCh      chan *listenerOp[S] // internal channel to synchronize maintenance
	resCh      chan any            // channel for command results
	active     bool                // is the signaller dispatching signals?
	maxLatency time.Duration       // max time for listener to respond
}

// NewSignaller instantiates a new signal manager:
func NewSignaller[S any]() *Signaller[S] {
	// create a new instance and initialize it.
	s := &Signaller[S]{
		inCh:       make(chan Signal[S]),
		outChs:     make(map[*Listener[S]]bool),
		cmdCh:      make(chan *listenerOp[S]),
		resCh:      make(chan any),
		active:     true,
		maxLatency: time.Second,
	}
	// run the dispatch loop as long as the signaller is active.
	go func() {
		for s.active {
			select {
			// handle listener list operation
			case cmd := <-s.cmdCh:
				switch cmd.op {
				// create a new listener channel
				case sigListenerAdd:
					listener := &Listener[S]{
						ch:     make(chan Signal[S]),
						refs:   0,
						mngr:   s,
						active: true,
					}
					s.outChs[listener] = true
					s.resCh <- listener

				// remove listener from list
				case sigListenerDrop:
					var err error
					if _, ok := s.outChs[cmd.lst]; !ok {
						err = ErrSigNoListener
					} else {
						// remove from list
						delete(s.outChs, cmd.lst)
						// close unreferenced channels
						if cmd.lst.refs == 0 {
							close(cmd.lst.ch)
						}
					}
					s.resCh <- err
				}

			// dispatch received signals
			case sig := <-s.inCh:
				// create a list of currently active listeners
				// so we can serve them in a Go routine.
				active := make([]*Listener[S], 0)
				for lst := range s.outChs {
					if lst.active {
						active = append(active, lst)
					}
					// increment pending count on listener
					lst.refs++
				}
				go func() {
					for _, listener := range active {
						done := make(chan struct{})
						go func() {
							defer func() {
								// decrease pending count on listener
								listener.refs--
							}()
							listener.ch <- sig
							close(done)
						}()
						select {
						case <-time.After(s.maxLatency):
							// listener not responding: drop it
							s.drop(listener)

						// signal sent
						case <-done:
						}
					}
				}()
			}
		}
	}()
	return s
}

// SetLatency sets the max latency for listener. A listener is removed from
// the list if it violates this policy.
func (s *Signaller[S]) SetLatency(d time.Duration) {
	s.maxLatency = d
}

// Retire a signaller: This will terminate the dispatch loop for signals; no
// further send operations are supported. A retired signaller cannot be
// re-activated. Running dispatches will not be interrupted.
func (s *Signaller[S]) Retire() {
	s.active = false
}

//----------------------------------------------------------------------

// Send a signal to be dispatched to all listeners.
func (s *Signaller[S]) Send(val S) error {
	// check for active signaller
	if !s.active {
		return ErrSigInactive
	}
	// send to dispatcher
	s.inCh <- NewSignal(val)
	return nil
}

//----------------------------------------------------------------------

// Listener returns a new channel to listen on each time it is called.
// Function interested in listening should get the channel, start the
// for/select loop and drop the channel if the loop terminates.
// Requesting an listener and than not reading from it will block all
// other listeners of the signaller.
func (s *Signaller[S]) Listener() (*Listener[S], error) {
	// check for active signaller
	if !s.active {
		return nil, ErrSigInactive
	}
	// trigger add operation.
	s.cmdCh <- &listenerOp[S]{op: sigListenerAdd}
	return (<-s.resCh).(*Listener[S]), nil
}

// drop removes a listener from the list. Failing to drop or close a
// listener will result in hanging go routines.
func (s *Signaller[S]) drop(listener *Listener[S]) error {
	// check for active signaller
	if !s.active {
		return ErrSigInactive
	}
	// tag the listener unavailable immediately
	listener.active = false

	// trigger delete operation
	s.cmdCh <- &listenerOp[S]{
		lst: listener,
		op:  sigListenerDrop,
	}
	// handle error return for command.
	var err error
	res := <-s.resCh
	if res != nil {
		err = res.(error)
	}
	return err
}

//----------------------------------------------------------------------

// ListenerOp codes
const (
	sigListenerAdd = iota
	sigListenerDrop
)

// listenerOp represents an operation on the listener list:
type listenerOp[S any] struct {
	op  int          // sigListener????
	lst *Listener[S] // listener reference
}
