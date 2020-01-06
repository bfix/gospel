package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

const (
	// GenesisHash is the hash value of the very first block in the blockchain.
	GenesisHash = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
	// GenesisHashTest is the hash value of the very first block in the
	// testnet blockchain.
	GenesisHashTest = "000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943"
)

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
	if ok, err := res.UnmarshalResult(block); !ok {
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
	if ok, err := res.UnmarshalResult(bc); !ok {
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

// GetBlockHeader gets a block header with a particular header hash from the
// local block database either as a JSON object or as a serialized block
// header.
func (s *Session) GetBlockHeader(hash string, asJSON bool) (string, error) {
	res, err := s.call("getblockheader", []Data{hash, asJSON})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// GetBlockTemplate gets a block template or proposal for use with mining
// software.
func (s *Session) GetBlockTemplate(caps []string) (*BlockTemplate, error) {
	param := new(BlockTemplateParameter)
	param.Capabilities = caps
	res, err := s.call("getblocktemplate", []Data{param})
	if err != nil {
		return nil, err
	}
	bt := new(BlockTemplate)
	if ok, err := res.UnmarshalResult(bt); !ok {
		return nil, err
	}
	return bt, nil
}

// GetChainTips returns information about the highest-height block (tip) of
// each local block chain.
func (s *Session) GetChainTips() ([]*ChainTip, error) {
	res, err := s.call("getchaintips", nil)
	if err != nil {
		return nil, err
	}
	var ct []*ChainTip
	if ok, err := res.UnmarshalResult(&ct); !ok {
		return nil, err
	}
	return ct, nil
}

// SubmitBlock accepts a block, verifies it is a valid addition to the block
// chain, and broadcasts it to the network. Extra parameters are ignored by
// Bitcoin Core but may be used by mining pools or other programs.
func (s *Session) SubmitBlock(block string, param interface{}) (string, error) {
	res, err := s.call("submitblock", []Data{block, param})
	if err != nil {
		return "", err
	}
	if res.Result == nil {
		return "", nil
	}
	return res.Result.(string), nil
}

// VerifyChain verifies each entry in the local block chain database.
// 'checkLevel' defines how thoroughly to check each block, from 0 to 4.
// Default is the level set with the -checklevel command line argument; if
// that isn’t set, the default is 3. Each higher level includes the tests
// from the lower levels. Levels are:
// -- 0. Read from disk to ensure the files are accessible
// -- 1. Ensure each block is valid
// -- 2. Make sure undo files can be read from disk and are in a valid format
// -- 3. Test each block undo to ensure it results in correct state
// -- 4. After undoing blocks, reconnect them to ensure they reconnect correctly
// 'numBlocks' is the number of blocks to verify. Set to 0 to check all blocks.
// Defaults to the value of the -checkblocks command-line argument; if that isn’t
// set, the default is 288.
func (s *Session) VerifyChain(checkLevel, numBlocks int) (bool, error) {
	data := []Data{}
	if checkLevel > 0 {
		data = append(data, checkLevel)
	}
	if numBlocks > 0 {
		data = append(data, numBlocks)
	}
	res, err := s.call("verifychain", data)
	if err != nil {
		return false, err
	}
	return res.Result.(bool), nil
}
