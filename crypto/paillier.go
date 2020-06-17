package crypto

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

import (
	"github.com/bfix/gospel/math"
)

// PaillierPublicKey data structure
type PaillierPublicKey struct {
	N, G *math.Int
}

// PaillierPrivateKey data structure
type PaillierPrivateKey struct {
	*PaillierPublicKey
	L, U *math.Int
	P, Q *math.Int
}

// NewPaillierPrivateKey generates a new Paillier private key (key pair).
//
// The key used in the Paillier crypto system consists of four integer
// values. The public key has two parameters; the private key has three
// parameters (one parameter is shared between the keys). As in RSA it
// starts with two random primes 'p' and 'q'; the public key parameter
// are computed as:
//
//   n := p * q
//   g := random number from interval [0,n^2[
//
// The private key parameters are computed as:
//
//   n := p * q
//   l := lcm (p-1,q-1)
//   u := (((g^l mod n^2)-1)/n) ^-1 mod n
//
// N.B. The division by n is integer based and rounds toward zero!
func NewPaillierPrivateKey(bits int) (key *PaillierPrivateKey, err error) {

	// generate primes 'p' and 'q' and their factor 'n'
	// repeat until the requested factor bitsize is reached
	var p, q, n *math.Int
	for {
		bitsP := (bits - 5) / 2
		bitsQ := bits - bitsP

		p = math.NewIntRndPrimeBits(bitsP)
		q = math.NewIntRndPrimeBits(bitsQ)
		n = p.Mul(q)
		if n.BitLen() == bits {
			break
		}
	}

	// initialize variables
	n2 := n.Mul(n)

	// compute public key parameter 'g' (generator)
	g := math.NewIntRnd(n2)

	// compute private key parameters
	p1 := p.Sub(math.ONE)
	q1 := q.Sub(math.ONE)
	l := q1.Mul(p1).Div(p1.GCD(q1))

	a := g.ModPow(l, n2).Sub(math.ONE).Div(n)
	u := a.ModInverse(n)

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

// GetPublicKey returns the corresponding public key from a private key.
func (p *PaillierPrivateKey) GetPublicKey() *PaillierPublicKey {
	return p.PaillierPublicKey
}

// Decrypt message with private key.
//
// The decryption function in the Paillier crypto scheme is:
//
//   m = D(c) = ((((c^l mod n^2)-1)/n) * u) mod n
//
// N.B. As in the key generation process the division by n is integer
//      based and rounds toward zero!
func (p *PaillierPrivateKey) Decrypt(c *math.Int) (m *math.Int, err error) {

	// initialize variables
	pub := p.GetPublicKey()
	n2 := pub.N.Mul(pub.N)

	// perform decryption function
	m = c.ModPow(p.L, n2).Sub(math.ONE).Div(pub.N).Mul(p.U).Mod(pub.N)
	return m, nil
}

// Encrypt message with public key.
//
// The encryption function in the Paillier crypto scheme is:
//
//   c = E(m) = (g^m * r^n) mod n^2
//
// where 'r' is a random number from the interval [0,n[. This encryption
// allows different encryption results for the same message, based on
// the actual value of 'r'.
func (p *PaillierPublicKey) Encrypt(m *math.Int) (c *math.Int, err error) {

	// initialize variables
	n2 := p.N.Mul(p.N)

	// compute decryption function
	c1 := p.G.ModPow(m, n2)
	r := math.NewIntRnd(p.N)
	c2 := r.ModPow(p.N, n2)
	c = c1.Mul(c2).Mod(n2)
	return c, nil
}
