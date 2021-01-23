
Gospel: Bitcoin tools
=====================

(c) 2010-2021 Bernd Fix <brf@hoi-polloi.org>   >Y<

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

## BIP39 seed generation from passphrase

Compile the program by running

```bash
go install github.com/bfix/gospel/bitcoin/tools/passphrase2seed
```

The program will prompt for a passphrase (CAUTION: input is echoed to the
terminal) and emit 24 BIP39-compliant seed words. The seed is computed from
the SHA256 hash of the passphrase (without a terminating LF or CR/LF)

## Vanity Bitcoin address

Compile the program by running

```bash
go install github.com/bfix/gospel/bitcoin/tools/vanityaddress
```

The program will find vanity Bitcoin addresses for given regular expressions
passed as command-line arguments, e.g.

```bash
bin/vanityaddress "^1myname" "test"
```

will try to generate Bitcoin addresses that start with "1myname" or have
the string "test" in them (at no specific position) in a case-insensitive
match. You can use the `-s` option to make the match case-sensitive, so

```bash
bin/vanityaddress -s "^1MyName"
```

will only return Bitcoin addresses that match "1MyName..." exactly.

The associated private key for the address is included in the output as
are some runtime statistics.
