package main

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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

import (
	"flag"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/bitcoin/wallet"
)

func main() {
	// get command-line arguments
	var caseSensitive bool
	flag.BoolVar(&caseSensitive, "s", false, "Case sensitive search")
	flag.Parse()
	prefixes := flag.Args()
	num := len(prefixes)
	if num == 0 {
		log.Println("No prefixes specified -- done.")
		return
	}
	// pre-compile regexp for given prefixes
	reg := make([]*regexp.Regexp, num)
	for i, p := range prefixes {
		if !caseSensitive {
			p = strings.ToLower(p)
		}
		reg[i] = regexp.MustCompile(p)
	}
	// try to find matches forever...
	start := time.Now()
	for i := 0; ; i++ {
		priv := bitcoin.GenerateKeys(true)
		addr, err := wallet.MakeAddress(&priv.PublicKey, 0, wallet.AddrP2PKH, wallet.NetwMain)
		if err != nil {
			log.Fatal(err)
		}
		test := addr
		if !caseSensitive {
			test = strings.ToLower(test)
		}
		for _, r := range reg {
			if r.MatchString(test) {
				elapsed := time.Since(start)
				kd := bitcoin.ExportPrivateKey(priv, false)
				log.Printf("%s [%s] (%d tries, %s elapsed)\n", addr, kd, i, elapsed)
				i = 0
				start = time.Now()
			}
		}
	}
}
