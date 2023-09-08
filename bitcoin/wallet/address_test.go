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

package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bfix/gospel/bitcoin"
)

func TestAddrP2WPKH(t *testing.T) {
	// see: https://en.bitcoin.it/wiki/Bech32
	data, err := hex.DecodeString("0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")
	if err != nil {
		t.Fatal(err)
	}
	addr, err := makeAddressSegWit(data, "bc", AddrP2WPKH)
	if err != nil {
		t.Fatal(err)
	}
	if addr != "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4" {
		t.Log(addr)
		t.Fatal("addr mismatch")
	}
}

type testData struct {
	path    string
	xpub    string
	coin    int
	version int
	netw    int
	addrs   []string
}

var (
	words = []string{
		"sketch", "blast", "judge", "ladder", "answer", "twice",
		"outer", "fiction", "finish", "dice", "true", "later",
		"vicious", "engage", "gravity", "diary", "inside", "ignore",
		"giggle", "surge", "turkey", "outside", "panther", "timber",
	}
	seed = "d761a2b872860dc981208a9bf3729d3c7234fb6bcf9446dc59ded41928d83999fc467f3622ed95d69c9d400f961f698abf5210a2529b4f0a1b4ec7557a1c4529"
	xprv = "xprv9s21ZrQH143K2rimReigYY8rKViMqHYQ2URn6PqRzNRa3Fs75nasDTJwzLkQmNB9PuhNh2U9Vfnxt1WY5qBua21dkJxXtA3byvBPyaJiKRG"

	testdata = []*testData{
		//----------------------------------------------------------
		// Bitcoin (P2PKH, P2WPKHinP2SH, P2WPK) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/0'/0'/0",
			xpub:    "xpub6E8YMhNAYXN7ida8ydegYYCvTSRmjAzvLTkpg3fqF4mbVQ5ygzc97fEbBNUEFcbtBXVsWjMUoTfP4fBQUiurBSzE1tV19ayw86mnAMbAEr2",
			coin:    0,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"1GNBLXbeQW3XiS9NBuAFUbUVWtdeBAW7Ks",
				"1BT99WDMk8LynsYdy5yVfvQyNeigMSJ6af",
				"18N1oUMgjmDxCb14ZYgJBvnvRwAZVZkqVh",
			},
		},
		{
			path:    "m/49'/0'/0'/0",
			xpub:    "ypub6YzfSXyntdLnruYh5nbmnxtvVoQbDw52JepTtDpS5a1TBxY6GuSveKwfx3aXWCDv6LqHWmQPwQSrBq1pc9Ro3wjjXJXPCY7tN4GqhLSML4H",
			coin:    0,
			version: AddrP2WPKHinP2SH,
			netw:    NetwMain,
			addrs: []string{
				"32FjpH9aEdrsoT7PBHuWsVZJQ1zzmRS7yp",
				"3HrttpWmv3TNmQGSyoujTccZYXqhXMMXXh",
				"3DRpWwqdDVmqdTMtFEzNiyW12vvr2M3JHj",
			},
		},
		{
			path:    "m/84'/0'/0'/0",
			xpub:    "zpub6t3DJ2MnEfmGZmyC5j9psWscdDDWgnbY9EeKS5UxeEBPr7HHmSsQmwsGm7fW5mtvX5j1sk3qYzxA5R9xN4ZqoVR2WFXNUUaWdY9ktiQC2Rt",
			coin:    0,
			version: AddrP2WPKH,
			netw:    NetwMain,
			addrs: []string{
				"bc1qfd2vv0ezx40e525x4m008u02a68sxlahvd0u20",
				"bc1qx4c27a4z70kpt70sz5syd7q3cjas23l8qsmaqp",
				"bc1qztt29vrmj82jrtujxah85pxjxxsvn62mgx2fg5",
			},
		},
		//----------------------------------------------------------
		// Litecoin (P2PKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/2'/0'/0",
			xpub:    "xpub6FLSDas21s8MHLfKKUyh2fgNKFzboV71BBF2TMWwwiJ9Q3q6Sciw4FXXxywLwY3psCUx1rgbyb5MtQxhtgCzESDEi2RMsyUu6iA9UqkP6Xn",
			coin:    2,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"LPpE1LKHJ3urKcexYzUGg29AqqPPmaRYfS",
				"LPiLNJxeQJdbNFU7PoPC1MkBBADCz2T6dm",
				"LYsmBRnSP9gt1ZCmuJHX9JYJTTSLfzFSrr",
			},
		},
		//----------------------------------------------------------
		// Dogecoin (P2PKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/3'/0'/0",
			xpub:    "dgub8uPNgcUAVx9LYQbFzCEhXbofX3EKZQu3S4soeg2uMg39J2auUCBZVxigd75FHvwyWoCFdrqw78o4AtdZ87MSq1uqmxCvYyymnEAzmPKeJwh",
			coin:    3,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"DFCvQ2fiAcypiz79Uc1pKnPGB3zyBapcof",
				"DGrPRX29ucppn3wSCCBiXpf956g2wWdXf2",
				"DTuZDuC7z9QtCKfC5yGuGLyPPJX18eWzg8",
			},
		},
		//----------------------------------------------------------
		// Dash (P2PKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/5'/0'/0",
			xpub:    "drkpS2VHVB4LqrEYwahzHoekQn8ZC1jJxFiYMkXvpVJxAAxFz6FcDiPKZA2qKS9pb1LDLjbtA4iuXpYssiZ2WgoVvqxVKeWZHybyWseEVxnDNR7",
			coin:    5,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"XjpM1S5ijJ7SV7NGzQL2psZYKFvPkm4WET",
				"XvQwicQbChzMa1UEU6hudNuHAQJUDJ3J5z",
				"Xsx5nfMs8XyVnPYWzFcLKgJJ1bM7Y7zJX7",
			},
		},
		//----------------------------------------------------------
		// Namecoin (P2PKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/7'/0'/0",
			xpub:    "xpub6DoCUxVGdMNrrmxJL2qYmbWcxMUF2nMX4MGE5kdaV2AaR5PFXVKMfVeZ9R217ex9z9CY4J15yL1K2fU14tsSC248U5Z4qPpnzKHXKzkShCN",
			coin:    7,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"Myn4hTwq74yVYFcfZdox7zBJChjD7pe62S",
				"N6voiyaKrWMnPcBkwD4iyp91VtPiS9K4zA",
				"NG5TJ7oEyLUC5XRBJNypefpnsyD5rj3FNi",
			},
		},
		//----------------------------------------------------------
		// Digibyte (P2PKH,P2WPKHinP2SH,P2WPKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/20'/0'/0",
			xpub:    "xpub6F8a29k7hbo1tabHPuRrynK8jBH2NJPWjVzyxsZJGPjVCqbQCSizn4UNsuB8W82ACr94dEbBa8QUkhitvtQqoj94hb9mAxEKakGVDvCZQp9",
			coin:    20,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"D5kyuUeGW8tehUuyLmCEqyyQLAEpi8N1KH",
				"DNbBpZ7euKXUUHAGY2JBTFH353pHwyv29X",
				"DEYDg68hggMGGrWCCkqpdpRFo8GtWYo8Ps",
			},
		},
		{
			path:    "m/49'/20'/0'/0",
			xpub:    "ypub6ZE3Ch8McCBR1zhr5dLR1NECkXac3BodVA8Jp3auSDxM2SrEiFS9Q7NKpa8QvnkjoWSS6vBuF82tJQfPDqVoMrH63EoBxLUgV3X1rEcUJea",
			coin:    20,
			version: AddrP2WPKHinP2SH,
			netw:    NetwMain,
			addrs: []string{
				"ScSW6s7aekqdTpnWFvCwgdq6v7uULZyCoR",
				"Sk41EPXz8EwN2n7ZrFLdcrUA5qgMNrKSZU",
				"SiYe3tupyeka8uAicYpA5QA1nMUwtUtWSe",
			},
		},
		{
			path:    "m/84'/20'/0'/0",
			xpub:    "zpub6trsrUqRjGgQ487C1gmR8R3WifqdfHaiNsnsF9XYNUZAwivfAMWdBZrHVXiKWmWYcyaXNMLVQcKrwryMZhLegT53BUAjdnEYehHXJCPcTAq",
			coin:    20,
			version: AddrP2WPKH,
			netw:    NetwMain,
			addrs: []string{
				"dgb1qlvl7ka8cakj332rthdx5yvfkrjlc9z6sn3faeh",
				"dgb1q5u8pxydx30pkd5zhhrwukrpt2mqdt2nmfap4cf",
				"dgb1q7eucwwyztd7m2cmgm0sqaw4nh5u46ak7rxmuun",
			},
		},
		//----------------------------------------------------------
		// Vertcoin (P2PKH,P2WPKHinP2SH,P2WPKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/28'/0'/0",
			xpub:    "xpub6F2hZN9fs9mqsM1oTyGLcjYihh6xjjf5rzZaYvGYRXJY5Y43gV5UWaVL6mhpLGy6CDsrbg7D8fgLxv1Duv5EBeCA8vwjNfGyfPCkjPVo83g",
			coin:    28,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"VtSpg52AY5k33VoKjz1zNvaVSyfryBEeJb",
				"VhbY94Va9kW8kEDKwBx5qgospmQ7NhPvDS",
				"Vwc66CTe1SBi8M8gCso6HEWyEbpx2J1oZk",
			},
		},
		{
			path:    "m/49'/28'/0'/0",
			xpub:    "xpub6FHDPcB4fiie725sZqFD2641Y2vFSkoZ3c36kRQ26dDtHie1cQhr2VrjbYgYicDmeccFhshYUhkrwmZinmvYPyutJJ3Pm5Ck6boyu6QeTYT",
			coin:    28,
			version: AddrP2WPKHinP2SH,
			netw:    NetwMain,
			addrs: []string{
				"3LQvZbae7mHvjNVV3AxjvaCohRLTJi2yBj",
				"3Hvjzfz9PQRzJfQBeKKj8kFoBd5Pt8Bwhk",
				"3FoEmcuSwpHu7jNB8z4fPeteVfiYqw3CpV",
			},
		},
		{
			path:    "m/84'/28'/0'/0",
			xpub:    "xpub6EmQfEi8DSi9fGCYPNKiznCCDycEt2nXEeDK8XfYWbcvnfK7G1iF1UhxHqfYwJ3u6zBYMN8yvU3wvEqjE6aMGcvbZ4ZnfHD7A9PYs8uRzAs",
			coin:    28,
			version: AddrP2WPKH,
			netw:    NetwMain,
			addrs: []string{
				"vtc1q0e9cp24vd5yfcw9a9m06ww0pc8t2ftu9unpk03",
				"vtc1qk5vg36emrls25cljqhp997p0gtkvrzlprct5an",
				"vtc1qwcw7705m4u3h97vy5u2msus6gn75e4w72zv8qq",
			},
		},
		//----------------------------------------------------------
		// ZCash (P2PKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/133'/0'/0",
			xpub:    "xpub6FHRtWCLBuxiaDgCYAGtec9MdDoSCSwSHCcW41mSGAH2ciMuKs5L1dWdAGqKpeWsBbixqH5MUebXR6JEEsdyKdjauE8kBg35uC2YFscmE8u",
			coin:    133,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"t1UH9hfYFjRjPpZjRjgVQT7cTwUekgaBUV3",
				"t1MrKJrrXBtPG7cJXUgCSEtbNgkt1b73Vpb",
				"t1cq8Q8kDwrwpp5P1gTf2fAXFy1pcb4Tv6E",
			},
		},
		//----------------------------------------------------------
		// BitcoinCash (P2PKH,P2WPKHinP2SH,P2WPKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/145'/0'/0",
			xpub:    "xpub6Dwp81H62AeMQdABvdKCmTtzagoUBxyebQU4tfhH6467QnSepzN87Y2sGgSbwWwyXNyjEWRjyu5mSjWmVp5ZxSE9B4H49oGyRczSZPNQtZ4",
			coin:    145,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"qr09zpf9ex5e0yktxuzyn8r300sw0h9adsk82c94pm",
				"qp9r70xkgz724gfxurm5de0mmyxr8rffjy95dxflyh",
				"qqzpgzkl5ec9khd7w74ajnlycn4p5zgufqnzrgzfrv",
			},
		},
		{
			path:    "m/49'/145'/0'/0",
			xpub:    "ypub6YQ3oQnVh3AXy2sdiZ4sguhNLV8eZp8ifVfAYnAnv1LWkEs8UnKcUdJbZHHJhtFMNqWd3T6YufJ9NFZYRqjoJ7AucHEnF21V86gJtM8cQMK",
			coin:    145,
			version: AddrP2WPKHinP2SH,
			netw:    NetwMain,
			addrs: []string{
				"39AhdB7euaHGrxUpJAHVe4WS1ccRspi1Wr",
				"37BkU5wzVGRHxP4i7Vspp4h5jKuDnALNuW",
				"3B87eHgk6H6QxBZNKz8RcjxWMLKdGXsU7C",
			},
		},
		{
			path:    "m/84'/145'/0'/0",
			xpub:    "zpub6tyomv3Hh7DbrXfGwHLSdgJqcDPtBnysLUGBn98cF4L8GicbSaDu3YQQEKMwXVqMgJsgHVPkxEyn38vmAG7JWLfy7xRA2gAVMY9aAVin2vz",
			coin:    145,
			version: AddrP2WPKH,
			netw:    NetwMain,
			addrs: []string{
				"bc1qm2747f3l92f8cplkfgatajzsumnlr4agu0as97",
				"bc1qzc2csynw50gggs2wa6gkezxptpms3kwj5cjvm4",
				"bc1qm2wekscdsaxl2fqstvzf7eh89zl60t5d6wkvkn",
			},
		},
		//----------------------------------------------------------
		// BitcoinGold (P2PKH,P2WPKHinP2SH,P2WPKH) (NetwMain)
		//----------------------------------------------------------
		{
			path:    "m/44'/156'/0'/0",
			xpub:    "xpub6FKmGqCHQ9B8TJApnGyfpHprWMQNj2chDnvuQ6VqokyPFq3DbroAQGj8XGJuxTMqbgH8SHBTURc87ptnYe8TWDcPiGD6kFE5vJcqJe1MP2m",
			coin:    156,
			version: AddrP2PKH,
			netw:    NetwMain,
			addrs: []string{
				"GKJiXkT4jw35tirbrWNbg6mrQ7e3bjxps8",
				"GJpWLy6rAni6LrHGMVQtdNMwnfQP6YXcnp",
				"GQee2mtTxn2BvGivKVpgB8nE1hUzYRKyYX",
			},
		},
		{
			path:    "m/49'/156'/0'/0",
			xpub:    "ypub6ZPapZb2XhNLQ1e5DeAUQMfwfB7Gfv95nVw91wNgRd9W1Skab2VzrTyhz4JAKbqBEThaNqTheTBHhvCFeq1SGpBGL9hN1U61A1YioWxS7qu",
			coin:    156,
			version: AddrP2WPKHinP2SH,
			netw:    NetwMain,
			addrs: []string{
				"AReJZiyQ71Fp5rBzz94yzfDtSFF2ShAkkr",
				"AR5KraC6hkQ1vYDHTTD9i1UjDUfJPEEGgx",
				"AV4CiYBnC6mVtmSgoEZDuJF3BeV33ctrwC",
			},
		},
		{
			path:    "m/84'/156'/0'/0",
			xpub:    "zpub6sFMAk1YgUdX5EtNevXBXkjPhqtqnUSDtABAxtWH6MsKVqPaBiaXpt31STMAbejMV5RSs78AYqd3XmpgEnbn171dZNMBGF67BWe4zJYKMbM",
			coin:    156,
			version: AddrP2WPKH,
			netw:    NetwMain,
			addrs: []string{
				"btg1qhz0rez3v2366n0rwqdekdfp8s745puqkshmtjx",
				"btg1qn7svv7nghkcxqs3t6kz9dw2xhlms4yym6jp7hl",
				"btg1qkc0255mmwr7muc0mm067a3lty069jf3jjdasah",
			},
		},
	}
)

