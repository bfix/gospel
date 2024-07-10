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

// Permutation of array elements
type Permutation[T any] struct {
	base []T
	idx  []int
}

// NewPermutation creates a shuffler for an array of elements
func NewPermutation[T any](in []T) *Permutation[T] {
	return &Permutation[T]{
		base: in,
		idx:  make([]int, len(in)),
	}
}

func (p *Permutation[T]) Next() (out []T, done bool) {
	n := len(p.idx)
	for i := n - 1; i >= 0; i-- {
		if i == 0 || p.idx[i] < n-i-1 {
			p.idx[i]++
			break
		}
		p.idx[i] = 0
	}
	out = make([]T, n)
	copy(out, p.base)
	done = (p.idx[0] == n)
	if !done {
		for i, v := range p.idx {
			out[i], out[i+v] = out[i+v], out[i]
		}
	}
	return
}
