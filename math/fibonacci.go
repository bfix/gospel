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

package math

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"

	"github.com/bfix/gospel/data"
)

// Knacci generates a k-step Fibonacci sequence in a cyclic group C_n with
// elements S_i. Each S_i is a list of factors S_{i,j} (0 <= j < k). The
// k-nacci result v_i is computed from the initial values a_j as:
// v_i = Sum_{j=1}^{k}(S_{i,j}*a_j)
type Knacci struct {
	n      *Int     // group order: if set, values are considered modulo N
	k      int64    // k-nacci step value
	state  [][]*Int // current state
	fac    []*Int   // current factors
	next   []*Int   // next factors
	pos    int64    // position of next insertion
	step   int64    // step counter
	offset *Int     // step offset
}

// NewKnacci instantiates a new k-nacci on the group modulo N
func NewKnacci(k int, N *Int) (kn *Knacci) {
	kn = new(Knacci)
	kn.n = N
	kn.k = int64(k)
	kn.state = make([][]*Int, kn.k)
	kn.next = make([]*Int, kn.k)
	kn.Reset()
	return
}

// Reset instance to initial conditions.
func (kn *Knacci) Reset() {
	for i := range kn.k {
		kn.next[i] = ONE
		kn.state[i] = make([]*Int, kn.k)
		for j := range kn.k {
			kn.state[i][j] = ZERO
			if i == j {
				kn.state[i][j] = ONE
			}
		}
	}
	kn.fac = slices.Clone(kn.next)
	kn.pos = 0
	kn.step = int64(kn.k + 1)
	kn.offset = ZERO
}

// Next returns the next position and factor list
func (kn *Knacci) Next() (step int64, f []*Int) {
	step, f = kn.step, kn.next
	kn.fac = slices.Clone(f)
	for i, v := range f {
		kn.next[i] = kn.next[i].Sub(kn.state[i][kn.pos]).Add(v).Mod(kn.n)
		kn.state[i][kn.pos] = v
	}
	kn.pos = (kn.pos + 1) % int64(kn.k)
	kn.step++
	f = kn.fac
	return
}

// Factors returns the factor list at given position. If the
// position is zero, the current factor list is returned.
func (kn *Knacci) Factors(n int64) []*Int {
	if n > 0 {
		k := int64(kn.k)
		if n < k {
			return nil
		}
		if n > 0 {
			kn.Reset()
			for range n - k {
				kn.Next()
			}
		}
	}
	return kn.fac
}

// Solve the equation for the first initial value
func (kn *Knacci) Solve(f1, f2, init []*Int) *Int {
	fd := f1[0].Sub(f2[0]).ModInverse(kn.n)
	var x = ZERO
	var j int64
	for j = 1; j < kn.k; j++ {
		x = x.Add(f2[j].Sub(f1[j]).Mul(init[j-1]).Mul(fd)).Mod(kn.n)
	}
	return x
}

// Kcheck is a callback to check if a sequence element is recurring.
// It is the task of the calback to compute the element from the
// list of factors and the initial values.
type Kcheck func(f []*Int, pos int64) int64

// Recurrence searches for a recurrence in the Fibonacci sequence.
// If successful, it returns the positions of the matching element
// (p1,p2) and the factor list at p2.
func (kn *Knacci) Recurrence(depth int64, check Kcheck) (f []*Int, p1, p2 int64) {
	kn.Reset()
	for p2 < depth {
		p2, f = kn.Next()
		if p1 = check(f, p2); p1 > 0 {
			return
		}
	}
	p1 = -1
	f = nil
	return
}

// Steps returns the real number of steps
func (kn *Knacci) Steps() *Int {
	return kn.offset.Add(NewInt(kn.step))
}

// ReadKnacci creates a Knacci from data in a file
func ReadKnacci(rdr io.Reader) (kn *Knacci, err error) {
	buf := make([]byte, 1024)
	readInt := func() (n *Int, err error) {
		if _, err = rdr.Read(buf[:1]); err != nil {
			return
		}
		s := buf[0]
		if _, err = rdr.Read(buf[:s]); err != nil {
			return
		}
		n = NewIntFromBytes(buf[:s])
		return
	}
	readUint64 := func() (n int64, err error) {
		if _, err = rdr.Read(buf[:8]); err != nil {
			return
		}
		n = int64(binary.BigEndian.Uint64(buf[:8]))
		return
	}

	var N *Int
	if N, err = readInt(); err != nil {
		return
	}
	var k int64
	if k, err = readUint64(); err != nil {
		return
	}
	kn = NewKnacci(int(k), N)
	if kn.offset, err = readInt(); err != nil {
		return
	}
	kn.step = 0
	if kn.pos, err = readUint64(); err != nil {
		return
	}
	for i := range kn.k {
		kn.next[i] = ZERO
		for j := range kn.k {
			if kn.state[i][j], err = readInt(); err != nil {
				return
			}
			kn.next[i] = kn.next[i].Add(kn.state[i][j]).Mod(kn.n)
		}
	}
	kn.fac = slices.Clone(kn.next)
	return
}

