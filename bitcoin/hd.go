package bitcoin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"

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
// (MAINNET only!)
//----------------------------------------------------------------------

const (
	xpubVersion = 0x0488b21e
	xprvVersion = 0x0488ade4
)

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

// NewExtendedData alloctates a new extended data object
func NewExtendedData() *ExtendedData {
	return &ExtendedData{
		Chaincode: make([]byte, 32),
		Keydata:   make([]byte, 33),
	}
}

// ParseExtended returns a new data object for a given extended key string
func ParseExtended(s string) (*ExtendedData, error) {
	v, err := Base58Decode(s)
	if err != nil {
		return nil, err
	}
	d := new(ExtendedData)
	if err = data.Unmarshal(d, v); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *ExtendedData) Convert() string {
	b, err := data.Marshal(d)
	if err != nil {
		return ""
	}
	var r []byte
	r = append(r, b...)
	cs := Hash256(b)
	r = append(r, cs[:4]...)
	return string(Base58Encode(r))
}

//----------------------------------------------------------------------
// Extended public and private keys
//----------------------------------------------------------------------

// ExtendedPublicKey represents a public key in a HD tree
type ExtendedPublicKey struct {
	data *ExtendedData
	key  *Point
}

// ParseExtendedPublicKey converts a xpub string to a public key
func ParseExtendedPublicKey(s string) (k *ExtendedPublicKey, err error) {
	k = new(ExtendedPublicKey)
	k.data, err = ParseExtended(s)
	if err != nil {
		return nil, err
	}
	if k.data.Version != xpubVersion {
		return nil, ErrHDVersion
	}
	k.key, _, err = NewPointFromBytes(k.data.Keydata)
	if err != nil {
		return nil, err
	}
	return
}

// String returns the string representation of an ExtendedPublicKey
func (e *ExtendedPublicKey) String() string {
	return e.data.Convert()
}

// Fingerprint returns the fingerprint of an ExtendedPublicKey
func (e *ExtendedPublicKey) Fingerprint() (i uint32) {
	fp := Hash160(e.key.Bytes(true))
	rdr := bytes.NewBuffer(fp)
	binary.Read(rdr, binary.BigEndian, &i)
	return
}

// ExtendedPrivateKey represents a private key in a HD tree
type ExtendedPrivateKey struct {
	data *ExtendedData
	key  *math.Int
}

// ParseExtendedPrivateKey converts a xprv string to a private key
func ParseExtendedPrivateKey(s string) (k *ExtendedPrivateKey, err error) {
	k = new(ExtendedPrivateKey)
	k.data, err = ParseExtended(s)
	if err != nil {
		return nil, err
	}
	if k.data.Version != xprvVersion {
		return nil, ErrHDVersion
	}
	k.key = math.NewIntFromBytes(k.data.Keydata)
	return k, nil
}

// Public returns the associated public key
func (k *ExtendedPrivateKey) Public() *ExtendedPublicKey {
	r := new(ExtendedPublicKey)
	r.key = MultBase(k.key)
	r.data = NewExtendedData()
	r.data.Version = xpubVersion
	r.data.Child = k.data.Child
	r.data.Depth = k.data.Depth
	r.data.ParentFP = k.data.ParentFP
	copy(r.data.Chaincode, k.data.Chaincode)
	copy(r.data.Keydata, r.key.Bytes(true))
	return r
}

// String returns the string representation of an ExtendedPrivateKey
func (e *ExtendedPrivateKey) String() string {
	return e.data.Convert()
}

//----------------------------------------------------------------------
// Hierarchically deterministic key space
//----------------------------------------------------------------------

// HD represents a hierarchically deterministic key space.
type HD struct {
	m *ExtendedPrivateKey
}

// NewHD initializes a new HD from a seed value.
func NewHD(s []byte) *HD {
	n := len(s)
	if n < 16 || n > 64 {
		return nil
	}
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(s)
	i := mac.Sum(nil)

	mKey := math.NewIntFromBytes(i[:32])
	if mKey.Equals(math.ZERO) || mKey.Cmp(c.N) >= 0 {
		return nil
	}

	hd := new(HD)
	hd.m = new(ExtendedPrivateKey)
	hd.m.key = mKey
	hd.m.data = NewExtendedData()
	hd.m.data.Version = xprvVersion
	copy(hd.m.data.Keydata, hd.m.key.FixedBytes(33))
	copy(hd.m.data.Chaincode, i[32:])
	hd.m.data.Child = 0
	hd.m.data.Depth = 0
	hd.m.data.ParentFP = 0
	return hd
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
		prv = hd.CKDprv(prv, i)
		if prv == nil {
			return nil, ErrHDKey
		}
	}
	return prv, nil
}

// Prublic returns an extended public key for a given path (BIP32,BIP44)
func (hd *HD) Public(path string) (pub *ExtendedPublicKey, err error) {
	prv, err := hd.Private(path)
	if err != nil {
		return nil, err
	}
	return prv.Public(), nil
}

// CKDprv is a key derivation function for private keys
func (hd *HD) CKDprv(k *ExtendedPrivateKey, i uint32) (ki *ExtendedPrivateKey) {
	mac := hmac.New(sha512.New, k.data.Chaincode)
	if i >= 1<<31 {
		mac.Write([]byte{0})
		mac.Write(k.key.FixedBytes(32))
	} else {
		p := MultBase(k.key)
		mac.Write(p.Bytes(true))
	}

	binary.Write(mac, binary.BigEndian, i)
	x := mac.Sum(nil)

	j := math.NewIntFromBytes(x[:32])
	if j.Equals(math.ZERO) || j.Cmp(c.N) >= 0 {
		return nil
	}
	ki = new(ExtendedPrivateKey)
	ki.key = j.Add(k.key).Mod(c.N)
	ki.data = NewExtendedData()
	ki.data.Version = k.data.Version
	ki.data.Depth = k.data.Depth + 1
	ki.data.Child = i
	ki.data.ParentFP = k.Public().Fingerprint()
	copy(ki.data.Chaincode, x[32:])
	copy(ki.data.Keydata, ki.key.FixedBytes(33))
	return
}

// CKDpub is a key derivation function for public keys
func (hd *HD) CKDpub(k *ExtendedPublicKey, i uint32) (ki *ExtendedPublicKey) {
	if i >= 1<<31 {
		return nil
	}
	mac := hmac.New(sha512.New, k.data.Chaincode)
	mac.Write(k.key.Bytes(true))
	binary.Write(mac, binary.BigEndian, i)
	x := mac.Sum(nil)

	j := math.NewIntFromBytes(x[:32])
	if j.Equals(math.ZERO) || j.Cmp(c.N) >= 0 {
		return nil
	}
	ki = new(ExtendedPublicKey)
	ki.key = MultBase(j).Add(k.key)
	ki.data = NewExtendedData()
	ki.data.Version = k.data.Version
	ki.data.Depth = k.data.Depth + 1
	ki.data.Child = i
	ki.data.ParentFP = k.Fingerprint()
	copy(ki.data.Chaincode, x[32:])
	copy(ki.data.Keydata, ki.key.Bytes(true))
	return
}
