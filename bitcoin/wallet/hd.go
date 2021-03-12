//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix  >Y<
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

package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/data"
	"github.com/bfix/gospel/math"
)

// Error codes
var (
	ErrHDVersion = errors.New("Version mismatch")
	ErrHDPath    = errors.New("Invalid HD path")
	ErrHDKey     = errors.New("Invalid HD key")
)

//----------------------------------------------------------------------
// ExtendedData objects represent public/private extended keys
//----------------------------------------------------------------------

// extended data version codes
const (
	// Generic versions (P2PKH)
	XpubVersion = 0x0488b21e
	XprvVersion = 0x0488ade4

	// Testnet versions
	UpubVersion = 0x043587cf
	UprvVersion = 0x04358394

	// Generic versions (P2SH)
	YpubVersion = 0x049d7cb2
	YprvVersion = 0x049d7878

	// Dash
	DrkpVersion = 0x02fe52cc

	// Dogecoin
	DgubVersion = 0x02facafd
	DgpvVersion = 0x02fac398

	// Litecoin
	MtubVersion = 0x01b26ef6
	MtpvVersion = 0x01b26792
)

// CheckVersion returns a status code:
//    -1 if extended data refers to a public key
//     1 if extended data refers to a private key
//     0 if version is unknown
func CheckVersion(version uint32) int {
	switch version {
	case XpubVersion, UpubVersion, YpubVersion, DrkpVersion, DgubVersion, MtubVersion:
		return -1
	case XprvVersion, UprvVersion:
		return 1
	}
	return 0
}

// ExtendedData is the data structure representing ExtendedKeys
// (both public and private) for exchange purposes.
type ExtendedData struct {
	Version   uint32 `order:"big"`
	Depth     uint8
	ParentFP  uint32 `order:"big"`
	Child     uint32 `order:"big"`
	Chaincode []byte `size:"32"`
	Keydata   []byte `size:"33"`
}

// NewExtendedData allocates a new extended data object
func NewExtendedData() *ExtendedData {
	return &ExtendedData{
		Chaincode: make([]byte, 32),
		Keydata:   make([]byte, 33),
	}
}

// ParseExtended returns a new data object for a given extended key string
func ParseExtended(s string) (*ExtendedData, error) {
	v, err := bitcoin.Base58Decode(s)
	if err != nil {
		return nil, err
	}
	d := new(ExtendedData)
	if err = data.Unmarshal(d, v); err != nil {
		return nil, err
	}
	return d, nil
}

// String converts an extended data object into a
// human-readable representation.
func (d *ExtendedData) String() string {
	b, err := data.Marshal(d)
	if err != nil {
		return ""
	}
	var r []byte
	r = append(r, b...)
	cs := bitcoin.Hash256(b)
	r = append(r, cs[:4]...)
	return string(bitcoin.Base58Encode(r))
}

//----------------------------------------------------------------------
// Extended public and private keys
//----------------------------------------------------------------------

// ExtendedPublicKey represents a public key in a HD tree
type ExtendedPublicKey struct {
	Data *ExtendedData
	Key  *bitcoin.Point
}

// ParseExtendedPublicKey converts a xpub string to a public key
func ParseExtendedPublicKey(s string) (k *ExtendedPublicKey, err error) {
	k = new(ExtendedPublicKey)
	k.Data, err = ParseExtended(s)
	if err != nil {
		return nil, err
	}
	// check for valid public key version field
	if CheckVersion(k.Data.Version) != -1 {
		return nil, ErrHDVersion
	}
	k.Key, _, err = bitcoin.NewPointFromBytes(k.Data.Keydata)
	if err != nil {
		return nil, err
	}
	return
}

// String returns the string representation of an ExtendedPublicKey
func (e *ExtendedPublicKey) String() string {
	return e.Data.String()
}

// Fingerprint returns the fingerprint of an ExtendedPublicKey
func (e *ExtendedPublicKey) Fingerprint() (i uint32) {
	fp := bitcoin.Hash160(e.Key.Bytes(true))
	rdr := bytes.NewBuffer(fp)
	binary.Read(rdr, binary.BigEndian, &i)
	return
}

