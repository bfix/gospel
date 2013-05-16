/*
 * Elliptic curve 'Secp256k1' engine for ECDSA.
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
	"math/big"
)

///////////////////////////////////////////////////////////////////////
/*
 * Sign hash value with private key.
 * [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 13f]
 * @param key *PrivateKey - key used to sign hash value
 * @param hash []byte - hash value to be signed
 * @return r,s *big.Int - signature values
 */
func Sign(key *PrivateKey, hash []byte) (r, s *big.Int) {

	var k, kInv *big.Int
	for {
		// compute value of 'r' as x-coordinate of k*G with random k
		for {
			// get random value
			k = n_rnd(three)
			// get its modular inverse
			kInv = n_inv(k)

			// compute k*G
			pnt := scalarMultBase(k)
			r = n_mod(pnt.x)
			if r.Sign() != 0 {
				break
			}
		}
		// compute value of 's := (rd + e)/k'
		e := convertHash(hash)
		s = n_mul(_add(n_mul(key.d, r), e), kInv)
		if s.Sign() != 0 {
			break
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////
/*
 * Verify hash value with public key.
 * [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 15f]
 * @param key *PublicKey - key used to verify signature
 * @param hash []byte - hash value of signed content
 * @param r,s *big.Int - signature values
 * @return bool - correct signature?
 */
func Verify(key *PublicKey, hash []byte, r, s *big.Int) bool {

	// sanity checks for arguments
	if r.Sign() == 0 || s.Sign() == 0 {
		return false
	}
	if r.Cmp(curve_n) >= 0 || s.Cmp(curve_n) >= 0 {
		return false
	}
	// check signature
	e := convertHash(hash)
	w := n_inv(s)

	u1 := e.Mul(e, w)
	u2 := w.Mul(r, w)

	p1 := scalarMultBase(u1)
	p2 := scalarMult(key.q, u2)
	if p1.x.Cmp(p2.x) == 0 {
		return false
	}
	p3 := add(p1, p2)
	rr := n_mod(p3.x)
	return rr.Cmp(r) == 0
}

///////////////////////////////////////////////////////////////////////
// convert hash value to integer
// [http://www.secg.org/download/aid-780/sec1-v2.pdf]

func convertHash(hash []byte) *big.Int {

	// trim hash value (if required)
	maxSize := (curve_n.BitLen() + 7) / 8
	if len(hash) > maxSize {
		hash = hash[:maxSize]
	}

	// convert to integer
	val := new(big.Int).SetBytes(hash)
	val.Rsh(val, uint(maxSize*8-curve_n.BitLen()))
	return val
}
