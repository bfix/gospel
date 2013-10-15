/*
 * Constants for elliptic curve 'Secp256k1'.
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
	"math/big"
	"github.com/bfix/gospel/math"
)

///////////////////////////////////////////////////////////////////////
// Curve constants.

// order of underlying field "F_p"
// p = 2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1
var curve_p = fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F")

// curve parameter (=7)
var curve_b = math.SEVEN

// base point
var curve_gx = fromHex("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798")
var curve_gy = fromHex("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8")

// order of G
var curve_n = fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141")

// cofactor
var curve_h = math.ONE

// bitsize
var curve_bits = 256

///////////////////////////////////////////////////////////////////////
// helper for initialization of bignum from hex string

func fromHex(s string) *big.Int {
	val, _ := new(big.Int).SetString(s, 16)
	return val
}
