package ed25519

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

	"github.com/bfix/gospel/v2/math"
)

var (
	g  = &Point{c.Gx, c.Gy}
	gm = g.Neg()
)

func TestParam(t *testing.T) {
	i := math.NewInt(121666).ModInverse(c.P)
	j := math.NewInt(121665).Neg().Mod(c.P)
	d := j.Mul(i).Mod(c.P)
	if d.Cmp(c.D) != 0 {
		t.Fatal("Curve.D mismatch (1)")
	}

	x2 := c.Gx.ModPow(math.TWO, c.P)
	y2 := c.Gy.ModPow(math.TWO, c.P)
	i = x2.Mul(y2).Mod(c.P).ModInverse(c.P)
	j = y2.Sub(x2).Sub(math.ONE).Mod(c.P)
	d = j.Mul(i).Mod(c.P)
	if d.Cmp(c.D) != 0 {
		t.Fatal("Curve.D mismatch (2)")
	}

	if math.TWO.ModPow(c.P.Sub(math.ONE), c.P).Cmp(math.ONE) != 0 {
		t.Fatal("2^(p-1) mod p != 1")
	}
	if c.P.Mod(math.FOUR).Cmp(math.ONE) != 0 {
		t.Fatal("p = 1 mod 4")

	}
	if math.TWO.ModPow(c.N.Sub(math.ONE), c.N).Cmp(math.ONE) != 0 {
		t.Fatal("2^(n-1) mod n != 1")
	}
	if c.N.Cmp(math.TWO.Pow(252)) < 0 {
		t.Fatal("n < 2^252")
	}
	if c.N.Cmp(math.TWO.Pow(253)) > 0 {
		t.Fatal("n > 2^253")
	}
	pm1 := c.P.Sub(math.ONE)
	if c.D.ModPow(pm1.Rsh(1), c.P).Cmp(pm1) != 0 {
		t.Fatal("d^(p-1) mod p != p-1")
	}
	if c.i.ModPow(math.TWO, c.P).Cmp(pm1) != 0 {
		t.Fatal("I^2 mod p != p-1")
	}
	if !c.BasePoint().IsOnCurve() {
		t.Fatal("Base point off curve")
	}
	if !c.BasePoint().Mult(c.N).Equals(c.Inf()) {
		t.Fatal("n*G != 0")
	}
}

func TestBase(t *testing.T) {
	if !g.IsOnCurve() {
		t.Fatal("base point not on curve")
	}
	gT := c.BasePoint()
	if !g.Equals(gT) {
		t.Fatal("GetBasePoint failed")
	}
	p := NewPoint(g.x, g.y)
	if !g.Equals(p) {
		t.Fatal("NewPoint failed")
	}
}

func TestInfinity(t *testing.T) {
	p1 := g.Mult(c.N)
	if !p1.IsInf() {
		t.Fatal("n*G is not infinity")
	}
	p1 = g.Add(gm)
	if !p1.IsInf() {
		t.Fatal("g-g is not infinity")
	}
	p1 = g.Add(c.Inf())
	if !p1.Equals(g) {
		t.Fatal("g+0 != g")
	}
	p1 = c.Inf().Mult(math.EIGHT)
	if !p1.IsInf() {
		t.Fatal("8*0 != 0")
	}
}

func TestAdd(t *testing.T) {
	p1 := g.Double()
	p2 := g.Add(p1)
	p3 := p1.Add(g)
	if !p2.Equals(p3) {
		t.Fatal("p+g != g+p")
	}
	p1 = g.Double().Add(g)
	p2 = g.Mult(math.THREE)
	if !p1.Equals(p2) {
		t.Fatal("G+G+G != 3*G")
	}

	for n := 0; n < 32; n++ {
		a := math.NewIntRnd(c.N)
		b := math.NewIntRnd(c.N)
		c := a.Add(b).Mod(c.N)
		p := g.Mult(a)
		q := g.Mult(b)
		r := g.Mult(c)
		p1 = p.Add(q)
		p2 = q.Add(p)
		if !p1.Equals(p2) || !p1.Equals(r) {
			t.Fatal("a*G + b*G != (a+b)*G")
		}
	}
}

func TestMult(t *testing.T) {
	p1 := g.Double()
	mult := func(n *math.Int) *Point {
		p := c.MultBase(n)
		if !p.IsOnCurve() {
			t.Fatalf("point not on curve for %v", n)
		}
		return p
	}
	p2 := mult(math.TWO)
	if !p1.Equals(p2) {
		t.Fatal("mult failed")
	}
	mult(math.THREE)
	mult(math.SEVEN)
	mult(math.EIGHT)
}

func TestCommute(t *testing.T) {
	dp := math.NewIntRndRange(math.THREE, c.N)
	dq := math.NewIntRndRange(math.THREE, c.N)
	p := c.MultBase(dp)
	q := c.MultBase(dq)
	p1 := p.Mult(dq)
	p2 := q.Mult(dp)
	if !p1.Equals(p2) {
		t.Fatal("failed commute")
	}
}

func TestInverse(t *testing.T) {
	for i := 0; i < 20; i++ {
		d := math.NewIntRndRange(math.THREE, c.N)
		di := d.ModInverse(c.N)
		x := d.Mul(di).Mod(c.N)
		if !x.Equals(math.ONE) {
			t.Fatal("failed inverse (1)")
		}
		d2 := d.Rsh(1)
		q := c.MultBase(d)
		q2 := c.MultBase(d2)
		if q2.Double().Equals(q) != (d.Bit(0) == 0) {
			t.Fatal("failed inverse (2)")
		}
	}
}
