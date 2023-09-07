package p2p

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
	"errors"
	"net"
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/data"
	"github.com/bfix/gospel/logger"
)

// Error codes
var (
	ErrNodeTimeout        = errors.New("operation timed out")
	ErrNodeRemote         = errors.New("no remote nodes allowed")
	ErrNodeConnectorSet   = errors.New("connector for node already set")
	ErrNodeSendNoReceiver = errors.New("send has no recipient")
	ErrNodeResolve        = errors.New("can't resolve network address")
	ErrNodeMsgType        = errors.New("invalid message type")
)

// constants
var (
	NodeTick = 1 * time.Minute
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
	addr   *Address            // address of the node in the P2P network

	// standard services
	ping   *PingService
	lookup *LookupService
	relay  *RelayService

	inCh chan Message // channel for incoming messages
	conn Connector    // send/receive stub

	buckets *BucketList  // routing table
	srvcs   *ServiceList // list of services

	lastID uint64 // last used identifier
}

// NewNode instantiates a new local node with given private key.
func NewNode(prv *ed25519.PrivateKey) (n *Node, err error) {
	// create new instance of a local node
	pub := prv.Public()
	addr := NewAddressFromKey(pub)
	logger.Printf(logger.INFO, "[%.8s] Creating node...\n", addr)
	n = &Node{
		prvKey: prv,
		addr:   addr,
		inCh:   make(chan Message),
		conn:   nil,
		srvcs:  NewServiceList(),
	}
	// add all standard services (P2P)
	n.ping = NewPingService()
	n.AddService(n.ping)
	n.lookup = NewLookupService()
	n.AddService(n.lookup)
	n.relay = NewRelayService()
	n.AddService(n.relay)

	// set node attributes with back references
	n.buckets = NewBucketList(addr, n.ping)
	return
}

//----------------------------------------------------------------------
// Address handling
//----------------------------------------------------------------------

// Address returns the P2P address of a node
func (n *Node) Address() *Address {
	return n.addr
}

//----------------------------------------------------------------------
// Service handling
//----------------------------------------------------------------------

// AddService to add a new service running on the node.
func (n *Node) AddService(s Service) bool {
	if n.srvcs.Add(s) {
		s.link(n)
		logger.Printf(logger.DBG, "[%.8s] Service '%s' added to node\n", n.addr, s.Name())
		return true
	}
	logger.Printf(logger.ERROR, "[%.8s] Failed to register service '%s' on node\n", n.addr, s.Name())
	return false
}

// Service returns the named service instance for node. Useful for external
// services that are unknown within the framework.
func (n *Node) Service(name string) Service {
	return n.srvcs.Get(name)
}

// PingService returns the PING service instance
func (n *Node) PingService() *PingService {
	return n.ping
}

// LookupService returns the PING service instance
func (n *Node) LookupService() *LookupService {
	return n.lookup
}

// RelayService returns the RELAY service instance
func (n *Node) RelayService() *RelayService {
	return n.relay
}

//----------------------------------------------------------------------
// Message exchange (incoming and outgoing messages)
//----------------------------------------------------------------------

