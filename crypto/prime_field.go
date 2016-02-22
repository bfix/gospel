/*
 * Prime field implementation.
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

// FieldP is a prime field
type FieldP struct {
	P *big.Int
}

///////////////////////////////////////////////////////////////////////
// Prime field methods
///////////////////////////////////////////////////////////////////////

// Random generates a random field value
func (f *FieldP) Random() *big.Int {
	return RandBigInt(big.NewInt(0), new(big.Int).Sub(f.P, big.NewInt(1)))
}

//---------------------------------------------------------------------

// Add field values
func (f *FieldP) Add(a, b *big.Int) *big.Int {
	c := new(big.Int).Add(a, b)
	return new(big.Int).Mod(c, f.P)
}

//---------------------------------------------------------------------

// Sub substracts field values
func (f *FieldP) Sub(a, b *big.Int) *big.Int {
	c := new(big.Int).Add(f.P, a)
	d := new(big.Int).Sub(c, b)
	return new(big.Int).Mod(d, f.P)
}

//---------------------------------------------------------------------

// Neg negates a field value
func (f *FieldP) Neg(a *big.Int) *big.Int {
	return new(big.Int).Sub(f.P, a)
}

//---------------------------------------------------------------------

// Mul multiplies field values
func (f *FieldP) Mul(a, b *big.Int) *big.Int {
	c := new(big.Int).Mul(a, b)
	return new(big.Int).Mod(c, f.P)
}

//---------------------------------------------------------------------

// Div divides field values
func (f *FieldP) Div(a, b *big.Int) *big.Int {
	c := new(big.Int).ModInverse(b, f.P)
	return f.Mul(a, c)
}
