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
	"context"
	"errors"
	"sync"

	"github.com/bfix/gospel/data"
)

const (
	// MaxMsgSize defines the maximum message size
	MaxMsgSize = 65530
)

//----------------------------------------------------------------------
//  message types and flags
//----------------------------------------------------------------------

// type values for node commands
const (
	//==================================================================
	// IMPORTANT: Requests MUST have odd values, responses MUST be even!
	// (If you have multiple responses for a single request, leave out
	// odd numbers from sequence)
	//==================================================================
	ReqPING  = 1 // PING to check if a node is alive
	RespPING = 2 // response to PING
	ReqNODE  = 3 // FIND_NODE returns a list of nodes "near" the requested one
	RespNODE = 4 // response to FIND_NODE
	ReqRELAY = 5 // relay message to another node (response-less)
)

// message flags
const (
	MsgfRelay = 1 // Message was forwarded (sender != originator)
	MsgfDrop  = 2 // Drop message without processing (cover traffic)
)

var (
	// ErrMessageParse if message parsing from binary data failed
	ErrMessageParse = errors.New("failed to parse message")
)

//----------------------------------------------------------------------
// Helper types for message fields
//----------------------------------------------------------------------

// String is a sequence of unicode runes in binary format
type String struct {
	Len  uint16 `order:"big"` // length of string
	Data []byte `size:"Len"`  // string data
}

// NewString encapsulates a standard string
func NewString(s string) *String {
	buf := []byte(s)
	return &String{
		Len:  uint16(len(buf)),
		Data: buf,
	}
}

// Size returns the total size of String in bytes
func (s *String) Size() uint16 {
	return s.Len + 2
}

// String returns the standard string object
func (s *String) String() string {
	return string(s.Data)
}

//----------------------------------------------------------------------
// Message interface
//----------------------------------------------------------------------

// Message interface
type Message interface {
	// Header of message (common data)
	Header() *MsgHeader

	// Data returns the binary representation of the message
	Data() []byte

	// String returns a human-readable message
	String() string
}

//----------------------------------------------------------------------
// Message header (shared by all message types)
//----------------------------------------------------------------------

// HdrSize is the size of the message header in bytes.
const HdrSize = 80

// MsgHeader (common header for requests and responses, 80 bytes)
type MsgHeader struct {
	Size     uint16 `order:"big"` // Size of message (including size)
	Type     uint16 `order:"big"` // Message type (see constants)
	Flags    uint32 `order:"big"` // Message flags (see constants)
	TxID     uint64 `order:"big"` // transaction identifier
	Sender   *Address
	Receiver *Address
}

// Header returns the common message header data structure
func (m *MsgHeader) Header() *MsgHeader {
	return m
}

// Data returns the binary representation the message
func (m *MsgHeader) Data() []byte {
	buf, _ := data.Marshal(m)
	return buf
}

//======================================================================
// Message factory
//======================================================================

// MessageFactory reassembles messages from binary data
type MessageFactory func([]byte) (Message, error)

// NewMessage returns a message (of specific type)
type NewMessage func() Message

//======================================================================
// Generic message handler list
//======================================================================

// Error codes used
var (
	ErrHandlerUsed   = errors.New("handler in use")
	ErrHandlerUnused = errors.New("handler not in use")
)

// MessageHandler is a (async) function that handles a message
type MessageHandler func(context.Context, Message) (bool, error)

// HandlerList maps an integer to a message handler instance.
type HandlerList struct {
	lock  sync.RWMutex           // lock for concurrent operations
	hdlrs map[int]MessageHandler // handler functions for given int value
}

// NewHandlerList instantiates a new list of handler mappings.
func NewHandlerList() *HandlerList {
	return &HandlerList{
		hdlrs: make(map[int]MessageHandler),
	}
}

// Add a message handler for given integer to list.
func (r *HandlerList) Add(id int, f MessageHandler) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// check if handler is already set
	if _, ok := r.hdlrs[id]; ok {
		return ErrHandlerUsed
	}
	// add handler to list
	r.hdlrs[id] = f
	return nil
}

// Remove handler for given integer from list.
func (r *HandlerList) Remove(id int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// check if listener is set
	if _, ok := r.hdlrs[id]; !ok {
		return ErrHandlerUnused
	}
	// remove handler
	delete(r.hdlrs, id)
	return nil
}

// Handle message
func (r *HandlerList) Handle(ctx context.Context, id int, m Message) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// check if handler for message is available
	if hdlr, ok := r.hdlrs[id]; ok {
		// call handler
		return hdlr(ctx, m)
	}
	return false, nil
}
