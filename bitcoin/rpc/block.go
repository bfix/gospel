package rpc

// Generate nearly instantly generates blocks.
// if maxTries < 0, use the default retry value (1000000).
func (s *Session) Generate(n, maxTries int) ([]string, error) {
	data := []Data{n}
	if maxTries > 0 {
		data = append(data, maxTries)
	}
	res, err := s.call("generate", data)
	if err != nil {
		return nil, err
	}
	var list []string
	val := res.Result.([]interface{})
	for _, v := range val {
		list = append(list, v.(string))
	}
	return list, nil
}

// GenerateToAddress mines blocks immediately to a specified address.
func (s *Session) GenerateToAddress(n int, addr string, maxTries int) ([]string, error) {
	data := []Data{n, addr}
	if maxTries > 0 {
		data = append(data, maxTries)
	}
	res, err := s.call("generate", data)
	if err != nil {
		return nil, err
	}
	var list []string
	val := res.Result.([]interface{})
	for _, v := range val {
		list = append(list, v.(string))
	}
	return list, nil
}

// GetBestBlockHash returns the header hash of the most recent block on the
// best block chain.
func (s *Session) GetBestBlockHash() (string, error) {
	res, err := s.call("getbestblockhash", nil)
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetBlock returns information about the given block hash.
func (s *Session) GetBlock(hash string) (*Block, error) {
	res, err := s.call("getblock", []Data{hash})
	if err != nil {
		return nil, err
	}

	block := new(Block)
	if err = res.UnmarshalResult(block); err != nil {
		return nil, err
	}
	return block, nil
}

// GetBlockAsJSON returns information about the given block hash in JSON.
func (s *Session) GetBlockAsJSON(hash string) (string, error) {
	res, err := s.call("getblock", []Data{hash, true})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetBlockchainInfo provides information about the current state of the
// block chain.
func (s *Session) GetBlockchainInfo() (*BlockchainInfo, error) {
	res, err := s.call("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}
	bc := new(BlockchainInfo)
	if err = res.UnmarshalResult(bc); err != nil {
		return nil, err
	}
	return bc, nil
}

// GetBlockCount returns the number of blocks in the longest
// block chain.
func (s *Session) GetBlockCount() (int, error) {
	res, err := s.call("getblockcount", nil)
	if err != nil {
		return -1, err
	}
	num := int(res.Result.(float64))
	return num, err
}

// GetBlockHash returns hash of block in best-block-chain at height.
func (s *Session) GetBlockHash(height int) (string, error) {
	res, err := s.call("getblockhash", []Data{height})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}
