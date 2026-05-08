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
)

// Combinations generates (order-independent) lists from
// a given list of elements.
//
//	Example: in=[1,2,3] returns [1],[2],[3],[1,2],[1,3],[2,3],[1,2,3]
func Combinations[T any](in []T) (list [][]T) {
	var backtrack func(start int, path []T)
	backtrack = func(start int, path []T) {
		if len(path) > 0 {
			list = append(list, path)
		}
		for i := start; i < len(in); i++ {
			backtrack(i+1, append(path, in[i]))
		}
	}
	backtrack(0, []T{})
	return
}

//----------------------------------------------------------------------

// Combinator yields lists of data assembled from source lists.
//
//	Example: data=[[1,7],[2,5,8],[4,9]] returns
//	[1,2,4],[7,2,4],[1,5,4],[7,5,4],[1,8,4],[7,8,4],
//	[1,2,9],[7,2,9],[1,5,9],[7,5,9],[1,8,9],[7,8,9]
type Combinator[T any] struct {
	data [][]T // source lists
	idx  []int // internal indexing
}

// NewCombinator from a list of index lists.
func NewCombinator[T any](list [][]T) *Combinator[T] {
	return &Combinator[T]{
		data: slices.Clone(list),
		idx:  make([]int, len(list)),
	}
}

// Next returns the next list combination
func (ci *Combinator[T]) Next() (res []T, done bool) {
	for i := range len(ci.idx) {
		res = append(res, ci.data[i][ci.idx[i]])
	}
	for i, v := range ci.idx {
		if v < len(ci.data[i])-1 {
			ci.idx[i] = v + 1
			return
		}
		ci.idx[i] = 0
	}
	done = true
	return
}
