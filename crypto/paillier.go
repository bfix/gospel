package crypto

import (
	"crypto/rand"
	"math/big"
)

// PaillierPublicKey data structure
type PaillierPublicKey struct {
	N, G *big.Int
}

// PaillierPrivateKey data structure
type PaillierPrivateKey struct {
	*PaillierPublicKey
	L, U *big.Int
	P, Q *big.Int
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
func (p *PaillierPrivateKey) Decrypt(c *big.Int) (m *big.Int, err error) {

	// initialize variables
	pub := p.GetPublicKey()
	n2 := new(big.Int).Mul(pub.N, pub.N)
	one := big.NewInt(1)

	// perform decryption function
	m = new(big.Int).Exp(c, p.L, n2)
	m.Sub(m, one)
	m.Div(m, pub.N)
	m.Mul(m, p.U)
	m.Mod(m, pub.N)
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
func (p *PaillierPublicKey) Encrypt(m *big.Int) (c *big.Int, err error) {

	// initialize variables
	n2 := new(big.Int).Mul(p.N, p.N)

	// compute decryption function
	c1 := new(big.Int).Exp(p.G, m, n2)
	r, err := rand.Int(rand.Reader, p.N)
	if err != nil {
		return nil, err
	}
	c2 := new(big.Int).Exp(r, p.N, n2)
	c = new(big.Int).Mul(c1, c2)
	c.Mod(c, n2)
	return c, nil
}
