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

const (
	// P2PKH is " Pay-to-PubKeyHash" scheme
	P2PKH = 0
	// P2SH is "Pay-to-ScriptHash" scheme
	P2SH = 1
)

var (
	addrVersion = [][2]byte{
		{0, 111}, // P2PKH
		{5, 196}, // P2SH
	}
)

// MakeAddress computes an address from public key for the "real" Bitcoin network
func MakeAddress(key *PublicKey, version int) string {
	return buildAddr(key, addrVersion[version][0])
}

// MakeTestAddress computes an address from public key for the test network
func MakeTestAddress(key *PublicKey, version int) string {
	return buildAddr(key, addrVersion[version][1])
}

// helper: compute address from public key using different (nested)
// hashes and identifiers.
func buildAddr(key *PublicKey, version byte) string {
	var addr []byte
	addr = append(addr, version)
	kh := Hash160(key.Bytes())
	addr = append(addr, kh...)
	cs := Hash256(addr)
	addr = append(addr, cs[:4]...)
	return string(Base58Encode(addr))
}
