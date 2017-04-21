package rpc

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
