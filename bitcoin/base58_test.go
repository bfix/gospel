package bitcoin

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
	"bytes"
	"testing"

	"github.com/bfix/gospel/math"
)

func TestBase58(t *testing.T) {
	if !test1(math.NewInt(57)) {
		t.Fatal("base58 failure")
	}
	if !test1(math.NewInt(58)) {
		t.Fatal("base58 failure")
	}
	if !test1(math.NewInt(255)) {
		t.Fatal("base58 failure")
	}
	if !test2([]byte{0, 255}) {
		t.Fatal("base58 failure")
	}
	if !test2([]byte{0, 0, 255}) {
		t.Fatal("base58 failure")
	}
	bound := math.NewInt(256)
	for n := 0; n < 128; n++ {
		if !test1(math.NewIntRndRange(math.ONE, bound)) {
			t.Fatal("base58 failure")
		}
		bound = bound.Lsh(1)
	}

	if _, err := Base58Decode("invalid"); err == nil {
		t.Fatal("base58 failure")
	}
}

func test1(x *math.Int) bool {
	s := Base58Encode(x.Bytes())
	b, err := Base58Decode(s)
	if err != nil {
		return false
	}
	y := math.NewIntFromBytes(b)
	return x.Equals(y)
}

func test2(x []byte) bool {
	s := Base58Encode(x)
	y, err := Base58Decode(s)
	if err != nil {
		return false
	}
	res := bytes.Equal(x, y)
	return res
}
