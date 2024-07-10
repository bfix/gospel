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
	"errors"

	gerr "github.com/bfix/gospel/errors"
	"github.com/bfix/gospel/math"
)

var (
	alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	b58      = math.NewInt(58)
)

// Error codes
var (
	ErrBtcBase58Decoding = errors.New("base58 decoding error -- unknown character")
)

// Base58Encode converts byte array to base58 string representation
func Base58Encode(in []byte) string {

	// convert byte array to integer
	val := math.NewIntFromBytes(in)

	// convert integer to base58 representation
	b := []byte{}
	var m *math.Int
	for val.Cmp(math.ZERO) > 0 {
		val, m = val.DivMod(b58)
		b = append(b, alphabet[int(m.Int64())])
	}
	// handle leading zero bytes in input
	for _, x := range in {
		if x == 0 {
			b = append(b, alphabet[0])
		} else {
			break
		}
	}
	// return base58 representation
	return string(reverse(b))
}

// Base58Decode converts a base58 representation into byte array
func Base58Decode(s string) (data []byte, err error) {

	// convert string to byte array
	in := []byte(s)

	// convert base58 to integer (ignores leading zeros)
	val := math.ZERO
	for _, b := range in {
		pos := bytes.IndexByte(alphabet, b)
		if pos == -1 {
			err = gerr.New(ErrBtcBase58Decoding, "char '%c'", b)
			return
		}
		val = val.Mul(b58).Add(math.NewInt(int64(pos)))
	}
	// prefix byte array with leading zeros
	pf := []byte{}
	for _, x := range s {
		if byte(x) == alphabet[0] {
			pf = append(pf, 0)
		} else {
			break
		}
	}
	// return final byte array
	data = append(pf, val.Bytes()...)
	return
}

// reverse byte array
func reverse(in []byte) []byte {
	n := len(in)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[n-i-1] = in[i]
	}
	return out
}
