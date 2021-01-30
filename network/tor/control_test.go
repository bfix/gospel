package network

import (
	"os"
	"testing"
)

var (
	testCtrl *Control = nil
	err      error
)

func TestTorControl(t *testing.T) {
	proto := os.Getenv("TOR_CONTROL_PROTO")
	if len(proto) == 0 {
		proto = "tcp"
	}
	endp := os.Getenv("TOR_CONTROL_ENDPOINT")
	if len(endp) == 0 {
		endp = "127.0.0.1:9052"
	}
	passwd := os.Getenv("TOR_CONTROL_PASSWORD")
	if len(passwd) == 0 {
		t.Skip("Skipping 'network/tor' tests!")
	}
	testCtrl, err = NewControl(proto, endp)
	if err != nil {
		t.Fatal(err)
	}
	if err = testCtrl.Authenticate(passwd); err != nil {
		t.Fatal(err)
	}
}

func TestGetConf(t *testing.T) {

}
