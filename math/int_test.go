package math

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
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
	"testing"
)

func TestIntBytes(t *testing.T) {
	c := TWO.Pow(256)
	for i := 0; i < 1000; i++ {
		a := NewIntRnd(c)
		b := NewIntFromBytes(a.Bytes())
		if !a.Equals(b) {
			t.Fatal("Bytes()/NewIntFromBytes() failed")
		}
	}
}

func TestExtendedEuclid(t *testing.T) {
	var (
		a, b *Int
		m    = NewInt(1000000000000000000)
	)
	test := func() {
		r := a.ExtendedEuclid(b)
		s := r[0].Mul(a).Add(r[1].Mul(b))
		if !s.Equals(ONE) {
			t.Fail()
		}
	}
	for i := 0; i < 10; {
		a = NewIntRnd(m).Add(ONE)
		b = NewIntRnd(a).Add(ONE)
		if !a.GCD(b).Equals(ONE) {
			continue
		}
		test()
		a, b = b, a
		test()
		i++
	}
}

func TestSqrt(t *testing.T) {
	p := NewIntRndPrimeBits(10)
	count := 0
	for i := 0; i < 1000; i++ {
		g := NewIntRnd(p)
		if g.Legendre(p) == 1 {
			count++
			h, err := SqrtModP(g, p)
			if err != nil {
				t.Fatal(err)
			}
			gg := h.ModPow(TWO, p)
			if !gg.Equals(g) {
				t.Fatalf("result error: %v != %v", g, gg)
			}
		}
	}
}
