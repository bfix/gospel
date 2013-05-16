/*
 * Run TCP/UDP service loop: Create listener for incoming connection
 * requests as a go-routine and call user-defined service handler as
 * a go-routine to handle client sessions.
 *
 * (c) 2010-2012 Bernd Fix   >Y<
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
	"github.com/bfix/gospel/logger"
	"net"
)

///////////////////////////////////////////////////////////////////////
// Public interfaces

/*
 * User-defined service handler: Handle TCP/UDP client sessions.
 * The interface defines four methods:
 * - Process (conn): Main handler routine for connection
 * - GetName(): Return service name (for logging output)
 * - CanHandle (protocol): Check if handler can process given
 *       network protocol (TCP or UDP on IPv4 or IPv6)
 * - IsAllowed (addr): Checkk if remote address is allowed to
 *       be served by the service handler.
 */
type Service interface {
	Process(conn net.Conn)          // main handler routine
	GetName() string                // get symbolic name of service
	CanHandle(protocol string) bool // check network protocol
	IsAllowed(remote string) bool   // check remote address
}

///////////////////////////////////////////////////////////////////////
// Public methods

/*
 * Run TCP/UDP network service with user-defined session handler.
 * @param network string - network identifier (TCP/UDP on IPv4/v6)
 * @param addr string - address:port specification of service
 * @param hdlr []Service - implementations of service interface
 * @return error - error object (or nil)
 */
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

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-11 08:19:55  brf
//  First release as free software (GPL3+)
//
//	Revision 1.0  2010-11-18 23:17:06  brf
//  Initial revision.
//
