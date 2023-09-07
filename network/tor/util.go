package tor

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
	"net"
	"strconv"
)

//======================================================================
// Tor utility functions
//======================================================================

// IsTorExit checks if source is a TOR exit node
func IsTorExit(src net.IP) bool {
	return checkTor(revAddr(src))
}

// Check if a TOR exit node is specified in name.
// name is a dotted list of "<src>.<port>.<dst>" where all addresses
// are IPv4 addresses.
func checkTor(name string) bool {
	addrs, err := net.LookupHost(name + ".dnsel.torproject.org")
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

// Get reversed order IPv4 address.
func revAddr(ip net.IP) string {
	if addr := ip.To4(); addr != nil {
		return strconv.Itoa(int(addr[3])) + "." +
			strconv.Itoa(int(addr[2])) + "." +
			strconv.Itoa(int(addr[1])) + "." +
			strconv.Itoa(int(addr[0]))
	}
	return "0.0.0.0"
}
