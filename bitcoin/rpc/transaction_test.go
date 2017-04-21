package rpc

import (
	"fmt"
	"testing"
)

func TestTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	txlist, err := sess.ListTransactions(accnt, 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(txlist) > 0 {
		txid := txlist[0].TxID
		tx, err := sess.GetTransaction(txid, false)
		if err != nil {
			t.Fatal(err)
		}
		if verbose {
			dumpObj("Transaction: %s\n", tx)
		}
		rtxHex, err := sess.GetRawTransaction(txid)
		if err != nil {
			t.Fatal(err)
		}
		rtx, err := sess.DecodeRawTransaction(rtxHex)
		if err != nil {
			t.Fatal(err)
		}
		if verbose {
			dumpObj("RawTransaction: %s\n", rtx)
		}
		ent, err := sess.GetMemPoolEntry(txid)
		if err == nil && verbose {
			dumpObj("MemPoolEntry: %v\n", ent)
		}
		anc, err := sess.GetMemPoolAncestorObjs(txid)
		if err == nil && verbose {
			dumpObj("MemPoolAncestors: %v\n", anc)
		}
		dec, err := sess.GetMemPoolDecendantObjs(txid)
		if err == nil && verbose {
			dumpObj("MemPoolDecendants: %v\n", dec)
		}
	}

	txlist, _, err = sess.ListSinceBlock("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(txlist) == 0 {
		fmt.Println("no transactions")
	}
}
