//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix  >Y<
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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/bfix/gospel/bitcoin"
)

// Test Bitcoin address functions
// ==============================

// Serialization format:
// 	--  4 byte: version bytes (mainnet: 0x0488B21E public, 0x0488ADE4 private; testnet: 0x043587CF public, 0x04358394 private)
//  --  1 byte: depth: 0x00 for master nodes, 0x01 for level-1 descendants, ....
//  --  4 bytes: the fingerprint of the parent's key (0x00000000 if master key)
//  --  4 bytes: child number. This is the number i in xi = xpar/i, with xi the key being serialized. This is encoded in MSB order. (0x00000000 if master key)
//  -- 32 bytes: the chain code
//  -- 33 bytes: the public key or private key data (0x02 + X or 0x03 + X for public keys, 0x00 + k for private keys)
//
//  pub_main: 0488b21e 00 00000000 00000000 873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d508 03+39a36013301597daef41fbe593a02cc513d0b55527ec2df1050e2e8ff49c85c2
//  prv_main: 0488ade4 00 00000000 00000000 873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d508 00+e8f32e723decf4051aefac8e2c93c9c5b214313817cdb01a1494b917c8436b35

type TestData struct {
	IDhex     string
	IDaddr    string
	PubHex    string
	Chain     string
	SerPubHex string
	SerPrvHex string
	SerPubB58 string
	SerPrvB58 string
}

var (
	testData = []TestData{
		{
			"3442193e1bb70916e914552172cd4e2dbc9df811",
			"15mKKb2eos1hWa6tisdPwwDC1a5J1y9nma",
			"0339a36013301597daef41fbe593a02cc513d0b55527ec2df1050e2e8ff49c85c2",
			"873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d508",
			"0488b21e000000000000000000873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d5080339a36013301597daef41fbe593a02cc513d0b55527ec2df1050e2e8ff49c85c2",
			"0488ade4000000000000000000873dff81c02f525623fd1fe5167eac3a55a049de3d314bb42ee227ffed37d50800e8f32e723decf4051aefac8e2c93c9c5b214313817cdb01a1494b917c8436b35",
			"xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
			"xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
		},
		{
			"bd16bee53961a47d6ad888e29545434a89bdfe95",
			"1JEoxevbLLG8cVqeoGKQiAwoWbNYSUyYjg",
			"03cbcaa9c98c877a26977d00825c956a238e8dddfbd322cce4f74b0b5bd6ace4a7",
			"60499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd9689",
			"0488b21e00000000000000000060499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd968903cbcaa9c98c877a26977d00825c956a238e8dddfbd322cce4f74b0b5bd6ace4a7",
			"0488ade400000000000000000060499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd9689004b03d6fc340455b363f51020ad3ecca4f0850280cf436c70c727923f6db46c3e",
			"xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
			"xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
		},
	}

	versionMainPublic  = "0488b21e"
	versionMainPrivate = "0488ade4"
	versionTestPublic  = "043587cf"
	versionTestPrivate = "04358394"
)

