/*
 * Raw transaction enhancement methods.
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
 * Insert (additional) data into raw transaction:
 * ==============================================
 * The binary encoded raw transaction looks like this:
 *
 *  ------+-------------+---------+---------------------------------
 *  Level | Field       | length  | Value/Comment
 *  ------+-------------+---------+---------------------------------
 *      0 | VERSION     |    4    | 01000000 = version 1
 *      0 | N_VOUT      |    1    | Number of VOUT defs to follow
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
 *      0 | N_VIN       |    1    | Number of VIN defs to follow
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
 */
func InjectData(raw string, data []byte, cost int) (string, error) {

	// limit costs to 2^16-1 satoshis
	if cost > 65535 {
		cost = 65535
	}
	// decode raw string from hex
	in_raw, err := hex.DecodeString(raw)
	if err != nil {
		return "", err
	}
	size := len(data)

	out_raw := make([]byte, 0)
	yscript := make([]byte, 0)

	// handle data depending on length
	switch {
	// too much data
	case size > 200:
		return "", errors.New("too much data to inject")

	// no data at all
	case size == 0:
		return "", errors.New("no data to inject")

	// small data (works on live Bitcoin network
	case size <= 20:
		// create a fake "PayToScriptHash"
		yscript = append(yscript, 0xa9)
		yscript = append(yscript, 0x14)
		yscript = append(yscript, data...)
		for n := size; n < 20; n++ {
			yscript = append(yscript, 0)
		}
		yscript = append(yscript, 0x87)

	// handle more data - works only on testnet!
	default:
		//---------------------------------------------------------
		//		<data>		; push <data> to stack, contains info
		//		OP_DROP		; drop info (not needed in computation)
		//---------------------------------------------------------
		switch {
		case size <= 75:
			yscript = append(yscript, byte(size&0xFF))
			yscript = append(yscript, data...)
		case size < 256:
			yscript = append(yscript, 76)
			yscript = append(yscript, byte(size&0xFF))
			yscript = append(yscript, data...)
		case size < 65536:
			yscript = append(yscript, 77)
			yscript = append(yscript, byte((size>>8)&0xFF))
			yscript = append(yscript, byte(size&0xFF))
			yscript = append(yscript, data...)
		default:
			yscript = append(yscript, 78)
			yscript = append(yscript, byte((size>>24)&0xFF))
			yscript = append(yscript, byte((size>>16)&0xFF))
			yscript = append(yscript, byte((size>>8)&0xFF))
			yscript = append(yscript, byte(size&0xFF))
			yscript = append(yscript, data...)
		}
		yscript = append(yscript, 117)
		yscript = append(yscript, 81)
	}

	//-----------------------------------------------------------------
	// dissect raw transaction and inject fake VOUT
	pos := 4
	n_vout := int(in_raw[pos])
	pos++
	for n := 0; n < n_vout; n++ {
		pos += 36
		scrlen := int(in_raw[pos])
		pos += scrlen + 5
	}
	out_raw = append(out_raw, in_raw[:pos]...)

	n_vin := int(in_raw[pos])
	pos++
	lastPos := pos

	out_raw = append(out_raw, byte(n_vin+1))
	for n := 0; n < n_vin; n++ {
		pos += 8
		scrlen := int(in_raw[pos])
		pos += scrlen + 1
	}
	out_raw = append(out_raw, in_raw[lastPos:pos]...)

	out_raw = append(out_raw, byte(cost&0xFF))
	out_raw = append(out_raw, byte((cost>>8)&0xFF))
	out_raw = append(out_raw, []byte{0, 0, 0, 0, 0, 0}...)
	out_raw = append(out_raw, byte(len(yscript)))
	out_raw = append(out_raw, yscript...)

	out_raw = append(out_raw, in_raw[pos:]...)

	// return new raw transaction
	return hex.EncodeToString(out_raw), nil
}
