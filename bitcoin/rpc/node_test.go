package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"testing"
)

func TestNetTotals(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	nt, err := sess.GetNetTotals()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("NetworkTotals: %s\n", nt)
	}
}

func TestMiningInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	mi, err := sess.GetMiningInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("MiningInfo: %s\n", mi)
	}
}

func TestNetworkHashPS(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	_, err := sess.GetNetworkHashPS(120, -1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNetworkInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	ni, err := sess.GetNetworkInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("NetworkInfo: %s\n", ni)
	}
}

func TestPeerInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	pi, err := sess.GetPeerInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("PeerInfo: %s\n", pi)
	}
}

func TestListBanned(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	bl, err := sess.ListBanned()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("BanList: %s\n", bl)
	}
}

func TestPing(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err := sess.Ping(); err != nil {
		t.Fatal(err)
	}
}
