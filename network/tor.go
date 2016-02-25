package network

import (
	"net"
	"strconv"
)

// IsTorExitToDest checks if source is a TOR exit node that can
// connect to dst:dport
func IsTorExitToDest(src, dst net.IP, dport int) bool {
	name := revAddr(src) + "." + strconv.Itoa(dport) + "." + revAddr(dst)
	return checkTor(name)
}

// IsTorExit checks if source is a TOR exit node
func IsTorExit(src net.IP) bool {
	return checkTor(revAddr(src) + ".80.1.2.3.4")
}

// Check if a TOR exit node is specified in name.
// name is a dotted list of "<src>.<port>.<dst>" where all addresses
// are IPv4 addresses.
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
