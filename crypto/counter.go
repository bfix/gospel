package crypto

import (
	"math/big"
)

// Counter is a cryptographic counter (encrypted value)
type Counter struct {
	pubkey *PaillierPublicKey // reference to public Paillier key
	data   *big.Int           // encrypted counter value
}

// NewCounter creates a new Counter instance for given public key.
func NewCounter(k *PaillierPublicKey) (c *Counter, err error) {

	// create a new counter with value "0"
	d, err := k.Encrypt(big.NewInt(0))
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
func (c *Counter) Get() *big.Int {
	return c.data
}

// Increment counter: usually called with step values of "0" (don't
// change counter, but change representation) and "1" (increment by
// one step).
func (c *Counter) Increment(step *big.Int) error {

	d, err := c.pubkey.Encrypt(step)
	if err != nil {
		return err
	}
	c.data.Mul(c.data, d)
	return nil
}
