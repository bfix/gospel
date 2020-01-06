package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

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
			dumpObj("Transaction HEX: %s\n", tx.Hex)
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
	dec, err := sess.GetMemPoolDescendantObjs(_txid)
	if err == nil && verbose {
		dumpObj("MemPoolDescendants: %v\n", dec)
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

func TestCheckTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	// get all transactions for the default account
	var txList []*Transaction
	skip := 0
	for {
		txs, err := sess.ListTransactions("", 25, skip)
		if err != nil {
			t.Fatal(err)
		}
		if len(txs) == 0 {
			break
		}
		skip += len(txs)
		txList = append(txList, txs...)
	}
	num := len(txList)
	fmt.Printf("CheckTransaction: found %d transactions.\n", num)
	if num < 2 {
		t.Skip("No enough transactions to run test")
	}
	// convert transactions to their raw format
	var err error
	rtxList := make(map[string]*RawTransaction)
	for _, tx := range txList {
		rtxList[tx.TxID], err = sess.GetRawTransactionObj(tx.TxID)
		if err != nil {
			t.Fatal(err)
		}
	}
	if verbose {
		for _, tx := range rtxList {
			dumpObj("%v\n", tx)
		}
	}
	// find all matching in/out pairs and validate them (run associated scripts)
	for _, tx := range rtxList {
		for j, vin := range tx.Vin {
			if len(vin.TxID) > 0 {
				vinTx, ok := rtxList[vin.TxID]
				if !ok {
					if vinTx, err = sess.GetRawTransactionObj(vin.TxID); err != nil {
						t.Fatal(err)
					}
				}
				fmt.Printf("Checking: %s[%d] --> %s[%d]\n", vin.TxID, vin.Vout, tx.TxID, j)
				if valid, err := VerifyTransfer(vinTx, vin.Vout, tx, j); err != nil {
					t.Fatal(err)
				} else if !valid {
					fmt.Println("==> FAILED")
				} else {
					fmt.Println("==> SUCCESS")
				}
			}
		}
	}
}
