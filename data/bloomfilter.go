package data

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
)

// BloomFilterBase for custom bloom filter implementations
// (e.g. simple, salted, counting, ...)
type BloomFilterBase struct {
	NumBits    uint32 `size:"big" json:"numBits"` // number of bits in filter
	NumIdx     uint8  `size:"big" json:"numIdx"`  // number of indices
	NumIdxBits uint8  `json:"numIdxBits"`         // number of bits per index
	NumHash    uint8  `json:"numHash"`            // number of SHA256 hashes needed
}

// Helper method to extract the list of indices for an entry.
func (bf *BloomFilterBase) indexList(entry []byte) []int {
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

//----------------------------------------------------------------------
// Generic bloomfilter
//----------------------------------------------------------------------

// A BloomFilter is a space/time efficient set of unique entries.
// It can not enumerate its elements, but can check if an entry is contained
// in the set. The check always succeeds for a contained entry, but can create
// "false-positives" (entries not contained in the map give a positive result).
// By adjusting the number of bits in the BloomFilter and the number of indices
// generated for an entry, a BloomFilter can handle a given number of entries
// with a desired upper-bound for the false-positive rate.
type BloomFilter struct {
	BloomFilterBase

	Bits []byte `size:"(BitsSize)" json:"bits"` // bit storage
}

// NewBloomFilterDirect creates a new BloomFilter based on the number of bits
// in the filter and the number of indices to be used.
func NewBloomFilterDirect(numBits, numIdx int) *BloomFilter {
	numIdxBits := int(math.Ceil(math.Log2(float64(numBits))))
	return &BloomFilter{
		BloomFilterBase: BloomFilterBase{
			NumBits:    uint32(numBits),
			NumIdx:     uint8(numIdx),
			NumIdxBits: uint8(numIdxBits),
			NumHash:    uint8((numIdxBits*numIdx + 255) / 256),
		},
		Bits: make([]byte, (numBits+7)/8),
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

// BitsSize returns the size of the byte array representing the filter bits.
func (bf *BloomFilter) BitsSize() uint {
	return uint((bf.NumBits + 7) / 8)
}

// Size returns the size of the binary representation
func (bf *BloomFilter) Size() uint {
	return uint(7 + len(bf.Bits))
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
		pos, mask := bf.resolve(idx)
		bf.Bits[pos] |= mask
	}
}

// Combine merges two BloomFilters (of same kind) into a new one.
func (bf *BloomFilter) Combine(bf2 *BloomFilter) *BloomFilter {
	if !bf.SameKind(bf2) {
		return nil
	}
	res := &BloomFilter{
		BloomFilterBase: BloomFilterBase{
			NumBits:    bf.NumBits,
			NumIdx:     bf.NumIdx,
			NumIdxBits: bf.NumIdxBits,
			NumHash:    bf.NumHash,
		},
		Bits: make([]byte, len(bf.Bits)),
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
		pos, mask := bf.resolve(idx)
		if (bf.Bits[pos] & mask) == 0 {
			return false
		}
	}
	return true
}

// Helper method to resolve an index into byte/bit positions in the data
// of the BloomFilter.
func (bf *BloomFilter) resolve(idx int) (int, byte) {
	return idx >> 3, byte(1 << uint(idx&7))
}

//----------------------------------------------------------------------
// Salted bloomfilter
//----------------------------------------------------------------------

// SaltedBloomFilter is a bloom filter where each entr is "salted" with
// a uint32 salt value before processing. As each filter have different
// salts, the same set of entries added to the filter will result in a
// different bit pattern for the filter resulting in different false-
// positives for the same set. Useful if a filter is repeatedly generated
// for the same (or similar) set of entries.
type SaltedBloomFilter struct {
	Salt []byte `size:"4"` // salt value
	BloomFilter
}

// NewSaltedBloomFilterDirect creates a new salted BloomFilter based on
// the number of bits in the filter and the number of indices to be used.
func NewSaltedBloomFilterDirect(salt uint32, numBits, numIdx int) *SaltedBloomFilter {
	bf := &SaltedBloomFilter{
		Salt:        make([]byte, 4),
		BloomFilter: *NewBloomFilterDirect(numBits, numIdx),
	}
	bf.setSalt(salt)
	return bf
}

// NewSaltedBloomFilter creates a new salted BloomFilter based on the
// upper-bounds for the number of entries and the "false-positive" rate.
func NewSaltedBloomFilter(salt uint32, numExpected int, falsePositiveRate float64) *SaltedBloomFilter {
	bf := &SaltedBloomFilter{
		Salt:        make([]byte, 4),
		BloomFilter: *NewBloomFilter(numExpected, falsePositiveRate),
	}
	bf.setSalt(salt)
	return bf
}

// Size returns the size of the binary representation
func (bf *SaltedBloomFilter) Size() uint {
	return bf.BloomFilter.Size() + 4
}

// Set salt for bloom filter
func (bf *SaltedBloomFilter) setSalt(salt uint32) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, salt)
	bf.Salt = buf.Bytes()
}

// Salt entry before processing
func (bf *SaltedBloomFilter) saltEntry(entry []byte) []byte {
	buf := make([]byte, len(entry)+4)
	copy(buf, bf.Salt)
	copy(buf[4:], entry)
	return buf
}

