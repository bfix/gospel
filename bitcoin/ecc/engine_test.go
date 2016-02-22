package ecc

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"fmt"
	"github.com/bfix/gospel/math"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestEngine(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("ecc/engine Test")
	fmt.Println("********************************************************")

	fmt.Println("Checking sign/verify chain:")
	fmt.Print("    ")
	failed := false
	for i := 0; i < 32; i++ {
		prv := GenerateKeys()
		hash := nRnd(math.ONE).Bytes()
		r, s := Sign(prv, hash)
		if Verify(&prv.PublicKey, hash, r, s) {
			fmt.Print("+")
		} else {
			fmt.Print("-")
		}
	}
	if failed {
		fmt.Println(" Failed")
		t.Fail()
	} else {
		fmt.Println(" O.K.")
	}
}
