package util

import (
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

// Hash160 computes RIPEMD-160(SHA-256(data))
func Hash160(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	ripemd := ripemd160.New()
	ripemd.Write(sha2.Sum(nil))
	return ripemd.Sum(nil)
}

// Hash256 computes SHA-256(SHA-256(data))
func Hash256(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	h := sha2.Sum(nil)
	sha2.Reset()
	sha2.Write(h)
	return sha2.Sum(nil)
}

// Sha256 computes SHA-256(data)
func Sha256(data []byte) []byte {
	sha2 := sha256.New()
	sha2.Write(data)
	return sha2.Sum(nil)
}
