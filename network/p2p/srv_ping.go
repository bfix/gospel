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
	"context"
	"fmt"
	"time"
)

//----------------------------------------------------------------------
// PING message (request)
//----------------------------------------------------------------------

// PingMsg for PING requests
type PingMsg struct {
	MsgHeader
}

// String returns human-readable message
func (m *PingMsg) String() string {
	return fmt.Sprintf("PING{%.8s -> %.8s, #%d}", m.Sender, m.Receiver, m.TxId)
}

// NewPingMsg creates an empty PING request
func NewPingMsg() Message {
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

// PongMsg for PONG responses
type PongMsg struct {
	MsgHeader
}

// String returns human-readable message
func (m *PongMsg) String() string {
	return fmt.Sprintf("PONG{%.8s -> %.8s, #%d}", m.Sender, m.Receiver, m.TxId)
}

// NewPongMsg creates an empty PONG response
func NewPongMsg() Message {
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

// PingService responds to PING requests from remote peer
type PingService struct {
	ServiceImpl
}

// NewPingService creates a new service instance
func NewPingService() *PingService {
	srv := &PingService{
		ServiceImpl: *NewServiceImpl(),
	}
	// defined message instantiators
	srv.factories[PING] = NewPingMsg
	srv.factories[PONG] = NewPongMsg

	// defined known labels
	srv.labels[PING] = "PING"
	srv.labels[PONG] = "PONG"
	return srv
}

// Name is a human-readble and short service description like "PING"
func (s *PingService) Name() string {
	return "ping"
}

// Respond to a service request from peer.
func (s *PingService) Respond(ctx context.Context, m Message) (bool, error) {
	// check we are responsible for this
	hdr := m.Header()
	if hdr.Type != PONG {
		return false, nil
	}
	// assemble PONG as response to PING
	resp := NewPongMsg().(*PongMsg)
	resp.TxId = hdr.TxId
	resp.Sender = hdr.Receiver
	resp.Receiver = hdr.Sender

	// send message
	if err := s.Send(ctx, resp); err != nil {
		return true, err
	}
	return true, nil
}

// NewMessage creates an empty service message of given type
func (s *PingService) NewMessage(mt int) Message {
	switch mt {
	case PING:
		return NewPingMsg()
	case PONG:
		return NewPongMsg()
	}
	return nil
}

//----------------------------------------------------------------------
// PING task
//----------------------------------------------------------------------

// Ping sends a ping to another node and waits for a response (with timeout)
func (s *PingService) Ping(ctx context.Context, rcv *Address, timeout time.Duration) error {
	// assemble request
	req := NewPingMsg().(*PingMsg)
	req.TxId = uint32(s.node.NextId())
	req.Sender = s.node.Address()
	req.Receiver = rcv

	// send request and process responses
	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) (bool, error) {
			// handle PING response (update node list etc.)
			return true, nil
		},
		timeout: timeout,
	}
	return s.Task(ctx, req, hdlr)
}
