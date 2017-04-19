package rpc

import ()

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
	if err = res.UnmarshalResult(&list); err != nil {
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
