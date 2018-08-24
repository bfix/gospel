package bitcoin

import (
	"crypto/sha1"
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

// RipeMD160 computes RIPEMD160(data)
func RipeMD160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)
	return ripemd.Sum(nil)
}

// Sha1 computes SHA1(data)
func Sha1(data []byte) []byte {
	sha1 := sha1.New()
	sha1.Write(data)
	return sha1.Sum(nil)
}
