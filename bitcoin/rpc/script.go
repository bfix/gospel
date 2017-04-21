package rpc

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
	if err = res.UnmarshalResult(ds); err != nil {
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
