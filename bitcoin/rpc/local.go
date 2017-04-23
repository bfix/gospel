package rpc

import ()

// GetInfo returns an object containing various state info.
func (s *Session) GetInfo() (*Info, error) {
	res, err := s.call("getinfo", nil)
	if err != nil {
		return nil, err
	}
	info := new(Info)
	if ok, err := res.UnmarshalResult(info); !ok {
		return nil, err
	}
	return info, nil
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
	if ok, err := res.UnmarshalResult(&anc); !ok {
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
	if ok, err := res.UnmarshalResult(&anc); !ok {
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
	if ok, err := res.UnmarshalResult(e); !ok {
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
	if ok, err := res.UnmarshalResult(mi); !ok {
		return nil, err
	}
	return mi, nil
}

// GetRawMemPoolList returns all transaction identifiers (TXIDs) in the memory.
// pool as a JSON array.
func (s *Session) GetRawMemPoolList() ([]string, error) {
	res, err := s.call("getrawmempool", []Data{false})
	if err != nil {
		return nil, err
	}
	var txids []string
	if ok, err := res.UnmarshalResult(&txids); !ok {
		return nil, err
	}
	return txids, nil
}

// GetRawMemPool returns all transaction identifiers (TXIDs) in the memory
// pool with detailed information about each transaction in the memory pool.
// GetMemPoolInfo returns information about the memory pool.
func (s *Session) GetRawMemPool() (map[string]*MemPoolTransaction, error) {
	res, err := s.call("getrawmempool", []Data{true})
	if err != nil {
		return nil, err
	}
	var list map[string]*MemPoolTransaction
	if ok, err := res.UnmarshalResult(&list); !ok {
		return nil, err
	}
	return list, nil
}

// Stop safely shuts down the Bitcoin Core server.
func (s *Session) Stop() error {
	_, err := s.call("stop", nil)
	return err
}
