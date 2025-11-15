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

//********************************************************************/
//*    PGMID.        FACTOR BASE INTERFACE.                          */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package sac

import (
	"log"

	"github.com/bfix/gospel/math"
)

// Factor base: contains a given number of suitable primes.<p>
// A prime 'p' can be included in the factor base if 'm' (the number
// to be factorized) is a quadratic residue modulo p.<p>
// An integer y is smooth over fb if all the prime factors of y are
// contained in the factor base fb.<p>
type FactorBase interface {

	// Prepare factor base.<p>
	// @param m BigInteger - number to be factorized
	// @return boolean - successful operation?
	Init(m *math.Int) bool

	// Get number of primes in the factor base.<P>
	// @return int - number of primes in factor base
	GetNumPrimes() int

	// Get the i.th primes in the factor base.<p>
	// @param i int - prime index
	// @return BigInteger - prime value
	GetPrime(i int) *math.Int

	// Solve the equation x^2 = m (mod p)
	// @param i int - prime index in factor base
	// @return BigInteger - solution (x)
	GetSqrt(i int) *math.Int

	// Get index of first (smallest) prime from factor base that
	// is a factor of f.<p>
	// @param f BigInteger - number (unsquare part of y)
	// @param pos int - last index (smaller than new index)
	// @return int - index of first occurring prime
	FirstPrimeIndex(f *math.Int, pos int) int

	// Extract a sub factor base.<p>
	// @param id int - identifier of sub factor base
	// @param start int - offset into full factor base
	// @param size int - number of successive primes
	// @return FactorBase - extracted sub factor base
	GetSubBase(id, start, size int) FactorBase

	// Check if number has a prime factor in factor base.<p>
	// @param a BigInteger - number to be tested
	// @return boolean - number has factor in factor base
	Covers(a *math.Int) bool

	// Get identifier of factor base.<p>
	// @return int - identifier
	GetId() int
}

//********************************************************************/
//*    PGMID.        FACTOR BASE IMPLEMENTATION.                     */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

// Factor base: contains a given number of suitable primes.<p>
// A prime 'p' can be included in the factor base if 'm' (the number
// to be factorized) is a quadratic residue modulo p.<p>
// An integer y is smooth over fb if all the prime factors of y are
// contained in the factor base fb.<p>
type FactorBaseImpl struct {
	Instance

	fbSize int         // number of primes
	fbData []*math.Int // list of primes p
	fbSqrt []*math.Int // list of x with x^2 = m (mod p)
	FbMult *math.Int   // multiplicative = Prod(i=1,B) p(i)
	fbId   int         // factor base identifier
}

// Prepare factor base.<p>
// @param m BigInteger - number to be factorized
// @return boolean - successful operation?
func (fb *FactorBaseImpl) Init(m *math.Int) bool {
	fb.Ident(-1, "")
	fb.fbId = 1

	// The size of the factor base is derived from the
	// number to be decomposed:   fbSize = log2(n)²/10
	// we also compute the solutions for x² - m = 0 (mod p)
	bits := m.BitLen()
	fb.fbSize = (bits * bits) / 10
	log.Printf("[fb %d] Generating factor base with %d primes...", fb.fbId, fb.fbSize)

	// allocate arrays
	fb.fbData = make([]*math.Int, fb.fbSize)
	fb.fbSqrt = make([]*math.Int, fb.fbSize)

	// fill factor base
	fb.fbData[0] = math.TWO
	fb.fbSqrt[0] = nil

	tested := 0
	fb.FbMult = math.ONE
	pf := math.THREE
	for n := 1; n < fb.fbSize; {
		// only primes where m is a quadratic residue (mod p)
		if m.Legendre(pf) == 1 {
			// store the quadratic residue...
			fb.fbSqrt[n], _ = math.SqrtModP(m, pf)
			// ...and the prime for later use.
			fb.fbData[n] = pf
			n++
			fb.FbMult = fb.FbMult.Mul(pf)
		}
		tested++
		pf = pf.NextProbablePrime(128)
	}
	log.Printf("[fb] Checked first %d primes; largest prime is %v", tested, fb.fbData[fb.fbSize-1])
	return true
}

// Get number of primes in the factor base.<P>
// @return int - number of primes in factor base
func (fb *FactorBaseImpl) GetNumPrimes() int {
	return fb.fbSize
}

// Get the i.th primes in the factor base.<p>
// @param i int - prime index
// @return BigInteger - prime value
func (fb *FactorBaseImpl) GetPrime(i int) *math.Int {
	return fb.fbData[i]
}

// Solve the equation x^2 = m (mod p)
// @param i int - prime index in factor base
// @return BigInteger - solution (x)
func (fb *FactorBaseImpl) GetSqrt(i int) *math.Int {
	return fb.fbSqrt[i]
}

// Get index of first (smallest) prime from factor base that
// is a factor of f.<p>
// @param f BigInteger - number (unsquare part of y)
// @param pos int - last index (smaller than new index)
// @return int - index of first occurring prime
func (fb *FactorBaseImpl) FirstPrimeIndex(f *math.Int, pos int) int {

	// if we could just get rid of this iteration...
	for n := pos + 1; n < fb.fbSize; n++ {
		p := fb.fbData[n]
		if f.Mod(p).Equals(math.ZERO) {
			return n
		}
	}
	return -1
}

// Extract a sub factor base.<p>
// @param id int - identifier of sub factor base
// @param start int - offset into full factor base
// @param size int - number of successive primes
// @return FactorBase - extracted sub factor base
func (fb *FactorBaseImpl) GetSubBase(id, start, size int) FactorBase {

	// instanciate sub factor base
	res := new(FactorBaseImpl)
	res.fbId = id
	res.fbSize = size

	// allocate arrays
	res.fbData = make([]*math.Int, res.fbSize)
	res.fbSqrt = make([]*math.Int, res.fbSize)

	// fill sub factor base
	res.FbMult = math.ONE
	for n := 0; n < res.fbSize; n++ {
		res.fbData[n] = fb.fbData[start+n]
		res.FbMult = res.FbMult.Mul(res.fbData[n])
		res.fbSqrt[n] = fb.fbSqrt[start+n]
	}
	return res
}

// Check if number has a prime factor in factor base.<p>
// @param a BigInteger - number to be tested
// @return boolean - number has factor in factor base
func (fb *FactorBaseImpl) Covers(a *math.Int) bool {
	return !fb.FbMult.GCD(a).Equals(math.ONE)
}

// Get identifier of factor base.<p>
// @return int - identifier
func (fb *FactorBaseImpl) GetId() int {
	return fb.fbId
}
