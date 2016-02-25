package util

import (
	"bytes"
	"github.com/bfix/gospel/crypto"
	"math/big"
	"testing"
)

func TestBase58(t *testing.T) {
	one := big.NewInt(1)
	if !test1(big.NewInt(57)) {
		t.Fatal()
	}
	if !test1(big.NewInt(58)) {
		t.Fatal()
	}
	if !test1(big.NewInt(255)) {
		t.Fatal()
	}
	if !test2([]byte{0, 255}) {
		t.Fatal()
	}
	if !test2([]byte{0, 0, 255}) {
		t.Fatal()
	}
	bound := big.NewInt(256)
	for n := 0; n < 128; n++ {
		if !test1(crypto.RandBigInt(one, bound)) {
			t.Fatal()
		}
		bound = new(big.Int).Lsh(bound, 1)
	}

	if _, err := Base58Decode("invalid"); err == nil {
		t.Fatal()
	}
}

func test1(x *big.Int) bool {
	s := Base58Encode(x.Bytes())
	b, err := Base58Decode(s)
	if err != nil {
		return false
	}
	y := new(big.Int).SetBytes(b)
	res := x.Cmp(y) == 0
	return res
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
