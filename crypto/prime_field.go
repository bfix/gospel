package crypto

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
