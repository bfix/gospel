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

package concurrent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/bfix/gospel/math"
	"golang.org/x/crypto/scrypt"
)

type TestDispatchable struct {
	busy  atomic.Int32
	best  atomic.Int32
	check func(i int64) (int32, []byte)
}

func NewTestDispatchable() *TestDispatchable {
	d := new(TestDispatchable)
	d.best.Store(257)
	d.busy.Store(0)

	d.check = func(i int64) (int32, []byte) {
		pp := fmt.Appendf(nil, "%d", i)
		buf, _ := scrypt.Key(pp, []byte("test"), 65536, 8, 1, 32)
		h := sha256.Sum256(buf)
		v := math.NewIntFromBytes(h[:])
		return int32(v.BitLen()), h[:]
	}
	return d
}

func (d *TestDispatchable) Worker(ctx context.Context, n int, taskCh chan int64, resCh chan int64) {
	for {
		select {
		case <-ctx.Done():
			return

		case i := <-taskCh:
			d.busy.Add(1)
			j, _ := d.check(i)
			if j < d.best.Load() {
				d.best.Store(j)
				resCh <- i
			}
			d.busy.Add(-1)
		}
	}
}

func (d *TestDispatchable) Eval(result int64) bool {
	j, h := d.check(result)
	fmt.Printf("got: %d -- [%d] %s\n", result, j, hex.EncodeToString(h))
	return j < 250
}

func (d *TestDispatchable) Busy() int {
	return int(d.busy.Load())
}

func TestWorker(t *testing.T) {

	// run dispatcher
	ctx, cancel := context.WithCancel(context.Background())
	d := NewDispatcher[int64, int64](ctx, 8, NewTestDispatchable())
	defer cancel()

	// process tasks until finished
	var i int64
	for i = 0; ; i++ {
		if !d.Process(i) {
			break
		}
	}
}
