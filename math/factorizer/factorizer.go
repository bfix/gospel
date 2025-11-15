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
//*    PGMID.        INTEGER PRIME DECOMPOSER.                       */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 07/02/05.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package factorizer

import (
	"slices"

	"github.com/bfix/gospel/math"
)

// list of known algorithms
const (
	POLLARD_RHO = iota
	POLLARD_PMINUS1
	WILLIAM_PPLUS1
	LENSTRA_ECM
	QUADRATIC_SIEVE
)

// FactorFinder interface implement by different algorithms
type FactorFinder interface {
	GetFactor(n *math.Int) *math.Int
}

// Factorizer uses various algorithm for integer factorization
type Factorizer struct {
	applied    []int
	algorithms map[int]FactorFinder
}

// Instantiate new Factorizer with given algorithms.
// @param algs - list of algorithm identifiers
func NewFactorizer(algs ...int) *Factorizer {
	fac := &Factorizer{
		applied:    slices.Clone(algs),
		algorithms: make(map[int]FactorFinder),
	}
	fac.algorithms[POLLARD_RHO] = new(Pollard_rho)
	fac.algorithms[POLLARD_PMINUS1] = new(Pollard_Pminus1)
	fac.algorithms[WILLIAM_PPLUS1] = new(William_Pplus1)
	fac.algorithms[LENSTRA_ECM] = new(Lenstra_ECM)
	fac.algorithms[QUADRATIC_SIEVE] = new(QuadraticSieve)
	return fac
}

// MAX_SMALL is the number of small primes to be (always) checked
var MAX_SMALL = math.NewInt(25000)

// Check next small prime factor
// @param n - number to be factorized
// @return rem - reminder (after n is divided by all found primes)
// @return list - list of found prime factors
func (f *Factorizer) smallPrimes(n *math.Int) (rem *math.Int, list []*math.Int) {
	rem = n
	for p := math.TWO; p.Cmp(MAX_SMALL) < 0; p = p.NextProbablePrime(128) {
		for rem.Mod(p).Equals(math.ZERO) {
			rem = rem.Div(p)
			list = append(list, p)
		}
	}
	return
}

// Decompose an integer value into its prime factors
// @param n - number to be factorized
// @return list - (unordered) list of prime factors
func (f *Factorizer) Decompose(n *math.Int) (list []*math.Int) {
	// Check for small prime factors
	n, list = f.smallPrimes(n)

	// apply algorithms to factorize n
	for pos, idx := range f.applied {
		// check for completion
		if n.Equals(math.ONE) {
			return
		}
		if n.ProbablyPrime(128) {
			break
		}
		// get factor with current algorithm.
		factor := f.algorithms[idx].GetFactor(n)

		// no factor found: try next algorithm
		if factor == nil {
			pos++
			continue
		}

		// check if found factor is prime.
		if factor.ProbablyPrime(128) {
			// yes: add to list of factors
			list = append(list, factor)
		} else {
			// decompose factor.
			v := f.Decompose(factor)
			list = append(list, v...)
		}

		// reduce value to be decomposed
		n = n.Div(factor)
	}
	// we have found the last factor - time to pass
	// back the list of primes to the caller.
	list = append(list, n)
	return
}
