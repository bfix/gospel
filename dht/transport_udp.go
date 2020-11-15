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
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/bfix/gospel/data"
)

//======================================================================
// Internet-based transport: send and receive packets as UDP datagrams
//======================================================================

// Error codes
var (
	ErrTransSenderMismatch = fmt.Errorf("Sender mismatch")
	ErrTransUnknownSender  = fmt.Errorf("Unknown sender")
	ErrTransPackaging      = fmt.Errorf("Failed to create packet")
	ErrTransMarshalling    = fmt.Errorf("Failed to marshal message")
	ErrTransClosed         = fmt.Errorf("Can't send on closed UDP connection")
	ErrTransWrite          = fmt.Errorf("Failed write to remote")
	ErrTransWriteShort     = fmt.Errorf("Short write to remote")
)

//----------------------------------------------------------------------
// Internet-based, routable packet transport (UDP based)
//----------------------------------------------------------------------

//----------------------------------------------------------------------
// UDP-based connector implementation
//----------------------------------------------------------------------

// UDPConnector is a stub between a node and the UDP transport implementation.
type UDPConnector struct {
	trans   *UDPTransport
	node    Node
	addr    *net.UDPAddr
	conn    net.PacketConn
	running bool
}

// Send message from node to the UDP network.
func (c *UDPConnector) Send(ctx context.Context, msg Message) (err error) {
	// log error message if send faild
	defer func() {
		if err != nil {
			log.Printf("[%.8s] Send failed: %.8s\n", c.node.Address(), err.Error())
		}
	}()

	// check if we have an UDP connection
	if c.conn == nil {
		err = ErrTransClosed
		return
	}
	// announce message transfer
	hdr := msg.Header()
	log.Printf("[%.8s] Sending %s to '%.8s'\n", c.node.Address(), MsgType(msg), hdr.Receiver)

	// only local nodes can send messages
	if c.node.IsRemote() {
		err = ErrTransRemote
		return
	}
	// only associated node can send message
	if !c.node.Address().Equals(hdr.Sender) {
		err = ErrTransSenderMismatch
		return
	}
	// get network addresses of peers
	nodeS := c.trans.Node(hdr.Sender)
	if nodeS == nil {
		err = ErrTransUnknownSender
		return
	}
	addrR := c.trans.Endpoint(hdr.Receiver)
	if addrR == nil {
		err = ErrTransUnknownReceiver
		return
	}
	// wrap the message into a packet
	pkt, err := NewPacket(msg, nodeS)
	if err != nil {
		err = ErrTransPackaging
		return
	}
	buf, err := data.Marshal(pkt)
	if err != nil {
		err = ErrTransMarshalling
		return
	}

	// do the UDP transfer
	n, err := c.conn.WriteTo(buf, addrR)
	if err != nil {
		err = ErrTransWrite
		return
	}
	if n != len(buf) {
		err = ErrTransWriteShort
		return
	}
	return nil
}

