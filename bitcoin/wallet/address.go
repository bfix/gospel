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
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/bfix/gospel/bitcoin"
	"golang.org/x/crypto/sha3"
)

// Address constants
const (
	// Mainnet/Testnet/Regnet
	AddrMain = 0
	AddrTest = 1
	AddrReg  = 2

	// Address usage
	AddrP2PKH        = 0
	AddrP2SH         = 1
	AddrP2WPKH       = 2
	AddrP2WSH        = 3
	AddrP2WPKHinP2SH = 4
	AddrP2WSHinP2SH  = 5
)

// Addresser is a function prototype for address conversion functions
type Addresser func(pk *bitcoin.PublicKey, coin, version, network, prefix int) string

// MakeAddress computes an address from public key for the "real" Bitcoin network
func MakeAddress(key *bitcoin.PublicKey, coin, version, network int) string {

	// get info for selected coin/version/network
	var prefix int = -1
	var conv Addresser = nil
	for _, addr := range AddrList {
		if addr.CoinID == coin {
			conv = addr.Conv
			v := addr.Formats[network]
			if v != nil {
				w := v.Versions[version]
				if w != nil {
					prefix = int(w.Version)
					break
				}
			}
		}
	}
	// call a custom conversion function
	if conv != nil {
		return conv(key, coin, version, network, prefix)
	}
	// if no prefix is found, we can't create address
	if prefix == -1 {
		return ""
	}

	// Generic address conversion:
	// get data for address
	var data []byte
	switch version {
	case AddrP2PKH:
		data = key.Bytes()
	case AddrP2SH:
		redeem := append([]byte(nil), 0)
		redeem = append(redeem, 0x14)
		kh := bitcoin.Hash160(key.Bytes())
		redeem = append(redeem, kh...)
		data = redeem
	}
	var addr []byte
	if prefix > 255 {
		addr = append(addr, byte((prefix>>8)&0xff))
	}
	addr = append(addr, byte(prefix&0xff))
	kh := bitcoin.Hash160(data)
	addr = append(addr, kh...)
	cs := bitcoin.Hash256(addr)
	addr = append(addr, cs[:4]...)
	return string(bitcoin.Base58Encode(addr))
}

// GetXDVersion returns the extended data version for a given coin mode
func GetXDVersion(coin, mode, network int, pub bool) uint32 {
	for _, addr := range AddrList {
		if addr.CoinID == coin {
			v := addr.Formats[network]
			if v != nil {
				w := v.Versions[mode]
				if w != nil {
					if pub {
						return w.PubVersion
					}
					return w.PrvVersion
				}
			}
		}
	}
	// return default
	if pub {
		return XpubVersion
	}
	return XprvVersion
}

// AddrVersion defines address version constants
type AddrVersion struct {
	Version    uint16 // version byte (address prefix)
	PubVersion uint32 // BIP32 key version (public)
	PrvVersion uint32 // BIP32 key version (private)
}

// AddrFormat defines formatting information for addresses
type AddrFormat struct {
	Bech32     string
	WifVersion byte
	Versions   []*AddrVersion
}

// AddrSpec defines a coin address format.
type AddrSpec struct {
	CoinID  int
	Formats []*AddrFormat
	Conv    Addresser
}

