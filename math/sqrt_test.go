package math

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestSqrt(t *testing.T) {
	p, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		t.Fatal()
	}
	count := 0
	for i := 0; i < 1000; i++ {
		g, err := rand.Int(rand.Reader, p)
		if err != nil {
			t.Fatal()
		}
		if isQuadraticResidue(g, p) {
			count++
			h, err := SqrtModP(g, p)
			if err != nil {
				t.Fatal()
			}
			gg := new(big.Int).Exp(h, TWO, p)
			if gg.Cmp(g) != 0 {
				t.Fatal()
			}
		}
	}
}
