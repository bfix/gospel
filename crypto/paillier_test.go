package crypto

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"crypto/rand"
	"fmt"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Test cases

func TestPaillier(t *testing.T) {

	fmt.Println("********************************************")
	fmt.Println("crypto/paillier Test")
	fmt.Println("********************************************")
	fmt.Println()

	for size := 128; size <= 2048; size *= 2 {
		fmt.Printf("%d: ", size)
		priv, err := NewPaillierPrivateKey(size)
		if err != nil {
			fmt.Printf("Error generating private key of size %d bits.\n", size)
			t.Fail()
		}
		pub := priv.GetPublicKey()
		for i := 0; i < 40; i++ {
			m, err := rand.Int(rand.Reader, pub.N)
			if err != nil {
				fmt.Println("Error generating message!")
				t.Fail()
			}
			c, err := pub.Encrypt(m)
			if err != nil {
				fmt.Println("Error encrypting message!")
				t.Fail()
			}
			d, err := priv.Decrypt(c)
			if err != nil {
				fmt.Println("Error decrypting message!")
				t.Fail()
			}
			if d.Cmp(m) != 0 {
				fmt.Println("Wrong result!")
				t.Fail()
			}
			fmt.Print(".")
		}
		fmt.Println()
	}
}
