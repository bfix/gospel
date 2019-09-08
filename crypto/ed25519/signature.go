package ed25519

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"hash"

	"github.com/bfix/gospel/math"
)

var (
	ErrSigInvalidPrvKey = fmt.Errorf("Private key not suitable for EdDSA")
	ErrSigNotEdDSA      = fmt.Errorf("Not a EdDSA signature")
	ErrSigNotEcDSA      = fmt.Errorf("Not a EcDSA signature")
	ErrSigInvalidEcDSA  = fmt.Errorf("Invalid EcDSA signature")
	ErrSigHashTooSmall  = fmt.Errorf("Hash value to small")
)

//----------------------------------------------------------------------
// EdDSA
//----------------------------------------------------------------------

// EdSignature
type EdSignature struct {
	R *Point
	S *math.Int
}

// NewEdSignatureFromBytes
func NewEdSignatureFromBytes(data []byte) (*EdSignature, error) {
	// check signature size
	if len(data) != 64 {
		return nil, ErrSigNotEdDSA
	}
	// extract R,S
	R, err := NewPointFromBytes(data[:32])
	if err != nil {
		return nil, err
	}
	S := math.NewIntFromBytes(reverse(data[32:]))
	// assemble signature
	return &EdSignature{
		R: R,
		S: S,
	}, nil
}

// Bytes
func (s *EdSignature) Bytes() []byte {
	buf := make([]byte, 64)
	copy(buf[:32], s.R.Bytes())
	copy(buf[32:], reverse(s.S.Bytes()))
	return buf
}

// EdSign creates an EdDSA signature (R,S) for a message
func (prv *PrivateKey) EdSign(msg []byte) (*EdSignature, error) {
	if prv.Seed == nil {
		return nil, ErrSigInvalidPrvKey
	}
	// hash the private seed and derive R, S
	md := sha512.Sum512(prv.Seed)
	r := h2i(md[32:], msg, nil)
	R := c.MultBase(r)
	S := r.Add(h2i(R.Bytes(), prv.Public().Bytes(), msg).Mul(prv.D)).Mod(c.N)

	return &EdSignature{R, S}, nil
}

// Verify checks an EdDSA signature of a message.
func (pub *PublicKey) EdVerify(msg []byte, sig *EdSignature) (bool, error) {
	h := h2i(sig.R.Bytes(), pub.Bytes(), msg).Mod(c.N)
	tl := c.MultBase(sig.S)
	tr := sig.R.Add(pub.Q.Mult(h))
	return tl.Equals(tr), nil
}

//----------------------------------------------------------------------
// EcDSA (classic or deterministic; see RFC 6979)
//----------------------------------------------------------------------

// EcSignature
type EcSignature struct {
	R *math.Int
	S *math.Int
}

// NewEcSignatureFromBytes
func NewEcSignatureFromBytes(data []byte) (*EcSignature, error) {
	// check signature size
	if len(data) != 64 {
		return nil, ErrSigNotEcDSA
	}
	// extract R,S
	R := math.NewIntFromBytes(data[:32])
	S := math.NewIntFromBytes(data[32:])
	// assemble signature
	return &EcSignature{
		R: R,
		S: S,
	}, nil
}

// Bytes
func (s *EcSignature) Bytes() []byte {
	buf := make([]byte, 64)
	copyBlock(buf[:32], s.R.Bytes())
	copyBlock(buf[32:], s.S.Bytes())
	return buf
}

// get_bounded constructs an integer of order 'n' from binary data (message hash).
func get_bounded(data []byte) *math.Int {
	z := math.NewIntFromBytes(data)
	shift := len(data)*8 - c.N.BitLen()
	if shift > 0 {
		z = z.Rsh(uint(shift))
	}
	return z
}

//----------------------------------------------------------------------
// kGenerator is an interface for standard or deterministic computation of the
// blinding factor 'k' in an EcDSA signature. It always uses SHA512 as a
// hashing algorithm.
type kGenerator interface {
	init(x *math.Int, h1 []byte) error
	next() (*math.Int, error)
}

// newKGenerator creates a new suitable generator for the binding factor 'k'.
func newKGenerator(det bool, x *math.Int, h1 []byte) (gen kGenerator, err error) {
	if det {
		gen = &kGenDet{}
	} else {
		gen = &kGenStd{}
	}
	err = gen.init(x, h1)
	return
}

//----------------------------------------------------------------------
// kGenDet is a RFC6979-compliant generator.
type kGenDet struct {
	x    *math.Int
	V, K []byte
	hmac hash.Hash
}

