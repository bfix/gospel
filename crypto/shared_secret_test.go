package crypto

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Test cases

func TestKeys(t *testing.T) {

	fmt.Println("********************************************")
	fmt.Println("crypto/shared_secret Test")
	fmt.Println("********************************************")
	fmt.Println()

	n, k := 10, 7
	s := big.NewInt(1234567890123456)
	fmt.Printf("           Secret: s = %v\n", s)
	fmt.Printf(" number of shares: n = %v\n", n)
	fmt.Printf("required treshold: k = %v\n", k)
	fmt.Println()

	p := nextPrime(s) // TEST ONLY!!! 'p' should be a random prime >> 's'!!!
	shares := Split(s, p, n, k)
	for i, sh := range shares {
		fmt.Printf("Share #%v = (%v,%v,%v)\n", i+1, sh.X, sh.Y, sh.P)
	}
	fmt.Println()

	for kk := 1; kk <= n; kk++ {
		fmt.Printf("Round #%v:", kk)
		perm := rand.Perm(n)
		coop := make([]Share, kk)
		for i, _ := range coop {
			coop[i] = shares[perm[i]]
			fmt.Printf(" %d", perm[i]+1)
		}

		s2 := Reconstruct(coop)
		fmt.Printf(" => %v\n", s2)

		switch {
		case s.Cmp(s2) != 0 && kk >= k:
			t.Fail()
		case s.Cmp(s2) == 0 && kk < k:
			t.Fail()
		}
	}
}

///////////////////////////////////////////////////////////////////////
// Find next prime number larger than p

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
