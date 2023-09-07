package parser

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
	"strings"
	"testing"
)

func TestData(t *testing.T) {

	//=================================================================
	// Basic
	//=================================================================
	testCase1(t, "Simple=Test", "/Simple", "", "Simple=\"Test\"", "111")
	testCase1(t, "Test", "/#1", "", "\"Test\"", "111")
	testCase1(t, "\"Test\"", "/#1", "", "", "111")
	testCase1(t, "Simple=\"Test\"", "/Simple", "", "", "111")
	testCase1(t, "Simple=\"Test mit Spaces\"", "/#1", "/Simple", "", "111")
	testCase1(t, "Simple=\"Test,mit=Spaces{}\"", "/Simple", "", "", "111")

	testCase1(t, "Simple=Test,", "/Simple", "", "Simple=\"Test\"", "111")
	testCase1(t, "Simple=Test,Muster", "/#2", "", "\"Muster\"", "111")
	testCase1(t, "Simple=Test", "/#3", "", "", "10X")
	testCase1(t, "Simple=Test", "/#-1", "", "", "10X")

	testCase1(t, "List={Entry1=Value1,Entry2=Value2}", "/List", "", "List={}", "111")
	testCase1(t, "List={Entry1=Value1,Entry2=Value2}", "/List/#1", "/List/Entry1", "Entry1=\"Value1\"", "111")
	testCase1(t, "List={Entry1=Value1,{{", "", "", "", "0XX")
	testCase1(t, "List={Entry1=Value1,{{}", "", "", "", "0XX")
	testCase1(t, "List={Entry1=Value1,{{}}", "", "", "", "0XX")
	testCase1(t, "List={Entry1=Value1,,{}}", "/List/#2", "", "~", "111")
	testCase1(t, "List={}", "/List", "", "", "111")
	testCase1(t, "List={{}}", "/List", "", "List={}", "111")
	testCase1(t, "List={,,,}", "/List/#1", "", "~", "111")
	testCase1(t, "List={,,,}", "/List/#2", "", "~", "111")
	testCase1(t, "List={,,,}", "/List/#3", "", "~", "111")
	testCase1(t, "List={,,,}", "/List/#4", "", "~", "111")
}

var err error

// Parse string into data object.
func getData(s string) *Data {
	rdr := bufio.NewReader(strings.NewReader(s))
	d := new(Data)
	err = d.Read(rdr)
	if err != nil {
		return nil
	}
	return d
}

// Run test case.
func testCase1(t *testing.T, data, access, path, elem string, flags string) {
	t.Helper()

	if len(elem) == 0 {
		elem = data
	}
	if len(path) == 0 {
		path = access
	}
	d := getData(data)
	if (flags[0] == '1') != (d != nil) {
		t.Fatal("getdata failed")
	}
	if d != nil {
		e := d.Lookup(access)
		p := ""
		if e != nil {
			p = e.GetPath()
		}
		if (flags[1] == '1') != (p == path) {
			t.Fatal("lookup failed")
		}
		if e != nil {
			p = e.String()
			if (flags[2] == '1') != (p == ("`" + elem + "`")) {
				t.Fatal("elem failed")
			}
		}
	}
}
