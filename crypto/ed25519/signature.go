package ed25519

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"

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

// getBounded returns integer with same bit length as 'n' from binary data.
func getBounded(data []byte) *math.Int {
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
	V, K []byte
}

var (
	b0 = []byte{0x00}
	b1 = []byte{0x01}
)

// init prepares a generator
func (k *kGenDet) init(x *math.Int, h1 []byte) error {
	// enforce 512 bit hash value (SHA512)
	if len(h1) != 64 {
		return ErrSigHashTooSmall
	}

	// initialize hmac'd data
	// data = int2octets(key) || bits2octets(hash)
	data := make([]byte, 64)
	copyBlock(data[:32], x.Bytes())
	h1i := getBounded(h1).Mod(c.N)
	copyBlock(data[32:], h1i.Bytes())

	// initialize K and V
	k.V = bytes.Repeat(b1, 64)
	k.K = bytes.Repeat(b0, 64)

	// start sequence for 'V' and 'K':
	// (1) K = HMAC_K(V || 0x00 || data)
	h := hmac.New(sha512.New, k.K)
	h.Write(k.V)
	h.Write(b0)
	h.Write(data)
	k.K = h.Sum(nil)

	// (2) V = HMAC_K(V)
	h = hmac.New(sha512.New, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

	// (3) K = HMAC_K(V || 0x01 || data)
	h.Reset()
	h.Write(k.V)
	h.Write(b1)
	h.Write(data)
	k.K = h.Sum(nil)

	// (4) V = HMAC_K(V)
	h = hmac.New(sha512.New, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

	return nil
}

// next returns the next 'k'
func (k *kGenDet) next() (*math.Int, error) {

	// (0) V = HMAC_K(V)
	h := hmac.New(sha512.New, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

	// extract 'k' from data
	kRes := getBounded(k.V)

	// (1) K = HMAC_K(V || 0x00
	h.Reset()
	h.Write(k.V)
	h.Write(b0)
	k.K = h.Sum(nil)

	// (2) V = HMAC_K(V)
	h = hmac.New(sha512.New, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

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
	z := getBounded(hv[:])

	// dsa_sign creates a deterministic signature (see RFC6979).
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
			if k.Cmp(c.N) >= 0 {
				continue
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
	return &EcSignature{
		R: r.Mod(c.N),
		S: s.Mod(c.N),
	}, nil
}

// EcVerify checks a EcDSA signature of a message.
func (pub *PublicKey) EcVerify(msg []byte, sig *EcSignature) (bool, error) {
	// Hash message
	hv := sha512.Sum512(msg)
	// compute z
	z := getBounded(hv[:])
	// compute u1, u2
	si := sig.S.ModInverse(c.N)
	u1 := si.Mul(z).Mod(c.N)
	u2 := si.Mul(sig.R).Mod(c.N)
	// compute P = u2 * Q + u1 * G
	P := pub.Q.Mult(u2).Add(c.MultBase(u1))
	// verify signature
	return sig.R.Cmp(P.X().Mod(c.N)) == 0, nil
}
