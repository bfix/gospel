package bitcoin

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
	"testing"

	"github.com/bfix/gospel/math"
)

func TestEngine(t *testing.T) {
	for i := 0; i < 32; i++ {
		prv := GenerateKeys(i&1 == 1)
		hash := nRnd(math.ONE).Bytes()
		sig := Sign(prv, hash)
		if !Verify(&prv.PublicKey, hash, sig) {
			t.Fatal("sign/verify failed")
		}
	}
}

func TestHash(t *testing.T) {
	i := nRnd(math.ONE)
	h := i.Bytes()
	j := convertHash(h)
	if i.Cmp(j) != 0 {
		t.Fatal("convertHash failed")
	}
}
