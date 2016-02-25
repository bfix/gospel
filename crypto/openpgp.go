package crypto

import (
	"bytes"
	"errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

const (
	// KeySign returns a signing key (GetKeyFromIdentity)
	KeySign = iota
	// KeyEncrypt returns a encryption key (GetKeyFromIdentity)
	KeyEncrypt
	// KeyAuth returns an authorization key (GetKeyFromIdentity)
	KeyAuth
)

// GetPublicKey converts an ASCII-armored public key representation
// into an OpenPGP key.
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

// GetArmoredPublicKey returns an armored public key for entity.
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

// GetKeyFromIdentity returns a suitable subkey from entity for the
// given operation.
func GetKeyFromIdentity(ent *openpgp.Entity, mode int) *openpgp.Key {
	key := new(openpgp.Key)
	key.Entity = ent
	ki := -1
	for i, sk := range ent.Subkeys {
		switch mode {
		case KeySign:
			if sk.PublicKey.PubKeyAlgo.CanSign() {
				ki = i
				break
			}
		case KeyEncrypt:
			if sk.PublicKey.PubKeyAlgo.CanEncrypt() {
				ki = i
				break
			}
		case KeyAuth:
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
