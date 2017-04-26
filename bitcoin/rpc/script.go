package rpc

import (
	"github.com/bfix/gospel/bitcoin/script"
)

// AddWitnessAddress adds a witness address for a script (with pubkey or
// redeemscript known).
func (s *Session) AddWitnessAddress(addr string) (string, error) {
	res, err := s.call("addwitnessaddress", []Data{addr})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// DecodeScript decodes a hex-encoded P2SH redeem script.
func (s *Session) DecodeScript(script string) (*DecodedScript, error) {
	res, err := s.call("decodescript", []Data{script})
	if err != nil {
		return nil, err
	}
	ds := new(DecodedScript)
	if ok, err := res.UnmarshalResult(ds); !ok {
		return nil, err
	}
	return ds, nil
}

// GetAddresses returns an array of addresses attached to the script.
func (s *ScriptPubKey) GetAddresses() []string {
	var res []string
	switch s.Addresses.(type) {
	case string:
		res = append(res, s.Addresses.(string))
	case []string:
		res = s.Addresses.([]string)
	}
	return res
}

// DataScript assembles a OP_RETURN script with data attached.
func DataScript(data []byte) (scr []byte) {
	scr = append(scr, script.Op_RETURN)
	scr = append(scr, PushData(data)...)
	return
}

// PushData creates an opcode instruction to push binary data onto the stack.
func PushData(data []byte) (res []byte) {
	size := len(data)
	switch {
	case size < 76:
		res = append(res, byte(size))
	case size < 256:
		res = append(res, script.Op_PUSHDATA1)
		res = append(res, byte(size))
	case size < 65536:
		// size of script
		res = append(res, script.Op_PUSHDATA2)
		res = append(res, byte(size&0xFF))
		res = append(res, byte((size>>8)&0xFF))
	default:
		res = append(res, script.Op_PUSHDATA4)
		res = append(res, byte(size&0xFF))
		res = append(res, byte((size>>8)&0xFF))
		res = append(res, byte((size>>16)&0xFF))
		res = append(res, byte((size>>24)&0xFF))
	}
	res = append(res, data...)
	return
}
