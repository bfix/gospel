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
	"encoding/hex"
	"testing"
)

var (
	t0 = "0200000001b762afdbca7d9cad9083a3a161eb550ed4553ec22c3e9d3902e43e" +
		"c4eeea3369010000006b483045022100fffea10aa251c1a46c01e980ac0f429c" +
		"866537e67adbdaa90304f528464bcf7e0220168b6fdf61fdb91b2b013dcaae3f" +
		"45467c7906004ee7336ad3dfd021d978d04b01210240de126ab3a20dfad69fa5" +
		"41bde2cc73eaaa6bcc07de96cf914b22c81cf598a6feffffff02d68580010000" +
		"00001976a914b5ea502cb15f248ed0e0cb7fa45a73cee0e061f388ac08d1cb05" +
		"000000001976a91423f583c822b89c65e37f18fa7e2f101ee1105c2a88ac4f15" +
		"1100"
)

func TestSign(t *testing.T) {
	// generate private key
	prv := GenerateKeys(true)
	pub := &prv.PublicKey
	pubHash := Hash160(pub.Bytes())

	// generate sig script
	s := hex.EncodeToString(pubHash)
	sigScr, err := hex.DecodeString("76a9" + s + "88ac")
	if err != nil {
		t.Fatal(err)
	}

	// prepare raw transaction
	tx, err := NewDissectedTransaction(t0)
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.PrepareForSign(0, sigScr); err != nil {
		t.Fatal(err)
	}

	// sign transaction
	sig, err := tx.Sign(prv, 0x01)
	if err != nil {
		t.Fatal(err)
	}

	// verify signature
	ok, err := tx.Verify(pub, sig)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Signature mismatch")
	}
}
