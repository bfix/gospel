package crypto

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

import (
	"github.com/bfix/gospel/math"
)

// FieldP is a prime field
type FieldP struct {
	P *math.Int
}

// Random generates a random field value
func (f *FieldP) Random() *math.Int {
	return math.NewIntRnd(f.P.Sub(math.ONE))
}

// Add field values
func (f *FieldP) Add(a, b *math.Int) *math.Int {
	return a.Add(b).Mod(f.P)
}

// Sub subtracts field values
func (f *FieldP) Sub(a, b *math.Int) *math.Int {
	return f.P.Add(a).Sub(b).Mod(f.P)
}

// Neg negates a field value
func (f *FieldP) Neg(a *math.Int) *math.Int {
	return f.P.Sub(a)
}

// Mul multiplies field values
func (f *FieldP) Mul(a, b *math.Int) *math.Int {
	return a.Mul(b).Mod(f.P)
}

// Div divides field values
func (f *FieldP) Div(a, b *math.Int) *math.Int {
	return b.ModInverse(f.P).Mul(a)
}
