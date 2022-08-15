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
	return fmt.Sprintf("PING{%.8s -> %.8s, #%d}", m.Sender, m.Receiver, m.TxID)
}

// NewPingMsg creates an empty PING request
func NewPingMsg() Message {
	return &PingMsg{
		MsgHeader: MsgHeader{
			Size:     HdrSize,
			TxID:     0,
			Type:     ReqPING,
			Flags:    0,
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
	return fmt.Sprintf("PONG{%.8s -> %.8s, #%d}", m.Sender, m.Receiver, m.TxID)
}

// NewPongMsg creates an empty PONG response
func NewPongMsg() Message {
	return &PongMsg{
		MsgHeader: MsgHeader{
			Size:     HdrSize,
			TxID:     0,
			Type:     RespPING,
			Flags:    0,
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
	srv.factories[ReqPING] = NewPingMsg
	srv.factories[RespPING] = NewPongMsg

	// defined known labels
	srv.labels[ReqPING] = "PING"
	srv.labels[RespPING] = "PONG"
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
	if hdr.Type != ReqPING {
		return false, nil
	}
	// assemble PONG as response to PING
	resp, _ := NewPongMsg().(*PongMsg)
	resp.TxID = hdr.TxID
	resp.Sender = hdr.Receiver
	resp.Receiver = hdr.Sender

	// send message
	return true, s.Send(ctx, resp)
}

// NewMessage creates an empty service message of given type
func (s *PingService) NewMessage(mt int) Message {
	switch mt {
	case ReqPING:
		return NewPingMsg()
	case RespPING:
		return NewPongMsg()
	}
	return nil
}

//----------------------------------------------------------------------
// PING task
//----------------------------------------------------------------------

// Ping sends a ping to another node and waits for a response (with timeout)
func (s *PingService) Ping(ctx context.Context, rcv *Address, timeout time.Duration, relays int) error {
	// assemble request
	req, _ := NewPingMsg().(*PingMsg)
	req.TxID = s.node.NextID()
	req.Sender = s.node.Address()
	req.Receiver = rcv

	// check for relayed message
	var msg Message = req
	if relays > 0 {
		// select hops for relay chain
		var err error
		hops := s.Node().Sample(relays, rcv)
		if hops != nil {
			// assemble relay message (chain)
			if msg, err = s.Node().RelayedMessage(msg, hops); err != nil {
				return err
			}
			// set transaction id of final request in outer relay message
			msg.Header().TxID = req.TxID
		}
	}

	// send request and process responses
	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) (bool, error) {
			// handle PING response (update node list etc.)
			return true, nil
		},
		timeout: timeout,
	}
	return s.Task(ctx, msg, hdlr)
}
