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
	"crypto/sha256"
	"errors"
	"github.com/bfix/gospel/bitcoin/ecc"
)

///////////////////////////////////////////////////////////////////////
/*
 * Export private key
 * @param k *PrivateKey - key to be exported
 * @return string - private key in SIPA format
 */
func ExportPrivateKey(k *ecc.PrivateKey) string {
	exp := make([]byte, 0)
	exp = append(exp, 0x80)
	exp = append(exp, k.Bytes()...)

	sha2 := sha256.New()
	sha2.Write(exp)
	h := sha2.Sum(nil)
	sha2.Reset()
	sha2.Write(h)
	cs := sha2.Sum(nil)
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
func ImportPrivateKey(keydata string) (*ecc.PrivateKey, error) {

	data, err := Base58Decode(keydata)
	if err != nil {
		return nil, err
	}
	if data[0] != 0x80 || len(data) != 37 {
		return nil, errors.New("Invalid key format")
	}

	k := data[1:33]
	c := data[33:]

	t := make([]byte, 0)
	t = append(t, 0x80)
	t = append(t, k...)

	sha2 := sha256.New()
	sha2.Write(t)
	h := sha2.Sum(nil)
	sha2.Reset()
	sha2.Write(h)
	cs := sha2.Sum(nil)
	if bytes.Compare(c, cs[:4]) != 0 {
		return nil, errors.New("Invalid key data")
	}
	return ecc.PrivateKeyFromBytes(k)
}
