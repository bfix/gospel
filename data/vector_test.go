package data

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
	"testing"
)

func TestVector(t *testing.T) {
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
	}

	vec := NewVector() // ""
	if vec.Len() != 0 {
		t.Fatal("new vector not empty")
	}
	for _, v := range list {
		vec.Add(v)
	}
	if vec.Len() != 5 { // "aa" "bb" "cc" "dd" "ee"
		t.Fatal("length mismatch")
	}
	vec.Insert(-3, "mm") // "mm" nil nil "aa" "bb" "cc" "dd" "ee"
	if vec.At(3).(string) != list[0] {
		t.Fatal("prepending failed")
	}
	if vec.Len() != 8 {
		t.Fatal("length mismatch")
	}
	vec.Insert(10, "nn") // "mm" nil nil "aa" "bb" "cc" "dd" "ee" nil nil "nn"
	if vec.Len() != 11 {
		t.Fatal("size mismatch")
	}
	if vec.Delete(7).(string) != list[4] { // "mm" nil nil "aa" "bb" "cc" "dd" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
		t.Fatal("delete failed")
	}
	if vec.Len() != 10 {
		t.Fatal("length mismatch")
	}
	vec.Insert(5, "pp") // "mm" nil nil "aa" "bb" "pp" "cc" "dd" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
	if vec.Len() != 11 {
		t.Fatal("length mismatch")
	}
}
