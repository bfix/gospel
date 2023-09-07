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
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/logger"
)

var (
	srv       *Service                    // service instance
	passwd    string                      // password for authentication
	testHost  string                      // host running hidden test server
	err       error                       // last error code
	socksPort = make(map[string][]string) // port mappings
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
	if len(endp) == 0 {
		endp = "127.0.0.1:9051"
	}
	testHost = os.Getenv("TOR_TEST_HOST")
	if len(testHost) == 0 {
		testHost = "127.0.0.1"
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
	conn, err := srv.DialTimeout("tcp", "ipify.org:80", time.Minute)
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

func TestDialOnion(t *testing.T) {
	// connect to Riseup through Tor
	conn, err := srv.DialTimeout("tcp", "vww6ybal4bd7szmgncyruucpgfkqahzddi37ktceo3ah7ngmcopnpyyd.onion:80", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	// get web page
	if _, err = conn.Write([]byte("GET /\n\n")); err != nil {
		t.Fatal(err)
	}
	if _, err = ioutil.ReadAll(conn); err != nil {
		t.Fatal(err)
	}
	// close connection
	if err = conn.Close(); err != nil {
		t.Fatal(err)
	}
}

//----------------------------------------------------------------------
// Hidden service tests (onion.go)
//----------------------------------------------------------------------

func TestOnion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping onion test in short mode.")
	}
	// start a simple echo server
	var echoErr error
	go func() {
		listener, err := net.Listen("tcp", "0.0.0.0:12345") //nolint:gosec // intentional
		if err != nil {
			echoErr = err
			return
		}
		conn, err := listener.Accept()
		if err != nil {
			echoErr = err
			return
		}
		defer conn.Close()
		rdr := bufio.NewReader(conn)
		data, err := rdr.ReadBytes(byte('\n'))
		if err != nil {
			echoErr = err
			return
		}
		_, echoErr = conn.Write(data)
	}()
	// start a hidden service
	_, prv := ed25519.NewKeypair()
	hs, err := NewOnion(prv)
	if err != nil {
		t.Fatal(err)
	}
	hs.AddPort(80, testHost+":12345")
	if err = hs.Start(srv); err != nil {
		t.Fatal(err)
	}
	host, err := hs.ServiceID()
	if err != nil {
		t.Fatal(err)
	}
	host += ".onion:80"
	// wait for hidden service to settle down
	t.Log("Waiting 60\" for hidden service to settle down...")
	time.Sleep(60 * time.Second)
	// connect to echo server through Tor to website
	t.Logf("Connecting to '%s'\n", host)
	conn, err := srv.DialTimeout("tcp", host, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write([]byte("TEST\n")); err != nil {
		t.Fatal(err)
	}
	var data []byte
	if data, err = ioutil.ReadAll(conn); err != nil {
		t.Fatal(err)
	}
	res := strings.TrimSpace(string(data))
	if res != "TEST" {
		t.Fatalf("Received '%s' instead of 'TEST'\n", res)
	}
	// close connection
	if err = conn.Close(); err != nil {
		t.Fatal(err)
	}
	// stop hidden service
	if err = hs.Stop(srv); err != nil {
		t.Fatal(err)
	}
	// check echo server status
	if echoErr != nil {
		t.Fatal(echoErr)
	}
}
