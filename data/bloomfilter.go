package data

import (
	"crypto/sha256"
	"hash"
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
	numBits    int       // number of bits in filter
	numIdx     int       // number of indices
	numIdxBits int       // number of bits per index
	numHash    int       // number of SHA256 hashes needed
	bits       []byte    // bit storage
	hasher     hash.Hash // SHA256 hasher
}

// NewBloomFilterDirect creates a new BloomFilter based on the number of bits
// in the filter and the number of indices to be used.
func NewBloomFilterDirect(numBits, numIdx int) *BloomFilter {
	numIdxBits := int(math.Ceil(math.Log2(float64(numBits))))
	return &BloomFilter{
		numBits:    numBits,
		numIdx:     numIdx,
		numIdxBits: numIdxBits,
		numHash:    (numIdxBits*numIdx + 255) / 256,
		bits:       make([]byte, (numBits+7)/8),
		hasher:     sha256.New(),
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

// Add an entry to the BloomFilter.
func (bf *BloomFilter) Add(entry []byte) {
	for _, idx := range bf.indexList(entry) {
		pos, mask := resolve(idx)
		bf.bits[pos] |= mask
	}
}

// Contains returns true if the BloomFilter contains the given entry, and
// false otherwise. If an entry was added to the set, this function will
// always return 'true'. It can return 'true' for entries not in the set
// ("false-positives").
func (bf *BloomFilter) Contains(entry []byte) bool {
	for _, idx := range bf.indexList(entry) {
		pos, mask := resolve(idx)
		if (bf.bits[pos] & mask) == 0 {
			return false
		}
	}
	return true
}

// Helper method to extract the list of indices for an entry.
func (bf *BloomFilter) indexList(entry []byte) []int {
	totalIdx := make([]byte, 0)
	bf.hasher.Reset()
	for i := 0; i < bf.numHash; i++ {
		bf.hasher.Write(entry)
		totalIdx = bf.hasher.Sum(totalIdx)
	}
	v := new(big.Int).SetBytes(totalIdx)
	mask := big.NewInt((1 << uint(bf.numIdxBits)) - 1)
	list := make([]int, bf.numIdx)
	for i := 0; i < bf.numIdx; i++ {
		j := new(big.Int).And(v, mask)
		list[i] = int(j.Int64()) % bf.numBits
		v = new(big.Int).Rsh(v, uint(bf.numIdxBits))
	}
	return list
}

// Helper method to resolve an index into byte/bit positions in the data
// of the BloomFilter.
func resolve(idx int) (int, byte) {
	return idx >> 3, byte(1 << uint(idx&7))
}
