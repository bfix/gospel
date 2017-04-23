package rpc

// AddNode attempts to add or remove a node from the addnode list, or to
// try a connection to a node once.
// The node to add as a string in the form of <IP address>:<port>. The IP
// address may be a hostname resolvable through DNS, an IPv4 address, an
// IPv4-as-IPv6 address, or an IPv6 address.
// Command specifies what to do with the IP address:
// - "add" to add a node to the addnode list. This will not connect
//   immediately if the outgoing connection slots are full.
// - "remove" to remove a node from the list. If currently connected,
//   this will disconnect immediately.
func (s *Session) AddNode(addr, command string) error {
	_, err := s.call("addnode", []Data{addr, command})
	return err
}

// ClearBanned clears list of banned nodes.
func (s *Session) ClearBanned() error {
	_, err := s.call("clearbanned", []Data{})
	return err
}

// DisconnectNode disconnects a node.
func (s *Session) DisconnectNode(addr string) error {
	_, err := s.call("disconnectnode", []Data{addr})
	return err
}

// GetAddedNodeInfo returns information about the given added node, or all
// added nodes (except onetry nodes). Only nodes which have been manually
// added using the addnode RPC will have their information displayed.
func (s *Session) GetAddedNodeInfo(detail bool, addr string) ([]*NodeInfo, error) {
	data := []Data{detail}
	if len(addr) > 0 {
		data = append(data, addr)
	}
	res, err := s.call("getaddednodeinfo", data)
	if err != nil {
		return nil, err
	}
	var list []*NodeInfo
	if ok, err := res.UnmarshalResult(&list); !ok {
		return nil, err
	}
	return list, nil
}

// GetConnectionCount returns the number of connections to other nodes.
func (s *Session) GetConnectionCount() (int, error) {
	res, err := s.call("getconnectioncount", nil)
	if err != nil {
		return -1, err
	}
	return int(res.Result.(float64)), nil
}

// GetMiningInfo returns returns various mining-related information.
func (s *Session) GetMiningInfo() (*MiningInfo, error) {
	res, err := s.call("getmininginfo", nil)
	if err != nil {
		return nil, err
	}
	mi := new(MiningInfo)
	if ok, err := res.UnmarshalResult(mi); !ok {
		return nil, err
	}
	return mi, nil
}

// GetNetTotals returns information about network traffic, including bytes in,
// bytes out, and the current time.
func (s *Session) GetNetTotals() (*NetworkStats, error) {
	res, err := s.call("getnettotals", nil)
	if err != nil {
		return nil, err
	}
	nt := new(NetworkStats)
	if ok, err := res.UnmarshalResult(nt); !ok {
		return nil, err
	}
	return nt, nil
}

// GetNetworkHashPS returns the estimated current or historical network
// hashes per second based on the last n blocks.
func (s *Session) GetNetworkHashPS(blocks, height int) (float64, error) {
	res, err := s.call("getnetworkhashps", []Data{blocks, height})
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// GetNetworkInfo returns information about the node’s connection to
// the network.
func (s *Session) GetNetworkInfo() (*NetworkInfo, error) {
	res, err := s.call("getnetworkinfo", nil)
	if err != nil {
		return nil, err
	}
	ni := new(NetworkInfo)
	if ok, err := res.UnmarshalResult(ni); !ok {
		return nil, err
	}
	return ni, nil
}

// GetPeerInfo returns data about each connected network node.
func (s *Session) GetPeerInfo() ([]*PeerInfo, error) {
	res, err := s.call("getpeerinfo", nil)
	if err != nil {
		return nil, err
	}
	var pi []*PeerInfo
	if ok, err := res.UnmarshalResult(&pi); !ok {
		return nil, err
	}
	return pi, nil
}

// ListBanned lists all banned IPs/Subnets.
func (s *Session) ListBanned() ([]*BannedNode, error) {
	res, err := s.call("listbanned", nil)
	if err != nil {
		return nil, err
	}
	var bl []*BannedNode
	if ok, err := res.UnmarshalResult(&bl); !ok {
		return nil, err
	}
	return bl, nil
}

// Ping sends a P2P ping message to all connected nodes to measure ping time.
// Results are provided by the getpeerinfo RPC pingtime and pingwait fields
// as decimal seconds. The P2P ping message is handled in a queue with all
// other commands, so it measures processing backlog, not just network ping.
func (s *Session) Ping() error {
	_, err := s.call("ping", nil)
	return err
}

// SetBan attempts add or remove a IP/Subnet from the banned list.
// Argument is the node to add or remove as a string in the form of
// <IP address>. The IP address may be a hostname resolvable through DNS, an
// IPv4 address, an IPv4-as-IPv6 address, or an IPv6 address.
func (s *Session) SetBan(addr, cmd string, banTime int) error {
	_, err := s.call("setban", []Data{addr, cmd, banTime})
	return err
}
