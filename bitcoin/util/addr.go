package util

import (
	"github.com/bfix/gospel/bitcoin/ecc"
)

// Address type (string-like base58 encoded data)
type Address string

// MakeAddress computes an address from public key for the "real" Bitcoin network
func MakeAddress(key *ecc.PublicKey) Address {
	return buildAddr(key, 0)
}

// MakeTestAddress computes an address from public key for the test network
func MakeTestAddress(key *ecc.PublicKey) Address {
	return buildAddr(key, 111)
}

// helper: compute address from public key using different (nested)
// hashes and identifiers.
func buildAddr(key *ecc.PublicKey, version byte) Address {
	var addr []byte
	addr = append(addr, version)
	kh := Hash160(key.Bytes())
	addr = append(addr, kh...)
	cs := Hash256(addr)
	addr = append(addr, cs[:4]...)
	return Address(Base58Encode(addr))
}
