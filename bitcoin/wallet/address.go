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
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/bitcoin/script"
	"golang.org/x/crypto/sha3"
)

// Address constants
const (
	// Mainnet/Testnet/Regnet
	NetwMain = 0
	NetwTest = 1
	NetwReg  = 2

	// Address usage
	AddrP2PKH        = 0
	AddrP2SH         = 1
	AddrP2WPKH       = 2
	AddrP2WSH        = 3
	AddrP2WPKHinP2SH = 4
	AddrP2WSHinP2SH  = 5
)

// Errors
var (
	ErrMkAddrPrefix         = errors.New("unknown prefix")
	ErrMkAddrVersion        = errors.New("unknown version")
	ErrMkAddrNotImplemented = errors.New("not implemented")
)

// GetAddrMode returns the numeric value for mode (P2PKH, P2SH, ...)
func GetAddrMode(label string) int {
	switch label {
	case "P2PKH":
		return AddrP2PKH
	case "P2SH":
		return AddrP2SH
	case "P2WPKH":
		return AddrP2WPKH
	case "P2WSH":
		return AddrP2WSH
	case "P2WPKHinP2SH":
		return AddrP2WPKHinP2SH
	case "P2WSHinP2SH":
		return AddrP2WSHinP2SH
	}
	return -1
}

// Addresser is a function prototype for address conversion functions
type Addresser func(pk *bitcoin.PublicKey, coin, version, network, prefix int) (string, error)

// MakeAddress computes an address from a public key for given coin, version
// and network. All cryptocoins based on the Bitcoin curve (secp256k1) are
// supported.
func MakeAddress(key *bitcoin.PublicKey, coin, version, network int) (string, error) {
	// get prefix (and optional addresser)
	prefix, hrp, conv := getPrefix(coin, version, network)

	// call a custom conversion function
	if conv != nil {
		return conv(key, coin, version, network, prefix)
	}
	// if no prefix is found, we can't create address
	if prefix == -1 {
		return "", ErrMkAddrPrefix
	}
	// sanity check: only certain versions allowed
	switch version {
	case AddrP2PKH, AddrP2WPKH, AddrP2WPKHinP2SH:
		return makeAddress(key.Bytes(), hrp, version, prefix)
	default:
		return "", ErrMkAddrVersion
	}
}

// MakeAddressScript creates an address for a P2SH script
func MakeAddressScript(scr *script.Script, coin, version, network int) (string, error) {
	// get prefix (and optional addresser)
	prefix, hrp, _ := getPrefix(coin, version, network)

	// if no prefix is found, we can't create address
	if prefix == -1 {
		return "", ErrMkAddrPrefix
	}
	// sanity check: only P2SH allowed
	if version != AddrP2SH {
		return "", ErrMkAddrVersion
	}
	return makeAddress(scr.Bytes(), hrp, version, prefix)
}

func makeAddress(obj []byte, hrp string, version, prefix int) (string, error) {
	// handle segwit addresses separately
	if version == AddrP2WPKH || version == AddrP2WSH {
		return makeAddressSegWit(obj, hrp, version, prefix)
	}
	// Generic address calculation:
	var data []byte
	switch version {
	case AddrP2PKH, AddrP2SH:
		data = obj
	case AddrP2WPKHinP2SH:
		redeem := append([]byte(nil), 0)
		redeem = append(redeem, 0x14)
		kh := bitcoin.Hash160(obj)
		redeem = append(redeem, kh...)
		data = redeem
	default:
		// can't create address for unknown version
		return "", ErrMkAddrVersion
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
	return bitcoin.Base58Encode(addr), nil
}

func makeAddressSegWit(obj []byte, hrp string, version, prefix int) (string, error) {
	// compute address data
	var data []byte
	switch version {
	case AddrP2WPKH:
		data = bitcoin.Hash160(obj)
	case AddrP2WSH:
	default:
		return "", ErrMkAddrVersion
	}
	// encode data to 5-bit sequence and add leading witness version
	buf := new(bytes.Buffer)
	buf.WriteByte(0) // witness version
	buf.Write(Bech32Bit5(data))

	// compute checksum and append to buffer
	crc := Bech32CRC(hrp, buf.Bytes())
	buf.Write(crc)

	// encode to Bech32 charset
	b32enc := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	addr := ""
	for _, v := range buf.Bytes() {
		addr += string(b32enc[v])
	}
	return hrp + "1" + addr, nil
}

//======================================================================
// custom address calculation functions
//======================================================================

// ETH (Ethereum) address
func makeAddressETH(key *bitcoin.PublicKey, coin, version, network, prefix int) (string, error) {
	pkData := key.Q.Bytes(false)
	hsh := sha3.NewLegacyKeccak256()
	hsh.Write(pkData[1:])
	val := hsh.Sum(nil)
	return "0x" + hex.EncodeToString(val[12:]), nil
}

// BCH (Bitcoin Cash) address
func makeAddressBCH(key *bitcoin.PublicKey, coin, version, network, prefix int) (string, error) {
	// segwit handling is generic
	if version == AddrP2WPKHinP2SH || version == AddrP2WPKH {
		return makeAddress(key.Bytes(), "bc", version, prefix)
	}
	// special handling for P2PKH addresses
	// get data for address
	var data []byte
	switch version {
	case AddrP2PKH:
		data = key.Bytes()
	case AddrP2WPKHinP2SH:
		redeem := append([]byte(nil), 0)
		redeem = append(redeem, 0x14)
		kh := bitcoin.Hash160(key.Bytes())
		redeem = append(redeem, kh...)
		data = redeem
	default:
		return "", ErrMkAddrVersion
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
	copy(values[12:], Bech32Bit5(buf))

	crc := func(values []byte) uint64 {
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
	res := new(bytes.Buffer)
	_ = binary.Write(res, binary.BigEndian, crc(values))
	return addr + strings.Trim(b32.EncodeToString(res.Bytes()[3:]), "="), nil
}
