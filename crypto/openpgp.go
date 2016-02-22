/*
 * OpenPGP helper functions.
 *
 * (c) 2013-2014 Bernd Fix    >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package crypto

///////////////////////////////////////////////////////////////////////
// Import external declarations.

import (
	"bytes"
	"errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

///////////////////////////////////////////////////////////////////////
// Module-global constants and variables

const (
	KEY_SIGN = iota
	KEY_ENCRYPT
	KEY_AUTH
)

///////////////////////////////////////////////////////////////////////
/*
 * Convert a ASCII-armored public key representation into
 * an OpenPGP key.
 * @param buf []byte - armored public key representation
 * @return *packet.PublicKey - OpenPGP key
 * @return error - error instance or nil
 */
func GetPublicKey(buf []byte) (*packet.PublicKey, error) {
	keyRdr, err := armor.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	keyData, err := packet.Read(keyRdr.Body)
	if err != nil {
		return nil, err
	}
	key, ok := keyData.(*packet.PublicKey)
	if !ok {
		return nil, errors.New("Invalid public key")
	}
	return key, nil
}

//---------------------------------------------------------------------
/*
 * Get armored public key for entity.
 * @param ent *openpgp.Entity - OpenPGP entity
 * @return []byte - armored public key representation
 * @return error - error instance or nil
 */
func GetArmoredPublicKey(ent *openpgp.Entity) ([]byte, error) {
	out := new(bytes.Buffer)
	wrt, err := armor.Encode(out, openpgp.PublicKeyType, nil)
	if err != nil {
		return nil, err
	}
	err = ent.Serialize(wrt)
	wrt.Close()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

//---------------------------------------------------------------------
/*
 * Get a suitable subkey from entity.
 * @param ent *openpgp.Entity - main entity
 * @param mode int - key mode (keyXXX)
 * @return *openpgp.Key - found subkey
 */
func GetKeyFromIdentity(ent *openpgp.Entity, mode int) *openpgp.Key {
	key := new(openpgp.Key)
	key.Entity = ent
	ki := -1
	for i, sk := range ent.Subkeys {
		switch mode {
		case KEY_SIGN:
			if sk.PublicKey.PubKeyAlgo.CanSign() {
				ki = i
				break
			}
		case KEY_ENCRYPT:
			if sk.PublicKey.PubKeyAlgo.CanEncrypt() {
				ki = i
				break
			}
		case KEY_AUTH:
			ki = i
			break
		}
	}
	if ki >= 0 {
		key.PublicKey = ent.Subkeys[ki].PublicKey
		key.PrivateKey = ent.Subkeys[ki].PrivateKey
		key.SelfSignature = ent.Subkeys[ki].Sig
	} else {
		key.PublicKey = ent.PrimaryKey
		key.PrivateKey = ent.PrivateKey
		for _, id := range ent.Identities {
			key.SelfSignature = id.SelfSignature
			break
		}
	}
	return key
}
