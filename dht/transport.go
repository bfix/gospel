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
	"fmt"
	"log"
	"sync"
)

// Error codes
var (
	ErrTransAddressDup      = fmt.Errorf("Address already registered")
	ErrTransRemote          = fmt.Errorf("No remote nodes allowed")
	ErrTransUnknownReceiver = fmt.Errorf("Unknown receiver")
)

//======================================================================
// Transport layer abstraction
//======================================================================

// Connector can send and receive message over a transport layer
type Connector interface {
	// Send message over transport
	Send(context.Context, Message) error

	// Listen to messages from transport
	Listen(context.Context, chan Message)
}

// Transport abstraction: Every endpoint (on a local machine) registers
// with its address and receives channels for communication (incoming
// and outgoing messages). The transfer process needs to be started
// with the 'Run()' method for its message pump to work.
type Transport interface {
	// Register a node for participation in this transport
	Register(context.Context, Node, string) error
}

//======================================================================
// LocalTransport (used as parent for specific transport mechanism)
//======================================================================

//----------------------------------------------------------------------
// Connector
//----------------------------------------------------------------------

// LocalConnector is used in LocalTransport
type LocalConnector struct {
	trans *LocalTransport
}

// Send a message locally.
func (c *LocalConnector) Send(ctx context.Context, msg Message) error {
	// send message to receiver
	hdr := msg.Header()
	if node, ok := c.trans.nodes[hdr.Receiver.String()]; ok {
		// send the message to receiver
		go func() {
			log.Printf("[%.8s] Sent %s to '%.8s'\n", hdr.Sender, MsgType(msg), hdr.Receiver)
			node.Handle() <- msg
		}()
		return nil
	}
	return ErrTransUnknownReceiver
}

// Listen to messages from "outside" not necessary in local transport
func (c *LocalConnector) Listen(ctx context.Context, ch chan Message) {
}

//----------------------------------------------------------------------
// Transport implementation
//----------------------------------------------------------------------

// LocalTransport handles all common transport functionality and is able
// to route messages to local nodes
type LocalTransport struct {

	// map of known endpoints
	nodes map[string]Node
	lock  sync.Mutex
}

// NewLocalTransport instantiates a local transport implementation
func NewLocalTransport() LocalTransport {
	return LocalTransport{
		nodes: make(map[string]Node),
	}
}

// Register a node for participation in the transport layer.
func (t *LocalTransport) Register(ctx context.Context, n Node, endp string) error {
	// synchronize access to node list
	t.lock.Lock()
	defer t.lock.Unlock()

	if n.IsRemote() {
		// no remote peers allowed in local transport
		return ErrTransRemote
	}
	// check for already registered address
	addr := n.Address().String()
	if _, ok := t.nodes[addr]; ok {
		return ErrTransAddressDup
	}
	t.nodes[addr] = n

	// create a connector for node
	n.Connect(&LocalConnector{t})
	return nil
}
