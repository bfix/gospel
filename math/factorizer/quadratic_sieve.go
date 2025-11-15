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
//*    PGMID.        QUADRATIC SIEVE ALGORITHM.                      */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/03/26.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package factorizer

import (
	"github.com/bfix/gospel/math"
	"github.com/bfix/gospel/math/factorizer/qs"
)

const (
	INTERVAL_SIZE = 1000000
)

// Decompose integer into two (hopefully prime) factors using the quadratic
type QuadraticSieve struct{}

// Find a factor of n.
// @param n - number to be factorized
// @return - factor of n (or nil)
func (qsieve *QuadraticSieve) GetFactor(n *math.Int) *math.Int {
	// allocate process objects
	f := qs.NewFunctionImpl(n)
	fb := qs.NewFactorBaseImpl(n)
	siever := qs.NewSieverImpl(f, fb, INTERVAL_SIZE)
	solver := qs.NewSolverImpl(n, f, fb)

	// Sieve intervals starting at x0 until a solution
	// (factorization) is found.
	for x0 := math.ONE; !solver.Done(); {
		x0 = siever.SieveInterval(x0, solver)
	}

	// return solution.
	return solver.GetSolution()
}
