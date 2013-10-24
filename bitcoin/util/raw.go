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

package util

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"encoding/hex"
	"errors"
)

///////////////////////////////////////////////////////////////////////
/*
 * Change "scriptPubKey" to "PayToScriptHash":
 * This only works if there is only one input/output slot defined in
 * the transaction. The old "scriptPubKey" is completely dropped.
 *
 * @param raw string - hex string of raw transaction
 * @param script []byte - serialized script
 * @return string - new hex-encoded raw string
 * @return error - error instance (or nil)
 */
func PayToScriptHash(raw string, script []byte) (string, error) {

	// decode raw string from hex
	in_raw, err := hex.DecodeString(raw)
	if err != nil {
		return "", err
	}

	// compute hash over data
	hash := Hash160(script)

	// create a "PayToScriptHash" script
	yscript := make([]byte, 0)
	yscript = append(yscript, 0xa9)
	yscript = append(yscript, 0x14)
	yscript = append(yscript, hash...)
	yscript = append(yscript, 0x87)

	// dissect raw transaction and change VOUT
	pos := 4
	n_vout := int(in_raw[pos])
	if n_vout != 1 {
		return "", errors.New("invalid number of vout (!= 1)")
	}
	pos += 37
	scrlen := int(in_raw[pos])
	if scrlen != 0 {
		return "", errors.New("invalid scriptSig size (!= 0)")
	}
	pos += scrlen + 5
	n_vin := int(in_raw[pos])
	if n_vin != 1 {
		return "", errors.New("invalid number of vin (!= 1)")
	}
	pos += 9
	scrlen = int(in_raw[pos])

	out_raw := make([]byte, 0)
	out_raw = append(out_raw, in_raw[:pos]...)
	out_raw = append(out_raw, 23)
	out_raw = append(out_raw, yscript...)
	out_raw = append(out_raw, in_raw[pos+scrlen+1:]...)

	// return new raw transaction
	return hex.EncodeToString(out_raw), nil
}

///////////////////////////////////////////////////////////////////////
// Helper methods:

/*
 * Create P2SH redeem script that includes data
 */
func ScriptFromData(data []byte, key []byte) []byte {
	out := make([]byte, 0)

	// push data
	size := len(data)
	out = append(out, LengthPrefix(size)...)
	out = append(out, data...)
	// OP_DROP
	out = append(out, 0x75)

	// push public key
	size = len(key)
	out = append(out, LengthPrefix(size)...)
	out = append(out, key...)
	// push OP_CHECKSIG
	return append(out, 0xac)
}

//---------------------------------------------------------------------
/*
 * Assemble the length prefix for data.
 */
func LengthPrefix(size int) []byte {
	prefix := make([]byte, 0)
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
