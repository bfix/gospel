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

package ed25519

import (
	"fmt"

	"github.com/bfix/gospel/math"
)

// Curve is the Ed25519 elliptic curve (twisted Edwards curve):
//
//	a x^2 + y^2 = 1 + d x^2 y^2,  d = -121665/121666, a = -1
type Curve struct {
	// P is the generator of the underlying field "F_p"
	// = 2^255 - 19
	P *math.Int
	// N is the order of G
	N *math.Int
	// curve D = 121665/121666
	D *math.Int
	// Gx is the x-coord of the base point
	Gx *math.Int
	// Gy is the y-coord of the base point
	Gy *math.Int
	// Ox is the x-coord of the identity point (Infinity)
	Ox *math.Int
	// Oy is the y-coord of the identity point (Infinity)
	Oy *math.Int
	// e = (P+3) / 8
	e *math.Int
	// i = 2^(P-1)/4 mod P
	i *math.Int
}

var (
	// Curve is the reference to a curve instance (meant as singleton)
	c = &Curve{
		P:  math.NewIntFromHex("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"),
		N:  math.NewIntFromHex("1000000000000000000000000000000014def9dea2f79cd65812631a5cf5d3ed"),
		D:  math.NewIntFromHex("52036cee2b6ffe738cc740797779e89800700a4d4141d8ab75eb4dca135978a3"),
		Gx: math.NewIntFromHex("216936d3cd6e53fec0a4e231fdd6dc5c692cc7609525a7b2c9562d608f25d51a"),
		Gy: math.NewIntFromHex("6666666666666666666666666666666666666666666666666666666666666658"),
		Ox: math.ZERO,
		Oy: math.ONE,
		e:  math.NewIntFromHex("0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe"),
		i:  math.NewIntFromHex("2b8324804fc1df0b2b4d00993dfbd7a72f431806ad2fe478c4ee1b274a0ea0b0"),
	}
)

// GetCurve returns the singleton curve instance
func GetCurve() *Curve {
	return c
}

// BasePoint returns the base Point of the curve
func (c *Curve) BasePoint() *Point {
	return NewPoint(c.Gx, c.Gy)
}

// Inf returns the identity point (infinity)
func (c *Curve) Inf() *Point {
	return NewPoint(c.Ox, c.Oy)
}

// MultBase multiplies the base Point of the curve with a scalar value k
func (c *Curve) MultBase(k *math.Int) *Point {
	return c.BasePoint().Mult(k)
}

// SolveX returns the positive solution of the curve equation for given y-coordinate
func (c *Curve) SolveX(y *math.Int) *math.Int {
	// compute 'x = +√((y^2 - 1) / (d y^2 + 1))
	y2 := y.ModPow(math.TWO, c.P)
	dy2 := c.D.Mul(y2).Mod(c.P)
	nom := y2.Sub(math.ONE).Mod(c.P)
	den := dy2.Add(math.ONE).Mod(c.P)
	x2 := den.ModInverse(c.P).Mul(nom).Mod(c.P)
	x := x2.ModPow(c.e, c.P)
	if x.Mul(x).Sub(x2).Mod(c.P).Cmp(math.ZERO) != 0 {
		x = x.Mul(c.i).Mod(c.P)
	}
	if x.Bit(0) != 0 {
		x = c.P.Sub(x)
	}
	return x
}

// Point (x,y) on the curve
type Point struct { // exported Point type
	x, y *math.Int // coordinate values
}

// NewPoint instaniates a new Point
func NewPoint(a, b *math.Int) *Point {
	return &Point{x: a, y: b}
}

// NewPointFromBytes reconstructs a Point from binary representation.
func NewPointFromBytes(b []byte) (p *Point, err error) {
	buf := reverse(b)
	neg := (buf[0] >> 7) == 1
	buf[0] &= 0x7f
	y := math.NewIntFromBytes(buf)
	x := c.SolveX(y)
	if neg {
		x = c.P.Sub(x)
	}
	return NewPoint(x, y), nil
}

// X returns the x-coordinate of a point.
func (p *Point) X() *math.Int {
	return p.x
}

// Y returns the y-coordinate of a point.
func (p *Point) Y() *math.Int {
	return p.y
}

// Neg returns -P for the point P.
func (p *Point) Neg() *Point {
	x := p.x.Neg().Mod(c.P)
	x2 := p.x.ModPow(math.TWO, c.P)
	dx2 := c.D.Mul(x2).Mod(c.P)
	nom := x2.Add(math.ONE)
	den := math.ONE.Sub(dx2).Mul(p.y).Mod(c.P)
	y := nom.Mul(den.ModInverse(c.P)).Mod(c.P)
	return NewPoint(x, y)
}

// String returns a human-readable representation of a point.
func (p *Point) String() string {
	return fmt.Sprintf("(%v,%v)", p.x, p.y)
}

// IsOnCurve checks if a Point (x,y) is on the curve
func (p *Point) IsOnCurve() bool {
	x2 := p.x.ModPow(math.TWO, c.P)
	y2 := p.y.ModPow(math.TWO, c.P)
	tl := y2.Sub(x2).Mod(c.P)
	tr := math.ONE.Add(c.D.Mul(x2).Mod(c.P).Mul(y2).Mod(c.P))
	return tl.Cmp(tr) == 0
}