// Write Knacci instance to writer
func (kn *Knacci) Write(wrt io.Writer) (err error) {
	buf := make([]byte, 1024)
	writeInt := func(n *Int) (err error) {
		d := n.Bytes()
		buf[0] = byte(len(d))
		copy(buf[1:], d)
		_, err = wrt.Write(buf[:len(d)+1])
		return
	}
	writeUint64 := func(n int64) (err error) {
		binary.BigEndian.PutUint64(buf[:8], uint64(n))
		_, err = wrt.Write(buf[:8])
		return
	}

	if err = writeInt(kn.n); err != nil {
		return
	}
	if err = writeUint64(kn.k); err != nil {
		return
	}
	offset := kn.offset.Add(NewInt(kn.step))
	if err = writeInt(offset); err != nil {
		return
	}
	if err = writeUint64(kn.pos); err != nil {
		return
	}
	for i := range kn.k {
		for j := range kn.k {
			if err = writeInt(kn.state[i][j]); err != nil {
				return
			}
		}
	}
	return
}

//----------------------------------------------------------------------

// KnacciInt is a k-nacci integer sequence over a cyclic group C_n.
type KnacciInt struct {
	*Knacci

	init []*Int // initial values
}

// NewKnacciInt instantiates a new integer-based k-nacci.
func NewKnacciInt(N *Int, init ...*Int) *KnacciInt {
	k := len(init)
	kn := &KnacciInt{
		Knacci: NewKnacci(k, N),
		init:   slices.Clone(init),
	}
	return kn
}

// Solve the equation for the first initial value
func (kn *KnacciInt) Solve(f1, f2 []*Int) *Int {
	return kn.Knacci.Solve(f1, f2, kn.init[1:])
}

// compute the current value of the sequence.
func (kn *KnacciInt) Value(f []*Int) (n *Int) {
	n = ZERO
	for i, v := range f {
		n = n.Add(kn.init[i].Mul(v)).Mod(kn.n)
	}
	return
}

// Recurrence detects a recurring value.
// The memory only spans a limit timeframe (1e6 steps), so recurrences could
// possibly be missed.
func (kn *KnacciInt) Recurrence(depth int64, descr string) (f []*Int, p1, p2 int64) {
	seen := data.NewMemory(1e6, func(e1, e2 any) bool {
		return e1.(*Int).Equals(e2.(*Int))
	})
	check := func(f []*Int, pos int64) int64 {
		if len(descr) > 0 {
			fmt.Printf("%10d/%s\r", pos, descr)
		}
		x := kn.Value(f)
		i := int64(seen.Contains(x))
		if i > 0 {
			return pos - i
		}
		seen.Add(x)
		return 0
	}
	return kn.Knacci.Recurrence(depth, check)
}

//----------------------------------------------------------------------

type Point interface {
	Equals(Point) bool
}

type Curve interface {
	N() *Int
	G() Point
	Inf() Point

	Add(p1, p2 Point) Point
	Mult(k *Int, p Point) Point
}

// KnacciECC is a k-nacci sequence of points on an elliptic
type KnacciECC struct {
	*Knacci

	c    Curve   // reference to the curve instance
	d    Point   // point with unknown scalar
	r    []*Int  // scalars of random points
	pnts []Point // list of points
}

// KnacciECC creates a k-nacci sequence over an elliptic
func NewKnacciECC(c Curve, d Point, r []*Int) *KnacciECC {
	k := len(r) + 1
	kn := &KnacciECC{
		Knacci: NewKnacci(k, c.N()),
		c:      c,
		d:      d,
		r:      slices.Clone(r),
		pnts:   make([]Point, k),
	}
	g := c.G()
	kn.pnts[0] = d
	for i, v := range r {
		kn.pnts[i+1] = c.Mult(v, g)
	}
	return kn
}

// Solve the equation for the scalar of the first initial point.
func (kn *KnacciECC) Solve(f1, f2 []*Int) *Int {
	return kn.Knacci.Solve(f1, f2, kn.r)
}

// compute point on E(p)
func (kn *KnacciECC) Value(f []*Int) (n Point) {
	n = kn.c.Inf()
	for i, v := range f {
		n = kn.c.Add(n, kn.c.Mult(v, kn.pnts[i]))
	}
	return
}

// Recurrence detects a recurring point.
// The memory only spans a limit timeframe (1e6 steps), so recurrences could
// possibly be missed.
func (kn *KnacciECC) Recurrence(depth int64, descr string) (f []*Int, p1, p2 int64) {
	seen := data.NewMemory(1e6, func(e1, e2 any) bool {
		return e1.(Point).Equals(e2.(Point))
	})
	check := func(f []*Int, pos int64) int64 {
		fmt.Printf("%10d/%s\r", pos, descr)
		x := kn.Value(f)
		i := int64(seen.Contains(x))
		if i > 0 {
			return pos - i
		}
		seen.Add(x)
		return 0
	}
	return kn.Knacci.Recurrence(depth, check)
}
