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
package rpc

///////////////////////////////////////////////////////////////////////
// import external declaratiuons

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

///////////////////////////////////////////////////////////////////////
// Call-related types

//---------------------------------------------------------------------
/*
 * Generic RPC data (in/out)
 */
type RPC_Data interface{}

//---------------------------------------------------------------------
/*
 * JSON-RPC request to running Bitcoin server
 */
type RPC_Request struct {
	Version string     `json:"jsonrpc"` // "1.0",
	Id      string     `json:"id"`      // ""
	Method  string     `json:"method"`
	Params  []RPC_Data `json:"params"`
}

//---------------------------------------------------------------------
/*
 * Response Error type
 */
type RPC_Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//---------------------------------------------------------------------
/*
 * JSON-RPC response from running Bitcoin server
 */
type RPC_Response struct {
	Result RPC_Data  `json:"result"`
	Error  RPC_Error `json:"error"`
}

///////////////////////////////////////////////////////////////////////
// Helper functions to retrieve intrinsic values from generic data

// get integer value
func GetInt(d RPC_Data, key string) int {
	return int(d.(map[string]interface{})[key].(float64))
}

// get string value
func GetString(d RPC_Data, key string) string {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return ""
	}
	return v.(string)
}

// egt double value
func GetFloat64(d RPC_Data, key string) float64 {
	v := d.(map[string]interface{})[key]
	if v == nil {
		return 0
	}
	return v.(float64)
}

// get boolean value
func GetBool(d RPC_Data, key string) bool {
	return d.(map[string]interface{})[key].(bool)
}

