/*
 * Address-related test functions.
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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// type for test data
//
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
	IdHex     string
	IdAddr    string
	SecHex    string
	SecWif    string
	PubHex    string
	Chain     string
	SerPubHex string
	SerPrvHex string
	SerPubB58 string
	SerPrvB58 string
}

///////////////////////////////////////////////////////////////////////
// test data definitions

var (
	data = []TestData{
		{
			"3442193e1bb70916e914552172cd4e2dbc9df811",
			"15mKKb2eos1hWa6tisdPwwDC1a5J1y9nma",
			"e8f32e723decf4051aefac8e2c93c9c5b214313817cdb01a1494b917c8436b35",
			"L52XzL2cMkHxqxBXRyEpnPQZGUs3uKiL3R11XbAdHigRzDozKZeW",
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
			"4b03d6fc340455b363f51020ad3ecca4f0850280cf436c70c727923f6db46c3e",
			"KyjXhyHF9wTphBkfpxjL8hkDXDUSbE3tKANT94kXSyh6vn6nKaoy",
			"03cbcaa9c98c877a26977d00825c956a238e8dddfbd322cce4f74b0b5bd6ace4a7",
			"60499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd9689",
			"0488b21e00000000000000000060499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd968903cbcaa9c98c877a26977d00825c956a238e8dddfbd322cce4f74b0b5bd6ace4a7",
			"0488ade400000000000000000060499f801b896d83179a4374aeb7822aaeaceaa0db1f85ee3e904c4defbd9689004b03d6fc340455b363f51020ad3ecca4f0850280cf436c70c727923f6db46c3e",
			"xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
			"xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
		},
	}

	VERSION_MAIN_PUBLIC  = "0488b21e"
	VERSION_MAIN_PRIVATE = "0488ade4"
	VERSION_TEST_PUBLIC  = "043587cf"
	VERSION_TEST_PRIVATE = "04358394"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestAddress(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("bitcoin/util/addr Test")
	fmt.Println("********************************************************")

	for _, d := range data {

		//-------------------------------------------------------------
		// check address data
		//-------------------------------------------------------------

		idhex, err := hex.DecodeString(d.IdHex)
		if err != nil {
			fmt.Println("IdHex decoding error")
			t.Fail()
			return
		}
		idaddr, err := Base58Decode(d.IdAddr)
		if err != nil {
			fmt.Println("IdAddr decoding error")
			t.Fail()
			return
		}

		if !bytes.Equal(idaddr[1:len(idhex)+1], idhex) {
			fmt.Println("IdAddr -- IdHex mismatch")
			t.Fail()
		}

		//-------------------------------------------------------------
		// check public key data
		//-------------------------------------------------------------

		pub, err := hex.DecodeString(d.PubHex)
		if err != nil {
			fmt.Println("PubHex decoding error")
			t.Fail()
			return
		}
		pubkey, err := ecc.PublicKeyFromBytes(pub)
		if err != nil {
			fmt.Println("PubHex conversion failed")
			t.Fail()
			return
		}
		if !ecc.IsOnCurve(pubkey.Q) {
			fmt.Println("PubKey no a point on curve")
			t.Fail()
			return
		}
		addr := MakeAddress(pubkey)
		if string(addr) != d.IdAddr {
			fmt.Println("Address(pubKey) failed")
			fmt.Println(">> " + addr)
			fmt.Println(">> " + d.IdAddr)
			t.Fail()
			return
		}

		pubkeyhex := hex.EncodeToString(pubkey.Bytes())
		if pubkeyhex != d.SerPubHex[90:] {
			fmt.Println("SerPubHex -- key mismatch")
			t.Fail()
		}
		if d.Chain != d.SerPubHex[26:90] {
			fmt.Println("Chain mismatch")
			t.Fail()
		}
		if VERSION_MAIN_PUBLIC != d.SerPubHex[:8] {
			fmt.Println("VERSION_MAIN_PUBLIC mismatch")
			t.Fail()
		}
		b, err := hex.DecodeString(d.SerPubHex)
		if err != nil {
			fmt.Println("SerPubHex decode failure")
			t.Fail()
		}
		b = append(b, prefix(b)...)
		pubser := Base58Encode(b)
		if pubser != d.SerPubB58 {
			fmt.Println("Public B58 mismatch")
			t.Fail()
		}

		//-------------------------------------------------------------
		// check private key data
		//-------------------------------------------------------------

		prv, err := hex.DecodeString(d.SerPrvHex[90:])
		if err != nil {
			fmt.Println("PrvHex decoding error: " + err.Error())
			t.Fail()
			return
		}
		// skip leading zeros
		if len(prv) == 33 {
			if prv[0] != 0 {
				fmt.Println("PrvHex wrong format")
				t.Fail()
				return
			}
			prv = prv[1:]
		}
		prvkey, err := ecc.PrivateKeyFromBytes(prv)
		if err != nil {
			fmt.Println("PrvHex conversion failed: " + err.Error())
			t.Fail()
			return
		}
		q := ecc.ScalarMultBase(prvkey.D)
		if !ecc.IsEqual(q, pubkey.Q) {
			fmt.Println("public private key mismatch")
			t.Fail()
		}
		if d.Chain != d.SerPubHex[26:90] {
			fmt.Println("Chain mismatch")
			t.Fail()
		}
		if VERSION_MAIN_PRIVATE != d.SerPrvHex[:8] {
			fmt.Println("VERSION_MAIN_PRIVATE mismatch")
			t.Fail()
		}
		b, err = hex.DecodeString(d.SerPrvHex)
		if err != nil {
			fmt.Println("SerPrvHex decode failure")
			t.Fail()
		}
		b = append(b, prefix(b)...)
		prvser := Base58Encode(b)
		if prvser != d.SerPrvB58 {
			fmt.Println("Public B58 mismatch")
			t.Fail()
		}
	}
}

// helper: compute double-SHA256 prefix (4 bytes)
func prefix(b []byte) []byte {

	sha256 := sha256.New()
	sha256.Write(b)
	h := sha256.Sum(nil)
	sha256.Reset()
	sha256.Write(h)
	cs := sha256.Sum(nil)
	return cs[:4]
}
