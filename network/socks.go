/*
 * Connect through SOCKS5 proxy as specified in RFC 1928. 
 *
 * (c) 2012 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package network

///////////////////////////////////////////////////////////////////////
// Import external declarations.

import (
	"errors"
	"github.com/bfix/gospel/logger"
	"net"
)

///////////////////////////////////////////////////////////////////////
// Local definitions

var socksState []string = []string{
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

///////////////////////////////////////////////////////////////////////
// Public methods

func Socks5Connect(proto string, addr string, port int, proxy string) (net.Conn, error) {

	if proto != "tcp" {
		logger.Printf(logger.ERROR, "[network] Unsupported protocol '%s'.\n", proto)
		return nil, errors.New("Unsupported protocol")
	}

	//-----------------------------------------------------------------
	// connect to SOCKS5 proxy via TCP
	//-----------------------------------------------------------------
	conn, err := net.Dial("tcp", proxy)
	if err != nil {
		logger.Printf(logger.ERROR, "[network] failed to connect to proxy server: %s\n", err.Error())
		return nil, err
	}

	data := make([]byte, 1024)

	//-----------------------------------------------------------------
	// negotiate authentication
	//-----------------------------------------------------------------
	data[0] = 5 // SOCKS version
	data[1] = 1 // One available authentication method
	data[2] = 0 // No authentication required
	if n, err := conn.Write(data[:3]); n != 3 {
		logger.Printf(logger.ERROR, "[network] failed to write to proxy server: %s\n", err.Error())
		conn.Close()
		return nil, err
	}
	if n, err := conn.Read(data); n != 2 {
		logger.Printf(logger.ERROR, "[network] failed to read from proxy server: %s\n", err.Error())
		conn.Close()
		return nil, err
	}
	if data[0] != 5 || data[1] == 0xFF {
		logger.Println(logger.ERROR, "[network] proxy server refuses non-authenticated connection.")
		conn.Close()
		return nil, err
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
	if n, err := conn.Write(data[:7+size]); n != (7 + size) {
		logger.Printf(logger.ERROR, "[network] failed to write to proxy server: %s\n", err.Error())
		conn.Close()
		return nil, err
	}
	conn.Read(data)
	if data[1] != 0 {
		err = errors.New(socksState[data[1]])
		logger.Printf(logger.ERROR, "[network] proxy server failed: %s\n", err.Error())
		conn.Close()
		return nil, err
	}

	//return connection	
	return conn, nil
}
