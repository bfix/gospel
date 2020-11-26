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
	"net"
)

// Error codes
var (
	ErrTransAddressDup      = fmt.Errorf("Address already registered")
	ErrTransUnknownReceiver = fmt.Errorf("Unknown receiver")
	ErrTransAddressInvalid  = fmt.Errorf("Invalid etwork address")
)

//======================================================================
// Transport layer abstraction
//======================================================================

// Connector can send and receive message over a transport layer.
type Connector interface {
	// Send packet to endpoint (low-level transport)
	Send(context.Context, net.Addr, *Packet) error

	// Listen to messages from transport
	Listen(context.Context, chan Message)

	// Learn network address of node address
	Learn(*Address, net.Addr) error

	// Resolve the network address of a node address
	Resolve(*Address) net.Addr

	// NewAddress to instaniate a new endpoint address
	NewAddress(string) (net.Addr, error)

	// Sample given number of nodes with network addresses
	// stored in cache.
	Sample(int, *Address) []*Address
}

// Transport abstraction: Every endpoint (on a local machine) registers
// with its address and receives channels for communication (incoming
// and outgoing messages). The transfer process needs to be started
// with the 'Run()' method for its message pump to work.
type Transport interface {
	// Register a node for participation in this transport
	Register(context.Context, *Node, string) error
}
