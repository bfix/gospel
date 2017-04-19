package rpc

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
		blockHash = "00000000000003fab35380c07f6773ae27727b21016a8821c88e47e241c86458"
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
