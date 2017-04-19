package rpc

//=====================================================================
// Data structures as used as input or output of Bitcoin JSON-RPC
// calls. The descriptions of fields is taken from the Bitcoin
// developer website at "https://bitcoin.org/en/developer-reference".
//=====================================================================

// Info contains various information about the node and the network.
type Info struct {
	// Version is this node’s version of Bitcoin Core in its internal
	// integer format. For example, Bitcoin Core 0.9.2 has the integer
	// version number 90200.
	Version int `json:"version"`
	// ProtocolVersion is the protocol version number used by this node.
	// See the protocol versions section for more information.
	ProtocolVersion int `json:"protocolversion"`
	// WalletVersion is the version number of the wallet. Only returned
	// if wallet support is enabled.
	WalletVersion int `json:"walletversion,omitempty"`
	// Balance of the wallet in bitcoins. Only returned if wallet support
	// is enabled.
	Balance float64 `json:"balance,omitempty"`
	// Blocks is the number of blocks in the local best block chain. A new
	// node with only the hardcoded genesis block will return 0.
	Blocks int `json:"blocks"`
	// TimeOffset is the offset of the node’s clock from the computer’s clock
	// (both in UTC) in seconds. The offset may be up to 4200 seconds
	// (70 minutes).
	TimeOffset int `json:"timeoffset"`
	// Connections is the total number of open connections (both outgoing
	// and incoming) between this node and other nodes.
	Connections int `json:"connections"`
	// Proxy is the hostname/IP address and port number of the proxy, if set,
	// or an empty string if unset.
	Proxy string `json:"proxy"`
	// Difficulty of the highest-height block in the local best block chain.
	Difficulty float64 `json:"difficulty"`
	// TestNet is set to true if this node is on testnet; set to false if this
	// node is on mainnet or a regtest.
	TestNet bool `json:"testnet"`
	// KeyPoolOldest marks the date as Unix epoch time when the oldest key in
	// the wallet key pool was created; useful for only scanning blocks
	// created since this date for transactions. Only returned if wallet
	// support is enabled.
	KeyPoolOldest int `json:"keypoololdest,omitempty"`
	// KeyPoolSize is the number of keys in the wallet keypool. Only returned
	// if wallet support is enabled.
	KeyPoolSize int `json:"keypoolsze,omitempty"`
	// PayTxFee is the minimum fee to pay per kilobyte of transaction; may be
	// 0. Only returned if wallet support is enabled.
	PayTxFee float64 `json:"paytxfee,omitempty"`
	// RelayFee is the minimum fee a low-priority transaction must pay in
	// order for this node to accept it into its memory pool.
	RelayFee float64 `json:"relayfee"`
	// UnlockedUntil is the Unix epoch time when the wallet will automatically
	// re-lock. Only displayed if wallet encryption is enabled. Set to 0 if
	// wallet is currently locked.
	UnlockedUntil int `json:"unlocked_until,omitempty"`
	// Errors is a plain-text description of any errors this node has
	// encountered or detected. If there are no errors, an empty string will
	// be returned. This is not related to the JSON-RPC error field.
	Errors string `json:"errors"`
}