// Listen on an UDP address/port for incoming packets
func (c *UDPConnector) Listen(ctx context.Context, ch chan Message) {

	// allocate buffer space
	buffer := make([]byte, MAX_MSGSIZE)
	nodeAddr := c.node.Address()

	// assemble listener configuration
	cfg := &net.ListenConfig{
		Control: func(netw string, addr string, raw syscall.RawConn) error {
			log.Printf("[%.8s] Starting listener at %s:%s...\n", nodeAddr, netw, addr)
			return nil
		},
		KeepAlive: 0,
	}

	// connector up and running
	c.running = true
	go func() {
		var err error
		for c.running {
			if c.conn, err = cfg.ListenPacket(ctx, "udp", c.addr.String()); err != nil {
				log.Printf("[%.8s] ERROR: Failed to (re-start) UDP connection", nodeAddr)
				log.Printf("       %s\n", err.Error())
				// wait some time, then retry
				time.Sleep(3 * time.Second)
				continue
			}
			for c.running {
				// read single UDP packet
				n, addr, err := c.conn.ReadFrom(buffer)
				if err != nil {
					log.Printf("[%.8s] Listener failed: %s\n", nodeAddr, err.Error())
					break
				}
				log.Printf("[%.8s] Packet received from %s\n", nodeAddr, addr)

				// convert to transport packet
				pkt := new(Packet)
				if err = data.Unmarshal(pkt, buffer[:n]); err != nil {
					log.Printf("[%.8s] Listener failed: %.8s\n", nodeAddr, err.Error())
					break
				}
				if n != int(pkt.Size) {
					log.Printf("[%.8s] Listener failed with invalid packet size %d (%d)\n", nodeAddr, n, int(pkt.Size))
					break
				}

				// decrypt packet into message
				msg, err := pkt.Unwrap(c.node.PrivateKey())
				if err != nil {
					log.Printf("[%.8s] Unwrapping packet failed: %s\n", nodeAddr, err.Error())
					continue
				}
				// is packet for this node?
				receiver := msg.Header().Receiver
				if !receiver.Equals(c.node.Address()) {
					// no: drop packet and continue
					log.Printf("[%.8s] Dropping packet from '%.8s'\n", nodeAddr, receiver)
					continue
				}
				// tell transport and node about the sender (in case it is unknown)
				sender := msg.Header().Sender
				c.trans.Learn(sender, addr.(*net.UDPAddr))
				c.node.Learn(sender)

				// let the node handle the message
				ch <- msg
			}
			// close the listener
			log.Printf("[%.8s] Closing listener\n", nodeAddr)
			c.conn.Close()
			c.conn = nil
			// wait before retrying
			time.Sleep(10 * time.Second)
		}
	}()
}

//----------------------------------------------------------------------
// Internet-based transport layer
//----------------------------------------------------------------------

// UDPTransport handles the transport of packets between nodes over the
// internet using the UDP protocol.
type UDPTransport struct {
	// network connectors for known nodes
	conns map[string]*UDPConnector
	lock  sync.Mutex
}

// NewUDPTransport instantiates a new UDP transport layer where the
// listening socket is bound to the specified address (host:port).
func NewUDPTransport() Transport {
	// instantiate transport
	return &UDPTransport{
		conns: make(map[string]*UDPConnector),
	}
}

// Register a node for participation in the transport layer.
func (t *UDPTransport) Register(ctx context.Context, n Node, endp string) error {
	// synchronize access to connector list
	t.lock.Lock()
	defer t.lock.Unlock()

	// check for already registered address
	addr := n.Address().String()
	if _, ok := t.conns[addr]; ok {
		return ErrTransAddressDup
	}
	// get the associated network address
	netwAddr, err := net.ResolveUDPAddr("udp", endp)
	if err != nil {
		return err
	}
	// assemble suitable connector
	conn := &UDPConnector{
		trans:   t,
		node:    n,
		addr:    netwAddr,
		running: false,
	}
	n.Connect(conn)
	t.conns[addr] = conn
	log.Printf("[%.8s] Registered with transport at %s\n", addr, netwAddr)
	return nil
}

// Endpoint returns the UDP address associated with an address
func (t *UDPTransport) Endpoint(addr *Address) *net.UDPAddr {
	// synchronize access to connector list
	t.lock.Lock()
	defer t.lock.Unlock()

	if conn, ok := t.conns[addr.String()]; ok {
		return conn.addr
	}
	return nil
}

// Node returns the peer instance for given address
func (t *UDPTransport) Node(addr *Address) Node {
	// synchronize access to connector list
	t.lock.Lock()
	defer t.lock.Unlock()

	if conn, ok := t.conns[addr.String()]; ok {
		return conn.node
	}
	return nil
}

// Learn about a node with given hdt and network address
func (t *UDPTransport) Learn(addr *Address, netw *net.UDPAddr) {
	// synchronize access to connector list
	t.lock.Lock()
	defer t.lock.Unlock()

	addrS := addr.String()
	if conn, ok := t.conns[addrS]; ok {
		// we already have a connector for the address. check if the
		// network address has changed.
		if netw.String() != conn.addr.String() {
			log.Printf("[transprt] New network address '%s' learned for node '%.8s'\n",
				netw, conn.node.Address())
			conn.addr = netw
		}
	} else {
		// create a new connector for address
		n := NewPeer(addr.PublicKey())
		conn = &UDPConnector{
			trans:   t,
			node:    n,
			addr:    netw,
			running: false,
		}
		n.Connect(conn)
		t.conns[addrS] = conn
	}
}
