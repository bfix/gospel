package math

import (
	"math"
	"testing"
)

const eps = 1e-9 // precision during compare

// Run test suite for FFT implementation.
func TestTransform(t *testing.T) {
	tf, err := NewTransformer(10)
	if err != nil {
		t.Fatal("failed to create new transformer")
	}
	size := tf.GetSize()
	in := make(Field, size)
	for i := 0; i < size; i++ {
		in[i] = complex(float64(i), 0.)
	}
	out, err := tf.Time2Freq(in)
	if err != nil {
		t.Fatal("failed t2f conversion")
	}
	in2, err := tf.Freq2Time(out)
	if err != nil {
		t.Fatal("failed f2t conversion")
	}
	for i := 0; i < size; i++ {
		if !isEqual(in[i], in2[i]) {
			t.Fatal("data mismatch")
		}
	}
}

// compare two complex numbers for equality.
func isEqual(a, b complex128) bool {
	if math.Abs(real(a)-real(b)) > eps {
		return false
	}
	if math.Abs(imag(a)-imag(b)) > eps {
		return false
	}
	return true
}
