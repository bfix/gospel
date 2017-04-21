package rpc

import (
	"errors"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
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
		return "", errors.New("Unknown parameter type #2")
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
	if err = res.UnmarshalResult(saddr); err != nil {
		return nil, err
	}
	return saddr, nil
}

// DumpPrivKey reveals the private key corresponding to <bitcoinaddress>
func (s *Session) DumpPrivKey(address string, testnet bool) (*ecc.PrivateKey, error) {
	res, err := s.call("dumpprivkey", []Data{address})
	if err != nil {
		return nil, err
	}
	return util.ImportPrivateKey(res.Result.(string), testnet)
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
	label := res.Result.(string)
	return label, err
}

// GetAccountAddress returns the first Bitcoin address matching label.
func (s *Session) GetAccountAddress(label string) (string, error) {
	res, err := s.call("getaccountaddress", []Data{label})
	if err != nil {
		return "", err
	}
	addr := res.Result.(string)
	return addr, nil
}

// GetAddressesByAccount returns the an array of bitcoin addresses
// matching label.
func (s *Session) GetAddressesByAccount(label string) ([]string, error) {
	res, err := s.call("getaddressesbyaccount", []Data{label})
	if err != nil {
		return nil, err
	}
	var list []string
	val := res.Result.([]interface{})
	for _, v := range val {
		list = append(list, v.(string))
	}
	return list, nil
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
	addr := res.Result.(string)
	return addr, nil
}

// ImportPrivateKey imports a private key into your bitcoin wallet.
// Private key must be in wallet import format (Sipa) beginning
// with a '5'.
// Remarks: Requires unlocked wallet
func (s *Session) ImportPrivateKey(key string) error {
	_, err := s.call("importprivkey", []Data{key})
	if err != nil {
		return err
	}
	return nil
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
func (s *Session) ListAccounts(confirmations int) (map[string]float64, error) {
	res, err := s.call("listaccounts", []Data{confirmations})
	if err != nil {
		return nil, err
	}
	list := make(map[string]float64)
	data := res.Result.(map[string]interface{})
	for key, value := range data {
		list[key] = value.(float64)
	}
	return list, nil
}

// ListReceivedByAccount returns an array of accounts with the the
// total received and more info.
func (s *Session) ListReceivedByAccount(minConf int, includeEmpty, watchOnly bool) ([]*AccountInfo, error) {
	res, err := s.call("listreceivedbyaccount", []Data{minConf, includeEmpty, watchOnly})
	if err != nil {
		return nil, err
	}
	var rcv []*AccountInfo
	if err = res.UnmarshalResult(&rcv); err != nil {
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
	if err = res.UnmarshalResult(&rcv); err != nil {
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
func (s *Session) ListUnspent(minconf, maxconf int) ([]*Output, error) {
	res, err := s.call("listunspent", []Data{minconf, maxconf})
	if err != nil {
		return nil, err
	}
	var unspent []*Output
	if err = res.UnmarshalResult(&unspent); err != nil {
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
	if err = res.UnmarshalResult(&list); err != nil {
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

// ValidateAddress returns information about a Bitcoin address.
// Remarks: Requires unlocked wallet
func (s *Session) ValidateAddress(addr string) (*Validity, error) {
	res, err := s.call("validateaddress", []Data{addr})
	if err != nil {
		return nil, err
	}
	v := new(Validity)
	if err = res.UnmarshalResult(v); err != nil {
		return nil, err
	}
	return v, nil
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
