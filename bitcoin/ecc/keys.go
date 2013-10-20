/*
 * Elliptic curve cryptography key handling.
 *
 * (c) 2011-2012 Bernd Fix   >Y<
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

package ecc

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"github.com/bfix/gospel/math"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
/*
 * PublicKey is a point on the elliptic curve: (x,y) = d*G, where
 * 'G' is the base point of the curve and 'd' is the secret private
 * factor (private key)
 */
type PublicKey struct {
	Q *point
}

///////////////////////////////////////////////////////////////////////
/*
 * Get byte representation of public key.
 * @param compressed bool - compressed representation?
 * @return []byte - byte array representing a public key
 */
func (k *PublicKey) Bytes(compressed bool) []byte {
	return pointAsBytes(k.Q, compressed)
}

///////////////////////////////////////////////////////////////////////
/*
 * Get public key from byte representation.
 */
func PublicKeyFromBytes(b []byte) (*PublicKey, error) {
	pnt, err := pointFromBytes(b)
	if err != nil {
		return nil, err
	}
	return &PublicKey{pnt}, nil
}

///////////////////////////////////////////////////////////////////////
/*
 * PrivateKey is a random factor 'd' for the base point that yields
 * the associated PublicKey (point on the curve (x,y) = d*G)
 */
type PrivateKey struct {
	PublicKey
	D *big.Int
}

///////////////////////////////////////////////////////////////////////
/*
 * Get byte representation of private key.
 * @return []byte - byte array representing a private key
 */
func (k *PrivateKey) Bytes() []byte {
	return coordAsBytes(k.D)
}

///////////////////////////////////////////////////////////////////////
/*
 * Get private key from byte representation.
 */
func PrivateKeyFromBytes(b []byte) (*PrivateKey, error) {
	key := &PrivateKey{}
	key.D = new(big.Int).SetBytes(b)
	g := GetBasePoint()
	key.Q = scalarMult(g, key.D)
	return key, nil
}

///////////////////////////////////////////////////////////////////////
/*
 * Generate a new set of keys.
 * [http://www.nsa.gov/ia/_files/ecdsa.pdf] page 19f but with a
 * different range (value 1 and 2 for exponent are not allowed)
 * @return *PrivateKey - generated key pair
 */
func GenerateKeys() *PrivateKey {

	prv := new(PrivateKey)
	for {
		// generate factor in range [3..n-1]
		prv.D = n_rnd(math.THREE)
		// generate point p = d*G
		prv.PublicKey.Q = ScalarMultBase(prv.D)

		// check for valid key
		if !isInf(prv.PublicKey.Q) {
			break
		}
	}
	return prv
}
