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
	"fmt"
	"testing"
)

func TestConnectionCount(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	conns, err := sess.GetConnectionCount()
	if err != nil {
		t.Fatal(err)
	}
	if conns != info.Connections {
		t.Fatal(fmt.Sprintf("session-count mismatch: %d != %d", conns, info.Connections))
	}
}

func TestDifficulty(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	diff, err := sess.GetDifficulty()
	if err != nil {
		t.Fatal(err)
	}
	if diff != info.Difficulty {
		t.Fatal("difficulty mismatch in info")
	}
}

func TestFee(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	if err := sess.SetTxFee(0.0001); err != nil {
		t.Fatal(err)
	}
}

func TestMemPoolInfo(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	mi, err := sess.GetMemPoolInfo()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("MemPoolInfo: %s\n", mi)
	}
}

func TestGetRawMemPool(t *testing.T) {
	if sess == nil {
		t.Skip("skipping test: session not available")
	}
	list, err := sess.GetRawMemPoolList()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("RawMemPoolList: %s\n", list)
	}
	tx, err := sess.GetRawMemPool()
	if err != nil {
		t.Fatal(err)
	}
	if verbose {
		dumpObj("RawMemPool: %s\n", tx)
	}
	for _, k := range list {
		if _, ok := tx[k]; !ok {
			t.Fatal(fmt.Sprintf("Unknown key '%s'", k))
		}
	}
}
