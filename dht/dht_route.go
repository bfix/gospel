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
	"time"
)

//----------------------------------------------------------------------
// FIND_NODE message (request)
//----------------------------------------------------------------------

// FindNodeMsg for FIND_NODE requests
type FindNodeMsg struct {
	MsgHeader

	Addr *Address // address of node we are looking for
}

// NewFindNodeMsg creates an empty FIND_NODE message
func NewFindNodeMsg() *FindNodeMsg {
	return &FindNodeMsg{
		MsgHeader: MsgHeader{
			Size:     uint16(HDR_SIZE + ADDR_SIZE),
			TxId:     0,
			Type:     FIND_NODE,
			Sender:   nil,
			Receiver: nil,
		},
		Addr: nil,
	}
}

// Set the additional address field
func (m *FindNodeMsg) Set(addr *Address) *FindNodeMsg {
	m.Addr = addr
	return m
}

//----------------------------------------------------------------------
// FIND_NODE_RESP message (response)
//----------------------------------------------------------------------

// FindNodeRespMsg for FIND_NODE responses
type FindNodeRespMsg struct {
	MsgHeader

	Addr *Address // address query
	Endp *String  // resolved network endpoint
}

// NewFindNodeRespMsg creates an empty FIND_NODE response
func NewFindNodeRespMsg() *FindNodeRespMsg {
	return &FindNodeRespMsg{
		MsgHeader: MsgHeader{
			Size:     uint16(HDR_SIZE + ADDR_SIZE),
			TxId:     0,
			Type:     FIND_NODE_RESP,
			Sender:   nil,
			Receiver: nil,
		},
		Addr: nil,
		Endp: nil,
	}
}

// Set the additional fields (address and enspoint)
func (m *FindNodeRespMsg) Set(addr *Address, endp string) *FindNodeRespMsg {
	m.Addr = addr
	m.Endp = NewString(endp)
	m.Size += m.Endp.Size()
	return m
}

//----------------------------------------------------------------------
// FIND_NODE service
//----------------------------------------------------------------------

// FindNodeService responds to FIND_NODE requests from remote peers.
func (n *LocalNode) FindNodeService(ctx context.Context, m Message) bool {
	switch msg := m.(type) {
	case *FindNodeMsg:
		// find endpoint assoicated with address or closest node in
		// our routing table
		panic("not implemented")
		var addr *Address = nil
		var endp string = ""

		// assemble FIND_NODE_RESP message
		hdr := msg.Header()
		resp := NewFindNodeRespMsg().Set(addr, endp)
		resp.TxId = hdr.TxId
		resp.Sender = hdr.Receiver
		resp.Receiver = hdr.Sender

		// send message
		if err := n.conn.Send(ctx, resp); err != nil {
			return false
		}
		return true
	default:
		return false
	}
}

//----------------------------------------------------------------------
// FIND_NODE task
//----------------------------------------------------------------------

// FindNodeTask is used to lookup a node endpoint address.
func (n *LocalNode) FindNodeTask(ctx context.Context, rcv, node *Address, timeout time.Duration) error {
	// assemble request
	req := NewFindNodeMsg()
	req.TxId = n.nextId()
	req.Sender = n.addr
	req.Receiver = rcv
	req.Addr = node

	// send request and process responses
	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) bool {
			// handle FIND_NODE response
			panic("not implemented")
			return true
		},
		timeout: timeout,
	}
	return n.Task(ctx, req, hdlr)
}
