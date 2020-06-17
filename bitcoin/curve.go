package bitcoin

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
	"errors"
	"fmt"

	"github.com/bfix/gospel/math"
)

// Curve is the elliptic curve used for Bitcoin (Secp256k1)
type Curve struct {
	// P is the generator of the underlying field "F_p"
	// = 2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1
	P *math.Int
	// Gx is the x-coord of the base point
	Gx *math.Int
	// Gy is the y-coord of the base point
	Gy *math.Int
	// N is the order of G
	N *math.Int
	// curve parameter (=7)
	B *math.Int
}

var (
	// Curve is the reference to a curve instance (meant as singleton)
	c = &Curve{
		P:  math.NewIntFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"),
		Gx: math.NewIntFromHex("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
		Gy: math.NewIntFromHex("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8"),
		N:  math.NewIntFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"),
		B:  math.SEVEN,
	}
)

// GetCurve returns the singleton curve instance
func GetCurve() *Curve {
	return c
}

// GetBasePoint returns the base Point of the curve
func GetBasePoint() *Point {
	return NewPoint(c.Gx, c.Gy)
}

// MultBase multiplies the base Point of the curve with a scalar value k
func MultBase(k *math.Int) *Point {
	return GetBasePoint().Mult(k)
}

// Solve the curve equation for given x-ccordinate (returns positive y)
func Solve(x *math.Int) (*math.Int, bool) {
	// compute 'y = +√(x³ + 7)'
	y2 := pAdd(pCub(x), math.SEVEN)
	y, err := math.SqrtModP(y2, c.P)
	return y, err == nil
}

// SignY of a point (y-coordinate) on the Bitcoin curve:
// +1 if above x-axis, -1 if below.
func SignY(p *Point) (int, error) {
	y, ok := Solve(p.X())
	if ok {
		// positive solution?
		if y.Equals(p.Y()) {
			return 1, nil
		}
		// negative solution?
		if y.Equals(c.P.Sub(p.Y())) {
			return -1, nil
		}
	}
	return 0, errors.New("Point not on curve")
}

// Inf is the point at "infinity"
var Inf = NewPoint(math.ZERO, math.ZERO)

// Point (x,y) on the curve
type Point struct { // exported Point type
	x, y *math.Int // coordinate values
}

// NewPoint instaniates a new Point
func NewPoint(a, b *math.Int) *Point {
	return &Point{x: a, y: b}
}

// NewPointFromBytes reconstructs a Point from binary representation.
func NewPointFromBytes(b []byte) (p *Point, compr bool, err error) {
	p = NewPoint(math.ZERO, math.ZERO)
	err = nil
	compr = true
	switch b[0] {
	case 0:
	case 4:
		p.x = math.NewIntFromBytes(b[1:33])
		p.y = math.NewIntFromBytes(b[33:])
		compr = false
	case 3:
		p.x = math.NewIntFromBytes(b[1:])
		p.y, err = computeY(p.x, 1)
		if err != nil {
			return
		}
	case 2:
		p.x = math.NewIntFromBytes(b[1:])
		p.y, err = computeY(p.x, 0)
		if err != nil {
			return
		}
	default:
		err = errors.New("Invalid binary Point representation")
	}
	return
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
	return NewPoint(p.x, p.y.Neg())
}

// String returns a human-readable representation of a point.
func (p *Point) String() string {
	return fmt.Sprintf("(%v,%v)", p.x, p.y)
}

// Equals checks if two Points are equal
func (p *Point) Equals(q *Point) bool {
	return p.x.Cmp(q.x) == 0 && p.y.Cmp(q.y) == 0
}

// IsInf checks if a Point is at infinity
func (p *Point) IsInf() bool {
	return p.x.Cmp(math.ZERO) == 0 && p.y.Cmp(math.ZERO) == 0
}

// IsOnCurve checks if a Point (x,y) is on the curve
func (p *Point) IsOnCurve() bool {
	// y² = x³ + 7
	y2 := pSqr(p.y)
	x3 := pCub(p.x)
	return y2.Cmp(pAdd(x3, c.B)) == 0
}

// Add two Points on the curve
func (p *Point) Add(q *Point) *Point {
	if p.Equals(q) {
		return p.Double()
	}
	if p.Equals(Inf) {
		return q
	}
	if q.Equals(Inf) {
		return p
	}
	_p1 := newJacPoint(p.x, p.y, math.ONE)
	_p2 := newJacPoint(q.x, q.y, math.ONE)
	return _p1.add(_p2).conv()
}

// Double a Point on the curve
func (p *Point) Double() *Point {
	if p.Equals(Inf) {
		return Inf
	}
	return newJacPoint(p.x, p.y, math.ONE).double().conv()
}

// Mult multiplies a Point on the curve with a scalar value k using
// a Montgomery multiplication approach
func (p *Point) Mult(k *math.Int) *Point {
	return newJacPoint(p.x, p.y, math.ONE).mult(k).conv()
}

// Bytes returns a byte representation of Point (compressed or uncompressed).
func (p *Point) Bytes(compressed bool) []byte {
	if p.Equals(Inf) {
		return []byte{0}
	}
	var res []byte
	if compressed {
		rc := byte(2)
		if p.y.Bit(0) == 1 {
			rc = 3
		}
		res = append(res, rc)
		res = append(res, coordAsBytes(p.x)...)
	} else {
		res = append(res, 4)
		res = append(res, coordAsBytes(p.x)...)
		res = append(res, coordAsBytes(p.y)...)
	}
	return res
}

// JacPoint is a point on the curve that is represented internally in
// Jacobian coordinates (X,Y,Z) with "x = X/Z^2" and "y = Y/Z^3". See:
// [http://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html]
type jacPoint struct { // internal Point type
	x, y, z *math.Int // using Jacobian coordinates
}

// NewJacPoint instaniates a new Point
func newJacPoint(a, b, c *math.Int) *jacPoint {
	return &jacPoint{x: a, y: b, z: c}
}

// Point at infinity
var jacInf = newJacPoint(Inf.x, Inf.y, math.ONE)

// check if a Point is at infinity
func (p *jacPoint) isInf() bool {
	return p.x.Equals(math.ZERO) && p.y.Equals(math.ZERO)
}

// convert internal Point to external representation
func (p *jacPoint) conv() *Point {
	if p.z.Equals(math.ZERO) {
		return NewPoint(math.ZERO, math.ZERO)
	}
	zi := pInv(p.z)
	x := pMul(p.x, pSqr(zi))
	y := pMul(p.y, pCub(zi))
	return NewPoint(x, y)
}

// add two jacPoints on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/addJacition/addJac-2007-bl]
func (p *jacPoint) add(q *jacPoint) *jacPoint {
	if p.isInf() {
		return q
	}
	if q.isInf() {
		return p
	}
	z1z1 := pSqr(p.z)
	z2z2 := pSqr(q.z)
	u1 := pMul(p.x, z2z2)
	u2 := pMul(q.x, z1z1)
	s1 := pMul(pMul(p.y, q.z), z2z2)
	s2 := pMul(pMul(q.y, p.z), z1z1)
	h := pSub(u2, u1)
	i := pSqr(pMul(math.TWO, h))
	j := pMul(h, i)
	r := pMul(math.TWO, pSub(s2, s1))
	v := pMul(u1, i)
	w := pAdd(p.z, q.z)
	x := pSub(pSub(pSqr(r), j), pMul(math.TWO, v))
	y := pSub(pMul(r, pSub(v, x)), pMul(math.TWO, pMul(s1, j)))
	z := pMul(pSub(pSub(pSqr(w), z1z1), z2z2), h)
	return newJacPoint(x, y, z)
}

// double a Point on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/doubling/dbl-2009-alnr]
func (p *jacPoint) double() *jacPoint {
	if p.isInf() {
		return p
	}
	a := pSqr(p.x)
	b := pSqr(p.y)
	zz := pSqr(p.z)
	c := pSqr(b)
	d := pMul(math.TWO, pSub(pSub(pSqr(p.x.Add(b)), a), c))
	e := pMul(math.THREE, a)
	f := pSqr(e)
	x := pSub(f, pMul(math.TWO, d))
	y := pSub(pMul(e, pSub(d, x)), pMul(math.EIGHT, c))
	z := pSub(pSub(pSqr(pAdd(p.y, p.z)), b), zz)
	return newJacPoint(x, y, z)
}

// Multiply a Point on the curve with a scalar value k using
// a Montgomery multiplication algorithm
func (p *jacPoint) mult(k *math.Int) *jacPoint {
	if p.isInf() {
		return p
	}
	if k.Cmp(math.ZERO) == 0 {
		return jacInf
	}
	r := jacInf
	x := jacInf
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

// helper: convert coordinate to byte array of correct length
func coordAsBytes(v *math.Int) []byte {
	bv := v.Bytes()
	plen := 32 - len(bv)
	if plen == 0 {
		return bv
	}
	b := make([]byte, plen)
	return append(b, bv...)
}

// helper: reconstruct y-coordinate of Point
func computeY(x *math.Int, m uint) (y *math.Int, err error) {
	y = math.ZERO
	err = nil
	y2 := pAdd(pCub(x), c.B)
	y, err = math.SqrtModP(y2, c.P)
	if err == nil {
		if y.Bit(0) != m {
			y = c.P.Sub(y)
		}
	}
	return
}

// modulus
func nMod(a *math.Int) *math.Int {
	return a.Mod(c.N)
}

func pInv(a *math.Int) *math.Int {
	return a.ModInverse(c.P)
}

func pMul(a, b *math.Int) *math.Int {
	return a.Mul(b).Mod(c.P)
}

func pSqr(a *math.Int) *math.Int {
	return pMul(a, a)
}

func pCub(a *math.Int) *math.Int {
	return pMul(pSqr(a), a)
}

func pSub(a, b *math.Int) *math.Int {
	x := a.Sub(b)
	if x.Sign() == -1 {
		x = x.Add(c.P)
	}
	return x
}

func pAdd(a, b *math.Int) *math.Int {
	return a.Add(b).Mod(c.P)
}

func nInv(a *math.Int) *math.Int {
	return a.ModInverse(c.N)
}

func nMul(a, b *math.Int) *math.Int {
	return a.Mul(b).Mod(c.N)
}

func nRnd(a *math.Int) *math.Int {
	return math.NewIntRndRange(a, c.N)
}
