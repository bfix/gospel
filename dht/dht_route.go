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

type FindNodeMsg struct {
	MsgHeader

	Addr *Address // address of node we are looking for
}

func NewFindNodeMsg() *FindNodeMsg {
	return &FindNodeMsg{
		MsgHeader: MsgHeader{
			Size:     uint16(HDR_SIZE + ADDR_SIZE),
			TxId:     0,
			Type:     FIND_NODE,
			Sender:   nil,
			Receiver: nil,
		},
		Addr: NewAddress(nil),
	}
}

//----------------------------------------------------------------------
// FIND_NODE_RESP message (response)
//----------------------------------------------------------------------

type FindNodeRespMsg struct {
	MsgHeader

	Addr *Address
	Endp string
}

func NewFindNodeRespMsg() *FindNodeRespMsg {
	return &FindNodeRespMsg{
		MsgHeader: MsgHeader{
			Size:     uint16(HDR_SIZE + ADDR_SIZE),
			TxId:     0,
			Type:     FIND_NODE_RESP,
			Sender:   nil,
			Receiver: nil,
		},
		Addr: NewAddress(nil),
		Endp: "",
	}
}

//----------------------------------------------------------------------
// FIND_NODE service
//----------------------------------------------------------------------

// FindNodeService responds to FIND_NODE requests from remote peers.
func (n *LocalNode) FindNodeService(ctx context.Context, m Message) bool {
	// assemble FIND_NODE_RESP message
	hdr := m.Header()
	resp := NewPongMsg()
	resp.TxId = hdr.TxId
	resp.Sender = hdr.Receiver
	resp.Receiver = hdr.Sender

	// send message
	if err := n.conn.Send(ctx, resp); err != nil {
		return false
	}
	return true
}

//----------------------------------------------------------------------
// FIND_NODE task
//----------------------------------------------------------------------

// FindTask
func (n *LocalNode) FindNodeTask(ctx context.Context, addr *Address, timeout time.Duration) error {
	// assemble request
	req := NewPingMsg()
	req.TxId = n.nextId()
	req.Sender = n.addr

	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) bool {
			// handle PING response (update node list etc.)
			return true
		},
		timeout: timeout,
	}
	return n.Task(ctx, nil, req, hdlr)
}
