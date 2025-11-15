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
//*    PGMID.        SIEVER INSTANCE INTERFACE.                      */
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

// Handle a sieved relation (x,y).<p>
// @param id int - identifier of siever
// @param x BigInteger - x value (function arg)
// @param y BigInteger - y value (function result)
// @return boolean - relation accepted?
type SieverCallback func(id int, x, y *math.Int) bool

// The Siever interface generically describes the methods of a siever
// instance.<p>
type Siever interface {
	Instance

	// Initialize solver instance.<p>
	// @param id int - siever identifier
	// @param m BigInteger - number to be factorized
	// @param fb FactorBase - factor base to be used
	// @param cb SieverCallback - callback for siever rersponses
	// @return boolean - successful operation?
	Init(id int, m *math.Int, fb FactorBase, cb SieverCallback) bool

	// Set siever interval for siever instance.<p>
	// @param x0 BigInteger - start of interval
	// @param x1 BigInteger - end of interval (inclusive)
	SetSieveInterval(x0, x1 *math.Int)
}

//********************************************************************/
//*    PGMID.        IMPLEMENTATION OF SIEVER INTERFACE.             */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

// Quadratic function used for sieving:  y = x^2 - m = (x'+r)^2 - m<p>
// x is always small so that y < 2m is ensured; x' = 1..x'(max)<p>
type Function struct {
	m *math.Int // number to be factorized
	r *math.Int // floor of square root of m
}

// Instanciate function (and compute/initialize helpers).<p>
// @param n BigInteger - number to be decomposed
func NewFunction(n *math.Int) *Function {
	return &Function{
		m: n,
		r: n.NthRoot(2, false),
	}
}

// Compute function result y = (x+r)^2 - m.<p>
// @param x BigInteger - x argument
// @return BigInteger - function result
func (f *Function) F(x *math.Int) *math.Int {
	return x.Add(f.r).Pow(2).Sub(f.m)
}

// Get expanded argument ("real" square argument).<p>
// @param x BigInteger - x value
// @return BigInteger - squared argument
func (f *Function) sqrArg(x *math.Int) *math.Int {
	return x.Add(f.r)
}

//----------------------------------------------------------------------

// Sieve relations (x,y) with x^2 = y (mod m) and y smooth over factor base
// from a given sieve interval.<p>
type SieverImpl struct {
	InstanceImpl

	cb SieverCallback // reference to response handler

	f  *Function  // (quadratic) function instance
	fb FactorBase // factor base

	ivStart, ivEnd *math.Int   // sieve boundary
	iv             []*math.Int // sub-interval data
	ivSize         int         // sub-interval size
}

// Initialize solver instance.<p>
// @param id int - siever identifier
// @param m BigInteger - number to be factorized
// @param fb FactorBase - factor base to be used
// @param cb SieverCallback - callback for siever rersponses
// @return boolean - successful operation?
func (s *SieverImpl) Init(id int, m *math.Int, fb FactorBase, cb SieverCallback) bool {
	s.Ident(id, "siever")
	s.cb = cb
	s.fb = fb

	// instanciate function object
	s.f = NewFunction(m)

	// allocate siever storage
	s.ivSize = 100000
	s.iv = make([]*math.Int, s.ivSize)
	log.Printf("Sieving interval size is %d", s.ivSize)
	return true
}

// Set siever interval for siever instance.<p>
// @param x0 BigInteger - start of interval
// @param x1 BigInteger - end of interval (inclusive)
func (s *SieverImpl) SetSieveInterval(x0, x1 *math.Int) {
	s.ivStart = x0
	s.ivEnd = x1
}

// Run instance process.<p>
func (s *SieverImpl) Run() {

	log.Println("Starting...")
	s.active = true
	x0 := s.ivStart

	// while we have something to sieve...
	for s.active {

		// compute size of actual sieve interval
		size := s.ivSize
		x1 := x0.Add(math.NewInt(int64(size)))
		if s.ivEnd.Cmp(x1) < 0 {
			size = int(s.ivEnd.Sub(x0).Int64() + 1)
			s.active = false
		}
		//log ("Starting new interval at x' = " + x0);

		// prepare sieving by filling the interval
		// with y values.
		x := x0
		for i := 0; i < size; i++ {
			s.iv[i] = s.f.F(x)
			x = x.Add(math.ONE)
		}

		// sieve with each prime in the factor base.
		for i := 0; i < s.fb.GetNumPrimes(); i++ {
			p := s.fb.GetPrime(i)
			pInt := int(p.Int64())

			// sieving possible?
			ss := s.fb.GetSqrt(i)
			if ss == nil {
				// no sieving for this prime.
				// process all values directly.
				for pos := 0; pos < size; pos++ {
					for s.iv[pos].Mod(p).Equals(math.ZERO) {
						s.iv[pos] = s.iv[pos].Div(p)
					}
				}
				// continue with next prime
				continue
			}

			// calculate the sieving offset parameters
			si := int(ss.Int64())
			xp := int(s.f.sqrArg(x0).Mod(p).Int64())

			// sieve with first solution
			shift := (si - xp + pInt) % pInt
			for pos := shift; pos < size; pos += pInt {
				for s.iv[pos].Mod(p).Equals(math.ZERO) {
					s.iv[pos] = s.iv[pos].Div(p)
				}
			}

			// sieve with second solution
			shift = (2*pInt - si - xp) % pInt
			for pos := shift; pos < size; pos += pInt {
				for s.iv[pos].Mod(p).Equals(math.ZERO) {
					s.iv[pos] = s.iv[pos].Div(p)
				}
			}
		}

		// collect smooth results.
		numSmooth := 0
		for i := 0; i < size; i++ {
			// found new relation?
			if s.iv[i].Equals(math.ONE) {
				numSmooth++
				// pass it to callback
				xx := x0.Add(math.NewInt(int64(i)))
				yy := s.f.F(xx)
				if !s.cb(s.id, s.f.sqrArg(xx), yy) {
					// we will terminate
					s.active = false
					break
				}
			}
		}
		//log (numSmooth + " smooth pairs found in interval");

		// start next interval
		x0 = x1
	}
	log.Println("Terminating...")
	s.cb(s.id, nil, nil)
}
