/*
 * Elliptic curve 'Secp256k1' methods. 
 *
 * (c) 2011-2012 Bernd Fix   >Y<
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
	"gospel/crypto"
	"math/big"
)

///////////////////////////////////////////////////////////////////////
// Point (x,y) on the curve

type point struct { // exported point type
	x, y *big.Int // coordinate values
}

// point at infinity
var inf = &point{zero, zero}

///////////////////////////////////////////////////////////////////////
// check if two points are equal

func isEqual(p1, p2 *point) bool {
	return p1.x.Cmp(p2.x) == 0 && p1.y.Cmp(p2.y) == 0
}

///////////////////////////////////////////////////////////////////////
// check if a point is at infinity

func isInf(p *point) bool {
	return p.x.Cmp(zero) == 0 && p.y.Cmp(zero) == 0
}

///////////////////////////////////////////////////////////////////////
// check if a point (x,y) is on the curve

func isOnCurve(p *point) bool {
	// y² = x³ + 7
	y2 := p_sqr(p.y)
	x3 := p_cub(p.x)
	return y2.Cmp(p_add(x3, curve_b)) == 0
}

///////////////////////////////////////////////////////////////////////
// Add two points on the curve

func add(p1, p2 *point) *point {
	if isEqual(p1, p2) {
		return double(p1)
	}
	if isEqual(p1, inf) {
		return p2
	}
	if isEqual(p2, inf) {
		return p1
	}
	_p1 := &point_{p1.x, p1.y, one}
	_p2 := &point_{p2.x, p2.y, one}
	return conv(add_(_p1, _p2))
}

///////////////////////////////////////////////////////////////////////
// Double a point on the curve

func double(p *point) *point {
	if isEqual(p, inf) {
		return inf
	}
	return conv(double_(&point_{p.x, p.y, one}))
}

///////////////////////////////////////////////////////////////////////
// Multiply a point on the curve with a scalar value k using
// a Montgomery multiplication approach

func scalarMult(p *point, k *big.Int) *point {
	return conv(scalarMult_(&point_{p.x, p.y, one}, k))
}

///////////////////////////////////////////////////////////////////////
// Multiply the base point of the curve with a scalar value k

func scalarMultBase(k *big.Int) *point {
	return scalarMult(&point{curve_gx, curve_gy}, k)
}

///////////////////////////////////////////////////////////////////////
// points (x,y) on the curve are represented internally in Jacobian
// coordinates (X,Y,Z) with "x = X/Z^2" and "y = Y/Z^3". See:
// [http://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html]

type point_ struct { // internal point type
	x, y, z *big.Int // using Jacobian coordinates
}

// point at infinity
var inf_ = &point_{inf.x, inf.y, one}

///////////////////////////////////////////////////////////////////////
// check if a point is at infinity

func isInf_(p *point_) bool {
	return p.x.Cmp(zero) == 0 && p.y.Cmp(zero) == 0
}

///////////////////////////////////////////////////////////////////////
// convert internal point to external representation

func conv(p *point_) *point {
	zi := p_inv(p.z)
	x := p_mul(p.x, p_sqr(zi))
	y := p_mul(p.y, p_cub(zi))
	return &point{x, y}
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
	i := p_sqr(p_mul(two, h))
	j := p_mul(h, i)
	r := p_mul(two, p_sub(s2, s1))
	v := p_mul(u1, i)
	w := p_add(p1.z, p2.z)
	x := p_sub(p_sub(p_sqr(r), j), p_mul(two, v))
	y := p_sub(p_mul(r, p_sub(v, x)), p_mul(two, p_mul(s1, j)))
	z := p_mul(p_sub(p_sub(p_sqr(w), z1z1), z2z2), h)
	return &point_{x, y, z}
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
	d := p_mul(two, p_sub(p_sub(p_sqr(_add(p.x, b)), a), c))
	e := p_mul(three, a)
	f := p_sqr(e)
	x := p_sub(f, p_mul(two, d))
	y := p_sub(p_mul(e, p_sub(d, x)), p_mul(eight, c))
	z := p_sub(p_sub(p_sqr(p_add(p.y, p.z)), b), zz)
	return &point_{x, y, z}
}

///////////////////////////////////////////////////////////////////////
// Multiply a point on the curve with a scalar value k using
// a Montgomery multiplication algorithm

func scalarMult_(p *point_, k *big.Int) *point_ {

	if isInf_(p) {
		return p
	}
	if k.Cmp(zero) == 0 {
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
