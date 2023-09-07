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

package wallet

import (
	"bytes"
	"math/big"
)

//----------------------------------------------------------------------
// Helper functions for Bech32
//----------------------------------------------------------------------

// Bech32Bit5 splits a byte array into 5-bit chunks
func Bech32Bit5(data []byte) []byte {
	size := len(data) * 8
	v := new(big.Int).SetBytes(data)
	pad := size % 5
	if pad != 0 {
		v = new(big.Int).Lsh(v, uint(5-pad))
	}
	num := (size + 4) / 5
	res := make([]byte, num)
	for i := num - 1; i >= 0; i-- {
		res[i] = byte(v.Int64() & 31)
		v = new(big.Int).Rsh(v, 5)
	}
	return res
}

func Bech32CRC(hrp string, data []byte) (crc []byte) {
	buf := new(bytes.Buffer)
	buf.Write(bech32ExpandHRP(hrp))
	buf.Write(data)
	buf.Write([]byte{0, 0, 0, 0, 0, 0})
	pm := bech32Polymod(buf.Bytes()) ^ 1
	crc = make([]byte, 6)
	for i := range crc {
		crc[i] = byte((pm >> (5 * (5 - i))) & 31)
	}
	return
}

func bech32Polymod(data []byte) (chk uint32) {
	gen := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk = 1
	for _, v := range data {
		b := (chk >> 25)
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i, g := range gen {
			if (b>>i)&1 == 1 {
				chk ^= g
			}
		}
	}
	return chk
}

func bech32ExpandHRP(hrp string) (buf []byte) {
	n := len(hrp)
	buf = make([]byte, 2*n+1)
	buf[n] = 0
	for i, c := range hrp {
		b := byte(c)
		buf[i] = b >> 5
		buf[i+n+1] = b & 31
	}
	return
}
