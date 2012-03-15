/*
 * Pseudo Random Number Generator based on cryptographically strong
 * source of randomness.
 *
 * (c) 2012 Bernd Fix   >Y<
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

package crypto

///////////////////////////////////////////////////////////////////////
// Import external declarations.

import (
	"math/big"
	"math/rand"
)

///////////////////////////////////////////////////////////////////////
// Random number generator instance

var rnd = rand.New(NewPrngSource())

///////////////////////////////////////////////////////////////////////
// Public Methods to generate randomized objects

/*
 * Return a random integer value with given range.
 * @param lower int - lower bound (inclusive)
 * @param upper int - upper bound (exclusive)
 * @return int - random number (in given range)
 */
func RandInt(lower, upper int) int {
	return lower + (rnd.Int() % (upper - lower + 1))
}

//=====================================================================
/*
 * Generate a byte array of given size with random content.
 * @param n int - size of resulting byte array
 * @return []byte - byte array with random content
 */
func RandBytes(n int) []byte {
	data := make([]byte, n)
	for n := 0; n < n; n++ {
		data[n] = byte(rnd.Int() & 0xFF)
	}
	return data
}

//=====================================================================
/*
 * Return a random big integer value with given range.
 * @param lower *big.Int - lower bound (inclusive)
 * @param upper *big.Int - upper bound (exclusive)
 * @return *big.Int - random number (in given range)
 */
func RandBigInt(lower, upper *big.Int) *big.Int {
	span := new(big.Int).Sub(upper, lower)
	span = new(big.Int).Add(span, big.NewInt(1))
	ofs := new(big.Int).Rand(rnd, span)
	return new(big.Int).Add(lower, ofs)
}
