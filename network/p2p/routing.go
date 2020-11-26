package p2p

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
	"fmt"
	"sync"
	"time"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/math"
)

const (
	// ALPHA is the concurrency parameter
	ALPHA = 3

	// BUCKET_TTL_SECS is the lifetime of a drop in a bucket (in seconds)
	BUCKET_TTL_SECS = 300

	// PING_TIMEOUT is the grace period for "still-alive" checks
	PING_TIMEOUT = 10 * time.Second
)

//----------------------------------------------------------------------
// Address (identifier for nodes and values alike) is a Ed25519 public
// key (32 bytes in size). Routing is based on addresses only.
//----------------------------------------------------------------------

var (
	// ADDR_SIZE is the size of an address in bytes
	ADDR_SIZE uint16 = 32

	// ErrAddressInvalid error message for malformed addresses
	ErrAddressInvalid = fmt.Errorf("Invalid address")
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

// NewAddressFromString recreates an address from string label
func NewAddressFromString(addr string) (*Address, error) {
	if len(addr) != 52 {
		return nil, ErrAddressInvalid
	}
	addr += "===="
	buf, err := base32.StdEncoding.DecodeString(addr)
	if err != nil {
		return nil, err
	}
	if len(buf) != 32 {
		return nil, ErrAddressInvalid
	}
	return NewAddress(buf), nil
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
	return base32.StdEncoding.EncodeToString(a.Data)[:52]
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
// Network endpoint
//----------------------------------------------------------------------

// Endpoint specifies a node on the network
type Endpoint struct {
	Addr *Address // address of node
	Endp *String  // listen address
}

// NewEndpoint creates a new instance for given address,endp pair
func NewEndpoint(addr *Address, endp string) *Endpoint {
	return &Endpoint{
		Addr: addr,
		Endp: NewString(endp),
	}
}

// Size of binary representation
func (e *Endpoint) Size() uint16 {
	return ADDR_SIZE + e.Endp.Size()
}

// String returns human-readable endpoint description
func (e *Endpoint) String() string {
	return fmt.Sprintf("%.8s@%s", e.Addr, e.Endp)
}

//----------------------------------------------------------------------
// Drops in a bucket...
//----------------------------------------------------------------------

// bucket entry
type drop struct {
	addr *Address
	seen time.Time
}

// create a new drop instance
func newDrop(addr *Address) *drop {
	return &drop{
		addr: addr,
		seen: time.Now(),
	}
}

// update drop instance
func (d *drop) update() {
	d.seen = time.Now()
}

// check if a drop has expired
func (d *drop) expired() bool {
	return time.Now().Sub(d.seen).Seconds() > BUCKET_TTL_SECS
}

//----------------------------------------------------------------------
// Routing buckets
//----------------------------------------------------------------------

var (
	// K_BUCKETS is the number of entries in a bucket per address bit.
	K_BUCKETS int = 20
)

// Bucket is used to store nodes depending on their distance to a
// reference node (the local node usually). All addresses in one bucket have
// the same distance value.
// A bucket is ordered: LRU addresses are at the beginning of the list, the
// MRU addresses are at the end (Kademlia scheme)
type Bucket struct {
	num   int        // bucket number (for log purposes)
	addrs []*drop    // list of addresses
	lock  sync.Mutex // lock for list access
	count int        // number of addresses in bucket
}

// NewBucket returns a new bucket of given length.
func NewBucket(n int) *Bucket {
	return &Bucket{
		num:   n,
		addrs: make([]*drop, K_BUCKETS),
		count: 0,
	}
}

// Add address to bucket. Returns false if the bucket is full.
func (b *Bucket) Add(addr *Address) bool {
	if b.count < K_BUCKETS {
		b.lock.Lock()
		b.addrs[b.count] = newDrop(addr)
		b.count++
		b.lock.Unlock()
		return true
	}
	return false
}

// Contains checks if address is already in the bucket and
// returns its index in the list (or -1 if not found)
func (b *Bucket) Contains(addr *Address) int {
	b.lock.Lock()
	defer b.lock.Unlock()

	for i, a := range b.addrs {
		if i == b.count {
			break
		}
		if addr.Equals(a.addr) {
			return i
		}
	}
	return -1
}

// Expired returns true if the indexed drop has expired.
func (b *Bucket) Expired(pos int) bool {
	// check range
	if pos < 0 || pos > b.count-2 {
		// skip if outside of range
		return false
	}
	return b.addrs[pos].expired()
}

// Update changes and moves the address from position to end of list (tail)
// If addr is nil, the currently stored address is moved.
func (b *Bucket) Update(pos int, drop *drop) {
	// check range
	if pos < 0 || pos > b.count-2 {
		// skip if outside of range
		return
	}
	b.lock.Lock()
	if drop == nil {
		drop = b.addrs[pos]
		drop.update()
	}
	copy(b.addrs[pos:b.count-1], b.addrs[pos+1:b.count])
	b.addrs[b.count-1] = drop
	b.lock.Unlock()
	return
}

// Count returns the number of addresses in bucket
func (b *Bucket) Count() int {
	return b.count
}

// MRU returns the most recently used entries in bucket, offs=0 refers to
// latest entry, offs=1 to second-latest and so on. offset must be smaller
// then the number of addresses in the bucket.
func (b *Bucket) MRU(offs int) *Address {
	// check range
	if offs < 0 || offs >= b.count {
		return nil
	}
	// return address
	return b.addrs[b.count-offs-1].addr
}

//----------------------------------------------------------------------

// BucketList is a list of buckets (one for each address bit)
type BucketList struct {
	addr  *Address             // base address for distance
	ping  *PingService         // ping helper
	list  []*Bucket            // distance-indexed buckets
	queue chan *BucketListTask // process queue
}

// NewBucketList returns a new BucketList with given address as reference
// point for distances.
func NewBucketList(addr *Address, ping *PingService) *BucketList {
	bl := &BucketList{
		addr:  addr,
		ping:  ping,
		list:  make([]*Bucket, 256),
		queue: make(chan *BucketListTask, 10), // buffered channel for tasks
	}
	for i := range bl.list {
		bl.list[i] = NewBucket(i)
	}
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
func (bl *BucketList) Add(addr *Address) {
	// compute the distance to reference
	k := addr.Distance(bl.addr).BitLen() - 1
	if k < 0 {
		// no need to add our own address :)
		return
	}
	// check if address is already in the bucket.
	b := bl.list[k]
	if pos := b.Contains(addr); pos != -1 {
		// found: move it to the tail of the list
		b.Update(pos, nil)
		return
	}
	// can we simply add the address to the bucket?
	if !b.Add(addr) {
		// no, we need to process address separately.
		if b.Expired(0) && bl.ping != nil {
			// LRU entry is expired and can be replaced
			bl.queue <- &BucketListTask{
				job:  1,
				addr: addr,
				k:    k,
			}
		}
	}
}

// Closest returns the n closest nodes we know of
// The number of returned nodes can be smaller if the node does not know
// about that many more nodes. Addresses are ordered by distance and MRU.
func (bl *BucketList) Closest(n int) (res []*Address) {
	// collect closest nodes from buckets
	tmp := make([]*Address, 0)
	for _, bkt := range bl.list {
		for i := 0; i < bkt.Count(); i++ {
			tmp = append(tmp, bkt.MRU(i))
			if len(tmp) == n {
				return
			}
		}
	}
	// reverse list (MRU first)
	num := len(tmp)
	res = make([]*Address, num)
	copy(res, tmp)
	for i := num/2 - 1; i >= 0; i-- {
		t := num - 1 - i
		res[i], res[t] = res[t], res[i]
	}
	return
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
					// get address to check from indexed bucker
					buck := bl.list[task.k]
					drop := buck.addrs[0]
					if drop.expired() {
						// ping address with short timeout
						if err := bl.ping.Ping(ctx, task.addr, PING_TIMEOUT, 0); err != nil {
							// ping failed: use new address drop
							drop = newDrop(task.addr)
							// update bucket
							buck.Update(0, drop)
						}
					}
				}

			// externally terminated
			case <-ctx.Done():
				return
			}
		}
	}()
}
