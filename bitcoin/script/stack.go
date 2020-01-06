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
	"github.com/bfix/gospel/math"
)

// Stack represents the FIFO stack used during the processing of a script.
// Objects on the stack are of type math.Int; byte arrays and intrinsic
// integers are converted in both way when necessary.
type Stack struct {
	d []*math.Int
}

// NewStack creates a new empty stack.
func NewStack() *Stack {
	return &Stack{
		d: make([]*math.Int, 0),
	}
}

// Len returns the length of the stack
func (s *Stack) Len() int {
	return len(s.d)
}

// Values returns the stack content
func (s *Stack) Values() []*math.Int {
	return s.d
}

// Push an object onto the stack.
// Objects can be of type int, []byte, *math.Int or bool; other types return
// a result code 'RcInvalidStackType'.
func (s *Stack) Push(v interface{}) int {
	var i *math.Int
	switch x := v.(type) {
	case bool:
		if x {
			i = math.NewInt(1)
		} else {
			i = math.NewInt(0)
		}
	case int:
		i = math.NewInt(int64(x))
	case []byte:
		i = math.NewIntFromBytes(x)
	case *math.Int:
		i = x
	default:
		return RcInvalidStackType
	}
	s.d = append(s.d, i)
	return RcOK
}

// Peek looks up the the top-level object on the stack without removing it.
func (s *Stack) Peek() (*math.Int, int) {
	return s.PeekAt(0)
}

// PeekAt looks up the object at depth 'i' of the stack (top-level is depth 0)
// without removing it.
func (s *Stack) PeekAt(i int) (*math.Int, int) {
	x := len(s.d)
	if x < i+1 {
		return nil, RcExceedsStack
	}
	v := s.d[x-1-i]
	return v, RcOK
}

// Pop removes the top-level element from the stack and returns it.
func (s *Stack) Pop() (*math.Int, int) {
	v, rc := s.Peek()
	if rc != RcOK {
		return nil, rc
	}
	if l := len(s.d); l > 1 {
		s.d = s.d[:l-1]
	} else {
		s.d = make([]*math.Int, 0)
	}
	return v, RcOK
}

// RemoveAt removes the element at depth 'i' from the stack (top-level is
// depth 0) and returns it.
func (s *Stack) RemoveAt(i int) (*math.Int, int) {
	x := len(s.d)
	if x < i+1 {
		return nil, RcExceedsStack
	}
	v := s.d[i]
	s.d = append(s.d[:i], s.d[i+1:]...)
	return v, RcOK
}

// Dup duplicates the top n.th elements of the stack.
func (s *Stack) Dup(n int) int {
	for i := 0; i < n; i++ {
		v, rc := s.PeekAt(n - 1)
		if rc != RcOK {
			return rc
		}
		if rc = s.Push(v); rc != RcOK {
			return rc
		}
	}
	return RcOK
}

// Compare compares the two top-level elements of the stack as integers.
func (s *Stack) Compare() (a, b *math.Int, cmp int, rc int) {
	b, rc = s.Pop()
	if rc != RcOK {
		return
	}
	a, rc = s.Pop()
	if rc != RcOK {
		return
	}
	cmp = a.Cmp(b)
	return
}
