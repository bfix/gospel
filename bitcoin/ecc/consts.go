package ecc

import (
	"github.com/bfix/gospel/math"
)

// order of underlying field "F_p"
// p = 2^256 - 2^32 - 2^9 - 2^8 - 2^7 - 2^6 - 2^4 - 1
var curveP = math.NewIntFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F")

// curve parameter (=7)
var curveB = math.SEVEN

// base point
var curveGx = math.NewIntFromHex("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798")
var curveGy = math.NewIntFromHex("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8")

// order of G
var curveN = math.NewIntFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141")

// cofactor
var curveH = math.ONE

// bitsize
var curveBits = 256
