package tor

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix
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
	"fmt"
	"os"
	"testing"
)

var (
	testCtrl *Control = nil
	passwd   string
	err      error
)

func TestMain(m *testing.M) {
	proto := os.Getenv("TOR_CONTROL_PROTO")
	if len(proto) == 0 {
		proto = "tcp"
	}
	endp := os.Getenv("TOR_CONTROL_ENDPOINT")
	if len(endp) == 0 {
		endp = "127.0.0.1:9052"
	}
	if passwd = os.Getenv("TOR_CONTROL_PASSWORD"); len(passwd) == 0 {
		fmt.Println("Skipping 'network/tor' tests!")
		return
	}
	testCtrl, err = NewControl(proto, endp)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}
	m.Run()
	testCtrl.Close()
}
