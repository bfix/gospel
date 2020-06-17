package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
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
	"encoding/json"
	"fmt"
	"os"
)

var (
	sess    *Session
	info    *Info
	verbose = false
)

func init() {
	fmt.Printf("init(): ")
	var err error
	rpcaddr := os.Getenv("BTC_HOST")
	user := os.Getenv("BTC_USER")
	passwd := os.Getenv("BTC_PASSWORD")
	if sess, err = NewSession(rpcaddr, user, passwd); err != nil {
		fmt.Println(err.Error())
		sess = nil
		return
	}
	strictCheck = true
	if info, err = sess.GetInfo(); err != nil {
		fmt.Println(err.Error())
		sess = nil
		return
	}
	if !info.TestNet {
		fmt.Println("Node not on testnet")
		sess = nil
		return
	}
	fmt.Println("Bitcoin JSON-RPC tests initialized!")
}

func dumpObj(fmtStr string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "    ")
	if err == nil {
		fmt.Printf(fmtStr, string(data))
	} else {
		fmt.Printf(fmtStr, "<"+err.Error()+">")
	}
}
