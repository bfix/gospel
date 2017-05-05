package script

import (
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"testing"
)

func TestCompile(t *testing.T) {
	for _, src := range scr {
		bin, err := Compile(src)
		if err != nil {
			t.Fatal(err)
		}
		src2, err := Decompile(bin)
		if err != nil {
			t.Fatal(err)
		}
		if src != src2 {
			if verbose {
				fmt.Println(">>> " + src)
				fmt.Println("    " + hex.EncodeToString(bin))
				fmt.Println("<<< " + src2)
			}
			t.Fatal("Script compile/decompile mismatch")
		}
	}
}

func TestSign(t *testing.T) {
	// generate private key
	prv := ecc.GenerateKeys(true)
	pub := &prv.PublicKey
	pubHash := util.Hash160(pub.Bytes())

	// generate sig script
	s := hex.EncodeToString(pubHash)
	sigScr, err := Compile("OP_DUP OP_HASH160 " + s + " OP_EQUALVERIFY OP_CHECKSIG")
	if err != nil {
		t.Fatal(err)
	}

	// prepare raw transaction
	tx, err := util.NewDissectedTransaction(t0)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.PrepareForSign(0, sigScr); err != nil {
		t.Fatal(err)
	}

	// sign transaction
	sig, err := Sign(prv, 0x01, tx)
	if err != nil {
		t.Fatal(err)
	}

	// verify signature
	ok, err := Verify(pub, sig, tx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Signature mismatch")
	}
}
