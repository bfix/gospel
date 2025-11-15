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

package factorizer

import "github.com/bfix/gospel/math"

// Simple point in Montgomery representation (affine coordinates).
type Point struct {
	x, y *math.Int // coordinate values
}

// Point in projective coordionates.
type ProjPoint struct {
	Point
	z *math.Int // projection coordinate
}

// Instanciate projective point from affine coordinates
// @param P - point in affine coordinates
// @return - projective point
func NewProjPoint(p *Point) *ProjPoint {
	pp := new(ProjPoint)
	if p != nil {
		pp.x = p.x
		pp.y = p.y
	}
	pp.z = math.ONE
	return pp
}

// Convert projective coordinate back to affine.
func (pp *ProjPoint) toAffine() *Point {
	return &Point{
		x: pp.x,
		y: math.ZERO, // y information is lost!
	}
}

// Simple elliptic curve in Montgomery representation:
// By^2= x^3 + Ax^2 + x"
type EllipticCurve struct {
	a, b *math.Int // curve parameters
	n    *math.Int // group parameter
	g    *Point    // base point
}

// Instanciate a random elliptic curve over Fn through point P
// @param n - field generator (mod n)
// @param g Point - base point (generator)
func NewEllipticCurve(n *math.Int, g *Point) *EllipticCurve {
	ec := new(EllipticCurve)
	ec.n = n
	ec.a = math.NewIntRndRange(math.THREE, n)
	ec.g = g

	// derive curve parameter "b" with (x,y) on the curve
	// using the Montgomery representation:
	t1 := g.y.ModPow(math.TWO, n)
	t2 := g.x.ModPow(math.THREE, n).Add(g.x.ModPow(math.TWO, n).Mul(ec.a)).Add(g.x).Mod(n)
	ec.b = t2.Mul(t1.ModInverse(n)).Mod(n)
	return ec
}

// Montgomery scalar multiplication on elliptic curve
// @param k - integer factor
// @param p - point on curve to be multiplied
// @return - resulting curve point
func (ec *EllipticCurve) multiply(k *math.Int, p *Point) *Point {
	r := NewProjPoint(nil)
	pp := NewProjPoint(p)
	for _, val := range k.Bytes() {
		for range 8 {
			r = ec.double(r)
			if val&0x80 == 0x80 {
				r = ec.add(pp, r)
			}
			val <<= 1
		}
	}
	return r.toAffine()
}

// Compute dpubled point 2P
// @param pp - point to be doubled
// @return - resulting curve point
func (ec *EllipticCurve) double(pp *ProjPoint) *ProjPoint {
	// x' = (x+z)^2 * (x-z)^2
	// z' = 4*x*z*((x-z)^2 + (A+2)*x*z)
	res := NewProjPoint(nil)
	t1 := pp.x.Add(pp.z).Pow(2).Mod(ec.n) // (x+z)^2
	t2 := pp.x.Sub(pp.z).Pow(2).Mod(ec.n) // (x-z)^2
	t3 := pp.x.Mul(pp.z).Mod(ec.n)        // x*z
	res.x = t1.Mul(t2).Mod(ec.n)
	t4 := t2.Add(ec.a.Add(math.TWO).Mul(t3)).Mod(ec.n)
	res.z = t3.Mul(math.FOUR).Mul(t4).Mod(ec.n)
	return res
}

// Compute sum of two points
// @param p - first point
// @param q - second point
// @return - resulting curve point
func (ec *EllipticCurve) add(p, q *ProjPoint) *ProjPoint {
	// x' = ( (xp-zp)(xq+zq) + (xp+zp)(xq-zq) )^2
	// z' = xr ( (xp-zp)(xq+zq) - (xp+zp)(xq-zq) )^2
	res := NewProjPoint(nil)
	t1 := p.x.Sub(p.z).Mul(q.x.Add(q.z)).Mod(ec.n)
	t2 := p.x.Add(p.z).Mul(q.x.Sub(q.z)).Mod(ec.n)
	res.x = t1.Add(t2).Pow(2).Mod(ec.n)
	res.z = t1.Sub(t2).Pow(2).Mul(ec.g.x).Mod(ec.n)
	return res
}