// Clone returns a deep copy of a public key
func (e *ExtendedPublicKey) Clone() *ExtendedPublicKey {
	r := new(ExtendedPublicKey)
	r.Key = bitcoin.NewPoint(e.Key.X(), e.Key.Y())
	r.Data = NewExtendedData()
	r.Data.Version = e.Data.Version
	r.Data.Depth = e.Data.Depth
	r.Data.Child = e.Data.Child
	r.Data.ParentFP = e.Data.ParentFP
	copy(r.Data.Chaincode, e.Data.Chaincode)
	copy(r.Data.Keydata, e.Data.Keydata)
	return r
}

// ExtendedPrivateKey represents a private key in a HD tree
type ExtendedPrivateKey struct {
	Data *ExtendedData
	Key  *math.Int
}

// ParseExtendedPrivateKey converts a xprv string to a private key
func ParseExtendedPrivateKey(s string) (k *ExtendedPrivateKey, err error) {
	k = new(ExtendedPrivateKey)
	k.Data, err = ParseExtended(s)
	if err != nil {
		return nil, err
	}
	if CheckVersion(k.Data.Version) != 1 {
		return nil, ErrHDVersion
	}
	k.Key = math.NewIntFromBytes(k.Data.Keydata)
	return k, nil
}

// Public returns the associated public key
func (k *ExtendedPrivateKey) Public() *ExtendedPublicKey {
	r := new(ExtendedPublicKey)
	r.Key = bitcoin.MultBase(k.Key)
	r.Data = NewExtendedData()
	r.Data.Version = XpubVersion
	r.Data.Child = k.Data.Child
	r.Data.Depth = k.Data.Depth
	r.Data.ParentFP = k.Data.ParentFP
	copy(r.Data.Chaincode, k.Data.Chaincode)
	copy(r.Data.Keydata, r.Key.Bytes(true))
	return r
}

// String returns the string representation of an ExtendedPrivateKey
func (k *ExtendedPrivateKey) String() string {
	return k.Data.String()
}

//----------------------------------------------------------------------
// Hierarchically deterministic key space (with private keys available)
//----------------------------------------------------------------------

var (
	c = bitcoin.GetCurve()
)

// HD represents a hierarchically deterministic key space.
type HD struct {
	m *ExtendedPrivateKey
}

// NewHD initializes a new HD from a seed value.
func NewHD(seed []byte) *HD {
	n := len(seed)
	if n < 16 || n > 64 {
		return nil
	}
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	i := mac.Sum(nil)

	mKey := math.NewIntFromBytes(i[:32])
	if mKey.Equals(math.ZERO) || mKey.Cmp(c.N) >= 0 {
		return nil
	}

	hd := new(HD)
	hd.m = new(ExtendedPrivateKey)
	hd.m.Key = mKey
	hd.m.Data = NewExtendedData()
	hd.m.Data.Version = XprvVersion
	copy(hd.m.Data.Keydata, hd.m.Key.FixedBytes(33))
	copy(hd.m.Data.Chaincode, i[32:])
	hd.m.Data.Child = 0
	hd.m.Data.Depth = 0
	hd.m.Data.ParentFP = 0
	return hd
}

// MasterPrivate returns the master private key.
func (hd *HD) MasterPrivate() *ExtendedPrivateKey {
	return hd.m
}

// MasterPublic returns the master public key.
func (hd *HD) MasterPublic() *ExtendedPublicKey {
	return hd.m.Public()
}

// Private returns an extended private key for a given path (BIP32,BIP44)
func (hd *HD) Private(path string) (prv *ExtendedPrivateKey, err error) {
	if !strings.HasPrefix(path, "m/") {
		return nil, ErrHDPath
	}
	prv = hd.m
	for _, id := range strings.Split(path[2:], "/") {
		var (
			j int64
			i uint32
		)
		if strings.HasSuffix(id, "'") {
			j, err = strconv.ParseInt(id[:len(id)-1], 10, 32)
			i = uint32(j) + (1 << 31)
		} else {
			j, err = strconv.ParseInt(id, 10, 32)
			i = uint32(j)
		}
		if err != nil {
			return
		}
		prv = CKDprv(prv, i)
		if prv == nil {
			return nil, ErrHDKey
		}
	}
	return prv, nil
}

