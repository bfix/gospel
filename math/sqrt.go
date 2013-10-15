/*
 * Square root of a quadratic residue mod p.
 *
 * (c) 2011-2013 Bernd Fix   >Y<
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

package math

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"errors"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// compute square root of a quadratic residue mod p
// Uses the Shanks-Tonelli algorithm to compute the square root
// see (http://en.wikipedia.org/wiki/Shanks%E2%80%93Tonelli_algorithm)

func Sqrt_modP (a, p *big.Int) (r *big.Int, err error) {
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

///////////////////////////////////////////////////////////////////////
// check if a number is a quadratic residue mod p

func isQuadraticResidue(a, p *big.Int) bool {
	p1 := new(big.Int).Sub(p, ONE)
	exp := new(big.Int).Rsh(p1, 1)
	v := new(big.Int).Exp(a, exp, p)
	return ONE.Cmp(v) == 0
}
