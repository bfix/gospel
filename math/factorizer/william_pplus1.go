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
//*    PGMID.        WILLIAM P+1 ALGORITHM.                          */
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
	PP1_MAXSTEP = 100
	PP1_AMAX    = 10000
)

// Find a factor of number n using Williams p+1 algorithm.
type William_Pplus1 struct{}

// Williams p+1 algorithm
// @param n - number to be factorized
// @return - found factor (or nil)
func (f *William_Pplus1) GetFactor(n *math.Int) *math.Int {
	for A := 3; A < PP1_AMAX; A++ {
		B := math.NewInt(int64(A))
		for step := 2; step < PP1_MAXSTEP; step++ {
			V := f.getVi(n, B, step)
			g := n.GCD(V.Sub(math.TWO))
			if g.Cmp(math.ONE) > 0 {
				return g
			}
			B = V
		}
	}
	// nothing found.
	return nil
}

// Get indexed element from sequence for base B.
// @param n - number to be factorized
// @param B - base value for sequence
// @param idx - index into sequence
// @return - sequence element
func (f *William_Pplus1) getVi(n, B *math.Int, idx int) *math.Int {
	x := B
	y := B.Pow(2).Sub(math.TWO).Mod(n)
	i := math.NewInt(int64(idx))
	for pos := i.BitLen() - 1; pos >= 0; pos-- {
		if i.Bit(pos) == 1 {
			x = x.Mul(y).Sub(B).Mod(n)
			y = y.Pow(2).Sub(math.TWO).Mod(n)
		} else {
			y = x.Mul(y).Sub(B).Mod(n)
			x = x.Pow(2).Sub(math.TWO).Mod(n)
		}
	}
	return x
}