// BlockchainInfo returns information about the current state of the
// blockchain as seen by the Bitcoin daemon.
type BlockchainInfo struct {
	// Chain is the name of the block chain. One of 'main' for mainnet,
	// 'test' for testnet, or 'regtest' for regtest.
	Chain string `json:"chain"`
	// Blocks is the number of validated blocks in the local best block
	// chain. For a new node with just the hardcoded genesis block, this
	// will be 0.
	Blocks int `json:"blocks"`
	// Headers is the number of validated headers in the local best headers
	// chain. For a new node with just the hardcoded genesis block, this will
	// be zero. This number may be higher than the number of blocks.
	Headers int `json:"headers"`
	// BestBlockHash is the hash of the header of the highest validated block
	// in the best block chain, encoded as hex in RPC byte order. This is
	// identical to the string returned by the getbestblockhash RPC.
	BestBlockHash string `json:"bestblockhash"`
	// Difficulty of the highest-height block in the best block chain.
	Difficulty float64 `json:"difficulty"`
	// MedianTime is the median time of the 11 blocks before the most recent
	// block on the blockchain. Used for validating transaction locktime
	// under BIP113
	MedianTime int `json:"mediantime"`
	// VerificationProgress is the estimate of what percentage of the block
	// chain transactions have been verified so far, starting at 0.0 and
	// increasing to 1.0 for fully verified. May slightly exceed 1.0 when
	// fully synced to account for transactions in the memory pool which have
	// been verified before being included in a block.
	VerificationProgress float64 `json:"verificationprogress"`
	// ChainWork is the estimated number of block header hashes checked from
	// the genesis block to this block, encoded as big-endian hex.
	ChainWork string `json:"chainwork"`
	// Pruned indicates if the blocks are subject to pruning.
	Pruned bool `json:"prune"`
	// PruneHeight is the lowest-height complete block stored if prunning
	// is activated.
	PruneHeight int `json:"pruneheight,omitempty"`
	// Softforks is an array of objects each describing a current or previous
	// soft fork.
	Softforks []*struct {
		// Id is the name of the softfork.
		ID string `json:"id"`
		// Version is the block version used for the softfork.
		Version int `json:"version"`
		// Enforce describes the progress toward enforcing the softfork rules for
		// new-version blocks. Could be either a string or a ForkProgress
		Enforce interface{} `json:"enforce,omitempty"`
		// Reject describes the progress toward enforcing the softfork rules for
		// new-version blocks.
		Reject *ForkProgress `json:"reject,omitempty"`
	} `json:"softforks"`
	// BIP9Softforks describes the status of BIP9 softforks in progress.
	BIP9Softforks map[string]*struct {
		// Status is set to one of the following reasons:
		// -- 'defined' if voting hasn’t started yet
		// -- 'started' if the voting has started
		// -- 'locked_in' if the voting was successful but the softfort hasn’t
		//    been activated yet
		// -- 'active' if the softfork was activated
		// -- 'failed' if the softfork has not receieved enough votes
		Status string `json:"status"`
		// Bit is the value of bit (0-28) in the block version field used to
		// signal this softfork. Field is only shown when status is started.
		Bit int `json:"bit,omitempty"`
		// StartTime is the Unix epoch time when the softfork voting begins.
		StartTime int `json:"startTime"`
		// Timeout is the Unix epoch time at which the deployment is considered
		// failed if not yet locked in.
		Timeout int `json:"timeout"`
	} `json:"bip9_softforks"`
}

// ForkProgress describes the progress toward enforcing the softfork rules for
// new-version blocks.
type ForkProgress struct {
	// Status indicates if the threshold was reached.
	Status bool `json:"status"`
	// Found is the number of blocks that support the softfork.
	Found int `json:"found,omitempty"`
	// Required is the number of blocks that are required to reach the
	// threshold.
	Required int `json:"required,omitempty"`
	// Window is the maximum size of examined window of recent blocks.
	Window int `json:"window,omitempty"`
}

// Block (element of the Bitcoin blockchain)
type Block struct {
	// Hash of this block’s block header encoded as hex in RPC byte order.
	// This is the same as the hash provided in parameter #1.
	Hash string `json:"hash"`
	// Confirmations is the number of confirmations the transactions in this
	// block have, starting at 1 when this block is at the tip of the best
	// block chain. This score will be -1 if the the block is not part of the
	// best block chain.
	Confirmations int `json:"confirmations"`
	// Size of this block in serialized block format, counted in bytes.
	Size int `json:"size"`
	// StrippedSize is the size of this block in serialized block format
	// excluding witness data, counted in bytes.
	StrippedSize int `json:"strippedsize"`
	// Weight is this block’s weight as defined in BIP141.
	Weight int `json:"weight"`
	// Height of this block on its block chain.
	Height int `json:"height"`
	// Version is this block’s version number. See block version numbers.
	Version int `json:"version"`
	// VersionHex is this block’s version formatted in hexadecimal.
	VersionHex string `json:"versionHex"`
	// MerkleRoot for this block, encoded as hex in RPC byte order.
	MerkleRoot string `json:"merkleroot"`
	// Tx is an array containing the TXIDs of all transactions in this block.
	// The transactions appear in the array in the same order they appear in
	// the serialized block.
	Tx []string `json:"tx"`
	// Time is the value of the time field in the block header, indicating
	// approximately when the block was created.
	Time int `json:"time"`
	// MediaTime is the median block time in Unix epoch time.
	MedianTime int `json:"mediantime"`
	// Nonce which was successful at turning this particular block into one
	// that could be added to the best block chain.
	Nonce int `json:"nonce"`
	// Bits is the value of the nBits field in the block header, indicating
	// the target threshold this block’s header had to pass.
	Bits string `json:"bits"`
	// Difficulty is the estimated amount of work done to find this block
	// relative to the estimated amount of work done to find block 0.
	Difficulty float64 `json:"difficulty"`
	// ChainWork is the estimated number of block header hashes miners had to
	// check from the genesis block to this block, encoded as big-endian hex.
	ChainWork string `json:"chainwork"`
	// PreviousBlockHash is the hash of the header of the previous block,
	// encoded as hex in RPC byte order. Not returned for genesis block.
	PreviousBlockHash string `json:"previousblockhash,omitempty"`
	// NextBlockHash is the hash of the next block on the best block chain,
	// if known, encoded as hex in RPC byte order.
	NextBlockHash string `json:"nextblockhash,omitempty"`
}

