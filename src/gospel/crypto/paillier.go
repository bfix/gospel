/*
 * Paillier crypto scheme implementation.
 *
 * (c) 2012 Bernd Fix   >Y<
 *
 *---------------------------------------------------------------------  
 * #### Key generation
 *
 * The key used in the Paillier crypto system consists of four integer
 * values. The public key has two parameters; the private key has three
 * parameters (one parameter is shared between the keys). As in RSA it
 * starts with two random primes 'p' and 'q'; the public key parameter
 * are computed as:
 *
 *   n := p * q
 *   g := random number from interval [0,n^2[
 *
 * The private key parameters are computed as:
 *
 *   n := p * q
 *   l := lcm (p-1,q-1)
 *   u := (((g^l mod n^2)-1)/n) ^-1 mod n
 *
 * N.B. The division by n is integer based and rounds toward zero!
 *
 * #### Encryption
 *
 * The encryption function in the Paillier crypto scheme is:
 *
 *   c = E(m) = (g^m * r^n) mod n^2
 *
 * where 'r' is a random number from the interval [0,n[. This encryption
 * allows different encryption results for the same message, based on
 * the actual value of 'r'.
 *
 * #### Decryption
 *
 * The decryption function in the Paillier crypto scheme is:
 *
 *   m = D(c) = ((((c^l mod n^2)-1)/n) * u) mod n 
 *
 * N.B. As in the key generation process the division by n is integer
 *      based and rounds toward zero!
 *---------------------------------------------------------------------  
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
	"crypto/rand"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// Paillier key pair

/*
 * Paillier public key
 */
type PaillierPublicKey struct {
	N, G *big.Int
}

/*
 * Paillier private key
 */
type PaillierPrivateKey struct {
	*PaillierPublicKey
	L, U *big.Int
	P, Q *big.Int
}

///////////////////////////////////////////////////////////////////////
/*
 * Generate a new Paillier private key (key pair).
 * @param bits int - number of bits of product 'N'
 * @return key *PaillierPrivateKey - generated private key (key pair)
 * @return err error - error object (or nil if successful)
 */
func NewPaillierPrivateKey(bits int) (key *PaillierPrivateKey, err error) {

	// generate primes 'p' and 'q' and their factor 'n'
	// repeat until the requested factor bitsize is reached
	var p, q, n *big.Int
	for {
		bitsP := (bits - 5) / 2
		bitsQ := bits - bitsP

		p, err = rand.Prime(rand.Reader, bitsP)
		if err != nil {
			return nil, err
		}
		q, err = rand.Prime(rand.Reader, bitsQ)
		if err != nil {
			return nil, err
		}

		n = new(big.Int).Mul(p, q)
		if n.BitLen() == bits {
			break
		}
	}

	// initialize variables
	one := big.NewInt(1)
	n2 := new(big.Int).Mul(n, n)

	// compute public key parameter 'g' (generator)
	g, err := rand.Int(rand.Reader, n2)
	if err != nil {
		return nil, err
	}

	// compute private key parameters	
	p1 := new(big.Int).Sub(p, one)
	q1 := new(big.Int).Sub(q, one)
	l := new(big.Int).Mul(q1, p1)
	l.Div(l, new(big.Int).GCD(nil, nil, p1, q1))

	a := new(big.Int).Exp(g, l, n2)
	a.Sub(a, one)
	a.Div(a, n)
	u := new(big.Int).ModInverse(a, n)

	// return key pair
	pubkey := &PaillierPublicKey{
		N: n,
		G: g,
	}
	prvkey := &PaillierPrivateKey{
		PaillierPublicKey: pubkey,
		L:                 l,
		U:                 u,
		P:                 p,
		Q:                 q,
	}
	return prvkey, nil
}

///////////////////////////////////////////////////////////////////////
// Methods related to the Paillier private key:

/*
 * Get public key from private key.
 * @param self *PaillierPrivateKey - this instance
 * @return *PaillierPublicKey - reference to public key
 */
func (self *PaillierPrivateKey) GetPublicKey() *PaillierPublicKey {
	return self.PaillierPublicKey
}

//---------------------------------------------------------------------
/*
 * Decrypt message with private key.
 * @param c *big.Int - encrypted message
 * @return m *big.Int - decrypted message 
 * @return err error - error object (or nil if successful)
 */
func (self *PaillierPrivateKey) Decrypt(c *big.Int) (m *big.Int, err error) {

	// initialize variables
	pub := self.GetPublicKey()
	n2 := new(big.Int).Mul(pub.N, pub.N)
	one := big.NewInt(1)

	// perform decryption function
	m = new(big.Int).Exp(c, self.L, n2)
	m.Sub(m, one)
	m.Div(m, pub.N)
	m.Mul(m, self.U)
	m.Mod(m, pub.N)
	return m, nil
}

///////////////////////////////////////////////////////////////////////
// Methods related to the Paillier public key:

/*
 * Encrypt message with private key.
 * @param m *big.Int - plaintext message
 * @return c *big.Int - encrypted message 
 * @return err error - error object (or nil if successful)
 */
func (self *PaillierPublicKey) Encrypt(m *big.Int) (c *big.Int, err error) {

	// initialize variables
	n2 := new(big.Int).Mul(self.N, self.N)

	// compute decryption function
	c1 := new(big.Int).Exp(self.G, m, n2)
	r, err := rand.Int(rand.Reader, self.N)
	if err != nil {
		return nil, err
	}
	c2 := new(big.Int).Exp(r, self.N, n2)
	c = new(big.Int).Mul(c1, c2)
	c.Mod(c, n2)
	return c, nil
}
