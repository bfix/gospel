package dht

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
	"encoding/hex"
	"testing"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/data"
)

func TestPacket(t *testing.T) {

	//------------------------------------------------------------------
	// (1) Prepare
	//------------------------------------------------------------------

	// sender S
	pubS, prvS := ed25519.NewKeypair()
	addrS := NewAddressFromKey(pubS)

	// receiver R
	pubR, prvR := ed25519.NewKeypair()
	addrR := NewAddressFromKey(pubR)

	// data to be transfered
	msg_out := &MsgHeader{
		Size:     HDR_SIZE,
		TxId:     23,
		Type:     PING,
		Receiver: addrR,
		Sender:   addrS,
	}
	buf_out, err := data.Marshal(msg_out)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("send message size is %d bytes\n", len(buf_out))
	t.Logf("=> %s\n", hex.EncodeToString(buf_out))

	//------------------------------------------------------------------
	// (2) Create packet with encrypted message
	//------------------------------------------------------------------

	pkt_out, err := NewPacketFromData(buf_out, prvS, pubR)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("packet size is %d bytes\n", 34+len(pkt_out.Body))

	//------------------------------------------------------------------
	// (3) Wire transfer
	//------------------------------------------------------------------

	wire, err := data.Marshal(pkt_out)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pkt buffer size is %d bytes\n", len(wire))
	t.Logf("=> %s\n", hex.EncodeToString(wire))

	//------------------------------------------------------------------
	// (4) Reconstruct packet
	//------------------------------------------------------------------

	pkt_in := new(Packet)
	if err := data.Unmarshal(pkt_in, wire); err != nil {
		t.Fatal(err)
	}

	//------------------------------------------------------------------
	// (5) Decrypt message from packet
	//------------------------------------------------------------------

	msg_in, err := pkt_in.Unwrap(prvR)
	if err != nil {
		t.Fatal(err)
	}
	buf_in, err := data.Marshal(msg_in)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("received message size is %d bytes\n", len(buf_in))

	//------------------------------------------------------------------
	// (6) Verify message
	//------------------------------------------------------------------

	if !bytes.Equal(buf_out, buf_in) {
		t.Fatal("Message mismatch")
	}
}
