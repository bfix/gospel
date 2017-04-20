package math

import (
	"fmt"
	"testing"
)

func TestIntBytes(t *testing.T) {
	c := TWO.Pow(256)
	for i := 0; i < 1000; i++ {
		a := NewIntRnd(c)
		b := NewIntFromBytes(a.Bytes())
		if !a.Equals(b) {
			t.Fatal("Bytes()/NewIntFromBytes() failed")
		}
	}
}

func TestExtendedEuclid(t *testing.T) {
	var (
		a, b *Int
		m    = NewInt(1000000000000000000)
	)
	test := func() {
		r := a.ExtendedEuclid(b)
		s := r[0].Mul(a).Add(r[1].Mul(b))
		if !s.Equals(ONE) {
			t.Fail()
		}
	}
	for i := 0; i < 10; {
		a = NewIntRnd(m).Add(ONE)
		b = NewIntRnd(a).Add(ONE)
		if !a.GCD(b).Equals(ONE) {
			continue
		}
		test()
		a, b = b, a
		test()
		i++
	}
}

func TestSqrt(t *testing.T) {
	p := NewIntRndPrimeBits(10)
	count := 0
	for i := 0; i < 1000; i++ {
		g := NewIntRnd(p)
		if g.Legendre(p) == 1 {
			count++
			h, err := SqrtModP(g, p)
			if err != nil {
				t.Fatal(err)
			}
			gg := h.ModPow(TWO, p)
			if !gg.Equals(g) {
				t.Fatal(fmt.Sprintf("result error: %v != %v", g, gg))
			}
		}
	}
}
