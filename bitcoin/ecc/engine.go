package ecc

import (
	"github.com/bfix/gospel/math"
)

// Sign a hash value with private key.
// [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 13f]
func Sign(key *PrivateKey, hash []byte) (r, s *math.Int) {

	var k, invK *math.Int
	for {
		// compute value of 'r' as x-coordinate of k*G with random k
		for {
			// get random value
			k = nRnd(math.THREE)
			// get its modular inverse
			invK = nInv(k)

			// compute k*G
			pnt := ScalarMultBase(k)
			r = nMod(pnt.x)
			if r.Sign() != 0 {
				break
			}
		}
		// compute value of 's := (rd + e)/k'
		e := convertHash(hash)
		s = nMul(nMul(key.D, r).Add(e), invK)
		if s.Sign() != 0 {
			break
		}
	}
	return
}

// Verify a hash value with public key.
// [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 15f]
func Verify(key *PublicKey, hash []byte, r, s *math.Int) bool {

	// sanity checks for arguments
	if r.Sign() == 0 || s.Sign() == 0 {
		return false
	}
	if r.Cmp(curveN) >= 0 || s.Cmp(curveN) >= 0 {
		return false
	}
	// check signature
	e := convertHash(hash)
	w := nInv(s)

	u1 := e.Mul(w)
	u2 := w.Mul(r)

	p1 := ScalarMultBase(u1)
	p2 := scalarMult(key.Q, u2)
	if p1.x.Cmp(p2.x) == 0 {
		return false
	}
	p3 := add(p1, p2)
	rr := nMod(p3.x)
	return rr.Cmp(r) == 0
}

// convert hash value to integer
// [http://www.secg.org/download/aid-780/sec1-v2.pdf]
func convertHash(hash []byte) *math.Int {
	// trim hash value (if required)
	maxSize := (curveN.BitLen() + 7) / 8
	if len(hash) > maxSize {
		hash = hash[:maxSize]
	}
	// convert to integer
	return math.NewIntFromBytes(hash).Rsh(uint(maxSize*8 - curveN.BitLen()))
}