// Public returns an extended public key for a given path (BIP32,BIP44)
func (hd *HD) Public(path string) (pub *ExtendedPublicKey, err error) {
	prv, err := hd.Private(path)
	if err != nil {
		return nil, err
	}
	return prv.Public(), nil
}

//----------------------------------------------------------------------
// Hierarchically deterministic key space (public keys only)
//----------------------------------------------------------------------

// HDPublic represents a public branch in a hierarchically deterministic
// key space.
type HDPublic struct {
	m    *ExtendedPublicKey
	path string
}

// NewHDPublic initializes a new HDPublic from an extended public key
// with a given path.
func NewHDPublic(key *ExtendedPublicKey, path string) *HDPublic {
	return &HDPublic{
		m:    key.Clone(),
		path: path,
	}
}

// Public returns an extended public key for a given path. The path MUST
// NOT contain hardened elements and must start with the path of the
// public key in HDPublic!
func (hd *HDPublic) Public(path string) (pub *ExtendedPublicKey, err error) {
	// check for matching relative path
	if !strings.HasPrefix(path, hd.path) {
		return nil, ErrHDPath
	}
	// trim to relative path
	path = path[len(hd.path)+1:]
	// check for hardened levels
	if strings.Index(path, "'") != -1 {
		return nil, ErrHDPath
	}
	// follow the path...
	pub = hd.m
	for _, id := range strings.Split(path, "/") {
		var j int64
		j, err = strconv.ParseInt(id, 10, 32)
		if err != nil {
			return
		}
		pub = CKDpub(pub, uint32(j))
		if pub == nil {
			return nil, ErrHDKey
		}
	}
	return
}

//----------------------------------------------------------------------
// Key derivation methods
//----------------------------------------------------------------------

// CKDprv is a key derivation function for private keys
func CKDprv(k *ExtendedPrivateKey, i uint32) (ki *ExtendedPrivateKey) {
	mac := hmac.New(sha512.New, k.Data.Chaincode)
	if i >= 1<<31 {
		mac.Write([]byte{0})
		mac.Write(k.Key.FixedBytes(32))
	} else {
		p := bitcoin.MultBase(k.Key)
		mac.Write(p.Bytes(true))
	}

	binary.Write(mac, binary.BigEndian, i)
	x := mac.Sum(nil)

	j := math.NewIntFromBytes(x[:32])
	if j.Equals(math.ZERO) || j.Cmp(c.N) >= 0 {
		return nil
	}
	ki = new(ExtendedPrivateKey)
	ki.Key = j.Add(k.Key).Mod(c.N)
	ki.Data = NewExtendedData()
	ki.Data.Version = k.Data.Version
	ki.Data.Depth = k.Data.Depth + 1
	ki.Data.Child = i
	ki.Data.ParentFP = k.Public().Fingerprint()
	copy(ki.Data.Chaincode, x[32:])
	copy(ki.Data.Keydata, ki.Key.FixedBytes(33))
	return
}

// CKDpub is a key derivation function for public keys
func CKDpub(k *ExtendedPublicKey, i uint32) (ki *ExtendedPublicKey) {
	if i >= 1<<31 {
		return nil
	}
	mac := hmac.New(sha512.New, k.Data.Chaincode)
	mac.Write(k.Key.Bytes(true))
	binary.Write(mac, binary.BigEndian, i)
	x := mac.Sum(nil)

	j := math.NewIntFromBytes(x[:32])
	if j.Equals(math.ZERO) || j.Cmp(c.N) >= 0 {
		return nil
	}
	ki = new(ExtendedPublicKey)
	ki.Key = bitcoin.MultBase(j).Add(k.Key)
	ki.Data = NewExtendedData()
	ki.Data.Version = k.Data.Version
	ki.Data.Depth = k.Data.Depth + 1
	ki.Data.Child = i
	ki.Data.ParentFP = k.Fingerprint()
	copy(ki.Data.Chaincode, x[32:])
	copy(ki.Data.Keydata, ki.Key.Bytes(true))
	return
}
