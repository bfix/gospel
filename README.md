
[![Build Status](https://travis-ci.org/bfix/gospel.svg?branch=master)](https://travis-ci.org/bfix/gospel)
[![Go Report Card](https://goreportcard.com/badge/github.com/bfix/gospel)](https://goreportcard.com/report/github.com/bfix/gospel)

Gospel: GO SPEcial Library
==========================

(c) 2010-2017 Bernd Fix <brf@hoi-polloi.org>   >Y<

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at
your option) any later version.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

Packages
--------

- gospel/parser: Read/access/write nested data structures
- gospel/network: Network-related functionality
    * services
    * packet handling
    * TOR
    * SOCKS5 connection handler
    * SMTP/POP3 mail handling
- gospel/bitcoin/ecc: Elliptic curve crypto
    * Secp256k1, as used by bitcoin
    * JSON-RPC to Bitcoin server
    * Utility methods (addresses, key exchange, raw
      transactions, hasing, Base58,...)
- gospel/math: Mathematical helpers
    * Fast Fourier Transformation
- gospel/crypto: cryptographic helpers
    * secret sharing
    * prime fields
    * PRNG
    * Paillier crypto scheme
    * cryptographic counters
- gospel/logger: logging facilities
- gospel/data: useful data structures like
    * stacks
    * vectors

Install
-------

This version is designed for the Go1.8 release.

Make sure that your Go environment - especially ${GOPATH} - is set up and
enter the following command:

    $ go get github.com/bfix/gospel/...
    
Test notes
----------

To successfully run the Bitcoin-JSON-RPC tests, follow these steps:

#### 1. Edit the 'bitcoin.conf' configuration file

Use an ASCII editor to set the RPC variables to appropriate values:
   
    server=1
    rpcuser=<username>
    rpcpassword=<password>
    rpctimeout=30
    rpcport=8332
   
#### 2. Prepare an encrypted test wallet

Make sure that the test wallet has some transactions in it, if you
want to run all implemented test paths.
   
#### 3. Start the Bitcoin daemon for the test wallet

    $ bitcoind -wallet=<your test wallet>

#### 4. Export environment variables for the RPC tests

    $ export BTC_HOST="http://127.0.0.1:8332"
    $ export BTC_USER="<username>"
    $ export BTC_PASSWORD="<password>"
    $ export BTC_WALLET_PP="<your test wallet passphrase>"
    $ export BTC_WALLET_COPY="/tmp/testwallet.dat"
   
   The following environment variables are optional and listed with
   their default value; you can replace them with other valid data:

    $ export BTC_PRIVKEY="L3W5UAHUmxYHF3iE7Biaky7JXA94o1NWrCFT3BMpq1FrzorfbPeM"
    $ export BTC_BLOCK_HASH="00000000000003fab35380c07f6773ae27727b21016a8821c88e47e241c86458"

#### 5. Run the tests

    $ go test github.com/bfix/gospel/bitcoin/rpc
