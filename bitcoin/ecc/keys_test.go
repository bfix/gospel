package ecc

import (
	"testing"
)

func TestKeys(t *testing.T) {
	var prv *PrivateKey
	for i := 0; i < 32; i++ {
		prv = GenerateKeys(i&1 == 1)
		b := prv.Bytes()
		if _, err := PrivateKeyFromBytes(b); err != nil {
			t.Fatal()
		}
		b = prv.PublicKey.Bytes()
		if _, err := PublicKeyFromBytes(b); err != nil {
			t.Fatal()
		}
		pnt := prv.Q
		tst := ScalarMultBase(prv.D)
		if !(IsOnCurve(pnt) && IsEqual(pnt, tst)) {
			t.Fatal()
		}
	}
}
