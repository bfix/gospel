package bitcoin

import (
	"bytes"
	"errors"

	"github.com/bfix/gospel/math"
)

var (
	alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	b58      = math.NewInt(58)
)

// Base58Encode converts byte array to base58 string representation
func Base58Encode(in []byte) string {

	// convert byte array to integer
	val := math.NewIntFromBytes(in)

	// convert integer to base58 representation
	b := []byte{}
	m := math.ZERO
	for val.Cmp(math.ZERO) > 0 {
		val, m = val.DivMod(b58)
		b = append(b, alphabet[int(m.Int64())])
	}
	// handle leading zero bytes in input
	for _, x := range in {
		if x == 0 {
			b = append(b, alphabet[0])
		} else {
			break
		}
	}
	// return base58 representation
	return string(reverse(b))
}

// Base58Decode converts a base58 representation into byte array
func Base58Decode(s string) ([]byte, error) {

	// convert string to byte array
	in := []byte(s)

	// convert base58 to integer (ignores leading zeros)
	val := math.ZERO
	for _, b := range in {
		pos := bytes.IndexByte(alphabet, b)
		if pos == -1 {
			return nil, errors.New("Base58 decoding error -- unknown character")
		}
		val = val.Mul(b58).Add(math.NewInt(int64(pos)))
	}
	// prefix byte array with leading zeros
	pf := []byte{}
	for _, x := range s {
		if byte(x) == alphabet[0] {
			pf = append(pf, 0)
		} else {
			break
		}
	}
	// return final byte array
	return append(pf, val.Bytes()...), nil
}

// reverse byte array
func reverse(in []byte) []byte {
	n := len(in)
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[n-i-1] = in[i]
	}
	return out
}
