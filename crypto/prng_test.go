package crypto

import (
	"math"
	"testing"
)

const (
	bitsPerBlock = 8
	numIntervals = 25
	intervalSize = 500000
	eTU          = 7.1836656
)

// The test is based on the paper "A Universal Statistical Test for
// Random Bit Generators" by Ueli Maurer, ETHZ 1992
// [ftp://ftp.inf.ethz.ch/pub/crypto/publications/Maurer92a.pdf]
func TestPrng(t *testing.T) {

	// allocate table
	v := 1 << bitsPerBlock
	tab := make([]int, v)
	for i := range tab {
		tab[i] = 0
	}

	// initial table
	q := 1
	for count := v; count > 0; q++ {
		i := RandInt(0, v-1)
		if tab[i] == 0 {
			count--
		}
		tab[i] = q
	}

	// compute statistics
	sum := 0.0
	n := 0
	for i := 0; i < numIntervals; i++ {
		for j := 0; j < intervalSize; j++ {
			i1 := n + q + 1
			i2 := RandInt(0, v-1)
			sum += math.Log(float64(i1 - tab[i2]))
			tab[i2] = i1
			n++
		}
		fTU := sum / (float64(n) * math.Log(2))
		eps := math.Abs(fTU - eTU)
		if eps > 0.002 {
			t.Fatal("random values correlate too much")
		}
	}
}
