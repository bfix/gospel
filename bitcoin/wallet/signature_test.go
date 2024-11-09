//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2024-present, Bernd Fix  >Y<
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

package wallet

import (
	"testing"

	"github.com/bfix/gospel/bitcoin"
)

func TestVerifyBitcoinMsg(t *testing.T) {
	addr := "1CypmgrpbP6ohTWBRYEaCKQAgUz6oTkon9"
	msg := "Secret message #1"
	b64Sig := "H0OFI1thq9kJXYGQ3E2lDc4dlD1o0XDM0mgaf6oKDq/vrAKERrV76P6kAejZzoSL9MOIgUoxqG3MQg1EpaT/0Jg="

	// verify signature
	ok, err := VerifyBitcoinMsg(addr, b64Sig, msg)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature NOT verified")
	}
}

func TestSignVerifyBitcoinMsg(t *testing.T) {

	// generate new Bitcoin address with private key
	pk := bitcoin.GenerateKeys(true)
	addr, err := MakeAddress(&pk.PublicKey, 0, AddrP2PKH, NetwMain)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("addr = '%s'", addr)

	// generate signature
	msg := "Binding commitment for John Doe"
	b64Sig, err := SignBitcoinMsg(pk, msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("b64sig = '%s'", b64Sig)

	// verify signature
	ok, err := VerifyBitcoinMsg(addr, b64Sig, msg)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature NOT verified")
	}
}
