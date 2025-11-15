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

//********************************************************************
//*    PGMID.        LENSTRA ECM FACTORIZATION.                      *
//*    AUTHOR.       BERND R. FIX   >Y<                              *
//*    DATE WRITTEN. 07/02/06.                                       *
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       *
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     *
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        *
//*    REMARKS.                                                      *
//********************************************************************

package factorizer

import (
	gmath "math"

	"github.com/bfix/gospel/math"
)

var (
	// ECM parameters for smoothness bound and number of curves to test.
	lenstra_ecm_params = [][]int{
		// digits         B1    curves
		{15, 2000, 25},
		{20, 11000, 90},
		{25, 50000, 300},
		{30, 250000, 700},
		{35, 1000000, 1800},
		{40, 3000000, 5100},
		{45, 11000000, 10600},
		{50, 43000000, 19300},
		{55, 110000000, 49000},
		{60, 260000000, 124000},
		{65, 850000000, 210000},
		{70, 2100000000, 340000},
	}
)

// Lenstra_ECM algorithm
type Lenstra_ECM struct{}

// Find a factor using the Lenstra ECM algorithm
// @param n - number to be factorized
// @return - factor of n (or nil)
func (f *Lenstra_ECM) GetFactor(n *math.Int) *math.Int {
	// Setup parameters (smoothness bounds, number of curves)
	// get number of digits in n
	numDigits := len(n.String())

	// estimate value for smoothness and number of curves.
	numParams := len(lenstra_ecm_params)
	B1 := lenstra_ecm_params[numParams-1][1]
	numCurves := lenstra_ecm_params[numParams-1][2]
	for i := range numParams {
		if lenstra_ecm_params[i][0] > numDigits {
			B1 = lenstra_ecm_params[i][1]
			numCurves = lenstra_ecm_params[i][2]
			break
		}
	}
	B2 := math.NewInt(int64(B1))
	ff := math.NewInt((int64)(gmath.Log(float64(B1)) + 3))
	B2 = B2.Mul(ff)

	// compute e = PI(p<B1) p^k  with k = log(B1)/log(p) and p prime
	kB1 := gmath.Log(float64(B1))
	e := math.ONE
	pi := 2
	for pi < B1 {
		k := (int)(kB1 / gmath.Log(float64(pi)))
		p := math.NewInt(int64(pi))
		e = e.Mul(p.ModPow(math.NewInt(int64(k)), n)).Mod(n)
		pi = int(p.NextProbablePrime(128).Int64())
	}
	// if e is zero, n has a prime factor below B1!
	// make sure you have tested all primes below B1 before.
	if e.Equals(math.ZERO) {
		return nil
	}

	//---------------------------------------------------------------------
	// iterate through all test curves
	//---------------------------------------------------------------------
	for curveCount := 0; curveCount < numCurves; curveCount++ {
		// generate a random initial point "G"
		G := &Point{
			x: math.NewIntRndRange(math.THREE, n),
			y: math.NewIntRndRange(math.THREE, n),
		}

		// instanciate elliptic curve
		ec := NewEllipticCurve(n, G)

		// test primes between B1 and B2
		p := math.NewInt(int64(B1)).NextProbablePrime(128)
		for p.Cmp(B2) < 0 {

			// compute R = peG
			R := ec.multiply(p.Mul(e).Mod(n), G)

			// check if we have a factor.
			g := n.GCD(R.x)

			// evaluate factor
			if g.Equals(n) {
				// curve failed. try next one.
				break
			}
			if g.Cmp(math.ONE) > 0 {
				return g
			}
			// test next prime
			p = p.NextProbablePrime(128)
		}
	}
	// no factor found.
	return nil
}
