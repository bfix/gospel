package ecc

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"fmt"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

func TestKeys(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("ecc/keys Test")
	fmt.Println("********************************************************")

	var prv *PrivateKey

	// check 32 keys
	fmt.Println("Checking if public key point is on curve and consistent:")
	fmt.Print("    ")
	failed := false
	for i := 0; i < 32; i++ {

		// generate new key
		prv = GenerateKeys()
		// get public key point
		pnt := prv.Q
		tst := ScalarMultBase(prv.D)

		if !(IsOnCurve(pnt) && IsEqual(pnt, tst)) {
			failed = true
			fmt.Print("-")
		} else {
			fmt.Print("+")
		}
	}

	if failed {
		t.Fail()
		fmt.Println(" Failed")
	} else {
		fmt.Println(" O.K.")
	}
}
