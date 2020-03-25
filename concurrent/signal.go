package concurrent

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"fmt"
)

// Error codes
var (
	ErrSignallerRetired = fmt.Errorf("Signaller retiered")
	ErrUnknownListener  = fmt.Errorf("Unknown signal listener")
)

// Signal can be any object (intrinsic or custom); it is the responsibility of
// the senders and receivers of signals to handle them accordingly.
type Signal interface{}

// ListenerOp codes
var (
	LISTENER_ADD  = 0
	LISTENER_DROP = 1
)

// ListenerOp represents an operation on the listener list:
type ListenerOp struct {
	ch chan Signal // listener channel
	op int         // 0=add, 1=delete
}

// Signaller manages signals send by senders for multiple listeners.
type Signaller struct {
	inCh   chan Signal          // channel for incoming signals
	outChs map[chan Signal]bool // channels for out-going signals

	cmdCh  chan *ListenerOp // internal channel to synchronize maintenance
	resCh  chan interface{} // channel for command results
	active bool             // is the signaller dispatching signals?
}

// NewSignaller instantiates a new signal manager:
func NewSignaller() *Signaller {
	// create a new instance and initialize it.
	s := &Signaller{
		inCh:   make(chan Signal),
		outChs: make(map[chan Signal]bool),
		cmdCh:  make(chan *ListenerOp),
		resCh:  make(chan interface{}),
		active: true,
	}
	// run the dispatch loop as long as the signaller is active.
	go func() {
		for s.active {
			select {
			case cmd := <-s.cmdCh:
				// handle listener list operation
				switch cmd.op {
				case LISTENER_ADD:
					// create a new listener channel
					out := make(chan Signal)
					s.outChs[out] = true
					s.resCh <- out
				case LISTENER_DROP:
					delete(s.outChs, cmd.ch)
					s.resCh <- true
				}
			case x := <-s.inCh:
				// dispatch received signals
				for out, active := range s.outChs {
					if active {
						out <- x
					}
				}
			default:
			}
		}
	}()
	return s
}

// Retire a signaller: This will terminate the dispatch loop for signals; no
// further send or listen operations are supported. A retired signaller cannot
// be re-activated.
func (s *Signaller) Retire() {
	s.active = false
}

// Send a signal to be dispatched to all listeners.
func (s *Signaller) Send(sig Signal) error {
	// check for active signaller
	if !s.active {
		return ErrSignallerRetired
	}
	s.inCh <- sig
	return nil
}

// Listen returns a channel to listen on
func (s *Signaller) Listen() chan Signal {
	// check for active signaller
	if !s.active {
		return nil
	}
	s.cmdCh <- &ListenerOp{op: LISTENER_ADD}
	return (<-s.resCh).(chan Signal)
}

// DropListener removes a listener from the list.
func (s *Signaller) Drop(out chan Signal) error {
	if _, ok := s.outChs[out]; !ok {
		return ErrUnknownListener
	}
	s.cmdCh <- &ListenerOp{
		ch: out,
		op: LISTENER_DROP,
	}
	<-s.resCh
	return nil
}
