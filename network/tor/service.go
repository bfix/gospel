package tor

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix
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
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/bfix/gospel/logger"
)

//======================================================================
// Tor management using control port protocol as defined in
// https://github.com/torproject/torspec/blob/master/control-spec.txt
//======================================================================

// Error codes
var (
	ErrTorNoSocksPort = fmt.Errorf("No SocksPort found")
	ErrTorNotLocal    = fmt.Errorf("Tor service not local")
)

// Service instance to communicate commands (and responses) with a
// running Tor process.
type Service struct {
	host string        // Tor service host
	conn net.Conn      // connection to control port (or socket)
	rdr  *bufio.Reader // buffered reader for responses
}

// NewService instantiates a new Tor controller
func NewService(schema, endp string) (srv *Service, err error) {
	// create a new service instance
	srv = new(Service)
	// connect to control port or socket
	if srv.conn, err = net.Dial(schema, endp); err != nil {
		return
	}
	srv.rdr = bufio.NewReader(srv.conn)
	// determine the service host
	srv.host = "127.0.0.1"
	if schema != "unix" && strings.LastIndex(endp, ":") != -1 {
		srv.host, _, err = net.SplitHostPort(endp)
	}
	return
}

// Close connection to Tor process
func (s *Service) Close() error {
	return s.conn.Close()
}

// Authenticate client by password defined in torrc as
// "HashedControlPassword 16:..."
func (s *Service) Authenticate(auth string) error {
	cmd := fmt.Sprintf("AUTHENTICATE \"%s\"", auth)
	_, err := s.execute(cmd)
	return err
}

// GetSocksPort returns the best-matching SocksPort definition for a given
// set of flags (only works for local Tor services)
func (s *Service) GetSocksPort(flags ...string) (string, error) {
	// check for local service
	// get list of defined proxy ports
	list, err := s.GetConf("SocksPort")
	if err != nil {
		return "", err
	}
	entries, ok := list["SocksPort"]
	if !ok {
		return "", ErrTorNoSocksPort
	}
	// find best port match
	bestProxy := ""
	bestDiff := 1000
	bestOff := 1000
	eval := func(list []string, flags ...string) int {
		found := 0
		for _, flag := range flags {
			for _, e := range list {
				if e == flag {
					found++
					break
				}
			}
		}
		return len(flags) - found
	}
	for _, e := range entries {
		parts := strings.Split(e, " ")
		if diff := eval(parts[1:], flags...); diff <= bestDiff {
			off := len(parts) - diff - 1
			if off < bestOff {
				bestDiff = diff
				bestProxy = parts[0]
			}
		}
	}
	if len(bestProxy) == 0 {
		return "", ErrTorNoSocksPort
	}
	return bestProxy, nil
}

//----------------------------------------------------------------------
// Low-level service and helper methods.
//----------------------------------------------------------------------

// GetConf returns the configuration settings associated with a config string.
// Key/value pairs are returned in a map.
func (s *Service) GetConf(cfg string) (map[string][]string, error) {
	cmd := fmt.Sprintf("GETCONF %s", cfg)
	return s.execute(cmd)
}

// SetConf sets a configration from a key/value pair.
func (s *Service) SetConf(key, value string) error {
	cmd := fmt.Sprintf("SETCONF %s=\"%s\"", key, value)
	_, err := s.execute(cmd)
	return err
}

// SetConfList sets all settings from a list of key/value pairs
func (s *Service) SetConfList(cfg map[string]string) error {
	cmd := "SETCONF"
	for k, v := range cfg {
		cmd += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	_, err := s.execute(cmd)
	return err
}

// ResetConf resets a configuration to its default value
func (s *Service) ResetConf(cfg string) error {
	cmd := fmt.Sprintf("RESETCONF %s", cfg)
	_, err := s.execute(cmd)
	return err
}

// SetEvents requests the server to inform the client about interesting events.
func (s *Service) SetEvents(evs []string, extended bool) error {
	cmd := "SETEVENTS"
	if extended {
		cmd += " EXTENDED"
	}
	for _, e := range evs {
		cmd += " " + e
	}
	_, err := s.execute(cmd)
	return err
}

// SaveConf instructs the server to write out its config options into its torrc.
func (s *Service) SaveConf(force bool) error {
	cmd := "SAVECONF"
	if force {
		cmd += " FORCE"
	}
	_, err := s.execute(cmd)
	return err
}

// Signal the server for action.
func (s *Service) Signal(sig string) error {
	cmd := fmt.Sprintf("SIGNAL %s", sig)
	_, err := s.execute(cmd)
	return err
}

// MapAddress tells the Tor servce that future SOCKS requests for connections
// to the original address should be replaced with connections to the
// specified replacement address.
func (s *Service) MapAddress(addrs map[string]string) (map[string][]string, error) {
	cmd := "MAPADDRESS"
	for oldAddr, newAddr := range addrs {
		cmd += fmt.Sprintf(" %s=%s", oldAddr, newAddr)
	}
	return s.execute(cmd)
}

// GetInfo from Tor service for non-torrc values
func (s *Service) GetInfo(keys []string) (map[string][]string, error) {
	cmd := "GETINFO"
	for _, k := range keys {
		cmd += fmt.Sprintf(" %s", k)
	}
	return s.execute(cmd)
}

// Execute (send) a control command to Tor and collect response(s).
// Reponses are formatted like "<status><cont><text>" where "status"
// is "250" for success and any other three digit number for failure.
// If "cont" is "-", more response lines will follow.
// The method returns a list of response texts (without a final "OK").
func (s *Service) execute(cmd string) (list map[string][]string, err error) {
	// send command
	logger.Printf(logger.DBG, "[TorService] <<< %s\n", cmd)
	if _, err = s.conn.Write([]byte(cmd + "\n")); err != nil {
		return
	}
	// read responses
	var (
		line []byte
		rc   int64
		resp []string
	)
	done := false
	for !done {
		// read next response line
		if line, _, err = s.rdr.ReadLine(); err != nil {
			if err != io.EOF {
				return
			}
			done = true
		}
		out := strings.Trim(string(line), " \t\n\v\r")
		logger.Printf(logger.DBG, "[TorService] >>> %s\n", out)
		// check for error status
		if rc, err = strconv.ParseInt(out[:3], 10, 32); err != nil {
			return
		}
		if rc != 250 {
			err = fmt.Errorf(out)
			return
		}
		// check for multi-line response
		tag := out[3]
		if tag == '+' {
			out = out[4:]
			for {
				// read next line
				if line, _, err = s.rdr.ReadLine(); err != nil {
					return
				}
				frag := string(line)
				// check for end-of-data
				if frag == "." {
					break
				}
				out += "\n" + frag
			}
		} else {
			out = out[4:]
		}
		// check for non-trivial content
		if out != "OK" {
			resp = append(resp, out)
		}
		// check for continuation
		if tag == ' ' {
			// no continuation
			break
		}
	}
	// build result list
	list = make(map[string][]string)
	for _, pair := range resp {
		parts := strings.SplitN(pair, "=", 2)
		entries, ok := list[parts[0]]
		if !ok {
			entries = make([]string, 0)
		}
		entry := ""
		if len(parts) == 2 {
			entry = parts[1]
		}
		list[parts[0]] = append(entries, entry)
	}
	return
}
