package p2p

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bfix/gospel/data"
	gerr "github.com/bfix/gospel/errors"
	"github.com/bfix/gospel/logger"
)

// Error codes
var (
	ErrLookupFailed = errors.New("Lookup failed")
)

//----------------------------------------------------------------------
// FIND_NODE message (request)
//----------------------------------------------------------------------

// FindNodeMsg for FIND_NODE requests
type FindNodeMsg struct {
	MsgHeader

	Addr *Address // address of node we are looking for
}

// String returns human-readable message
func (m *FindNodeMsg) String() string {
	return fmt.Sprintf("FIND_NODE(%.8s){%.8s -> %.8s, #%d}", m.Addr, m.Sender, m.Receiver, m.TxID)
}

// NewFindNodeMsg creates an empty FIND_NODE message
func NewFindNodeMsg() Message {
	return &FindNodeMsg{
		MsgHeader: MsgHeader{
			Size:     HdrSize + AddrSize,
			TxID:     0,
			Type:     ReqNODE,
			Flags:    0,
			Sender:   nil,
			Receiver: nil,
		},
		Addr: nil,
	}
}

// Set the additional address field
func (m *FindNodeMsg) Set(addr *Address) *FindNodeMsg {
	m.Addr = addr
	return m
}

//----------------------------------------------------------------------
// FIND_NODE_RESP message (response)
//----------------------------------------------------------------------

// FindNodeRespMsg for FIND_NODE responses
type FindNodeRespMsg struct {
	MsgHeader

	// List of response entries
	List []*Endpoint `size:"*"`
}

// String returns human-readable message
func (m *FindNodeRespMsg) String() string {
	return fmt.Sprintf("FIND_NODE_RESP{%.8s -> %.8s, #%d}[%d]", m.Sender, m.Receiver, m.TxID, len(m.List))
}

// NewFindNodeRespMsg creates an empty FIND_NODE response
func NewFindNodeRespMsg() Message {
	return &FindNodeRespMsg{
		MsgHeader: MsgHeader{
			Size:     HdrSize,
			TxID:     0,
			Type:     RespNODE,
			Flags:    0,
			Sender:   nil,
			Receiver: nil,
		},
		List: make([]*Endpoint, 0),
	}
}

// Add an lookup entry to the list
func (m *FindNodeRespMsg) Add(e *Endpoint) {
	m.List = append(m.List, e)
	m.Size += e.Size()
}

//----------------------------------------------------------------------
// Lookup service
//----------------------------------------------------------------------

// LookupService to resolve node addresses (routing)
type LookupService struct {
	ServiceImpl
}

// NewLookupService creates a new service instance
func NewLookupService() *LookupService {
	srv := &LookupService{
		ServiceImpl: *NewServiceImpl(),
	}
	// defined message instantiators
	srv.factories[ReqNODE] = NewFindNodeMsg
	srv.factories[RespNODE] = NewFindNodeRespMsg

	// defined known labels
	srv.labels[ReqNODE] = "FIND_NODE"
	srv.labels[ReqNODE] = "FIND_NODE_RESP"
	return srv
}

// Name is a human-readble and short service description like "PING"
func (s *LookupService) Name() string {
	return "lookup"
}

// Respond to a service request from peer.
func (s *LookupService) Respond(ctx context.Context, m Message) (bool, error) {
	// check we are responsible for this
	hdr := m.Header()
	if hdr.Type != ReqNODE {
		return false, nil
	}
	// cast will succeed because type of message is checked
	msg, _ := m.(*FindNodeMsg)

	// assemble FIND_NODE_RESP message
	resp, _ := NewFindNodeRespMsg().(*FindNodeRespMsg)
	resp.TxID = hdr.TxID
	resp.Sender = hdr.Receiver
	resp.Receiver = hdr.Sender

	// get address to lookup and try to resolve it locally.
	addr := msg.Addr
	netw := s.Node().Resolve(addr)
	if netw != nil {
		// we know the address being resolved: send it
		// as the only result in the response
		resp.Add(&Endpoint{
			Addr: addr,
			Endp: NewString(netw.String()),
		})
	} else {
		// return closest nodes in our routing table
		for _, addr := range s.Node().Closest(KBuckets) {
			netw = s.Node().Resolve(addr)
			resp.Add(&Endpoint{
				Addr: addr,
				Endp: NewString(netw.String()),
			})
		}
	}
	// send message
	return true, s.Send(ctx, resp)
}

// NewMessage creates an empty service message of given type
func (s *LookupService) NewMessage(mt int) Message {
	switch mt {
	case ReqNODE:
		return NewFindNodeMsg()
	case RespNODE:
		return NewFindNodeRespMsg()
	}
	return nil
}

//----------------------------------------------------------------------
// lookup task
//----------------------------------------------------------------------

