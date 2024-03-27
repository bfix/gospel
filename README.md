
# Gospel: GO SPEcial Library (v1.2.28)

(c) 2011-2023 Bernd Fix <brf@hoi-polloi.org>   >Y<

Gospel is free software: you can redistribute it and/or modify it
under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

Gospel is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

SPDX-License-Identifier: AGPL3.0-or-later

[![Build Status](https://travis-ci.org/bfix/gospel.svg?branch=master)](https://travis-ci.org/bfix/gospel)
[![Go Report Card](https://goreportcard.com/badge/github.com/bfix/gospel)](https://goreportcard.com/report/github.com/bfix/gospel)
[![GoDoc](https://godoc.org/github.com/bfix/gospel?status.svg)](https://godoc.org/github.com/bfix/gospel)

## Packages

- gospel/network: Network-related functionality
  - services
  - packet handling
  - SOCKS5 connection handler
  - SMTP/POP3 mail handling
- gospel/network/p2p:
  - P2P core library
- gospel/network/tor:
  - Tor controller
  - hidden services (onion handling)
  - Tor utilities
- gospel/network/tor/tools:
  - TorAuthCookie
- gospel/bitcoin:
  - Elliptic curve crypto (Secp256k1)
  - Bitcoin addresses (and derived coins)
  - key exchange
  - hash functions (Hash160, Hash256)
  - base58 encoding
- gospel/bitcoin/wallet:
  - HD key space
  - BIP39 seed words
- gospel/bitcoin/script: Bitcoin script parser/interpreter
- gospel/bitcoin/tools:
  - passphrase2seed
  - vanityaddress
- gospel/math: Mathematical helpers
  - Fast Fourier Transformation
  - Arbitrary precision integers with chainable methods
- gospel/crypto: cryptographic helpers
  - secret sharing
  - prime fields
  - PRNG
  - Paillier crypto scheme
  - cryptographic counters
- gospel/crypto/ed25519:
  - general purpose Ed25519 crypto
- gospel/logger: logging facilities
- gospel/concurrent:
  - Signaller (signal relay)
  - Dispatcher (Workload distribution to go-routine)
- gospel/data:
  - Marshal/Unmarshal Golang objects
  - Bloom filter
  - Generators
  - S-expressions
- gospel/parser: Read/access/write nested data structures

## Install

This version (`v1.2.28`) is designed for the Go1.20+ release to make use
of new language features.

If you only want to use the library in your projects, you don't have to
install anything. Just include `github.com/bfix/gospel v1.2.28` in your
`go.mod` file and do a `go mod tidy`.

You can install Gospel locally if desired. Make sure that your Go
environment - especially ${GOPATH} - is set up and enter the
following commands:

```bash
git clone get https://github.com/bfix/gospel
cd gospel
go mod tidy
```

## Test notes

### Network-related tests

#### 1. Tor-related tests

To run the Tor-related tests you need to set some environment variables to
access the Tor control port:

```bash
export TOR_CONTROL_PROTO="tcp"
export TOR_CONTROL_ENDPOINT="127.0.0.1:9051"
export TOR_CONTROL_PASSWORD="my_torcontrol_secret"
export TOR_TEST_HOST="127.0.0.1"
```

Only `TOR_CONTROL_PASSWORD` is mandatory; `TOR_CONTROL_PROTO`,
`TOR_CONTROL_ENDPOINT` and `TOR_TEST_HSHOST` default to the above values.

If Tor is not running on localhost (127.0.0.1), but remotely (either on
a separate machine or in a Docker container), the environment variables
for Tor tests above need to be adjusted. Let `1.2.3.4` be the IP address
of the Tor instance and `5.6.7.8` the IP address of the system running
the tests. Both systems must be interconnected, so they can talk to each
other. The settings in this case would look like:

```bash
export TOR_CONTROL_PROTO="tcp"
export TOR_CONTROL_ENDPOINT="1.2.3.4:9051"
export TOR_CONTROL_PASSWORD="my_torcontrol_secret"
export TOR_TEST_HOST="5.6.7.8"
```

N.B.: You have to make sure that host `5.6.7.8` can access the Tor control
port and Tor socks proxy ports (see `torrc` settings on the Tor instance).
