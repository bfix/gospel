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
//*    PGMID.        POLLARD RHO ALGORITHM.                          */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 07/02/05.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package factorizer

import "github.com/bfix/gospel/math"

// Algorithm constants
const (
	RHO_RETRY = 100
	RHO_LOOP  = 8192
)

// Find a factor of number n using Pollards rho algorithm.
type Pollard_rho struct{}

// Pollards rho algorithm
// @param n - number to be factorized
// @return - found factor (or nil)
func (f *Pollard_rho) GetFactor(n *math.Int) *math.Int {

	// setup variables
	x := math.TWO
	y := math.TWO
	d := math.ONE
	rnd := math.NewIntRndRange(math.THREE, n)

	// try a number of different pseudo randoms...
	for range RHO_RETRY {

		// use a kind of Floyd's cycle finding algorithm
		// to find a factor of n.
		for loop := 0; d.Equals(math.ONE) && loop < RHO_LOOP; loop++ {
			x = x.ModPow(rnd, n)
			y = y.ModPow(rnd, n).ModPow(rnd, n)
			d = n.GCD(x.Sub(y).Abs())
		}
		// if the factor is in range [2 .. n-1]
		// it's valid (and likely to be prime)
		if d.Cmp(math.ONE) > 0 && d.Cmp(n) < 0 {
			// yes: return factor.
			return d
		}
		// our pseudo random sequence doesn't work. try another one...
		rnd = math.NewIntRndRange(math.THREE, n)
	}
	// algorithm failed on input value.
	return nil
}
