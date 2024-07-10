//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	// "/run/tor/control.authcookie" base64-encoded
	enc := "L3J1bi90b3IvY29udHJvbC5hdXRoY29va2ll"
	fname, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file, err := os.Open(string(fname))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(hex.EncodeToString(data))
	os.Exit(0)
}
