package math

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
	"errors"
	"math"
)

// Transformer type declaration (worker object for FF transformations).
type Transformer struct {
	depth   int          // depth of binary field
	size    int          // helper: 2^depth
	twiddle []complex128 // helper: factor constants [0..size/2[
	scale   complex128   // helper: scale constant
}

// Field instances are input/output objects for transformation methods.
type Field []complex128

// NewTransformer creates a new transformer worker instance.
func NewTransformer(n int) (*Transformer, error) {

	// check for valid argument
	if n < 1 {
		return nil, errors.New("NewTransformer - invalid exponent")
	}

	// allocate new worker instance and preset scalars
	t := new(Transformer)
	t.depth = n
	t.size = 1 << uint(n)
	t.scale = complex(math.Sqrt(float64(t.size)), 0)

	// precompute constants
	p0 := 2 * math.Pi / float64(n)
	t.twiddle = make([]complex128, t.size/2)
	for k := 0; k < t.size/2; k++ {
		x := float64(k) * p0
		t.twiddle[k] = complex(math.Cos(x), -math.Sin(x))
	}
	// return worker instance
	return t, nil
}

// GetSize returns the field size for a transformation worker instance.
func (t *Transformer) GetSize() int {
	return t.size
}

// Time2Freq transforms a time series into the frequency domain.
func (t *Transformer) Time2Freq(in Field) (Field, error) {

	// check for matching array length
	if len(in) != t.size {
		return nil, errors.New("Time2Freq: invalid input field size")
	}

	// preset output with input
	out := make(Field, t.size)
	copy(out, in)

	// perform reduction
	n := t.size / 2
	for i := 0; i < t.depth; i++ {
		for j := 0; j < t.size; j++ {
			k := t.index(j, n)
			z := t.twiddle[k] * out[n+j]
			out[n+j] = out[j] - z
			out[j] = out[j] + z
			if (n + j + 1) == (j/n+2)*n {
				j += n
			}
		}
		n >>= 1
	}
	// re-order array
	for j := 0; j < t.size; j++ {
		k := t.index(j, 1)
		if k > j {
			z := out[j]
			out[j] = out[k]
			out[k] = z
		}
		// scale down transformed values.
		out[j] = out[j] / t.scale
	}
	// return result
	return out, nil
}

// Freq2Time transforms a frequency series into the time domain.
func (t *Transformer) Freq2Time(in Field) (Field, error) {

	// check for matching array length
	if len(in) != t.size {
		return nil, errors.New("Freq2Time: invalid input field size")
	}

	// pre-set output with input
	out := make(Field, t.size)
	copy(out, in)

	// re-order input array
	for j := 0; j < t.size; j++ {
		k := t.index(j, 1)
		if k <= j {
			continue
		}
		z := out[j]
		out[j] = out[k]
		out[k] = z
	}
	// perform composition
	n := 1
	for i := 0; i < t.depth; i++ {
		for j := 0; j < t.size; j++ {
			k := t.index(j, n)
			out[j] = (out[n+j] + out[j]) / 2.
			out[n+j] = (out[j] - out[n+j]) / t.twiddle[k]
			if (n + j + 1) == (j/n+2)*n {
				j += n
			}
		}
		n <<= 1
	}
	// re-scale array
	for j := 0; j < t.size; j++ {
		out[j] *= t.scale
	}
	// return result
	return out, nil
}

// Helper method for index computation.
func (t *Transformer) index(j, n int) int {
	a := j / n
	d := 0
	for k := 0; k < t.depth; k++ {
		d = 2*d + (a & 1)
		a >>= 1
	}
	return d
}
