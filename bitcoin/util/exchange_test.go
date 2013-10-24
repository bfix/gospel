/*
 * Exchange-related test functions.
 *
 * (c) 2013 Bernd Fix   >Y<
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
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestExchange(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("bitcoin/util/exchange Test")
	fmt.Println("********************************************************")

	for n := 0; n < 100; n++ {
		testnet := (n & 1 == 1)
		key := ecc.GenerateKeys()

		s := ExportPrivateKey(key, testnet)

		kk, err := ImportPrivateKey(s, testnet)
		if err != nil {
			fmt.Println("ImportPrivateKey() failed: " + err.Error())
			t.Fail()
			return
		}

		if kk.D.Cmp(key.D) != 0 {
			fmt.Println("key mismatch")
			t.Fail()
			return
		}
	}
}
