package rpc

/*
 * ====================================================================
 * Bitcoin RPC to local server.
 * ====================================================================
 *
 * This functions allow to create a session to the local Bitcoin server
 * (local instance of bitcoind) and to query the instance for data.
 *
 * (c) 2011-2019 Bernd Fix   >Y<
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
 *
 * ====================================================================
 *
 * The 'Session' methods are implementing Bitcoin JSON-RPC API calls,
 * that are documented on the following webpages:
 *
 * -- http://blockchain.info/api/json_rpc_api
 * -- https://bitcoin.org/en/developer-reference#bitcoin-core-apis
 * -- https://en.bitcoin.it/wiki/Original_Bitcoin_client/API_calls_list
 * -- https://en.bitcoin.it/wiki/Elis-API
 * -- https://en.bitcoin.it/wiki/Raw_Transactions
 *
 * ====================================================================
 *
 *  Method:                 File:
 *  ----------------------+--------------------------
 *  AbandonTransaction      wallet.go
 *  AddMultiSigAddress      wallet.go
 *  AddNode                 node.go
 *  AddWitnessAddress       script.go
 *  BackupWallet            wallet.go
 *  ClearBanned             node.go
 *  CreateMultiSig          wallet.go
 *  CreateRawTransaction    transaction.go
 *  DecodeRawTransaction    transaction.go
 *  DecodeScript            script.go
 *  DisconnectNode          node.go
 *  DumpPrivKey             wallet.go
 *  DumpWallet              wallet.go
 *  EncryptWallet           wallet.go
 *  EstimateFee             transaction.go
 *  EstimatePriority        transaction.go
 *  FundRawTransaction      transaction.go
 *  Generate                block.go
 *  GenerateToAddress       block.go
 *  GetAccountAddress       wallet.go
 *  GetAccount              wallet.go
 *  GetAddedNodeInfo        node.go
 *  GetAddressesByAccount   wallet.go
 *  GetBalance              wallet.go
 *  GetBestBlockHash        block.go
 *  GetBlock                block.go
 *  GetBlockChainInfo       block.go
 *  GetBlockCount           block.go
 *  GetBlockHash            block.go
 *  GetBlockHeader          block.go
 *  GetBlockTemplate        block.go
 *  GetChainTips            block.go
 *  GetConnectionCount      node.go
 *  GetDifficulty           local.go
 *  GetGenerate             ---
 *  GetHashesPerSec         ---
 *  GetInfo                 local.go
 *  GetMemPoolAncestors     local.go
 *  GetMemPoolDescendants   local.go
 *  GetMemPoolEntry         local.go
 *  GetMemPoolInfo          local.go
 *  GetMiningInfo           node.go
 *  GetNetTotals            node.go
 *  GetNetworkHashPS        node.go
 *  GetNetworkInfo          node.go
 *  GetNewAddress           wallet.go
 *  GetPeerInfo             node.go
 *  GetRawChangeAddress     wallet.go
 *  GetRawMemPool           local.go
 *  GetRawTransaction       transaction.go
 *  GetReceivedByAccount    wallet.go
 *  GetReceivedByAddress    wallet.go
 *  GetTransaction          transaction.go
 *  GetTxOut                transaction.go
 *  GetTxOutProof			transaction.go
 *  GetTxOutSetInfo         transaction.go
 *  GetUnconfirmedBalance   wallet.go
 *  GetWalletInfo           wallet.go
 *  GetWork                 ---
 *  Help                    ---
 *  ImportAddress           wallet.go
 *  ImportPrivKey           wallet.go
 *  ImportPrunedFunds       wallet.go
 *  ImportWallet            wallet.go
 *  KeyPoolRefill           wallet.go
 *  ListAccounts            wallet.go
 *  ListAddressGroupings    wallet.go
 *  ListBanned              node.go
 *  ListLockUnspent         wallet.go
 *  ListReceivedByAccount   wallet.go
 *  ListReceivedByAddress   wallet.go
 *  ListSinceBlock          transaction.go
 *  ListTransactions        transaction.go
 *  ListUnspent             wallet.go
 *  LockUnspent             wallet.go
 *  Move                    wallet.go
 *  Ping                    node.go
 *  PrioritiseTransaction   transaction.go
 *  RemovePrunedFunds       wallet.go
 *  SendFrom                wallet.go
 *  SendMany                wallet.go
 *  SendRawTransaction      transaction.go
 *  SendToAddress           wallet.go
 *  SetAccount              wallet.go
 *  SetBan                  node.go
 *  SetGenerate             ---
 *  SetTxFee                wallet.go
 *  SignMessage             wallet.go
 *  SignMessageWithPrivKey  wallet.go
 *  SignRawTransaction      transaction.go
 *  Stop                    local.go
 *  SubmitBlock             block.go
 *  ValidateAddress         wallet.go
 *  VerifyChain             block.go
 *  VerifyMessage           wallet.go
 *  VerifyTxOutProof        transaction.go
 *  WalletLock              wallet.go
 *  WalletPassphrase        wallet.go
 *  WalletPassphraseChange  wallet.go
 */
