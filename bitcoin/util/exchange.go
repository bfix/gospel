/*
 * Bitcoin import/export methods.
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
	"bytes"
	"errors"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
)

///////////////////////////////////////////////////////////////////////
/*
 * Export private key
 * @param k *PrivateKey - key to be exported
 * @return string - private key in SIPA format
 */
func ExportPrivateKey(k *ecc.PrivateKey, testnet bool) string {
	exp := make([]byte, 0)
	if testnet {
		exp = append(exp, 0xEF)
	} else {
		exp = append(exp, 0x80)
	}
	exp = append(exp, k.Bytes()...)

	cs := Hash256(exp)
	exp = append(exp, cs[:4]...)

	return Base58Encode(exp)
}

///////////////////////////////////////////////////////////////////////
/*
 * Import private key
 * @param keydata string - private key in SIPA format
 * @return *PrivateKey - imported private key
 * @return error
 */
func ImportPrivateKey(keydata string, testnet bool) (*ecc.PrivateKey, error) {
	// decode and check data
	data, err := Base58Decode(keydata)
	if err != nil {
		return nil, err
	}
	if testnet {
		if data[0] != 0xEF {
			msg := fmt.Sprintf("Invalid key version: %d (testnet)\n", int(data[0]))
			return nil, errors.New(msg)
		}
	} else {
		if data[0] != 0x80 {
			msg := fmt.Sprintf("Invalid key version: %d\n", int(data[0]))
			return nil, errors.New(msg)
		}
	}
	// copy key data
	var k, c []byte
	if len(data) == 37 {
		// uncompressed public key
		k = data[1:33]
		c = data[33:37]
	} else if len(data) == 38 {
		// compressed public key
		k = data[1:34]
		c = data[34:38]
		if data[33] != 1 {
			msg := fmt.Sprintf("Invalid key compression indicator: %d\n", int(data[33]))
			return nil, errors.New(msg)
		}
	} else {
		return nil, errors.New("Invalid key format")
	}
	// recompute and verify checksum
	t := make([]byte, 0)
	t = append(t, data[0])
	t = append(t, k...)
	cs := Hash256(t)
	if bytes.Compare(c, cs[:4]) != 0 {
		return nil, errors.New("Invalid key data")
	}
	// return key
	return ecc.PrivateKeyFromBytes(k)
}
