package ed25519

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"crypto/rand"
	"crypto/sha512"

	"github.com/bfix/gospel/math"
)

//----------------------------------------------------------------------
// Public key
//----------------------------------------------------------------------

// PublicKey is a point on the Ed25519 curve.
type PublicKey struct {
	Q *Point
}

// NewPublicKeyFromBytes creates a new public key from a binary representation.
func NewPublicKeyFromBytes(data []byte) *PublicKey {
	q, err := NewPointFromBytes(data)
	if err != nil {
		return nil
	}
	return &PublicKey{
		Q: q,
	}
}

// Bytes returns the binary representation of a public key.
func (pub *PublicKey) Bytes() []byte {
	return pub.Q.Bytes()
}

// Mult returns P = n*Q
func (pub *PublicKey) Mult(n *math.Int) *PublicKey {
	return &PublicKey{
		Q: pub.Q.Mult(n),
	}
}

//----------------------------------------------------------------------
// Private Key
//----------------------------------------------------------------------

// PrivateKey is a Ed25519 private key.
type PrivateKey struct {
	PublicKey
	Seed []byte    // 32-byte seed data
	D    *math.Int // private scalar
}

// NewPrivateKeyFromSeed returns a private key for a given seed.
// If the seed has no matching length, no key is returned
func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	// check seed size
	if len(seed) != 32 {
		return nil
	}
	// create private key and save seed
	key := &PrivateKey{}
	key.Seed = make([]byte, 32)
	copy(key.Seed, seed)
	// compute hash from seed to derive 'd'
	md := sha512.Sum512(seed)
	d := reverse(md[:32])
	d[0] = (d[0] & 0x3f) | 0x40
	d[31] &= 0xf8
	// set private scalar and public point
	key.D = math.NewIntFromBytes(d)
	key.Q = c.MultBase(key.D)
	// return new private key instance
	return key
}

// NewPrivateKeyFromD returns a private key for a given factor.
func NewPrivateKeyFromD(d *math.Int) *PrivateKey {
	k := &PrivateKey{}
	k.D = d
	k.Q = c.MultBase(d)
	k.Seed = nil
	return k
}

// Bytes returns the binary representation of a private key.
func (prv *PrivateKey) Bytes() []byte {
	buf := make([]byte, 64)
	if prv.Seed != nil {
		copy(buf[:32], prv.Seed)
	}
	copy(buf[32:], prv.Q.Bytes())
	return buf
}

// Public returns the public key for a private key.
func (prv *PrivateKey) Public() *PublicKey {
	return &PublicKey{
		Q: prv.Q,
	}
}

// Mult returns a new private key with d' = n*d
func (prv *PrivateKey) Mult(n *math.Int) *PrivateKey {
	return NewPrivateKeyFromD(prv.D.Mul(n))
}

// NewKeypair creates a new Ed25519 key pair.
func NewKeypair() (*PublicKey, *PrivateKey) {
	seed := make([]byte, 32)
	rand.Read(seed)
	prv := NewPrivateKeyFromSeed(seed)
	pub := prv.Public()
	return pub, prv
}
