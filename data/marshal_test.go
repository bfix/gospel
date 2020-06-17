package data

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
	"encoding/hex"
	"fmt"
	"testing"
)

type NestedStruct struct {
	A int64 `order:"big"`
	B int32
}

func (n *NestedStruct) String() string {
	return fmt.Sprintf("NS(%v)", *n)
}

type SubStruct struct {
	G int32
}

func (s *SubStruct) String() string {
	return fmt.Sprintf("SS(%v)", *s)
}

type MainStruct struct {
	C uint64 `order:"big"`
	D string
	F *SubStruct
	G uint32
	E []*NestedStruct `size:"G"`
	H []uint32        `size:"5" order:"big"`
}

func TestNested(t *testing.T) {
	r := new(MainStruct)
	r.C = 19031962
	r.D = "Just a test"
	r.E = make([]*NestedStruct, 3)
	r.F = new(SubStruct)
	r.G = 3
	r.F.G = 0x23
	for i := 0; i < 3; i++ {
		n := new(NestedStruct)
		n.A = int64(255 - i)
		n.B = int32(815 * (i + 1))
		r.E[i] = n
	}
	r.H = make([]uint32, 5)
	for i := range r.H {
		r.H[i] = uint32(i * 23)
	}

	data, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("<<< %v\n", r)
	fmt.Printf("    [%s]\n", hex.EncodeToString(data))

	s := new(MainStruct)
	if err = Unmarshal(s, data); err != nil {
		t.Fatal(err)
	}
	fmt.Printf(">>> %v\n", s)
}
