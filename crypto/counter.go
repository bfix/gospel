package crypto

import (
	"github.com/bfix/gospel/math"
)

// Counter is a cryptographic counter (encrypted value)
type Counter struct {
	pubkey *PaillierPublicKey // reference to public Paillier key
	data   *math.Int          // encrypted counter value
}

// NewCounter creates a new Counter instance for given public key.
func NewCounter(k *PaillierPublicKey) (c *Counter, err error) {
	// create a new counter with value "0"
	d, err := k.Encrypt(math.ZERO)
	if err != nil {
		return nil, err
	}
	c = &Counter{
		pubkey: k,
		data:   d,
	}
	return c, nil
}

// Get the encrypted counter value.
func (c *Counter) Get() *math.Int {
	return c.data
}

// Increment counter: usually called with step values of "0" (don't
// change counter, but change representation) and "1" (increment by
// one step).
func (c *Counter) Increment(step *math.Int) error {

	d, err := c.pubkey.Encrypt(step)
	if err != nil {
		return err
	}
	c.data = c.data.Mul(d)
	return nil
}
