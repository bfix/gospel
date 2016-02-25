package crypto

import (
	"math/big"
)

// FieldP is a prime field
type FieldP struct {
	P *big.Int
}

// Random generates a random field value
func (f *FieldP) Random() *big.Int {
	return RandBigInt(big.NewInt(0), new(big.Int).Sub(f.P, big.NewInt(1)))
}

// Add field values
func (f *FieldP) Add(a, b *big.Int) *big.Int {
	c := new(big.Int).Add(a, b)
	return new(big.Int).Mod(c, f.P)
}

// Sub subtracts field values
func (f *FieldP) Sub(a, b *big.Int) *big.Int {
	c := new(big.Int).Add(f.P, a)
	d := new(big.Int).Sub(c, b)
	return new(big.Int).Mod(d, f.P)
}

// Neg negates a field value
func (f *FieldP) Neg(a *big.Int) *big.Int {
	return new(big.Int).Sub(f.P, a)
}

// Mul multiplies field values
func (f *FieldP) Mul(a, b *big.Int) *big.Int {
	c := new(big.Int).Mul(a, b)
	return new(big.Int).Mod(c, f.P)
}

// Div divides field values
func (f *FieldP) Div(a, b *big.Int) *big.Int {
	c := new(big.Int).ModInverse(b, f.P)
	return f.Mul(a, c)
}
