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

// Vector data structure
type Vector struct {
	data [](interface{}) // list of elements
}

// NewVector instantiates a new (empty) Vector object.
func NewVector() *Vector {
	return &Vector{
		data: make([](interface{}), 0),
	}
}

// Len returns the number of elements in the vector.
func (vec *Vector) Len() int {
	return len(vec.data)
}

// Add element to the end of the vector.
func (vec *Vector) Add(v interface{}) {
	vec.data = append(vec.data, v)
}

// Insert element at given position. Add 'nil' elements if index
// is beyond the end of the vector.
func (vec *Vector) Insert(i int, v interface{}) {

	if i < 0 {
		// create a prepending slice
		pre := make([](interface{}), -i)
		pre[0] = v
		vec.data = append(pre, vec.data...)
	} else if i >= len(vec.data) {
		// create appending slice
		idx := i - len(vec.data) + 1
		app := make([](interface{}), idx)
		app[idx-1] = v
		vec.data = append(vec.data, app...)
	} else {
		pre := vec.data[:i]
		app := vec.data[i:]
		vec.data = append(append(pre, v), app...)
	}
}

// Drop the last element from the vector.
func (vec *Vector) Drop() (v interface{}) {
	pos := len(vec.data) - 1
	v, vec.data = vec.data[pos], vec.data[:pos]
	return
}

// Delete indexed element from the vector.
func (vec *Vector) Delete(i int) (v interface{}) {
	if i < 0 || i > len(vec.data)-1 {
		return nil
	}
	v = vec.data[i]
	vec.data = append(vec.data[:i], vec.data[i+1:]...)
	return
}

// At return the indexed element from vector.
func (vec *Vector) At(i int) (v interface{}) {
	if i < 0 || i > len(vec.data)-1 {
		return nil
	}
	return vec.data[i]
}
