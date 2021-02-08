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
	"fmt"
	"net"

	upnp "github.com/huin/goupnp/dcps/internetgateway2"
)

// Port mapping modes
const (
	PmNONE = iota
	PmDIRECT
	PmUPNP
	PmSTUN
)

// Error messages
var (
	ErrPortMapperInit    = fmt.Errorf("Port mapper initialized")
	ErrPortMapperNoInit  = fmt.Errorf("Port mapper not initialized")
	ErrPortMapperConfig  = fmt.Errorf("Can't configure port mapper")
	ErrPortMapperUnknown = fmt.Errorf("Unknown port mapping")
)

// Mapping of an "external" port to "internal" port for a network protocol.
// The IP addresses (external, internal) are the same for all mappings
type Mapping struct {
	netw string
	port int
}

// PortMapper implements a mapping between an external (globally visible)
// service address (ip:port) to an internal address.
type PortMapper struct {
	mode       int                    // PM_???
	name       string                 // network name
	extIP      net.IP                 // external IP address
	lclIP      net.IP                 // local listener address
	server     net.IP                 // involved server (gateway, stun)
	upnpClient *upnp.WANIPConnection2 // UPNP client connection
	lastID     int                    // last identifier used
	assigns    map[string]*Mapping    // port mappings
}

var prvBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			prvBlocks = append(prvBlocks, block)
		}
	}
}

func isRoutable(addr net.Addr) bool {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	for _, block := range prvBlocks {
		if block.Contains(ip) {
			return false
		}
	}
	return true
}

// NewPortMapper instaniates a new port mapping mechanism with given name.
func NewPortMapper(name string) (*PortMapper, error) {
	pm := new(PortMapper)
	pm.assigns = make(map[string]*Mapping)
	pm.name = name

	//------------------------------------------------------------------
	// (1) check local interfaces for global address
	//------------------------------------------------------------------
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && isRoutable(addr) {
			pm.mode = PmDIRECT
			pm.extIP = ip.IP
			pm.lclIP = nil
			return pm, nil
		}
	}
	//------------------------------------------------------------------
	// (2) try UPNP port forwarding
	//------------------------------------------------------------------
	clients, _, err := upnp.NewWANIPConnection2Clients()
	if err != nil {
		return nil, err
	}
	for _, c := range clients {
		pm.upnpClient = c
		host, _, _ := net.SplitHostPort(c.ServiceClient.Location.Host)
		pm.server = net.ParseIP(host)
		srv := c.ServiceClient.Service
		scpd, err := srv.RequestSCPD()
		if err != nil {
			continue
		}
		if scpd == nil || (scpd.GetAction("GetExternalIPAddress") != nil && scpd.GetAction("AddPortMapping") != nil) {
			ip, err := c.GetExternalIPAddress()
			if err == nil {
				pm.extIP = net.ParseIP(ip)
				for _, addr := range addrs {
					if ipn, ok := addr.(*net.IPNet); ok {
						if ipn.Contains(pm.server) {
							pm.lclIP = ipn.IP
							pm.mode = PmUPNP
							return pm, nil
						}
					}
				}
			}
		}
	}
	return nil, ErrPortMapperConfig
}

// Assign a port mapping for a given port and protocol.
// Returns the mapping identifier, external and internal service addresses
// and an optional error code.
func (pm *PortMapper) Assign(network string, port int) (string, string, string, error) {
	if pm.mode == PmNONE {
		return "", "", "", ErrPortMapperNoInit
	}
	concat := func(ip net.IP, port int) string {
		if ip.To4() == nil {
			return fmt.Sprintf("[%s]:%d", ip.String(), port)
		}
		return fmt.Sprintf("%s:%d", ip.String(), port)
	}
	if pm.mode == PmDIRECT {
		ext := concat(pm.extIP, port)
		return "", ext, ext, nil
	}
	pm.lastID++
	descr := fmt.Sprintf("%s:%d", pm.name, pm.lastID)
	err := pm.upnpClient.AddPortMapping("", uint16(port), network, uint16(port), pm.lclIP.String(), true, descr, 0)
	ext := concat(pm.extIP, port)
	lcl := concat(pm.lclIP, port)
	pm.assigns[descr] = &Mapping{
		netw: network,
		port: port,
	}
	return descr, ext, lcl, err
}

// Unassign removes a port mapping
func (pm *PortMapper) Unassign(id string) error {
	if pm.mode == PmNONE {
		return ErrPortMapperNoInit
	}
	if pm.mode == PmDIRECT {
		return nil
	}
	if m, ok := pm.assigns[id]; ok {
		if err := pm.upnpClient.DeletePortMapping("", uint16(m.port), m.netw); err != nil {
			return err
		}
		delete(pm.assigns, id)
		return nil
	}
	return ErrPortMapperUnknown
}

// Close port mapper
func (pm *PortMapper) Close() error {
	if pm.mode == PmNONE {
		return ErrPortMapperNoInit
	}
	if pm.mode == PmUPNP {
		for id := range pm.assigns {
			if err := pm.Unassign(id); err != nil {
				return err
			}
		}
	}
	pm.mode = PmNONE
	return nil
}
