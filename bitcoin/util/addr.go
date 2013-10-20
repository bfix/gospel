/*
 * Bitcoin address: base58 encoded binary data from hashes and version
 *
 * (c) 2011-2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package util

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"code.google.com/p/go.crypto/ripemd160"
	"crypto/sha256"
	"github.com/bfix/gospel/bitcoin/ecc"
)

///////////////////////////////////////////////////////////////////////
// Address type (string-like base58 encoded data)

type Address string

///////////////////////////////////////////////////////////////////////
// compute address from public key for either the "real" bitcoin
// network or the test network

func MakeAddress(key *ecc.PublicKey) Address {
	return buildAddr(key, 0)
}

func MakeTestAddress(key *ecc.PublicKey) Address {
	return buildAddr(key, 111)
}

///////////////////////////////////////////////////////////////////////
// helper: compute address from public key using different (nested)
// hashes and identifiers.

func buildAddr(key *ecc.PublicKey, version byte) Address {

	addr := make([]byte, 0)
	addr = append(addr, version)

	sha2 := sha256.New()
	sha2.Write(key.Bytes(true))
	h := sha2.Sum(nil)

	ripemd := ripemd160.New()
	ripemd.Write(h)
	kh := ripemd.Sum(nil)
	addr = append(addr, kh...)

	sha2.Reset()
	sha2.Write(addr)
	h = sha2.Sum(nil)
	sha2.Reset()
	sha2.Write(h)
	cs := sha2.Sum(nil)
	addr = append(addr, cs[:4]...)

	return Address(Base58Encode(addr))
}