// Equals checks if two Points are equal
func (p *Point) Equals(q *Point) bool {
	return p.x.Cmp(q.x) == 0 && p.y.Cmp(q.y) == 0
}

// IsInf checks if a Point is at infinity
func (p *Point) IsInf() bool {
	return p.x.Cmp(c.Ox) == 0 && p.y.Cmp(c.Oy) == 0
}

// Add two Points on the curve
func (p *Point) Add(q *Point) *Point {
	_p := newPrjPoint(p)
	_q := newPrjPoint(q)
	return _p.add(_q).conv()
}

// Double a Point on the curve
func (p *Point) Double() *Point {
	_p := newPrjPoint(p)
	return _p.double().conv()
}

// Mult multiplies a Point on the curve with a scalar value k using
// a Montgomery multiplication approach
func (p *Point) Mult(k *math.Int) *Point {
	_p := newPrjPoint(p)
	return _p.mult(k).conv()
}

// Bytes returns a byte representation of Point (compressed or uncompressed).
func (p *Point) Bytes() []byte {
	buf := make([]byte, 32)
	b := p.Y().Bytes()
	off := 32 - len(b)
	copy(buf[off:], b)
	if p.X().Bit(0) == 1 {
		buf[0] |= 0x80
	}
	return reverse(buf)
}

//----------------------------------------------------------------------
// Projective coordinates
//----------------------------------------------------------------------

// prjPoint (x,y,z) on the curve
type prjPoint struct {
	x, y, z *math.Int // coordinate values
}

// Convert affine point to projective coordinates
func newPrjPoint(p *Point) *prjPoint {
	return &prjPoint{
		x: p.x,
		y: p.y,
		z: math.ONE,
	}
}

// String returns a human-readable representation of a point.
func (p *prjPoint) String() string {
	return fmt.Sprintf("(%v,%v,%v)", p.x, p.y, p.z)
}

// Convert projective coordinates to affine point.
func (p *prjPoint) conv() *Point {
	zi := p.z.ModInverse(c.P)
	return NewPoint(p.x.Mul(zi).Mod(c.P), p.y.Mul(zi).Mod(c.P))
}

// Add two projective points
// (see https://hyperelliptic.org/EFD/g1p/data/twisted/projective/addition/add-2008-bbjlp)
//     A = Z1*Z2
//     B = A2
//     C = X1*X2
//     D = Y1*Y2
//     E = d*C*D
//     F = B-E
//     G = B+E
//     X3 = A*F*((X1+Y1)*(X2+Y2)-C-D)
//     Y3 = A*G*(D-a*C)
//     Z3 = F*G

func (p *prjPoint) add(q *prjPoint) *prjPoint {
	_a := p.z.Mul(q.z)
	_b := _a.Mul(_a)
	_c := p.x.Mul(q.x)
	_d := p.y.Mul(q.y)
	_e := c.D.Mul(_c.Mul(_d))
	_f := _b.Sub(_e)
	_g := _b.Add(_e)
	_h := p.x.Add(p.y).Mul(q.x.Add(q.y))
	_i := _h.Sub(_c.Add(_d))
	_j := _d.Add(_c)
	return &prjPoint{
		x: _a.Mul(_f.Mul(_i)).Mod(c.P),
		y: _a.Mul(_g.Mul(_j)).Mod(c.P),
		z: _f.Mul(_g).Mod(c.P),
	}
}

// Doubling a projective point.
// (see https://hyperelliptic.org/EFD/g1p/data/twisted/projective/doubling/dbl-2008-bbjlp)
//
//	B = (X1+Y1)2
//	C = X12
//	D = Y12
//	E = a*C
//	F = E+D
//	H = Z12
//	J = F-2*H
//	X3 = (B-C-D)*J
//	Y3 = F*(E-D)
//	Z3 = F*J
func (p *prjPoint) double() *prjPoint {
	_b := p.x.Add(p.y)
	_b = _b.Mul(_b)
	_c := p.x.Mul(p.x)
	_d := p.y.Mul(p.y)
	_e := _c.Neg()
	_f := _e.Add(_d)
	_h := p.z.Mul(p.z)
	_j := _f.Sub(math.TWO.Mul(_h))
	return &prjPoint{
		x: _j.Mul(_b.Sub(_c.Add(_d))).Mod(c.P),
		y: _f.Mul(_e.Sub(_d)).Mod(c.P),
		z: _j.Mul(_f).Mod(c.P),
	}
}

// Scalar multiplication of a curve point
func (p *prjPoint) mult(k *math.Int) *prjPoint {
	r := &prjPoint{c.Ox, c.Oy, math.ONE}
	x := &prjPoint{c.Ox, c.Oy, math.ONE}
	for _, val := range k.Bytes() {
		for pos := 0; pos < 8; pos++ {
			r = r.double()
			if val&0x80 == 0x80 {
				r = p.add(r)
			} else {
				x = p.add(x)
			}
			val <<= 1
		}
	}
	return r
}
