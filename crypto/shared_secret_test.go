package crypto

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestKeys(t *testing.T) {

	n, k := 10, 7
	s := big.NewInt(1234567890123456)

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
		case s.Cmp(s2) != 0 && kk >= k:
			t.Fatal()
		case s.Cmp(s2) == 0 && kk < k:
			t.Fatal()
		}
	}
}

func nextPrime(p *big.Int) *big.Int {

	// make sure p is odd
	if p.Bit(0) == 0 {
		p = new(big.Int).Add(p, big.NewInt(1))
	}

	step := big.NewInt(2)
	for {
		p = new(big.Int).Add(p, step)
		if p.ProbablyPrime(128) {
			break
		}
	}
	return p
}
