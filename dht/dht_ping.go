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
// PING message (request)
//----------------------------------------------------------------------

type PingMsg struct {
	MsgHeader
}

func NewPingMsg() *PingMsg {
	return &PingMsg{
		MsgHeader: MsgHeader{
			Size:     HDR_SIZE,
			TxId:     0,
			Type:     PING,
			Sender:   nil,
			Receiver: nil,
		},
	}
}

//----------------------------------------------------------------------
// PONG message (response)
//----------------------------------------------------------------------

type PongMsg struct {
	MsgHeader
}

func NewPongMsg() *PongMsg {
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
// PING service
//----------------------------------------------------------------------

// PingService responds to PING requests from remote peers.
func (n *LocalNode) PingService(ctx context.Context, m Message) bool {
	// assemble PONG as response to PING
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
// PING task
//----------------------------------------------------------------------

// Ping a node
func (n *LocalNode) PingTask(ctx context.Context, o Node, timeout time.Duration) error {
	// assemble request
	req := NewPingMsg()
	req.TxId = n.nextId()
	req.Sender = n.addr
	req.Receiver = o.Address()

	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) bool {
			// handle PING response (update node list etc.)
			return true
		},
		timeout: timeout,
	}
	return n.Task(ctx, o, req, hdlr)
}
