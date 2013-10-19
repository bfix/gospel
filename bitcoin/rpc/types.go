/*
 * Bitcoin RPC return types.
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
// public types

//---------------------------------------------------------------------
/*
 * Generic infomation about running Bitcoin server.
 */
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

//---------------------------------------------------------------------
/*
 * Block (element of the Bitcoin blockchain)
 */
type Block struct {
	TxIDs             []string
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

//---------------------------------------------------------------------
/*
 * Bitcoin transaction
 */
type Transaction struct {
	Amount        float64
	Fee           float64
	BlockIndex    int
	Confirmations int
	TxID          string
	BlockHash     string
	Time int
	BlockTime int
	TimeReceived int
}

//---------------------------------------------------------------------
/*
 * Unspent transactions for accounts
 */
type Unspent struct {
	Id            string
	Vout          int
	ScriptPubkey  string
	Amount        float64
	Confirmations int
}

//---------------------------------------------------------------------
/*
 * Received transactions for account/address (accumulated)
 */
type Received struct {
	Account       string
	Label         string
	Address       string
	Amount        float64
	Confirmations int
}

//---------------------------------------------------------------------
/*
 * Validity check on address
 */
type Validity struct {
	Address      string
	IsCompressed bool
	Account      string
	PubKey       string
	IsMine       bool
	IsValid      bool
}

//---------------------------------------------------------------------
/*
 * Balance of Bitcoin address (used for outgoing transactions as well)
 */
type Balance struct {
	Address string
	Amount float64
}
