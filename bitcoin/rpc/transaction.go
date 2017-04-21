package rpc

// CreateRawTransaction [{"txid":txid,"vout":n},...] {address:amount,...}
// Create a transaction spending given inputs (array of objects containing
// transaction outputs to spend), sending to given address(es). Returns the
// hex-encoded transaction in a string. Note that the transaction's inputs
// are not signed, and it is not stored in the wallet or transmitted to the
// network.
// Also note that NO transaction validity checks are done; it is easy to
// create invalid transactions or transactions that will not be relayed/mined
// by the network because they contain insufficient fees.
func (s *Session) CreateRawTransaction(slots []Outpoint, targets []Balance, lockTime int) (string, error) {
	ins := make(map[string]float64)
	for _, t := range targets {
		ins[t.Address] = t.Amount
	}
	data := []Data{slots, ins}
	if lockTime != 0 {
		data = append(data, lockTime)
	}
	res, err := s.call("createrawtransaction", data)
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// DecodeRawTransaction <hex string>
// Returns instance with information about a serialized, hex-encoded
// transaction.
func (s *Session) DecodeRawTransaction(raw string) (*RawTransaction, error) {
	res, err := s.call("decoderawtransaction", []Data{raw})
	if err != nil {
		return nil, err
	}
	t := new(RawTransaction)
	if err = res.UnmarshalResult(t); err != nil {
		return nil, err
	}
	return t, nil
}

// EstimateFee estimates the transaction fee per kilobyte that needs to be
// paid for a transaction to be included within a certain number of blocks.
func (s *Session) EstimateFee(waitBlocks int) (float64, error) {
	res, err := s.call("estimatefee", []Data{waitBlocks})
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// EstimatePriority estimates the priority that a transaction needs in order
// to be included within a certain number of blocks as a free high-priority
// transaction.
func (s *Session) EstimatePriority(waitBlocks int) (float64, error) {
	res, err := s.call("estimatepriority", []Data{waitBlocks})
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// FundRawTransaction adds inputs to a transaction until it has enough in
// value to meet its out value. This will not modify existing inputs, and
// will add one change output to the outputs. Note that inputs which were
// signed may need to be resigned after completion since in/outputs have
// been added. The inputs added will not be signed, use signrawtransaction
// for that. All existing inputs must have their previous output transaction
// be in the wallet.
func (s *Session) FundRawTransaction(raw string, opts *Options) (*TransactionInfo, error) {
	res, err := s.call("fundrawtransaction", []Data{raw, opts})
	if err != nil {
		return nil, err
	}
	fi := new(TransactionInfo)
	if err = res.UnmarshalResult(fi); err != nil {
		return nil, err
	}
	return fi, nil
}

// GetRawTransaction returns a serialized, hex-encoded data for
// transaction txid.
func (s *Session) GetRawTransaction(txid string) (string, error) {
	res, err := s.call("getrawtransaction", []Data{txid, 0})
	if err != nil {
		return "", err
	}
	if res.Result == nil {
		return "", nil
	}
	return res.Result.(string), nil
}

// GetRawTransactionObj returns a RawTransaction object for transaction txid.
func (s *Session) GetRawTransactionObj(txid string) (*RawTransaction, error) {
	res, err := s.call("getrawtransaction", []Data{txid, 1})
	if err != nil {
		return nil, err
	}
	if res.Result == nil {
		return nil, nil
	}
	tx := new(RawTransaction)
	if err = res.UnmarshalResult(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// GetTransaction returns an object about the given transaction hash.
func (s *Session) GetTransaction(hash string, watchOnly bool) (*Transaction, error) {
	res, err := s.call("gettransaction", []Data{hash, watchOnly})
	if err != nil {
		return nil, err
	}
	t := new(Transaction)
	if err = res.UnmarshalResult(t); err != nil {
		return nil, err
	}
	return t, nil
}

// GetTxOut returns details about a transaction output. Only unspent
// transaction outputs (UTXOs) are guaranteed to be available.
func (s *Session) GetTxOut(txid string, vout int, unconfirmed bool) (*OutputInfo, error) {
	res, err := s.call("gettxout", []Data{txid, vout, unconfirmed})
	if err != nil {
		return nil, err
	}
	oi := new(OutputInfo)
	if err = res.UnmarshalResult(oi); err != nil {
		return nil, err
	}
	return oi, nil
}

// GetTxOutProof returns a hex-encoded proof that one or more specified
// transactions were included in a block. NOTE: By default this function only
// works when there is an unspent output in the UTXO set for this transaction.
// To make it always work, you need to maintain a transaction index, using the
// -txindex command line option, or specify the block in which the transaction
// is included in manually (by block header hash).
func (s *Session) GetTxOutProof(txids []string, header string) (string, error) {
	data := []Data{txids}
	if len(header) > 0 {
		data = append(data, header)
	}
	res, err := s.call("gettxoutproof", data)
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetTxOutSetInfo returns statistics about the confirmed unspent transaction
// output (UTXO) set. Note that this call may take some time and that it only
// counts outputs from confirmed transactions—it does not count outputs from
// the memory pool.
func (s *Session) GetTxOutSetInfo() (*TxOutSetInfo, error) {
	res, err := s.call("gettxoutsetinfo", nil)
	if err != nil {
		return nil, err
	}
	si := new(TxOutSetInfo)
	if err = res.UnmarshalResult(si); err != nil {
		return nil, err
	}
	return si, nil
}

// ListSinceBlock gets all transactions in blocks since block
// [blockhash] (not inclusive), or all transactions if omitted.
// Max 25 at a time.
func (s *Session) ListSinceBlock(hash string, minConf int) ([]*Transaction, string, error) {
	res, err := s.call("listsinceblock", []Data{hash, minConf})
	if err != nil {
		return nil, "", err
	}
	type txList struct {
		LastBlock    string         `json:"lastblock"`
		Transactions []*Transaction `json:"transactions"`
	}
	list := new(txList)
	if err = res.UnmarshalResult(list); err != nil {
		return nil, "", err
	}
	return list.Transactions, list.LastBlock, nil
}

// ListTransactions returns up to [count] most recent transactions
// skipping the first [from] transactions for account [account].
func (s *Session) ListTransactions(accnt string, count, offset int) ([]*Transaction, error) {
	res, err := s.call("listtransactions", []Data{accnt, count, offset})
	if err != nil {
		return nil, err
	}
	var list []*Transaction
	if err = res.UnmarshalResult(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// ListAllTransactions returns up to [count] most recent transactions
// skipping the first [from] transactions for all accounts.
func (s *Session) ListAllTransactions(count, offset int) ([]*Transaction, error) {
	res, err := s.call("listtransactions", []Data{nil, count, offset})
	if err != nil {
		return nil, err
	}
	var list []*Transaction
	if err = res.UnmarshalResult(&list); err != nil {
		return nil, err
	}
	return list, nil
}

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
	res, err := s.call("signrawtransaction", []Data{raw, ins, keys, mode})
	if err != nil {
		return raw, false, err
	}
	type sigResult struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}
	sr := new(sigResult)
	if err = res.UnmarshalResult(sr); err != nil {
		return raw, false, err
	}
	return sr.Hex, sr.Complete, nil
}
