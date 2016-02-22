package rpc

/*
 * Bitcoin RPC session calls to running Bitcoin server
 *
 * (c) 2011-2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

///////////////////////////////////////////////////////////////////////
// import external declaratiuons

import (
	"encoding/json"
	"errors"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"io/ioutil"
	"net/http"
	"strings"
)

///////////////////////////////////////////////////////////////////////
// Call-related types

//---------------------------------------------------------------------

// Data is a generic data structure for RPC data (in/out)
type Data interface{}

//---------------------------------------------------------------------

// Request is a JSON-RPC task to a running Bitcoin server
type Request struct {
	Version string `json:"jsonrpc"` // "1.0",
	ID      string `json:"id"`      // ""
	Method  string `json:"method"`
	Params  []Data `json:"params"`
}

//---------------------------------------------------------------------

// Error is a Response-related failure code.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//---------------------------------------------------------------------

// Response is a JSON-encoded reply from a running Bitcoin server
type Response struct {
	Result Data  `json:"result"`
	Error  Error `json:"error"`
}

///////////////////////////////////////////////////////////////////////
// Helper functions to retrieve intrinsic values from generic data

// GetInt returns an integer value from named RPC data
func GetInt(d Data, key string) int {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return -1
	}
	return int(v.(float64))
}

// GetString returns a string value from named RPC data
func GetString(d Data, key string) string {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetFloat64 returns a double value from named RPC data
func GetFloat64(d Data, key string) float64 {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return 0
	}
	return v.(float64)
}

// GetBool returns a boolean value from named RPC data
func GetBool(d Data, key string) bool {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return false
	}
	return v.(bool)
}

// GetObject returns a generic value from named RPC data
func GetObject(d Data, key string) interface{} {
	return d.(map[string]interface{})[key]
}

///////////////////////////////////////////////////////////////////////

// Session type
type Session struct {
	Address string // server address/name
	User    string // user name
	Passwd  string // user password

	client *http.Client
}

//---------------------------------------------------------------------

// NewSession allocates a new Session instance for communication
func NewSession(addr, user, pw string) (*Session, error) {
	s := &Session{
		Address: addr,
		User:    user,
		Passwd:  pw,
	}
	s.client = &http.Client{}
	return s, nil
}

//---------------------------------------------------------------------
/*
 * Generic call to running server: Handles input parameters and
 * returns generic result data.
 */
func (s *Session) call(methodname string, args []Data) (result *Response, err error) {
	request := &Request{
		Version: "1.0",
		ID:      "",
		Method:  methodname,
		Params:  args,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.Address, strings.NewReader(string(data)))
	req.SetBasicAuth(s.User, s.Passwd)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}
	if response.Error.Code != 0 {
		return nil, errors.New(response.Error.Message)
	}
	return response, nil
}

///////////////////////////////////////////////////////////////////////
// Session methods (Bitcoin JSON-RPC API calls)
// Method descriptions from "http://blockchain.info/api/json_rpc_api" and
// "https://en.bitcoin.it/wiki/Raw_Transactions" (with corrections from
// the author, where required)

// BackupWallet securely copies the wallet to another file/directory.
func (s *Session) BackupWallet(dest string) error {
	_, err := s.call("backupwallet", []Data{dest})
	return err
}

//---------------------------------------------------------------------

