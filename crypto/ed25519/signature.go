//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
	"hash"

	"github.com/bfix/gospel/math"
)

// Error codes for signing / verifying
var (
	ErrSigInvalidPrvKey    = errors.New("private key not suitable for EdDSA")
	ErrSigNotEdDSA         = errors.New("not a EdDSA signature")
	ErrSigNotEcDSA         = errors.New("not a EcDSA signature")
	ErrSigInvalidEcDSA     = errors.New("invalid EcDSA signature")
	ErrSigHashSizeMismatch = errors.New("hash size mismatch")
	ErrSigInvalid          = errors.New("invalid signature")
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
	init(x, n *math.Int, size int, h []byte, hshNew func() hash.Hash) error
	next() (*math.Int, error)
}

// newKGenerator creates a new suitable generator for the binding factor 'k'.
func newKGenerator(det bool, x, n *math.Int, size int, h []byte, hsh func() hash.Hash) (gen kGenerator, err error) {
	if det {
		gen = &kGenDet{}
	} else {
		gen = &kGenStd{}
	}
	err = gen.init(x, n, size, h, hsh)
	return
}

// ----------------------------------------------------------------------
// kGenDet is a RFC6979-compliant generator.
type kGenDet struct {
	V, K   []byte
	n      *math.Int
	size   int
	hshNew func() hash.Hash
}

var (
	b0 = []byte{0x00}
	b1 = []byte{0x01}
)

// init prepares a generator
func (k *kGenDet) init(x, n *math.Int, size int, h []byte, hshNew func() hash.Hash) error {
	// enforce correct hash value size
	hshSize := hshNew().Size()
	if len(h) != hshSize {
		return ErrSigHashSizeMismatch
	}

	// initialize hmac'd data
	// data = int2octets(key) || bits2octets(hash)
	data := make([]byte, 2*size)
	copyBlock(data[:size], x.Bytes())
	hi := getBounded(h, n).Mod(n)
	copyBlock(data[size:], hi.Bytes())

	// initialize K and V
	k.V = bytes.Repeat(b1, hshSize)
	k.K = bytes.Repeat(b0, hshSize)
	k.n = n
	k.size = size
	k.hshNew = hshNew

	// start sequence for 'V' and 'K':
	// (1) K = HMAC_K(V || 0x00 || data)
	hsh := hmac.New(hshNew, k.K)
	hsh.Write(k.V)
	hsh.Write(b0)
	hsh.Write(data)
	k.K = hsh.Sum(nil)

	// (2) V = HMAC_K(V)
	hsh = hmac.New(hshNew, k.K)
	hsh.Write(k.V)
	k.V = hsh.Sum(nil)

	// (3) K = HMAC_K(V || 0x01 || data)
	hsh.Reset()
	hsh.Write(k.V)
	hsh.Write(b1)
	hsh.Write(data)
	k.K = hsh.Sum(nil)

	// (4) V = HMAC_K(V)
	hsh = hmac.New(hshNew, k.K)
	hsh.Write(k.V)
	k.V = hsh.Sum(nil)

	return nil
}

// next returns the next 'k'
func (k *kGenDet) next() (*math.Int, error) {

	t := new(bytes.Buffer)
	for t.Len()*8 < k.n.BitLen() {
		// (0) V = HMAC_K(V)
		h := hmac.New(k.hshNew, k.K)
		h.Write(k.V)
		k.V = h.Sum(nil)
		t.Write(k.V)
	}

	// extract 'k' from data
	kRes := getBounded(t.Bytes(), k.n)

	// (1) K = HMAC_K(V || 0x00)
	h := hmac.New(k.hshNew, k.K)
	h.Write(k.V)
	h.Write(b0)
	k.K = h.Sum(nil)

	// (2) V = HMAC_K(V)
	h = hmac.New(k.hshNew, k.K)
	h.Write(k.V)
	k.V = h.Sum(nil)

	return kRes, nil
}

// getBounded takes as input a sequence of 'len(data)*8' bits and
// outputs a non-negative integer that is less than 2^(bitLen(n)).
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
func (k *kGenStd) init(x, n *math.Int, size int, h []byte, hshNew func() hash.Hash) error {
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
		gen, err := newKGenerator(det, prv.D, c.N, 32, hv[:], sha512.New)
		if err != nil {
			return nil, nil, err
		}
		for {
			// generate next possible 'k'
			var k *math.Int
			if k, err = gen.next(); err != nil {
				return
			}
			// check 'k' within range [1,c.N-1]
			if k.Equals(math.ZERO) || k.Cmp(c.N) >= 0 {
				continue
			}
			// compute r = x-coordinate of point k*G; must be non-zero
			if r = c.MultBase(k).X().Mod(c.N); r.Equals(math.ZERO) {
				continue
			}
			// compute non-zero s
			ki := k.ModInverse(c.N)
			if s = ki.Mul(z.Add(r.Mul(prv.D))).Mod(c.N); s.Equals(math.ZERO) {
				continue
			}
			return
		}
	}
	// assemble signature
	r, s, err := dsaSign(true)
	if err != nil {
		return nil, err
	}
	return &EcSignature{
		R: r,
		S: s,
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
