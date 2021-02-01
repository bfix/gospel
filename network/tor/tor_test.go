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
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bfix/gospel/logger"
)

var (
	srv       *Service = nil
	passwd    string
	torProxy  string
	err       error
	socksPort = make(map[string][]string)
)

//----------------------------------------------------------------------
// Main test entry point
//----------------------------------------------------------------------

func TestMain(m *testing.M) {
	logger.SetLogLevel(logger.INFO)
	rc := 0
	defer func() {
		os.Exit(rc)
	}()

	// handle environment variables
	proto := os.Getenv("TOR_CONTROL_PROTO")
	if len(proto) == 0 {
		proto = "tcp"
	}
	endp := os.Getenv("TOR_CONTROL_ENDPOINT")
	isLocal := (proto == "unix")
	if len(endp) == 0 {
		endp = "127.0.0.1:9051"
		isLocal = true
	} else {
		// check for local service instance
		host, _, err := net.SplitHostPort(endp)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			rc = 1
			return
		}
		isLocal = isLocal || (host == "localhost" || host == "127.0.0.1")
	}
	if passwd = os.Getenv("TOR_CONTROL_PASSWORD"); len(passwd) == 0 {
		fmt.Println("Skipping 'network/tor' tests!")
		return
	}
	// instaniate new service for tests
	srv, err = NewService(proto, endp)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		rc = 1
		return
	}
	// determine Tor proxy spec
	torProxy = os.Getenv("TOR_PROXY")
	if len(torProxy) == 0 {
		if isLocal {
			proxy, err := srv.GetSocksPort()
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				rc = 1
				return
			}
			_, port, err := net.SplitHostPort(proxy)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				rc = 1
				return
			}
			torProxy = "socks5://127.0.0.1:" + port
		} else {
			torProxy = "socks5://127.0.0.1:9050"
		}
	}

	// run test cases
	rc = m.Run()

	// clean-up
	if err = srv.Close(); err != nil {
		rc = 1
	}
}

//----------------------------------------------------------------------
// Service test (service.go)
//----------------------------------------------------------------------

func TestAuthentication(t *testing.T) {
	if err = srv.Authenticate(passwd); err != nil {
		t.Fatal(err)
	}
}

func TestGetConf(t *testing.T) {
	list, err := srv.GetConf("SocksPort")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Result:")
	for k, v := range list {
		t.Logf(">>> '%s':\n", k)
		for _, e := range v {
			t.Logf(">>>     '%s'\n", e)
			parts := strings.Split(e, " ")
			socksPort[parts[0]] = parts[1:]
		}
	}
}

func TestSocksPort(t *testing.T) {
	for proxy, flags := range socksPort {
		found, err := srv.GetSocksPort(flags...)
		if err != nil {
			t.Fatal(err)
		}
		if found != proxy {
			t.Fatalf("Proxy mismatch: %s != %s\n", proxy, found)
		}
	}
}

//----------------------------------------------------------------------
// Connection tests (conn.go)
//----------------------------------------------------------------------

func TestDial(t *testing.T) {
	// connect through Tor to website
	conn, err := DialTimeout("tcp", "ipify.org:80", time.Minute, torProxy)
	if err != nil {
		t.Fatal(err)
	}
	// get my IP address
	if _, err = conn.Write([]byte("GET /\n\n")); err != nil {
		t.Fatal(err)
	}
	var data []byte
	if data, err = ioutil.ReadAll(conn); err != nil {
		t.Fatal(err)
	}
	// check for Tor exit node
	if !IsTorExit(net.ParseIP(string(data))) {
		t.Fatal("Invalid exit node address")
	}
	// close connection
	if err = conn.Close(); err != nil {
		t.Fatal(err)
	}
}
