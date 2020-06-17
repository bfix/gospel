package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

import (
	"os"
	"testing"
)

func TestBlock(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	blks, err := sess.GetBlockCount()
	if err != nil {
		t.Fatal(err)
	}
	if blks != info.Blocks {
		t.Fatal("blockcount mismatch in info")
	}
	blockHash := os.Getenv("BTC_BLOCK_HASH")
	if len(blockHash) == 0 {
		blockHash = "000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943"
	}
	block, err := sess.GetBlock(blockHash)
	if err != nil {
		t.Fatal(err)
	}
	blkhash, err := sess.GetBlockHash(block.Height)
	if err != nil {
		t.Fatal(err)
	}
	if blkhash != block.Hash {
		t.Fatal("blockhash mismatch")
	}
}

func TestBlockchainInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	bci, err := sess.GetBlockchainInfo()
	if err != nil {
		sess = nil
		t.Fatal(err)
	}
	if verbose {
		dumpObj("BlockchainInfo: %s\n", bci)
	}
}

func TestBlockTemplate(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	caps := []string{"coinbasetxn", "workid", "coinbase/append"}
	bt, err := sess.GetBlockTemplate(caps)
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("BlockTemplate: %s\n", bt)
	}
}

func TestChainTips(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	ct, err := sess.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("ChainTips: %s\n", ct)
	}
}

func TestVerifyChain(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err := sess.VerifyChain(-1, -1); err != nil {
		t.Fatal(err)
	}
}
