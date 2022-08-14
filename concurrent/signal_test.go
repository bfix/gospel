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
	"sync"
	"testing"
	"time"
)

// Test #1: Create signaller without listeners; send signal and retire the
// signaller; a subsequent send should fail with an error.
func TestSginallerEmpty(t *testing.T) {
	s := NewSignaller()
	if err := s.Send(true); err != nil {
		t.Fatal(err)
	}
	s.Retire()
	if s.Send(false) != ErrSigInactive {
		t.Fail()
	}
}

// Test #2: create a signaller to serve multiple listeners
func TestSignallerGroup(t *testing.T) {
	s := NewSignaller()
	// create listener routines
	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			t.Logf("Listener #%d started...\n", id)
			listener, err := s.Listener()
			if err != nil {
				t.Error(err)
				return
			}
			sig := <-listener.Signal()
			t.Logf("Listener #%d received signal '%v'\n", id, sig)
			listener.Close()
		}(i + 1)
	}
	time.Sleep(time.Second)
	if err := s.Send(true); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestSignallerHanging(t *testing.T) {
	s := NewSignaller()
	// create listener routine
	var (
		listener *Listener
		err      error
	)
	ready := make(chan bool)
	quit := false
	go func() {
		t.Log("Listener started...")
		if listener, err = s.Listener(); err != nil {
			return
		}
		ready <- true
	loop:
		for sig := range listener.Signal() {
			switch x := sig.(type) {
			case nil:
				t.Log("Listener closed")
				break loop
			case bool:
				t.Log("Listener received 'quit' signal")
				quit = true
				break loop
			case int:
				t.Logf("Listener received number '%d'\n", x)
				// now waste time
				time.Sleep(2 * time.Second)
			default:
				t.Logf("Listener received signal '%v'\n", x)
			}
		}
		listener.Close()
	}()

	// wait for listener ready
	<-ready
	if err != nil {
		t.Fatal(err)
	}

	// number signal
	if err = s.Send(23); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	if err = s.Send(42); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	// 'quit' signal
	// (not be seen as listener should have been dropped earlier)
	if err = s.Send(true); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	if quit {
		t.Fatal("'quit' signal received")
	}
}
