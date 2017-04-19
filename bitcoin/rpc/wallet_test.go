package rpc

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	addr  = ""
	accnt = ""
)

func TestWallet(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	walletCopy := os.Getenv("BTC_WALLET_COPY")
	if len(walletCopy) == 0 {
		t.Skip("skipping test: no copy location for wallet specified")
	}
	if err = sess.BackupWallet(walletCopy); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(walletCopy); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(walletCopy); err != nil {
		t.Fatal(err)
	}
	passphrase := os.Getenv("BTC_WALLET_PP")
	if len(passphrase) == 0 {
		t.Skip("skipping test: no passphrase for wallet specified")
	}
	if err = sess.WalletLock(); err != nil {
		t.Fatal(err)
	}
	if err = sess.WalletPassphrase(passphrase, 3600); err != nil {
		t.Fatal(err)
	}
}

func TestKeypool(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	err = sess.KeypoolRefill()
	if err != nil {
		t.Fatal(err)
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

func TestListUnspent(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err = sess.ListUnspent(1, 999999); err != nil {
		t.Fatal(err)
	}
}

func TestAccount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	label := fmt.Sprintf("Account %d", time.Now().Unix())
	addr, err = sess.GetNewAddress(label)
	if err != nil {
		t.Fatal(err)
	}
	accnt = "Renamed " + label
	err = sess.SetAccount(addr, accnt)
	if err != nil {
		t.Fatal(err)
	}
	label, err = sess.GetAccount(addr)
	if err != nil {
		t.Fatal(err)
	}
	if accnt != label {
		t.Fatal("account label mismatch")
	}
	_, err = sess.GetAccountAddress(accnt)
	if err != nil {
		t.Fatal(err)
	}
	accnts, err := sess.ListAccounts(0)
	if err != nil {
		t.Fatal(err)
	}
	for label := range accnts {
		addrList, err := sess.GetAddressesByAccount(label)
		if err != nil {
			t.Fatal(err)
		}
		if len(addrList) == 0 {
			continue
		}
		if len(label) > 0 {
			bal, err := sess.GetBalance(label)
			if err != nil {
				t.Fatal(err)
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
		t.Fatal(err)
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
		t.Fatal(err)
	}
	if _, err = sess.GetBalanceAll(); err != nil {
		t.Fatal(err)
	}
}

func TestReceived(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	rcv1, err := sess.ListReceivedByAccount(1, false, false)
	if err != nil {
		t.Fatal(err)
	}
	rcv2, err := sess.ListReceivedByAddress(1, false, false)
	if err != nil {
		t.Fatal(err)
	}
	l1 := len(rcv1)
	l2 := len(rcv2)
	if l1 != l2 {
		if verbose {
			dumpObj("[1] %s\n", rcv1)
			dumpObj("[2] %s\n", rcv2)
		}
		t.Fatal(fmt.Sprintf("receivers mismatch: %d != %d", l1, l2))
	}
}

func TestAddress(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	val, err := sess.ValidateAddress(addr)
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("Validity: %s\n", val)
	}
	if val.Address != addr {
		t.Fatal("address mismatch")
	}
	if !val.IsMine {
		t.Fatal("owner mismatch")
		return
	}
	/*
		found := false
		for _, a := range val.Addresses {
			if a.Account == accnt {
				found = true
			}
		}
		if !found {
			t.Fatal("account mismatch")
		}
	*/
}
