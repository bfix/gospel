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
)

//----------------------------------------------------------------------
// RELAY message
//----------------------------------------------------------------------

// RelayMsg is used to forward packets to nodes. The forwarded packet can
// itself be a RelayMsg, thus allowing a nested relay path (onion-like)
type RelayMsg struct {
	MsgHeader

	NextHop *Endpoint // next hop address
	Pkt     *Packet   // packet to deliver to next hop
}

// String returns human-readable message
func (m *RelayMsg) String() string {
	return fmt.Sprintf("RELAY{%.8s -> %.8s => %.8s}", m.Sender, m.Receiver, m.NextHop)
}

// NewRelayMsg creates an empty forward message
func NewRelayMsg() Message {
	return &RelayMsg{
		MsgHeader: MsgHeader{
			Size:     HDR_SIZE,
			TxId:     0,
			Type:     RELAY,
			Flags:    0,
			Sender:   nil,
			Receiver: nil,
		},
		NextHop: nil,
		Pkt:     nil,
	}
}

// Set the forward parameters
func (m *RelayMsg) Set(e *Endpoint, pkt *Packet) {
	m.NextHop = e
	m.Pkt = pkt
	m.Size += e.Size() + pkt.Size
}

//----------------------------------------------------------------------
// Relay service.
//----------------------------------------------------------------------

// RelayService to forward messages to other nodes.
type RelayService struct {
	ServiceImpl
}

// NewRelayService creates a new service instance
func NewRelayService() *RelayService {
	srv := &RelayService{
		ServiceImpl: *NewServiceImpl(),
	}
	// defined message instantiators
	srv.factories[RELAY] = NewRelayMsg

	// defined known labels
	srv.labels[RELAY] = "RELAY"
	return srv
}

// Name is a human-readble and short service description like "PING"
func (s *RelayService) Name() string {
	return "relay"
}

// Relay message to next hop
func (s *RelayService) Relay(ctx context.Context, msg *RelayMsg) error {

	return nil
}

// Respond to a service request from peer.
func (s *RelayService) Respond(ctx context.Context, m Message) (bool, error) {
	// check we are responsible for this
	hdr := m.Header()
	if hdr.Type != RELAY {
		return false, nil
	}
	// cast will succeed because type of message is checked
	msg := m.(*RelayMsg)

	// check if this is a relay to ourself
	if msg.NextHop.Addr.Equals(s.Node().Address()) {
		// Unwrap the packet
		inMsg, err := s.Node().Unwrap(msg.Pkt)
		if err == nil {
			hdr := msg.Header()
			// relay request endpoint is the network address of the sender
			s.Node().Learn(hdr.Sender, msg.NextHop.Endp.String())
			// process unwrapped message
			go func() {
				s.Node().Handle() <- inMsg
			}()
		}
		return true, err
	}

	// resolve receiver
	netw := s.Node().Resolve(msg.NextHop.Addr)
	var err error = nil
	if netw == nil {
		// cache miss: try provided network address
		endp := msg.NextHop.Endp.String()
		netw, err = s.Node().NewNetworkAddr(endp)
	}
	if err == nil {
		s.Node().SendRaw(ctx, netw, msg.Pkt)
	}
	return true, err
}

// NewMessage creates an empty service message of given type
func (s *RelayService) NewMessage(mt int) Message {
	switch mt {
	case RELAY:
		return NewRelayMsg()
	}
	return nil
}
