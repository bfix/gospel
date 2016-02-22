package util

/*
 * Raw transaction manipulation methods.
 * Implement "non-default" scriptSig/scriptPubkey combinations
 * (contracts).
 *
 * (c) 2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *#####################################################################
 *
 * The binary encoded raw transaction looks like this:
 *
 *  ------+-------------+---------+---------------------------------
 *  Level | Field       | length  | Value/Comment
 *  ------+-------------+---------+---------------------------------
 *      0 | VERSION     |    4    | 01000000 = version 1
 *      0 | N_VOUT      |    1    | Number of VOUT defs to follow =1
 *  ------+-------------+---------+---------------------------------
 *      1 | VOUT:TXID   |   32    | Transaction ID
 *      1 | VOUT:N      |    4    | VOUT number in transaction
 *      1 | VOUT:SCRIPT |         | scriptSig (hex-encoded)
 *  ------+-------------+---------+---------------------------------
 *      2 | SCRIPT:LEN  |    1    | Length of script
 *      2 | SCRIPT:DATA |   <n>   | Script data
 *  ------+-------------+---------+---------------------------------
 *      1 | VOUT:SEQ    |         | FFFFFFFF = sequence number (-1)
 *  ------+-------------+---------+---------------------------------
 *      0 | N_VIN       |    1    | Number of VIN defs to follow =1
 *  ------+-------------+---------+---------------------------------
 *      1 | VIN:VALUE   |    4    | Number of Satoshis (1e-8 btc)
 *      1 | VIN:???     |    4    | 00000000 = ????
 *      1 | VIN:SCRIPT  |         | scriptPubkey
 *  ------+-------------+---------+---------------------------------
 *      2 | SCRIPT:LEN  |    1    | Length of script
 *      2 | SCRIPT:DATA |   <n>   | Script data
 *  ------+-------------+---------+---------------------------------
 *      0 | LOCKTIME    |    4    | 00000000 = locktime
 *  ------+-------------+---------+---------------------------------
 *
 *#####################################################################
 */

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"encoding/hex"
	"errors"
)

///////////////////////////////////////////////////////////////////////

// NullDataScript assembles a TX_NULL_DATA script.
func NullDataScript(data []byte) ([]byte, error) {
	size := len(data)
	if size > 75 {
		return nil, errors.New("attached data to big")
	}

	var script []byte
	script = append(script, 0x6a) // OP_RETURN
	script = append(script, LengthPrefix(size)...)
	script = append(script, data...)
	return script, nil
}

///////////////////////////////////////////////////////////////////////

// ReplaceScriptPubKey changes "scriptPubKey" to a new script.
// This only works if there is only one input/output slot defined in
// the transaction. The old "scriptPubKey" is completely dropped.
func ReplaceScriptPubKey(raw string, script []byte) (string, error) {

	// decode raw string from hex
	inRaw, err := hex.DecodeString(raw)
	if err != nil {
		return "", err
	}

	// dissect raw transaction and change VOUT
	pos := 4
	nVout := int(inRaw[pos])
	if nVout != 1 {
		return "", errors.New("invalid number of vout (!= 1)")
	}
	pos += 37
	scrlen := int(inRaw[pos])
	if scrlen != 0 {
		return "", errors.New("invalid scriptSig size (!= 0)")
	}
	pos += scrlen + 5
	nVin := int(inRaw[pos])
	if nVin != 1 {
		return "", errors.New("invalid number of vin (!= 1)")
	}
	pos += 9
	scrlen = int(inRaw[pos])

	var outRaw []byte
	outRaw = append(outRaw, inRaw[:pos]...)
	outRaw = append(outRaw, LengthPrefix(len(script))...)
	outRaw = append(outRaw, script...)
	outRaw = append(outRaw, inRaw[pos+scrlen+1:]...)

	// return new raw transaction
	return hex.EncodeToString(outRaw), nil
}

///////////////////////////////////////////////////////////////////////

// LengthPrefix assembles the length prefix for data.
func LengthPrefix(size int) []byte {
	var prefix []byte
	switch {
	case size < 76:
		prefix = append(prefix, byte(size))
	case size < 256:
		prefix = append(prefix, 0x4c)
		prefix = append(prefix, byte(size))
	case size < 65536:
		// size of script
		prefix = append(prefix, 0x4d)
		prefix = append(prefix, byte(size&0xFF))
		prefix = append(prefix, byte((size>>8)&0xFF))
	case size < 65536:
		prefix = append(prefix, 0x4d)
		prefix = append(prefix, byte(size&0xFF))
		prefix = append(prefix, byte((size>>8)&0xFF))
		prefix = append(prefix, byte((size>>16)&0xFF))
		prefix = append(prefix, byte((size>>24)&0xFF))
	}
	return prefix
}
