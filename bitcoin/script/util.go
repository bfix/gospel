package script

import (
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"github.com/bfix/gospel/math"
	"strings"
)

// Compile compiles a Bitcoin script source into its binary representation.
func Compile(src string) (bin []byte, err error) {
	add := func(b []byte) {
		lb := uint(len(b))
		if lb < 76 {
			bin = append(bin, util.PutUint(lb, 1)...)
		} else if lb < 65536 {
			bin = append(bin, 0xfd)
			bin = append(bin, util.PutUint(lb, 2)...)
		} else {
			bin = append(bin, 0xfe)
			bin = append(bin, util.PutUint(lb, 4)...)
		}
		if lb > 0 {
			bin = append(bin, b...)
		}
	}
	for _, op := range strings.Split(src, " ") {
		if len(op) == 0 {
			continue
		}
		if strings.HasPrefix(op, "OP_") {
			found := false
			for _, opcode := range OpCodes {
				if opcode.Name == op {
					bin = append(bin, opcode.Value)
					found = true
					break
				}
			}
			if !found {
				return bin, fmt.Errorf("Unknown opcode '%s'", op)
			}
		} else if strings.HasPrefix(op, "#") {
			v := math.NewIntFromString(op[1:])
			add(v.Bytes())
		} else {
			b, err := hex.DecodeString(op)
			if err != nil {
				return nil, err
			}
			add(b)
		}
	}
	return
}

// Decompile returns a human-readable Bitcoin script source from a
// binary script representation.
func Decompile(bin []byte) (src string, err error) {
	convert := func(i, s int) {
		if s < 5 {
			v := math.NewIntFromBytes(bin[i : i+s])
			src += "#" + v.String()
		} else {
			src += hex.EncodeToString(bin[i : i+s])
		}
	}
	lb := len(bin)
	for i := 0; i < lb; {
		op := bin[i]
		if len(src) > 0 {
			src += " "
		}
		if op > 0 && op < 76 {
			i++
			s := int(op)
			convert(i, s)
			i += s
		} else if op == 76 {
			i++
			s, err := util.GetUint(bin, i, 1)
			if err != nil {
				return src, err
			}
			i++
			convert(i, int(s))
			i += int(s)
		} else if op == 77 {
			i++
			s, err := util.GetUint(bin, i, 2)
			if err != nil {
				return src, err
			}
			i += 2
			convert(i, int(s))
			i += int(s)
		} else if op == 78 {
			i++
			s, err := util.GetUint(bin, i, 4)
			if err != nil {
				return src, err
			}
			i += 4
			convert(i, int(s))
			i += int(s)
		} else {
			found := false
			for _, opcode := range OpCodes {
				if opcode.Value == op {
					src += opcode.Name
					found = true
					break
				}
			}
			if !found {
				return src, fmt.Errorf("Unknown opcode '%v' at pos %d", op, i)
			}
			i++
		}
	}
	return
}

// Sign signs a prepared transaction with a private key
func Sign(prv *ecc.PrivateKey, hashType byte, tx *util.DissectedTransaction) ([]byte, error) {
	// compute hash of amended transaction
	txSign := append(tx.Signable, []byte{hashType, 0, 0, 0}...)
	txHash := util.Hash256(txSign)
	// sign the hash
	sig := ecc.Sign(prv, txHash)
	sigData, err := sig.Bytes()
	if err != nil {
		return nil, err
	}
	sigData = append(sigData, hashType)
	return sigData, nil
}

// Verify checks the signature on a prepared transaction with a public key.
func Verify(pub *ecc.PublicKey, sig []byte, tx *util.DissectedTransaction) (bool, error) {
	// extract ASN.1 signature object and hashtype
	lsig := len(sig)
	hashType := sig[lsig-1]
	tSig, err := ecc.NewSignatureFromASN1(sig[:lsig-1])
	if err != nil {
		return false, err
	}
	// compute hash of amended transaction
	txSign := append(tx.Signable, []byte{hashType, 0, 0, 0}...)
	txHash := util.Hash256(txSign)
	// perform signature verify
	return ecc.Verify(pub, txHash, tSig), nil
}
