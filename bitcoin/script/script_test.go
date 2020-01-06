package script

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
	"encoding/hex"
	"fmt"
	"testing"
)

var (
	scr = []string{
		"3045022074f35af390c41ef1f5395d11f6041cf55a6d7dab0acdac8ee746c1f2de7a43b3022100b3dc3d916b557d378268a856b8f9a98b9afaf45442f5c9d726fce343de835a5801 " +
			"02c34538fc933799d972f55752d318c0328ca2bacccd5c7482119ea9da2df70a2f " +
			"OP_DUP " +
			"OP_HASH160 " +
			"5e4ff47ceb3a51cdf7ddd80afc4acc5a692dac2d " +
			"OP_EQUALVERIFY " +
			"OP_CHECKSIG",
		"#12 OP_DUP OP_ADD #24 OP_EQUALVERIFY",
		"5af390c41ef1f539 28ca2bacccd5c748 OP_2DUP #4 OP_DUP #8 OP_DUP #16 OP_DUP #32 OP_DUP #64 OP_DUP #128 OP_DUP #256 OP_DUP #512 OP_DUP #1024 OP_DUP #1953 OP_DUP #23 38fc933799d972f5",
	}
)

func TestParse(t *testing.T) {
	for _, hexScript := range s {
		scr, rc := Parse(hexScript)
		if rc != RcOK {
			t.Fatal(fmt.Sprintf("Parse failed: rc=%s", RcString[rc]))
		}
		if verbose {
			fmt.Printf("Statements: %v\n", scr.Stmts)
		}
		h2 := hex.EncodeToString(scr.Bytes())
		if h2 != hexScript {
			if true {
				fmt.Println("<<< " + hexScript)
				fmt.Println(">>> " + h2)
			}
			t.Fatal(fmt.Sprintf("Hex script mismatch"))
		}
	}
}

func TestCompile(t *testing.T) {
	for _, src := range scr {
		bin, err := Compile(src)
		if err != nil {
			t.Fatal(err)
		}
		src2 := bin.Decompile()
		if src != src2 {
			if true {
				fmt.Println(">>> " + src)
				fmt.Println("    " + hex.EncodeToString(bin.Bytes()))
				fmt.Println("<<< " + src2)
			}
			t.Fatal("Script compile/decompile mismatch")
		}
	}
}
