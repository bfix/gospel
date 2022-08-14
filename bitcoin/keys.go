package bitcoin

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

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
	return k.Q.Bytes(k.IsCompressed)
}

// PublicKeyFromBytes returns a public key from a byte representation.
func PublicKeyFromBytes(b []byte) (*PublicKey, error) {
	pnt, compr, err := NewPointFromBytes(b)
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
			return nil, errors.New("invalid private key format (compression flag)")
		}
	} else if len(b) != 32 {
		return nil, errors.New("invalid private key format (length)")
	}
	// set private factor.
	key := &PrivateKey{}
	key.D = math.NewIntFromBytes(kd)
	// compute public key
	g := GetBasePoint()
	key.Q = g.Mult(key.D)
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
		prv.Q = MultBase(prv.D)
		prv.IsCompressed = compr

		// check for valid key
		if !prv.PublicKey.Q.IsInf() {
			break
		}
	}
	return prv
}
