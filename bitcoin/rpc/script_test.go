package rpc

import (
	"testing"
)

func TestPushData(t *testing.T) {
	check := func(n, s int) {
		b := PushData(make([]byte, n))
		if len(b)-n != s {
			t.Fatal("PushData failed")
		}
	}
	check(64, 1)
	check(128, 2)
	check(512, 3)
	check(72000, 5)
}
