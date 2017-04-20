package util

import (
	"bytes"
	"github.com/bfix/gospel/math"
	"testing"
)

func TestBase58(t *testing.T) {
	if !test1(math.NewInt(57)) {
		t.Fatal("base58 failure")
	}
	if !test1(math.NewInt(58)) {
		t.Fatal("base58 failure")
	}
	if !test1(math.NewInt(255)) {
		t.Fatal("base58 failure")
	}
	if !test2([]byte{0, 255}) {
		t.Fatal("base58 failure")
	}
	if !test2([]byte{0, 0, 255}) {
		t.Fatal("base58 failure")
	}
	bound := math.NewInt(256)
	for n := 0; n < 128; n++ {
		if !test1(math.NewIntRndRange(math.ONE, bound)) {
			t.Fatal("base58 failure")
		}
		bound = bound.Lsh(1)
	}

	if _, err := Base58Decode("invalid"); err == nil {
		t.Fatal("base58 failure")
	}
}

func test1(x *math.Int) bool {
	s := Base58Encode(x.Bytes())
	b, err := Base58Decode(s)
	if err != nil {
		return false
	}
	y := math.NewIntFromBytes(b)
	return x.Equals(y)
}

func test2(x []byte) bool {
	s := Base58Encode(x)
	y, err := Base58Decode(s)
	if err != nil {
		return false
	}
	res := bytes.Equal(x, y)
	return res
}
