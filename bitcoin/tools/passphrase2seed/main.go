package main

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix  >Y<
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
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bfix/gospel/bitcoin/wallet"
)

func main() {

	fmt.Printf(">>> Passphrase: ")
	rdr := bufio.NewReader(os.Stdin)
	in, _, err := rdr.ReadLine()
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	ent := sha256.Sum256(in)
	words, err := wallet.EntropyToWords(ent[:])
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	seed, _ := wallet.WordsToSeed(words, "")

	fmt.Printf("<<<    Entropy: %s\n", hex.EncodeToString(ent[:]))
	fmt.Printf("<<<       Seed: %s\n", hex.EncodeToString(seed))
	fmt.Println("<<< Seed words:")
	n := len(words) / 2
	for i := 0; i < n; i++ {
		fmt.Printf("<<<    %2d: %-20s %2d: %-20s\n", i+1, words[i], i+n+1, words[i+n])
	}
}