// Transaction is a Bitcoin transaction
type Transaction struct {
	// Amount is a positive number of bitcoins if this transaction increased
	// the total wallet balance; a negative number of bitcoins if this
	// transaction decreased the total wallet balance, or 0 if the transaction
	// had no net effect on wallet balance.
	Amount float64 `json:"amount"`
	// Fee for an outgoing transaction; paid by the transaction reported
	// as negative bitcoins
	Fee float64 `json:"fee,omitempty"`
	// Confirmation is the number of confirmations the transaction has
	// received. Will be 0 for unconfirmed and -1 for conflicted.
	Confirmations int `json:"confirmations"`
	// Generated marks a transaction if it is a coinbase. Not returned
	// for regular transactions.
	Generated bool `json:"generated,omitempty"`
	// BlockHash is the hash of the block on the local best block chain which
	// includes this transaction, encoded as hex in RPC byte order. Only
	// returned for confirmed transactions.
	BlockHash string `json:"blockhash,omitempty"`
	// BlockIndex is the index of the transaction in the block that includes
	// it. Only returned for confirmed transactions.
	BlockIndex int `json:"blockindex,omitempty"`
	// BlockTime is the block header time (Unix epoch time) of the block on
	// the local best block chain which includes this transaction. Only
	// returned for confirmed transactions.
	BlockTime int `json:"blocktime,omitempty"`
	// TxID of the transaction, encoded as hex in RPC byte order.
	TxID string `json:"txid,omitempty"`
	// WalletConflicts is an array containing the TXIDs of other transactions
	// that spend the same inputs (UTXOs) as this transaction. Array may be
	// empty.
	WalletConflicts []string `json:"walletconflicts,omitempty"`
	// Time is a Unix epoch time when the transaction was added to the wallet.
	Time int `json:"time"`
	// TimeReceived is a Unix epoch time when the transaction was detected by
	// the local node, or the time of the block on the local best block chain
	// that included the transaction.
	TimeReceived int `json:"timereceived,omitempty"`
	// BIP125Replaceable indicates if a transaction is replaceable under
	// BIP 125:
	// -- 'yes' is replaceable
	// -- 'no' not replaceable
	// -- 'unknown' for unconfirmed transactions not in the mempool
	BIP125Replaceable string `json:"bip125-replaceable"`
	// Comment is added to a transaction originating with this wallet. Only
	// returned if a comment was added.
	Comment string `json:"comment,omitempty"`
	// To is added as a comment to a transaction originating with this wallet.
	// Only returned if a comment-to was added
	To      string `json:"to,omitempty"`
	Details []*struct {
		// InvolvesWatchOnly is set to true if the input or output involves
		// a watch-only address. Otherwise not returned.
		InvolvesWatchOnly bool `json:"involvesWatchonly,omitempty"`
		// Account which the payment was credited to or debited from. May be
		// an empty string ("") for the default account.
		Account string `json:"account"`
		// Address on output is the address paid (may be someone else’s
		// address not belonging to this wallet). If an input, the address
		// paid in the previous output. May be empty if the address is
		// unknown, such as when paying to a non-standard pubkey script.
		Address string `json:"address,omitempty"`
		// Category is set to one of the following values:
		// -- 'send' if sending payment
		// -- 'receive' if this wallet received payment in a regular transaction
		// -- 'generate' if a matured and spendable coinbase
		// -- 'immature' if a coinbase that is not spendable yet
		// -- 'orphan' if a coinbase from a block that’s not in the local best block chain
		Category string `json:"category"`
		// Amount is a negative bitcoin amount if sending payment; a positive
		// bitcoin amount if receiving payment (including coinbases).
		Amount float64 `json:"amount"`
		// Vout on an output is the output index (vout) for this output in
		// this transaction. For an input, the output index for the output
		// being spent in its transaction. Because inputs list the output
		// indexes from previous transactions, more than one entry in the
		// details array may have the same output index.
		Vout int `json:"vout"`
		// Fee is paid as a negative bitcoins value on a sending payment. May
		// be 0. Not returned if receiving payment.
		Fee float64 `json:"fee,omitempty"`
		// Abandoned indicates if a transaction is was abandoned:
		// -- 'true' if it was abandoned (inputs are respendable)
		// -- 'false' if it was not abandoned
		// Only returned by send category payments
		Abandoned bool `json:"abandoned,omitempty"`
		// Hex represents the transaction in serialized transaction format.
		Hex string `json:"hex"`
	} `json:"details"`
}

