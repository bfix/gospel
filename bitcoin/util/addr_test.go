package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"testing"
)

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
	data = []TestData{
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
	for _, d := range data {
		idhex, err := hex.DecodeString(d.IDhex)
		if err != nil {
			t.Fatal("test data failure")
		}
		idaddr, err := Base58Decode(d.IDaddr)
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
		pubkey, err := ecc.PublicKeyFromBytes(pub)
		if err != nil {
			t.Fatal("test data failure")
		}
		if !pubkey.Q.IsOnCurve() {
			t.Fatal("public point not on curve")
		}

		addr := MakeTestAddress(pubkey)

		addr = MakeAddress(pubkey)
		if string(addr) != d.IDaddr {
			t.Fatal("makeaddress failed")
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
		pubser := Base58Encode(b)
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
		prvkey, err := ecc.PrivateKeyFromBytes(prv)
		if err != nil {
			t.Fatal("privatekeyfrombytes failed")
		}
		q := ecc.MultBase(prvkey.D)
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
		prvser := Base58Encode(b)
		if prvser != d.SerPrvB58 {
			t.Fatal("test data failure")
		}
	}
}

var (
	privKey = "L35JWBbB2nXH6pEzmTGjTnQkRS4fWT7tRKyQhfH9oW9JqffVMgVL"
)

func TestPrivKeyAddress(t *testing.T) {
	b, err := Base58Decode(privKey)
	if err != nil {
		t.Fatal("Base58 decoder failed.: " + err.Error())
	}
	fmt.Printf("*** %s [%d]\n", hex.EncodeToString(b), len(b))
	for i := 1; i < 5; i++ {
		b = b[i : 33+i]
		prv, err := ecc.PrivateKeyFromBytes(b)
		if err != nil {
			t.Fatal("PrivateKeyFromBytes failed: " + err.Error())
		}
		addr := MakeAddress(&prv.PublicKey)
		fmt.Printf("*** %d: %s\n", i, addr)
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
