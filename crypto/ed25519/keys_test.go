package ed25519

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
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bfix/gospel/math"
)

var (
	seed = []byte{
		0x20, 0x3f, 0x2f, 0x8c, 0x54, 0xf4, 0x1a, 0xd3,
		0x01, 0x9a, 0x56, 0x92, 0x19, 0xda, 0xee, 0x4f,
		0xd2, 0x53, 0x55, 0xa6, 0x3c, 0xfc, 0x57, 0x40,
		0x8a, 0xb0, 0x86, 0x88, 0xf6, 0x86, 0xf4, 0x9c,
	}

	d = []byte{
		0x7F, 0xDE, 0x7A, 0xAA, 0xEA, 0x0D, 0xA1, 0x7A,
		0x7B, 0xCB, 0x4F, 0x57, 0x49, 0xCC, 0xA9, 0xBE,
		0xA7, 0xFB, 0x2B, 0x85, 0x77, 0xAD, 0xC9, 0x55,
		0xDA, 0xB2, 0x68, 0xB2, 0xB4, 0xCC, 0x24, 0x78,
	}

	q = []byte{
		0x92, 0xDC, 0xBF, 0x39, 0x40, 0x2D, 0xC6, 0x3C,
		0x97, 0xA6, 0x81, 0xE0, 0xFC, 0xD8, 0x7C, 0x74,
		0x17, 0xD3, 0xA3, 0x8C, 0x52, 0xFD, 0xE0, 0x49,
		0xBC, 0xD0, 0x1C, 0x0A, 0x0B, 0x8C, 0x02, 0x51,
	}

	prv = NewPrivateKeyFromSeed(seed)
	pub = prv.Public()
)

func TestPrvKey(t *testing.T) {
	//	if testing.Verbose() {
	//		fmt.Printf("PRIVATE (seed=%s, d=%s)\n",
	//			hex.EncodeToString(seed), hex.EncodeToString(prv.D.Bytes()))
	//	}

	pubB := pub.Bytes()
	//	if testing.Verbose() {
	//		fmt.Printf("PUBLIC  = %s\n", hex.EncodeToString(pubB))
	//	}

	if !bytes.Equal(pubB, q) {
		fmt.Printf("PUBLIC(computed) = %s\n", hex.EncodeToString(pubB))
		fmt.Printf("PUBLIC(expected) = %s\n", hex.EncodeToString(q))
		t.Fatal("Public key mismatch")
	}

	dVal := math.NewIntFromBytes(d)
	if !dVal.Equals(prv.D) {
		t.Fatal("Private exponent mismatch")
	}

	pub2 := NewPrivateKeyFromD(dVal).Public().Bytes()
	if !bytes.Equal(pubB, pub2) {
		fmt.Printf("PUBLIC2(computed) = %s\n", hex.EncodeToString(pub2))
		fmt.Printf("PUBLIC2(expected) = %s\n", hex.EncodeToString(pubB))
		t.Fatal("Public key mismatch")
	}
}
