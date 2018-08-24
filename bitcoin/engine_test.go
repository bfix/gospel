package bitcoin

import (
	"testing"

	"github.com/bfix/gospel/math"
)

func TestEngine(t *testing.T) {
	for i := 0; i < 32; i++ {
		prv := GenerateKeys(i&1 == 1)
		hash := nRnd(math.ONE).Bytes()
		sig := Sign(prv, hash)
		if !Verify(&prv.PublicKey, hash, sig) {
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