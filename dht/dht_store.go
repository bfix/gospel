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
	"context"
	"crypto/sha256"
	"time"
)

//----------------------------------------------------------------------
// Value (object that is stored in the DHT under a key)
//----------------------------------------------------------------------

// Value encapsulates data representing a binary object of varying size (up
// to 2^16 = 65536 bytes)
type Value struct {
	Size uint16 `order:"big"`
	Data []byte `size:"Size"`
}

// Address returns the identifier for a value object
func (v *Value) Address() *Address {
	h := sha256.New()
	h.Write(v.Data)
	return NewAddress(h.Sum(nil))
}

//----------------------------------------------------------------------
// STORE message (request)
//----------------------------------------------------------------------

// StoreMsg for STORE requests
type StoreMsg struct {
	MsgHeader

	Key *Address // DHT key for data
	Val *Value   // DHT data blob
}

//
func (m *StoreMsg) Set(addr *Address, value *Value) {
	m.Key = addr
	m.Val = value
	m.Size = HDR_SIZE + 34 + value.Size
}

func NewStoreMsg() *StoreMsg {
	return &StoreMsg{
		MsgHeader: MsgHeader{
			Size:     HDR_SIZE + 32, // size of message with empty value
			TxId:     0,
			Type:     PING,
			Sender:   nil,
			Receiver: nil,
		},
		Key: NewAddress(nil),
		Val: nil,
	}
}

//----------------------------------------------------------------------
// STORE_RESP message (response)
//----------------------------------------------------------------------

type StoreRespMsg struct {
	MsgHeader
}

func NewStoreRespMsg() *PongMsg {
	return &PongMsg{
		MsgHeader: MsgHeader{
			Size:     HDR_SIZE,
			TxId:     0,
			Type:     PONG,
			Sender:   nil,
			Receiver: nil,
		},
	}
}

//----------------------------------------------------------------------
// STORE service
//----------------------------------------------------------------------

// StoreService responds to requests to store value under given key.
func (n *LocalNode) StoreService(ctx context.Context, m Message) bool {
	// assemble response
	hdr := m.Header()
	resp := NewStoreRespMsg()
	resp.TxId = hdr.TxId
	resp.Sender = hdr.Receiver
	resp.Receiver = hdr.Sender
	// store data
	panic("not implemented")

	// send message
	if err := n.conn.Send(ctx, resp); err != nil {
		return false
	}
	return true
}

//----------------------------------------------------------------------
// STORE task
//----------------------------------------------------------------------

// Store data in the DHT
func (n *LocalNode) StoreTask(ctx context.Context, rcv, key *Address, val *Value, timeout time.Duration) error {
	// assemble request
	req := NewPingMsg()
	req.TxId = n.nextId()
	req.Sender = n.addr
	req.Receiver = rcv

	// send request and process responses
	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) bool {
			// handle STORE responses
			panic("not implemented")
			return true
		},
		timeout: timeout,
	}
	return n.Task(ctx, req, hdlr)
}