func TestMakeAddress(t *testing.T) {
	// generate HD wallet
	s1, check := WordsToSeed(words, "")
	if len(check) > 0 {
		t.Fatal(check)
	}
	s2, err := hex.DecodeString(seed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s1, s2) {
		t.Fatal("seed mismatch")
	}
	hd, err = NewHD(s1)
	if err != nil {
		t.Fatal(err)
	}
	prv := hd.MasterPrivate().String()
	if prv != xprv {
		t.Fatal("xprv mismatch")
	}

	// test addresses
	for _, test := range testdata {
		t.Log(test.path)
		pub, err := hd.Public(test.path)
		if err != nil {
			t.Fatal(err)
		}
		pub.Data.Version = GetXDVersion(test.coin, test.version, test.netw, true)

		if pub.String() != test.xpub {
			t.Log(test.xpub)
			t.Log(pub.String())
			t.Fatal("xpub mismatch")
		}
		pubhd := NewHDPublic(pub, test.path)
		for i, addr := range test.addrs {
			epk, err := pubhd.Public(test.path + fmt.Sprintf("/%d", i))
			if err != nil {
				t.Fatal(err)
			}
			pk, err := bitcoin.PublicKeyFromBytes(epk.Data.Keydata)
			if err != nil {
				t.Fatal(err)
			}
			a, err := MakeAddress(pk, test.coin, test.version, test.netw)
			if err != nil {
				t.Fatal(err)
			}
			if a != addr {
				t.Logf("%s != %s\n", a, addr)
				t.Fatal("addr mismatch")
			}

		}
	}
}
