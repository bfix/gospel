package ecc

import (
	"github.com/bfix/gospel/math"
)

// Signature is a Bitcoin signature in scripts.
type Signature struct {
	R, S *math.Int
}

// Sign a hash value with private key.
// [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 13f]
func Sign(key *PrivateKey, hash []byte) *Signature {
	sig := new(Signature)
	var k, invK *math.Int
	for {
		// compute value of 'r' as x-coordinate of k*G with random k
		for {
			// get random value
			k = nRnd(math.THREE)
			// get its modular inverse
			invK = nInv(k)

			// compute k*G
			pnt := MultBase(k)
			sig.R = nMod(pnt.x)
			if sig.R.Sign() != 0 {
				break
			}
		}
		// compute value of 's := (rd + e)/k'
		e := convertHash(hash)
		sig.S = nMul(nMul(key.D, sig.R).Add(e), invK)
		if sig.S.Sign() != 0 {
			break
		}
	}
	return sig
}

// Verify a hash value with public key.
// [http://www.nsa.gov/ia/_files/ecdsa.pdf, page 15f]
func Verify(key *PublicKey, hash []byte, sig *Signature) bool {

	// sanity checks for arguments
	if sig.R.Sign() == 0 || sig.S.Sign() == 0 {
		return false
	}
	if sig.R.Cmp(c.N) >= 0 || sig.S.Cmp(c.N) >= 0 {
		return false
	}
	// check signature
	e := convertHash(hash)
	w := nInv(sig.S)

	u1 := e.Mul(w)
	u2 := w.Mul(sig.R)

	p1 := MultBase(u1)
	p2 := key.Q.Mult(u2)
	if p1.x.Cmp(p2.x) == 0 {
		return false
	}
	p3 := p1.Add(p2)
	rr := nMod(p3.x)
	return rr.Cmp(sig.R) == 0
}

// convert hash value to integer
// [http://www.secg.org/download/aid-780/sec1-v2.pdf]
func convertHash(hash []byte) *math.Int {
	// trim hash value (if required)
	maxSize := (c.N.BitLen() + 7) / 8
	if len(hash) > maxSize {
		hash = hash[:maxSize]
	}
	// convert to integer
	return math.NewIntFromBytes(hash).Rsh(uint(maxSize*8 - c.N.BitLen()))
}
