package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

const (
	verbose = false
)

var (
	sess *Session
	err  error
	info *Info
)

func init() {
	rpcaddr := os.Getenv("BTC_HOST")
	user := os.Getenv("BTC_USER")
	passwd := os.Getenv("BTC_PASSWORD")
	sess, err = NewSession(rpcaddr, user, passwd)
	if err != nil {
		sess = nil
	}
	info, err = sess.GetInfo()
	if err != nil {
		sess = nil
		fmt.Printf("ERROR: " + err.Error())
	} else if verbose {
		dumpObj("Info: %s\n", info)
	}
}

func dumpObj(fmtStr string, v interface{}) {
	data, err := json.Marshal(v)
	if err == nil {
		fmt.Printf(fmtStr, string(data))
	} else {
		fmt.Printf(fmtStr, "<"+err.Error()+">")
	}
}

func TestConnectionCount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	conns, err := sess.GetConnectionCount()
	if err != nil {
		t.Fatal("getsessioncount failed")
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
		t.Fatal("getdifficulty failed")
	}
	if diff != info.Difficulty {
		t.Fatal("difficulty mismatch in info")
	}
}

func TestFee(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err = sess.SetTxFee(0.0001); err != nil {
		t.Fatal("settxfee failed")
	}
}
