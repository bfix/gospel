package bitcoin

import (
	"testing"
)

func TestKeys(t *testing.T) {
	var prv *PrivateKey
	for i := 0; i < 32; i++ {
		prv = GenerateKeys(i&1 == 1)
		b := prv.Bytes()
		if _, err := PrivateKeyFromBytes(b); err != nil {
			t.Fatal("PrivateKeyFromBytes failed")
		}
		b = prv.PublicKey.Bytes()
		if _, err := PublicKeyFromBytes(b); err != nil {
			t.Fatal("PublicKeyFromBytes failed")
		}
		pnt := prv.Q
		tst := MultBase(prv.D)
		if !(pnt.IsOnCurve() && pnt.Equals(tst)) {
			t.Fatal("public point failed")
		}
	}
}
