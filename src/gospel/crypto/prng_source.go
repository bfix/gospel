/*
 * Cryptographically strong source of randomness.  
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
	"big"
	"crypto/rand"
)

///////////////////////////////////////////////////////////////////////
// Cryptographically strong source of random bits 

/*
 * Source of randomness: 
 */
type prng struct {
}

//=====================================================================
/*
 * Get next (unsigned) 64-bit integer value.
 * @return int64 - random integer
 */
func (p *prng) Int63() int64 {

	val,err := rand.Int (rand.Reader, new(big.Int).Lsh (big.NewInt(1), 63))
	if err != nil {
		panic ("PRNG failure: " + err.String())
	}
	return val.Int64()
}

//---------------------------------------------------------------------
/*
 * Seeding a source: not necessary, because random bits are generated
 * on a system level by either a hardware RNG or a cryptographically
 * secure PRNG algorithm.
 * @param seed int64 - seeding value
 */
func (p *prng) Seed (seed int64) {
	// intentionally not implemented
}

//=====================================================================
/*
 * Instantiate a new source for random bits.
 * @return *prng - reference to rand.Source instance
 */
func NewPrngSource() *prng {
	return &prng{}
}
