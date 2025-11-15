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
//*    PGMID.        GENERIC SIEVER.                                 */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/03/27.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package qs

import (
	"fmt"

	"github.com/bfix/gospel/math"
)

// Interface for (quadratic) sieve implementations.<p>
type Siever interface {
	Init(f Function, fb FactorBase, size int) bool

	SieveInterval(x0 *math.Int, solver Solver) *math.Int
}

type SieverImpl struct {
	f  Function   // (quadratic) function instance
	fb FactorBase // factor base

	iv     []*math.Int // Interval data
	ivSize int         // Interval size
}

func NewSieverImpl(f Function, fb FactorBase, size int) *SieverImpl {
	si := new(SieverImpl)
	si.Init(f, fb, size)
	return si
}

func (si *SieverImpl) Init(f Function, fb FactorBase, size int) bool {
	si.f = f
	si.fb = fb
	si.ivSize = size
	si.iv = make([]*math.Int, si.ivSize)
	//fmt.Printf("   [siever] Sieving interval size is %d\n", si.ivSize)
	return true
}

func (si *SieverImpl) SieveInterval(x0 *math.Int, solver Solver) *math.Int {
	//fmt.Printf("   [siever] Starting new interval at %v\n", x0)

	// prepare sieving by filling the interval
	// with y values.
	x := x0
	for i := 0; i < si.ivSize; i++ {
		x = x.Add(math.ONE)
		si.iv[i] = si.f.F(x)
	}

	// sieve with each prime in the factor base.
	for i := 0; i < si.fb.GetNumPrimes(); i++ {
		p := si.fb.GetPrime(i)
		pInt := p.Int64()

		// sieving possible?
		ss := si.fb.GetSqrt(i)
		if ss == nil {
			// no sieving for this prime.
			// process all values directly.
			for pos := 0; pos < si.ivSize; pos++ {
				for si.iv[pos].Mod(p).Equals(math.ZERO) {
					si.iv[pos] = si.iv[pos].Div(p)
				}
			}
			// continue with next prime
			continue
		}

		// calculate the sieving offset parameters
		s := ss.Int64()
		xp := si.f.ModP(x0, p).Int64()

		// sieve with first solution
		shift := (s - xp + pInt) % pInt
		for pos := shift; pos < int64(si.ivSize); pos += pInt {
			for si.iv[pos].Mod(p).Equals(math.ZERO) {
				si.iv[pos] = si.iv[pos].Div(p)
			}
		}

		// sieve with second solution
		shift = (2*pInt - s - xp) % pInt
		for pos := shift; pos < int64(si.ivSize); pos += pInt {
			for si.iv[pos].Mod(p).Equals(math.ZERO) {
				si.iv[pos] = si.iv[pos].Div(p)
			}
		}
	}

	// collect smooth results.
	numSmooth := 0
	inserted := 0
	for i := 0; i < si.ivSize; i++ {
		// found new relation?
		if si.iv[i].Equals(math.ONE) {
			numSmooth++
			// pass it to solver.
			rc := solver.Process(x0.Add(math.NewInt(int64(i))))
			if rc == SOLVED {
				break
			} else if rc == INSERTED {
				inserted++
			}
		}
	}

	// if we added relations, let the solver
	// make some optimizations.
	if inserted > 0 {
		solver.Optimize()
	}

	fmt.Printf("   [siever] %d smooth pairs found in interval\n", numSmooth)

	// return start of next interval.
	return x0.Add(math.NewInt(int64(si.ivSize)))
}
