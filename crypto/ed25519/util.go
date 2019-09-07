package ed25519

import (
	"crypto/sha512"

	"github.com/bfix/gospel/math"
)

func reverse(b []byte) []byte {
	bl := len(b)
	r := make([]byte, bl)
	for i := 0; i < bl; i++ {
		r[bl-i-1] = b[i]
	}
	return r
}

func clone(d []byte) []byte {
	r := make([]byte, len(d))
	copy(r, d)
	return r
}

// CopyBlock copies 'in' to 'out' so that 'out' is filled completely.
// - If 'in' is larger than 'out', it is left-truncated before copy
// - If 'in' is smaller than 'out', it is left-padded with 0 before copy
func copyBlock(out, in []byte) {
	count := len(in)
	size := len(out)
	from, to := 0, 0
	if count > size {
		from = count - size
	} else if count < size {
		to = size - count
		for i := 0; i < to; i++ {
			out[i] = 0
		}
	}
	copy(out[to:], in[from:])
}

// h2u hashes successive blocks and converts the resulting SHA512 value
// to integer (from little endian representation)
func h2i(m1, m2, m3 []byte) *math.Int {
	hsh := sha512.New()
	hsh.Write(m1)
	hsh.Write(m2)
	if m3 != nil {
		hsh.Write(m3)
	}
	md := hsh.Sum(nil)
	return math.NewIntFromBytes(reverse(md))
}
