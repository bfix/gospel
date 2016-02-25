package crypto

import (
	"crypto/rand"
	"math/big"
)

// Prng is a pseudo random number generator; a source of randomness
type Prng struct {
	mask *big.Int
}

// Int63 returns the next random (unsigned) 64-bit integer value.
func (p *Prng) Int63() int64 {

	val, err := rand.Int(rand.Reader, p.mask)
	if err != nil {
		panic("PRNG failure: " + err.Error())
	}
	return val.Int64()
}

// Seed for a random source: not necessary, because random bits are
// generated on a system level by either a hardware RNG or a
// cryptographically secure PRNG algorithm.
func (p *Prng) Seed(seed int64) {
	// intentionally not implemented
}

// NewPrngSource instantiates a new source for random bits.
func NewPrngSource() *Prng {
	return &Prng{
		mask: new(big.Int).Lsh(big.NewInt(1), 63),
	}
}
