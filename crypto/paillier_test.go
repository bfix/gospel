package crypto

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
	"github.com/bfix/gospel/math"
	"testing"
)

func TestPaillier(t *testing.T) {
	for size := 128; size <= 2048; size *= 2 {
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			t.Fatal("newpaillierprivatekey failed")
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 5; i++ {
			m := math.NewIntRnd(pub.N)
			c, err := pub.Encrypt(m)
			if err != nil {
				t.Fatal("encrypt failed")
			}
			d, err := priv.Decrypt(c)
			if err != nil {
				t.Fatal("decrypt failed")
			}
			if !d.Equals(m) {
				t.Fatal("message mismatch")
			}
		}
	}
}
