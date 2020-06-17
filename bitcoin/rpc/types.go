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
	Balance float64 `json:"balance"`
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
	KeyPoolSize int `json:"keypoolsize,omitempty"`
	// PayTxFee is the minimum fee to pay per kilobyte of transaction; may be
	// 0. Only returned if wallet support is enabled.
	PayTxFee float64 `json:"paytxfee"`
	// RelayFee is the minimum fee a low-priority transaction must pay in
	// order for this node to accept it into its memory pool.
	RelayFee float64 `json:"relayfee"`
	// UnlockedUntil is the Unix epoch time when the wallet will automatically
	// re-lock. Only displayed if wallet encryption is enabled. Set to 0 if
	// wallet is currently locked.
	UnlockedUntil *int `json:"unlocked_until,omitempty"`
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
	Pruned bool `json:"pruned"`
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
		// Since is the block number when the softfork took place.
		Since int `json:"since"`
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

// ChainTip is an object describing a particular chain tip. The first object
// in an array returned by 'GetChainTips()' will always describe the active
// chain (the local best block chain).
type ChainTip struct {
	// Height of the highest block in the chain. A new node with only the
	// genesis block will have a single tip with height of 0.
	Height int `json:"height"`
	// Hash of the highest block in the chain, encoded as hex in RPC byte order
	Hash string `json:"hash"`
	// BranchLen is the number of blocks that are on this chain but not on the
	// main chain. For the local best block chain, this will be 0; for all
	// other chains, it will be at least 1.
	BranchLen int `json:"branchlen"`
	// Status  of this chain. Valid values are:
	// -- 'active' for the local best block chain
	// -- 'invalid' for a chain that contains one or more invalid blocks
	// -- 'headers-only' for a chain with valid headers whose corresponding
	//    blocks both haven’t been validated and aren’t stored locally
	// -- 'valid-headers' for a chain with valid headers whose corresponding
	//    blocks are stored locally, but which haven’t been fully validated
	// -- 'valid-fork' for a chain which is fully validated but which isn’t
	//    part of the local best block chain (it was probably the local best
	//    block chain at some point)
	// -- 'unknown' for a chain whose reason for not being the active chain
	//    is unknown
	Status string `json:"status"`
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

// BlockTemplate is a data structure needed by mining applications.
type BlockTemplate struct {
	Capabilities      []string       `json:"capabilities"`
	Version           int            `json:"version"`
	Rules             []string       `json:"rules"`
	VbAvailable       map[string]int `json:"vbavailable"`
	VbRequired        int            `json:"vbrequired"`
	PreviousBlockHash string         `json:"previousblockhash"`
	Transactions      []*struct {
		Data    string `json:"data"`
		TxID    string `json:"txid"`
		Hash    string `json:"hash"`
		Depends []int  `json:"depends"`
		Fee     int    `json:"fee"`
		SigOps  int    `json:"sigops"`
		Weight  int    `json:"weight"`
	} `json:"transactions"`
	CoinbaseAux   map[string]string `json:"coinbaseaux"`
	CoinbaseValue int               `json:"coinbasevalue"`
	LongPollID    string            `json:"longpollid"`
	Target        string            `json:"target"`
	MinTime       int               `json:"mintime"`
	Mutable       []string          `json:"mutable"`
	NonceRange    string            `json:"noncerange"`
	SigOpLimit    int               `json:"sigoplimit"`
	SizeLimit     int               `json:"sizelimit"`
	CurTime       int               `json:"curtime"`
	Bits          string            `json:"bits"`
	Height        int               `json:"height"`
	WeightLimit   int               `json:"weightlimit"`
}

// BlockTemplateParameter defines parameters for the GetBlockTemplate call.
type BlockTemplateParameter struct {
	Capabilities []string `json:"capabilities"`
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
	// Abandoned indicates if a transaction is was abandoned:
	// -- 'true' if it was abandoned (inputs are respendable)
	// -- 'false' if it was not abandoned
	// Only returned by send category payments
	Abandoned *bool `json:"abandoned,omitempty"`
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
	WalletConflicts []string `json:"walletconflicts"`
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
	// Label is the label for the transaction.
	Label *string `json:"label,omitempty"`
	// Comment is added to a transaction originating with this wallet. Only
	// returned if a comment was added.
	Comment string `json:"comment,omitempty"`
	// To is added as a comment to a transaction originating with this wallet.
	// Only returned if a comment-to was added
	To string `json:"to,omitempty"`
	// Category is set to one of the following values:
	// -- 'send' if sending payment
	// -- 'receive' if this wallet received payment in a regular transaction
	// -- 'generate' if a matured and spendable coinbase
	// -- 'immature' if a coinbase that is not spendable yet
	// -- 'orphan' if a coinbase from a block that’s not in the local best block chain
	Category string `json:"category,omitempty"`
	// Vout on an output is the output index (vout) for this output in
	// this transaction. For an input, the output index for the output
	// being spent in its transaction. Because inputs list the output
	// indexes from previous transactions, more than one entry in the
	// details array may have the same output index.
	Vout *int `json:"vout,omitempty"`
	// Account which the payment was credited to or debited from. May be
	// an empty string ("") for the default account.
	Account *string `json:"account,omitempty"`
	// Address on output is the address paid (may be someone else’s
	// address not belonging to this wallet). If an input, the address
	// paid in the previous output. May be empty if the address is
	// unknown, such as when paying to a non-standard pubkey script.
	Address string `json:"address,omitempty"`
	// Hex represents the transaction in serialized transaction format.
	Hex *string `json:"hex,omitempty"`
	// Details is an array of detail information
	Details []*struct {
		// InvolvesWatchOnly is set to true if the input or output involves
		// a watch-only address. Otherwise not returned.
		InvolvesWatchOnly bool `json:"involvesWatchonly,omitempty"`
		// Label is the label for the transaction.
		Label string `json:"label"`
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
		Abandoned *bool `json:"abandoned,omitempty"`
		// Hex represents the transaction in serialized transaction format.
		Hex *string `json:"hex,omitempty"`
	} `json:"details,omitempty"`
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

// MemPoolTransaction is an object describing a transaction in the memory pool.
type MemPoolTransaction struct {
	// Size of the serialized transaction in bytes.
	Size int `json:"size"`
	// Fee paid by the transaction in decimal bitcoins.
	Fee float64 `json:"fee"`
	// ModifiedFee with fee deltas used for mining priority in decimal bitcoins
	ModifiedFee float64 `json:"modifiedfee"`
	// Time the transaction entered the memory pool, Unix epoch time format.
	Time int `json:"time"`
	// Height is the block height when the transaction entered the memory pool.
	Height int `json:"height"`
	// StartingPriority is the priority of the transaction when it first
	// entered the memory pool.
	StartingPriority int `json:"startingpriority,omitempty"`
	// CurrentPriority is the current priority of the transaction.
	CurrentPriority int `json:"currentpriority,omitempty"`
	// DescendantCount is the number of in-mempool descendant transactions
	// (including this one).
	DescendantCount int `json:"descendantcount"`
	// DescendantSize is the size of in-mempool descendants (including this one)
	DescendantSize int `json:"descendantsize"`
	// DescendantFees is the modified fees (see modifiedfee above) of in-mempool
	// descendants (including this one).
	DescendantFees float64 `json:"descendantfees"`
	// AncestorCount is the number of in-mempool ancestor transactions
	// (including this one).
	AncestorCount int `json:"ancestorcount"`
	// AncestorSize is the size of in-mempool ancestors (including this one)
	AncestorSize int `json:"ancestorsize"`
	// AncestorFees is the modified fees (see modifiedfee above) of in-mempool
	// ancestors (including this one).
	AncestorFees float64 `json:"ancestorfees"`
	// Depends is an array holding TXIDs of unconfirmed transactions this
	// transaction depends upon (parent transactions). Those transactions must
	// be part of a block before this transaction can be added to a block,
	// although all transactions may be included in the same block. The array
	// may be empty.
	Depends []string `json:"depends"`
}

// RawTransaction is a Bitcoin transaction in raw format
type RawTransaction struct {
	// Hex is the serialized, hex-encoded data for the provided txid.
	Hex *string `json:"hex,omitempty"`
	// TxID of the transaction encoded as hex in RPC byte order.
	TxID string `json:"txid"`
	// Confirmation is the number of confirmations the transaction has
	// received. Will be 0 for unconfirmed and -1 for conflicted.
	Confirmations int `json:"confirmations,omitempty"`
	// Hash is the transaction hash. Differs from txid for witness
	// transactions.
	Hash string `json:"hash"`
	// BlockHash is the hash of the block on the local best block chain which
	// includes this transaction, encoded as hex in RPC byte order. Only
	// returned for confirmed transactions.
	BlockHash string `json:"blockhash,omitempty"`
	// BlockTime is the block header time (Unix epoch time) of the block on
	// the local best block chain which includes this transaction. Only
	// returned for confirmed transactions.
	BlockTime int `json:"blocktime,omitempty"`
	// Size is the byte count of the serialized transaction.
	Size int `json:"size"`
	// VSize is the virtual transaction size. Differs from size for
	// witness transactions.
	VSize int `json:"vsize"`
	// Version is the transaction format version number.
	Version int `json:"version"`
	// Time
	Time int `json:"time,omitempty"`
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
	Addresses []string `json:"addresses"`
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
	// returned for non-standard script types.
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

// Unspent describes an unspent output.
type Unspent struct {
	Output
	// Solvable is set to true if the wallet knows how to spend this output.
	// Set to false if the wallet does not know how to spend the output. It
	// is ignored if the private keys are available.
	Solvable bool `json:"solvable"`
	// Spendable is set to true if the private key or keys needed to spend
	// this output are part of the wallet. Set to false if not (such as for
	// watch-only addresses).
	Spendable bool `json:"spendable"`
	// Address is the P2PKH or P2SH address the output paid. Only returned for
	// P2PKH or P2SH output scripts.
	Address string `json:"address"`
	// Account is set if the address returned belongs to an account.
	Account *string `json:"account,omitempty"`
	// Amount is paid to the output in bitcoins.
	Amount float64 `json:"amount"`
	// Confirmations is the number of confirmations received for the
	// transaction containing this output.
	Confirmations int `json:"confirmations"`
	// Safe
	Safe bool `json:"safe"`
}

// OutputInfo contains info about the output of a transaction.
type OutputInfo struct {
	// BestBlock is the hash of the header of the block on the local best
	// block chain which includes this transaction. The hash will encoded as
	// hex in RPC byte order. If the transaction is not part of a block, the
	// string will be empty.
	BestBlock string `json:"bestblock"`
	// Confirmations is the number of confirmations received for the
	// transaction containing this output or 0 if the transaction hasn’t been
	// confirmed yet.
	Confirmations int `json:"confirmations"`
	// Value is the amount of bitcoins spent to this output. May be 0.
	Value float64 `json:"value"`
	// ScriptPubKey is an object with information about the pubkey script.
	// This may be null if there was no pubkey script.
	ScriptPubKey *ScriptPubKey `json:"scriptPubKey"`
	// Version is the transaction version number of the transaction containing
	// the pubkey script.
	Version int `json:"version"`
	// Coinbase is set to true if the transaction output belonged to a coinbase
	// transaction; set to false for all other transactions. Coinbase
	// transactions need to have 101 confirmations before their outputs can be
	// spent.
	Coinbase bool `json:"coinbase"`
}

// TxOutSetInfo contains information about the UTXO set.
type TxOutSetInfo struct {
	// Height is the height of the local best block chain. A new node with
	// only the hardcoded genesis block will have a height of 0.
	Height int `json:"height"`
	// BestBlock is the hash of the header of the highest block on the
	// local best block chain, encoded as hex in RPC byte order.
	BestBlock string `json:"bestblock"`
	// Transactions is the number of transactions with unspent outputs.
	Transactions int `json:"transactions"`
	// TxOuts is the number of unspent transaction outputs.
	TxOuts int `json:"txouts"`
	// BytesSerialized is the size of the serialized UTXO set in bytes; not
	// counting overhead, this is the size of the chainstate directory in the
	// Bitcoin Core configuration directory.
	BytesSerialized int `json:"bytes_serialized"`
	// HashSerialized is a SHA256 hash of the serialized UTXO set; useful for
	// comparing two nodes to see if they have the same set (they should, if
	// they always used the same serialization format and currently have the
	// same best block). The hash is encoded as hex in RPC byte order.
	HashSerialized string `json:"hash_serialized"`
	// TotalAmount is the total number of bitcoins in the UTXO set.
	TotalAmount float64 `json:"total_amount"`
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

// AddressGroup is a group of addresses that may have had their common
// ownership made public by common use as inputs in the same transaction or
// from being used as change from a previous transaction.
type AddressGroup []*AddressDetail

// AddressDetail is information about an address in an address group.
type AddressDetail []interface{}

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
	IsWatchOnly *bool `json:"iswatchonly,omitempty"`
	// IsScript is set to true if a P2SH address; otherwise set to false. Only
	// returned if the address is in the wallet.
	IsScript *bool `json:"isscript,omitempty"`
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
	// PubKey corresponding to this address. Only returned if the address
	// is a P2PKH address in the wallet.
	PubKey string `json:"pubkey,omitempty"`
	// IsCompressed is set to true if a compressed public key or set to
	// false if an uncompressed public key. Only returned if the address
	// is a P2PKH address in the wallet.
	IsCompressed bool `json:"iscompressed,omitempty"`
	// TimeStamp
	TimeStamp int `json:"timestamp,omitempty"`
	// Account this address belong to. May be an empty string for the
	// default account. Only returned if the address belongs to the wallet.
	Account string `json:"account,omitempty"`
	// HDKeyPath is the HD keypath if the key is HD and available.
	HDKeyPath string `json:"hdkeypath,omitempty"`
	// HDMasterKeyID is the Hash160 of the HD master public key.
	HDMasterKeyID string `json:"hdmasterkeyid,omitempty"`
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

// WalletInfo describes the wallet.
type WalletInfo struct {
	// WalletVersion is the version number of the wallet.
	WalletVersion int `json:"walletversion"`
	// UnlockedUntil is only returned if the wallet was encrypted with the
	// encryptwallet RPC. A Unix epoch date when the wallet will be locked, or
	// 0 if the wallet is currently locked.
	UnlockedUntil int `json:"unlocked_until,omitempty"`
	// Balance is the balance of the wallet. The same as returned by the
	// getbalance RPC with default parameters.
	Balance float64 `json:"balance"`
	// UnconfirmedBalance is the balance of the unconfirmed transactions in
	// the wallet. The same as returned by the getbalance RPC with default
	// parameters.
	UnconfirmedBalance float64 `json:"unconfirmed_balance"`
	// ImmatureBalance
	ImmatureBalance float64 `json:"immature_balance"`
	// PayTxFee
	PayTxFee float64 `json:"paytxfee"`
	// TxCount is the total number of transactions in the wallet (both spends
	// and receives).
	TxCount int `json:"txcount"`
	// KeypoolOldest is the date as Unix epoch time when the oldest key in the
	// wallet key pool was created; useful for only scanning blocks created
	// since this date for transactions.
	KeypoolOldest int `json:"keypoololdest"`
	// KeypoolSize is the number of keys in the wallet keypool.
	KeypoolSize int `json:"keypoolsize"`
	// KeypoolSizeHDInternal
	KeypoolSizeHDInternal int `json:"keypoolsize_hd_internal"`
	// HDMasterKeyID
	HDMasterKeyID string `json:"hdmasterkeyid"`
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

// BannedNode is an entry in the ban list.
type BannedNode struct {
	// Address is the IP/Subnet of the entry.
	Address string `json:"address"`
	// BannedUntil is the Unix epoch time when the entry was added to the ban
	// list.
	BannedUntil int `json:"banned_until"`
	// BanCreated is the Unix epoch time until the IP/Subnet is banned.
	BanCreated int `json:"ban_created"`
	// BanReason is set to one of the following reasons:
	// -- 'node misbehaving' if the node was banned by the client because of
	//    DoS violations
	// -- 'manually added' if the node was manually banned by the user
	BanReason string `json:"ban_reason"`
}

// MemPoolInfo is an object containing information about the memory pool.
type MemPoolInfo struct {
	// Size is the number of transactions currently in the memory pool.
	Size int `json:"size"`
	// Bytes is the total number of bytes in the transactions in the
	// memory pool.
	Bytes int `json:"bytes"`
	// Usage is total memory usage for the mempool in bytes.
	Usage int `json:"usage"`
	// MaxMemPool is the maximum memory usage for the mempool in bytes.
	MaxMemPool int `json:"maxmempool"`
	// MemPoolMinFee is the lowest fee per kilobyte paid by any transaction
	// in the memory pool.
	MemPoolMinFee float64 `json:"mempoolminfee"`
}

// MiningInfo contains various mining-related information.
type MiningInfo struct {
	// Blocks is the height of the highest block on the local best block chain.
	Blocks int `json:"blocks"`
	// CurrentBlockSize is the size in bytes of the last block built by this
	// node for header hash checking if generation was enabled since the last
	// time this node was restarted. Otherwise, the value is 0.
	CurrentBlockSize int `json:"currentblocksize"`
	// CurrentBlockWeight
	CurrentBlockWeight int `json:"currentblockweight"`
	// CurrentBlockTx  is the number of transactions in the last block built
	// by this node for header hash checking if generation was enabled since
	// the last time this node was restarted. Otherwise, this is the value 0.
	CurrentBlockTx int `json:"currentblocktx"`
	// Difficulty of the highest-height block in the local best block chain if
	// generation was enabled since the last time this node was restarted.
	// Otherwise, this is the value 0.
	Difficulty float64 `json:"difficulty"`
	// Errors is a a plain-text description of any errors this node has
	// encountered or detected. If there are no errors, an empty string will
	// be returned.
	Errors string `json:"errors"`
	// GenProcLimit is the limit on the number of processors to use for
	// generation. If generation was enabled since the last time this node was
	// restarted, this is the number used in the second parameter of the
	// setgenerate RPC (or the default). Otherwise, it is -1.
	GenProcLimit int `json:"genproclimit,omitempty"`
	// NetworkHashPS is an estimate of the number of hashes per second the
	// network is generating to maintain the current difficulty. See the
	// getnetworkhashps RPC for configurable access to this data.
	NetworkHashPS float64 `json:"networkhashps"`
	// PooledTx is the number of transactions in the memory pool.
	PooledTx int `json:"pooledtx"`
	// Testnet is set to true if this node is running on testnet. Set to false
	// if this node is on mainnet or a regtest.
	Testnet bool `json:"testnet,omitempty"`
	// Chain is set to 'main' for mainnet, 'test' for testnet, and 'regtest'
	// for regtest.
	Chain string `json:"chain"`
	// Generate is set to true if generation is currently enabled; set to
	// false if generation is currently disabled. Only returned if the node
	// has wallet support enabled.
	Generate bool `json:"generate,omitempty"`
	// HashesPerSec is the approximate number of hashes per second this node
	// is generating across all CPUs, if generation is enabled. Otherwise 0.
	// Only returned if the node has wallet support enabled
	HashesPerSec float64 `json:"hashespersec,omitempty"`
}

// PeerInfo describes a particular connected node.
type PeerInfo struct {
	// ID is the node’s index number in the local node address database.
	ID int `json:"id"`
	// Addr is the IP address and port number used for the connection to the
	// remote node.
	Addr string `json:"addr"`
	// AddrLocal is our IP address and port number according to the remote
	// node. May be incorrect due to error or lying. Most SPV nodes set this
	// to 127.0.0.1:8333
	AddrLocal string `json:"addrlocal"`
	// Services as advertised by the remote node in its version message.
	Services string `json:"services"`
	// LastSend is the Unix epoch time when we last successfully sent data to
	// the TCP socket for this node.
	LastSend int `json:"lastsend"`
	// LastRecv is the Unix epoch time when we last received data from this
	// node.
	LastRecv int `json:"lastrecv"`
	// BytesSent is the total number of bytes we’ve sent to this node.
	BytesSent int `json:"bytessent"`
	// BytesRecv is the total number of bytes we’ve received from this node.
	BytesRecv int `json:"bytesrecv"`
	// ConnTime is the Unix epoch time when we connected to this node.
	ConnTime int `json:"conntime"`
	// TimeOffset is the time offset in seconds.
	TimeOffset int `json:"timeoffset"`
	// PingTime is the number of seconds this node took to respond to our
	// last P2P ping message.
	PingTime float64 `json:"pingtime,omitempty"`
	// MinPing is the minimum observed ping time (if any at all).
	MinPing float64 `json:"minping,omitempty"`
	// PingWait is the number of seconds we’ve been waiting for this node to
	// respond to a P2P ping message. Only shown if there’s an outstanding
	// ping message.
	PingWait float64 `json:"pingwait,omitempty"`
	// Version is the protocol version number used by this node. See the
	// protocol versions section for more information.
	Version int `json:"version"`
	// Subver is the user agent this node sends in its version message.
	// This string will have been sanitized to prevent corrupting the JSON
	// results. May be an empty string.
	SubVer string `json:"subver"`
	// Inbound is set to true if this node connected to us; set to false if
	// we connected to this node.
	Inbound bool `json:"inbound"`
	// AddNode
	AddNode bool `json:"addnode"`
	// RelayTxes is set true if this node relays transactions.
	RelayTxes bool `json:"relaytxes"`
	// StartingHeight is the height of the remote node’s block chain when it
	// connected to us as reported in its version message.
	StartingHeight int `json:"startingheight"`
	// BanScore is the ban score we’ve assigned the node based on any
	// misbehavior it’s made. By default, Bitcoin Core disconnects when the
	// ban score reaches 100.
	BanScore float64 `json:"banscore"`
	// SyncedHeaders is the highest-height header we have in common with this
	// node based the last P2P headers message it sent us. If a headers message
	// has not been received, this will be set to -1.
	SyncedHeaders int `json:"synced_headers"`
	// SyncedBlocks is the highest-height block we have in common with this
	// node based on P2P inv messages this node sent us. If no block inv
	// messages have been received from this node, this will be set to -1.
	SyncedBlocks int `json:"synced_blocks"`
	// Inflight is an array of blocks which have been requested from this
	// peer. May be empty.
	Inflight []int `json:"inflight"`
	// Whitelisted is set to true if the remote peer has been whitelisted;
	// otherwise, set to false. Whitelisted peers will not be banned if their
	// ban score exceeds the maximum (100 by default). By default, peers
	// connecting from localhost are whitelisted.
	Whitelisted bool `json:"whitelisted"`
	// BytesSentPerMessage is the total sent bytes aggregated by message type
	BytesSentPerMessage map[string]int `json:"bytessent_per_msg"`
	// BytesRecvPerMessage is the total received bytes aggregated by
	// message type
	BytesRecvPerMessage map[string]int `json:"bytesrecv_per_msg"`
}

// NetworkInfo contains information about this node’s connection to the network.
type NetworkInfo struct {
	// Version is this node’s version of Bitcoin Core in its internal integer
	// format. For example, Bitcoin Core 0.9.2 has the integer version
	// number 90200.
	Version int `json:"version"`
	// Subversion is the user agent this node sends in its version message.
	Subversion string `json:"subversion"`
	// ProtocolVersion is the protocol version number used by this node. See
	// the protocol versions section for more information.
	ProtocolVersion int `json:"protocolversion"`
	// LocalServices list the services supported by this node as advertised in
	// its version message.
	LocalServices string `json:"localservices"`
	// LocalRelay is set if this node acts as a relay.
	LocalRelay bool `json:"localrelay"`
	// TimeOffset is the offset of the node’s clock from the computer’s clock
	// (both in UTC) in seconds. The offset may be up to 4200 seconds
	// (70 minutes).
	TimeOffset int `json:"timeoffset"`
	// NetworkActive
	NetworkActive bool `json:"networkactive"`
	// Connections is the total number of open connections (both outgoing and
	// incoming) between this node and other nodes.
	Connections int `json:"connections"`
	// Networks is an array with three objects: one describing the IPv4
	// connection, one describing the IPv6 connection, and one describing the
	// Tor hidden service (onion) connection.
	Networks []*struct {
		// Name is the name of the network. Either ipv4, ipv6, or onion.
		Name string `json:"name"`
		// Limited is set to true if only connections to this network are
		// allowed according to the -onlynet Bitcoin Core command-line/
		// configuration-file parameter. Otherwise set to false.
		Limited bool `json:"limited"`
		// Reachable is set to true if connections can be made to or from this
		// network. Otherwise set to false.
		Reachable bool `json:"reachable"`
		// Proxy is the hostname and port of any proxy being used for this
		// network. If a proxy is not in use, an empty string.
		Proxy string `json:"proxy"`
		// ProxyRandomizeCredentials is set to true if randomized credentials
		// are set for this proxy. Otherwise set to false.
		ProxyRandomizeCredentials bool `json:"proxy_randomize_credentials"`
	} `json:"networks"`
	// RelayFee is the minimum fee a low-priority transaction must pay in
	// order for this node to accept it into its memory pool.
	RelayFee float64 `json:"relayfee"`
	// IncrementalFee
	IncrementalFee float64 `json:"incrementalfee"`
	// LocalAddresses is an array of objects each describing the local
	// addresses this node believes it listens on.
	LocalAddresses []*struct {
		// Address is an IP address or .onion address this node believes it
		// listens on. This may be manually configured, auto detected, or based
		// on version messages this node received from its peers.
		Address string `json:"address"`
		// Port number this node believes it listens on for the associated
		// address. This may be manually configured, auto detected, or based
		// on version messages this node received from its peers.
		Port int `json:"port"`
		// Score is the number of incoming connections during the uptime of
		// this node that have used this address in their version message.
		Score int `json:"score"`
	} `json:"localaddresses"`
	// Warnings is a plain-text description of any network warnings. If there
	// are no warnings, an empty string will be returned.
	Warnings string `json:"warnings"`
}

// NetworkStats contains information about the node’s network totals.
type NetworkStats struct {
	// TotalBytesRecv is the total number of bytes received since the node was
	// last restarted.
	TotalBytesRecv int `json:"totalbytesrecv"`
	// TotalBytesSent is the total number of bytes sent since the node was
	// last restarted.
	TotalBytesSent int `json:"totalbytessent"`
	// TimeMillis is the Unix epoch time in milliseconds according to the
	// operating system’s clock (not the node adjusted time).
	TimeMillis int `json:"timemillis"`
	// UploadTarget is the upload target information.
	UploadTarget *struct {
		// Timeframe is the length of the measuring timeframe in seconds. The
		// timeframe is currently set to 24 hours.
		Timeframe int `json:"timeframe"`
		// Target is the maximum allowed outbound traffic in bytes. The default is
		// 0. Can be changed with -maxuploadtarget.
		Target int `json:"target"`
		// TargetReached indicates if the target is reached. If the target is
		// reached the node won’t serve SPV and historical block requests anymore.
		TargetReached bool `json:"target_reached"`
		// ServeHistoricalBlocks indicates if historical blocks are served.
		ServerHistoricalBlocks bool `json:"serve_historical_blocks"`
		// BytesLeftInCycle is the amount of bytes left in current time cycle.
		// 0 is displayed if no upload target is set.
		BytesLeftInCycle int `json:"bytes_left_in_cycle"`
		// TimeLeftInCycle is the number of seconds left in current time cycle.
		// 0 is displayed if no upload target is set.
		TimeLeftInCycle int `json:"time_left_in_cycle"`
	} `json:"uploadtarget"`
}
