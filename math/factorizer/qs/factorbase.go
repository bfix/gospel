//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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

package qs

import "github.com/bfix/gospel/math"

// Factor base: contains a given number of suitable primes.<p>
// An integer x is smooth if all the prime factors of x are
// contained in the factor base.<p>
type FactorBase interface {
	// Prepare factor base.<p>
	// @param m BigInteger - number to be factorized
	Init(m *math.Int) bool

	// Get number of primes in the factor base.<P>
	// @return int - number of primes in factor base
	GetNumPrimes() int

	// Get the i.th primes in the factor base.<p>
	// @param i int - prime index
	// @return BigInteger - prime value
	GetPrime(i int) *math.Int

	// Solve the equation x^2 = m (mod p)
	// @param i
	// @return
	GetSqrt(i int) *math.Int
}

type FactorBaseImpl struct {
	fbSize int
	fbData []*math.Int
	fbSqrt []*math.Int
}

func NewFactorBaseImpl(m *math.Int) *FactorBaseImpl {
	fb := new(FactorBaseImpl)
	fb.Init(m)
	return fb
}

func (fb *FactorBaseImpl) Init(m *math.Int) bool {

	// The size of the factor base is derived from the
	// number to be decomposed:   fbSize = log2(n)²/10
	// we also compute the solutions for x² - m = 0 (mod p)
	bits := m.BitLen()
	fb.fbSize = (bits * bits) / 10
	//fmt.Printf("   [fb] Generating factor base with %d primes: ",fbSize)

	// allocate arrays
	fb.fbData = make([]*math.Int, fb.fbSize)
	fb.fbSqrt = make([]*math.Int, fb.fbSize)

	// fill factor base
	fb.fbData[0] = math.TWO
	fb.fbSqrt[0] = nil

	tested := 0
	mult := math.ONE
	pf := math.THREE
	for n := 1; n < fb.fbSize; {
		// only primes where m is a quadratic residue (mod p)
		if m.Legendre(pf) == 1 {
			// store the quadratic residue...
			fb.fbSqrt[n], _ = math.SqrtModP(m, pf)
			// ...and the prime for later use.
			fb.fbData[n] = pf
			n++
			mult = mult.Mul(pf)
		}
		tested++
		pf = pf.NextProbablePrime(128)
	}
	// fmt.Println ("done.")
	// fmt.Printf  ("   [fb] Checked first %d primes; largest prime is %v\n", tested, fbData[fbSize-1])
	// fmt.Printf  ("   [fb] Product over all primes in factorbase about 2^%d\n", mult.BitLen())
	return true
}

func (fb *FactorBaseImpl) GetNumPrimes() int {
	return fb.fbSize
}

func (fb *FactorBaseImpl) GetPrime(i int) *math.Int {
	return fb.fbData[i]
}

func (fb *FactorBaseImpl) GetSqrt(i int) *math.Int {
	return fb.fbSqrt[i]
}
