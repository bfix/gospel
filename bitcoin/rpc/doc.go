package rpc

/*
 * ====================================================================
 * Bitcoin RPC to local server.
 * ====================================================================
 *
 * This functions allow to create a session to the local Bitcoin server
 * (local instance of bitcoind) and to query the instance for data.
 *
 * (c) 2011-2017 Bernd Fix   >Y<
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
 * -- https://en.bitcoin.it/wiki/Raw_Transactions (with corrections from
 *    the author, where required)
 */
