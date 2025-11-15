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

package math

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

func TestFibIO(t *testing.T) {

	kn1 := NewKnacci(8, NewInt(1048573))
	for range 10000 {
		kn1.Next()
	}
	d1 := new(bytes.Buffer)
	kn1.Write(d1)
	b1 := d1.Bytes()

	d2 := new(bytes.Buffer)
	kn2 := NewKnacci(8, NewInt(1048573))
	var err error
	for range 10000 {
		step, _ := kn2.Next()
		if step%1000 == 0 {
			kn2.Write(d2)
			rdr := bytes.NewReader(d2.Bytes())
			if kn2, err = ReadKnacci(rdr); err != nil {
				t.Fatal(err)
			}
			d2 = new(bytes.Buffer)
		}
	}
	kn2.Write(d2)
	b2 := d2.Bytes()

	if !slices.Equal(b1, b2) {
		t.Fatal("mismatch")
	}
}

func TestFib1(t *testing.T) {
	N := NewInt(1048573)
	d := NewIntRndRange(THREE, N)
	init := []*Int{
		d,
		NewInt(2),
		NewInt(3),
		NewInt(5),
	}
	kn := NewKnacciInt(N, init...)
	f2, p1, p2 := kn.Recurrence(1e9, "")
	if p1 > 0 {
		fmt.Printf("recurrence after %d steps.\n", p2)
		f1 := kn.Factors(p1)
		x := kn.Solve(f1, f2)
		if x.Equals(d) {
			fmt.Println("SUCCESS!")
		} else {
			fmt.Println("mismatch")
		}
	} else {
		fmt.Println("failed.")
	}
}
