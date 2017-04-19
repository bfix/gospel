package rpc

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	sess  *Session
	err   error
	info  *Info
	addr  = ""
	accnt = ""
)

func init() {
	rpcaddr := os.Getenv("BTC_HOST")
	user := os.Getenv("BTC_USER")
	passwd := os.Getenv("BTC_PASSWORD")
	sess, err = NewSession(rpcaddr, user, passwd)
	if err != nil {
		sess = nil
	}
}

func TestSession(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	info, err = sess.GetInfo()
	if err != nil {
		sess = nil
		t.Fatal(err)
	}
}

func TestConnectionCount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	conns, err := sess.GetConnectionCount()
	if err != nil {
		t.Fatal("getsessioncount failed")
	}
	if conns != info.Connections {
		t.Fatal("num-seesions in info mismatch")
	}
}

func TestDifficulty(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	diff, err := sess.GetDifficulty()
	if err != nil {
		t.Fatal("getdifficulty failed")
	}
	if diff != info.Difficulty {
		t.Fatal("difficulty mismatch in info")
	}
}

func TestWallet(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	walletCopy := os.Getenv("BTC_WALLET_COPY")
	if len(walletCopy) == 0 {
		t.Skip("skipping test: no copy location for wallet specified")
	}
	if err = sess.BackupWallet(walletCopy); err != nil {
		t.Fatal("backupwallet failed")
	}
	if _, err = os.Stat(walletCopy); err != nil {
		t.Fatal("no wallet copy created")
	}
	if err = os.Remove(walletCopy); err != nil {
		t.Fatal("failed to remove wallet copy")
	}
	passphrase := os.Getenv("BTC_WALLET_PP")
	if len(passphrase) == 0 {
		t.Skip("skipping test: no passphrase for wallet specified")
	}
	if err = sess.WalletLock(); err != nil {
		t.Fatal("walletlock failed")
	}
	if err = sess.WalletPassphrase(passphrase, 3600); err != nil {
		t.Fatal("walletpassphrase failed")
	}
}

func TestKeypool(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	err = sess.KeypoolRefill()
	if err != nil {
		t.Fatal("keypoolrefill failed")
	}
}

func TestImport(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	prvKey := os.Getenv("BTC_PRIVKEY")
	if len(prvKey) == 0 {
		t.Skip("skipping test: no private key for import found")
	}
	if err = sess.ImportPrivateKey(prvKey); err != nil {
		t.Fatal(err)
		return
	}
}

func TestListUnspend(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err = sess.ListUnspent(1, 999999); err != nil {
		t.Fatal("listunspend failed")
	}

}

func TestAccount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	label := fmt.Sprintf("Account %d", time.Now().Unix())
	addr, err = sess.GetNewAddress(label)
	if err != nil {
		t.Fatal("getnewaddress failed")
	}
	accnt = "Renamed " + label
	err = sess.SetAccount(addr, accnt)
	if err != nil {
		t.Fatal("setaccount failed")
	}
	label, err = sess.GetAccount(addr)
	if err != nil {
		t.Fatal("getaccount failed")
	}
	if accnt != label {
		t.Fatal("account label mismatch")
	}
	_, err = sess.GetAccountAddress(accnt)
	if err != nil {
		t.Fatal("getaccountaddress failed")
	}
	accnts, err := sess.ListAccounts(0)
	if err != nil {
		t.Fatal("listaccounts failed")
	}
	for label := range accnts {
		addrList, err := sess.GetAddressesByAccount(label)
		if err != nil {
			t.Fatal("getaddressbyaccount failed")
		}
		if len(addrList) == 0 {
			continue
		}
		if len(label) > 0 {
			bal, err := sess.GetBalance(label)
			if err != nil {
				t.Fatal("getbalance failed")
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
			t.Fatal("account list failure")
		}
	}

	addrList, err := sess.GetAddressesByAccount(accnt)
	if err != nil {
		t.Fatal("getaddressbyaccount failed")
	}
	found := false
	for _, a := range addrList {
		if a == addr {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("address not found")
	}
}

func TestBalance(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err = sess.GetBalance(accnt); err != nil {
		t.Fatal("getbalance failed")
	}
	if _, err = sess.GetBalanceAll(); err != nil {
		t.Fatal("getbalanceall failed")
	}
}

func TestReceived(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	rcv1, err := sess.ListReceivedByAccount(1, false)
	if err != nil {
		t.Fatal("listreceivedbyaccount failed")
	}
	rcv2, err := sess.ListReceivedByAddress(1, false)
	if err != nil {
		t.Fatal("listreceivedbyaddress failed")
	}
	if len(rcv1) != len(rcv2) {
		t.Fatal("receivers mismatch")
	}
}

func TestAddress(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	val, err := sess.ValidateAddress(addr)
	if err != nil {
		t.Fatal("validateaddress failed")
	}
	if val.Address != addr {
		t.Fatal("address mismatch")
	}
	if val.Account != accnt {
		t.Fatal("account mismatch")
	}
	if !val.IsMine {
		t.Fatal("owner mismatch")
		return
	}
}

func TestBlock(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	blks, err := sess.GetBlockCount()
	if err != nil {
		t.Fatal("getblockcount failed")
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
		t.Fatal("getblock failed")
	}
	blkhash, err := sess.GetBlockHash(block.Height)
	if err != nil {
		t.Fatal("getblockhash failed")
	}
	if blkhash != block.Hash {
		t.Fatal("blockhash mismatch")
	}
}

func TestTransaction(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	txlist, err := sess.ListTransactions(accnt, 25, 0)
	if err != nil {
		t.Fatal("listtransactions failed")
	}
	if len(txlist) > 0 {
		txid := txlist[0].ID
		if _, err = sess.GetTransaction(txid); err != nil {
			t.Fatal("gettransaction failed")
		}
	}

	txlist, _, err = sess.ListSinceBlock("", 1)
	if err != nil {
		t.Fatal("listsinceblock failed")
	}
	if len(txlist) == 0 {
		fmt.Println("no transactions")
	}
}

func TestFee(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err = sess.SetTxFee(0.0001); err != nil {
		t.Fatal("settxfee failed")
	}
}
