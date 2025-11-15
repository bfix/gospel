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

import "github.com/bfix/gospel/math"

type Function interface {

	// Instanciate function (and compute/initialize helpers).<p>
	// @param n BigInteger - number to be decomposed
	Init(n *math.Int) bool

	F(x *math.Int) *math.Int

	SqrArg(x *math.Int) *math.Int

	ModP(a, p *math.Int) *math.Int
}

type FunctionImpl struct {
	m *math.Int // number to be factorized
	r *math.Int // floor of square root of m
}

// Instanciate function (and compute/initialize helpers).<p>
// @param n BigInteger - number to be decomposed
func NewFunctionImpl(n *math.Int) *FunctionImpl {
	fb := new(FunctionImpl)
	fb.Init(n)
	return fb
}

// Prepare function for given integer.<p>
// @param n BigInteger - number to be decomposed
func (f *FunctionImpl) Init(n *math.Int) bool {
	f.m = n
	f.r = n.NthRoot(2, false)
	return true
}

func (f *FunctionImpl) F(x *math.Int) *math.Int {
	return x.Add(f.r).Pow(2).Sub(f.m)
}

func (f *FunctionImpl) ModP(x, p *math.Int) *math.Int {
	return x.Add(f.r).Mod(p)
}

func (f *FunctionImpl) SqrArg(x *math.Int) *math.Int {
	return x.Add(f.r)
}
