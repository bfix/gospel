package p2p

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
	"crypto/rand"
	"testing"
)

func TestAddressString(t *testing.T) {
	a := make([]byte, 32)
	for i := 0; i < 23; i++ {
		n, err := rand.Read(a)
		if n != 32 {
			t.Fatal("rand.Read() failed short")
		}
		if err != nil {
			t.Fatal(err)
		}
		addr := NewAddress(a)
		s := addr.String()
		t.Log(s)
		addr2, err := NewAddressFromString(s)
		if err != nil {
			t.Fatal(err)
		}
		if !addr.Equals(addr2) {
			t.Fatal("address mismatch")
		}
	}
}