// RawTransaction is a Bitcoin transaction in raw format
type RawTransaction struct {
	// Hex is the serialized, hex-encoded data for the provided txid.
	Hex string `json:"hex"`
	// TxID of the transaction encoded as hex in RPC byte order.
	TxID string `json:"txid"`
	// Hash is the transaction hash. Differs from txid for witness
	// transactions.
	Hash string `json:"hash"`
	// Size is the byte count of the serialized transaction.
	Size int `json:"size"`
	// VSize is the virtual transaction size. Differs from size for
	// witness transactions.
	VSize int `json:"vsize"`
	// Version is the transaction format version number.
	Version int `json:"version"`
	// LockTime is the transaction’s locktime: either a Unix epoch date or
	// block height; see the Locktime parsing rules.
	LockTime int `json:"locktime"`
	// Vin is an array of objects with each object being an input vector (vin)
	// for this transaction. Input objects will have the same order within the
	// array as they have in the transaction, so the first input listed will
	// be input 0.
	Vin []*struct {
		// TxID of the outpoint being spent, encoded as hex in RPC byte order.
		// Not present if this is a coinbase transaction.
		TxID string `json:"txid"`
		// Vout is the output index number (vout) of the outpoint being spent.
		// The first output in a transaction has an index of 0. Not present
		// if this is a coinbase transaction.
		Vout int `json:"vout"`
		// ScriptSig is an object describing the signature script of this
		// input. Not present if this is a coinbase transaction.
		ScriptSig *Script `json:"scriptSig"`
		// Coinbase (similar to the hex field of a scriptSig) encoded as hex.
		// Only present if this is a coinbase transaction.
		Coinbase string `json:"coinbase,omitempty"`
		// Sequence is the input sequence number.
		Sequence int `json:"sequence"`
		// TxInWitness is the hex-encoded witness data. Only for segregated
		// witness transactions
		TxInWitness string `json:"txinwitness,omitempty"`
	} `json:"vin"`
	// Vout is an object describing one of this transaction’s outputs.
	Vout []*struct {
		// Value is the number of bitcoins paid to this output. May be 0.
		Value float64 `json:"value"`
		// N is the output index number of this output within this transaction.
		N int `json:"n"`
		// ScriptPubKey is an object describing the pubkey script.
		ScriptPubKey *ScriptPubKey `json:"scriptPubKey"`
	} `json:"vout"`
}

// Script is a generic script base object either used for signature scripts
// (input) or pubkey scripts (output).
type Script struct {
	// Asm is the script in decoded form with non-data-pushing
	// opcodes listed.
	Asm string `json:"asm"`
	// Hex is the script encoded as hex.
	Hex string `json:"hex"`
}

// ScriptPubKey is a public key scripts used in outputs.
type ScriptPubKey struct {
	Script
	// ReqSigs is the number of signatures required; this is always 1
	// for P2PK, P2PKH, and P2SH (including P2SH multisig because the
	// redeem script is not available in the pubkey script). It may be
	// greater than 1 for bare multisig. This value will not be
	// returned for nulldata or nonstandard script types (see the type
	// key below)
	ReqSigs int `json:"reqSigs,emitempty"`
	// Type of script. This will be one of the following:
	// -- 'pubkey' for a P2PK script
	// -- 'pubkeyhash' for a P2PKH script
	// -- 'scripthash' for a P2SH script
	// -- 'multisig' for a bare multisig script
	// -- 'nulldata' for nulldata scripts
	// -- 'nonstandard' for unknown scripts
	Type string `json:"type"`
	// Addresses
	Addresses interface{} `json:"addresses"`
}

