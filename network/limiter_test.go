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

package network

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	prt := func(stats *RateStats) {
		t.Log("-----------------------------------")
		t.Log(" Second Minute   Hour    Day   Week")
		t.Logf("%7d%7d%7d%7d%7d\n", stats.pSec, stats.pMin, stats.pHr, stats.pDay, stats.pWeek)
		t.Logf("%7d%7d%7d%7d%7d\n", stats.rSec, stats.rMin, stats.rHr, stats.rDay, stats.rWeek)
		t.Log("-----------------------------------")
	}

	lim := NewRateLimiter(5, 150, 450, 5000, 20000)
	ctx := context.Background()
	for range 100 {
		time.Sleep(100 * time.Millisecond)
		stats := lim.Stats()
		prt(stats)
		lim.Pass(ctx)
	}
}