var (
	// AddrList for selected coins
	// (see page source for "https://iancoleman.io/bip39/")
	AddrList = []*AddrSpec{
		//--------------------------------------------------------------
		// BTC (Bitcoin)
		//--------------------------------------------------------------
		{0, []*AddrFormat{
			// Mainnet
			{"bc", 0x80, []*AddrVersion{
				{0x00, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2SH
				{0x00, 0x04b24746, 0x04b2430c}, // P2WPKH
				{0x05, 0x02aa7ed3, 0x02aa7a99}, // P2WSH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				{0x05, 0x0295b43f, 0x0295b005}, // P2WSHinP2SH
			}},
			// Testnet
			{"tb", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
			// Regnet
			{"bcrt", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
		}, nil},
		//--------------------------------------------------------------
		// LTC (Litecoin)
		//--------------------------------------------------------------
		{2, []*AddrFormat{
			// Mainnet
			{"ltc", 0xb0, []*AddrVersion{
				{0x30, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x32, 0x01b26ef6, 0x01b26792}, // P2SH
				{0x30, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x32, 0x01b26ef6, 0x01b26792}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"litecointestnet", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x043587cf, 0x04358394}, // P2WPKH
				nil,                            // P2WSH
				{0xc4, 0x043587cf, 0x04358394}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// DOGE
		//--------------------------------------------------------------
		{3, []*AddrFormat{
			// Mainnet
			{"", 0x9e, []*AddrVersion{
				{0x1e, 0x02facafd, 0x02fac398}, // P2PKH
				{0x16, 0x02facafd, 0x02fac398}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"dogecointestnet", 0xf1, []*AddrVersion{
				{0x71, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x71, 0x043587cf, 0x04358394}, // P2WPKH
				nil,                            // P2WSH
				{0xc4, 0x043587cf, 0x04358394}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// DASH
		//--------------------------------------------------------------
		{5, []*AddrFormat{
			// Mainnet
			{"", 0xcc, []*AddrVersion{
				{0x4c, 0x02fe52cc, 0x0488ade4}, // P2PKH
				{0x10, 0x02fe52cc, 0x0488ade4}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"", 0xef, []*AddrVersion{
				{0x8c, 0x043587cf, 0x04358394}, // P2PKH
				{0x13, 0x043587cf, 0x04358394}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// NMC (Namecoin)
		//--------------------------------------------------------------
		{7, []*AddrFormat{
			// Mainnet
			{"", 0xb4, []*AddrVersion{
				{0x34, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x0d, 0x0488b21e, 0x0488ade4}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// DGB (Digibyte)
		//--------------------------------------------------------------
		{20, []*AddrFormat{
			// Mainnet
			{"dgb", 0x80, []*AddrVersion{
				{0x1e, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x3f, 0x049d7cb2, 0x049d7878}, // P2SH
				{0x1e, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x3f, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// VTC (Vertcoin)
		//--------------------------------------------------------------
		{28, []*AddrFormat{
			// Mainnet
			{"", 0x80, []*AddrVersion{
				{0x47, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2SH
				{0x47, 0x0488b21e, 0x0488ade4}, // P2WPKH
				nil,                            // P2WSH
				{0x47, 0x0488b21e, 0x0488ade4}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// ETH (Ethereum)
		//--------------------------------------------------------------
		{60, []*AddrFormat{
			// Mainnet
			nil,
			// Testnet
			nil,
			// Regnet
			nil,
		}, makeAddressETH},
		//--------------------------------------------------------------
		// ETC (Ethereum Classic)
		//--------------------------------------------------------------
		{61, []*AddrFormat{
			// Mainnet
			nil,
			// Testnet
			nil,
			// Regnet
			nil,
		}, makeAddressETH},
		//--------------------------------------------------------------
		// ZEC (ZCash)
		//--------------------------------------------------------------
		{133, []*AddrFormat{
			// Mainnet
			{"", 0x80, []*AddrVersion{
				{0x1cb8, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x1cbd, 0x0488b21e, 0x0488ade4}, // P2SH
				nil,                              // P2WPKH
				nil,                              // P2WSH
				nil,                              // P2WPKHinP2SH
				nil,                              // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}, nil},
		//--------------------------------------------------------------
		// BCH
		//--------------------------------------------------------------
		{145, []*AddrFormat{
			// Mainnet
			{"", 0x80, []*AddrVersion{
				{0x00, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x05, 0x0488b21e, 0x0488ade4}, // P2SH
				{0x00, 0x04b24746, 0x04b2430c}, // P2WPKH
				{0x05, 0x02aa7ed3, 0x02aa7a99}, // P2WSH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				{0x05, 0x0295b43f, 0x0295b005}, // P2WSHinP2SH
			}},
			// Testnet
			{"", 0xef, []*AddrVersion{
				{0x6f, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0xc4, 0x0488b21e, 0x0488ade4}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
			// Regnet
			{"", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
		}, makeAddressBCH},
		//--------------------------------------------------------------
		// BTG
		//--------------------------------------------------------------
		{156, []*AddrFormat{
			// Mainnet
			{"btg", 0x80, []*AddrVersion{
				{0x26, 0x0488b21e, 0x0488ade4}, // P2PKH
				{0x17, 0x049d7cb2, 0x049d7878}, // P2SH
				{0x26, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x17, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}, nil},
	}
)

//======================================================================
// custom address conversion functions
//======================================================================

// ETH (Ethereum) address
func makeAddressETH(key *bitcoin.PublicKey, coin, version, network, prefix int) string {
	pkData := key.Q.Bytes(false)
	hsh := sha3.NewLegacyKeccak256()
	hsh.Write(pkData[1:])
	val := hsh.Sum(nil)
	return "0x" + hex.EncodeToString(val[12:])
}

// BCH (Bitcoin Cash) address
func makeAddressBCH(key *bitcoin.PublicKey, coin, version, network, prefix int) string {

	// bit5 splits a byte array into 5-bit chunks
	bit5 := func(data []byte) []byte {
		size := len(data) * 8
		v := new(big.Int).SetBytes(data)
		pad := size % 5
		if pad != 0 {
			v = new(big.Int).Lsh(v, uint(5-pad))
		}
		num := (size + 4) / 5
		res := make([]byte, num)
		for i := num - 1; i >= 0; i-- {
			res[i] = byte(v.Int64() & 31)
			v = new(big.Int).Rsh(v, 5)
		}
		return res
	}

	// polymod computes a CRC for 5-bit sequences
	polymod := func(values []byte) uint64 {
		var c uint64 = 1
		for _, d := range values {
			c0 := c >> 35
			c = ((c & 0x07ffffffff) << 5) ^ uint64(d)
			if c0&0x01 != 0 {
				c ^= 0x98f2bc8e61
			}
			if c0&0x02 != 0 {
				c ^= 0x79b76d99e2
			}
			if c0&0x04 != 0 {
				c ^= 0xf33e5fb3c4
			}
			if c0&0x08 != 0 {
				c ^= 0xae2eabe2a8
			}
			if c0&0x10 != 0 {
				c ^= 0x1e4f43e470
			}
		}
		return c ^ 1
	}

	// get data for address
	var data []byte
	switch version {
	case AddrP2PKH:
		data = key.Bytes()
	case AddrP2SH:
		redeem := append([]byte(nil), 0)
		redeem = append(redeem, 0x14)
		kh := bitcoin.Hash160(key.Bytes())
		redeem = append(redeem, kh...)
		data = redeem
	}
	var buf []byte
	if prefix > 255 {
		buf = append(buf, byte((prefix>>8)&0xff))
	}
	buf = append(buf, byte(prefix&0xff))
	kh := bitcoin.Hash160(data)
	buf = append(buf, kh...)

	b32 := base32.NewEncoding("qpzry9x8gf2tvdw0s3jn54khce6mua7l")
	addr := strings.Trim(b32.EncodeToString(buf), "=")
	values := make([]byte, 54)
	copy(values, []byte{2, 9, 20, 3, 15, 9, 14, 3, 1, 19, 8, 0})
	copy(values[12:], bit5(buf))
	crc := polymod(values)
	res := new(bytes.Buffer)
	_ = binary.Write(res, binary.BigEndian, crc)
	return addr + strings.Trim(b32.EncodeToString(res.Bytes()[3:]), "=")
}
