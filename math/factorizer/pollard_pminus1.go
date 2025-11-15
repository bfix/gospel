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
//*    PGMID.        POLLARD P-1 ALGORITHM.                          */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 07/02/05.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package factorizer

import (
	"github.com/bfix/gospel/math"
)

// Algorithm constants
const (
	PM1_RETRY = 100
	PM1_BMAX  = 10000
)

// Find a factor of number n using Pollards p-1 algorithm.
type Pollard_Pminus1 struct{}

// Pollards p-1 algorithm
// @param n - number to be factorized
// @return - found factor (or null)
func (f *Pollard_Pminus1) GetFactor(n *math.Int) *math.Int {
	Bmax := math.NewInt(PM1_BMAX)

	for range PM1_RETRY {

		// create random coprime to n
		a := math.NewIntRnd(n)
		d := a.GCD(n)
		// make sure numbers are relatively prime.
		if d.Cmp(math.ONE) > 0 {
			// lucky punch: we have a factor!
			return d
		}

		// try different smoothness bounds B
		M := math.ONE
		for B := math.TWO; B.Cmp(Bmax) <= 0; B = B.Add(math.ONE) {

			// compute M = lcm {2..B}
			M = M.Mul(B).Div(M.GCD(B)).Mod(n)

			// compute a^M-1 mod n
			t := a.ModPow(M, n).Sub(math.ONE).Mod(n)
			// compute factor.
			d = t.GCD(n)

			// if the factor is in range [2 .. n-1] it is valid
			if d.Cmp(math.ONE) > 0 && d.Cmp(n) < 0 {
				return d
			}

			// if no factor is found, try with higher B
			if d.Equals(math.ONE) {
				continue
			}

			// if factor is n, retry with another coprime.
			if d.Equals(n) {
				break
			}
		}
	}
	// nothing found.
	return nil
}
