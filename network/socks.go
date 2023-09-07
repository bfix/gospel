package network

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
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	gerr "github.com/bfix/gospel/errors"
)

var socksState = []string{
	"succeeded",
	"general SOCKS server failure",
	"connection not allowed by ruleset",
	"Network unreachable",
	"Host unreachable",
	"Connection refused",
	"TTL expired",
	"Command not supported",
	"Address type not supported",
	"to X'FF' unassigned",
}

// Error codes
var (
	ErrSocksUnsupportedProtocol = errors.New("unsupported protocol (TCP only)")
	ErrSocksInvalidProxyScheme  = errors.New("invalid proxy scheme")
	ErrSocksInvalidHost         = errors.New("invalid host definition (missing port)")
	ErrSocksInvalidPort         = errors.New("invalid host definition (port out of range)")
	ErrSocksProxyFailed         = errors.New("proxy server failed")
)

// Socks5Connect connects to a SOCKS5 proxy.
func Socks5Connect(proto string, addr string, port int, proxy string) (net.Conn, error) {
	return Socks5ConnectTimeout(proto, addr, port, proxy, 0)
}

// Socks5ConnectTimeout connects to a SOCKS5 proxy with timeout.
func Socks5ConnectTimeout(proto string, addr string, port int, proxy string, timeout time.Duration) (conn net.Conn, err error) {
	if proto != "tcp" {
		err = ErrSocksUnsupportedProtocol
		return
	}
	p, err := url.Parse(proxy)
	if err != nil {
		return
	}
	if len(p.Scheme) > 0 && p.Scheme != "socks5" {
		err = gerr.New(ErrSocksInvalidProxyScheme, "scheme %s", p.Scheme)
		return
	}
	idx := strings.Index(p.Host, ":")
	if idx == -1 {
		err = ErrSocksInvalidHost
		return
	}
	var pPort int
	if pPort, err = strconv.Atoi(p.Host[idx+1:]); err != nil || pPort < 1 || pPort > 65535 {
		err = gerr.New(ErrSocksInvalidPort, "port %d", pPort)
		return
	}
	if timeout == 0 {
		conn, err = net.Dial("tcp", p.Host)
	} else {
		conn, err = net.DialTimeout("tcp", p.Host, timeout)
	}
	if err != nil {
		return
	}

	data := make([]byte, 1024)

	//-----------------------------------------------------------------
	// negotiate authentication
	//-----------------------------------------------------------------
	data[0] = 5 // SOCKS version
	data[1] = 1 // One available authentication method
	data[2] = 0 // No authentication required
	if timeout > 0 {
		if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
	}
	var n int
	if n, err = conn.Write(data[:3]); n != 3 {
		err = gerr.New(err, "failed to write to proxy server")
		conn.Close()
		return
	}
	if timeout > 0 {
		if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
	}
	if n, err = conn.Read(data); n != 2 {
		err = gerr.New(err, "failed to read from proxy server")
		conn.Close()
		return
	}
	if data[0] != 5 || data[1] == 0xFF {
		err = gerr.New(err, "proxy server refuses non-authenticated connection")
		conn.Close()
		return
	}

	//-----------------------------------------------------------------
	// connect to target (request/reply processing)
	//-----------------------------------------------------------------
	dn := []byte(addr)
	size := len(dn)

	data[0] = 5          // SOCKS versions
	data[1] = 1          // connect to target
	data[2] = 0          // reserved
	data[3] = 3          // domain name specified
	data[4] = byte(size) // length of domain name
	for i, v := range dn {
		data[5+i] = v
	}
	data[5+size] = (byte)(port / 256)
	data[6+size] = (byte)(port % 256)
	if timeout > 0 {
		if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
	}
	if n, err = conn.Write(data[:7+size]); n != (7 + size) {
		err = gerr.New(err, "failed to write to proxy server")
		conn.Close()
		return
	}
	if timeout > 0 {
		if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
	}
	if _, err = conn.Read(data); err != nil {
		conn.Close()
		return
	}
	if data[1] != 0 {
		err = gerr.New(ErrSocksProxyFailed, socksState[data[1]])
		conn.Close()
		return
	}
	// remove timeout from connection
	var zero time.Time
	err = conn.SetDeadline(zero)
	// return connection
	return
}
