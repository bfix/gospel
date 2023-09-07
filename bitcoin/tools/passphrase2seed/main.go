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
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bfix/gospel/bitcoin/wallet"
)

func main() {

	fmt.Printf(">>> Passphrase (Entropy): ")
	rdr := bufio.NewReader(os.Stdin)
	pp1, _, err := rdr.ReadLine()
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	ent := sha256.Sum256(pp1)
	words, err := wallet.EntropyToWords(ent[:])
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	fmt.Printf(">>>    Password (BIP 39): ")
	pp2, _, err := rdr.ReadLine()
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	seed, _ := wallet.WordsToSeed(words, string(pp2))

	fmt.Printf("<<<     Entropy: %s\n", hex.EncodeToString(ent[:]))
	fmt.Printf("<<< BIP-32 seed: %s\n", hex.EncodeToString(seed))
	fmt.Println("<<< BIP-39 seed:")
	n := len(words) / 2
	for i := 0; i < n; i++ {
		fmt.Printf("<<<    %2d: %-20s %2d: %-20s\n", i+1, words[i], i+n+1, words[i+n])
	}

	hd, err := wallet.NewHD(seed)
	if err != nil {
		fmt.Println("<<< ERROR: " + err.Error())
		return
	}
	pk := hd.MasterPublic()
	fmt.Printf("<<< BIP-32 Master/Root Pub: %s\n", pk)
	sk := hd.MasterPrivate()
	fmt.Printf("<<< BIP-32 Master/Root Prv: %s\n", sk)

	type task struct {
		av   *wallet.AddrVersion
		name string
		path string
	}
	var tasks = []task{
		{wallet.AddrList[0].Formats[0].Versions[0], "BIP-44", "m/44'/0'/0'"},
		{wallet.AddrList[0].Formats[0].Versions[4], "BIP-49", "m/49'/0'/0'"},
		{wallet.AddrList[0].Formats[0].Versions[2], "BIP-84", "m/84'/0'/0'"},
	}

	for _, t := range tasks {
		// create a HD wallet for the given seed
		fmt.Printf("<<< %s\n", t.name)
		bsk, err := hd.Private(t.path)
		if err != nil {
			fmt.Println("<<< ERROR: " + err.Error())
			continue
		}
		bsk.Data.Version = t.av.PrvVersion
		fmt.Printf("<<<     %s\n", bsk)

		bpk, err := hd.Public(t.path)
		if err != nil {
			fmt.Println("<<< ERROR: " + err.Error())
			continue
		}
		bpk.Data.Version = t.av.PubVersion
		fmt.Printf("<<<     %s\n", bpk)
	}
}
