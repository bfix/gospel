package parser

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
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
	testCase1("Simple=Test", "/Simple", "", "Simple=\"Test\"", t, "111")
	testCase1("Test", "/#1", "", "\"Test\"", t, "111")
	testCase1("\"Test\"", "/#1", "", "", t, "111")
	testCase1("Simple=\"Test\"", "/Simple", "", "", t, "111")
	testCase1("Simple=\"Test mit Spaces\"", "/#1", "/Simple", "", t, "111")
	testCase1("Simple=\"Test,mit=Spaces{}\"", "/Simple", "", "", t, "111")

	testCase1("Simple=Test,", "/Simple", "", "Simple=\"Test\"", t, "111")
	testCase1("Simple=Test,Muster", "/#2", "", "\"Muster\"", t, "111")
	testCase1("Simple=Test", "/#3", "", "", t, "10X")
	testCase1("Simple=Test", "/#-1", "", "", t, "10X")

	testCase1("List={Entry1=Value1,Entry2=Value2}", "/List", "", "List={}", t, "111")
	testCase1("List={Entry1=Value1,Entry2=Value2}", "/List/#1", "/List/Entry1", "Entry1=\"Value1\"", t, "111")
	testCase1("List={Entry1=Value1,{{", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,{{}", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,{{}}", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,,{}}", "/List/#2", "", "~", t, "111")
	testCase1("List={}", "/List", "", "", t, "111")
	testCase1("List={{}}", "/List", "", "List={}", t, "111")
	testCase1("List={,,,}", "/List/#1", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#2", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#3", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#4", "", "~", t, "111")
}

var err error

// Parse string into data object.
func getData(s string, t *testing.T) *Data {
	rdr := bufio.NewReader(strings.NewReader(s))
	d := new(Data)
	err = d.Read(rdr)
	if err != nil {
		return nil
	}
	return d
}

// Run test case.
func testCase1(data, access, path, elem string, t *testing.T, flags string) {
	if len(elem) == 0 {
		elem = data
	}
	if len(path) == 0 {
		path = access
	}
	d := getData(data, t)
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
