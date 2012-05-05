/*
 * TOR (The Onion Router) helper methods. 
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
	"net"
	"strconv"
)

///////////////////////////////////////////////////////////////////////
// Public functions

/*
 * Check if source is a TOR exit node that can connect to dst:dport
 * @param src net.IP - source address to be checked for TOR exit node 
 * @param dst net.IP - destination address (usually addr of this instance) 
 * @param dport int - destination port (usually port of this instance) 
 * @return bool - TOR exit node involved?
 */
func IsTorExitToDest(src, dst net.IP, dport int) bool {
	name := revAddr(src) + "." + strconv.Itoa(dport) + "." + revAddr(dst)
	return checkTor(name)
}

//---------------------------------------------------------------------
/*
 * Check if source is a TOR exit node
 * @param src net.IP - source address to be checked for TOR exit node 
 * @return bool - TOR exit node involved?
 */
func IsTorExit(src net.IP) bool {
	return checkTor(revAddr(src) + ".80.1.2.3.4")
}

///////////////////////////////////////////////////////////////////////
// Helper functions

/*
 * Check if a TOR exit node is specified in name.
 * name is a dotted list of "<src>.<port>.<dst>" where all addresses
 * are IPv4 addresses.
 * @param name string - TOR specification
 * @return bool - TOR exit node involved?
 */
func checkTor(name string) bool {
	addrs, err := net.LookupHost(name + ".ip-port.exitlist.torproject.org")
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if addr == "127.0.0.2" {
			return true
		}
	}
	return false
}

//---------------------------------------------------------------------
/*
 * Get reversed order IPv4 address.
 * @param ip net.IP - network address (a.b.c.d)
 * @return string - reversed address ("d.c.b.a")
 */
func revAddr(ip net.IP) string {
	if addr := ip.To4(); addr != nil {
		return strconv.Itoa(int(addr[3])) + "." +
			strconv.Itoa(int(addr[2])) + "." +
			strconv.Itoa(int(addr[1])) + "." +
			strconv.Itoa(int(addr[0]))
	}
	return "0.0.0.0"
}