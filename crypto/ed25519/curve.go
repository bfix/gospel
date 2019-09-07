package ed25519

import (
	"fmt"

	"github.com/bfix/gospel/math"
)

// Curve is the Ed25519 elliptic curve (twisted Edwards curve):
//     a x^2 + y^2 = 1 + d x^2 y^2,  d = -121665/121666, a = -1
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

// Solve the curve equation for given y-coordinate (returns positive x)
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
	if p.IsInf() {
		return q
	}
	if q.IsInf() {
		return p
	}

	// r.x = (p.x*q.y + q.x*p.y) / (1 + d*p.x*q.x*p.y*q.y)
	// r.y = (p.y*q.y + p.x*q.x) / (1 - d*p.x*q.x*p.y*q.y)
	k1 := p.x.Mul(q.y).Mod(c.P)
	k2 := p.y.Mul(q.x).Mod(c.P)
	dxy := c.D.Mul(k1).Mul(k2).Mod(c.P)
	nom := k1.Add(k2).Mod(c.P)
	den := math.ONE.Add(dxy).Mod(c.P)
	rx := nom.Mul(den.ModInverse(c.P)).Mod(c.P)
	k3 := p.x.Mul(q.x).Mod(c.P)
	k4 := p.y.Mul(q.y).Mod(c.P)
	nom = k3.Add(k4).Mod(c.P)
	den = math.ONE.Sub(dxy).Mod(c.P)
	ry := nom.Mul(den.ModInverse(c.P)).Mod(c.P)
	return NewPoint(rx, ry)
}

// Double a Point on the curve
func (p *Point) Double() *Point {
	return p.Add(p)
}

// Mult multiplies a Point on the curve with a scalar value k using
// a Montgomery multiplication approach
func (p *Point) Mult(k *math.Int) *Point {
	if p.IsInf() {
		return p
	}
	r := NewPoint(c.Ox, c.Oy)
	if k.Cmp(math.ZERO) == 0 {
		return r
	}
	for _, val := range k.Bytes() {
		for pos := 0; pos < 8; pos++ {
			r = r.Double()
			if val&0x80 == 0x80 {
				r = p.Add(r)
			}
			val <<= 1
		}
	}
	return r
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
