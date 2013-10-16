/*
 * Base58 en- and decoding test functions.
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
	"fmt"
	"github.com/bfix/gospel/crypto"
	"math/big"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestBase58(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("bitcoin/util/base58 Test")
	fmt.Println("********************************************************")

	fmt.Println("Checking Base58 conversion functions:")

	one := big.NewInt(1)

	if !test1(big.NewInt(57)) {
		t.Fail()
	}
	if !test1(big.NewInt(58)) {
		t.Fail()
	}
	if !test1(big.NewInt(255)) {
		t.Fail()
	}
	if !test2([]byte{0, 255}) {
		t.Fail()
	}
	if !test2([]byte{0, 0, 255}) {
		t.Fail()
	}
	bound := big.NewInt(256)
	for n := 0; n < 128; n++ {
		if !test1(crypto.RandBigInt(one, bound)) {
			t.Fail()
		}
		bound = new(big.Int).Lsh(bound, 1)
	}
}

///////////////////////////////////////////////////////////////////////
// test helper function

func test1(x *big.Int) bool {
	s := Base58Encode(x.Bytes())
	b, err := Base58Decode(s)
	if err != nil {
		return false
	}
	y := new(big.Int).SetBytes(b)
	res := x.Cmp(y) == 0
	return res
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
