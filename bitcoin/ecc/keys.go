package ecc

import (
	"errors"
	"github.com/bfix/gospel/math"
)

// PublicKey is a Point on the elliptic curve: (x,y) = d*G, where
// 'G' is the base Point of the curve and 'd' is the secret private
// factor (private key)
type PublicKey struct {
	Q            *Point
	IsCompressed bool
}

// Bytes returns the byte representation of public key.
func (k *PublicKey) Bytes() []byte {
	return pointAsBytes(k.Q, k.IsCompressed)
}

// PublicKeyFromBytes returns a public key from a byte representation.
func PublicKeyFromBytes(b []byte) (*PublicKey, error) {
	pnt, compr, err := pointFromBytes(b)
	if err != nil {
		return nil, err
	}
	key := &PublicKey{
		Q:            pnt,
		IsCompressed: compr,
	}
	return key, nil
}

// PrivateKey is a random factor 'd' for the base Point that yields
// the associated PublicKey (Point on the curve (x,y) = d*G)
type PrivateKey struct {
	PublicKey
	D *math.Int
}

// Bytes returns a byte representation of private key.
func (k *PrivateKey) Bytes() []byte {
	b := coordAsBytes(k.D)
	if k.IsCompressed {
		b = append(b, 1)
	}
	return b
}

// PrivateKeyFromBytes returns a private key from a byte representation.
func PrivateKeyFromBytes(b []byte) (*PrivateKey, error) {
	// check compressed/uncompressed
	var (
		kd    = b
		compr = false
	)
	if len(b) == 33 {
		kd = b[:32]
		if b[32] == 1 {
			compr = true
		} else {
			return nil, errors.New("Invalid private key format (compression flag)")
		}
	} else if len(b) != 32 {
		return nil, errors.New("Invalid private key format (length)")
	}
	// set private factor.
	key := &PrivateKey{}
	key.D = math.NewIntFromBytes(kd)
	// compute public key
	g := GetBasePoint()
	key.Q = scalarMult(g, key.D)
	key.IsCompressed = compr
	return key, nil
}

// GenerateKeys creates a new set of keys.
// [http://www.nsa.gov/ia/_files/ecdsa.pdf] page 19f but with a
// different range (value 1 and 2 for exponent are not allowed)
func GenerateKeys(compr bool) *PrivateKey {

	prv := new(PrivateKey)
	for {
		// generate factor in range [3..n-1]
		prv.D = nRnd(math.THREE)
		// generate Point p = d*G
		prv.Q = ScalarMultBase(prv.D)
		prv.IsCompressed = compr

		// check for valid key
		if !isInf(prv.PublicKey.Q) {
			break
		}
	}
	return prv
}
