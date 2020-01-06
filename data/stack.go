package data

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

// Stack for generic data types.
type Stack struct {
	data [](interface{}) // list of stack elements
}

// NewStack instantiates a new generic Stack object.
func NewStack() *Stack {
	return &Stack{
		data: make([](interface{}), 0),
	}
}

// Pop last entry from stack and return it to caller.
func (s *Stack) Pop() (v interface{}) {
	pos := len(s.data) - 1
	v, s.data = s.data[pos], s.data[:pos]
	return
}

// Push generic entry to stack.
func (s *Stack) Push(v interface{}) {
	s.data = append(s.data, v)
}

// Len returns the number of elements on stack.
func (s *Stack) Len() int {
	return len(s.data)
}

// Peek at the last element pushed to stack without dropping it.
func (s *Stack) Peek() (v interface{}) {
	pos := len(s.data) - 1
	if pos < 0 {
		return nil
	}
	return s.data[pos]
}

// IntStack is an Integer-based Stack type and implementation.
type IntStack struct {
	data []int // list of stack elements
}

// NewIntStack instantiates a new integer-based Stack object.
func NewIntStack() *IntStack {
	return &IntStack{
		data: make([]int, 0),
	}
}

// Pop last entry from stack and return it to caller.
func (s *IntStack) Pop() (v int) {
	pos := len(s.data) - 1
	v, s.data = s.data[pos], s.data[:pos]
	return
}

// Push entry to stack.
func (s *IntStack) Push(v int) {
	s.data = append(s.data, v)
}

// Len returns the number of elements on stack.
func (s *IntStack) Len() int {
	return len(s.data)
}

// Peek at the last element pushed to stack without dropping it.
func (s *IntStack) Peek() (v int) {
	return s.data[len(s.data)-1]
}

// IsTop compares last element with given value.
func (s *IntStack) IsTop(v int) bool {
	pos := len(s.data) - 1
	if pos < 0 {
		return false
	}
	return s.data[pos] == v
}
