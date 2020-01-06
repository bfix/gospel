package wallet

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"encoding/hex"
	"fmt"
	"testing"
)

var (
	xserData = []string{
		"xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
		"xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
		"xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
		"xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
	}
	pathData = [][]string{
		{"m/0'",
			"xpub68Gmy5EdvgibQVfPdqkBBCHxA5htiqg55crXYuXoQRKfDBFA1WEjWgP6LHhwBZeNK1VTsfTFUHCdrfp1bgwQ9xv5ski8PX9rL2dZXvgGDnw",
			"xprv9uHRZZhk6KAJC1avXpDAp4MDc3sQKNxDiPvvkX8Br5ngLNv1TxvUxt4cV1rGL5hj6KCesnDYUhd7oWgT11eZG7XnxHrnYeSvkzY7d2bhkJ7",
		},
		{"m/0'/1",
			"xpub6ASuArnXKPbfEwhqN6e3mwBcDTgzisQN1wXN9BJcM47sSikHjJf3UFHKkNAWbWMiGj7Wf5uMash7SyYq527Hqck2AxYysAA7xmALppuCkwQ",
			"xprv9wTYmMFdV23N2TdNG573QoEsfRrWKQgWeibmLntzniatZvR9BmLnvSxqu53Kw1UmYPxLgboyZQaXwTCg8MSY3H2EU4pWcQDnRnrVA1xe8fs",
		},
		{"m/0'/1/2'",
			"xpub6D4BDPcP2GT577Vvch3R8wDkScZWzQzMMUm3PWbmWvVJrZwQY4VUNgqFJPMM3No2dFDFGTsxxpG5uJh7n7epu4trkrX7x7DogT5Uv6fcLW5",
			"xprv9z4pot5VBttmtdRTWfWQmoH1taj2axGVzFqSb8C9xaxKymcFzXBDptWmT7FwuEzG3ryjH4ktypQSAewRiNMjANTtpgP4mLTj34bhnZX7UiM",
		},
		{"m/0'/1/2'/2",
			"xpub6FHa3pjLCk84BayeJxFW2SP4XRrFd1JYnxeLeU8EqN3vDfZmbqBqaGJAyiLjTAwm6ZLRQUMv1ZACTj37sR62cfN7fe5JnJ7dh8zL4fiyLHV",
			"xprvA2JDeKCSNNZky6uBCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334",
		},
		{"m/0'/1/2'/2/1000000000",
			"xpub6H1LXWLaKsWFhvm6RVpEL9P4KfRZSW7abD2ttkWP3SSQvnyA8FSVqNTEcYFgJS2UaFcxupHiYkro49S8yGasTvXEYBVPamhGW6cFJodrTHy",
			"xprvA41z7zogVVwxVSgdKUHDy1SKmdb533PjDz7J6N6mV6uS3ze1ai8FHa8kmHScGpWmj4WggLyQjgPie1rFSruoUihUZREPSL39UNdE3BBDu76",
		},
	}

	hd *HD
)

func init() {
	seed, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	hd = NewHD(seed)
}

func TestHDPublic(t *testing.T) {
	path := pathData[2][0]
	pub, err := hd.Public(path)
	if err != nil {
		t.Fatal(err)
	}
	hdp := NewHDPublic(pub, path)

	path2 := pathData[4][0]
	pub2, err := hdp.Public(path2)
	if err != nil {
		t.Fatal(err)
	}
	pubS := pub2.String()
	s := pathData[4][1]
	if pubS != s {
		t.Fatalf("%s != %s", pubS, s)
	}
}

func TestParseExtended(t *testing.T) {
	for _, s := range xserData {
		d, err := ParseExtended(s)
		if err != nil {
			t.Fatal(err)
		}
		u := d.String()
		if u != s {
			t.Fatalf("Mismatch: %s != %s", s, u)
		}
	}
}

func TestM(t *testing.T) {
	xprv := "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi"
	xpub := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"
	prv := hd.m
	prvS := prv.String()
	if prvS != xprv {
		t.Fatalf("prv mismatch: %s\n", prvS)
	}
	pub := prv.Public()
	pubS := pub.String()
	if pubS != xpub {
		t.Fatalf("pub mismatch: %s\n", pubS)
	}
}

func TestPath(t *testing.T) {
	for i, p := range pathData {
		prv, err := hd.Private(p[0])
		if err != nil {
			t.Fatal(err)
		}
		prvS := prv.String()
		if prvS != p[2] {
			fmt.Printf("[%d] d=%v\n", i, prv.data)
			d, _ := ParseExtended(p[2])
			fmt.Printf("d=%v\n", d)
			t.Fatalf("prv mismatch: %s\n", prvS)
		}

		pub, err := hd.Public(p[0])
		if err != nil {
			t.Fatal(err)
		}
		pubS := pub.String()
		if pubS != p[1] {
			fmt.Printf("[%d] d=%v\n", i, pub.data)
			d, _ := ParseExtended(p[1])
			fmt.Printf("d=%v\n", d)
			t.Fatalf("pub mismatch: %s\n", pubS)
		}

		if !pub.key.Equals(prv.Public().key) {
			t.Fatal("public key mismach")
		}
	}
}
