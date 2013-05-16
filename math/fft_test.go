package math

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"fmt"
	"math"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Constants

const eps = 1e-9 // precision during compare

///////////////////////////////////////////////////////////////////////
//	Public test method

//---------------------------------------------------------------------
/*
 * Run test suite for FFT implementation.
 * @param test *testing.T - test handler
 */
func TestTransform(test *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("math/fft Test")
	fmt.Println("********************************************************")

	// create worker instance
	t, err := NewTransformer(10)
	if err != nil {
		fmt.Println("Can't create FFT worker instance!")
		test.Fail()
		return
	}

	// display size
	size := t.GetSize()
	fmt.Printf("Size is %d\n", size)

	// allocate and fill input data field
	in := make(Field, size)
	for i := 0; i < size; i++ {
		in[i] = complex(float64(i), 0.)
	}

	// transform time series into frequency domain
	out, err := t.Time2Freq(in)
	if err != nil {
		fmt.Println("Failed transformation into frequency domain!")
		test.Fail()
		return
	}
	/*
		// display intermediate output
		for i := 0; i < size; i++ {
			fmt.Printf ("[%d] %v\n", i+1, out[i])
		}
	*/
	// Re-transform from frequency domain to time series
	in2, err := t.Freq2Time(out)
	if err != nil {
		fmt.Println("Failed transformation into time domain!")
		test.Fail()
		return
	}

	// compare result of both transformation with input field.
	for i := 0; i < size; i++ {
		if !isEqual(in[i], in2[i]) {
			fmt.Printf("[%d] %v != %v\n", i+1, in[i], in2[i])
			fmt.Println("Failed transformations!")
			test.Fail()
			return
		}
	}
}

///////////////////////////////////////////////////////////////////////
// private helper methods

//---------------------------------------------------------------------
/**
 * compare two complex numbers for equality.
 * @param a complex128 - first complex number
 * @param b complex128 - second complex number
 * @return bool - numbers equals within eps range?
 */
//---------------------------------------------------------------------
func isEqual(a, b complex128) bool {

	if math.Abs(real(a)-real(b)) > eps {
		return false
	}
	if math.Abs(imag(a)-imag(b)) > eps {
		return false
	}
	return true
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-11 02:45:03  brf
//  First release as free software (GPL3+)
//
//	Revision 1.3  2010-11-28 14:18:06  brf
//	Added comments.
//
