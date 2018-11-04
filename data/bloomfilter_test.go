package data

import (
	"bytes"
	"crypto/rand"
	"sort"
	"testing"
)

type Entry []byte

type EntryList []Entry

func (list EntryList) Len() int           { return len(list) }
func (list EntryList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }
func (list EntryList) Less(i, j int) bool { return bytes.Compare(list[i], list[j]) < 0 }

func (list EntryList) Contains(e Entry) bool {
	size := len(list)
	i := sort.Search(size, func(i int) bool { return bytes.Compare(list[i], e) >= 0 })
	return i != size
}

func TestBloomfilter(t *testing.T) {

	n := 500
	fpRate := 0.0001

	// generate positives (entries in the set)
	positives := make(EntryList, n)
	for i := 0; i < n; i++ {
		data := make(Entry, 32)
		if _, err := rand.Read(data); err != nil {
			t.Fatal(err)
		}
		positives[i] = data
	}
	sort.Sort(positives)

	// generate negatives (entries outside the set)
	negatives := make(EntryList, n)
	for i := 0; i < n; {
		data := make(Entry, 32)
		if _, err := rand.Read(data); err != nil {
			t.Fatal(err)
		}
		if !positives.Contains(data) {
			negatives[i] = data
			i++
		}
	}

	// create BloomFilter
	bf := NewBloomFilter(n, fpRate)

	// add positives to bloomfilter
	for _, e := range positives {
		bf.Add(e)
	}

	// check lookup of positives
	count := 0
	for _, e := range positives {
		if !bf.Contains(e) {
			count++
		}
	}
	if count > 0 {
		t.Fatalf("FAILED with %d false-negatives", count)
	}

	// check lookup of negatives
	count = 0
	for _, e := range negatives {
		if bf.Contains(e) {
			count++
		}
	}
	fpReal := float64(count) / float64(n)
	if fpReal > fpRate {
		t.Logf("false-positive rate %f > %f", fpReal, fpRate)
	}
}
