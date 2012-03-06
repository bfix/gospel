
package crypto

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"testing"
	"math"
	"fmt"
)

///////////////////////////////////////////////////////////////////////

const (
	bitsPerBlock	=         8
	numIntervals	=        25
	intervalSize	=    500000
	eTU				= 7.1836656
)

///////////////////////////////////////////////////////////////////////
// Test case for PRNG:
// The test is based on the paper "A Universal Statistical Test for
// Random Bit Generators" by Ueli Maurer, ETHZ 1992
// [ftp://ftp.inf.ethz.ch/pub/crypto/publications/Maurer92a.pdf]

func TestPrng (t *testing.T) {

	fmt.Println ("********************************************")
	fmt.Println ("crypto/prng Test")
	fmt.Println ("********************************************")
	fmt.Println ()
	
	// allocate table
	v := 1 << bitsPerBlock;
	tab := make ([]int, v)
	for i,_ := range tab {	
		tab[i] = 0
	}
		
	// initial table	
	q := 1
	for count := v; count > 0; q++ {
		i := RandInt (0, v-1)
		if tab[i] == 0 {
			count--
		}
		tab[i] = q
	}
	fmt.Printf ("Initializing table required %f over-sampling.\n", float32(q)/float32(v))
	
	// compute statistics
	sum := 0.0
	n := 0
	for i := 0; i < numIntervals; i++ {
		for j := 0; j < intervalSize; j++ {
			i1 := n + q + 1
			i2 := RandInt (0, v-1)
			sum += math.Log (float64(i1 - tab[i2]))
			tab[i2] = i1
			n++
		}
		fTU := sum / (float64(n) * math.Log(2))
		fmt.Printf ("%d: %f (%f)\n", n , fTU, fTU - eTU)
	}
}
