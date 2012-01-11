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

package math

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"big"
	"rand"
	"time"
)

///////////////////////////////////////////////////////////////////////
// Pre-defined field values

var _ZERO = big.NewInt (0)
var _ONE  = big.NewInt (1)
var _TWO  = big.NewInt (2)

///////////////////////////////////////////////////////////////////////
// PRNG

var rnd	= rand.New (rand.NewSource (time.UTC().Nanoseconds()))

///////////////////////////////////////////////////////////////////////
/*
 * Prime field
 */
type FieldP struct {
	P	*big.Int
}

///////////////////////////////////////////////////////////////////////
// Prime field methods
///////////////////////////////////////////////////////////////////////

/*
 * Generate random field value
 * @return *big.Int - random value in field
 */
func (f *FieldP) Random () *big.Int {
	return new(big.Int).Rand (rnd, f.P)
} 

//---------------------------------------------------------------------
/*
 * Add field values
 * @param a,b *big.Int - numbers to be added
 * @return *big.Int - resulting number
 */
func (f *FieldP)  Add (a, b *big.Int) *big.Int {
	c := new(big.Int).Add (a, b)
	return new(big.Int).Mod (c, f.P)
}

//---------------------------------------------------------------------
/*
 * Subtract field values
 * @param a,b *big.Int - numbers to be subtracted
 * @return *big.Int - resulting number
 */
func (f *FieldP) Sub (a, b *big.Int) *big.Int {
	c := new(big.Int).Add (f.P, a)
	d := new(big.Int).Sub (c, b)
	return new(big.Int).Mod (d, f.P)
}

//---------------------------------------------------------------------
/*
 * Negate field value
 * @param a *big.Int - numbers to be negated
 * @return *big.Int - resulting number
 */
func (f *FieldP) Neg (a *big.Int) *big.Int {
	return new(big.Int).Sub (f.P, a)
}

//---------------------------------------------------------------------
/*
 * Multiply field values
 * @param a,b *big.Int - numbers to be multiplied
 * @return *big.Int - resulting number
 */
func (f *FieldP)  Mul (a, b *big.Int) *big.Int {
	c := new(big.Int).Mul (a, b)
	return new(big.Int).Mod (c, f.P)
}

//---------------------------------------------------------------------
/*
 * Divide field values
 * @param a,b *big.Int - numbers to be divided
 * @return *big.Int - resulting number
 */
func (f *FieldP) Div (a, b *big.Int) *big.Int {
	c := new(big.Int).ModInverse (b, f.P)
	return f.Mul (a, c)
} 

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-11 03:02:41  brf
//  First release as free software (GPL3+)
//
//	Revision 1.2  2011-12-21 23:33:19  brf
//	Comments added. Code clean-up.
//
