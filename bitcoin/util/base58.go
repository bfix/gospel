/*
 * Base58 en- and decoding.
 *
 * (c) 2011-2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package util

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"bytes"
	"errors"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// global variables

var (
	alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	zero     = big.NewInt(0)
	b58      = big.NewInt(58)
)

///////////////////////////////////////////////////////////////////////
// convert byte array to base58 string representation

func Base58Encode(in []byte) string {

	// convert byte array to integer
	val := new(big.Int).SetBytes(in)

	// convert integer to base58 representation
	b := []byte{}
	m := big.NewInt(0)
	for val.Cmp(zero) > 0 {
		val, m = new(big.Int).DivMod(val, b58, m)
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

///////////////////////////////////////////////////////////////////////
// convert base58 representation into byte array

func Base58Decode(s string) ([]byte, error) {

	// convert string to byte array
	in := []byte(s)

	// convert base58 to integer (ignores leading zeros)
	val := big.NewInt(0)
	for _, b := range in {
		pos := bytes.IndexByte(alphabet, b)
		if pos == -1 {
			return nil, errors.New("Base58 decoding error -- unknown character")
		}
		val = new(big.Int).Mul(val, b58)
		val = new(big.Int).Add(val, big.NewInt(int64(pos)))
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
	return append(pf, val.Bytes()...), nil
}

///////////////////////////////////////////////////////////////////////
// reverse byte array

func reverse(in []byte) []byte {
	n := len(in)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[n-i-1] = in[i]
	}
	return out
}