// Add an entry to the BloomFilter.
func (bf *SaltedBloomFilter) Add(entry []byte) {
	bf.BloomFilter.Add(bf.saltEntry(entry))
}

// Combine merges two salted BloomFilters (of same kind) into a new one.
func (bf *SaltedBloomFilter) Combine(bf2 *SaltedBloomFilter) *SaltedBloomFilter {
	if !bytes.Equal(bf.Salt, bf2.Salt) || !bf.BloomFilter.SameKind(&bf2.BloomFilter) {
		return nil
	}
	res := new(SaltedBloomFilter)
	res.Salt = make([]byte, 4)
	copy(res.Salt, bf.Salt)
	res.BloomFilter = *bf.BloomFilter.Combine(&bf2.BloomFilter)
	return res
}

// Contains returns true if the salted BloomFilter contains the given entry,
// and false otherwise. If an entry was added to the set, this function will
// always return 'true'. It can return 'true' for entries not in the set
// ("false-positives").
func (bf *SaltedBloomFilter) Contains(entry []byte) bool {
	return bf.BloomFilter.Contains(bf.saltEntry(entry))
}

//----------------------------------------------------------------------
// Counting bloomfilter
//----------------------------------------------------------------------

// CountingBloomFilter is an extension of a generic bloomfilter that
// keeps a count instead of a single bit for masking. This allows the
// deletion of entries for the cost of 32x emory increase.
type CountingBloomFilter struct {
	BloomFilterBase

	Counts []uint32 `size:"(NumBits)" json:"bits"` // counter storage
}

// NewCoutingBloomFilterDirect creates a new BloomFilter based on the
// number of bits in the filter and the number of indices to be used.
func NewCountingBloomFilterDirect(numBits, numIdx int) *CountingBloomFilter {
	numIdxBits := int(math.Ceil(math.Log2(float64(numBits))))
	return &CountingBloomFilter{
		BloomFilterBase: BloomFilterBase{
			NumBits:    uint32(numBits),
			NumIdx:     uint8(numIdx),
			NumIdxBits: uint8(numIdxBits),
			NumHash:    uint8((numIdxBits*numIdx + 255) / 256),
		},
		Counts: make([]uint32, numBits),
	}
}

// NewCoutingBloomFilter creates a new BloomFilter based on the upper-bounds
// for the number of entries and the "false-positive" rate.
func NewCountingBloomFilter(numExpected int, falsePositiveRate float64) *CountingBloomFilter {
	// do some math and calculate the number of indices and number of bits
	// in the new BloomFilter given an upper-bound for the number of entries
	// and the "false-positive" rate.
	numIdx := int(math.Ceil(-math.Log2(falsePositiveRate)))
	numBits := int(float64(numIdx*numExpected) / math.Ln2)
	return NewCountingBloomFilterDirect(numBits, numIdx)
}

// Size returns the size of the binary representation
func (bf *CountingBloomFilter) Size() uint {
	return uint(7 + 4*bf.NumBits)
}

// Add an entry to the BloomFilter.
func (bf *CountingBloomFilter) Add(entry []byte) {
	for _, idx := range bf.indexList(entry) {
		bf.Counts[idx] += 1
	}
}

// SameKind checks if two BloomFilter have the same parameters.
func (bf *CountingBloomFilter) SameKind(bf2 *CountingBloomFilter) bool {
	return bf.NumBits == bf2.NumBits &&
		bf.NumHash == bf2.NumHash &&
		bf.NumIdx == bf2.NumIdx &&
		bf.NumIdxBits == bf2.NumIdxBits
}

// Combine merges two BloomFilters (of same kind) into a new one.
func (bf *CountingBloomFilter) Combine(bf2 *CountingBloomFilter) *CountingBloomFilter {
	if !bf.SameKind(bf2) {
		return nil
	}
	res := &CountingBloomFilter{
		BloomFilterBase: BloomFilterBase{
			NumBits:    bf.NumBits,
			NumIdx:     bf.NumIdx,
			NumIdxBits: bf.NumIdxBits,
			NumHash:    bf.NumHash,
		},
		Counts: make([]uint32, bf.NumBits),
	}
	for i := range res.Counts {
		res.Counts[i] = bf.Counts[i] + bf2.Counts[i]
	}
	return res
}

// Contains returns true if the BloomFilter contains the given entry, and
// false otherwise. If an entry was added to the set, this function will
// always return 'true'. It can return 'true' for entries not in the set
// ("false-positives").
func (bf *CountingBloomFilter) Contains(entry []byte) bool {
	for _, idx := range bf.indexList(entry) {
		if bf.Counts[idx] == 0 {
			return false
		}
	}
	return true
}

// Remove an entry from the bloomfilter
func (bf *CountingBloomFilter) Remove(entry []byte) bool {
	// make sure the entry is stored in the filter
	idxList := bf.indexList(entry)
	for _, idx := range idxList {
		if bf.Counts[idx] == 0 {
			return false
		}
	}
	for _, idx := range idxList {
		bf.Counts[idx]--
	}
	return true
}
