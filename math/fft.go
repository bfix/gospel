/*
 * Fast Fourier Transformation. 
 *
 * (c) 2011-2012 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package math

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"errors"
	"math"
)

///////////////////////////////////////////////////////////////////////
// Public types 

//---------------------------------------------------------------------
/*
 * Transformer type declaration (worker object for FF transformations).
 */
type Transformer struct {
	depth   int          // depth of binary field
	size    int          // helper: 2^depth
	twiddle []complex128 // helper: factor constants [0..size/2[
	scale   complex128   // helper: scale constant
}

//---------------------------------------------------------------------
/*
 * Fields are input/output objects for transformation methods.
 */
type Field []complex128

///////////////////////////////////////////////////////////////////////
// Public methods

//---------------------------------------------------------------------
/*
 * Create a new transformer worker instance.
 * @param n int - handle fields of size 2^n
 * @return *Transformer - new worker instance
 * @return error - error encountered (or nil if successful)
 */
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

//---------------------------------------------------------------------
/*
 * Get field size for transformation worker instance.
 * @this t *Transformer - worker instance for transformation
 * @return int - expected width of field
 */
func (t *Transformer) GetSize() int {
	return t.size
}

//---------------------------------------------------------------------
/*
 * Transform time series into frequency domain.
 * @this t *Transformer - worker instance for transformation
 * @param in Field - input data for transformation
 * @return Field - output data for transformation
 * @return error - processing status (nil = O.K.)
 */
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

//---------------------------------------------------------------------
/*
 * Transform frequemcy series into time domain.
 * @this t *Transformer - worker instance for transformation
 * @param in Field - input data for transformation
 * @return Field - output data for transformation
 * @return error - processing status (nil = O.K.)
 */
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

//---------------------------------------------------------------------
/*
 * Helper method for index computation.
 * @this t *Transformer - worker instance for transformation
 * @param j int - current field index
 * @oaram n int - current sub-field size
 * @return int - associated index
 */
func (t *Transformer) index(j, n int) int {
	a := j / n
	d := 0
	for k := 0; k < t.depth; k++ {
		d = 2*d + (a & 1)
		a >>= 1
	}
	return d
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-11 02:45:03  brf
//  First release as free software (GPL3+)
//
//	Revision 1.2  2010-11-10 21:08:16  brf
//	Corrected allocation and scaling of output field.
//
//	Revision 1.1  2010-11-10 06:50:40  brf
//	Initial revision.
//
