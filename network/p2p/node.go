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
	"log"
	"net"
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
)

// Error codes
var (
	ErrNodeTimeout        = fmt.Errorf("Operation timed out...")
	ErrNodeRemote         = fmt.Errorf("No remote nodes allowed")
	ErrNodeConnectorSet   = fmt.Errorf("Connector for node already set")
	ErrNodeSendNoReceiver = fmt.Errorf("Send has no recipient")
)

// constants
var (
	NODE_TICK = 1 * time.Minute
)

//======================================================================
// Network peer:
// A peer participating in the P2P network has a crypto address (public
// key in binary representation); all that peers need to know about
// other peers is their public key address.
// On the transport level a remote peer also needs to have a network
// address but that is managed on the transport layer, not inside a
// node implementation.
//======================================================================

// Node represents a local network peer
type Node struct {
	prvKey *ed25519.PrivateKey // Private Ed25519 key
	pubKey *ed25519.PublicKey  // node public key
	addr   *Address            // address of the node in the P2P network

	inCh chan Message // channel for incoming messages
	conn Connector    // send/receive stub

	buckets *BucketList  // routing table
	srvcs   *ServiceList // list of services

	lastID int64 // last used identifier
}

// NewNode instantiates a new local node with given private key.
func NewNode(prv *ed25519.PrivateKey) (n *Node, err error) {
	// create new instance of a local node
	pub := prv.Public()
	n = &Node{
		prvKey:  prv,
		pubKey:  pub,
		addr:    NewAddressFromKey(pub),
		inCh:    make(chan Message),
		conn:    nil,
		buckets: NewBucketList(),
		srvcs:   NewServiceList(),
	}
	// add all standard services (P2P)
	n.AddService(NewPingService())
	n.AddService(NewLookupService())
	return
}

//----------------------------------------------------------------------
// Address and key handling
//----------------------------------------------------------------------

// Address returns the P2P address of a node
func (n *Node) Address() *Address {
	return n.addr
}

// PublicKey returns the Ed25519 key from the node address
func (n *Node) PublicKey() *ed25519.PublicKey {
	return n.pubKey
}

// PrivateKey returns the private Ed25519 key for the node
func (n *Node) PrivateKey() *ed25519.PrivateKey {
	return n.prvKey
}

//----------------------------------------------------------------------
// Service handling
//----------------------------------------------------------------------

// AddService to add a new service running on the node.
func (n *Node) AddService(s Service) bool {
	if n.srvcs.Add(s) {
		s.link(n)
		log.Printf("[%.8s] Service '%s' added to node\n", n.addr, s.Name())
		return true
	}
	log.Printf("[%.8s] Failed to register service '%s' on node\n", n.addr, s.Name())
	return false
}

// Service returns the named service instance for node
func (n *Node) Service(name string) Service {
	return n.srvcs.Get(name)
}

//----------------------------------------------------------------------
// Message exchange (incoming and outgoing messages)
//----------------------------------------------------------------------

// Send a message from this node to a peer on the network
func (n *Node) Send(ctx context.Context, msg Message) error {
	return n.conn.Send(ctx, msg)
}

// Handle messages from channel
func (n *Node) Handle() chan Message {
	return n.inCh
}

// Connect to the transport network with given connector.
func (n *Node) Connect(c Connector) error {
	if n.conn != nil {
		return ErrNodeConnectorSet
	}
	n.conn = c
	return nil
}

//----------------------------------------------------------------------
// Endpoint management (mapping node addresses to network addresses)
//----------------------------------------------------------------------

// Learn about a new peer in the network
func (n *Node) Learn(addr *Address, endp string) (err error) {
	// compute the distance between node and addr
	k := addr.Distance(n.Address()).BitLen() - 1
	if k < 0 {
		// no need to learn our own address :)
		return
	}
	// add peer to routing table
	n.buckets.Add(k, addr)
	// learn network endpoint if specified
	if len(endp) > 0 {
		// get the associated network address
		netw, err := net.ResolveUDPAddr("udp", endp)
		if err != nil {
			return err
		}
		err = n.conn.Learn(addr, netw)
	}
	return
}

// Resolve peer address to network address
func (n *Node) Resolve(addr *Address) string {
	netw := n.conn.Resolve(addr)
	if netw == nil {
		return ""
	}
	return netw.String()
}

//----------------------------------------------------------------------
// Run the node with services
//----------------------------------------------------------------------

// Run the local node
func (n *Node) Run(ctx context.Context) {

	// we do periodic jobs once every minute
	// and remember the epoch we are in
	epoch := 0
	period := time.Tick(NODE_TICK)

	// run bucket list processor
	n.buckets.Run(ctx)

	// run the connector listener
	n.conn.Listen(ctx, n.inCh)

	// operate message pump
	for {
		select {
		// process incoming message
		case msg := <-n.inCh:
			go func() {
				hdr := msg.Header()

				switch hdr.Type % 2 {
				//----------------------------------------------------------
				// Incoming request
				//----------------------------------------------------------
				case 1:
					// lookup service handling the request
					n.srvcs.Respond(ctx, msg)

				//----------------------------------------------------------
				// Incoming response
				//----------------------------------------------------------
				case 0:
					// lookup service listening to response
					n.srvcs.Listen(ctx, msg)
				}
			}()

		// periodic jobs
		case <-period:
			epoch++

		// externally cancelled
		case <-ctx.Done():
			return
		}
	}
}

//----------------------------------------------------------------------
// Helper methods
//----------------------------------------------------------------------

// Factory produces messages from a binary representation
func (n *Node) MessageFactory(buf []byte) (Message, error) {
	return n.srvcs.MessageFactory(buf)
}

// Closest returns the n closest nodes we know of
func (n *Node) Closest(num int) []*Address {
	return n.buckets.Closest(num)
}

// NextId returns the next unique identifier for this node context
func (n *Node) NextId() int64 {
	n.lastID++
	return n.lastID

}
