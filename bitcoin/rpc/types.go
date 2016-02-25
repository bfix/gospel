package rpc

// Info is a generic information about running Bitcoin server.
type Info struct {
	Version         int
	ProtocolVersion int
	WalletVersion   int
	Proxy           string
	TestNet         bool
	Connections     int
	KeyPoolSize     int
	TimeOffset      int
	KeyPoolOldest   int
	Balance         float64
	Errors          string
	PayTxFee        float64
	Difficulty      float64
	Blocks          int
}

// Block (element of the Bitcoin blockchain)
type Block struct {
	IDList            []string
	Time              int
	Height            int
	Nonce             int
	Confirmations     int
	Hash              string
	PreviousBlockHash string
	NextBlockHash     string
	Bits              string
	Difficulty        int
	MerkleRoot        string
	Version           int
	Size              int
}

//Transaction is a Bitcoin transaction
type Transaction struct {
	Amount        float64
	Fee           float64
	BlockIndex    int
	Confirmations int
	ID            string
	BlockHash     string
	Time          int
	BlockTime     int
	TimeReceived  int
}

// Vinput is a raw transaction input slot
type Vinput struct {
	ID        string
	Vout      int
	ScriptSig string
	Sequence  int
}

// Voutput is a raw transaction output slot
type Voutput struct {
	Value        float64
	N            int
	ScriptPubKey string
	ReqSigs      int
	Type         string
	Addresses    []string
}

// RawTransaction is a Bitcoin transaction in raw format
type RawTransaction struct {
	ID       string
	Version  int
	LockTime int
	Vin      []Vinput
	Vout     []Voutput
}

// Output of a transaction
type Output struct {
	ID           string
	Vout         int
	ScriptPubKey string
	RedeemScript string
}

// Unspent transactions for accounts
type Unspent struct {
	Output
	Amount        float64
	Confirmations int
}

// Received transactions for account/address (accumulated)
type Received struct {
	Account       string
	Label         string
	Address       string
	Amount        float64
	Confirmations int
}

// Validity check on address
type Validity struct {
	Address      string
	IsCompressed bool
	Account      string
	PubKey       string
	IsMine       bool
	IsValid      bool
}

// Balance of Bitcoin address (used for outgoing transactions as well)
type Balance struct {
	Address string
	Amount  float64
}
