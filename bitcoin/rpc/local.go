package rpc

import ()

// GetInfo returns an object containing various state info.
func (s *Session) GetInfo() (*Info, error) {
	res, err := s.call("getinfo", nil)
	if err != nil {
		return nil, err
	}
	info := new(Info)
	if err = res.UnmarshalResult(info); err != nil {
		return nil, err
	}
	return info, err
}

// GetDifficulty returns the proof-of-work difficulty as a multiple
// of the minimum difficulty.
func (s *Session) GetDifficulty() (float64, error) {
	res, err := s.call("getdifficulty", nil)
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// GetMemPoolAncestors returns all in-mempool ancestors for a transaction
// in the mempool as an array of TXIDs belonging to transactions in the
// memory pool. The array may be empty if there are no transactions in the
// memory pool.
func (s *Session) GetMemPoolAncestors(addr string) ([]string, error) {
	res, err := s.call("getmempoolancestors", []Data{addr, false})
	if err != nil {
		return nil, err
	}
	return res.Result.([]string), nil
}

// GetMemPoolAncestorObjs returns all in-mempool ancestors for a transaction
// in the mempool as an array of MemPoolTransaction objects.
func (s *Session) GetMemPoolAncestorObjs(addr string) ([]*MemPoolTransaction, error) {
	res, err := s.call("getmempoolancestors", []Data{addr, true})
	if err != nil {
		return nil, err
	}
	var anc []*MemPoolTransaction
	if err = res.UnmarshalResult(&anc); err != nil {
		return nil, err
	}
	return anc, nil
}

// GetMemPoolDecendants returns all in-mempool decendants for a transaction
// in the mempool as an array of TXIDs belonging to transactions in the
// memory pool. The array may be empty if there are no transactions in the
// memory pool.
func (s *Session) GetMemPoolDecendants(addr string) ([]string, error) {
	res, err := s.call("getmempooldecendants", []Data{addr, false})
	if err != nil {
		return nil, err
	}
	return res.Result.([]string), nil
}

// GetMemPoolDecendantObjs returns all in-mempool decendants for a transaction
// in the mempool as an array of MemPoolTransaction objects.
func (s *Session) GetMemPoolDecendantObjs(addr string) ([]*MemPoolTransaction, error) {
	res, err := s.call("getmempooldecendants", []Data{addr, true})
	if err != nil {
		return nil, err
	}
	var anc []*MemPoolTransaction
	if err = res.UnmarshalResult(&anc); err != nil {
		return nil, err
	}
	return anc, nil
}

// GetMemPoolEntry returns mempool data for given transaction (must be in
// mempool).
func (s *Session) GetMemPoolEntry(addr string) (*MemPoolTransaction, error) {
	res, err := s.call("getmempoolentry", []Data{addr})
	if err != nil {
		return nil, err
	}
	e := new(MemPoolTransaction)
	if err = res.UnmarshalResult(e); err != nil {
		return nil, err
	}
	return e, nil
}

// GetMemPoolInfo returns information about the memory pool.
func (s *Session) GetMemPoolInfo() (*MemPoolInfo, error) {
	res, err := s.call("getmempoolinfo", nil)
	if err != nil {
		return nil, err
	}
	mi := new(MemPoolInfo)
	if err = res.UnmarshalResult(mi); err != nil {
		return nil, err
	}
	return mi, nil
}
