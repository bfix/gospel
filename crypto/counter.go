package crypto

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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
