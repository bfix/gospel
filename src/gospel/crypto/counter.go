/*
 * Cryptographic counter implementation based on the Paillier
 * crypto scheme.
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
// import external declarations

import (
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// type definitions:

/*
 * Cryptographic counter
 */
type Counter struct {
	pubkey *PaillierPublicKey // reference to public Paillier key
	data   *big.Int           // encrypted counter value
}

///////////////////////////////////////////////////////////////////////
// Public counter methods:

/*
 * Create a new Counter instance for given public key.
 * @param k *PaillierPublicKey - associated public Paillier key
 * @return c *Counter - reference to Counter instance
 * @return err error - error object (or nil if successful)
 */
func NewCounter(k *PaillierPublicKey) (c *Counter, err error) {

	// create a new counter with value "0"
	d, err := k.Encrypt(big.NewInt(0))
	if err != nil {
		return nil, err
	}
	c = &Counter{
		pubkey: k,
		data:   d,
	}
	return c, nil
}

//---------------------------------------------------------------------
/*
 * Get the encrypted counter value.
 * @param self *Counter - this instance
 * @return *big.Int - encrypted counter value
 */
func (self *Counter) Get() *big.Int {
	return self.data
}

//---------------------------------------------------------------------
/*
 * Increment counter: usually called with step values of "0" (don't
 * change counter, but change representation) and "1" (increment by
 * one step).
 * @param self *Counter - this instance
 * @param step *big.Int - increment
 * @return error - error object (or nil if successful)
 */
func (self *Counter) Increment(step *big.Int) error {

	d, err := self.pubkey.Encrypt(step)
	if err != nil {
		return err
	}
	self.data.Mul(self.data, d)
	return nil
}
