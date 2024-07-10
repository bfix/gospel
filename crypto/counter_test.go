package crypto

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

import (
	"testing"

	"github.com/bfix/gospel/math"
)

func TestCounter(t *testing.T) {
	for size := 128; size <= 2048; size *= 2 {
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			t.Fatal("newpaillierprivatekey failed")
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 3; i++ {
			cnt, err := NewCounter(pub)
			if err != nil {
				t.Fatal("newcounter failed")
			}
			var inc int64
			for i := 0; i < 5; i++ {
				v := math.NewIntRnd(math.TWO)
				if err = cnt.Increment(v); err != nil {
					t.Fatal(err)
				}
				if v.Bit(0) == 1 {
					inc++
				}
			}
			tt := cnt.Get()
			tt, err = priv.Decrypt(tt)
			if err != nil {
				t.Fatal("decrypt failed")
			}
			v := tt.Int64()
			if v != inc {
				t.Fatal("counter mismatch")
			}
		}
	}
}
