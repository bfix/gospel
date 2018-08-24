package bitcoin

// MakeAddress computes an address from public key for the "real" Bitcoin network
func MakeAddress(key *PublicKey) string {
	return buildAddr(key, 0)
}

// MakeTestAddress computes an address from public key for the test network
func MakeTestAddress(key *PublicKey) string {
	return buildAddr(key, 111)
}

// helper: compute address from public key using different (nested)
// hashes and identifiers.
func buildAddr(key *PublicKey, version byte) string {
	var addr []byte
	addr = append(addr, version)
	kh := Hash160(key.Bytes())
	addr = append(addr, kh...)
	cs := Hash256(addr)
	addr = append(addr, cs[:4]...)
	return string(Base58Encode(addr))
}
