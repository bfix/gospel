package crypto

import (
	"math/big"
	"math/rand"
)

var rnd = rand.New(NewPrngSource())

// RandInt returns a random integer value with given range (inclusive).
func RandInt(lower, upper int) int {
	return lower + (rnd.Int() % (upper - lower + 1))
}

// RandBytes generates a byte array of given size with random content.
func RandBytes(n int) []byte {
	data := make([]byte, n)
	for i := 0; i < n; i++ {
		data[i] = byte(rnd.Int() & 0xFF)
	}
	return data
}

// RandBigInt returns a random big integer value with given range.
func RandBigInt(lower, upper *big.Int) *big.Int {
	span := new(big.Int).Sub(upper, lower)
	span = new(big.Int).Add(span, big.NewInt(1))
	ofs := new(big.Int).Rand(rnd, span)
	return new(big.Int).Add(lower, ofs)
}
