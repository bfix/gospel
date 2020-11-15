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
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
)

// Error codes
var (
	ErrNodeTimeout      = fmt.Errorf("Operation timed out...")
	ErrNodeConnectorSet = fmt.Errorf("Connector for node already set")
)

//======================================================================
// Node interface and implementation:s There are two implementations of
// the Node interface; one for a local node and one for remote nodes.
//======================================================================

// Node interface for nodes in the DHT network
type Node interface {
	IsRemote() bool                  // returns true if node is a remote peer
	Address() *Address               // DHT address of node
	Learn(*Address)                  // Learn about a new peer in the network
	PublicKey() *ed25519.PublicKey   // Node public key
	PrivateKey() *ed25519.PrivateKey // Node private key
	Handle() chan Message            // channel of incoming messages
	Connect(Connector) error         // Use a given connector for exchange
}

//----------------------------------------------------------------------
// Peer implements a Node interface and represents an actual (remote)
// participant in the P2P network.
//----------------------------------------------------------------------

// Peer is a node in the network
type Peer struct {
	// address of the node in the DHT
	addr   *Address           // node address
	pubKey *ed25519.PublicKey // node public key

	inCh chan Message // channel for incoming messages
	conn Connector    // send/receive stub
}

// NewPeer instantiates a new remote node
func NewPeer(pub *ed25519.PublicKey) *Peer {
	return &Peer{
		pubKey: pub,
		addr:   NewAddressFromKey(pub),
		inCh:   make(chan Message),
		conn:   nil,
	}
}

// IsRemote returns true if node is a remote peer
func (n *Peer) IsRemote() bool {
	return true
}

// Address returns the DHT address of a node
func (n *Peer) Address() *Address {
	return n.addr
}

// Learn about a new peer in the network
func (n *Peer) Learn(addr *Address) {
	// remote nodes stay dumb
}

// PublicKey returns the Ed25519 key from the node address
func (n *Peer) PublicKey() *ed25519.PublicKey {
	return n.pubKey
}

// PrivateKey returns the private Ed25519 key for the node
func (n *Peer) PrivateKey() *ed25519.PrivateKey {
	// we don't have the private key of (remote) peers
	return nil
}

// Handle messages from channel
func (n *Peer) Handle() chan Message {
	return n.inCh
}

// Connect to the transport network with given connector.
func (n *Peer) Connect(c Connector) error {
	if n.conn != nil {
		return ErrNodeConnectorSet
	}
	n.conn = c
	return nil
}

//----------------------------------------------------------------------
// LocalNode
//----------------------------------------------------------------------

// LocalNode represents an owned node running on a local machine
type LocalNode struct {
	Peer

	prvKey *ed25519.PrivateKey // Private Ed25519 key

	buckets *BucketList  // routing table
	hdlrs   *HandlerList // response handlers
	srvs    *HandlerList // request handlers (services)

	lastID uint32 // last used ID (for requests)
}

// NewLocalNode instantiates a new local node with given private key.
func NewLocalNode(prv *ed25519.PrivateKey) (n *LocalNode, err error) {
	// create new instance of a local node
	pub := prv.Public()
	n = &LocalNode{
		Peer: Peer{
			pubKey: pub,
			addr:   NewAddressFromKey(pub),
			inCh:   make(chan Message),
			conn:   nil,
		},
		prvKey:  prv,
		buckets: NewBucketList(),
		hdlrs:   NewHandlerList(),
		srvs:    NewHandlerList(),
		lastID:  0,
	}
	// add all standard services (DHT)
	n.srvs.Add(PING, n.PingService)
	return
}

// IsRemote returns true if node is a remote peer
func (n *LocalNode) IsRemote() bool {
	return false
}

// Learn about a new peer in the network
func (n *LocalNode) Learn(addr *Address) {
	// compute the distance between node and addr
	k := addr.Distance(n.Address()).BitLen() - 1
	// add peer to routing table
	n.buckets.Add(k, addr)
}

// PrivateKey returns the private Ed25519 key for the node
func (n *LocalNode) PrivateKey() *ed25519.PrivateKey {
	return n.prvKey
}

// Run the local node
func (n *LocalNode) Run(ctx context.Context) {

	// we do periodic jobs once every minute
	// and remember the epoch we are in
	epoch := 0
	period := time.Tick(time.Minute)

	// run the connector listener
	n.conn.Listen(ctx, n.inCh)

	// operate message pump
	for {
		select {
		// process incoming message
		case msg := <-n.inCh:
			hdr := msg.Header()
			log.Printf("[%.8s] Received %s from '%.8s'\n", hdr.Receiver, MsgType(msg), hdr.Sender)

			mt := int(hdr.Type)
			switch mt % 2 {
			//----------------------------------------------------------
			// Incoming request
			//----------------------------------------------------------
			case 1:
				// lookup service handling the request
				if !n.srvs.Handle(ctx, mt, msg) {
					log.Printf("Unknown request type '%d'...\n", hdr.Type)
				}

			//----------------------------------------------------------
			// Incoming response
			//----------------------------------------------------------
			case 0:
				if !n.hdlrs.Handle(ctx, int(hdr.TxId), msg) {
					log.Printf("Unknown response for id '%d'...\n", hdr.TxId)
				}
			}

		// periodic jobs
		case <-period:
			epoch++

		// externally cancelled
		case <-ctx.Done():
			return
		}
	}
}

// TaskCallback is used to notify listener of events during task processing
type TaskHandler struct {

	// Message handler for responses in task
	msgHdlr MessageHandler

	// Timeout duration
	timeout time.Duration
}

// Task is a generic wrapper for tasks running on a local node. It sends
// an initial message 'm' from sender 'n' to receiver 'o'; responses from
// 'o' are handled by 'f'. If 'f' returns true, no further responses from
// the receiver are expected.
func (n *LocalNode) Task(ctx context.Context, o Node, m Message, f *TaskHandler) (err error) {
	// register for responses
	ctrl := make(chan int)
	txid := int(m.Header().TxId)
	n.hdlrs.Add(txid, func(ctx context.Context, m Message) bool {
		// call the "real" handler
		rc := f.msgHdlr(ctx, m)
		if rc {
			// no more responses expected
			ctrl <- 0
		}
		return rc
	})
	// send message
	ctx_send, _ := context.WithDeadline(ctx, time.Now().Add(f.timeout))
	if err = n.conn.Send(ctx_send, m); err != nil {
		return
	}
	// timeout for response(s)
	tick := time.Tick(f.timeout)
	select {
	case <-ctrl:
	case <-tick:
		err = ErrNodeTimeout
	}
	// unregister handler
	n.hdlrs.Remove(txid)
	return
}

//----------------------------------------------------------------------
// Helper methods
//----------------------------------------------------------------------

// nextId returns the next integer identifier
func (n *LocalNode) nextId() uint32 {
	n.lastID++
	return n.lastID
}
