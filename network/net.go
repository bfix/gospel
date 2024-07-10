package network

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
	"errors"
	"net"
	"os"
	"strings"
	"time"

	gerr "github.com/bfix/gospel/errors"
	"github.com/bfix/gospel/logger"
)

var (
	delay, _   = time.ParseDuration("1ms")
	retries    = 1000
	timeout, _ = time.ParseDuration("100us")
)

// Error codes
var (
	ErrNetInvalidEndpoint = errors.New("invalid endpoint")
	ErrNetInvalidNetwork  = errors.New("invalid network")
)

// SendData sends data over network connection (stream-oriented).
func SendData(conn net.Conn, data []byte, srv string) bool {

	count := len(data) // total length of data
	start := 0         // start position of slice
	retry := 0         // retry counter

	// write data to socket buffer
	for count > 0 {
		// set timeout
		if conn.SetDeadline(time.Now().Add(timeout)) != nil {
			return false
		}
		// get (next) chunk to be send
		chunk := data[start : start+count]
		if num, err := conn.Write(chunk); num > 0 {
			// advance slice on partial write
			start += num
			count -= num
			retry = 0
		} else if err != nil {
			// handle error condition
			switch nerr := err.(type) {
			case net.Error:
				// network error: retry...
				if nerr.Timeout() {
					retry++
					time.Sleep(delay)
					if retry == retries {
						logger.Printf(logger.ERROR, "[%s] Write failed after retries: %s", srv, err.Error())
						return false
					}
				}
			default:
				logger.Printf(logger.INFO, "[%s] Connection closed by peer", srv)
				return false
			}
		}
	}
	// report success
	if retry > 0 {
		logger.Printf(logger.INFO, "[%s] %d retries needed to send data.", srv, retry)
	}
	return true
}

// RecvData receives data over network connection (stream-oriented).
func RecvData(conn net.Conn, data []byte, srv string) (int, bool) {

	for retry := 0; retry < retries; {
		// set timeout
		if conn.SetDeadline(time.Now().Add(timeout)) != nil {
			return 0, false
		}
		// read data from socket buffer
		n, err := conn.Read(data)
		if err != nil {
			// handle error condition
			switch nerr := err.(type) {
			case net.Error:
				// network error: retry...
				if nerr.Timeout() {
					retry++
					continue
				}
			default:
				logger.Printf(logger.INFO, "[%s] Connection closed by peer", srv)
				return 0, false
			}
		}
		// report success
		if retry > 0 {
			logger.Printf(logger.INFO, "[%s] %d retries needed to receive data.", srv, retry)
		}
		return n, true
	}
	// retries failed
	logger.Printf(logger.ERROR, "[%s] Read failed after retries...", srv)
	return 0, false
}

// SplitNetworkEndpoint splits a string like "tcp:127.0.0.1:80" or
// "unix:/run/app/app.sock" into components.
func SplitNetworkEndpoint(networkendp string) (network string, endp string, err error) {
	pos := strings.Index(networkendp, ":")
	if pos == -1 {
		err = ErrNetInvalidEndpoint
		return
	}
	network = networkendp[:pos]
	endp = networkendp[pos+1:]
	switch network {
	// local Unix domain socket
	case "unix":
		_, err = os.Stat(endp)
		return
	// IP-based transport
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		return
	}
	err = gerr.New(ErrNetInvalidNetwork, network)
	return
}
