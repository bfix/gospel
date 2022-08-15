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
	"errors"
)

// AbandonTransaction marks an in-wallet transaction and all its in-wallet
// descendants as abandoned. This allows their inputs to be respent.
func (s *Session) AbandonTransaction(txid string) error {
	_, err := s.call("abandontransaction", []Data{txid})
	return err
}

// AddMultiSigAddress adds a P2SH multisig address to the wallet.
// 'm' is the minimum number of signatures required to spend this
// m-of-n multisig script.
// 'list' is either an array of strings with each string being a public key
// or address; or a public key against which signatures will be checked. If
// wallet support is enabled, this may be a P2PKH address belonging to the
// wallet—the corresponding public key will be substituted. There must be at
// least as many keys as specified by the Required parameter, and there may
// be more keys.
func (s *Session) AddMultiSigAddress(m int, list interface{}, account string) (string, error) {
	var data []Data
	data = append(data, m)
	switch list.(type) {
	case []string:
		data = append(data, list)
	case string:
		data = append(data, list)
	default:
		return "", errors.New("unknown parameter type #2")
	}
	data = append(data, account)
	res, err := s.call("addmultisigaddress", data)
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// BackupWallet securely copies the wallet to another file/directory.
func (s *Session) BackupWallet(dest string) error {
	_, err := s.call("backupwallet", []Data{dest})
	return err
}

// CreateMultiSig adds a P2SH multisig address to the wallet.
// 'm' is the minimum number of signatures required to spend this
// m-of-n multisig script.
// 'list' is either an array of strings with each string being a public key
// or address; or a public key against which signatures will be checked. If
// wallet support is enabled, this may be a P2PKH address belonging to the
// wallet—the corresponding public key will be substituted. There must be at
// least as many keys as specified by the Required parameter, and there may
// be more keys.
func (s *Session) CreateMultiSig(m int, list []string) (*MultiSigAddr, error) {
	res, err := s.call("createmultisig", []Data{m, list})
	if err != nil {
		return nil, err
	}
	saddr := new(MultiSigAddr)
	if ok, err := res.UnmarshalResult(saddr); !ok {
		return nil, err
	}
	return saddr, nil
}

// DumpPrivKey reveals the private key corresponding to <bitcoinaddress>
func (s *Session) DumpPrivKey(address string) (string, error) {
	res, err := s.call("dumpprivkey", []Data{address})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// DumpWallet creates or overwrites a file with all wallet keys in a human-
// readable format.
func (s *Session) DumpWallet(file string) error {
	_, err := s.call("dumpwallet", []Data{file})
	return err
}

// EncryptWallet encrypts the wallet with a passphrase. This is only to enable
// encryption for the first time. After encryption is enabled, you will need
// to enter the passphrase to use private keys.
func (s *Session) EncryptWallet(passphrase string) (string, error) {
	res, err := s.call("encryptwallet", []Data{passphrase})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetAccount returns the label for a Bitcoin address.
func (s *Session) GetAccount(address string) (string, error) {
	res, err := s.call("getaccount", []Data{address})
	if err != nil {
		return "", err
	}
	label, _ := res.Result.(string)
	return label, err
}

// GetAccountAddress returns the first Bitcoin address matching label.
func (s *Session) GetAccountAddress(label string) (string, error) {
	res, err := s.call("getaccountaddress", []Data{label})
	if err != nil {
		return "", err
	}
	addr, _ := res.Result.(string)
	return addr, nil
}

// GetAddressesByAccount returns the an array of bitcoin addresses
// matching label.
func (s *Session) GetAddressesByAccount(label string) (list []string, err error) {
	var res *Response
	if res, err = s.call("getaddressesbyaccount", []Data{label}); err != nil {
		return
	}
	list, _ = res.Result.([]string)
	return
}

// GetBalance returns the balance in the account.
func (s *Session) GetBalance(label string) (float64, error) {
	res, err := s.call("getbalance", []Data{label})
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

// GetBalanceAll returns the balance in all accounts in the wallet.
func (s *Session) GetBalanceAll() (float64, error) {
	res, err := s.call("getbalance", nil)
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

// GetNewAddress returns a new bitcoin address for receiving payments.
// It is added to the address book for the given account, so payments
// received with the address will be credited to [account].
func (s *Session) GetNewAddress(account string) (string, error) {
	res, err := s.call("getnewaddress", []Data{account})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetRawChangeAddress returns a new Bitcoin address for receiving change.
// This is for use with raw transactions, not normal use.
func (s *Session) GetRawChangeAddress() (string, error) {
	res, err := s.call("getrawchangeaddress", nil)
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetUnconfirmedBalance returns the wallet’s total unconfirmed balance.
func (s *Session) GetUnconfirmedBalance() (float64, error) {
	res, err := s.call("getunconfirmedbalance", nil)
	if err != nil {
		return 0.0, err
	}
	return res.Result.(float64), nil
}

// GetWalletInfo provides information about the wallet.
func (s *Session) GetWalletInfo() (*WalletInfo, error) {
	res, err := s.call("getwalletinfo", nil)
	if err != nil {
		return nil, err
	}
	wi := new(WalletInfo)
	if ok, err := res.UnmarshalResult(wi); !ok {
		return nil, err
	}
	return wi, nil
}

// ImportAddress adds an address or pubkey script to the wallet without the
// associated private key, allowing you to watch for transactions affecting
// that address or pubkey script without being able to spend any of its outputs.
func (s *Session) ImportAddress(addr, account string, rescan bool) error {
	_, err := s.call("importaddress", []Data{addr, account, rescan})
	return err
}

// ImportPrivateKey imports a private key into your bitcoin wallet.
// Private key must be in wallet import format (Sipa) beginning
// with a '5'.
// Remarks: Requires unlocked wallet
func (s *Session) ImportPrivateKey(key string) error {
	_, err := s.call("importprivkey", []Data{key})
	return err
}

// ImportPrunedFunds imports funds without the need of a rescan. Meant for use
// with pruned wallets. Corresponding address or script must previously be
// included in wallet. The end-user is responsible to import additional
// transactions that subsequently spend the imported outputs or rescan after
// the point in the blockchain the transaction is included.
func (s *Session) ImportPrunedFunds(rtx, proof string) error {
	_, err := s.call("importprunedfunds", []Data{rtx, proof})
	return err
}

// ImportWallet imports private keys from a file in wallet dump file format
// (see the dumpwallet RPC). These keys will be added to the keys currently in
// the wallet. This call may need to rescan all or parts of the block chain for
// transactions affecting the newly-added keys, which may take several minutes.
func (s *Session) ImportWallet(file string) error {
	_, err := s.call("importwallet", []Data{file})
	return err
}

// KeypoolRefill creates a number of new Bitcoin addresses for later use.
// Remarks: Requires unlocked wallet
func (s *Session) KeypoolRefill() error {
	_, err := s.call("getnewaddress", nil)
	if err != nil {
		return err
	}
	return nil
}

// ListAccounts returns Object that has account names as keys and
// account balances as values.
func (s *Session) ListAccounts(confirmations int) (list map[string]float64, err error) {
	var res *Response
	if res, err = s.call("listaccounts", []Data{confirmations}); err == nil {
		list, _ = res.Result.(map[string]float64)
	}
	return
}

// ListAddressGroupings lists groups of addresses that may have had their
// common ownership made public by common use as inputs in the same
// transaction or from being used as change from a previous transaction.
func (s *Session) ListAddressGroupings() ([]*AddressGroup, error) {
	res, err := s.call("listaddressgroupings", nil)
	if err != nil {
		return nil, err
	}
	var ag []*AddressGroup
	if ok, err := res.UnmarshalResult(&ag); !ok {
		return nil, err
	}
	return ag, nil
}

// ListReceivedByAccount returns an array of accounts with the the
// total received and more info.
func (s *Session) ListReceivedByAccount(minConf int, includeEmpty, watchOnly bool) ([]*AccountInfo, error) {
	res, err := s.call("listreceivedbyaccount", []Data{minConf, includeEmpty, watchOnly})
	if err != nil {
		return nil, err
	}
	var rcv []*AccountInfo
	if ok, err := res.UnmarshalResult(&rcv); !ok {
		return nil, err
	}
	return rcv, nil
}

// ListReceivedByAddress returns an array of addresses with the the
// total received and more info.
func (s *Session) ListReceivedByAddress(minConf int, includeEmpty, watchOnly bool) ([]*AddressInfo, error) {
	res, err := s.call("listreceivedbyaddress", []Data{minConf, includeEmpty, watchOnly})
	if err != nil {
		return nil, err
	}
	var rcv []*AddressInfo
	if ok, err := res.UnmarshalResult(&rcv); !ok {
		return nil, err
	}
	return rcv, nil
}

// ListUnspent [minconf=1] [maxconf=999999]
// Returns an array of unspent transaction outputs in the wallet that have
// between minconf and maxconf (inclusive) confirmations. Each output is a
// 5-element object with keys: txid, output, scriptPubKey, amount,
// confirmations. txid is the hexadecimal transaction id, output is which
// output of that transaction, scriptPubKey is the hexadecimal-encoded CScript
// for that output, amount is the value of that output and confirmations is
// the transaction's depth in the chain.
func (s *Session) ListUnspent(minconf, maxconf int) ([]*Unspent, error) {
	res, err := s.call("listunspent", []Data{minconf, maxconf})
	if err != nil {
		return nil, err
	}
	var unspent []*Unspent
	if ok, err := res.UnmarshalResult(&unspent); !ok {
		return nil, err
	}
	return unspent, nil
}

// ListLockUnspent lists all temporarily locked transaction outputs.
func (s *Session) ListLockUnspent() ([]*Output, error) {
	res, err := s.call("listlockunspent", nil)
	if err != nil {
		return nil, err
	}
	var list []*Output
	if ok, err := res.UnmarshalResult(&list); !ok {
		return nil, err
	}
	return list, nil
}

// LockUnspent temporarily locks (true) or unlocks (false) specified
// transaction outputs. A locked transaction output will not be chosen
// by automatic coin selection, when spending bitcoins. Locks are
// stored in memory only. Nodes start with zero locked outputs, and
// the locked output list is always cleared (by virtue of process exit)
// when a node stops or fails.
func (s *Session) LockUnspent(lock bool, slots []*Output) error {
	_, err := s.call("lockunspent", []Data{lock, slots})
	if err != nil {
		return err
	}
	return nil
}

// Move shifts funds from one account in your wallet to another.
func (s *Session) Move(fromAccount, toAccount string, amount float64) (string, error) {
	res, err := s.call("move", []Data{fromAccount, toAccount, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// RemovePrunedFunds deletes the specified transaction from the wallet. Meant
// for use with pruned wallets and as a companion to importprunedfunds. This
// will effect wallet balances.
func (s *Session) RemovePrunedFunds(txid string) error {
	_, err := s.call("removeprunedfunds", []Data{txid})
	return err
}

// SendFrom sends a given amount (real; rounded to 8 decimal places).
// Will send the given amount to the given address, ensuring the account
// has a valid balance using [minconf] confirmations. Returns the
// transaction ID if successful.
// Remarks: Requires unlocked wallet
func (s *Session) SendFrom(fromAccount, toAddress string, amount float64) (string, error) {
	res, err := s.call("sendfrom", []Data{fromAccount, toAddress, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// SendMany sends numerous amounts to different receiving addresses.
// Amounts are double-precision floating point numbers.
// Remarks: Requires unlocked wallet
func (s *Session) SendMany(fromAccount string, targets []Balance) (string, error) {
	list := make(map[string]float64)
	for _, t := range targets {
		list[t.Address] = t.Amount
	}
	res, err := s.call("sendmany", []Data{fromAccount, list})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// SendToAddress sends an amount (real, rounded to 8 decimal places) to
// a receiving address. Returns the transaction hash if successful.
// Remarks: Requires unlocked wallet
func (s *Session) SendToAddress(addr string, amount float64) (string, error) {
	res, err := s.call("sendtoaddress", []Data{addr, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// SetAccount set a label for a Bitcoin address.
func (s *Session) SetAccount(address, label string) error {
	_, err := s.call("setaccount", []Data{address, label})
	return err
}

// SetTxFee  sets the default transaction fee for 24 hours (must be
// called every 24 hours).
func (s *Session) SetTxFee(amount float64) error {
	_, err := s.call("settxfee", []Data{amount})
	return err
}

// SignMessage signs a message with the private key of an address.
func (s *Session) SignMessage(addr, msg string) (string, error) {
	res, err := s.call("signmessage", []Data{addr, msg})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// SignMessageWithPrivKey signs a message with the private key.
func (s *Session) SignMessageWithPrivKey(key, msg string) (string, error) {
	res, err := s.call("signmessagewithprivkey", []Data{key, msg})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// ValidateAddress returns information about a Bitcoin address.
// Remarks: Requires unlocked wallet
func (s *Session) ValidateAddress(addr string) (*Validity, error) {
	res, err := s.call("validateaddress", []Data{addr})
	if err != nil {
		return nil, err
	}
	v := new(Validity)
	if ok, err := res.UnmarshalResult(v); !ok {
		return nil, err
	}
	return v, nil
}

// VerifyMessage verifies a signed message.
func (s *Session) VerifyMessage(addr, sig, msg string) (bool, error) {
	res, err := s.call("verifymessage", []Data{addr, sig, msg})
	if err != nil {
		return false, err
	}
	return res.Result.(bool), nil
}

// WalletLock removes second password from memory.
func (s *Session) WalletLock() error {
	_, err := s.call("walletlock", []Data{})
	return err
}

// WalletPassphrase stores the wallet second password in cache for
// timeout in seconds. Only required for double encrypted wallets.
func (s *Session) WalletPassphrase(passphrase string, timeout int) error {
	_, err := s.call("walletpassphrase", []Data{passphrase, timeout})
	return err
}

//======================================================================
// Various type methods related to wallets.

// GetAddress returns the Bitcoin address of an address group member.
func (a AddressDetail) GetAddress() string {
	if len(a) < 1 {
		return ""
	}
	return a[0].(string)
}

// GetBalance returns the balance of an address group member.
func (a AddressDetail) GetBalance() float64 {
	if len(a) < 2 {
		return -1.0
	}
	return a[1].(float64)
}

// GetAccount returns the account name of an address group member.
func (a AddressDetail) GetAccount() string {
	if len(a) < 3 {
		return ""
	}
	return a[2].(string)
}
