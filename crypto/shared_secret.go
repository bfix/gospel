package crypto

import (
	"github.com/bfix/gospel/math"
)

// Share is a data structure for a partial secret.
type Share struct {
	X, Y, P *math.Int
}

// Split a 'secret' into 'n' shares, where a 'k' shares are sufficient
// to reconstruct 'secret'.
func Split(secret, p *math.Int, n, k int) []Share {

	f := &FieldP{p}
	// coefficients for a k-1 polynominal
	a := make([]*math.Int, k)
	a[0] = secret
	// generate remaining coefficients
	for i := 1; i < k; i++ {
		a[i] = f.Random()
	}

	// construct shares
	shares := make([]Share, n)
	for i := range shares {
		x := f.Random()
		y := a[0]
		xi := x
		for j := 1; j < k; j++ {
			yi := f.Mul(a[j], xi)
			y = f.Add(y, yi)
			xi = f.Mul(xi, x)
		}
		shares[i] = Share{x, y, f.P}
	}
	return shares
}

// Reconstruct secrets from number of shares: if not sufficient shares
// are available, the resulting secret is "random"
func Reconstruct(shares []Share) *math.Int {

	// compute value of Lagrangian polynominal at 0
	k := len(shares)
	y := math.ZERO
	f := &FieldP{shares[0].P}
	for i, s := range shares {
		if !s.P.Equals(f.P) {
			return nil
		}
		li := math.ONE
		for j := 0; j < k; j++ {
			if j == i {
				continue
			}
			a := f.Neg(shares[j].X)
			b := f.Sub(s.X, shares[j].X)
			li = f.Mul(li, f.Div(a, b))
		}
		y = f.Add(y, f.Mul(s.Y, li))
	}
	return y
}
