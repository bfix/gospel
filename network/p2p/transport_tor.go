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
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/bfix/gospel/logger"
	"github.com/bfix/gospel/network"
	"github.com/bfix/gospel/network/tor"
)

//======================================================================
// Tor-based transport: send and receive packets via Tor curcuits and
// hidden services.
//
// Each node (representent by its Ed25519 private key) has a P2P
// network address (identifier) and a Tor hidden service address
// (onion or hidden service address); both network addresses are
// derived from the associated public key of the node and are
// interchangable so that the Tor-based transport (and its connector)
// does not need to resolve network addresses for given P2P identifiers
// (they can simply be "computed"). As a consequence the "lookup
// service" is usually not called to find nodes; the routing table
// (buckets) is only filled when DHT-related lookups are performed
// (not part of the base P2P package).
//======================================================================

//----------------------------------------------------------------------
// Tor-based network address (onion address of hidden service)
//----------------------------------------------------------------------

// TorAddress is an onion address as each node is listing for
// incoming packets as a "hidden service".
type TorAddress struct {
	addr string
}

// NewTorAddress creates a new onion address from the P2P address of a node.
func NewTorAddress(addr *Address) (*TorAddress, error) {
	onion, err := tor.ServiceID(addr.PublicKey())
	if err != nil {
		return nil, err
	}
	return &TorAddress{
		addr: onion + ".onion",
	}, nil
}

// Network returns the network label of the address
func (a *TorAddress) Network() string {
	return "tor"
}

// String returns the human-readable network address
func (a *TorAddress) String() string {
	return a.addr
}

//----------------------------------------------------------------------
// Tor-based transport configuration
//----------------------------------------------------------------------

// TorTransportConfig specifies the configuration parameters required to
// instanciate a Tor-based transport for the P2P network.
type TorTransportConfig struct {
	// Ctrl specifies the Tor control interface. It is formatted like
	// "network:endp", where 'network' is either "tcp" (for TCP/IP) or
	// "unix" (for a local Unix socket). 'endp' specifies the endpoint
	// depending on the network; it is "host:port" for "tcp" and a file
	// path in case of a local Unix socket.
	// The same host (either "host" as deined or localhos) must have the
	// defined SOCKS ports open for access to Tor cuircuits.
	Ctrl string `json:"ctrl"`
	// Auth is the authentication password/cookie. If a cookie is used
	// (only applicable for local Tor service instances), the value is
	// set dynamically (and not in a persistent configuration file).
	Auth string `json:"auth"`
}

// TransportType returns the kind of transport implementation targeted
// by the configuration information.
func (c *TorTransportConfig) TransportType() string {
	return "tor"
}

//----------------------------------------------------------------------
// Tor-based connector implementation
//----------------------------------------------------------------------

// TorConnector is a stub between a node and the Tor-based transport
// implementation.
type TorConnector struct {
	trans   *TorTransport
	node    *Node
	addr    net.Addr
	conn    net.Listener
	running bool

	sample []*Address
	pos    int
	lock   sync.Mutex
}

// NewTorConnector creates a connector on transport for a given node
func NewTorConnector(trans *TorTransport, node *Node) (*TorConnector, error) {
	addr, err := NewTorAddress(node.Address())
	if err != nil {
		return nil, err
	}
	return &TorConnector{
		trans:   trans,
		node:    node,
		addr:    addr,
		conn:    nil,
		running: false,
		sample:  make([]*Address, SampleCache),
		pos:     0,
	}, nil
}

// NewAddress returns a new onion address for an endpoint
func (c *TorConnector) NewAddress(endp string) (net.Addr, error) {
	return &TorAddress{
		addr: endp,
	}, nil
}

