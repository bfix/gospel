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
//*    PGMID.        RELATION INTERFACE.                             */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package sac

import (
	"fmt"

	"github.com/bfix/gospel/math"
)

// A relation (x,y) with x square and y smooth over factor base.<p>
// Relations can be combined and eventually lead to a y value that is
// a quadratic residue modulo m.<p>
type Relation interface {

	// Initialize new relation for (x,y)<p>
	// @param x BigInteger - argument to square function
	// @param y BigInteger - fb-smooth function result
	Init(x, y *math.Int)

	// Normalize relation over factor base (reduce yh)<p>
	// @param fb FactorBase - factor base for reduction
	// @param m BigInteger - modulus
	Normalize(fb FactorBase, m *math.Int)

	// Multiply this relation with r:  this = this * r<p>
	// @param r Relation - multiplicator from list
	// @param fb FactorBase - factor base for reduction
	// @param m BigInteger - modulus
	Multiply(r Relation, fb FactorBase, m *math.Int)

	// Get smallest prime factor of f.<p>
	// @param fb FactorBase - list of primes
	// @return int - index of first prime
	FirstPrimeIndex(fb FactorBase, pos int) int

	//=================================================================
	// Convenience methods
	//=================================================================

	// Returns (x-s) if relation is square (x^2 = s^2 mod m).<p>
	// @return BigInteger - (x-s) if square or null otherwise
	IsSquared() *math.Int

	// Check if h has factors in factor base fb.<p>
	// @param fb FactorBase - factor base to be used
	// @return boolean - h has factors in fb
	IsReducable(fb FactorBase) bool

	// Check if relation is completely reduced.<p>
	// @return boolean - relation reduced?
	IsReduced() bool

	// Check if f has factors in factor base fb.<p>
	// @param fb FactorBase - factor base to be used
	// @return boolean - f has factors in fb
	IsCovered(fb FactorBase) bool

	// Generate printable representation of relation.<p<
	// @return String - printable (and readable) relation
	String() string
}

//********************************************************************/
//*    PGMID.        IMPLEMENTATION OF RELATION INTERFACE.           */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

// A relation (x,y) with x square and y smooth over factor base.<p>
// Relations can be combined and eventually lead to a y value that is
// a quadratic residue modulo m.<p>
type RelationImpl struct {
	x          *math.Int
	ys, yf, yh *math.Int // y = ys^2 * yf * yh
}

// Initialize new relation for (x,y)<p>
// @param x BigInteger - argument to square function
// @param y BigInteger - fb-smooth function result
func (r *RelationImpl) Init(x, y *math.Int) {
	r.x = x
	r.ys = math.ONE
	r.yf = math.ONE
	r.yh = y
}

// Normalize relation over factor base (reduce yh)<p>
// @param fb FactorBase - factor base for reduction
// @param m BigInteger - modulus
func (r *RelationImpl) Normalize(fb FactorBase, m *math.Int) {

	// if we have no factors in fb
	if !r.IsReducable(fb) {
		return
	}

	// Reduce yh over factor base (include into ys,yf)
	count := fb.GetNumPrimes()
	y := r.yh

	for n := 0; n < count; n++ {
		p := fb.GetPrime(n)
		t := false
		s := y
		for y.Mod(p).Equals(math.ZERO) {
			t = !t
			y = y.Div(p)
		}
		if t {
			s = s.Div(p)
			r.yf = r.yf.Mul(p)
		}
		s = s.Div(y)
		if !s.Equals(math.ONE) {
			s = s.NthRoot(2, false)
			r.ys = r.ys.Mul(s).Mod(m)
		}
	}
	r.yh = y
}

// Multiply this relation with r:  this = this * r<p>
// @param r Relation - multiplicator from list
// @param fb FactorBase - factor base for reduction
// @param m BigInteger - modulus
func (r *RelationImpl) Multiply(re Relation, fb FactorBase, m *math.Int) {
	// cast to implementation type
	ri := re.(*RelationImpl)

	// transform x
	r.x = (r.x.Mul(ri.x)).Mod(m)

	// transform y
	f := r.yf.GCD(ri.yf)
	r.ys = r.ys.Mul(ri.ys).Mod(m)
	r.ys = r.ys.Mul(f).Mod(m)
	r.yf = r.yf.Mul(ri.yf).Div(f.Pow(2))
	// no need to transform yh: relations are reduced
}

// Get smallest prime factor of f.<p>
// @param fb FactorBase - list of primes
// @return int - index of first prime
func (r *RelationImpl) FirstPrimeIndex(fb FactorBase, pos int) int {
	return fb.FirstPrimeIndex(r.yf, pos)
}

//=================================================================
// Convenience methods
//=================================================================

// Returns (x-s) if relation is square (x^2 = s^2 mod m).<p>
// @return BigInteger - (x-s) if square or null otherwise
func (r *RelationImpl) IsSquared() *math.Int {
	if r.yf.Equals(math.ONE) && r.yh.Equals(math.ONE) {
		return r.x.Sub(r.ys).Abs()
	}
	return nil
}

// Check if h has factors in factor base fb.<p>
// @param fb FactorBase - factor base to be used
// @return boolean - h has factors in fb
func (r *RelationImpl) IsReducable(fb FactorBase) bool {
	return fb.Covers(r.yh)
}

// Check if relation is completely reduced.<p>
// @return boolean - relation reduced?
func (r *RelationImpl) IsReduced() bool {
	return r.yh.Equals(math.ONE)
}

// Check if f has factors in factor base fb.<p>
// @param fb FactorBase - factor base to be used
// @return boolean - f has factors in fb
func (r *RelationImpl) IsCovered(fb FactorBase) bool {
	return r.IsReduced() && fb.Covers(r.yf)
}

// Generate printable representation of relation.<p<
// @return String - printable (and readable) relation
func (r *RelationImpl) String() string {
	return fmt.Sprintf("(%v,%v,%v,%v)", r.x, r.ys, r.yf, r.yh)
}
