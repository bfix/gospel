package math

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	Public test method

func TestSqrt(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("math/sqrt Test")
	fmt.Println("********************************************************")

	p, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		t.Fail()
		return
	}

	count := 0
	for i := 0; i < 1000; i++ {
		g, err := rand.Int(rand.Reader, p)
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
			return
		}
		if isQuadraticResidue(g, p) {
			count++
			h, err := SqrtModP(g, p)
			if err != nil {
				fmt.Println(err.Error())
				t.Fail()
				return
			}
			gg := new(big.Int).Exp(h, TWO, p)
			if gg.Cmp(g) != 0 {
				fmt.Printf("%d: %v == %v\n", i+1, g, gg)
				t.Fail()
				return
			}
		}
	}
}