// Sample returns a random collection of node/network address pairs this node
// has learned during up-time.
func (c *TorConnector) Sample(num int, skip *Address) []*Address {
	// limit number of hops
	if num > MaxSample {
		num = MaxSample
	}
	// check if request can be satisfied
	if num > c.pos-2 {
		// too few entries
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
func (c *TorConnector) Send(ctx context.Context, dst net.Addr, pkt *Packet) error {
	// connect to peer
	endp := fmt.Sprintf("%s:14235", dst.String())
	conn, err := c.trans.ctrl.DialTimeout("tcp", endp, time.Minute)
	if err != nil {
		return err
	}
	// send packet
	if _, err = conn.Write(pkt.Body); err != nil {
		return err
	}
	// close connection
	return conn.Close()
}

// Listen on an UDP address/port for incoming packets
func (c *TorConnector) Listen(ctx context.Context, ch chan Message) {

	// allocate buffer space
	buffer := make([]byte, MaxMsgSize)
	nodeAddr := c.node.Address()

	// assemble listener configuration
	cfg := &net.ListenConfig{
		Control: func(netw string, addr string, raw syscall.RawConn) error {
			logger.Printf(logger.DBG, "[%.8s] Starting listener at %s:%s...", nodeAddr, netw, addr)
			return nil
		},
		KeepAlive: 0,
	}

	// connector up and running
	c.running = true
	go func() {
		var err error
		for c.running {
			if c.conn, err = cfg.Listen(ctx, "tcp", "0.0.0.0:14235"); err != nil {
				logger.Printf(logger.ERROR, "[%.8s] ERROR: Failed to (re-)start TCP listener", nodeAddr)
				logger.Printf(logger.ERROR, "       %s", err.Error())
				// wait some time, then retry
				time.Sleep(3 * time.Second)
				continue
			}
			for c.running {
				// wait for incoming data
				conn, err := c.conn.Accept()
				if err != nil {
					logger.Printf(logger.ERROR, "[%.8s] Listener failed: %s", nodeAddr, err.Error())
					break
				}
				go func(cn net.Conn) {
					// read packet
					n, err := cn.Read(buffer)
					if err != nil {
						logger.Printf(logger.ERROR, "[%.8s] Reading packet failed: %s", nodeAddr, err.Error())
						return
					}
					defer cn.Close()

					// convert to message
					msg, err := c.node.Unpack(buffer, n)
					if err != nil {
						logger.Printf(logger.ERROR, "[%.8s] Unwrapping packet failed: %s", nodeAddr, err.Error())
						return
					}
					hdr := msg.Header()
					// is packet for this node?
					if !hdr.Receiver.Equals(c.node.Address()) || (hdr.Flags&MsgfDrop != 0) {
						// no: drop packet and continue
						logger.Printf(logger.WARN, "[%.8s] Dropping packet from '%.8s'", nodeAddr, hdr.Receiver)
						return
					}
					// tell transport and node about the sender (in case it is unknown and not forwarded)
					if hdr.Flags&MsgfRelay == 0 {
						c.Learn(hdr.Sender, nil)
						c.node.Learn(hdr.Sender, "")
					}
					// let the node handle the message
					ch <- msg

				}(conn)
			}
			// close the listener
			logger.Printf(logger.WARN, "[%.8s] Closing listener", nodeAddr)
			c.conn.Close()
			c.conn = nil
			// wait before retrying
			time.Sleep(10 * time.Second)
		}
	}()

	// start a hidden service
	hs, err := tor.NewOnion(c.node.prvKey)
	if err != nil {
		logger.Printf(logger.ERROR, "[%.8s] Failed to create Tor onion", nodeAddr)
		return
	}
	hs.AddPort(14235, "127.0.0.1:12345")
	if err = hs.Start(c.trans.ctrl); err != nil {
		logger.Printf(logger.ERROR, "[%.8s] Failed to start Tor onion", nodeAddr)
		return
	}
}

// Learn network address of node address is obsolete if Tor transport
// is used; the network address can be computed from the P2P address.
func (c *TorConnector) Learn(addr *Address, endp net.Addr) error {
	// just keep a list of sampled addresses
	c.sample[c.pos%SampleCache] = addr
	c.pos++
	return nil
}

// Resolve node address into a network address is a deterministic function
// if Tor transport is used.
func (c *TorConnector) Resolve(addr *Address) net.Addr {
	netw, err := NewTorAddress(addr)
	if err != nil {
		netw = nil
	}
	return netw
}

//----------------------------------------------------------------------
// Tor-based transport layer
//----------------------------------------------------------------------

// TorTransport handles the transport of packets between nodes over
// Tor curcuits / hidden services.
type TorTransport struct {
	// Tor service controller
	ctrl *tor.Service
	// nodes registered with transport
	registry map[string]bool
	// transport initialized (opened)?
	active bool
}

// NewTorTransport instantiates a new Tor transport layer where the
// listening socket is bound to the specified address (host:port).
func NewTorTransport() *TorTransport {
	// instantiate Tor transport
	return &TorTransport{
		ctrl:     nil,
		registry: make(map[string]bool),
		active:   false,
	}
}

// Open transport based on configuration
func (t *TorTransport) Open(cfg TransportConfig) (err error) {
	// check for inactive transport
	if t.active {
		return ErrTransOpened
	}
	// check for matching configuration type
	if cfg.TransportType() != "tor" {
		return ErrTransInvalidConfig
	}
	torCfg, ok := cfg.(*TorTransportConfig)
	if !ok {
		return ErrTransInvalidConfig
	}
	// connect to the Tor service through the control port
	netw, endp, err := network.SplitNetworkEndpoint(torCfg.Ctrl)
	t.ctrl, err = tor.NewService(netw, endp)
	if err != nil {
		return
	}
	// perform authentication
	if err = t.ctrl.Authenticate(torCfg.Auth); err == nil {
		t.active = true
	}
	return
}

// Register a node for participation in the transport layer. The 'endp'
// argument is ignored as the network address of the node is computed
// from the P2P address of the node.
func (t *TorTransport) Register(ctx context.Context, n *Node, endp string) error {
	// check for opened transport
	if !t.active {
		return ErrTransClosed
	}
	// check for already registered address
	addr := n.Address().String()
	if _, ok := t.registry[addr]; ok {
		return ErrTransAddressDup
	}
	// connect to suitable connector
	conn, err := NewTorConnector(t, n)
	if err != nil {
		return err
	}
	n.Connect(conn)
	logger.Printf(logger.DBG, "[%.8s] Registered with transport at %s\n", addr, conn.addr)
	return nil
}

// Close transport
func (t *TorTransport) Close() error {
	// check for active (open) transport
	if !t.active {
		return ErrTransClosed
	}
	// close controller
	return t.ctrl.Close()
}
