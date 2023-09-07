package bitcoin

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
	"bytes"
	"errors"
	"fmt"
)

// ExportPrivateKey returns a private key in SIPA format
func ExportPrivateKey(k *PrivateKey, testnet bool) string {
	var exp []byte
	if testnet {
		exp = append(exp, 0xEF)
	} else {
		exp = append(exp, 0x80)
	}
	exp = append(exp, k.Bytes()...)

	cs := Hash256(exp)
	exp = append(exp, cs[:4]...)

	return Base58Encode(exp)
}

// ImportPrivateKey imports a private key in SIPA format
func ImportPrivateKey(keydata string, testnet bool) (*PrivateKey, error) {
	// decode and check data
	data, err := Base58Decode(keydata)
	if err != nil {
		return nil, err
	}
	if testnet {
		if data[0] != 0xEF {
			msg := fmt.Sprintf("Invalid key version: %d (testnet)\n", int(data[0]))
			return nil, errors.New(msg)
		}
	} else {
		if data[0] != 0x80 {
			msg := fmt.Sprintf("Invalid key version: %d\n", int(data[0]))
			return nil, errors.New(msg)
		}
	}
	// copy key data
	var k, c []byte
	if len(data) == 37 {
		// uncompressed public key
		k = data[1:33]
		c = data[33:37]
	} else if len(data) == 38 {
		// compressed public key
		k = data[1:34]
		c = data[34:38]
		if data[33] != 1 {
			return nil, fmt.Errorf("invalid key compression indicator: %d", int(data[33]))
		}
	} else {
		return nil, errors.New("invalid key format")
	}
	// recompute and verify checksum
	var t []byte
	t = append(t, data[0])
	t = append(t, k...)
	cs := Hash256(t)
	if !bytes.Equal(c, cs[:4]) {
		return nil, errors.New("invalid key data")
	}
	// return key
	return PrivateKeyFromBytes(k)
}