// CreateRawTransaction [{"txid":txid,"vout":n},...] {address:amount,...}
// Create a transaction spending given inputs (array of objects containing
// transaction outputs to spend), sending to given address(es). Returns the
// hex-encoded transaction in a string. Note that the transaction's inputs
// are not signed, and it is not stored in the wallet or transmitted to the
// network.
// Also note that NO transaction validity checks are done; it is easy to
// create invalid transactions or transactions that will not be relayed/mined
// by the network because they contain insufficient fees.
func (s *Session) CreateRawTransaction(slots []Output, targets []Balance) (string, error) {
	var outs []map[string]interface{}
	for _, s := range slots {
		e := make(map[string]interface{})
		e["txid"] = s.ID
		e["vout"] = s.Vout
		outs = append(outs, e)
	}
	ins := make(map[string]interface{})
	for _, t := range targets {
		ins[t.Address] = t.Amount
	}
	res, err := s.call("createrawtransaction", []Data{outs, ins})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------

// DecodeRawTransactionAsObject <hex string>
// Returns JSON object with information about a serialized, hex-encoded
// transaction.
func (s *Session) DecodeRawTransactionAsObject(raw string) (interface{}, error) {
	res, err := s.call("decoderawtransaction", []Data{raw})
	if err != nil {
		return nil, err
	}
	return res.Result, nil
}

//---------------------------------------------------------------------

// DecodeRawTransaction <hex string>
// Returns instance with information about a serialized, hex-encoded
// transaction.
func (s *Session) DecodeRawTransaction(raw string) (*RawTransaction, error) {
	res, err := s.DecodeRawTransactionAsObject(raw)
	if err != nil {
		return nil, err
	}
	t := new(RawTransaction)
	t.ID = GetString(res, "txid")
	t.Version = GetInt(res, "version")
	t.LockTime = GetInt(res, "locktime")
	// fill input slots
	t.Vin = make([]Vinput, 0)
	data := GetObject(res, "vin").([]interface{})
	for _, d := range data {
		in := Vinput{
			ID:        GetString(d, "txid"),
			Vout:      GetInt(d, "vout"),
			ScriptSig: GetString(GetObject(d, "scriptSig"), "hex"),
			Sequence:  GetInt(d, "sequence"),
		}
		t.Vin = append(t.Vin, in)
	}
	// fill output slots
	t.Vout = make([]Voutput, 0)
	data = GetObject(res, "vout").([]interface{})
	for _, d := range data {
		script := GetObject(d, "scriptPubKey")
		out := Voutput{
			Value:        GetFloat64(d, "value"),
			N:            GetInt(d, "n"),
			ScriptPubKey: GetString(script, "hex"),
			ReqSigs:      GetInt(script, "reqSigs"),
			Type:         GetString(script, "type"),
			Addresses:    make([]string, 0),
		}
		list := GetObject(script, "addresses").([]interface{})
		for _, l := range list {
			out.Addresses = append(out.Addresses, l.(string))
		}
		t.Vout = append(t.Vout, out)
	}
	// return assembled raw transaction
	return t, nil
}

//---------------------------------------------------------------------

// DumpPrivKey reveals the private key corresponding to <bitcoinaddress>
func (s *Session) DumpPrivKey(address string, testnet bool) (*ecc.PrivateKey, error) {
	res, err := s.call("dumpprivkey", []Data{address})
	if err != nil {
		return nil, err
	}
	return util.ImportPrivateKey(res.Result.(string), testnet)
}

//---------------------------------------------------------------------

// GetAccount returns the label for a Bitcoin address.
func (s *Session) GetAccount(address string) (string, error) {
	res, err := s.call("getaccount", []Data{address})
	if err != nil {
		return "", err
	}
	label := res.Result.(string)
	return label, err
}

//---------------------------------------------------------------------

// GetAccountAddress returns the first Bitcoin address matching label.
func (s *Session) GetAccountAddress(label string) (string, error) {
	res, err := s.call("getaccountaddress", []Data{label})
	if err != nil {
		return "", err
	}
	addr := res.Result.(string)
	return addr, nil
}

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

// GetBalance returns the balance in the account.
func (s *Session) GetBalance(label string) (float64, error) {
	res, err := s.call("getbalance", []Data{label})
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------

// GetBalanceAll returns the balance in all accounts in the wallet.
func (s *Session) GetBalanceAll() (float64, error) {
	res, err := s.call("getbalance", nil)
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------

// GetBlock returns information about the given block hash.
func (s *Session) GetBlock(hash string) (*Block, error) {
	res, err := s.call("getblock", []Data{hash})
	if err != nil {
		return nil, err
	}
	block := new(Block)
	block.IDList = make([]string, 0)
	list := GetObject(res.Result, "tx").([]interface{})
	for _, txid := range list {
		block.IDList = append(block.IDList, txid.(string))
	}
	block.Time = GetInt(res.Result, "time")
	block.Height = GetInt(res.Result, "height")
	block.Nonce = GetInt(res.Result, "nonce")
	block.Hash = GetString(res.Result, "hash")
	block.NextBlockHash = GetString(res.Result, "nextblockhash")
	block.PreviousBlockHash = GetString(res.Result, "previousblockhash")
	block.Bits = GetString(res.Result, "bits")
	block.Difficulty = GetInt(res.Result, "difficulty")
	block.MerkleRoot = GetString(res.Result, "merkleroot")
	block.Version = GetInt(res.Result, "version")
	block.Size = GetInt(res.Result, "size")
	return block, nil
}

//---------------------------------------------------------------------

// GetBlockCount returns the number of blocks in the longest
// block chain.
func (s *Session) GetBlockCount() (int, error) {
	res, err := s.call("getblockcount", nil)
	if err != nil {
		return -1, err
	}
	num := int(res.Result.(float64))
	return num, err
}

//---------------------------------------------------------------------

// GetBlockHash returns hash of block in best-block-chain at height.
func (s *Session) GetBlockHash(height int) (string, error) {
	res, err := s.call("getblockhash", []Data{height})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------

// GetConnectionCount returns the number of connections to other nodes.
func (s *Session) GetConnectionCount() (int, error) {
	res, err := s.call("getconnectioncount", nil)
	if err != nil {
		return -1, err
	}
	return int(res.Result.(float64)), nil
}

//---------------------------------------------------------------------

// GetDifficulty returns the proof-of-work difficulty as a multiple
// of the minimum difficulty.
func (s *Session) GetDifficulty() (float64, error) {
	res, err := s.call("getdifficulty", nil)
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------

// GetInfo returns an object containing various state info.
func (s *Session) GetInfo() (*Info, error) {
	res, err := s.call("getinfo", nil)
	if err != nil {
		return nil, err
	}
	info := new(Info)
	info.Version = GetInt(res.Result, "version")
	info.ProtocolVersion = GetInt(res.Result, "protocolversion")
	info.WalletVersion = GetInt(res.Result, "walletversion")
	info.Connections = GetInt(res.Result, "connections")
	info.KeyPoolSize = GetInt(res.Result, "keypoolsize")
	info.TimeOffset = GetInt(res.Result, "timeoffset")
	info.KeyPoolOldest = GetInt(res.Result, "keypoololdest")
	info.Balance = GetFloat64(res.Result, "balance")
	info.Errors = GetString(res.Result, "errors")
	info.PayTxFee = GetFloat64(res.Result, "paytxfee")
	info.Proxy = GetString(res.Result, "proxy")
	info.TestNet = GetBool(res.Result, "testnet")
	info.Difficulty = GetFloat64(res.Result, "difficulty")
	info.Blocks = GetInt(res.Result, "blocks")
	return info, err
}

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

// GetRawTransaction returns a serialized, hex-encoded data for
// transaction txid.
func (s *Session) GetRawTransaction(txid string) (string, error) {
	res, err := s.call("getrawtransaction", []Data{txid})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------

// GetTransaction returns an object about the given transaction hash.
func (s *Session) GetTransaction(hash string) (*Transaction, error) {
	res, err := s.call("gettransaction", []Data{hash})
	if err != nil {
		return nil, err
	}
	t := &Transaction{
		Amount:        GetFloat64(res.Result, "amount"),
		Fee:           GetFloat64(res.Result, "fee"),
		BlockIndex:    GetInt(res.Result, "blockindex"),
		Confirmations: GetInt(res.Result, "confirmations"),
		ID:            GetString(res.Result, "txid"),
		BlockHash:     GetString(res.Result, "blockhash"),
		Time:          GetInt(res.Result, "time"),
		BlockTime:     GetInt(res.Result, "blocktime"),
		TimeReceived:  GetInt(res.Result, "timereceived"),
	}
	return t, nil
}

//---------------------------------------------------------------------

// KeypoolRefill creates a number of new Bitcoin addresses for later use.
// Remarks: Requires unlocked wallet
func (s *Session) KeypoolRefill() error {
	_, err := s.call("getnewaddress", nil)
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------

// ImportPrivateKey imports a private key into your bitcoin wallet.
// Private key must be in wallet import format (Sipa) beginning
// with a '5'.
// Remarks: Requires unlocked wallet
func (s *Session) ImportPrivateKey(key string) error {
	_, err := s.call("importprivatekey", []Data{key})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

// ListReceivedByAccount returns an array of accounts with the the
// total received and more info.
func (s *Session) ListReceivedByAccount(minConf int, includeEmpty bool) ([]Received, error) {
	res, err := s.call("listreceivedbyaccount", []Data{minConf, includeEmpty})
	if err != nil {
		return nil, err
	}
	var rcv []Received
	data := res.Result.([]interface{})
	for _, entry := range data {
		r := Received{
			Account:       GetString(entry, "account"),
			Label:         GetString(entry, "label"),
			Address:       "<unknown>",
			Amount:        GetFloat64(entry, "amount"),
			Confirmations: GetInt(entry, "confirmations"),
		}
		rcv = append(rcv, r)
	}
	return rcv, nil
}

//---------------------------------------------------------------------

// ListReceivedByAddress returns an array of addresses with the the
// total received and more info.
func (s *Session) ListReceivedByAddress(minConf int, includeEmpty bool) ([]Received, error) {
	res, err := s.call("listreceivedbyaddress", []Data{minConf, includeEmpty})
	if err != nil {
		return nil, err
	}
	var rcv []Received
	data := res.Result.([]interface{})
	for _, entry := range data {
		r := Received{
			Account:       GetString(entry, "account"),
			Label:         "<unknown>",
			Address:       GetString(entry, "address"),
			Amount:        GetFloat64(entry, "amount"),
			Confirmations: GetInt(entry, "confirmations"),
		}
		rcv = append(rcv, r)
	}
	return rcv, nil
}

//---------------------------------------------------------------------

// ListSinceBlock gets all transactions in blocks since block
// [blockhash] (not inclusive), or all transactions if omitted.
// Max 25 at a time.
func (s *Session) ListSinceBlock(hash string, minConf int) ([]Transaction, string, error) {
	res, err := s.call("listsinceblock", []Data{hash, minConf})
	if err != nil {
		return nil, "", err
	}
	last := GetString(res.Result, "lastblock")
	data := GetObject(res.Result, "transactions").([]interface{})
	var list []Transaction
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			ID:            GetString(e, "txid"),
			BlockHash:     GetString(e, "blockhash"),
			Time:          GetInt(e, "time"),
			BlockTime:     GetInt(e, "blocktime"),
			TimeReceived:  GetInt(e, "timereceived"),
		}
		list = append(list, t)
	}
	return list, last, nil
}

//---------------------------------------------------------------------

// ListTransactions returns up to [count] most recent transactions
// skipping the first [from] transactions for account [account].
func (s *Session) ListTransactions(accnt string, count, offset int) ([]Transaction, error) {
	res, err := s.call("listtransactions", []Data{accnt, count, offset})
	if err != nil {
		return nil, err
	}
	var list []Transaction
	data := res.Result.([]interface{})
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			ID:            GetString(e, "txid"),
			BlockHash:     GetString(e, "blockhash"),
			Time:          GetInt(e, "time"),
			BlockTime:     GetInt(e, "blocktime"),
			TimeReceived:  GetInt(e, "timereceived"),
		}
		list = append(list, t)
	}
	return list, nil
}

//---------------------------------------------------------------------

// ListAllTransactions returns up to [count] most recent transactions
// skipping the first [from] transactions for all accounts.
func (s *Session) ListAllTransactions(count, offset int) ([]Transaction, error) {
	res, err := s.call("listtransactions", []Data{nil, count, offset})
	if err != nil {
		return nil, err
	}
	var list []Transaction
	data := res.Result.([]interface{})
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			ID:            GetString(e, "txid"),
			BlockHash:     GetString(e, "blockhash"),
			Time:          GetInt(e, "time"),
			BlockTime:     GetInt(e, "blocktime"),
			TimeReceived:  GetInt(e, "timereceived"),
		}
		list = append(list, t)
	}
	return list, nil
}

//---------------------------------------------------------------------

// ListUnspent [minconf=1] [maxconf=999999]
// Returns an array of unspent transaction outputs in the wallet that have
// between minconf and maxconf (inclusive) confirmations. Each output is a
// 5-element object with keys: txid, output, scriptPubKey, amount,
// confirmations. txid is the hexadecimal transaction id, output is which
// output of that transaction, scriptPubKey is the hexadecimal-encoded CScript
// for that output, amount is the value of that output and confirmations is
// the transaction's depth in the chain.
func (s *Session) ListUnspent(minconf, maxconf int) ([]Unspent, error) {
	args := []Data{minconf, maxconf}
	res, err := s.call("listunspent", args)
	if err != nil {
		return nil, err
	}
	data := res.Result.([]interface{})
	num := len(data)
	unspent := make([]Unspent, num)
	for i, entry := range data {
		unspent[i].ID = GetString(entry, "txid")
		unspent[i].Amount = GetFloat64(entry, "amount")
		unspent[i].Vout = GetInt(entry, "vout")
		unspent[i].Confirmations = GetInt(entry, "confirmations")
		unspent[i].ScriptPubKey = GetString(entry, "scriptPubKey")
	}
	return unspent, nil
}

//---------------------------------------------------------------------

// ListLockUnspent lists all temporarily locked transaction outputs.
func (s *Session) ListLockUnspent() ([]Output, error) {
	res, err := s.call("listlockunspent", nil)
	if err != nil {
		return nil, err
	}
	var list []Output
	data := res.Result.([]interface{})
	for _, d := range data {
		o := Output{
			ID:   GetString(d, "txid"),
			Vout: GetInt(d, "vout"),
		}
		list = append(list, o)
	}
	return list, nil
}

//---------------------------------------------------------------------

// LockUnspent temporarily locks (true) or unlocks (false) specified
// transaction outputs. A locked transaction output will not be chosen
// by automatic coin selection, when spending bitcoins. Locks are
// stored in memory only. Nodes start with zero locked outputs, and
// the locked output list is always cleared (by virtue of process exit)
// when a node stops or fails.
func (s *Session) LockUnspent(lock bool, slots []Output) error {
	var list []map[string]interface{}
	for _, s := range slots {
		e := make(map[string]interface{})
		e["txid"] = s.ID
		e["vout"] = s.Vout
		list = append(list, e)
	}
	_, err := s.call("lockunspent", []Data{lock, list})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------

// Move shifts funds from one account in your wallet to another.
func (s *Session) Move(fromAccount, toAccount string, amount float64) (string, error) {
	res, err := s.call("move", []Data{fromAccount, toAccount, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

// SetAccount set a label for a Bitcoin address.
func (s *Session) SetAccount(address, label string) error {
	_, err := s.call("setaccount", []Data{address, label})
	return err
}

//---------------------------------------------------------------------

// SetTxFee  sets the default transaction fee for 24 hours (must be
// called every 24 hours).
func (s *Session) SetTxFee(amount float64) error {
	_, err := s.call("settxfee", []Data{amount})
	return err
}

//---------------------------------------------------------------------

// SendRawTransaction submits a raw transaction (serialized, hex-encoded)
// to local node and network. Returns transaction id, or an error if the
// transaction is invalid for any reason.
func (s *Session) SendRawTransaction(raw string) error {
	_, err := s.call("sendrawtransaction", []Data{raw})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------

// SignRawTransaction <hex string> [{"txid":txid,"vout":n,"scriptPubKey":hex},...] [<privatekey1>,...] [sighash="ALL"]
// Sign as many inputs as possible for raw transaction (serialized,
// hex-encoded). The first argument may be several variations of the same
// transaction concatenated together; signatures from all of them will be
// combined together, along with signatures for keys in the local wallet. The
// optional second argument is an array of parent transaction outputs, so you
// can create a chain of raw transactions that depend on each other before
// sending them to the network. Third optional argument is an array of
// base58-encoded private keys that, if given, will be the only keys used to
// sign the transaction. The fourth optional argument is a string that specifies
// how the signature hash is computed, and can be "ALL", "NONE", "SINGLE",
// "ALL|ANYONECANPAY", "NONE|ANYONECANPAY", or "SINGLE|ANYONECANPAY".
// Returns json object with keys:
//     hex : raw transaction with signature(s) (hex-encoded string)
//     complete : 1 if rawtx is completely signed, 0 if signatures are missing.
// If no private keys are given and the wallet is locked, requires that the
// wallet be unlocked with walletpassphrase first.
func (s *Session) SignRawTransaction(raw string, ins []Output, keys []string, mode string) (string, bool, error) {
	var inList [](map[string]interface{})
	if len(ins) == 0 {
		inList = nil
	} else {
		for _, i := range ins {
			e := make(map[string]interface{})
			e["txid"] = i.ID
			e["vout"] = i.Vout
			e["scriptPubKey"] = i.ScriptPubKey
			e["redeemScript"] = i.RedeemScript
			inList = append(inList, e)
		}
	}
	var keyList []interface{}
	if len(keys) == 0 {
		keyList = nil
	} else {
		for _, k := range keys {
			keyList = append(keyList, k)
		}
	}
	res, err := s.call("signrawtransaction", []Data{raw, inList, keyList, mode})
	if err != nil {
		return "", false, err
	}
	signed := GetString(res.Result, "hex")
	complete := GetBool(res.Result, "complete")
	return signed, complete, nil
}

//---------------------------------------------------------------------

// ValidateAddress returns information about a Bitcoin address.
// Remarks: Requires unlocked wallet
func (s *Session) ValidateAddress(addr string) (*Validity, error) {
	res, err := s.call("validateaddress", []Data{addr})
	if err != nil {
		return nil, err
	}
	v := &Validity{
		Address:      GetString(res.Result, "address"),
		Account:      GetString(res.Result, "account"),
		PubKey:       GetString(res.Result, "pubkey"),
		IsCompressed: GetBool(res.Result, "iscompressed"),
		IsMine:       GetBool(res.Result, "ismine"),
		IsValid:      GetBool(res.Result, "isvalid"),
	}
	return v, nil
}

//---------------------------------------------------------------------

// WalletLock removes second password from memory.
func (s *Session) WalletLock() error {
	_, err := s.call("walletlock", []Data{})
	return err
}

//---------------------------------------------------------------------

// WalletPassphrase stores the wallet second password in cache for
// timeout in seconds. Only required for double encrypted wallets.
func (s *Session) WalletPassphrase(passphrase string, timeout int) error {
	_, err := s.call("walletpassphrase", []Data{passphrase, timeout})
	return err
}
