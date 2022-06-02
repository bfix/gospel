package concurrent

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
// the senders and receivers of signals to handle them accordingly.
type Signal interface{}

//----------------------------------------------------------------------

// Listener for signals managed by Signaller
type Listener struct {
	ch   chan Signal // channel to receive on
	refs int         // number of pending dispatches
}

// Signal returns the channel from which to read the signal.
func (l *Listener) Signal() <-chan Signal {
	return l.ch
}

//----------------------------------------------------------------------
// Signaller (signal dispatcher to listeners)
//----------------------------------------------------------------------

// Signaller dispatches signals to multiple concurrent listeners. The sequence
// in which listeners are served is stochastic.
//
// In highly concurrent environments with a lot of messages the sequence of
// signals seen by a listener can vary. This is due to the fact that a signal
// gets dispatched in a Go routine, so the next signal can be dispatched
// before a listener got the first one if the second Go routine handles the
// listener earlier. It is therefore mandatory that received signals from
// a listener get handled in a Go routine as well to keep latency low. If
// a listener violates that promise, it got removed from the list.
type Signaller struct {
	inCh   chan Signal        // channel for incoming signals
	outChs map[*Listener]bool // channels for out-going signals

	cmdCh      chan *listenerOp // internal channel to synchronize maintenance
	resCh      chan interface{} // channel for command results
	active     bool             // is the signaller dispatching signals?
	maxLatency time.Duration    // max time for listener to respond
}

// NewSignaller instantiates a new signal manager:
func NewSignaller() *Signaller {
	// create a new instance and initialize it.
	s := &Signaller{
		inCh:       make(chan Signal),
		outChs:     make(map[*Listener]bool),
		cmdCh:      make(chan *listenerOp),
		resCh:      make(chan interface{}),
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
					listener := &Listener{
						ch:   make(chan Signal),
						refs: 0,
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
				active := make([]*Listener, 0)
				for lst := range s.outChs {
					active = append(active, lst)
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
							s.Drop(listener)

						// message sent
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
func (s *Signaller) SetLatency(d time.Duration) {
	s.maxLatency = d
}

// Retire a signaller: This will terminate the dispatch loop for signals; no
// further send or listen operations are supported. A retired signaller cannot
// be re-activated.
func (s *Signaller) Retire() {
	s.active = false
}

//----------------------------------------------------------------------

// Send a signal to be dispatched to all listeners.
func (s *Signaller) Send(sig Signal) error {
	// check for active signaller
	if !s.active {
		return ErrSigInactive
	}
	s.inCh <- sig
	return nil
}

//----------------------------------------------------------------------

// Listener returns a new channel to listen on each time it is called.
// Function interested in listening should get the channel, start the
// for/select loop and drop the channel if the loop terminates.
// Requesting an listener and than not reading from it will block all
// other listeners of the signaller.
func (s *Signaller) Listener() (*Listener, error) {
	// check for active signaller
	if !s.active {
		return nil, ErrSigInactive
	}
	// trigger add operation.
	s.cmdCh <- &listenerOp{op: sigListenerAdd}
	return (<-s.resCh).(*Listener), nil
}

// Drop removes a listener from the list. Failing to drop or close a
// listener will result in hanging go routines.
func (s *Signaller) Drop(listener *Listener) error {
	// check for active signaller
	if !s.active {
		return ErrSigInactive
	}
	// trigger delete operation
	s.cmdCh <- &listenerOp{
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
	sigListenerRef
	sigListenerUnref
)

// listenerOp represents an operation on the listener list:
type listenerOp struct {
	op  int       // sigListener????
	lst *Listener // listener reference
}
