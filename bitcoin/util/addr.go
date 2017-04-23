package util

import (
	"github.com/bfix/gospel/bitcoin/ecc"
)

// MakeAddress computes an address from public key for the "real" Bitcoin network
func MakeAddress(key *ecc.PublicKey) string {
	return buildAddr(key, 0)
}

// MakeTestAddress computes an address from public key for the test network
func MakeTestAddress(key *ecc.PublicKey) string {
	return buildAddr(key, 111)
}

// helper: compute address from public key using different (nested)
// hashes and identifiers.
func buildAddr(key *ecc.PublicKey, version byte) string {
	var addr []byte
	addr = append(addr, version)
	kh := Hash160(key.Bytes())
	addr = append(addr, kh...)
	cs := Hash256(addr)
	addr = append(addr, cs[:4]...)
	return string(Base58Encode(addr))
}
