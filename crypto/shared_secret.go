/*
 * --------------------------------------------------------------------
 * Shamir Secret Sharing Scheme:
 * --------------------------------------------------------------------
 * Split a secret number into 'n' shares where any 'k' combined shares
 * yield the original secret number. The underlying prime field needs a
 * generator that is larger than the secret to be shared.
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

package crypto

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"math/big"
)

///////////////////////////////////////////////////////////////////////

// Share is a data structure for a partial secret.
type Share struct {
	X, Y, P *big.Int
}

///////////////////////////////////////////////////////////////////////
// Public methods

// Split a 'secret' into 'n' shares, where a 'k' shares are sufficient
// to reconstruct 'secret'.
func Split(secret, p *big.Int, n, k int) []Share {

	f := &FieldP{p}
	// coefficients for a k-1 polynominal
	a := make([]*big.Int, k)
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

//---------------------------------------------------------------------

// Reconstruct secrets from number of shares: if not sufficient shares
// are available, the resulting secret is "random"
func Reconstruct(shares []Share) *big.Int {

	// compute value of Lagrangian polynominal at 0
	k := len(shares)
	y := big.NewInt(0)
	f := &FieldP{shares[0].P}
	for i, s := range shares {
		if s.P.Cmp(f.P) != 0 {
			return nil
		}
		li := big.NewInt(1)
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
