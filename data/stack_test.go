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
	"testing"
)

func TestIntStack(t *testing.T) {
	is := NewIntStack()
	if is.Len() != 0 {
		t.Fatal("new stack not empty")
	}
	for i := 0; i < 10; i++ {
		is.Push(i)
		if !is.IsTop(i) {
			t.Fatal("push/istop failed")
		}
	}
	if is.Len() != 10 {
		t.Fatal("length mismatch")
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if is.Peek() != j {
			t.Fatal("peek failed")
		}
		if is.Pop() != j {
			t.Fatal("pop failed")
		}
		if is.Len() != j {
			t.Fatal("length mismatch")
		}
	}
}

func TestStringStack(t *testing.T) {
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
		"ff", "gg", "hh", "ii", "kk",
	}
	ss := NewStack()
	if ss.Len() != 0 {
		t.Fatal("new stack not empty")
	}
	for _, v := range list {
		ss.Push(v)
	}
	if ss.Len() != 10 {
		t.Fatal("length mismatch")
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if ss.Peek().(string) != list[j] {
			t.Fatal("peek failed")
		}
		if ss.Pop().(string) != list[j] {
			t.Fatal("pop failed")
		}
		if ss.Len() != j {
			t.Fatal("length mismatch")
		}
	}
}
