package rpc

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

const (
	walletCopy = "/tmp/testwallet.dat"
	newAccount = false
	passphrase = "HellFreezesOver"
	prvKey     = ""
	blockHash  = "00000000000003fab35380c07f6773ae27727b21016a8821c88e47e241c86458"
)

var (
	sess  *Session
	err   error
	info  *Info
	addr  = ""
	accnt = ""
)

func init() {
	sess, err = NewSession("http://localhost:18332", "DonaldDuck", "MoneyMakesTheWorldGoRound")
	if err != nil {
		sess = nil
	}
}

func TestSession(t *testing.T) {
	if sess == nil {
		t.Fatal()
	}
	info, err = sess.GetInfo()
	if err != nil {
		t.Fatal()
	}
}

func TestConnectionCount(t *testing.T) {
	conns, err := sess.GetConnectionCount()
	if err != nil {
		t.Fatal()
	}
	if conns != info.Connections {
		t.Fatal()
	}
}

func TestDifficulty(t *testing.T) {
	diff, err := sess.GetDifficulty()
	if err != nil {
		t.Fatal()
	}
	if diff != info.Difficulty {
		t.Fatal()
	}
}

func TestWallet(t *testing.T) {
	if err = sess.BackupWallet(walletCopy); err != nil {
		t.Fatal()
	}
	if _, err = os.Stat(walletCopy); err != nil {
		t.Fatal()
	}
	if err = os.Remove(walletCopy); err != nil {
		t.Fatal()
	}
	if err = sess.WalletLock(); err != nil {
		t.Fatal()
	}
	if err = sess.WalletPassphrase(passphrase, 600); err != nil {
		t.Fatal()
	}
}

func TestKeypool(t *testing.T) {
	err = sess.KeypoolRefill()
	if err != nil {
		t.Fatal()
	}
}

func TestImport(t *testing.T) {
	if newAccount {
		if err = sess.ImportPrivateKey(prvKey); err != nil {
			t.Fatal()
			return
		}
	}
}

func TestListUnspend(t *testing.T) {
	if _, err = sess.ListUnspent(1, 999999); err != nil {
		t.Fatal()
	}

}

func TestAccount(t *testing.T) {
	if newAccount {
		label := fmt.Sprintf("Account %d", time.Now().Unix())
		addr, err = sess.GetNewAddress(label)
		if err != nil {
			t.Fatal()
		}
		accnt = "Renamed " + label
		err = sess.SetAccount(addr, accnt)
		if err != nil {
			t.Fatal()
		}
		label, err = sess.GetAccount(addr)
		if err != nil {
			t.Fatal()
		}
		if accnt != label {
			t.Fatal()
		}
		_, err = sess.GetAccountAddress(accnt)
		if err != nil {
			t.Fatal()
		}
	}

	accnts, err := sess.ListAccounts(0)
	if err != nil {
		t.Fatal()
	}
	for label := range accnts {
		addrList, err := sess.GetAddressesByAccount(label)
		if err != nil {
			t.Fatal()
		}
		if len(addrList) == 0 {
			continue
		}
		if len(label) > 0 {
			bal, err := sess.GetBalance(label)
			if err != nil {
				t.Fatal()
			}
			if bal > 0 {
				accnt = label
				addr = addrList[0]
				break
			}
		}
	}
	if len(accnt) > 0 {
		if _, ok := accnts[accnt]; !ok {
			t.Fatal()
		}
	}

	addrList, err := sess.GetAddressesByAccount(accnt)
	if err != nil {
		t.Fatal()
	}
	found := false
	for _, a := range addrList {
		if a == addr {
			found = true
			break
		}
	}
	if !found {
		t.Fatal()
	}
}

func TestBalance(t *testing.T) {
	if _, err = sess.GetBalance(accnt); err != nil {
		t.Fatal()
	}
	if _, err = sess.GetBalanceAll(); err != nil {
		t.Fatal()
	}
}

func TestReceived(t *testing.T) {
	rcv1, err := sess.ListReceivedByAccount(1, false)
	if err != nil {
		t.Fatal()
	}
	rcv2, err := sess.ListReceivedByAddress(1, false)
	if err != nil {
		t.Fatal()
	}
	if len(rcv1) != len(rcv2) {
		t.Fatal()
	}
}

func TestAddress(t *testing.T) {
	val, err := sess.ValidateAddress(addr)
	if err != nil {
		t.Fatal()
	}
	if val.Address != addr {
		t.Fatal()
	}
	if val.Account != accnt {
		t.Fatal()
	}
	if !val.IsMine {
		t.Fatal()
		return
	}
}

func TestBlock(t *testing.T) {
	blks, err := sess.GetBlockCount()
	if err != nil {
		t.Fatal()
	}
	if blks != info.Blocks {
		t.Fatal()
	}
	block, err := sess.GetBlock(blockHash)
	if err != nil {
		t.Fatal()
	}
	blkhash, err := sess.GetBlockHash(block.Height)
	if err != nil {
		t.Fatal()
	}
	if blkhash != block.Hash {
		t.Fatal()
	}
}

func TestTransaction(t *testing.T) {
	txlist, err := sess.ListTransactions(accnt, 25, 0)
	if err != nil {
		t.Fatal()
	}
	if len(txlist) > 0 {
		txid := txlist[0].ID
		if _, err = sess.GetTransaction(txid); err != nil {
			t.Fatal()
		}
	}

	txlist, _, err = sess.ListSinceBlock("", 1)
	if err != nil {
		t.Fatal()
	}
	if len(txlist) == 0 {
		t.Fatal()
	}
}

func TestFee(t *testing.T) {
	if err = sess.SetTxFee(0.0001); err != nil {
		t.Fatal()
	}
}
