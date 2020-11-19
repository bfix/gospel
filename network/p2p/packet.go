package p2p

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
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/data"
	"github.com/bfix/gospel/math"
	chacha "golang.org/x/crypto/chacha20poly1305"
)

// Error messages
var (
	ErrPacketSenderMismatch = fmt.Errorf("Sender not matching message header")
	ErrPacketIntegrity      = fmt.Errorf("Packet integrity violated")
)

//----------------------------------------------------------------------
// Packet is the unit of transfer in a network. It wraps messages
// so they are encrypted and authenticated during transit; only
// the sending and the receiving peers can read the message. A new
// encryption key is generated for each Packet.
//
// The sending and receiving peers (with indices 's' and 'r') both have
// a public/private Ed25519 key pair: 'd_s' and 'd_r' are the private
// keys, 'P_s = [d_s]G' and 'P_r = [d_r]G' are the public keys (with G
// as the base point of the Ed25519 group). The public keys are mutually
// known to the peers.
//
// Encryption:
// ===========
// (1) The sender generates the SHA256 hash of the message to be wrapped
//     into a packet and derives a value 'r' as 'r = SHA256(m) mod N'
//     where 'N' is the group order of Ed25519.
// (2) The sender computes 'Q = [r*d_s]P_r' (=[r*d_s*d_r]G) as a shared
//     secret and derives a symmetric encryption key from it.
// (3) The message is encrypted and stored in the 'Body' field of the
//     packet; the length of the encrypted message is the same as the
//     plain message.
// (4) The sender computes 'KXT = [r]P_s' as the key exchange token and
//     stores it in the 'KXT' field of the packet.
//
// Decryption:
// ===========
// (1) The receiver computes 'Q = [d_r]*KXT' to get the shared secret
//     and to derive a symmetric decryption key from it. The 'Body'
//     is decrypted.
// (2) The receiver computes the SHA256 hash value of the decrypted
//     body and derives a value 'r = SHA256(Body) mod N'
// (3) The receiver verifies that the equation '[r]P_s == KXT' holds to
//     verify the integrity of the packet. 'P_s' is part of the message
//     header as defined in this framework. The receiver now knows that
//     the (plain text) message is originating from the sender.
//
//----------------------------------------------------------------------

// Packet data structure
type Packet struct {
	Size uint16 `order:"big"` // size of packet (including this field)
	KXT  []byte `size:"32"`   // Key Exchange Token
	Body []byte `size:"*"`    // encrypted body
}

// NewPacket creates a new packet from a message.
func NewPacket(msg Message, sender *Node) (*Packet, error) {
	// check if sender is correctly specified in message
	hdr := msg.Header()
	if !sender.Address().Equals(hdr.Sender) {
		return nil, ErrPacketSenderMismatch
	}
	// convert message to binary object
	buf, err := data.Marshal(msg)
	if err != nil {
		return nil, err
	}
	// get keys from peers
	skey := sender.PrivateKey()
	rkey := hdr.Receiver.PublicKey()
	return NewPacketFromData(buf, skey, rkey)
}

// NewPacketFromData creates a new packet from a binary object.
func NewPacketFromData(buf []byte, sender *ed25519.PrivateKey, receiver *ed25519.PublicKey) (*Packet, error) {

	// compute 'r = SHA256(b) mod N'
	rb := sha256.Sum256(buf)
	r := math.NewIntFromBytes(rb[:])
	r = r.Mod(ed25519.GetCurve().N)

	// compute shared secret and derive encryption key
	Q := receiver.Mult(r.Mul(sender.D))

	// encrypt body with ChaCha20-Poly1305 AEAD
	aead, err := chacha.New(Q.Bytes())
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(buf)+aead.Overhead())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	enc := aead.Seal(nonce, nonce, buf, nil)

	// assemble packet.
	pubS := sender.Public()
	return &Packet{
		Size: uint16(34 + len(enc)),
		KXT:  pubS.Mult(r).Bytes(),
		Body: enc,
	}, nil
}

// Unwrap a packet
func (p *Packet) Unwrap(receiver *ed25519.PrivateKey, mf MessageFactory) (Message, error) {

	// compute shared secret and derive encryption key
	Q := ed25519.NewPublicKeyFromBytes(p.KXT).Mult(receiver.D)

	// decrypt with ChaCha20-Poly1305 AEAD
	aead, err := chacha.New(Q.Bytes())
	if err != nil {
		return nil, err
	}
	nonce := p.Body[:aead.NonceSize()]
	enc := p.Body[aead.NonceSize():]
	body, err := aead.Open(nil, nonce, enc, nil)
	if err != nil {
		return nil, err
	}

	// compute 'r = SHA256(b) mod N'
	rb := sha256.Sum256(body)
	r := math.NewIntFromBytes(rb[:])
	r = r.Mod(ed25519.GetCurve().N)

	// reconstruct message
	msg, err := mf(body)
	if err != nil {
		return nil, err
	}

	// check message integrity
	k := msg.Header().Sender.PublicKey().Mult(r).Bytes()
	if !bytes.Equal(k, p.KXT) {
		return nil, ErrPacketIntegrity
	}
	// return decrypted message
	return msg, nil
}
