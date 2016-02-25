package math

import (
	"math"
	"testing"
)

const eps = 1e-9 // precision during compare

// Run test suite for FFT implementation.
func TestTransform(t *testing.T) {
	t, err := NewTransformer(10)
	if err != nil {
		t.Fatal()
	}
	size := t.GetSize()
	in := make(Field, size)
	for i := 0; i < size; i++ {
		in[i] = complex(float64(i), 0.)
	}
	out, err := t.Time2Freq(in)
	if err != nil {
		t.Fatal()
	}
	in2, err := t.Freq2Time(out)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < size; i++ {
		if !isEqual(in[i], in2[i]) {
			t.Fatal()
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
