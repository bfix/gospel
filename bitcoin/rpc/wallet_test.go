package rpc

import (
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"os"
	"testing"
	"time"
)

var (
	_addr  = ""
	_accnt = ""
)

func TestWallet(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	walletCopy := os.Getenv("BTC_WALLET_COPY")
	if len(walletCopy) == 0 {
		t.Skip("skipping test: no copy location for wallet specified")
	}
	if err := sess.BackupWallet(walletCopy); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(walletCopy); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(walletCopy); err != nil {
		t.Fatal(err)
	}
	passphrase := os.Getenv("BTC_WALLET_PP")
	if len(passphrase) == 0 {
		t.Skip("skipping test: no passphrase for wallet specified")
	}
	if err := sess.WalletLock(); err != nil {
		t.Fatal(err)
	}
	if err := sess.WalletPassphrase(passphrase, 3600); err != nil {
		t.Fatal(err)
	}
}

func TestKeypool(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err := sess.KeypoolRefill(); err != nil {
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
	if err := sess.ImportPrivateKey(prvKey); err != nil {
		t.Fatal(err)
		return
	}
}

func TestListUnspent(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err := sess.ListUnspent(1, 999999); err != nil {
		t.Fatal(err)
	}
}

func TestAccount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	var err error
	label := fmt.Sprintf("Account %d", time.Now().Unix())
	if _addr, err = sess.GetNewAddress(label); err != nil {
		t.Fatal(err)
	}
	_accnt = "Renamed " + label
	if err = sess.SetAccount(_addr, _accnt); err != nil {
		t.Fatal(err)
	}
	if label, err = sess.GetAccount(_addr); err != nil {
		t.Fatal(err)
	}
	if _accnt != label {
		t.Fatal("account label mismatch")
	}
	if _, err := sess.GetAccountAddress(_accnt); err != nil {
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
				_accnt = label
				_addr = addrList[0]
				break
			}
		}
	}
	if len(_accnt) > 0 {
		if _, ok := accnts[_accnt]; !ok {
			t.Fatal("account list failure")
		}
	}

	addrList, err := sess.GetAddressesByAccount(_accnt)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range addrList {
		if a == _addr {
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
	if _, err := sess.GetBalance(_accnt); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.GetBalanceAll(); err != nil {
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
	val, err := sess.ValidateAddress(_addr)
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("Validity: %s\n", val)
	}
	if val.Address != _addr {
		t.Fatal("address mismatch")
	}
	if !val.IsMine {
		t.Fatal("owner mismatch")
		return
	}
}

func TestGetRawChangeAddress(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err := sess.GetRawChangeAddress(); err != nil {
		t.Fatal(err)
	}
}

func TestGetUnconfirmedBalance(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if _, err := sess.GetUnconfirmedBalance(); err != nil {
		t.Fatal(err)
	}
}

func TestGetWalletInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	wi, err := sess.GetWalletInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("WalletInfo: %s\n", wi)
	}
	b, err := sess.GetBalance("*")
	if err != nil {
		t.Fatal(err)
	}
	if b != wi.Balance {
		t.Fatal("Balance mismatch")
	}
}

func TestImportAddress(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	k := ecc.GenerateKeys(true)
	a := util.MakeTestAddress(&k.PublicKey)
	if err := sess.ImportAddress(a, "", false); err != nil {
		t.Fatal(err)
	}
}

func TestListAddressGroupings(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	ag, err := sess.ListAddressGroupings()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("AddressGroups: %s\n", ag)
	}
}

func TestSignMessage(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	msg := "This is a test message"
	sig, err := sess.SignMessage(_addr, msg)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := sess.VerifyMessage(_addr, sig, msg)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("signature verification failed.")
	}
}

func TestSignMessageWithPrivKey(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	key, err := sess.DumpPrivKey(_addr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.SignMessageWithPrivKey(key, "Test message"); err != nil {
		t.Fatal(err)
	}
}
