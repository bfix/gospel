package util

import (
	"encoding/hex"
	"errors"
)

/*
 * A binary encoded raw transaction looks like this:
 *
 *  ------+-------------+---------+---------------------------------
 *  Level | Field       | length  | Value/Comment
 *  ------+-------------+---------+---------------------------------
 *      0 | VERSION     |    4    | version
 *      0 | N_VIN       |  var    | Number of Vin defs to follow
 *  ------+-------------+---------+---------------------------------
 *      1 | VIN:TXID    |   32    | Transaction ID
 *      1 | VIN:N       |    4    | VOUT number in transaction
 *      1 | VIN:SCRIPT  |         | scriptSig (hex-encoded)
 *  ------+-------------+---------+---------------------------------
 *      2 | SCRIPT:LEN  |   var   | Length of script
 *      2 | SCRIPT:DATA |   <n>   | Script data
 *  ------+-------------+---------+---------------------------------
 *      1 | VIN:SEQ     |         | sequence number
 *  ------+-------------+---------+---------------------------------
 *      0 | N_VOUT      |  var    | Number of Vout defs to follow
 *  ------+-------------+---------+---------------------------------
 *      1 | VOUT:AMOUNT |    4    | Number of Satoshis (1e-8 btc)
 *      1 | VOUT:INDEX  |    4    | Output index
 *      1 | VOUT:SCRIPT |         | scriptPubkey
 *  ------+-------------+---------+---------------------------------
 *      2 | SCRIPT:LEN  |   var   | Length of script
 *      2 | SCRIPT:DATA |   <n>   | Script data
 *  ------+-------------+---------+---------------------------------
 *      0 | LOCKTIME    |    4    | locktime
 *  ------+-------------+---------+---------------------------------
 */

// GetUint converts 'n' bytes in a buffer 'buf' starting at position 'p' into
// an unsigned integer.
func GetUint(buf []byte, p, n int) (v uint, err error) {
	if p+n > len(buf) {
		err = errors.New("GetUint: buffer too small")
	} else {
		v = 0
		for i := n; i > 0; i-- {
			v = uint(buf[p+i-1]) + 256*v
		}
	}
	return
}

// GetVarUint gets a variable length unsigned integer from a buffer.
func GetVarUint(buf []byte, p int) (uint, int, error) {
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
		return uint(buf[p]), 1, nil
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

// DissectRawTransaction dissects a raw transaction into its defining segments.
func DissectRawTransaction(rawHex string) (res [][]byte, err error) {
	var buf []byte
	if buf, err = hex.DecodeString(rawHex); err != nil {
		return nil, err
	}
	pos := 0
	add := func(n int) int {
		c := make([]byte, n)
		for i := 0; i < n; i++ {
			c[i] = buf[pos+i]
		}
		pos += n
		res = append(res, c)
		if n > 4 {
			return 0
		}
		v, _ := GetUint(c, 0, n)
		return int(v)
	}
	add(4)                            // version
	n, j, err := GetVarUint(buf, pos) // number of inputs
	if err != nil {
		return nil, err
	}
	res = append(res, buf[pos:pos+j])
	pos += j
	for i := 0; i < int(n); i++ {
		add(32)                           // input address hash
		add(4)                            // input index
		s, j, err := GetVarUint(buf, pos) // size of script
		if err != nil {
			return nil, err
		}
		res = append(res, buf[pos:pos+j])
		pos += j
		if s == 0 {
			res = append(res, []byte{})
		} else {
			add(int(s)) // script code
		}
		add(4) // sequence
	}
	n, j, err = GetVarUint(buf, pos) // number of outputs
	if err != nil {
		return nil, err
	}
	res = append(res, buf[pos:pos+j])
	pos += j
	for i := 0; i < int(n); i++ {
		add(4)                            // amount
		add(4)                            // output index
		s, j, err := GetVarUint(buf, pos) // size of script
		if err != nil {
			return nil, err
		}
		res = append(res, buf[pos:pos+j])
		pos += j
		if s == 0 {
			res = append(res, []byte{})
		} else {
			add(int(s)) // script code
		}
	}
	add(4) // locktime
	return
}

// PrepareTxForSign prepares a dissected raw transaction for a given
// vin script for signature.
func PrepareTxForSign(tx [][]byte, vin int, scr []byte) ([]byte, error) {
	n, _, err := GetVarUint(tx[1], 0)
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(n); i++ {
		tx[5*i+4] = []byte{0}
		tx[5*i+5] = []byte{}
	}
	tx[5*vin+4] = PutVarUint(uint(len(scr)))
	tx[5*vin+5] = scr
	var buf []byte
	for _, v := range tx {
		if len(v) > 0 {
			buf = append(buf, v...)
		}
	}
	return buf, nil
}
