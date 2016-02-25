package util

import (
	"math/rand"
	"testing"
)

var raw = "01000000017b1eabe0209b1fe794124575ef807057c77ada2138ae4fa8d6c4de" +
	"0398a14f3f0000000000ffffffff01f0ca052a010000001976a914cbc20a7664" +
	"f2f69e5355aa427045bc15e7c6c77288ac00000000"

func TestScript(t *testing.T) {
	testScr := func(n int) error {
		b := make([]byte, n)
		n, err := rand.Read(b)
		if err != nil || n != len(b) {
			t.Fatal()
		}
		_, err = NullDataScript(b)
		return err
	}
	if testScr(64) != nil {
		t.Fatal()
	}
	if testScr(128) == nil {
		t.Fatal()
	}
}

func TestPrefix(t *testing.T) {
	check := func(n, s int) {
		p := LengthPrefix(n)
		if len(p) != s {
			t.Fatal()
		}
	}
	check(64, 1)
	check(128, 2)
	check(512, 3)
	check(72000, 5)
}

func TestRaw(t *testing.T) {
	b := rnd(t, 64)
	s, err := NullDataScript(b)
	if err != nil {
		t.Fatal()
	}
	// OK
	if _, err = ReplaceScriptPubKey(raw, s); err != nil {
		t.Fatal()
	}
	// #vout != 1
	r := raw[:9] + "2" + raw[10:]
	if _, err = ReplaceScriptPubKey(r, s); err == nil {
		t.Fatal()
	}
	// sigSize != 0
	r = raw[:83] + "1" + raw[84:]
	if _, err = ReplaceScriptPubKey(r, s); err == nil {
		t.Fatal()
	}
	// #vout != 1
	r = raw[:93] + "2" + raw[94:]
	if _, err = ReplaceScriptPubKey(r, s); err == nil {
		t.Fatal()
	}
	// invalid raw
	if _, err = ReplaceScriptPubKey("invalid", s); err == nil {
		t.Fatal()
	}
}

func rnd(t *testing.T, n int) []byte {
	b := make([]byte, n)
	i, err := rand.Read(b)
	if err != nil || i != n {
		t.Fatal()
	}
	return b
}