// get generic value
func GetObject(d RPC_Data, key string) interface{} {
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
/*
 * Get a new Session instance for communication
 */
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
func (s *Session) call(methodname string, args []RPC_Data) (result *RPC_Response, err error) {
	request := &RPC_Request{
		Version: "1.0",
		Id:      "",
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
	response := new(RPC_Response)
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

/*
 * Method: backupwallet
 * Parameters: String destination
 * Description: Securely copies the wallet to another file/directory.
 *              Uploads the wallet to Google Drive, Dropbox or Email.
 *              Must login through the web interface first.
 * Returns: error
 */
func (s *Session) BackupWallet(dest string) error {
	_, err := s.call("backupwallet", []RPC_Data{dest})
	return err
}

//---------------------------------------------------------------------
/*
 * createrawtransaction [{"txid":txid,"vout":n},...] {address:amount,...}
 * Create a transaction spending given inputs (array of objects containing
 * transaction outputs to spend), sending to given address(es). Returns the
 * hex-encoded transaction in a string. Note that the transaction's inputs
 * are not signed, and it is not stored in the wallet or transmitted to the
 * network.
 * Also note that NO transaction validity checks are done; it is easy to
 * create invalid transactions or transactions that will not be relayed/mined
 * by the network because they contain insufficient fees.
 */
func (s *Session) CreateRawTransaction(slots []Output, targets []Balance) (string, error) {
	outs := make([]map[string]interface{}, 0)
	for _, s := range slots {
		e := make(map[string]interface{})
		e["txid"] = s.Id
		e["vout"] = s.Vout
		outs = append(outs, e)
	}
	ins := make(map[string]interface{})
	for _, t := range targets {
		ins[t.Address] = t.Amount
	}
	res, err := s.call("createrawtransaction", []RPC_Data{outs, ins})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * decoderawtransaction <hex string>
 * Returns JSON object with information about a serialized, hex-encoded transaction.
 */
func (s *Session) DecodeRawTransaction(raw string) (interface{}, error) {
	res, err := s.call("decoderawtransaction", []RPC_Data{raw})
	if err != nil {
		return nil, err
	}
	return res.Result, nil
}

//---------------------------------------------------------------------
/*
 * Method: getaccount
 * Parameters: (String bitcoinAddress)
 * Description: Get the label for a bitcoin address.
 * Returns: String, error
 */
func (s *Session) GetAccount(address string) (string, error) {
	res, err := s.call("getaccount", []RPC_Data{address})
	if err != nil {
		return "", err
	}
	label := res.Result.(string)
	return label, err
}

//---------------------------------------------------------------------
/*
 * Method: getaccountaddress
 * Parameters: (String label)
 * Description: Get the first bitcoin address matching label.
 * Returns: String, error
 */
func (s *Session) GetAccountAddress(label string) (string, error) {
	res, err := s.call("getaccountaddress", []RPC_Data{label})
	if err != nil {
		return "", err
	}
	addr := res.Result.(string)
	return addr, nil
}

//---------------------------------------------------------------------
/*
 * Method: getaddressesbyaccount
 * Parameters: (String label)
 * Description: Get the an array of bitcoin addresses matching label.
 * Returns: [string], error
 */
func (s *Session) GetAddressesByAccount(label string) ([]string, error) {
	res, err := s.call("getaddressesbyaccount", []RPC_Data{label})
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	val := res.Result.([]interface{})
	for _, v := range val {
		list = append(list, v.(string))
	}
	return list, nil
}

//---------------------------------------------------------------------
/*
 * Method: getbalance
 * Parameters: (String account = null, int minimumConfirmations = 1)
 * Description: Returns the balance in the account.
 * Returns: double, error
 */
func (s *Session) GetBalance(label string) (float64, error) {
	res, err := s.call("getbalance", []RPC_Data{label})
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------
/*
 * Method: getbalance
 * Parameters: (String account = null, int minimumConfirmations = 1)
 * Description: Returns the balance in all accounts in the wallet.
 * Returns: double, error
 */
func (s *Session) GetBalanceAll() (float64, error) {
	res, err := s.call("getbalance", nil)
	if err != nil {
		return 0, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------
/*
 * Method: getblock
 * Parameters: (String blockHash)
 * Description: Returns information about the given block hash.
 * Returns: Block, error
 */
func (s *Session) GetBlock(hash string) (*Block, error) {
	res, err := s.call("getblock", []RPC_Data{hash})
	if err != nil {
		return nil, err
	}
	block := new(Block)
	block.IdList = make([]string, 0)
	list := GetObject(res.Result, "tx").([]interface{})
	for _, txid := range list {
		block.IdList = append(block.IdList, txid.(string))
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
/*
 * Method: getblockcount
 * Parameters: None
 * Description: Returns the number of blocks in the longest block chain.
 * Returns: int, error
 */
func (s *Session) GetBlockCount() (int, error) {
	res, err := s.call("getblockcount", nil)
	if err != nil {
		return -1, err
	}
	num := int(res.Result.(float64))
	return num, err
}

//---------------------------------------------------------------------
/*
 * Method: getblockhash
 * Alias: getblocknumber
 * Parameters: (blockHeight)
 * Description: Returns hash of block in best-block-chain at height.
 * Returns: String, error
 */
func (s *Session) GetBlockHash(height int) (string, error) {
	res, err := s.call("getblockhash", []RPC_Data{height})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: getconnectioncount
 * Parameters: None
 * Description: Returns the number of connections to other nodes.
 * Returns: int, error
 */
func (s *Session) GetConnectionCount() (int, error) {
	res, err := s.call("getconnectioncount", nil)
	if err != nil {
		return -1, err
	}
	return int(res.Result.(float64)), nil
}

//---------------------------------------------------------------------
/*
 * Method: getdifficulty
 * Parameters: None
 * Description: Returns the proof-of-work difficulty as a multiple of the minimum difficulty.
 * Returns: double, error
 */
func (s *Session) GetDifficulty() (float64, error) {
	res, err := s.call("getdifficulty", nil)
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

//---------------------------------------------------------------------
/*
 * Method: getinfo
 * Parameters: None
 * Description: Returns an object containing various state info.
 * Returns: Info reference, error
 */
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
/*
 * Method: getnewaddress
 * Parameters: (String label = null)
 * Description: Returns a new bitcoin address for receiving payments. It is
 *              added to the address book for the given account, so payments
 *              received with the address will be credited to [account].
 * Returns: String, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) GetNewAddress(account string) (string, error) {
	res, err := s.call("getnewaddress", []RPC_Data{account})
	if err != nil {
		return "", err
	}
	addr := res.Result.(string)
	return addr, nil
}

//---------------------------------------------------------------------
/*
 * getrawtransaction <txid>
 * returns serialized, hex-encoded data for transaction txid.
 */
func (s *Session) GetRawTransaction(txid string) (string, error) {
	res, err := s.call("getrawtransaction", []RPC_Data{txid})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: gettransaction
 * Parameters: (String hash)
 * Description: Returns an object about the given transaction hash.
 * Returns: Transaction ref, error
 */
func (s *Session) GetTransaction(hash string) (*Transaction, error) {
	res, err := s.call("gettransaction", []RPC_Data{hash})
	if err != nil {
		return nil, err
	}
	t := &Transaction{
		Amount:        GetFloat64(res.Result, "amount"),
		Fee:           GetFloat64(res.Result, "fee"),
		BlockIndex:    GetInt(res.Result, "blockindex"),
		Confirmations: GetInt(res.Result, "confirmations"),
		Id:            GetString(res.Result, "txid"),
		BlockHash:     GetString(res.Result, "blockhash"),
		Time:          GetInt(res.Result, "time"),
		BlockTime:     GetInt(res.Result, "blocktime"),
		TimeReceived:  GetInt(res.Result, "timereceived"),
	}
	return t, nil
}

//---------------------------------------------------------------------
/*
 * Method: importprivkey
 * Parameters: (String privateKey)
 * Description: Import a private key into your bitcoin wallet. Private key
 *              must be in wallet import format (Sipa) beginning with a '5'.
 * Returns: error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) ImportPrivateKey(key string) error {
	_, err := s.call("importprivatekey", []RPC_Data{key})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------
/*
 * Method: listaccounts
 * Parameters: (int confirmations = 1)
 * Description: Returns Object that has account names as keys, account balances as values.
 * Returns: [{account,balance}]
 */
func (s *Session) ListAccounts(confirmations int) (map[string]float64, error) {
	res, err := s.call("listaccounts", []RPC_Data{confirmations})
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
/*
 * Method: listreceivedbyaccount
 * Parameters: (int minConfirmations = 1, boolean includeempty = false)
 * Description: Returns An array of accounts with the the total received and more info.
 * Returns: []Received, error
 */
func (s *Session) ListReceivedByAccount(minConf int, includeEmpty bool) ([]Received, error) {
	res, err := s.call("listreceivedbyaccount", []RPC_Data{minConf, includeEmpty})
	if err != nil {
		return nil, err
	}
	rcv := make([]Received, 0)
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
/*
 * Method: listreceivedbyaddress
 * Parameters: (int minConfirmations = 1, boolean includeempty = false)
 * Description: Returns An array of addresses with the the total received and more info.
 * Returns: []Received, error
 */
func (s *Session) ListReceivedByAddress(minConf int, includeEmpty bool) ([]Received, error) {
	res, err := s.call("listreceivedbyaddress", []RPC_Data{minConf, includeEmpty})
	if err != nil {
		return nil, err
	}
	rcv := make([]Received, 0)
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
/*
 * Method: listsinceblock
 * Parameters: (String blockHash, int minConfirmations = 1)
 * Description: Get all transactions in blocks since block [blockhash] (not inclusive),
 *              or all transactions if omitted. Max 25 at a time.
 * Returns: lastBlock, []Transaction, error
 */
func (s *Session) ListSinceBlock(hash string, minConf int) ([]Transaction, string, error) {
	res, err := s.call("listsinceblock", []RPC_Data{hash, minConf})
	if err != nil {
		return nil, "", err
	}
	last := GetString(res.Result, "lastblock")
	data := GetObject(res.Result, "transactions").([]interface{})
	list := make([]Transaction, 0)
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			Id:            GetString(e, "txid"),
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
/*
 * Method: listtransactions
 * Parameters: (String account, int count = 25, int offset = 0)
 * Description: Returns up to [count] most recent transactions skipping the first [from] transactions for account [account].
 * Returns: []Transaction, error
 */
func (s *Session) ListTransactions(accnt string, count, offset int) ([]Transaction, error) {
	res, err := s.call("listtransactions", []RPC_Data{accnt, count, offset})
	if err != nil {
		return nil, err
	}
	list := make([]Transaction, 0)
	data := res.Result.([]interface{})
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			Id:            GetString(e, "txid"),
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
/*
 * Method: listtransactions
 * Parameters: (int count = 25, int offset = 0)
 * Description: Returns up to [count] most recent transactions skipping the first [from] transactions for all accounts.
 * Returns: []Transaction, error
 */
func (s *Session) ListAllTransactions(count, offset int) ([]Transaction, error) {
	res, err := s.call("listtransactions", []RPC_Data{nil, count, offset})
	if err != nil {
		return nil, err
	}
	list := make([]Transaction, 0)
	data := res.Result.([]interface{})
	for _, e := range data {
		t := Transaction{
			Amount:        GetFloat64(e, "amount"),
			Fee:           GetFloat64(e, "fee"),
			BlockIndex:    GetInt(e, "blockindex"),
			Confirmations: GetInt(e, "confirmations"),
			Id:            GetString(e, "txid"),
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
/*
 * listunspent [minconf=1] [maxconf=999999]
 * Returns an array of unspent transaction outputs in the wallet that have
 * between minconf and maxconf (inclusive) confirmations. Each output is a
 * 5-element object with keys: txid, output, scriptPubKey, amount,
 * confirmations. txid is the hexadecimal transaction id, output is which
 * output of that transaction, scriptPubKey is the hexadecimal-encoded CScript
 * for that output, amount is the value of that output and confirmations is
 * the transaction's depth in the chain.
 */
func (s *Session) ListUnspent(minconf, maxconf int) ([]Unspent, error) {
	args := []RPC_Data{minconf, maxconf}
	res, err := s.call("listunspent", args)
	if err != nil {
		return nil, err
	}
	data := res.Result.([]interface{})
	num := len(data)
	unspent := make([]Unspent, num)
	for i, entry := range data {
		unspent[i].Id = GetString(entry, "txid")
		unspent[i].Amount = GetFloat64(entry, "amount")
		unspent[i].Vout = GetInt(entry, "vout")
		unspent[i].Confirmations = GetInt(entry, "confirmations")
		unspent[i].ScriptPubkey = GetString(entry, "scriptPubKey")
	}
	return unspent, nil
}

//---------------------------------------------------------------------
/*
 * ListLockUnspent: List all temporarily locked transaction outputs.
 */
func (s *Session) ListLockUnspent() ([]Output, error) {
	res, err := s.call("listlockunspent", nil)
	if err != nil {
		return nil, err
	}
	list := make([]Output, 0)
	data := res.Result.([]interface{})
	for _, d := range data {
		o := Output{
			Id:   GetString(d, "txid"),
			Vout: GetInt(d, "vout"),
		}
		list = append(list, o)
	}
	return list, nil
}

//---------------------------------------------------------------------
/*
 * LockUspent: Temporarily lock (true) or unlock (false) specified transaction
 * outputs. A locked transaction output will not be chosen by automatic coin
 * selection, when spending bitcoins. Locks are stored in memory only. Nodes
 * start with zero locked outputs, and the locked output list is always cleared
 * (by virtue of process exit) when a node stops or fails.
 */
func (s *Session) LockUnspent(lock bool, slots []Output) error {
	list := make([]map[string]interface{}, 0)
	for _, s := range slots {
		e := make(map[string]interface{})
		e["txid"] = s.Id
		e["vout"] = s.Vout
		list = append(list, e)
	}
	_, err := s.call("lockunspent", []RPC_Data{lock, list})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------
/*
 * Method: move
 * Parameters: (String fromAccount, String toAccount, long amount)
 * Description: Move funds from one account in your wallet to another.
 * Returns: String, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) Move(fromAccount, toAccount string, amount float64) (string, error) {
	res, err := s.call("move", []RPC_Data{fromAccount, toAccount, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: sendfrom
 * Parameters: (String fromAccount, String bitcoinAddress, long amount)
 * Description: amount is a real and is rounded to 8 decimal places. Will send the given amount
 *              to the given address, ensuring the account has a valid balance using [minconf]
 *              confirmations. Returns the transaction ID if successful.
 * Returns: String, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) SendFrom(fromAccount, toAddress string, amount float64) (string, error) {
	res, err := s.call("sendfrom", []RPC_Data{fromAccount, toAddress, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: sendmany
 * Parameters: (String fromAccount, []addressAmountPairs)
 * Address Amount Pairs: {address:amount,...} e.g. {"1yeTWjh876opYp6R5VRj8rzkLFPE4dP3Uw":10,"1yeTWjh876opYp6R5VRj8rzkLFPE4dP3Uw":15}
 * Description: amounts are double-precision floating point numbers.
 * Returns: String, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) SendMany(fromAccount string, targets []Balance) (string, error) {
	list := make(map[string]float64)
	for _, t := range targets {
		list[t.Address] = t.Amount
	}
	res, err := s.call("sendmany", []RPC_Data{fromAccount, list})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: sendtoaddress
 * Parameters: (String bitcoinAddress, double amount)
 * Description: amount is a real and is rounded to 8 decimal places. Returns the transaction hash if successful. .
 * Returns: String, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) SendToAddress(addr string, amount float64) (string, error) {
	res, err := s.call("sendtoaddress", []RPC_Data{addr, amount})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

//---------------------------------------------------------------------
/*
 * Method: setaccount
 " Parameters: (String bitcoinAddress, String label)
 " Description: Set the label for a bitcoin address.
 " Returns: error
*/
func (s *Session) SetAccount(address, label string) error {
	_, err := s.call("setaccount", []RPC_Data{address, label})
	return err
}

//---------------------------------------------------------------------
/*
 * Method: settxfee
 * Parameters: (double amount)
 * Description: Sets the default transaction fee for 24 hours (must be called every 24 hours).
 * Returns: error
 */
func (s *Session) SetTxFee(amount float64) error {
	_, err := s.call("settxfee", []RPC_Data{amount})
	return err
}

//---------------------------------------------------------------------
/*
 * sendrawtransaction <hex string>
 * Submits raw transaction (serialized, hex-encoded) to local node and network.
 * Returns transaction id, or an error if the transaction is invalid for any
 * reason.
 */
func (s *Session) SendRawTransaction(raw string) error {
	_, err := s.call("sendrawtransaction", []RPC_Data{raw})
	if err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------
/*
 * signrawtransaction <hex string> [{"txid":txid,"vout":n,"scriptPubKey":hex},...] [<privatekey1>,...] [sighash="ALL"]
 * Sign as many inputs as possible for raw transaction (serialized,
 * hex-encoded). The first argument may be several variations of the same
 * transaction concatenated together; signatures from all of them will be
 * combined together, along with signatures for keys in the local wallet. The
 * optional second argument is an array of parent transaction outputs, so you
 * can create a chain of raw transactions that depend on each other before
 * sending them to the network. Third optional argument is an array of
 * base58-encoded private keys that, if given, will be the only keys used to
 * sign the transaction. The fourth optional argument is a string that specifies
 * how the signature hash is computed, and can be "ALL", "NONE", "SINGLE",
 * "ALL|ANYONECANPAY", "NONE|ANYONECANPAY", or "SINGLE|ANYONECANPAY".
 * Returns json object with keys:
 *     hex : raw transaction with signature(s) (hex-encoded string)
 *     complete : 1 if rawtx is completely signed, 0 if signatures are missing.
 * If no private keys are given and the wallet is locked, requires that the
 * wallet be unlocked with walletpassphrase first.
 */
func (s *Session) SignRawTransaction(raw string) (string, bool, error) {
	res, err := s.call("signrawtransaction", []RPC_Data{raw})
	if err != nil {
		return "", false, err
	}
	signed := GetString(res.Result, "hex")
	complete := GetBool(res.Result, "complete")
	return signed, complete, nil
}

//---------------------------------------------------------------------
/*
 * Method: validateaddress
 * Parameters: (String bitcoinAddress)
 * Description: Return information about bitcoinaddress.
 * Returns: Validity ref, error
 * Remarks: Requires unlocked wallet
 */
func (s *Session) ValidateAddress(addr string) (*Validity, error) {
	res, err := s.call("validateaddress", []RPC_Data{addr})
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
/*
 * Method: walletlock
 * Parameters: None
 * Description: Remove second password from memory.
 * Returns: error
 */
func (s *Session) WalletLock() error {
	_, err := s.call("walletlock", []RPC_Data{})
	return err
}

//---------------------------------------------------------------------
/*
 * Method: walletpassphrase
 * Parameters: (String password, int timeout)
 * Description: Stores the wallet second password in cache for timeout in seconds. Only requred for double encrypted wallets.
 * Returns: error
 */
func (s *Session) WalletPassphrase(passphrase string, timeout int) error {
	_, err := s.call("walletpassphrase", []RPC_Data{passphrase, timeout})
	return err
}