// Request resolves an address 'addr' on peer 'rcv' synchronously
func (s *LookupService) Request(ctx context.Context, rcv, addr *Address, timeout time.Duration) (res []*Endpoint, err error) {
	// assemble request
	req, _ := NewFindNodeMsg().(*FindNodeMsg)
	req.TxID = s.node.NextID()
	req.Sender = s.node.Address()
	req.Receiver = rcv
	req.Addr = addr

	// send request and process responses
	hdlr := &TaskHandler{
		msgHdlr: func(ctx context.Context, m Message) (bool, error) {
			// handle FIND_NODE response
			switch x := m.(type) {
			case *FindNodeRespMsg:
				res = x.List
				return true, nil
			}
			return false, nil
		},
		timeout: timeout,
	}
	err = s.Task(ctx, req, hdlr)
	return
}

// Query remote peer for given address; result depends on query implementation.
// It is either an error, a boolean "done" signal (no result) or a result
// instance. In this service the "final" result is of type Endoint; other
// services (like DHT) can use their own results (Value). If the result is an
// address list, the referenced nodes are queried for a result.
type Query func(ctx context.Context, peer, addr *Address) interface{}

// LookupNode a node endpoint address.
func (s *LookupService) LookupNode(ctx context.Context, addr *Address, timeout time.Duration) (entry *Endpoint, err error) {
	sAddr := s.Node().Address()
	logger.Printf(logger.INFO, "[%.8s] Lookup for '%.8s':\n", sAddr, addr)

	query := func(ctx context.Context, peer, addr *Address) interface{} {
		// perform query
		logger.Printf(logger.INFO, "[%.8s] Lookup for '%.8s' on '%.8s'...\n", sAddr, addr, peer)
		var list []*Endpoint
		if list, err = s.Request(ctx, peer, addr, timeout); err != nil {
			return gerr.New(ErrLookupFailed, "[%.8s] Lookup for '%.8s' on '%.8s'", sAddr, addr, peer)
		}
		// learn all entries
		peers := make([]*Address, 0)
		for _, e := range list {
			peers = append(peers, e.Addr)
			_ = s.Node().Learn(e.Addr, e.Endp.String())
		}
		// check if we got a final result
		if len(list) == 1 && list[0].Addr.Equals(addr) {
			// node endpoint is resolved
			return list[0]
		}
		// otherwise return the list of closer nodes.
		return peers
	}
	// call the resolver
	var res interface{}
	res, err = s.Lookup(ctx, addr, query, timeout)
	if err != nil {
		return
	}
	if res == nil {
		return
	}
	return res.(*Endpoint), nil
}

// Lookup with specific resolver logic to handle mutlitple lookup scenarios.
func (s *LookupService) Lookup(ctx context.Context, addr *Address, resolver Query, timeout time.Duration) (res interface{}, err error) {
	// create internal state
	wg := new(sync.WaitGroup)
	bf := data.NewBloomFilter(1000, 1e-5)
	ctxLookup, cancel := context.WithTimeout(ctx, timeout)
	running := true

	// Ask a bunch of peers in parallel to resolve address
	var query func(peer, addr *Address, ch chan interface{})
	queryPeers := func(peers []*Address, addr *Address, ch chan interface{}) {
		// query only limited number of peers
		if len(peers) > Alpha {
			peers = peers[:Alpha]
		}
		// start query for all of them
		for _, peer := range peers {
			go query(peer, addr, ch)
		}
	}
	// Query single peer to resolve given address
	query = func(peer, addr *Address, ch chan interface{}) {
		if !running {
			return
		}
		if !bf.Contains(peer.Data) {
			bf.Add(peer.Data)
			wg.Add(1)
			defer wg.Done()
			switch x := resolver(ctxLookup, peer, addr).(type) {
			// process list recursively
			case []*Address:
				go queryPeers(x, addr, ch)
			// return result (or error) in other cases
			default:
				ch <- x
			}
		}
	}
	// start resolver with closest nodes
	closest := s.Node().Closest(KBuckets)
	out := make(chan interface{})
	defer close(out)
	for {
		// query the next group of peers
		go queryPeers(closest, addr, out)

		// wait for all queries to end
		go func() {
			wg.Wait()
			out <- true
		}()

		// wait for final result, error or unresolved event
		select {
		case in := <-out:
			switch x := in.(type) {
			// leave with error message
			case error:
				cancel()
				running = false
				return nil, x
			// all processing done but no result (unresolved)
			case bool:
				// re-try with next set of closest
				pending := len(closest)
				if pending == 0 {
					cancel()
					running = false
					return nil, ErrLookupFailed
				}
				if pending > Alpha {
					closest = closest[:Alpha]
				}
				continue
			// leave with final lookup result
			default:
				cancel()
				running = false
				return in, nil
			}
		// externally cancelled
		case <-ctx.Done():
			cancel()
			running = false
			return nil, ErrNodeTimeout
		}
	}
}
