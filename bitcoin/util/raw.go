package util

import (
	"encoding/hex"
	"errors"
)

// DissectedTransaction is dissected raw transaction for easier manipulation.
type DissectedTransaction struct {
	Version  uint64   // version number of the transaction
	VinSlot  int      // active vin slot index
	VinSeq   []uint64 // list of sequence numbers of vin slots
	LockTime uint64   // locktime of the transaction
	Content  [][]byte // list of transaction segments
	Signable []byte   // signable transaction
}

// NewDissectedTransaction dissects a raw transaction into its defining
// segments. A binary encoded raw transaction looks like this:
//
//  ------+-------------+---------+---------------------------------
//  Level | Field       | length  | Value/Comment
//  ------+-------------+---------+---------------------------------
//      0 | VERSION     |    4    | version
//      0 | N_VIN       |  var    | Number of Vin defs to follow
//  ------+-------------+---------+---------------------------------
//      1 | VIN:TXID    |   32    | Transaction ID
//      1 | VIN:N       |    4    | VOUT number in transaction
//      1 | VIN:SCRIPT  |         | scriptSig (hex-encoded)
//  ------+-------------+---------+---------------------------------
//      2 | SCRIPT:LEN  |   var   | Length of script
//      2 | SCRIPT:DATA |   <n>   | Script data
//  ------+-------------+---------+---------------------------------
//      1 | VIN:SEQ     |         | sequence number
//  ------+-------------+---------+---------------------------------
//      0 | N_VOUT      |  var    | Number of Vout defs to follow
//  ------+-------------+---------+---------------------------------
//      1 | VOUT:AMOUNT |    4    | Number of Satoshis (1e-8 btc)
//      1 | VOUT:INDEX  |    4    | Output index
//      1 | VOUT:SCRIPT |         | scriptPubkey
//  ------+-------------+---------+---------------------------------
//      2 | SCRIPT:LEN  |   var   | Length of script
//      2 | SCRIPT:DATA |   <n>   | Script data
//  ------+-------------+---------+---------------------------------
//      0 | LOCKTIME    |    4    | locktime
//  ------+-------------+---------+---------------------------------
//
func NewDissectedTransaction(rawHex string) (dt *DissectedTransaction, err error) {
	var buf []byte
	if buf, err = hex.DecodeString(rawHex); err != nil {
		return nil, err
	}
	dt = &DissectedTransaction{
		Version:  0,
		VinSeq:   make([]uint64, 0),
		LockTime: 0,
		Content:  make([][]byte, 0),
		VinSlot:  -1,
		Signable: nil,
	}
	pos := 0
	add := func(n int) uint64 {
		c := make([]byte, n)
		for i := 0; i < n; i++ {
			c[i] = buf[pos+i]
		}
		pos += n
		dt.Content = append(dt.Content, c)
		if n > 4 {
			return 0
		}
		v, _ := GetUint(c, 0, n)
		return v
	}
	dt.Version = add(4)               // version
	n, j, err := GetVarUint(buf, pos) // number of inputs
	if err != nil {
		return nil, err
	}
	dt.Content = append(dt.Content, buf[pos:pos+j])
	pos += j
	for i := 0; i < int(n); i++ {
		add(32)                           // input address hash
		add(4)                            // input index
		s, j, err := GetVarUint(buf, pos) // size of script
		if err != nil {
			return nil, err
		}
		dt.Content = append(dt.Content, buf[pos:pos+j])
		pos += j
		if s == 0 {
			dt.Content = append(dt.Content, []byte{})
		} else {
			add(int(s)) // script code
		}
		seq := add(4) // sequence
		dt.VinSeq = append(dt.VinSeq, seq)
	}
	n, j, err = GetVarUint(buf, pos) // number of outputs
	if err != nil {
		return nil, err
	}
	dt.Content = append(dt.Content, buf[pos:pos+j])
	pos += j
	for i := 0; i < int(n); i++ {
		add(4)                            // amount
		add(4)                            // output index
		s, j, err := GetVarUint(buf, pos) // size of script
		if err != nil {
			return nil, err
		}
		dt.Content = append(dt.Content, buf[pos:pos+j])
		pos += j
		if s == 0 {
			dt.Content = append(dt.Content, []byte{})
		} else {
			add(int(s)) // script code
		}
	}
	dt.LockTime = add(4) // locktime
	return
}

// Bytes returns a flat binary array of as a raw transaction.
func (d *DissectedTransaction) Bytes() (buf []byte) {
	for _, v := range d.Content {
		if len(v) > 0 {
			buf = append(buf, v...)
		}
	}
	return buf
}

// PrepareForSign prepares a dissected raw transaction for a given
// vin script for signature.
func (d *DissectedTransaction) PrepareForSign(vin int, scr []byte) error {
	d.VinSlot = vin
	// create a local copy of the content.
	var tx [][]byte
	for _, b := range d.Content {
		bl := len(b)
		r := make([]byte, bl)
		copy(r, b)
		tx = append(tx, r)
	}
	// reset all vin scripts
	n, _, err := GetVarUint(tx[1], 0)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		tx[5*i+4] = []byte{0}
		tx[5*i+5] = []byte{}
	}
	// set the script for the given vin slot
	tx[5*vin+4] = PutVarUint(uint(len(scr)))
	tx[5*vin+5] = scr
	// flatten content to binary array.
	d.Signable = make([]byte, 0)
	for _, v := range tx {
		if len(v) > 0 {
			d.Signable = append(d.Signable, v...)
		}
	}
	return nil
}

// GetUint converts 'n' bytes in a buffer 'buf' starting at position 'p' into
// an unsigned integer.
func GetUint(buf []byte, p, n int) (v uint64, err error) {
	if p+n > len(buf) {
		err = errors.New("GetUint: buffer too small")
	} else {
		v = 0
		for i := n; i > 0; i-- {
			v = uint64(buf[p+i-1]) + 256*v
		}
	}
	return
}

// GetVarUint gets a variable length unsigned integer from a buffer.
func GetVarUint(buf []byte, p int) (uint64, int, error) {
	switch buf[p] {
	case 0xfd:
		v, err := GetUint(buf, p+1, 2)
		return v, 3, err
	case 0xfe:
		v, err := GetUint(buf, p+1, 4)
		return v, 5, err
	case 0xff:
		return 0, 1, errors.New("Invalid VarUint")
	default:
		return uint64(buf[p]), 1, nil
	}
}

// PutUint encodes an uint into a byte array of given length (1,2 or 4).
func PutUint(n uint, j int) []byte {
	b := make([]byte, j)
	switch j {
	case 4:
		b[3] = byte((n >> 24) & 0xFF)
		b[2] = byte((n >> 16) & 0xFF)
		fallthrough
	case 2:
		b[1] = byte((n >> 8) & 0xFF)
		fallthrough
	case 1:
		b[0] = byte(n & 0xFF)
	}
	return b
}

// PutVarUint encodes a var_uint into a byte array.
func PutVarUint(n uint) []byte {
	switch {
	case n < 253:
		return PutUint(n, 1)
	case n < 65536:
		return PutUint(n, 2)
	default:
		return PutUint(n, 4)
	}
}
