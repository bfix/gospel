//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2024 Bernd Fix  >Y<
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
	"fmt"
	"testing"
)

func gen(out GeneratorChannel[string]) {
	for i := 0; i < 100; i++ {
		val := fmt.Sprintf("Item #%d", i+1)
		if !out.Yield(val) {
			return
		}
	}
	out.Done()
}

func TestGenerator(t *testing.T) {
	g := NewGenerator(gen)
	for s := range g.Run() {
		t.Logf("got: %s\n", s)
		if s == "Item #23" {
			g.Stop()
			break
		}
	}
}
