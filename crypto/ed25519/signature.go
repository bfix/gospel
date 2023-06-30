//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

package ed25519

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"errors"

	"github.com/bfix/gospel/math"
)

// Error codes for signing / verifying
var (
	ErrSigInvalidPrvKey = errors.New("private key not suitable for EdDSA")
	ErrSigNotEdDSA      = errors.New("not a EdDSA signature")
	ErrSigNotEcDSA      = errors.New("not a EcDSA signature")
	ErrSigInvalidEcDSA  = errors.New("invalid EcDSA signature")
	ErrSigHashTooSmall  = errors.New("hash value to small")
	ErrSigInvalid       = errors.New("invalid signature")
)

//----------------------------------------------------------------------
// EdDSA
//----------------------------------------------------------------------

// EdSignature is a EdDSA signature
type EdSignature struct {
	R *Point
	S *math.Int
}

// NewEdSignatureFromBytes builds a EdDSA signature from its binary
// representation.
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

// Bytes returns the binary representation of an EdDSA signature
func (s *EdSignature) Bytes() []byte {
	buf := make([]byte, 64)
	copy(buf[:32], s.R.Bytes())
	copy(buf[32:], reverse(s.S.Bytes()))
	return buf
}

// EdSign creates an EdDSA signature (R,S) for a message
func (prv *PrivateKey) EdSign(msg []byte) (*EdSignature, error) {
	// hash the private seed and derive R, S
	r := h2i(prv.Nonce, msg, nil)
	R := c.MultBase(r)
	S := r.Add(h2i(R.Bytes(), prv.Public().Bytes(), msg).Mul(prv.D)).Mod(c.N)

	return &EdSignature{R, S}, nil
}

// EdVerify checks an EdDSA signature of a message.
func (pub *PublicKey) EdVerify(msg []byte, sig *EdSignature) (bool, error) {
	h := h2i(sig.R.Bytes(), pub.Bytes(), msg).Mod(c.N)
	tl := c.MultBase(sig.S)
	tr := sig.R.Add(pub.Q.Mult(h))
	return tl.Equals(tr), nil
}

//----------------------------------------------------------------------
// EcDSA (classic or deterministic; see RFC 6979)
//----------------------------------------------------------------------

// EcSignature is a ECDSA signature
type EcSignature struct {
	R *math.Int
	S *math.Int
}

// NewEcSignatureFromBytes creates a ECDSA signature from its binary
// representation.
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

// Bytes returns the binary representation of a ECDSA signature.
func (s *EcSignature) Bytes() []byte {
	buf := make([]byte, 64)
	copyBlock(buf[:32], s.R.Bytes())
	copyBlock(buf[32:], s.S.Bytes())
	return buf
}

// ----------------------------------------------------------------------

// kGenerator is an interface for standard or deterministic computation of the
// blinding factor 'k' in an EcDSA signature. It always uses SHA512 as a
// hashing algorithm.
type kGenerator interface {
	init(x, n *math.Int, h1 []byte) error
	next() (*math.Int, error)
}

// newKGenerator creates a new suitable generator for the binding factor 'k'.
func newKGenerator(det bool, x, n *math.Int, h1 []byte) (gen kGenerator, err error) {
	if det {
		gen = &kGenDet{}
	} else {
		gen = &kGenStd{}
	}
	err = gen.init(x, n, h1)
	return
}

// ----------------------------------------------------------------------
// kGenDet is a RFC6979-compliant generator.
type kGenDet struct {
	V, K []byte
	n    *math.Int
}

var (
	b0 = []byte{0x00}
	b1 = []byte{0x01}
)

// init prepares a generator
func (k *kGenDet) init(x, n *math.Int, h []byte) error {
	// enforce 512 bit hash value (SHA512)
	if len(h) != 64 {
		return ErrSigHashTooSmall
	}

	// initialize hmac'd data
	// data = int2octets(key) || bits2octets(hash)
	data := make([]byte, 64)
	copyBlock(data[:32], x.Bytes())
	h1i := getBounded(h, n)
	copyBlock(data[32:], h1i.Bytes())

	// initialize K and V
	k.V = bytes.Repeat(b1, 64)
	k.K = bytes.Repeat(b0, 64)
	k.n = n

	// start sequence for 'V' and 'K':
	// (1) K = HMAC_K(V || 0x00 || data)
	hsh := hmac.New(sha512.New, k.K)
	hsh.Write(k.V)
	hsh.Write(b0)
	hsh.Write(data)
	k.K = hsh.Sum(nil)

	// (2) V = HMAC_K(V)
	hsh = hmac.New(sha512.New, k.K)
	hsh.Write(k.V)
	k.V = hsh.Sum(nil)

	// (3) K = HMAC_K(V || 0x01 || data)
	hsh.Reset()
	hsh.Write(k.V)
	hsh.Write(b1)
	hsh.Write(data)
	k.K = hsh.Sum(nil)

	// (4) V = HMAC_K(V)
	hsh = hmac.New(sha512.New, k.K)
	hsh.Write(k.V)
	k.V = hsh.Sum(nil)

	return nil
}

// next returns the next 'k'
func (k *kGenDet) next() (*math.Int, error) {

	// (0) V = HMAC_K(V)
	h := hmac.New(sha512.New, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

	// extract 'k' from data
	kRes := getBounded(k.V, k.n)

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

// getBounded returns integer with same bit length as 'n' from binary data.
func getBounded(data []byte, n *math.Int) *math.Int {
	z := math.NewIntFromBytes(data)
	shift := len(data)*8 - n.BitLen()
	if shift > 0 {
		z = z.Rsh(uint(shift))
	}
	return z
}

// ----------------------------------------------------------------------
// kGenStd is a random generator.
type kGenStd struct {
}

// init prepares a generator
func (k *kGenStd) init(x, n *math.Int, h1 []byte) error {
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
	z := getBounded(hv[:], c.N)

	// dsaSign creates a deterministic signature (see RFC6979).
	dsaSign := func(det bool) (r, s *math.Int, err error) {
		zero := math.NewInt(0)
		gen, err := newKGenerator(det, prv.D, c.N, hv[:])
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
	r, s, err := dsaSign(true)
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
	z := getBounded(hv[:], c.N)
	// compute u1, u2
	si := sig.S.ModInverse(c.N)
	if si == nil {
		return false, ErrSigInvalid
	}
	u1 := si.Mul(z).Mod(c.N)
	u2 := si.Mul(sig.R).Mod(c.N)
	// compute P = u2 * Q + u1 * G
	P := pub.Q.Mult(u2).Add(c.MultBase(u1))
	// verify signature
	return sig.R.Cmp(P.X().Mod(c.N)) == 0, nil
}
