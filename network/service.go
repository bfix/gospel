package network

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
	"github.com/bfix/gospel/logger"
	"net"
)

// Service is a user-defined service handler that handles TCP/UDP
// client sessions. The interface defines four methods:
//
// - Process (conn): Main handler routine for connection
//
// - GetName(): Return service name (for logging output)
//
// - CanHandle (protocol): Check if handler can process given
//   network protocol (TCP or UDP on IPv4 or IPv6)
//
// - IsAllowed (addr): Checkk if remote address is allowed to
//   be served by the service handler.
type Service interface {
	Process(conn net.Conn)          // main handler routine
	GetName() string                // get symbolic name of service
	CanHandle(protocol string) bool // check network protocol
	IsAllowed(remote string) bool   // check remote address
}

// RunService runs a TCP/UDP network service with user-defined
// session handler.
func RunService(network, addr string, hdlr []Service) error {

	// initialize control service
	service, err := net.Listen(network, addr)
	if err != nil {
		logger.Println(logger.ERROR, "[network] service start-up failed for '"+network+"/"+addr+"': "+err.Error())
		return err
	}

	// handle connection requests
	go func() {
		for {
			// wait for connection request
			client, err := service.Accept()
			if err != nil {
				logger.Println(logger.ERROR, "[network] accept failed for '"+network+"/"+addr+"': "+err.Error())
				continue
			}
			// find service interface that can handle the request
			accepted := false
			for _, srv := range hdlr {
				// check if connection is allowed:
				remote := client.RemoteAddr().String()
				protocol := client.RemoteAddr().Network()
				// check for matching protocol
				if !srv.CanHandle(protocol) {
					logger.Printf(logger.WARN, "["+srv.GetName()+"] rejected connection protocol '%s' from %s\n", protocol, remote)
					continue
				}
				// check for matching remote address
				if !srv.IsAllowed(remote) {
					logger.Printf(logger.WARN, "["+srv.GetName()+"] rejected connection from %s\n", remote)
					continue
				}
				// connection accepted
				logger.Printf(logger.INFO, "["+srv.GetName()+"] accepted connection from %s\n", remote)
				accepted = true

				// start handler
				go srv.Process(client)
				break
			}
			// close unhandled connections
			if !accepted {
				client.Close()
			}
		}
	}()

	// report success
	logger.Println(logger.INFO, "[network] service started on '"+network+"/"+addr+"'...")
	return nil
}
