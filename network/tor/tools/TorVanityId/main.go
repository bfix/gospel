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

package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/network/tor"
)

func main() {
	// handle command-line arguments
	flag.Parse()

	// handle custom patterns
	patterns := flag.Args()
	num := len(patterns)
	if num == 0 {
		fmt.Println("No patterns specified -- done.")
		return
	}
	// pre-compile regexp
	reg := make([]*regexp.Regexp, num)
	for i, p := range patterns {
		reg[i] = regexp.MustCompile(p)
	}
	// allocate key material
	seed := make([]byte, 32)

	// generate new keys in a loop
	fmt.Println("Generating vanity key")
	start := time.Now()
	for i := 0; ; i++ {
		_, _ = rand.Read(seed)
		prv := ed25519.NewPrivateKeyFromSeed(seed)
		id, err := tor.ServiceID(prv.Public())
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range reg {
			if r.MatchString(id) {
				elapsed := time.Since(start)
				s1 := hex.EncodeToString(seed)
				s2 := hex.EncodeToString(prv.D.Bytes())
				fmt.Printf("%s [%s][%s] (%d tries, %s elapsed)\n", id, s1, s2, i, elapsed)
				i = 0
				start = time.Now()
			}
		}
	}
}
