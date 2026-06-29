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

// Stack for generic data types.
type Stack[T comparable] struct {
	data []T // list of stack elements
}

// NewStack instantiates a new generic Stack object.
func NewStack[T comparable]() *Stack[T] {
	return &Stack[T]{
		data: make([]T, 0),
	}
}

// Pop last entry from stack and return it to caller.
func (s *Stack[T]) Pop() (v T) {
	pos := len(s.data) - 1
	v, s.data = s.data[pos], s.data[:pos]
	return
}

// Push generic entry to stack.
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Len returns the number of elements on stack.
func (s *Stack[T]) Len() int {
	return len(s.data)
}

// Peek at the last element pushed to stack without dropping it.
func (s *Stack[T]) Peek() (v T) {
	pos := len(s.data) - 1
	if pos < 0 {
		var null T
		return null
	}
	return s.data[pos]
}

// IsTop compares the last element with given value.
func (s *Stack[T]) IsTop(v T) bool {
	pos := len(s.data) - 1
	if pos < 0 {
		return false
	}
	return s.data[pos] == v
}
