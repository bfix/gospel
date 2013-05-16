package crypto

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Test cases

func TestCounter(t *testing.T) {

	fmt.Println("********************************************")
	fmt.Println("crypto/counter Test")
	fmt.Println("********************************************")
	fmt.Println()

	for size := 128; size <= 2048; size *= 2 {
		fmt.Printf("%d: ", size)
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			fmt.Printf("Error generating private key of size %d bits.\n", size)
			t.Fail()
		}
		for i := 0; i < 10; i++ {
			if !test_counter(priv) {
				t.Fail()
			}
			fmt.Print(".")
		}
		fmt.Println()
	}
}

///////////////////////////////////////////////////////////////////////
/*
 * Test Counter
 */
func test_counter(priv *PaillierPrivateKey) bool {

	max := big.NewInt(2)
	pub := priv.GetPublicKey()
	cnt, err := NewCounter(pub)
	if err != nil {
		return false
	}
	var inc int64 = 0
	for i := 0; i < 100; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return false
		}
		cnt.Increment(v)
		if v.Bit(0) == 1 {
			inc++
		}
	}
	t := cnt.Get()
	t, err = priv.Decrypt(t)
	if err != nil {
		return false
	}
	v := t.Int64()
	if v == inc {
		return true
	}
	fmt.Printf("Counter mismatch: %d -- %d\n", v, inc)
	return false
}
