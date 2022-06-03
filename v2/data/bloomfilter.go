package data

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

import (
	"crypto/sha256"
	"math"
	"math/big"
)

// A BloomFilter is a space/time efficient set of unique entries.
// It can not enumerate its elements, but can check if an entry is contained
// in the set. The check always succeeds for a contained entry, but can create
// "false-positives" (entries not contained in the map give a positive result).
// By adjusting the number of bits in the BloomFilter and the number of indices
// generated for an entry, a BloomFilter can handle a given number of entries
// with a desired upper-bound for the false-positive rate.
type BloomFilter struct {
	NumBits    uint32 `json:"numBits"`       // number of bits in filter
	NumIdx     uint8  `json:"numIdx"`        // number of indices
	NumIdxBits uint8  `json:"numIdxBits"`    // number of bits per index
	NumHash    uint8  `json:"numHash"`       // number of SHA256 hashes needed
	Bits       []byte `json:"bits" size:"*"` // bit storage
}

// NewBloomFilterDirect creates a new BloomFilter based on the number of bits
// in the filter and the number of indices to be used.
func NewBloomFilterDirect(numBits, numIdx int) *BloomFilter {
	numIdxBits := int(math.Ceil(math.Log2(float64(numBits))))
	return &BloomFilter{
		NumBits:    uint32(numBits),
		NumIdx:     uint8(numIdx),
		NumIdxBits: uint8(numIdxBits),
		NumHash:    uint8((numIdxBits*numIdx + 255) / 256),
		Bits:       make([]byte, (numBits+7)/8),
	}
}

// NewBloomFilter creates a new BloomFilter based on the upper-bounds for the
// number of entries and the "false-positive" rate.
func NewBloomFilter(numExpected int, falsePositiveRate float64) *BloomFilter {
	// do some math and calculate the number of indices and number of bits
	// in the new BloomFilter given an upper-bound for the number of entries
	// and the "false-positive" rate.
	numIdx := int(math.Ceil(-math.Log2(falsePositiveRate)))
	numBits := int(float64(numIdx*numExpected) / math.Ln2)
	return NewBloomFilterDirect(numBits, numIdx)
}

// SameKind checks if two BloomFilter have the same parameters.
func (bf *BloomFilter) SameKind(bf2 *BloomFilter) bool {
	return bf.NumBits == bf2.NumBits &&
		bf.NumHash == bf2.NumHash &&
		bf.NumIdx == bf2.NumIdx &&
		bf.NumIdxBits == bf2.NumIdxBits
}

// Add an entry to the BloomFilter.
func (bf *BloomFilter) Add(entry []byte) {
	for _, idx := range bf.indexList(entry) {
		pos, mask := resolve(idx)
		bf.Bits[pos] |= mask
	}
}

// Combine merges two BloomFilters (of same kind) into a new one.
func (bf *BloomFilter) Combine(bf2 *BloomFilter) *BloomFilter {
	if !bf.SameKind(bf2) {
		return nil
	}
	res := &BloomFilter{
		NumBits:    bf.NumBits,
		NumIdx:     bf.NumIdx,
		NumIdxBits: bf.NumIdxBits,
		NumHash:    bf.NumHash,
		Bits:       make([]byte, len(bf.Bits)),
	}
	for i := range res.Bits {
		res.Bits[i] = bf.Bits[i] | bf2.Bits[i]
	}
	return res
}

// Contains returns true if the BloomFilter contains the given entry, and
// false otherwise. If an entry was added to the set, this function will
// always return 'true'. It can return 'true' for entries not in the set
// ("false-positives").
func (bf *BloomFilter) Contains(entry []byte) bool {
	for _, idx := range bf.indexList(entry) {
		pos, mask := resolve(idx)
		if (bf.Bits[pos] & mask) == 0 {
			return false
		}
	}
	return true
}

// Helper method to extract the list of indices for an entry.
func (bf *BloomFilter) indexList(entry []byte) []int {
	totalIdx := make([]byte, 0)
	hasher := sha256.New()
	var i uint8
	for i = 0; i < bf.NumHash; i++ {
		hasher.Write(entry)
		totalIdx = hasher.Sum(totalIdx)
	}
	v := new(big.Int).SetBytes(totalIdx)
	mask := big.NewInt((1 << uint(bf.NumIdxBits)) - 1)
	list := make([]int, bf.NumIdx)
	for i = 0; i < bf.NumIdx; i++ {
		j := new(big.Int).And(v, mask)
		list[i] = int(j.Int64()) % int(bf.NumBits)
		v = new(big.Int).Rsh(v, uint(bf.NumIdxBits))
	}
	return list
}

// Helper method to resolve an index into byte/bit positions in the data
// of the BloomFilter.
func resolve(idx int) (int, byte) {
	return idx >> 3, byte(1 << uint(idx&7))
}
