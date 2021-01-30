package network

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
	"net"
	"strconv"
	"strings"

	"github.com/bfix/gospel/logger"
)

//======================================================================
// Tor management using control port protocol as defined in
// https://github.com/torproject/torspec/blob/master/control-spec.txt
//======================================================================

// Control instance to communicate commands (and responses) with a
// running Tor process.
type Control struct {
	conn net.Conn      // connection to control port (or socket)
	rdr  *bufio.Reader // buffered reader for responses
}

// NewTorControl instantiates a new Tor controller
func NewControl(schema, endp string) (*Control, error) {
	// connect to control port or socket
	conn, err := net.Dial(schema, endp)
	if err != nil {
		return nil, err
	}
	// instantiate controller object
	rdr := bufio.NewReader(conn)
	tc := &Control{
		conn: conn,
		rdr:  rdr,
	}
	return tc, nil
}

// Close connection to Tor process
func (c *Control) Close() error {
	return c.conn.Close()
}

// Authenticate client by password defined in torrc as
// "HashedControlPassword 16:..."
func (c *Control) Authenticate(auth string) error {
	cmd := fmt.Sprintf("AUTHENTICATE \"%s\"", auth)
	_, err := c.execute(cmd)
	return err
}

// GetConf returns the configuration settings associated with a config string.
// Key/value pairs are returned in a map.
func (c *Control) GetConf(cfg string) (map[string]string, error) {
	cmd := fmt.Sprintf("GETCONF %s", cfg)
	return c.execute(cmd)
}

// SetConf sets a configration from a key/value pair.
func (c *Control) SetConf(key, value string) error {
	cmd := fmt.Sprintf("SETCONF %s=\"%s\"", key, value)
	_, err := c.execute(cmd)
	return err
}

// SetConfList sets all settings from a list of key/value pairs
func (c *Control) SetConfList(cfg map[string]string) error {
	cmd := "SETCONF"
	for k, v := range cfg {
		cmd += fmt.Sprintf(" %s=\"%s\"", k, v)
	}
	_, err := c.execute(cmd)
	return err
}

// ResetConf resets a configration to its default value
func (c *Control) ResetConf(cfg string) error {
	cmd := fmt.Sprintf("RESETCONF %s", cfg)
	_, err := c.execute(cmd)
	return err
}

// SetEvents requests the server to inform the client about interesting events.
func (c *Control) SetEvents(evs []string, extended bool) error {
	cmd := "SETEVENTS"
	if extended {
		cmd += " EXTENDED"
	}
	for _, e := range evs {
		cmd += " " + e
	}
	_, err := c.execute(cmd)
	return err
}

// SaveConf instructs the server to write out its config options into its torrc.
func (c *Control) SaveConf(force bool) error {
	cmd := "SAVECONF"
	if force {
		cmd += " FORCE"
	}
	_, err := c.execute(cmd)
	return err
}

// Signal the server for action.
func (c *Control) Signal(sig string) error {
	cmd := fmt.Sprintf("SIGNAL %s", sig)
	_, err := c.execute(cmd)
	return err
}

// MapAddress tells the Tor servce that future SOCKS requests for connections
// to the original address should be replaced with connections to the
// specified replacement address.
func (c *Control) MapAddress(addrs map[string]string) (map[string]string, error) {
	cmd := "MAPADDRESS"
	for oldAddr, newAddr := range addrs {
		cmd += fmt.Sprintf(" %s=%s", oldAddr, newAddr)
	}
	return c.execute(cmd)
}

// GetInfo from Tor service for non-torrc values
func (c *Control) GetInfo(keys []string) (map[string]string, error) {
	cmd := "GETINFO"
	for _, k := range keys {
		cmd += fmt.Sprintf(" %s", k)
	}
	return c.execute(cmd)
}

// Execute (send) a control command to Tor and collect response(s).
// Reponses are formatted like "<status><cont><text>" where "status"
// is "250" for success and any other three digit number for failure.
// If "cont" is "-", more response lines will follow.
// The method returns a list of response texts (without a final "OK").
func (c *Control) execute(cmd string) (list map[string]string, err error) {
	// send command
	logger.Printf(logger.DBG, "[TorControl] <<< %s\n", cmd)
	if _, err = c.conn.Write([]byte(cmd + "\n")); err != nil {
		return
	}
	// read responses
	var (
		line []byte
		rc   int64
		resp []string
	)
	for {
		// read next reponse line
		if line, _, err = c.rdr.ReadLine(); err != nil {
			return
		}
		out := strings.Trim(string(line), " \t\n\v\r")
		logger.Printf(logger.DBG, "[TorControl] >>> %s\n", out)
		// check for error status
		if rc, err = strconv.ParseInt(out[:3], 10, 32); err != nil {
			return
		}
		if rc != 250 {
			err = fmt.Errorf(out)
			return
		}
		// check for multi-line response
		if out[3] == '+' {
			out = out[4:]
			for {
				// read next line
				if line, _, err = c.rdr.ReadLine(); err != nil {
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
		if out[3] == ' ' {
			// no continuation
			break
		}
	}
	for _, pair := range resp {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			list[parts[0]] = parts[1]
		} else {
			list[parts[0]] = ""
		}
	}
	return
}
