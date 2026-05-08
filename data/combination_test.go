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

package data

import (
	"slices"
	"testing"
)

func TestCombinations(t *testing.T) {
	in := []int{1, 2, 3}
	out := [][]int{
		{1}, {1, 2}, {1, 2, 3}, {1, 3}, {2}, {2, 3}, {3},
	}
	for i, c := range Combinations(in) {
		if !slices.Equal(c, out[i]) {
			t.Fatalf("%v != %v", c, out[i])
		}
	}
}

func TestCombinator(t *testing.T) {
	in := [][]int{
		{1, 7},
		{2, 5, 8},
		{4, 9},
	}
	out := [][]int{
		{1, 2, 4}, {7, 2, 4}, {1, 5, 4}, {7, 5, 4}, {1, 8, 4}, {7, 8, 4},
		{1, 2, 9}, {7, 2, 9}, {1, 5, 9}, {7, 5, 9}, {1, 8, 9}, {7, 8, 9},
	}
	ci := NewCombinator(in)
	for i := 0; ; i++ {
		list, done := ci.Next()
		if !slices.Equal(list, out[i]) {
			t.Fatalf("%v != %v", list, out[i])
		}
		if done {
			break
		}
	}
}
