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

package qs

import (
	"github.com/bfix/gospel/math"
)

const (
	UNSMOOTH = iota
	TRIVIAL
	INSERTED
	SOLVED
)

type Solver interface {
	Init(n *math.Int, f Function, fb FactorBase) bool

	Process(x *math.Int) int

	Done() bool

	Optimize()

	GetSolution() *math.Int
}

type Relation struct {
	Rel_maxBits int
	x           *math.Int
	ys, yf      *math.Int // y = ys^2 * yf
	fb          FactorBase
	modulus     *math.Int
	firstBit    int
	m           *math.Int
}

func (r *Relation) Set(x, y *math.Int) bool {
	r.x = x

	// Flag odd primes powers in f(x)
	// remember the smallest prime with odd power
	count := r.fb.GetNumPrimes()
	r.firstBit = -1
	r.yf = math.ONE
	z := y
	for n := 0; n < count; n++ {
		p := r.fb.GetPrime(n)
		t := false
		for z.Mod(p).Equals(math.ZERO) {
			t = !t
			z = z.Div(p)
		}
		if t {
			r.yf = r.yf.Mul(p)
			if r.firstBit < 0 {
				r.firstBit = n
			}
		}
	}
	r.ys = math.ONE
	if !y.Equals(r.yf) {
		y.Div(r.yf).NthRoot(2, false)
	}
	return z.Equals(math.ONE)
}

func (r *Relation) Multiply(r2 *Relation) {
	r.x = r.x.Mul(r2.x).Mod(r.m)
	r.ys = r.ys.Mul(r2.ys).Mod(r.m)

	f := r.yf.GCD(r2.yf)
	r.ys = r.ys.Mul(f).Mod(r.m)
	r.yf = r.yf.Mul(r2.yf).Div(f.Pow(2))

	sb := r.yf.BitLen()
	if sb > r.Rel_maxBits {
		r.Rel_maxBits = sb
	}
	r.firstBit++
	for !r.yf.Mod(r.fb.GetPrime(r.firstBit)).Equals(math.ZERO) {
		r.firstBit++
		if r.firstBit == r.fb.GetNumPrimes() {
			r.firstBit = -1
			break
		}
	}
}

type SolverImpl struct {
	Relation

	rows    []*Relation
	numRows int
	numRel  int

	factor *math.Int
	f      Function
}

func NewSolverImpl(m *math.Int, f Function, fb FactorBase) *SolverImpl {
	si := new(SolverImpl)
	si.Init(m, f, fb)
	return si
}

func (si *SolverImpl) Init(m *math.Int, f Function, fb FactorBase) bool {
	si.f = f
	si.fb = fb
	si.modulus = m

	size := fb.GetNumPrimes()
	si.rows = make([]*Relation, size)
	si.numRows = 0
	si.numRel = 0

	return true
}

func (si *SolverImpl) Process(x *math.Int) int {

	// compute function value
	yy := si.f.F(x)
	xx := si.f.SqrArg(x)

	// instanciate relation
	e := new(Relation)
	if !e.Set(xx, yy) {
		return UNSMOOTH
	}
	si.numRel++

	// reduce odd prime powers.
	for {
		// all odd prime powers removed?
		if e.yf.Equals(math.ONE) {
			// Yes: try factorization
			t := si.modulus.GCD(e.x.Sub(e.ys).Abs())
			if !t.Equals(si.modulus) && !t.Equals(math.ONE) {
				// we found a factor!!
				si.factor = t
				// fmt.Printf("   [solver] solution found: %v\n", si.factor)
				// fmt.Printf("   [solver] Max size of yf: %d bits\n", si.Rel_maxBits)
				return SOLVED
			}
			// trival factor ignored.
			return TRIVIAL
		}

		// insert the relation at the (unused) row 'pos'
		// ('pos' is the position of the smallest prime
		// with odd power).
		pos := e.firstBit
		if si.rows[pos] == nil {
			// store in empty slot
			si.rows[pos] = e
			si.numRows++
			// we are done for this relation
			//fmt.Printf ("   [solver] relation #%d of %d (%d)\n", numRows, fb.GetNumPrimes(), numRel)
			return INSERTED
		}

		// Next multiply the input row with the existing row.
		// This removes the smallest odd prime power and the
		// transformed input row is suitable as new input...
		e.Multiply(si.rows[pos])
	}
}

func (si *SolverImpl) Optimize() {}

func (si *SolverImpl) Done() bool {
	return si.factor != nil
}

func (si *SolverImpl) GetSolution() *math.Int {
	return si.factor
}
