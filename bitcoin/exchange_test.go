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

func TestExchange(t *testing.T) {
	var err error

	if _, err = ImportPrivateKey("invalid", true); err == nil {
		t.Fatal("importprivatekey failed")
	}

	for n := 0; n < 100; n++ {
		testnet := (n&1 == 1)
		compr := (n%3 == 0)
		key := GenerateKeys(compr)
		s := ExportPrivateKey(key, testnet)
		x, _ := Base58Decode(s)
		if len(x) != 37 && len(x) != 38 {
			t.Fatal("invalid key size")
		}

		kk, err := ImportPrivateKey(s, testnet)
		if err != nil || kk.D.Cmp(key.D) != 0 {
			t.Fatal("key mismatch")
		}

		tt := make([]byte, len(x))
		copy(tt, x)
		tt[0] = 0
		ss := Base58Encode(tt)
		if _, err = ImportPrivateKey(ss, testnet); err == nil {
			t.Fatal("version check failed")
		}

		copy(tt, x)
		tt[len(tt)-1] ^= 255
		ss = Base58Encode(tt)
		if _, err = ImportPrivateKey(ss, testnet); err == nil {
			t.Fatal("hash test failed")
		}

		if compr {
			tt = x
			tt[33] = 0
			ss = Base58Encode(tt)
			if _, err = ImportPrivateKey(ss, testnet); err == nil {
				t.Fatal("compression check failed")
			}
		}

		copy(tt, x)
		if len(tt) == 37 {
			tt = tt[:36]
		} else {
			tt = append(tt, 0)
		}
		ss = Base58Encode(tt)
		if _, err = ImportPrivateKey(ss, testnet); err == nil {
			t.Fatal("key size check failed")
		}
	}
}
