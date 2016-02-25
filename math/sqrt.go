package math

import (
	"errors"
	"math/big"
)

// SqrtModP computes the square root of a quadratic residue mod p
// It uses the Shanks-Tonelli algorithm to compute the square root
// see (http://en.wikipedia.org/wiki/Shanks%E2%80%93Tonelli_algorithm)
func SqrtModP(a, p *big.Int) (r *big.Int, err error) {
	r = ZERO
	err = nil
	if !isQuadraticResidue(a, p) {
		err = errors.New("No quadratic residue")
		return
	}

	s := 0
	p1 := new(big.Int).Sub(p, ONE)
	m := p1
	for m.Bit(0) == 0 {
		s++
		m = new(big.Int).Rsh(m, 1)
	}
	m2 := new(big.Int).Add(m, ONE)
	m2.Rsh(m2, 1)
	z := big.NewInt(1)
	for isQuadraticResidue(z, p) {
		z.Add(z, ONE)
	}
	c := new(big.Int).Exp(z, m, p)
	u := new(big.Int).Exp(a, m, p)
	r = new(big.Int).Exp(a, m2, p)
	if s < 2 {
		return
	}

	pow := new(big.Int).Lsh(ONE, uint(s-2))
	for i := 1; i < s; i++ {
		c2 := new(big.Int).Mul(c, c)
		uu := new(big.Int).Exp(u, pow, p)
		if uu.Cmp(p1) == 0 {
			u.Mul(u, c2)
			r.Mul(r, c)
			pow.Rsh(pow, 1)
		}
		c = c2
	}
	return
}

// check if a number is a quadratic residue mod p
func isQuadraticResidue(a, p *big.Int) bool {
	p1 := new(big.Int).Sub(p, ONE)
	exp := new(big.Int).Rsh(p1, 1)
	v := new(big.Int).Exp(a, exp, p)
	return ONE.Cmp(v) == 0
}