// DecodedScript is an object describing a decoded script.
type DecodedScript struct {
	// Asm is the redeem script in decoded form with non-data-pushing opcodes
	// listed. May be empty.
	Asm string `json:"asm"`
	// Type of script. This will be one of the following:
	// -- 'pubkey' for a P2PK script inside P2SH
	// -- 'pubkeyhash' for a P2PKH script inside P2SH
	// -- 'multisig' for a multisig script inside P2SH
	// -- 'nonstandard' for unknown scripts
	Type string `json:"type,omitempty"`
	// ReqSigs is the number of signatures required; this is always 1 for P2PK
	// or P2PKH within P2SH. It may be greater than 1 for P2SH multisig. This
	// value will not be returned for nonstandard script types (see the type
	// key above).
	ReqSigs int `json:"reqsigs,omitempty"`
	// Addresses is a P2PKH addresses used in this script, or the computed
	// P2PKH addresses of any pubkeys in this script. This array will not be
	// returned for nonstandard script types.
	Addresses []string `json:"addresses,omitempty"`
	// P2SH address of this redeem script.
	P2SH string `json:"p2sh"`
}

// Outpoint of a transaction
type Outpoint struct {
	// TxID of the outpoint encoded as hex in RPC byte order.
	TxID string `json:"txid"`
	// Vout is the output index number (vout) of the outpoint; the first
	// output in a transaction is index 0.
	Vout int `json:"vout"`
}

// Output is being spent.
type Output struct {
	Outpoint
	// ScriptPubKey is the output’s pubkey script encoded as hex.
	ScriptPubKey string `json:"scriptPubKey"`
	// RedeemScript is the corresponding redeem script, if the pubkey script
	// was a script hash.
	ReddemScript string `json:"redeemScript,omitempty"`
}

// Balance of Bitcoin address.
type Balance struct {
	// Address refers to a Bitcoin address.
	Address string `json:"address"`
	// Amount is the Bitcoin value associated with the address.
	Amount float64 `json:"amount"`
}

// AccountInfo list informations about an acount.
type AccountInfo struct {
	// InvolvesWatchOnly is set to true if the balance of this account
	// includes a watch-only address which has received a spendable payment
	// (that is, a payment with at least the specified number of confirmations
	// and which is not an immature coinbase). Otherwise not returned.
	InvolvesWatchonly bool `json:"involvesWatchonly,omitempty"`
	// Account is the name of the account.
	Account string `json:"account"`
	// Amount is the total amount received by this account in bitcoins.
	Amount float64 `json:"amount"`
	// Confirmations is the number of confirmations received by the last
	// transaction received by this account. May be 0.
	Confirmations int `json:"confirmations"`
}

// AddressInfo contains information for a Bitcoin address.
type AddressInfo struct {
	AccountInfo
	// Address being described encoded in base58check.
	Address string `json:"address"`
	// Label is the name of the account the address belongs to. May be the
	// default account, an empty string ("").
	Label string `json:"label"`
	// TxIDs is an array of TXIDs belonging to transactions that pay the
	// address.
	TxIDs []string `json:"txids"`
}

