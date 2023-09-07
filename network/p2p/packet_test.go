package p2p

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

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/data"
)

type TestMsg struct {
	MsgHeader
}

func (m *TestMsg) String() string {
	return "TEST"
}

func TestPacket(t *testing.T) {

	NewMessage := func(buf []byte) (Message, error) {
		msg := new(TestMsg)
		if err := data.Unmarshal(msg, buf); err != nil {
			return nil, err
		}
		return msg, nil
	}

	//------------------------------------------------------------------
	// (1) Prepare
	//------------------------------------------------------------------

	// sender S
	pubS, prvS := ed25519.NewKeypair()
	addrS := NewAddressFromKey(pubS)

	// receiver R
	pubR, prvR := ed25519.NewKeypair()
	addrR := NewAddressFromKey(pubR)

	// data to be transferred
	msgOut := &TestMsg{
		MsgHeader: MsgHeader{
			Size:     HdrSize,
			TxID:     23,
			Type:     ReqPING,
			Receiver: addrR,
			Sender:   addrS,
		},
	}
	bufOut, err := data.Marshal(msgOut)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("send message size is %d bytes\n", len(bufOut))
	t.Logf("=> %s\n", hex.EncodeToString(bufOut))

	//------------------------------------------------------------------
	// (2) Create packet with encrypted message
	//------------------------------------------------------------------

	pktOut, err := NewPacketFromData(bufOut, prvS, pubR)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("packet size is %d bytes\n", 34+len(pktOut.Body))

	//------------------------------------------------------------------
	// (3) Wire transfer
	//------------------------------------------------------------------

	wire, err := data.Marshal(pktOut)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pkt buffer size is %d bytes\n", len(wire))
	t.Logf("=> %s\n", hex.EncodeToString(wire))

	//------------------------------------------------------------------
	// (4) Reconstruct packet
	//------------------------------------------------------------------

	pktIn := new(Packet)
	if err = data.Unmarshal(pktIn, wire); err != nil {
		t.Fatal(err)
	}

	//------------------------------------------------------------------
	// (5) Decrypt message from packet
	//------------------------------------------------------------------

	msgIn, err := pktIn.Unwrap(prvR, NewMessage)
	if err != nil {
		t.Fatal(err)
	}
	bufIn, err := data.Marshal(msgIn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("received message size is %d bytes\n", len(bufIn))

	//------------------------------------------------------------------
	// (6) Verify message
	//------------------------------------------------------------------

	if !bytes.Equal(bufOut, bufIn) {
		t.Fatal("Message mismatch")
	}
}