// RelayedMessage creates a nested relay message
func (n *Node) RelayedMessage(msg Message, peers []*Address) (Message, error) {
	// onion-wrapping
	envelope := func(m Message, peer, rcv *Address) (Message, error) {
		// resolve next hop
		hop := m.Header().Receiver
		netw := n.Resolve(rcv)
		// skip hops without network address
		if netw == nil {
			return nil, ErrNodeResolve
		}
		endp := &Endpoint{
			Addr: hop,
			Endp: NewString(netw.String()),
		}
		// wrap message into packet
		pkt, err := n.Wrap(m)
		if err != nil {
			return nil, err
		}
		// assemble relay message
		wrp, ok := n.relay.NewMessage(ReqRELAY).(*RelayMsg)
		if !ok {
			return nil, ErrNodeMsgType
		}
		wrp.Set(endp, pkt)
		wrp.TxID = n.NextID()
		wrp.Flags = MsgfRelay
		wrp.Sender = n.Address()
		wrp.Receiver = peer
		return wrp, nil
	}
	// prepare original message
	hdr := msg.Header()
	hdr.Flags |= MsgfRelay
	// create self-relayed message to announce network address
	m, err := envelope(msg, hdr.Receiver, n.Address())
	if err != nil {
		return nil, err
	}
	// assemble nested relay message
	for _, hop := range peers {
		m, err = envelope(m, hop, m.Header().Receiver)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Send a message from this node to a peer on the network
func (n *Node) Send(ctx context.Context, msg Message) (err error) {
	// log error message if send faild
	defer func() {
		if err != nil {
			logger.Printf(logger.ERROR, "[%.8s] Send failed: %s\n", n.addr, err.Error())
		}
	}()
	// announce message transfer
	logger.Printf(logger.INFO, "[%.8s] Sending message %s\n", n.addr, msg)

	// only associated node can send message
	hdr := msg.Header()
	if !n.addr.Equals(hdr.Sender) {
		err = ErrTransSenderMismatch
		return
	}
	// wrap message into packet
	pkt, err := n.Wrap(msg)
	if err != nil {
		return
	}
	// resolve receiver
	netw := n.Resolve(hdr.Receiver)
	if netw == nil {
		err = ErrTransUnknownReceiver
		return
	}
	// send message
	err = n.SendRaw(ctx, netw, pkt)
	return
}

// SendRaw message from this node to a peer on the network
func (n *Node) SendRaw(ctx context.Context, dst net.Addr, pkt *Packet) error {
	return n.conn.Send(ctx, dst, pkt)
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
// Packet-related methods
//----------------------------------------------------------------------

// Pack message into a byte array
func (n *Node) Pack(msg Message) ([]byte, error) {
	pkt, err := n.Wrap(msg)
	if err != nil {
		return nil, err
	}
	return data.Marshal(pkt)
}

// Unpack byte array into a message
func (n *Node) Unpack(buf []byte, size int) (msg Message, err error) {
	// convert data to packet
	pkt := new(Packet)
	if err = data.Unmarshal(pkt, buf[:size]); err != nil {
		return
	}
	if size != int(pkt.Size) {
		err = ErrPacketSizeMismatch
		return
	}
	// decrypt packet into message
	return n.Unwrap(pkt)
}

// Wrap message into a packet
func (n *Node) Wrap(msg Message) (pkt *Packet, err error) {
	// wrap the message into a packet
	return NewPacket(msg, n.prvKey)
}

// Unwrap packet into a message
func (n *Node) Unwrap(pkt *Packet) (msg Message, err error) {
	// decrypt packet into message
	return pkt.Unwrap(n.prvKey, n.srvcs.MessageFactory)
}

//----------------------------------------------------------------------
// Endpoint management (mapping node addresses to network addresses)
//----------------------------------------------------------------------

// Learn about a new peer in the network
func (n *Node) Learn(addr *Address, endp string) error {
	// add peer to routing table
	n.buckets.Add(addr)
	// learn network endpoint if specified
	if len(endp) > 0 {
		// get the associated network address
		netw, err := n.conn.NewAddress(endp)
		if err != nil {
			return err
		}
		return n.conn.Learn(addr, netw)
	}
	return nil
}

// Resolve peer address to network address
// This will only deliver a result if the address has been learned before by
// the transport connector. Unknown endpoints must be resolved with the
// LookupService.
func (n *Node) Resolve(addr *Address) net.Addr {
	res := n.conn.Resolve(addr)
	if res == nil {
		return nil
	}
	return res
}

// NewNetworkAddr returns the network address for endpoint (transport-specific)
func (n *Node) NewNetworkAddr(endp string) (net.Addr, error) {
	return n.conn.NewAddress(endp)
}

// Sample returns a random collection of node/network address pairs this node
// has learned during up-time.
func (n *Node) Sample(num int, skip *Address) []*Address {
	return n.conn.Sample(num, skip)
}

//----------------------------------------------------------------------
// Run the node with services
//----------------------------------------------------------------------

// Run the local node
func (n *Node) Run(ctx context.Context) {

	// we do periodic jobs once every minute
	// and remember the epoch we are in
	epoch := 0
	period := time.NewTicker(NodeTick)

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
					if _, err := n.srvcs.Respond(ctx, msg); err != nil {
						logger.Printf(logger.ERROR, "[%.8s] Respond failed: %s\n", n.addr, err.Error())
					}

				//----------------------------------------------------------
				// Incoming response
				//----------------------------------------------------------
				case 0:
					// lookup service listening to response
					if _, err := n.srvcs.Listen(ctx, msg); err != nil {
						logger.Printf(logger.ERROR, "[%.8s] Listen failed: %s\n", n.addr, err.Error())
					}
				}
			}()

		// periodic jobs
		case <-period.C:
			epoch++
			n.conn.Epoch(epoch)

		// externally cancelled
		case <-ctx.Done():
			return
		}
	}
}

//----------------------------------------------------------------------
// Helper methods
//----------------------------------------------------------------------

// Closest returns the 'num' closest nodes we know of
func (n *Node) Closest(num int) []*Address {
	return n.buckets.Closest(num)
}

// NextID returns the next unique identifier for this node context
func (n *Node) NextID() uint64 {
	n.lastID++
	return n.lastID

}
