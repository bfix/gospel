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
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/bfix/gospel/data"
)

const (
	// MAX_MSGSIZE defines the maximum message size
	MAX_MSGSIZE = 65530
)

// Error codes
var (
	ErrNoNodeList = fmt.Errorf("No node list in response")
	ErrNoValue    = fmt.Errorf("No value in response")
	ErrNoData     = fmt.Errorf("No enough data")
)

//----------------------------------------------------------------------
//  message types
//----------------------------------------------------------------------

const (
	// type values for node commands
	//==================================================================
	// IMPORTANT: Requests MUST have odd values, responses MUST be even!
	// (If you have multiple responses for a single request, leave out
	// odd numbers from sequence)
	//==================================================================
	PING            = 1 // PING to check if a node is alive
	PONG            = 2 // response to PING
	STORE           = 3 // STORE a key/value pair in the DHZ
	STORE_RESP      = 4 // response to STORE
	FIND_NODE       = 5 // FIND NODE returns a list of nodes "near" the requested one
	FIND_NODE_RESP  = 6 // response to FIND_NODE
	FIND_VALUE      = 7 // FIND VALUE recursively
	FIND_VALUE_RESP = 8 // response to FIND_VALUE
)

var (
	// human-readable message type names
	msgType = []string{
		"PING", "PONG",
		"STORE", "STORE_RESP",
		"FIND_NODE", "FIND_NODE_RESP",
		"FIND_VALUE", "FIND_VALUE_RESP",
	}

	// Error message
	ErrMessageParse = fmt.Errorf("Failed to parse message")
)

// MsgType returns a human-readable message type
func MsgType(m Message) string {
	hdr := m.Header()
	if i := int(hdr.Type) - 1; i >= 0 && i < len(msgType) {
		return msgType[i]
	}
	return "<Unknown>"
}

//----------------------------------------------------------------------
// Message interface
//----------------------------------------------------------------------

// Message interface
type Message interface {
	Header() *MsgHeader
	Data() []byte
}

//----------------------------------------------------------------------
// Message header (shared by all message types)
//----------------------------------------------------------------------

// HDR_SIZE is the size of the message header in bytes.
const HDR_SIZE = 72

// MsgHeader (common header for requests and responses, 80 bytes)
type MsgHeader struct {
	Size     uint16 `order:"big"` // Size of message (including size)
	Type     uint16 `order:"big"` // Message type (see constants)
	TxId     uint32 `order:"big"` // transaction identifier
	Sender   *Address
	Receiver *Address
}

// Header returns the common message header data structure
func (m *MsgHeader) Header() *MsgHeader {
	return m
}

// Data() returns the binary representation the message
func (m *MsgHeader) Data() []byte {
	buf, _ := data.Marshal(m)
	return buf
}

//======================================================================
// Message factory
//======================================================================

// ParseMessage returns a message implementation from binary data
func ParseMessage(buf []byte) (Message, error) {
	// read the type of the message
	var mt uint16
	binary.Read(bytes.NewBuffer(buf[2:4]), binary.BigEndian, &mt)
	// create empty message of given type
	msg := CreateMessage(mt)
	if msg == nil {
		return nil, ErrMessageParse
	}
	// Unmarshal binary data into message instance
	if err := data.Unmarshal(msg, buf); err != nil {
		return nil, err
	}
	return msg, nil
}

// CreateMessage instantiates an empty message of given type
func CreateMessage(t uint16) Message {
	switch t {
	case PING:
		return NewPingMsg()
	case PONG:
		return NewPongMsg()
	}
	return nil
}

//======================================================================
// Generic message handler list
//======================================================================

// Error codes used
var (
	ErrHandlerUsed   = fmt.Errorf("Handler in use")
	ErrHandlerUnused = fmt.Errorf("Handler not in use")
)

// MessageHandler is a (async) function that handles a message
type MessageHandler func(context.Context, Message) bool

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

// Handle message; returns true if the message was processed (successfully).
func (r *HandlerList) Handle(ctx context.Context, id int, m Message) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// check if handler for message is available
	if hdlr, ok := r.hdlrs[id]; ok {
		// call handler
		return hdlr(ctx, m)
	}
	return false
}
