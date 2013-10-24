package rpc

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	WALLET_COPY  = "/tmp/testwallet.dat"
	NEW_ACCOUNTS = false
	PASSPHRASE   = "HellFreezesOver"
	PRV_KEY      = ""
	BLOCK_HASH   = "00000000000003fab35380c07f6773ae27727b21016a8821c88e47e241c86458"
)

func TestRPC(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("bitcoin/rpc/calls Test")
	fmt.Println("********************************************************")

	sess, err := NewSession("http://localhost:18332", "DonaldDuck", "MoneyMakesTheWorldGoRound")
	if err != nil {
		fmt.Println("session creation failed: " + err.Error())
		t.Fail()
		return
	}

	//=================================================================
	// GENERIC methods
	//=================================================================

	info, err := sess.GetInfo()
	if err != nil {
		fmt.Println("GetInfo() failed: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetConnectionCount
	conns, err := sess.GetConnectionCount()
	if err != nil {
		fmt.Println("GetConnectionCount() failed: " + err.Error())
		t.Fail()
		return
	}
	if conns != info.Connections {
		fmt.Println("connection count mismatch")
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetDifficulty
	diff, err := sess.GetDifficulty()
	if err != nil {
		fmt.Println("GetDifficulty() failed: " + err.Error())
		t.Fail()
		return
	}
	if diff != info.Difficulty {
		fmt.Println("difficulty mismatch")
		t.Fail()
		return
	}

	//=================================================================
	// WALLET-related methods
	//=================================================================

	//-----------------------------------------------------  BackupWallet
	err = sess.BackupWallet(WALLET_COPY)
	if err != nil {
		fmt.Println("BackupWallet(): " + err.Error())
		t.Fail()
		return
	}
	_, err = os.Stat(WALLET_COPY)
	if err != nil {
		fmt.Println("Wallet file: " + err.Error())
		t.Fail()
		return
	}
	if err = os.Remove(WALLET_COPY); err != nil {
		fmt.Println("Wallet file: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  WalletLock
	err = sess.WalletLock()
	if err != nil {
		fmt.Println("WalletLock(): " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  WalletPassphrase
	err = sess.WalletPassphrase(PASSPHRASE, 600)
	if err != nil {
		fmt.Println("WalletPassphrase(): " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  KeypoolRefill
	err = sess.KeypoolRefill()
	if err != nil {
		fmt.Println("KeypoolRefill(): " + err.Error())
		t.Fail()
		return
	}

	if NEW_ACCOUNTS {
		//-------------------------------------------------  ImportPrivateKey
		err = sess.ImportPrivateKey(PRV_KEY)
		if err != nil {
			fmt.Println("ImportPrivateKey(): " + err.Error())
			t.Fail()
			return
		}
	}

	//=================================================================
	// ACCOUNT-related methods
	//=================================================================

	//-----------------------------------------------------  ListUnspent
	_, err = sess.ListUnspent(1, 999999)
	if err != nil {
		fmt.Println("ListUnspent() failed: " + err.Error())
		t.Fail()
		return
	}

	var (
		addr  = ""
		accnt = ""
	)
	if NEW_ACCOUNTS {
		//-------------------------------------------------  GetNewAddress
		label := fmt.Sprintf("Account %d", time.Now().Unix)
		addr, err = sess.GetNewAddress(label)
		if err != nil {
			fmt.Println("GetNewAddress() failed: " + err.Error())
			t.Fail()
			return
		}

		//-------------------------------------------------  SetAccount
		accnt = "Renamed " + label
		err = sess.SetAccount(addr, accnt)
		if err != nil {
			fmt.Println("SetAccount() failed: " + err.Error())
			t.Fail()
			return
		}

		//-------------------------------------------------  GetAccount
		label, err = sess.GetAccount(addr)
		if err != nil {
			fmt.Println("GetAccount() failed: " + err.Error())
			t.Fail()
			return
		}
		if accnt != label {
			fmt.Println("account label mismatch")
			t.Fail()
			return
		}

		//-------------------------------------------------  GetAccountAddress
		_, err = sess.GetAccountAddress(accnt)
		if err != nil {
			fmt.Println("GetAccountAddress() failed: " + err.Error())
			t.Fail()
			return
		}
	} else {
		//-------------------------------------------------  ListAccounts
		accnts, err := sess.ListAccounts(0)
		if err != nil {
			fmt.Println("ListAccounts() failed: " + err.Error())
			t.Fail()
			return
		}
		for label, _ := range accnts {
			//---------------------------------------------  GetAddressByAccount
			addr_list, err := sess.GetAddressesByAccount(label)
			if err != nil {
				fmt.Println("GetAddressesByAccount() failed: " + err.Error())
				t.Fail()
				return
			}
			if len(addr_list) == 0 {
				continue
			}
			// use first valid pair
			if len(label) > 0 {
				//-------------------------------------------------  GetBalance
				bal, err := sess.GetBalance(label)
				if err != nil {
					fmt.Println("GetBalance(label) failed: " + err.Error())
					t.Fail()
					return
				}
				if bal > 0 {
					accnt = label
					addr = addr_list[0]
					break
				}
			}
		}
		if len(accnt) == 0 {
			fmt.Println("No valid account label found")
			t.Fail()
			return
		}
	}
	//fmt.Println ("Using account '" + accnt + "' with address '" + addr + "'")

	//-----------------------------------------------------  ListAccounts
	accnts, err := sess.ListAccounts(0)
	if err != nil {
		fmt.Println("ListAccounts() failed: " + err.Error())
		t.Fail()
		return
	}
	if _, ok := accnts[accnt]; !ok {
		fmt.Println("ListAccounts(label) failed")
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetAddressesByAccount
	addr_list, err := sess.GetAddressesByAccount(accnt)
	if err != nil {
		fmt.Println("GetAddressesByAccount('" + accnt + "') failed: " + err.Error())
		t.Fail()
		return
	}
	found := false
	for _, a := range addr_list {
		if a == addr {
			found = true
			break
		}
	}
	if !found {
		fmt.Println("GetAddressesByAccount() fail to deliver known address")
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetBalance
	_, err = sess.GetBalance(accnt)
	if err != nil {
		fmt.Println("GetBalance(accnt) failed: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetBalanceAll
	_, err = sess.GetBalanceAll()
	if err != nil {
		fmt.Println("GetBalance() failed: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  ListReceivedByAccount
	rcv_1, err := sess.ListReceivedByAccount(1, false)
	if err != nil {
		fmt.Println("ListReceivedByAccount() failed: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  ListReceivedByAddress
	rcv_2, err := sess.ListReceivedByAddress(1, false)
	if err != nil {
		fmt.Println("ListReceivedByAddress() failed: " + err.Error())
		t.Fail()
		return
	}
	if len(rcv_1) != len(rcv_2) {
		fmt.Println("received count mismatch")
		t.Fail()
		return
	}

	//-----------------------------------------------------  ValidateAddress
	val, err := sess.ValidateAddress(addr)
	if err != nil {
		fmt.Println("ValidateAddress() failed: " + err.Error())
		t.Fail()
		return
	}
	if val.Address != addr {
		fmt.Println("ValidateAddress(): address mismatch")
		t.Fail()
		return
	}
	if val.Account != accnt {
		fmt.Println("ValidateAddress(): account mismatch")
		t.Fail()
		return
	}
	if !val.IsMine {
		fmt.Println("ValidateAddress(): owner mismatch")
		t.Fail()
		return
	}

	//=================================================================
	// BLOCKCHAIN-related methods
	//=================================================================

	//-----------------------------------------------------  GetBlockCount
	blks, err := sess.GetBlockCount()
	if err != nil {
		fmt.Println("GetBlockCount() failed: " + err.Error())
		t.Fail()
		return
	}
	if blks != info.Blocks {
		fmt.Println("block count mismatch")
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetBlock
	block, err := sess.GetBlock(BLOCK_HASH)
	if err != nil {
		fmt.Println("GetBlock() failed: " + err.Error())
		t.Fail()
		return
	}

	//-----------------------------------------------------  GetBlockHash
	blkhash, err := sess.GetBlockHash(block.Height)
	if err != nil {
		fmt.Println("GetBlockHash() failed: " + err.Error())
		t.Fail()
		return
	}
	if blkhash != block.Hash {
		fmt.Println("GetBlockHash() mismatch")
		t.Fail()
		return
	}

	//=================================================================
	// TRANSACTION-related methods
	//=================================================================

	//-----------------------------------------------------  ListTransactions
	txlist, err := sess.ListTransactions(accnt, 25, 0)
	if err != nil {
		fmt.Println("ListTransactions() failed: " + err.Error())
		t.Fail()
		return
	}
	if len(txlist) > 0 {
		txid := txlist[0].Id
		//-------------------------------------------------  GetTransaction
		_, err = sess.GetTransaction(txid)
		if err != nil {
			fmt.Println("GetTransaction() failed: " + err.Error())
			t.Fail()
			return
		}
	}

	//-----------------------------------------------------  ListSinceBlock
	txlist, _, err = sess.ListSinceBlock("", 1)
	if err != nil {
		fmt.Println("ListSinceBlock() failed: " + err.Error())
		t.Fail()
		return
	}
	if len(txlist) == 0 {
		fmt.Println("ListSinceBlock() with no results")
		t.Fail()
		return
	}

	//-----------------------------------------------------  SetTxFee
	err = sess.SetTxFee(0.0001)
	if err != nil {
		fmt.Println("SetTxFee() failed: " + err.Error())
		t.Fail()
		return
	}
}
