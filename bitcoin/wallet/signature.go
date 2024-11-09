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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/math"
)

var (
	ErrSigRecoverFail = errors.New("invalid key recovery from signature")
	ErrSigInvalidHdr  = errors.New("invalid signature header")

	msgHdr = []byte("Bitcoin Signed Message:\n")
)

// VerifyBitcoinMsg verifies a message signature from a Bitcoin address.
// Since different wallets generate signatures with different interpretations
// of the header byte values, all possible signing keys are tested independent
// from the actual value of the header byte; only the lower two bits (recID)
// are used. They are handled correctly by all wallets generating signatures.
func VerifyBitcoinMsg(addr, b64sig, msg string) (ok bool, err error) {

	// handle signature format ("trezor"/"electrum")
	b64sig = strings.TrimPrefix(b64sig, "trezor:")

	// decode base64-encoded Bitcoin signature
	var sigBuf []byte
	if sigBuf, err = base64.StdEncoding.DecodeString(b64sig); err != nil {
		return
	}
	var sig *bitcoin.Signature
	if sig, err = bitcoin.NewSignatureFromBytes(sigBuf[1:]); err != nil {
		return
	}

	// extract recID from header (according to BIP-0137)
	hdr := sigBuf[0]
	compr := true
	var recID byte
	if hdr >= 43 {
		err = ErrSigInvalidHdr
		return
	} else if hdr >= 39 {
		// bech32 signature
		recID = hdr - 39
	} else if hdr >= 35 {
		// segwit p2sh signature
		recID = hdr - 35
	} else if hdr >= 31 {
		// compressed key signature
		recID = hdr - 31
	} else if hdr >= 27 {
		// uncompressed key signature
		recID = hdr - 27
		compr = false
	} else {
		err = ErrSigInvalidHdr
		return
	}

	// recover public key
	pk, ok := recoverFromSignature(recID, sig, msg, compr)
	if !ok {
		err = ErrSigRecoverFail
		return
	}

	// ignore the header value and generate all possible addresses:
	ah_u := bitcoin.Hash160(pk.Q.Bytes(false))
	ah_c := bitcoin.Hash160(pk.Q.Bytes(true))

	ok = func() bool {
		// (1) ECDSA verification, uncompressed P2PKH address
		a := append([]byte{0}, ah_u...)
		cs := bitcoin.Hash256(a)
		a = append(a, cs[:4]...)
		da := bitcoin.Base58Encode(a)
		if da == addr {
			return true
		}

		// (2) ECDSA verification, compressed P2PKH address
		a = append([]byte{0}, ah_c...)
		cs = bitcoin.Hash256(a)
		a = append(a, cs[:4]...)
		da = bitcoin.Base58Encode(a)
		if da == addr {
			return true
		}

		// (3) ECDSA verification, P2WPKH-P2SH compressed address
		rs := append([]byte{0x00, 0x14}, ah_c...)
		rh := bitcoin.Hash160(rs)
		a = append([]byte{5}, rh...)
		cs = bitcoin.Hash256(a)
		a = append(a, cs[:4]...)
		da = bitcoin.Base58Encode(a)
		if da == addr {
			return true
		}

		// (3) ECDSA verification, P2WPKH compressed address
		da = Bech32("bc", ah_c)
		if da == addr {
			return true
		}

		// (4) Schnorr verification, P2TR (Taproot) compressed address
		//     Not implemented yet...
		da = ""

		return false
	}()

	return
}

//----------------------------------------------------------------------
// helper functions
//----------------------------------------------------------------------

// recoverFromSignature returns the public key of the signing key used to
// generate the Bitcoin signature for a message.
func recoverFromSignature(recId byte, sig *bitcoin.Signature, msg string, compr bool) (pk *bitcoin.PublicKey, ok bool) {

	// create message hash from formatted message
	hash := bitcoin.Hash256(formatMessageForSigning(msg))

	// reconstruct public key from signature
	n := bitcoin.GetCurve().N
	p := bitcoin.GetCurve().P

	x := sig.R
	if recId&2 != 0 {
		x = x.Add(n)
	}
	if x.Cmp(p) >= 0 {
		return
	}

	y, ok := bitcoin.Solve(x)
	if !ok {
		return
	}
	yp := y.Bit(0)
	if uint(recId&1) != yp {
		y = p.Sub(y)
	}
	R := bitcoin.NewPoint(x, y)
	if !R.Mult(n).IsInf() {
		return
	}

	e := bitcoin.ConvertHash(hash).Mod(n)
	ei := math.ZERO.Sub(e).Mod(n)
	ri := sig.R.ModInverse(n)
	s_ri := ri.Mul(sig.S).Mod(n)
	ei_ri := ri.Mul(ei).Mod(n)
	Q := bitcoin.MultBase(ei_ri).Add(R.Mult(s_ri))

	pk = &bitcoin.PublicKey{Q: Q, IsCompressed: compr}
	ok = true
	return
}

// formatMessageForSigning prepares a message to be signed
func formatMessageForSigning(msg string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(len(msgHdr)))
	buf.Write(msgHdr)
	mb := []byte(msg)
	buf.Write(encode(len(mb)))
	buf.Write(mb)
	return buf.Bytes()
}

// encode an integer into its smallest binary representation
func encode(n int) (i []byte) {
	buf := new(bytes.Buffer)
	if n < 0xFD {
		buf.WriteByte(byte(n % 0xFD))
	} else if n < 0xFFFF {
		buf.WriteByte(253)
		binary.Write(buf, binary.LittleEndian, uint16(n))
	} else if n < 0xFFFFFFFF {
		buf.WriteByte(254)
		binary.Write(buf, binary.LittleEndian, uint32(n))
	} else {
		buf.WriteByte(255)
		binary.Write(buf, binary.LittleEndian, uint64(n))
	}
	return buf.Bytes()
}
