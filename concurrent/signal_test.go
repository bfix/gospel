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
	"fmt"
	"sync"
	"testing"
	"time"
)

// Test #1: Create signaller without listeners; send signal and retire the
// signaller; a subsequent send should fail with an error.
func TestSginallerEmpty(t *testing.T) {
	s := NewSignaller()
	if err := s.Send(true); err != nil {
		t.Fail()
	}
	s.Retire()
	if s.Send(false) != ErrSignallerRetired {
		t.Fail()
	}
}

// Test #2: create a signaller to serve multiple listeners
func TestSginallerGroup(t *testing.T) {
	s := NewSignaller()
	// create listener routines
	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if testing.Verbose() {
				fmt.Printf("Listener #%d started...\n", id)
			}
			listener := s.Listen()
			if listener == nil {
				t.Fail()
			}
		loop:
			for {
				select {
				case sig := <-listener:
					if testing.Verbose() {
						fmt.Printf("Listener #%d received signal '%v'\n", id, sig)
					}
					break loop
				default:
				}
			}
			s.Drop(listener)
		}(i + 1)
	}
	time.Sleep(2 * time.Second)
	s.Send(true)
	wg.Wait()
}
