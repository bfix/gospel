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

type Memory struct {
	list  []any
	num   int
	wrpos int
	equal func(any, any) bool
}

func NewMemory(num int, equal func(any, any) bool) *Memory {
	return &Memory{
		list:  make([]any, num),
		num:   num,
		wrpos: 0,
		equal: equal,
	}
}

func (s *Memory) Add(e any) {
	s.list[s.wrpos] = e
	s.wrpos = (s.wrpos + 1) % s.num
}

func (s *Memory) Contains(e any) int {
	for i, v := range s.list {
		if s.equal(v, e) {
			return (s.wrpos - i + s.num) % s.num
		}
	}
	return 0
}
