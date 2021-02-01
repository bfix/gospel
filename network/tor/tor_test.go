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
	tor_proxy string
	err       error
	socksPort = make(map[string][]string)
)

//----------------------------------------------------------------------
// Main test entry point
//----------------------------------------------------------------------

func TestMain(m *testing.M) {
	logger.SetLogLevel(logger.INFO)

	proto := os.Getenv("TOR_CONTROL_PROTO")
	if len(proto) == 0 {
		proto = "tcp"
	}
	endp := os.Getenv("TOR_CONTROL_ENDPOINT")
	if len(endp) == 0 {
		endp = "127.0.0.1:9051"
	}
	tor_proxy = os.Getenv("TOR_PROXY")
	if len(tor_proxy) == 0 {
		tor_proxy = "127.0.0.1:9050"
	}
	if passwd = os.Getenv("TOR_CONTROL_PASSWORD"); len(passwd) == 0 {
		fmt.Println("Skipping 'network/tor' tests!")
		return
	}
	srv, err = NewService(proto, endp)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}
	rc := m.Run()
	srv.Close()
	os.Exit(rc)
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
	if !srv.isLocal {
		t.Skip("Skipping SocksPort test on non-local service")
	}
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
	var (
		conn net.Conn
		err  error
	)
	if srv.isLocal {
		conn, err = srv.DialTimeout("tcp", "ipify.org:80", time.Minute)
	} else {
		conn, err = srv.DialTimeout("tcp", "ipify.org:80", time.Minute, tor_proxy)
	}
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
