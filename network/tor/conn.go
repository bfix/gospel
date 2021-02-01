package tor

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix
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
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bfix/gospel/network"
)

//======================================================================
// Tor connectivity functions
//======================================================================

// Error codes
var (
	ErrTorInvalidProto = fmt.Errorf("Only TCP protocol allowed")
)

// Dial a Tor-based connection
func Dial(netw, address string, proxy string) (net.Conn, error) {
	return DialTimeout(netw, address, 0, proxy)
}

// DialTimeout to establish a Tor-based connection with timeout
func DialTimeout(netw, address string, timeout time.Duration, proxy string) (net.Conn, error) {
	// check protocol
	if netw != "tcp" {
		return nil, ErrTorInvalidProto
	}
	// split address
	host, portS, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseInt(portS, 10, 32)
	if err != nil {
		return nil, err
	}
	// connect through Tor proxy
	return network.Socks5ConnectTimeout(netw, host, int(port), proxy, timeout)
}
