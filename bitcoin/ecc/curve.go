/*
 * Elliptic curve 'Secp256k1' methods.
 *
 * (c) 2011-2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ecc

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"errors"
	"github.com/bfix/gospel/crypto"
	"github.com/bfix/gospel/math"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// Point (x,y) on the curve

type point struct { // exported point type
	x, y *big.Int // coordinate values
}

// instaniate a new point
func NewPoint(a, b *big.Int) *point {
	p := &point{}
	p.x = new(big.Int).Set(a)
	p.y = new(big.Int).Set(b)
	return p
}

// point at infinity
var inf = NewPoint(math.ZERO, math.ZERO)

// get base point
func GetBasePoint() *point {
	return NewPoint(curve_gx, curve_gy)
}

/////////////////////////////////////////////////////////////////////
// get byte representation of point (compressed or uncompressed).

func pointAsBytes(p *point, compressed bool) []byte {
	if IsEqual(p, inf) {
		return []byte{0}
	}
	res := make([]byte, 0)
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

// helper: convert coordinate to byte array of correct length
func coordAsBytes(v *big.Int) []byte {
	bv := v.Bytes()
	plen := 32 - len(bv)
	if plen == 0 {
		return bv
	}
	b := make([]byte, plen)
	return append(b, bv...)
}

/////////////////////////////////////////////////////////////////////
// reconstruct point from binary representation

func pointFromBytes(b []byte) (p *point, compr bool, err error) {
	p = NewPoint(math.ZERO, math.ZERO)
	err = nil
	compr = true
	switch b[0] {
	case 0:
	case 4:
		p.x.SetBytes(b[1:33])
		p.y.SetBytes(b[33:])
		compr = false
	case 3:
		p.x.SetBytes(b[1:])
		p.y, err = computeY(p.x, 1)
		if err != nil {
			return
		}
	case 2:
		p.x.SetBytes(b[1:])
		p.y, err = computeY(p.x, 0)
		if err != nil {
			return
		}
	default:
		err = errors.New("Invalid binary point representation")
	}
	return
}

// helper: reconstruct y-coordinate of point
func computeY(x *big.Int, m uint) (y *big.Int, err error) {
	y = big.NewInt(0)
	err = nil
	y2 := p_add(p_cub(x), curve_b)
	y, err = math.Sqrt_modP(y2, curve_p)
	if err == nil {
		if y.Bit(0) != m {
			y = new(big.Int).Sub(curve_p, y)
		}
	}
	return
}

/////////////////////////////////////////////////////////////////////
// check if two points are equal

func IsEqual(p1, p2 *point) bool {
	return p1.x.Cmp(p2.x) == 0 && p1.y.Cmp(p2.y) == 0
}

///////////////////////////////////////////////////////////////////////
// check if a point is at infinity

func isInf(p *point) bool {
	return p.x.Cmp(math.ZERO) == 0 && p.y.Cmp(math.ZERO) == 0
}

///////////////////////////////////////////////////////////////////////
// check if a point (x,y) is on the curve

func IsOnCurve(p *point) bool {
	// y² = x³ + 7
	y2 := p_sqr(p.y)
	x3 := p_cub(p.x)
	return y2.Cmp(p_add(x3, curve_b)) == 0
}

///////////////////////////////////////////////////////////////////////
// Add two points on the curve

func add(p1, p2 *point) *point {
	if IsEqual(p1, p2) {
		return double(p1)
	}
	if IsEqual(p1, inf) {
		return p2
	}
	if IsEqual(p2, inf) {
		return p1
	}
	_p1 := NewPoint_(p1.x, p1.y, math.ONE)
	_p2 := NewPoint_(p2.x, p2.y, math.ONE)
	return conv(add_(_p1, _p2))
}

///////////////////////////////////////////////////////////////////////
// Double a point on the curve

func double(p *point) *point {
	if IsEqual(p, inf) {
		return inf
	}
	return conv(double_(NewPoint_(p.x, p.y, math.ONE)))
}

///////////////////////////////////////////////////////////////////////
// Multiply a point on the curve with a scalar value k using
// a Montgomery multiplication approach

func scalarMult(p *point, k *big.Int) *point {
	return conv(scalarMult_(NewPoint_(p.x, p.y, math.ONE), k))
}

///////////////////////////////////////////////////////////////////////
// Multiply the base point of the curve with a scalar value k

func ScalarMultBase(k *big.Int) *point {
	return scalarMult(GetBasePoint(), k)
}

///////////////////////////////////////////////////////////////////////
// points (x,y) on the curve are represented internally in Jacobian
// coordinates (X,Y,Z) with "x = X/Z^2" and "y = Y/Z^3". See:
// [http://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html]

type point_ struct { // internal point type
	x, y, z *big.Int // using Jacobian coordinates
}

// instaniate a new point
func NewPoint_(a, b, c *big.Int) *point_ {
	p := &point_{}
	p.x = new(big.Int).Set(a)
	p.y = new(big.Int).Set(b)
	p.z = new(big.Int).Set(c)
	return p
}

// point at infinity
var inf_ = NewPoint_(inf.x, inf.y, math.ONE)

///////////////////////////////////////////////////////////////////////
// check if a point is at infinity

func isInf_(p *point_) bool {
	return p.x.Cmp(math.ZERO) == 0 && p.y.Cmp(math.ZERO) == 0
}

///////////////////////////////////////////////////////////////////////
// convert internal point to external representation

func conv(p *point_) *point {
	zi := p_inv(p.z)
	x := p_mul(p.x, p_sqr(zi))
	y := p_mul(p.y, p_cub(zi))
	return NewPoint(x, y)
}

///////////////////////////////////////////////////////////////////////
// add two points on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/addition/add-2007-bl]

func add_(p1, p2 *point_) *point_ {
	if isInf_(p1) {
		return p2
	}
	if isInf_(p2) {
		return p1
	}
	z1z1 := p_sqr(p1.z)
	z2z2 := p_sqr(p2.z)
	u1 := p_mul(p1.x, z2z2)
	u2 := p_mul(p2.x, z1z1)
	s1 := p_mul(p_mul(p1.y, p2.z), z2z2)
	s2 := p_mul(p_mul(p2.y, p1.z), z1z1)
	h := p_sub(u2, u1)
	i := p_sqr(p_mul(math.TWO, h))
	j := p_mul(h, i)
	r := p_mul(math.TWO, p_sub(s2, s1))
	v := p_mul(u1, i)
	w := p_add(p1.z, p2.z)
	x := p_sub(p_sub(p_sqr(r), j), p_mul(math.TWO, v))
	y := p_sub(p_mul(r, p_sub(v, x)), p_mul(math.TWO, p_mul(s1, j)))
	z := p_mul(p_sub(p_sub(p_sqr(w), z1z1), z2z2), h)
	return NewPoint_(x, y, z)
}

///////////////////////////////////////////////////////////////////////
// double a point on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/doubling/dbl-2009-alnr]

func double_(p *point_) *point_ {
	if isInf_(p) {
		return p
	}
	a := p_sqr(p.x)
	b := p_sqr(p.y)
	zz := p_sqr(p.z)
	c := p_sqr(b)
	d := p_mul(math.TWO, p_sub(p_sub(p_sqr(_add(p.x, b)), a), c))
	e := p_mul(math.THREE, a)
	f := p_sqr(e)
	x := p_sub(f, p_mul(math.TWO, d))
	y := p_sub(p_mul(e, p_sub(d, x)), p_mul(math.EIGHT, c))
	z := p_sub(p_sub(p_sqr(p_add(p.y, p.z)), b), zz)
	return NewPoint_(x, y, z)
}

///////////////////////////////////////////////////////////////////////
// Multiply a point on the curve with a scalar value k using
// a Montgomery multiplication algorithm

func scalarMult_(p *point_, k *big.Int) *point_ {

	if isInf_(p) {
		return p
	}
	if k.Cmp(math.ZERO) == 0 {
		return inf_
	}

	r := inf_
	for _, val := range k.Bytes() {
		for pos := 0; pos < 8; pos++ {
			r = double_(r)
			if val&0x80 == 0x80 {
				r = add_(p, r)
			}
			val <<= 1
		}
	}
	return r
}

///////////////////////////////////////////////////////////////////////
// helper methods for arithmetic operations on curve points

//---------------------------------------------------------------------
//	modulus
//---------------------------------------------------------------------

func _mod(a, n *big.Int) *big.Int {
	return new(big.Int).Mod(a, n)
}

func n_mod(a *big.Int) *big.Int {
	return _mod(a, curve_n)
}

//---------------------------------------------------------------------
//	modular inverse
//---------------------------------------------------------------------

func _inv(a, n *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, n)
}

func p_inv(a *big.Int) *big.Int {
	return _inv(a, curve_p)
}

func n_inv(a *big.Int) *big.Int {
	return _inv(a, curve_n)
}

//---------------------------------------------------------------------
//	multiplication
//---------------------------------------------------------------------

func _mul(a, b, n *big.Int) *big.Int {
	return _mod(new(big.Int).Mul(a, b), n)
}

func p_mul(a, b *big.Int) *big.Int {
	return _mul(a, b, curve_p)
}

func n_mul(a, b *big.Int) *big.Int {
	return _mul(a, b, curve_n)
}

//---------------------------------------------------------------------
//	squares and cubes
//---------------------------------------------------------------------

func p_sqr(a *big.Int) *big.Int {
	return p_mul(a, a)
}

func p_cub(a *big.Int) *big.Int {
	return p_mul(p_sqr(a), a)
}

//---------------------------------------------------------------------
//	addition and subtraction
//---------------------------------------------------------------------

func p_sub(a, b *big.Int) *big.Int {
	x := new(big.Int).Sub(a, b)
	if x.Sign() == -1 {
		x.Add(x, curve_p)
	}
	return x
}

func _add(a, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

func p_add(a, b *big.Int) *big.Int {
	return _mod(_add(a, b), curve_p)
}

//---------------------------------------------------------------------
//	generate random integer value in given range
//---------------------------------------------------------------------

func n_rnd(a *big.Int) *big.Int {
	return crypto.RandBigInt(a, curve_n)
}
