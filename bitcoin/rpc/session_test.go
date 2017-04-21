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
		return
	}
	strictCheck = true
	info, err = sess.GetInfo()
	if err != nil {
		sess = nil
		fmt.Println("ERROR: " + err.Error())
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

func TestSession(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
}
