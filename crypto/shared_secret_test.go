package crypto

import (
	"github.com/bfix/gospel/math"
	"math/rand"
	"testing"
)

func TestKeys(t *testing.T) {

	n, k := 10, 7
	s := math.NewInt(1234567890123456)

	p := nextPrime(s) // TEST ONLY!!! 'p' should be a random prime >> 's'!!!
	shares := Split(s, p, n, k)

	for kk := 1; kk <= n; kk++ {
		perm := rand.Perm(n)
		coop := make([]Share, kk)
		for i := range coop {
			coop[i] = shares[perm[i]]
		}

		s2 := Reconstruct(coop)

		switch {
		case !s.Equals(s2) && kk >= k:
			t.Fatal("failed reconstruction")
		case s.Equals(s2) && kk < k:
			t.Fatal("pre-mature reconstruction")
		}
	}
}

func nextPrime(p *math.Int) *math.Int {
	// make sure p is odd
	if p.Bit(0) == 0 {
		p = p.Add(math.ONE)
	}
	step := math.TWO
	for {
		p = p.Add(step)
		if p.ProbablyPrime(128) {
			break
		}
	}
	return p
}
