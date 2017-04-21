package rpc

import (
	"fmt"
	"testing"
)

func TestConnectionCount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	conns, err := sess.GetConnectionCount()
	if err != nil {
		t.Fatal(err)
	}
	if conns != info.Connections {
		t.Fatal(fmt.Sprintf("session-count mismatch: %d != %d", conns, info.Connections))
	}
}

func TestDifficulty(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	diff, err := sess.GetDifficulty()
	if err != nil {
		t.Fatal(err)
	}
	if diff != info.Difficulty {
		t.Fatal("difficulty mismatch in info")
	}
}

func TestFee(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err := sess.SetTxFee(0.0001); err != nil {
		t.Fatal(err)
	}
}

func TestMemPoolInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	mi, err := sess.GetMemPoolInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("MemPoolInfo: %s\n", mi)
	}
}

func TestGetRawMemPool(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	list, err := sess.GetRawMemPoolList()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("RawMemPoolList: %s\n", list)
	}
	tx, err := sess.GetRawMemPool()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("RawMemPool: %s\n", tx)
	}
	for _, k := range list {
		if _, ok := tx[k]; !ok {
			t.Fatal(fmt.Sprintf("Unknown key '%s'", k))
		}
	}
}
