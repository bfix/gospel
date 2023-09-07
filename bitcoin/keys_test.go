package bitcoin

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
)

func TestKeys(t *testing.T) {
	var prv *PrivateKey
	for i := 0; i < 32; i++ {
		prv = GenerateKeys(i&1 == 1)
		b := prv.Bytes()
		if _, err := PrivateKeyFromBytes(b); err != nil {
			t.Fatal("PrivateKeyFromBytes failed")
		}
		b = prv.PublicKey.Bytes()
		if _, err := PublicKeyFromBytes(b); err != nil {
			t.Fatal("PublicKeyFromBytes failed")
		}
		pnt := prv.Q
		tst := MultBase(prv.D)
		if !(pnt.IsOnCurve() && pnt.Equals(tst)) {
			t.Fatal("public point failed")
		}
	}
}
