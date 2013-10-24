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
	"errors"
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
	Q            *point
	IsCompressed bool
}

///////////////////////////////////////////////////////////////////////
/*
 * Get byte representation of public key.
 * @return []byte - byte array representing a public key
 */
func (k *PublicKey) Bytes() []byte {
	return pointAsBytes(k.Q, k.IsCompressed)
}

///////////////////////////////////////////////////////////////////////
/*
 * Get public key from byte representation.
 */
func PublicKeyFromBytes(b []byte) (*PublicKey, error) {
	pnt, compr, err := pointFromBytes(b)
	if err != nil {
		return nil, err
	}
	key := &PublicKey{
		Q:            pnt,
		IsCompressed: compr,
	}
	return key, nil
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
	b := coordAsBytes(k.D)
	if k.IsCompressed {
		b = append(b, 1)
	}
	return b
}

///////////////////////////////////////////////////////////////////////
/*
 * Get private key from byte representation.
 */
func PrivateKeyFromBytes(b []byte) (*PrivateKey, error) {
	// check compressed/uncompressed
	var (
		kd    []byte = b
		compr bool   = false
	)
	if len(b) == 33 {
		kd = b[:32]
		if b[32] == 1 {
			compr = true
		} else {
			return nil, errors.New("Invalid private key format (compression flag)")
		}
	} else if len(b) != 32 {
		return nil, errors.New("Invalid private key format (length)")
	}
	// set private factor.
	key := &PrivateKey{}
	key.D = new(big.Int).SetBytes(kd)
	// compute public key
	g := GetBasePoint()
	key.Q = scalarMult(g, key.D)
	key.IsCompressed = compr
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
		prv.Q = ScalarMultBase(prv.D)
		prv.IsCompressed = true

		// check for valid key
		if !isInf(prv.PublicKey.Q) {
			break
		}
	}
	return prv
}
