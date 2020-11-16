package dht

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
	"context"
	"encoding/base32"
	"log"
	"sync"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/math"
)

//----------------------------------------------------------------------
// Address (identifier for nodes and values alike) is a Ed25519 public
// key (32 bytes in size). Routing is based on addresses only.
//----------------------------------------------------------------------

var (
	// ADDR_SIZE is the size of an address in bytes
	ADDR_SIZE = 32
)

// Address encapsulates data representing the object identifier.
type Address struct {
	Data []byte `size:"32"` // address data of size "(ADDRESS_BITS+7)/8"
}

// NewAddressFromKey generate an address from public key
func NewAddressFromKey(pub *ed25519.PublicKey) *Address {
	return NewAddress(pub.Bytes())
}

// NewAddress creates a new address from a binary object
func NewAddress(b []byte) *Address {
	if b == nil {
		b = make([]byte, 32)
	} else if len(b) != 32 {
		panic("Invalid address data")
	}
	return &Address{
		Data: b,
	}
}

// Distance returns the distance between two addresses. The distance metric is
// based on XOR'ing the address values.
func (a *Address) Distance(b *Address) *math.Int {
	buf := make([]byte, len(a.Data))
	for i, v := range a.Data {
		buf[i] = v ^ b.Data[i]
	}
	return math.NewIntFromBytes(buf)
}

// String returns a human-readable address
func (a *Address) String() string {
	return base32.StdEncoding.EncodeToString(a.Data)
}

// Equals checks if two addresses are the same
func (a *Address) Equals(o *Address) bool {
	return bytes.Equal(a.Data, o.Data)
}

// PublicKey returns the public node key (from its address)
func (a *Address) PublicKey() *ed25519.PublicKey {
	return ed25519.NewPublicKeyFromBytes(a.Data)
}

//----------------------------------------------------------------------
// Routing buckets
//----------------------------------------------------------------------

var (
	// K_BUCKETS is the number of entries in a bucket per address bit.
	K_BUCKETS int = 20
)

// Bucket is used to store nodes depending on their distance to a
// reference node (the local node usually). A bucket is ordered: Least recently
// seen addresses are at the beginning of the list, the most recent entries at
// the end.
type Bucket struct {
	node []*Address
}

// NewBucket returns a new bucket of given length.
func NewBucket() *Bucket {
	return &Bucket{
		node: make([]*Address, 0, K_BUCKETS),
	}
}

// BucketList is a list of buckets (one for each address bit)
type BucketList struct {
	list  []*Bucket
	lock  sync.Mutex
	queue chan *BucketListTask
}

// NewBucketList returns a new BucketList.
func NewBucketList() *BucketList {
	bl := &BucketList{
		list: make([]*Bucket, 256),
	}
	for i := range bl.list {
		bl.list[i] = NewBucket()
	}
	// use a buffered channel for tasks
	bl.queue = make(chan *BucketListTask, 10)
	return bl
}

// BucketListTask describes a maintenance job on the bucket list
type BucketListTask struct {
	// what to do:
	// 0 = delete address
	// 1 = add address (if oldest peer is unresponsive)
	job int

	// Address to be processed (with bucket number)
	addr *Address
	k    int
}

// Add a new peer to the routing table (possibly)
func (bl *BucketList) Add(k int, addr *Address) {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	// check if address is already in the bucket.
	list := bl.list[k].node
	num := len(list)
	for i, a := range bl.list[k].node {
		if addr.Equals(a) {
			// found: move it to the tail of the list
			copy(list[i:num-1], list[i+1:num])
			list[num-1] = a
			log.Printf("[buckets ] Moved '%.8s' from pos %d:%d to tail\n", addr, k, i)
			return
		}
	}
	// can we simply add the address to the bucket?
	if num < K_BUCKETS {
		// yes, we can
		list = append(list, addr)
		log.Printf("[buckets ] Appended '%.8s' in bucket #%d\n", addr, k)
		return
	}
	// we need to process address serialized.
	bl.queue <- &BucketListTask{
		job:  1,
		addr: addr,
		k:    k,
	}
}

// Run the processing loop for the bucket list.
func (bl *BucketList) Run(ctx context.Context) {
	go func() {
		for {
			select {
			// process new addresses
			case task := <-bl.queue:
				switch task.job {
				// delete address from bucket
				case 0:
					panic("not implemented")

				// add address if oldest peer is unresponsive
				case 1:
					bl.lock.Lock()
					//PingTask(ctx context.Context, rcv *Address, timeout time.Duration)
					bl.lock.Unlock()
				}

			// externally terminated
			case <-ctx.Done():
				return
			}
		}
	}()
}
