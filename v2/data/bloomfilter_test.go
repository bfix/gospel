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
