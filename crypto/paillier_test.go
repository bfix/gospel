package crypto

import (
	"github.com/bfix/gospel/math"
	"testing"
)

func TestPaillier(t *testing.T) {
	for size := 128; size <= 2048; size *= 2 {
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			t.Fatal("newpaillierprivatekey failed")
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 5; i++ {
			m := math.NewIntRnd(pub.N)
			c, err := pub.Encrypt(m)
			if err != nil {
				t.Fatal("encrypt failed")
			}
			d, err := priv.Decrypt(c)
			if err != nil {
				t.Fatal("decrypt failed")
			}
			if !d.Equals(m) {
				t.Fatal("message mismatch")
			}
		}
	}
}
