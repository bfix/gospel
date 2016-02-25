package crypto

import (
	"crypto/rand"
	"github.com/bfix/gospel/math"
	"testing"
)

func TestCounter(t *testing.T) {
	for size := 128; size <= 2048; size *= 2 {
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			t.Fatal()
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 10; i++ {
			cnt, err := NewCounter(pub)
			if err != nil {
				t.Fatal()
			}
			var inc int64
			for i := 0; i < 100; i++ {
				v, err := rand.Int(rand.Reader, math.TWO)
				if err != nil {
					t.Fatal()
				}
				cnt.Increment(v)
				if v.Bit(0) == 1 {
					inc++
				}
			}
			tt := cnt.Get()
			tt, err = priv.Decrypt(tt)
			if err != nil {
				t.Fatal()
			}
			v := tt.Int64()
			if v != inc {
				t.Fatal()
			}
		}
	}
}