func TestAddress(t *testing.T) {
	for _, d := range testData {
		idhex, err := hex.DecodeString(d.IDhex)
		if err != nil {
			t.Fatal("test data failure")
		}
		idaddr, err := bitcoin.Base58Decode(d.IDaddr)
		if err != nil {
			t.Fatal("test data failure")
		}
		if !bytes.Equal(idaddr[1:len(idhex)+1], idhex) {
			t.Fatal("test data mismatch")
		}
		pub, err := hex.DecodeString(d.PubHex)
		if err != nil {
			t.Fatal("test data failure")
		}
		pubkey, err := bitcoin.PublicKeyFromBytes(pub)
		if err != nil {
			t.Fatal("test data failure")
		}
		if !pubkey.Q.IsOnCurve() {
			t.Fatal("public point not on curve")
		}

		addr := MakeAddress(pubkey, 0, AddrP2PKH, AddrMain)
		if string(addr) != d.IDaddr {
			t.Fatalf("makeaddress failed: '%s' != '%s'\n", string(addr), d.IDaddr)
		}
		pubkeyhex := hex.EncodeToString(pubkey.Bytes())
		if pubkeyhex != d.SerPubHex[90:] {
			t.Fatal("pubkey mismatch")
		}
		if d.Chain != d.SerPubHex[26:90] {
			t.Fatal("chain mismatch")
		}
		if versionMainPublic != d.SerPubHex[:8] {
			t.Fatal("version mismatch")
		}
		b, err := hex.DecodeString(d.SerPubHex)
		if err != nil {
			t.Fatal("test data failure")
		}
		b = append(b, prefix(b)...)
		pubser := bitcoin.Base58Encode(b)
		if pubser != d.SerPubB58 {
			t.Fatal("test data failure")
		}
		prv, err := hex.DecodeString(d.SerPrvHex[90:])
		if err != nil {
			t.Fatal("test data failure")
		}
		if len(prv) == 33 {
			if prv[0] != 0 {
				t.Fatal("no leading zero")
			}
			prv = prv[1:]
		}
		prvkey, err := bitcoin.PrivateKeyFromBytes(prv)
		if err != nil {
			t.Fatal("privatekeyfrombytes failed")
		}
		q := bitcoin.MultBase(prvkey.D)
		if !q.Equals(pubkey.Q) {
			t.Fatal("pub/private mismatch")
		}
		if d.Chain != d.SerPubHex[26:90] {
			t.Fatal("chain mismatch")
		}
		if versionMainPrivate != d.SerPrvHex[:8] {
			t.Fatal("version mismatch")
		}
		b, err = hex.DecodeString(d.SerPrvHex)
		if err != nil {
			t.Fatal("test data failure")
		}
		b = append(b, prefix(b)...)
		prvser := bitcoin.Base58Encode(b)
		if prvser != d.SerPrvB58 {
			t.Fatal("test data failure")
		}
	}
}

var (
	tPrivKey = "L35JWBbB2nXH6pEzmTGjTnQkRS4fWT7tRKyQhfH9oW9JqffVMgVL"
	tAddr    = "14Wf6fPLEawQq5zSaCkAJ1Upgaekvy1Hiy"
)

func TestPrivKeyAddress(t *testing.T) {
	b, err := bitcoin.Base58Decode(tPrivKey)
	if err != nil {
		t.Fatal("Base58 decoder failed.: " + err.Error())
	}
	prv, err := bitcoin.PrivateKeyFromBytes(b[1:34])
	if err != nil {
		t.Fatal("PrivateKeyFromBytes failed: " + err.Error())
	}
	addr := MakeAddress(&prv.PublicKey, 0, AddrP2PKH, AddrMain)
	if addr != tAddr {
		t.Fatal("address mismatch")
	}
}

func TestBCHAddress(t *testing.T) {
	pk := "0316b88b26b842eb141031cb3d29e2bb4ccccf595cfa7bb895cbbaa3f1536223d1"
	tAddr := "bitcoincash:qpnfc27ttwqky82emu6mvwtqphg94y4ahc957hjwhp"

	pub, err := hex.DecodeString(pk)
	if err != nil {
		t.Fatal("test data failure")
	}
	pubkey, err := bitcoin.PublicKeyFromBytes(pub)
	if err != nil {
		t.Fatal("test data failure")
	}
	if !pubkey.Q.IsOnCurve() {
		t.Fatal("public point not on curve")
	}
	addr := MakeAddress(pubkey, 145, AddrP2PKH, AddrMain)
	if addr != tAddr {
		t.Fatalf("failed: '%s' != '%s'\n", addr, tAddr)
	}
}

func TestETHAddress(t *testing.T) {
	pk := "034a5823d9d25a434ae893518e121c39fbdcb7ec688974ab23994ddbdb776152d3"
	tAddr := "0xa5b14fd4d99bd75ae22897fb827d76e3310e36f8"

	pub, err := hex.DecodeString(pk)
	if err != nil {
		t.Fatal("test data failure")
	}
	pubkey, err := bitcoin.PublicKeyFromBytes(pub)
	if err != nil {
		t.Fatal("test data failure")
	}
	if !pubkey.Q.IsOnCurve() {
		t.Fatal("public point not on curve")
	}
	addr := MakeAddress(pubkey, 60, AddrP2PKH, AddrMain)
	if addr != tAddr {
		t.Fatalf("failed: '%s' != '%s'\n", addr, tAddr)
	}
}

func prefix(b []byte) []byte {
	sha256 := sha256.New()
	sha256.Write(b)
	h := sha256.Sum(nil)
	sha256.Reset()
	sha256.Write(h)
	cs := sha256.Sum(nil)
	return cs[:4]
}