// Validity check on address
type Validity struct {
	// IsValid is set to true if the address is a valid P2PKH or P2SH address;
	// set to false otherwise.
	IsValid bool `json:"isvalid"`
	// Address is the bitcoin address given as parameter.
	Address string `json:"address,omitempty"`
	// ScriptPubKey is the hex encoded scriptPubKey generated by the address.
	ScriptPubKey string `json:"scriptPubKey,omitempty"`
	// IsMine is set to true if the address belongs to the wallet; set to
	// false if it does not. Only returned if wallet support enabled.
	IsMine bool `json:"ismine,omitempty"`
	// IsWatchOnly is set to true if the address is watch-only. Otherwise set
	// to false. Only returned if address is in the wallet.
	IsWatchOnly bool `json:"iswatchonly,omitempty"`
	// IsScript is set to true if a P2SH address; otherwise set to false. Only
	// returned if the address is in the wallet.
	IsScript bool `json:"isscript,omitempty"`
	// Script is only returned for P2SH addresses belonging to this wallet.
	// This is the type of script:
	// -- 'pubkey' for a P2PK script inside P2SH
	// -- 'pubkeyhash' for a P2PKH script inside P2SH
	// -- 'multisig' for a multisig script inside P2SH
	// -- 'nonstandard' for unknown scripts
	Script string `json:"script,omitempty"`
	// Hex is only returned for P2SH addresses belonging to this wallet. This
	// is the redeem script encoded as hex.
	Hex string `json:"hex,omitempty"`
	// Addresses
	Addresses []*struct {
		// SigRequired is only returned for multisig P2SH addresses belonging
		// to the wallet. The number of signatures required by this script.
		SigRequired int `json:"sigrequired,omitempty"`
		// PubKey corresponding to this address. Only returned if the address
		// is a P2PKH address in the wallet.
		PubKey string `json:"pubkey,omitempty"`
		// IsCompressed is set to true if a compressed public key or set to
		// false if an uncompressed public key. Only returned if the address
		// is a P2PKH address in the wallet.
		IsCompressed bool `json:"iscompressed"`
		// Account this address belong to. May be an empty string for the
		// default account. Only returned if the address belongs to the wallet.
		Account string `json:"account,omitempty"`
		// HDKeyPath is the HD keypath if the key is HD and available.
		HDKeyPath string `json:"hdkeypath,omitempty"`
		// HDMasterKeyID is the Hash160 of the HD master public key.
		HDMasterKeyID string `json:"hdmasterkeyid,omitempty"`
	} `json:"addresses,omitempty"`
}

// Options holds additional data for FundRawTransaction
type Options struct {
	// ChangeAddress is the bitcoin address to receive the change. If not set,
	// the address is chosen from address pool.
	ChangeAddress string `json:"changeAddress,omitempty"`
	// ChangePosition is the index of the change output. If not set, the
	// change position is randomly chosen.
	ChangePosition int `json:"changePosition,omitempty"`
	// IncludeWatching decides if inputs from watch-only addresses are also
	// considered. The default is false.
	IncludeWatching bool `json:"includeWatching,omitempty"`
	// LockUnspent flags that selected outputs are locked after running the
	// rpc call. The default is false.
	LockUnspents bool `json:"lockUnspents,omitempty"`
	// FeeRate you are willing to pay (BTC per KB). If not set, the wallet
	// determines the fee.
	FeeRate float64 `json:"feeRate,omitempty"`
}

// MultiSigAddr is a multi-signature address object
type MultiSigAddr struct {
	// Address is the P2SH address for this multisig redeem script.
	Address string `json:"address"`
	// RedeemScript is the multisig redeem script encoded as hex.
	RedeemScript string `json:"redeemScript"`
}

// TransactionInfo contains information about a created transaction.
type TransactionInfo struct {
	// Hex is the resulting unsigned raw transaction in serialized transaction
	// format encoded as hex.
	Hex string `json:"hex"`
	// Fee in BTC the resulting transaction pays.
	Fee float64 `json:"fee"`
	// ChangePos is the position of the added change output, or -1 if no
	// change output was added
	ChangePos int `json:"changepos"`
}

// NodeInfo holds information about added nodes.
type NodeInfo struct {
	// AddedNode is an added node in the same <IP address>:<port> format as
	// used in the addnode RPC. This element is present for any added node
	// whether or not the Details parameter was set to true.
	AddedNode string `json:"addednode"`
	// Connected will be set to true if the node is currently connected and
	// false if it is not. This is only set if the Details parameter was set
	// to true.
	Connected bool `json:"connected,omitempty"`
	// Addresses will be an array of addresses belonging to the added node if
	// the Details parameter was set to true.
	Addresses []*struct {
		// Address is an IP address and port number of the node. If the node
		// was added using a DNS address, this will be the resolved IP
		// address.
		Address string `json:"address"`
		// Connected flags if the local node is connected to this addnode
		// using this IP address. Valid values are:
		// -- 'false' for not connected
		// -- 'inbound' if the addnode connected to us
		// -- 'outbound' if we connected to the addnode
		Connected string `json:"connected"`
	} `json:"addresses"`
}
