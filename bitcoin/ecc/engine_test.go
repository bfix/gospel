package ecc

import (
	"github.com/bfix/gospel/math"
	"testing"
)

func TestEngine(t *testing.T) {

	for i := 0; i < 32; i++ {
		prv := GenerateKeys(i&1 == 1)
		hash := nRnd(math.ONE).Bytes()
		r, s := Sign(prv, hash)
		if !Verify(&prv.PublicKey, hash, r, s) {
			t.Fatal("sign/verify failed")
		}
	}
}

func TestHash(t *testing.T) {
	i := nRnd(math.ONE)
	h := i.Bytes()
	j := convertHash(h)
	if i.Cmp(j) != 0 {
		t.Fatal("convertHash failed")
	}
}
