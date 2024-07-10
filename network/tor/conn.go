package tor

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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
	"strings"
	"time"

	"github.com/bfix/gospel/network"
)

//======================================================================
// Tor connectivity functions
//======================================================================

// Error codes
var (
	ErrTorInvalidProto = fmt.Errorf("only TCP protocol allowed")
)

// Dial a Tor-based connection
func (s *Service) Dial(netw, address string, flags ...string) (net.Conn, error) {
	return s.DialTimeout(netw, address, 0, flags...)
}

// DialTimeout to establish a Tor-based connection with timeout
func (s *Service) DialTimeout(netw, address string, timeout time.Duration, flags ...string) (net.Conn, error) {
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
	// determine best proxy port
	socks, err := s.GetSocksPort(flags...)
	if err != nil {
		return nil, err
	}
	if strings.LastIndex(socks, ":") != -1 {
		_, socks, err = net.SplitHostPort(socks)
		if err != nil {
			return nil, err
		}
	}
	proxy := fmt.Sprintf("socks5://%s:%s", s.host, socks)
	// connect through Tor proxy
	return network.Socks5ConnectTimeout(netw, host, int(port), proxy, timeout)
}
