package crypto

import (
	"crypto/rand"
	"testing"
)

func TestPaillier(t *testing.T) {
	for size := 128; size <= 2048; size *= 2 {
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			t.Fatal("newpaillierprivatekey failed")
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 40; i++ {
			m, err := rand.Int(rand.Reader, pub.N)
			if err != nil {
				t.Fatal("rand failed")
			}
			c, err := pub.Encrypt(m)
			if err != nil {
				t.Fatal("encrypt failed")
			}
			d, err := priv.Decrypt(c)
			if err != nil {
				t.Fatal("decrypt failed")
			}
			if d.Cmp(m) != 0 {
				t.Fatal("message mismatch")
			}
		}
	}
}
