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
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/bfix/gospel/data"
	"github.com/bfix/gospel/logger"
)

//======================================================================
// Internet-based transport: send and receive packets as UDP datagrams
//======================================================================

//----------------------------------------------------------------------
// UDP-based connector implementation
//----------------------------------------------------------------------

// UDPConnector is a stub between a node and the UDP transport implementation.
type UDPConnector struct {
	trans   *UDPTransport
	node    *Node
	addr    *net.UDPAddr
	conn    net.PacketConn
	running bool

	cache  map[string]*net.UDPAddr
	sample []*Address
	pos    int
	lock   sync.Mutex
}

// NewUDPConnector creates an empty instance for node at given network address
func NewUDPConnector(trans *UDPTransport, node *Node, addr *net.UDPAddr) *UDPConnector {
	// assemble connector
	conn := &UDPConnector{
		trans:   trans,
		node:    node,
		addr:    addr,
		conn:    nil,
		running: false,
		cache:   make(map[string]*net.UDPAddr),
		sample:  make([]*Address, SampleCache),
		pos:     0,
	}
	// register our own node
	_ = conn.Learn(node.Address(), addr)
	return conn
}

// NewAddress returns a new UDP address for an endpoint
func (c *UDPConnector) NewAddress(endp string) (net.Addr, error) {
	return net.ResolveUDPAddr("udp", endp)
}

// Sample returns a random collection of node/network address pairs this node
// has learned during up-time.
func (c *UDPConnector) Sample(num int, skip *Address) []*Address {
	// limit number of hops
	if num > MaxSample {
		num = MaxSample
	}
	// check if request can be satisfied
	if num > len(c.cache)-2 {
		// too few cache entries
		return nil
	}
	// collect random addresses
	res := make([]*Address, num)
loop:
	for i := 0; i < num; {
		pos := rand.Intn(SampleCache)
		addr := c.sample[pos]
		if addr == nil || addr.Equals(c.node.addr) || addr.Equals(skip) {
			continue
		}
		for _, v := range res {
			if v == nil {
				break
			}
			if addr.Equals(v) {
				continue loop
			}
		}
		res[i] = addr
		i++
	}
	return res
}

// Send message from node to the UDP network.
func (c *UDPConnector) Send(ctx context.Context, dst net.Addr, pkt *Packet) error {
	// check if we have an UDP connection
	if c.conn == nil {
		return ErrTransClosed
	}
	buf, err := data.Marshal(pkt)
	if err != nil {
		return err
	}
	// do the UDP transfer
	n, err := c.conn.WriteTo(buf, dst)
	if err != nil {
		return ErrTransWrite
	}
	if n != len(buf) {
		return ErrTransWriteShort
	}
	return nil
}

// Listen on an UDP address/port for incoming packets
func (c *UDPConnector) Listen(ctx context.Context, ch chan Message) {

	// allocate buffer space
	buffer := make([]byte, MaxMsgSize)
	nodeAddr := c.node.Address()

	// assemble listener configuration
	cfg := &net.ListenConfig{
		Control: func(netw string, addr string, raw syscall.RawConn) error {
			logger.Printf(logger.DBG, "[%.8s] Starting listener at %s:%s...\n", nodeAddr, netw, addr)
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
				logger.Printf(logger.ERROR, "[%.8s] ERROR: Failed to (re-start) UDP connection", nodeAddr)
				logger.Printf(logger.ERROR, "       %s\n", err.Error())
				// wait some time, then retry
				time.Sleep(3 * time.Second)
				continue
			}
			for c.running {
				// read single UDP packet
				n, addr, err := c.conn.ReadFrom(buffer)
				if err != nil {
					logger.Printf(logger.ERROR, "[%.8s] Listener failed: %s\n", nodeAddr, err.Error())
					break
				}
				//log.Printf("[%.8s] Packet received from %s\n", nodeAddr, addr)

				// convert to message
				msg, err := c.node.Unpack(buffer, n)
				if err != nil {
					logger.Printf(logger.ERROR, "[%.8s] Unwrapping packet failed: %s\n", nodeAddr, err.Error())
					continue
				}
				hdr := msg.Header()
				// is packet for this node?
				if !hdr.Receiver.Equals(c.node.Address()) || (hdr.Flags&MsgfDrop != 0) {
					// no: drop packet and continue
					logger.Printf(logger.WARN, "[%.8s] Dropping packet from '%.8s'\n", nodeAddr, hdr.Receiver)
					continue
				}
				// tell transport and node about the sender (in case it is unknown and not forwarded)
				if hdr.Flags&MsgfRelay == 0 {
					_ = c.Learn(hdr.Sender, addr)
					_ = c.node.Learn(hdr.Sender, "")
				}
				// let the node handle the message
				ch <- msg
			}
			// close the listener
			logger.Printf(logger.WARN, "[%.8s] Closing listener\n", nodeAddr)
			c.conn.Close()
			c.conn = nil
			// wait before retrying
			time.Sleep(10 * time.Second)
		}
	}()
}

// Learn network address of node address
func (c *UDPConnector) Learn(addr *Address, endp net.Addr) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	switch x := endp.(type) {
	case *net.UDPAddr:
		c.cache[addr.String()] = x
		c.sample[c.pos] = addr
		c.pos = (c.pos + 1) % SampleCache
		//log.Printf("[connectr] Learned network address '%s' for node '%.8s'\n", x, addr)
	default:
		//log.Printf("[connectr] Can't learn network address '%s' for node '%.8s'\n", x, addr)
	}
	return nil
}

// Resolve node address into a network address
func (c *UDPConnector) Resolve(addr *Address) net.Addr {
	c.lock.Lock()
	defer c.lock.Unlock()

	endp := c.cache[addr.String()]
	if endp == nil {
		return nil
	}
	//log.Printf("[connectr] Resolve node '%.8s' ==> %s\n", addr, endp)
	return endp
}

// Epoch step: perform periodic tasks
func (c *UDPConnector) Epoch(epoch int) {
}

//----------------------------------------------------------------------
// Internet-based transport layer
//----------------------------------------------------------------------

// UDPTransport handles the transport of packets between nodes over the
// internet using the UDP protocol.
type UDPTransport struct {
	// nodes registered with transport
	registry map[string]bool
}

// NewUDPTransport instantiates a new UDP transport layer where the
// listening socket is bound to the specified address (host:port).
func NewUDPTransport() *UDPTransport {
	// instantiate transport
	return &UDPTransport{
		registry: make(map[string]bool),
	}
}

// Open transport based on configuration
func (t *UDPTransport) Open(cfg TransportConfig) error {
	// nothing to setup...
	return nil
}

// Register a node for participation in the transport layer.
func (t *UDPTransport) Register(ctx context.Context, n *Node, endp string) error {
	// get the associated network address
	netwAddr, err := net.ResolveUDPAddr("udp", endp)
	if err != nil {
		return err
	}
	// check for already registered address
	addr := n.Address().String()
	if _, ok := t.registry[addr]; ok {
		return ErrTransAddressDup
	}
	// connect to suitable connector
	if err = n.Connect(NewUDPConnector(t, n, netwAddr)); err != nil {
		return err
	}
	logger.Printf(logger.DBG, "[%.8s] Registered with transport at %s\n", addr, netwAddr)
	return nil
}

// Close transport
func (t *UDPTransport) Close() error {
	return nil
}
