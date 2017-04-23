package rpc

import (
	"fmt"
	"testing"
)

var (
	_txid  string
	_block string
)

func TestTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	txlist, err := sess.ListTransactions(_accnt, 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(txlist) > 0 {
		_txid = txlist[0].TxID
		tx, err := sess.GetTransaction(_txid, false)
		if err != nil {
			t.Fatal(err)
		}
		if verbose {
			dumpObj("Transaction: %s\n", tx)
		}
		_block = tx.BlockHash
	}
}

func TestRawTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	rtxHex, err := sess.GetRawTransaction(_txid)
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
}

func TestMemPoolEntry(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	ent, err := sess.GetMemPoolEntry(_txid)
	if err == nil && verbose {
		dumpObj("MemPoolEntry: %v\n", ent)
	}
}

func TestMemPoolAncestors(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	anc, err := sess.GetMemPoolAncestorObjs(_txid)
	if err == nil && verbose {
		dumpObj("MemPoolAncestors: %v\n", anc)
	}
}

func TestMemPoolDescendants(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	dec, err := sess.GetMemPoolDecendantObjs(_txid)
	if err == nil && verbose {
		dumpObj("MemPoolDecendants: %v\n", dec)
	}
}

func TestListSinceBlock(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	txlist, _, err := sess.ListSinceBlock("", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(txlist) == 0 {
		fmt.Println("no transactions")
	}
}

func TestGetTxOut(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	oi, err := sess.GetTxOut(_txid, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("OutputInfo: %s\n", oi)
	}
}

func TestGetTxOutProof(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	p, err := sess.GetTxOutProof([]string{_txid}, "")
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("TxOutProof: %s\n", p)
	}
	a, err := sess.VerifyTxOutProof(p)
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("VerifyTxOutProof result: %s\n", a)
	}
}

func TestGetTxOutSetInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	/*
		si, err := sess.GetTxOutSetInfo()
		if err != nil {
			t.Fatal(err)
		}
		if verbose {
			dumpObj("TxOutSetInfo: %s\n", si)
		}
	*/
}

func TestPrioritiseTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if len(_txid) == 0 {
		t.Skip("skipping test: no transaction available")
	}
	if _, err := sess.PrioritiseTransaction(_txid, 23); err != nil {
		t.Fatal(err)
	}
}
