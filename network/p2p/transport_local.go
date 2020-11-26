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
	"net"
	"sync"
)

//======================================================================
// LocalTransport (used as parent for specific transport mechanism)
//======================================================================

// LocalAddress is the network address (net.Addr) of a local node.
type LocalAddress struct {
	Name string
}

// NewLocalAddress creates a new address with give name
func NewLocalAddress(name string) *LocalAddress {
	return &LocalAddress{
		Name: name,
	}
}

// Network returns the network label of the address
func (a *LocalAddress) Network() string {
	return "local"
}

// String returns the human-rwadable network address
func (a *LocalAddress) String() string {
	return a.Name
}

//----------------------------------------------------------------------
// Connector
//----------------------------------------------------------------------

// LocalConnector is used in LocalTransport
type LocalConnector struct {
	trans *LocalTransport

	cache map[string]string
	lock  sync.Mutex
}

// NewAddress returns a new network address for the transport based on an
// endpoint specification.
func (c *LocalConnector) NewAddress(endp string) (net.Addr, error) {
	return NewLocalAddress(endp), nil
}

// Sample returns a random collection of node/network address pairs this node
// has learned during up-time.
func (c *LocalConnector) Sample(num int, skip *Address) []*Address {
	return nil
}

// Send a message locally.
func (c *LocalConnector) Send(ctx context.Context, dst net.Addr, pkt *Packet) error {
	// resolve node for endpoint
	switch dst.(type) {
	case *LocalAddress:
		// the type we are looking for...
	default:
		return ErrTransAddressInvalid
	}
	addr := dst.String()
	var node *Node = nil
	for id, endp := range c.cache {
		if endp == addr {
			node = c.trans.getNode(id)
		}
	}
	// send message to receiver
	if node != nil {
		// send the message to receiver
		msg, err := node.Unwrap(pkt)
		if err != nil {
			return err
		}
		go func() {
			node.Handle() <- msg
		}()
		return nil
	}
	return ErrTransUnknownReceiver
}

// Listen to messages from "outside" not necessary in local transport
func (c *LocalConnector) Listen(ctx context.Context, ch chan Message) {
}

// Learn network address of node address
func (c *LocalConnector) Learn(addr *Address, endp net.Addr) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := addr.String()
	c.cache[key] = endp.String()
	return nil
}

// Resolve node address into a network address
func (c *LocalConnector) Resolve(addr *Address) net.Addr {
	c.lock.Lock()
	defer c.lock.Unlock()

	netw, ok := c.cache[addr.String()]
	if !ok {
		return nil
	}
	return NewLocalAddress(netw)
}

//----------------------------------------------------------------------
// Transport implementation
//----------------------------------------------------------------------

// LocalTransport handles all common transport functionality and is able
// to route messages to local nodes
type LocalTransport struct {

	// map of known endpoints
	nodes map[string]*Node
	lock  sync.Mutex
}

// NewLocalTransport instantiates a local transport implementation
func NewLocalTransport() LocalTransport {
	return LocalTransport{
		nodes: make(map[string]*Node),
	}
}

// Register a node for participation in the transport layer.
func (t *LocalTransport) Register(ctx context.Context, n *Node, endp string) error {
	// synchronize access to node list
	t.lock.Lock()
	defer t.lock.Unlock()

	// check for already registered address
	addr := n.Address().String()
	if _, ok := t.nodes[addr]; ok {
		return ErrTransAddressDup
	}
	t.nodes[addr] = n

	// create a connector for node
	// assemble suitable connector
	conn := &LocalConnector{
		trans: t,
		cache: make(map[string]string),
	}
	n.Connect(conn)
	return nil
}

// get the node associated with given address
func (t *LocalTransport) getNode(addr string) *Node {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.nodes[addr]
}
