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
	"crypto/sha1" //nolint:gosec // required for BTC
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160" //nolint:staticcheck // required for BTC
)

// Hash160 computes RIPEMD-160(SHA-256(data))
func Hash160(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	ripemd := ripemd160.New()
	ripemd.Write(sha2.Sum(nil))
	return ripemd.Sum(nil)
}

// Hash256 computes SHA-256(SHA-256(data))
func Hash256(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	h := sha2.Sum(nil)
	sha2.Reset()
	sha2.Write(h)
	return sha2.Sum(nil)
}

// Sha256 computes SHA-256(data)
func Sha256(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	return sha2.Sum(nil)
}

// RipeMD160 computes RIPEMD160(data)
func RipeMD160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)
	return ripemd.Sum(nil)
}

// Sha1 computes SHA1(data)
func Sha1(data []byte) []byte {
	sha1 := sha1.New() //nolint:gosec // required
	sha1.Write(data)
	return sha1.Sum(nil)
}
