/*
 * Run TCP/UDP service loop: Create listener for incoming connection
 * requests as a go-routine and call user-defined service handler as
 * a go-routine to handle client sessions. 
 *
 * (c) 2010 Bernd Fix   >Y<
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
	"os"
	"net"
	"log"
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
	Process (conn net.Conn)				// main handler routine
	GetName() string					// get symbolic name of service
	CanHandle (protocol string) bool	// check network protocol
	IsAllowed (remote string) bool		// check remote address
}

///////////////////////////////////////////////////////////////////////
// Public methods

/*
 * Run TCP/UDP network service with user-defined session handler.
 * @param network string - network identifier (TCP/UDP on IPv4/v6)
 * @param addr string - address:port specification of service
 * @param hdlr Service - implementation of service interface
 */
func Run (network, addr string, hdlr Service) os.Error {

	// initialize control service	
	service, err := net.Listen (network, addr)
	if err != nil {
		log.Println ("[" + hdlr.GetName() + "] service start-up failed: " + err.String())
		return err
	}
	
	// handle connection requests
	ch := make (chan net.Conn)
	go func() {
		for {
			// wait for connection request
			client, err := service.Accept()
			if err != nil {
				log.Println ("[" + hdlr.GetName() + "] Accept(): " + err.String())
				continue
			}
			// check for TCP connection from localhost
			remote  := client.RemoteAddr().String()
			protocol := client.RemoteAddr().Network()
			if !hdlr.CanHandle (protocol)  {
				log.Printf("[" + hdlr.GetName() + "] rejected non-TCP connection from %v\n", protocol)
				continue
			}
			if !hdlr.IsAllowed (remote)  {
				log.Printf("[" + hdlr.GetName() + "] rejected remote connection from %v\n", addr)
				continue
			}
			// connection accepted
			log.Printf("[" + hdlr.GetName() + "] accepted %v\n", addr)
			ch <- client
		}
	}()

	// handle incoming client connections	
	for {
		go hdlr.Process (<-ch)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 1.0  2010-11-18 23:17:06  brf
//  Initial revision.
//