// init prepares a generator
func (k *kGenDet) init(x *math.Int, h1 []byte) error {
	// enforce 512 bit hash value (SHA512)
	if len(h1) != 64 {
		return ErrSigHashTooSmall
	}

	// initialize generator specs
	k.x = x
	k.hmac = hmac.New(sha512.New, x.Bytes())

	// initialize hmac'd data
	// data = int2octets(key) || bits2octets(hash)
	data := make([]byte, 128)
	copyBlock(data[0:64], x.Bytes())
	copyBlock(data[64:128], h1)

	k.V = bytes.Repeat([]byte{0x01}, 64)
	k.K = bytes.Repeat([]byte{0x00}, 64)

	// start sequence for 'V' and 'K':
	// (1) K = HMAC_K(V || 0x00 || data)
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.hmac.Write([]byte{0x00})
	k.hmac.Write(data)
	k.K = k.hmac.Sum(nil)
	// (2) V = HMAC_K(V)
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.V = k.hmac.Sum(nil)
	// (3) K = HMAC_K(V || 0x01 || data)
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.hmac.Write([]byte{0x01})
	k.hmac.Write(data)
	k.K = k.hmac.Sum(nil)
	// (4) V = HMAC_K(V)
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.V = k.hmac.Sum(nil)

	return nil
}

// next returns the next 'k'
func (k *kGenDet) next() (*math.Int, error) {
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.V = k.hmac.Sum(nil)

	// extract 'k' from data
	kRes := get_bounded(k.V)

	// prepare for possible next round
	// (1) K = HMAC_K(V || 0x00
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.hmac.Write([]byte{0x00})
	k.K = k.hmac.Sum(nil)
	// (2) V = HMAC_K(V)
	k.hmac.Reset()
	k.hmac.Write(k.V)
	k.V = k.hmac.Sum(nil)

	return kRes, nil
}

//----------------------------------------------------------------------
// kGenStd is a random generator.
type kGenStd struct {
}

// init prepares a generator
func (k *kGenStd) init(x *math.Int, h1 []byte) error {
	return nil
}

// next returns the next 'k'
func (*kGenStd) next() (*math.Int, error) {
	// generate random k
	return math.NewIntRnd(c.N), nil
}

//----------------------------------------------------------------------
// EcSign creates an EcDSA signature for a message.
func (prv *PrivateKey) EcSign(msg []byte) (*EcSignature, error) {
	// Hash message
	hv := sha512.Sum512(msg)

	// compute z
	z := get_bounded(hv[:])

	// dsa_sign creates a signature. A deterministic signature implements RFC6979.
	dsa_sign := func(det bool) (r, s *math.Int, err error) {
		zero := math.NewInt(0)
		gen, err := newKGenerator(det, prv.D, hv[:])
		if err != nil {
			return nil, nil, err
		}
		for {
			// generate next possible 'k'
			k, err := gen.next()
			if err != nil {
				return nil, nil, err
			}

			// compute r = x-coordinate of point k*G; must be non-zero
			r := c.MultBase(k).X()
			if r.Cmp(zero) == 0 {
				continue
			}

			// compute non-zero s
			ki := k.ModInverse(c.N)
			s := ki.Mul(z.Add(r.Mul(prv.D))).Mod(c.N)

			if s.Cmp(zero) == 0 {
				continue
			}
			return r, s, nil
		}
	}

	// assemble signature
	r, s, err := dsa_sign(true)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 64)
	copyBlock(data[:32], r.Bytes())
	copyBlock(data[32:], s.Bytes())
	return &EcSignature{
		R: r,
		S: s,
	}, nil
}

// EcVerify checks a EcDSA signature of a message.
func (pub *PublicKey) EcVerify(msg []byte, sig *EcSignature) (bool, error) {
	// check r,s values
	if sig.R.Cmp(c.N) != -1 || sig.S.Cmp(c.N) != -1 {
		return false, ErrSigInvalidEcDSA
	}
	// Hash message
	hv := sha512.Sum512(msg)
	// compute z
	z := get_bounded(hv[:])
	// compute u1, u2
	si := sig.S.ModInverse(c.N)
	u1 := si.Mul(z).Mod(c.N)
	u2 := si.Mul(sig.R).Mod(c.N)
	// compute P = u2 * Q + u1 * G
	P := pub.Q.Mult(u2).Add(c.MultBase(u1))
	// verify signature
	return sig.R.Cmp(P.X()) == 0, nil
}
