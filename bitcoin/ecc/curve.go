package ecc

import (
	"errors"
	"github.com/bfix/gospel/math"
)

// Point (x,y) on the curve
type Point struct { // exported Point type
	x, y *math.Int // coordinate values
}

// NewPoint instaniates a new Point
func NewPoint(a, b *math.Int) *Point {
	return &Point{x: a, y: b}
}

// Point at infinity
var inf = NewPoint(math.ZERO, math.ZERO)

// GetBasePoint returns the base Point of the curve
func GetBasePoint() *Point {
	return NewPoint(curveGx, curveGy)
}

// get byte representation of Point (compressed or uncompressed).
func pointAsBytes(p *Point, compressed bool) []byte {
	if IsEqual(p, inf) {
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

// reconstruct Point from binary representation
func pointFromBytes(b []byte) (p *Point, compr bool, err error) {
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

// helper: reconstruct y-coordinate of Point
func computeY(x *math.Int, m uint) (y *math.Int, err error) {
	y = math.ZERO
	err = nil
	y2 := pAddJac(pCub(x), curveB)
	y, err = math.SqrtModP(y2, curveP)
	if err == nil {
		if y.Bit(0) != m {
			y = curveP.Sub(y)
		}
	}
	return
}

// IsEqual checks if two Points are equal
func IsEqual(p1, p2 *Point) bool {
	return p1.x.Cmp(p2.x) == 0 && p1.y.Cmp(p2.y) == 0
}

// isInf checks if a Point is at infinity
func isInf(p *Point) bool {
	return p.x.Cmp(math.ZERO) == 0 && p.y.Cmp(math.ZERO) == 0
}

// IsOnCurve checks if a Point (x,y) is on the curve
func IsOnCurve(p *Point) bool {
	// y² = x³ + 7
	y2 := pSqr(p.y)
	x3 := pCub(p.x)
	return y2.Cmp(pAddJac(x3, curveB)) == 0
}

// Add two Points on the curve
func add(p1, p2 *Point) *Point {
	if IsEqual(p1, p2) {
		return double(p1)
	}
	if IsEqual(p1, inf) {
		return p2
	}
	if IsEqual(p2, inf) {
		return p1
	}
	_p1 := newJacPoint(p1.x, p1.y, math.ONE)
	_p2 := newJacPoint(p2.x, p2.y, math.ONE)
	return conv(addJac(_p1, _p2))
}

// Double a Point on the curve
func double(p *Point) *Point {
	if IsEqual(p, inf) {
		return inf
	}
	return conv(doubleJac(newJacPoint(p.x, p.y, math.ONE)))
}

// Multiply a Point on the curve with a scalar value k using
// a Montgomery multiplication approach
func scalarMult(p *Point, k *math.Int) *Point {
	return conv(scalarMultJac(newJacPoint(p.x, p.y, math.ONE), k))
}

// ScalarMultBase multiplies the base Point of the curve with a scalar value k
func ScalarMultBase(k *math.Int) *Point {
	return scalarMult(GetBasePoint(), k)
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
var jacInf = newJacPoint(inf.x, inf.y, math.ONE)

// check if a Point is at infinity
func isInfJac(p *jacPoint) bool {
	return p.x.Equals(math.ZERO) && p.y.Equals(math.ZERO)
}

// convert internal Point to external representation
func conv(p *jacPoint) *Point {
	zi := pInv(p.z)
	x := pMul(p.x, pSqr(zi))
	y := pMul(p.y, pCub(zi))
	return NewPoint(x, y)
}

// addJac two Points on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/addJacition/addJac-2007-bl]
func addJac(p1, p2 *jacPoint) *jacPoint {
	if isInfJac(p1) {
		return p2
	}
	if isInfJac(p2) {
		return p1
	}
	z1z1 := pSqr(p1.z)
	z2z2 := pSqr(p2.z)
	u1 := pMul(p1.x, z2z2)
	u2 := pMul(p2.x, z1z1)
	s1 := pMul(pMul(p1.y, p2.z), z2z2)
	s2 := pMul(pMul(p2.y, p1.z), z1z1)
	h := pSub(u2, u1)
	i := pSqr(pMul(math.TWO, h))
	j := pMul(h, i)
	r := pMul(math.TWO, pSub(s2, s1))
	v := pMul(u1, i)
	w := pAddJac(p1.z, p2.z)
	x := pSub(pSub(pSqr(r), j), pMul(math.TWO, v))
	y := pSub(pMul(r, pSub(v, x)), pMul(math.TWO, pMul(s1, j)))
	z := pMul(pSub(pSub(pSqr(w), z1z1), z2z2), h)
	return newJacPoint(x, y, z)
}

// double a Point on the curve
// [http://www.hyperelliptic.org/EFD/g1p/data/shortw/jacobian-0/doubling/dbl-2009-alnr]
func doubleJac(p *jacPoint) *jacPoint {
	if isInfJac(p) {
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
	z := pSub(pSub(pSqr(pAddJac(p.y, p.z)), b), zz)
	return newJacPoint(x, y, z)
}

// Multiply a Point on the curve with a scalar value k using
// a Montgomery multiplication algorithm
func scalarMultJac(p *jacPoint, k *math.Int) *jacPoint {
	if isInfJac(p) {
		return p
	}
	if k.Cmp(math.ZERO) == 0 {
		return jacInf
	}
	r := jacInf
	for _, val := range k.Bytes() {
		for pos := 0; pos < 8; pos++ {
			r = doubleJac(r)
			if val&0x80 == 0x80 {
				r = addJac(p, r)
			}
			val <<= 1
		}
	}
	return r
}

// modulus
func nMod(a *math.Int) *math.Int {
	return a.Mod(curveN)
}

// modular inverse
func pInv(a *math.Int) *math.Int {
	return a.ModInverse(curveP)
}

func nInv(a *math.Int) *math.Int {
	return a.ModInverse(curveN)
}

// multiplication
func pMul(a, b *math.Int) *math.Int {
	return a.Mul(b).Mod(curveP)
}

func nMul(a, b *math.Int) *math.Int {
	return a.Mul(b).Mod(curveN)
}

// squares and cubes
func pSqr(a *math.Int) *math.Int {
	return pMul(a, a)
}

func pCub(a *math.Int) *math.Int {
	return pMul(pSqr(a), a)
}

//	addJacition and subtraction
func pSub(a, b *math.Int) *math.Int {
	x := a.Sub(b)
	if x.Sign() == -1 {
		x = x.Add(curveP)
	}
	return x
}

func xaddIntJac(a, b *math.Int) *math.Int {
	return a.Add(b)
}

func pAddJac(a, b *math.Int) *math.Int {
	return a.Add(b).Mod(curveP)
}

// generate random integer value in given range
func nRnd(a *math.Int) *math.Int {
	return math.NewIntRndRange(a, curveN)
}
