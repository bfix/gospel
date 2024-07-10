package crypto

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

import (
	"math/rand"
	"testing"

	"github.com/bfix/gospel/math"
)

func TestKeys(t *testing.T) {

	n, k := 10, 7
	s := math.NewInt(1234567890123456)

	p := nextPrime(s) // TEST ONLY!!! 'p' should be a random prime >> 's'!!!
	shares := Split(s, p, n, k)

	for kk := 1; kk <= n; kk++ {
		perm := rand.Perm(n)
		coop := make([]Share, kk)
		for i := range coop {
			coop[i] = shares[perm[i]]
		}

		s2 := Reconstruct(coop)

		switch {
		case !s.Equals(s2) && kk >= k:
			t.Fatal("failed reconstruction")
		case s.Equals(s2) && kk < k:
			t.Fatal("pre-mature reconstruction")
		}
	}
}

func nextPrime(p *math.Int) *math.Int {
	// make sure p is odd
	if p.Bit(0) == 0 {
		p = p.Add(math.ONE)
	}
	step := math.TWO
	for {
		p = p.Add(step)
		if p.ProbablyPrime(128) {
			break
		}
	}
	return p
}
