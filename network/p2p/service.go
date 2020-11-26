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
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/bfix/gospel/data"
)

//======================================================================
// Services
//======================================================================

// Error messages
var (
	ErrServiceRequest         = fmt.Errorf("Request failed or invalid")
	ErrServiceRequestUnknown  = fmt.Errorf("Unknown request")
	ErrServiceResponseUnknown = fmt.Errorf("Unknown response")
)

//----------------------------------------------------------------------
// Interface
//----------------------------------------------------------------------

// Service is running on a node and provides:
// - a responder for requests with matching message type
// - a listener for service messages with matching TxId
// Service is the building block for upper levels of the protocol stack.
type Service interface {
	// Name is a human-readble and short service description like "PING"
	Name() string

	// Node the service is running on
	Node() *Node

	// Respond to a service request from peer.
	// The bool return value indicates whether the message has been processed
	// or not and if there was an error message with it.
	Respond(context.Context, Message) (bool, error)

	// Listen to service responses from peer.
	// The bool return value indicates whether the message has been processed
	// or not and if there was an error message with it.
	Listen(context.Context, Message) (bool, error)

	// Send a message to network
	Send(context.Context, Message) error

	// NewMessage creates an empty service message of given type.
	// The type must be known to the service.
	NewMessage(mt int) Message

	// Link service to node it is running on. Package-local method
	// can only be called in the "p2p" namespace (by a Node).
	link(*Node) bool
}

//----------------------------------------------------------------------
// Base implementation
//----------------------------------------------------------------------

// ServiceImpl is a basic implementation of a Service instance and the
// building block for custom services implementing new functionality.
type ServiceImpl struct {
	// back reference to node the service is running on
	// (set by the node while registering a service)
	node *Node

	responders *HandlerList       // list of responders
	listeners  *HandlerList       // list of listeners
	factories  map[int]NewMessage // map of message instantiators
	labels     map[int]string     // map of message type labels
}

// NewServiceImpl returns a new basic service instance (NOP)
func NewServiceImpl() *ServiceImpl {
	return &ServiceImpl{
		node:       nil,
		responders: NewHandlerList(),
		listeners:  NewHandlerList(),
		factories:  make(map[int]NewMessage),
		labels:     make(map[int]string),
	}
	return nil
}

// Respond to a service request
func (s *ServiceImpl) Respond(ctx context.Context, msg Message) (bool, error) {
	return s.responders.Handle(ctx, int(msg.Header().Type), msg)
}

// Listen to response message
func (s *ServiceImpl) Listen(ctx context.Context, msg Message) (bool, error) {
	return s.listeners.Handle(ctx, int(msg.Header().TxId), msg)
}

// Send a message from this node to the network
func (s *ServiceImpl) Send(ctx context.Context, msg Message) error {
	return s.node.Send(ctx, msg)
}

// Node returns the node the service is running on
func (s *ServiceImpl) Node() *Node {
	return s.node
}

// MessageFactory returns an empty message for a given type
func (s *ServiceImpl) MessageFactory(mt int) Message {
	// check if we handle this message type
	if fac, ok := s.factories[mt]; ok {
		// yes: return new message
		return fac()
	}
	return nil
}

// Internal: link node to service
func (s *ServiceImpl) link(n *Node) bool {
	if s.node != nil {
		return false
	}
	s.node = n
	return true
}

//======================================================================
// Tasks
//======================================================================

// TaskHandler is used to notify listener of messages during task processing
type TaskHandler struct {

	// Message handler for responses in task
	msgHdlr MessageHandler

	// Timeout duration
	timeout time.Duration
}

// Task is a generic wrapper for tasks running on a local node. It sends
// an initial message 'm'; responses are handled by 'f'. If 'f' returns
// true, no further responses from the receiver are expected.
func (s *ServiceImpl) Task(ctx context.Context, m Message, f *TaskHandler) (err error) {
	// register for responses
	ctrl := make(chan int)
	txid := int(m.Header().TxId)
	s.listeners.Add(txid, func(ctx context.Context, m Message) (bool, error) {
		// call the "real" handler
		rc, err := f.msgHdlr(ctx, m)
		if rc {
			// no more responses expected
			ctrl <- 0
		}
		return rc, err
	})
	// send message
	ctxSend, cancel := context.WithDeadline(ctx, time.Now().Add(f.timeout))
	defer cancel()
	if err = s.node.Send(ctxSend, m); err != nil {
		return
	}
	// timeout for response(s)
	tick := time.Tick(f.timeout)
	select {
	case <-tick:
		err = ErrNodeTimeout
	case <-ctrl:
	}
	// unregister handler
	s.listeners.Remove(txid)
	return
}

//======================================================================
// List of services for a node
//======================================================================

// ServiceList for all services registered on a node
type ServiceList struct {
	srvcs []Service
}

// NewServiceList returns a new list instance
func NewServiceList() *ServiceList {
	return &ServiceList{
		srvcs: make([]Service, 0),
	}
}

// MessageFactory re-creates a message from binary data.
func (sl *ServiceList) MessageFactory(buf []byte) (Message, error) {
	// read the type of the message
	var mt uint16
	binary.Read(bytes.NewBuffer(buf[2:4]), binary.BigEndian, &mt)

	// create empty message of given type
	for _, srv := range sl.srvcs {
		msg := srv.NewMessage(int(mt))
		if msg != nil {
			// parse binary data
			if err := data.Unmarshal(msg, buf); err != nil {
				return nil, err
			}
			return msg, nil
		}
	}
	// we failed to re-create the message
	return nil, ErrMessageParse
}

// Add service to list
func (sl *ServiceList) Add(s Service) bool {
	name := s.Name()
	for _, ss := range sl.srvcs {
		if ss.Name() == name {
			return false
		}
	}
	sl.srvcs = append(sl.srvcs, s)
	return true
}

// Get service from list
func (sl *ServiceList) Get(name string) Service {
	for _, s := range sl.srvcs {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

// Respond to service requests
func (sl *ServiceList) Respond(ctx context.Context, msg Message) (bool, error) {
	hdr := msg.Header()
	rcv := hdr.Receiver.String()
	log.Printf("[%.8s] Received request: %s\n", rcv, msg)

	for _, srvc := range sl.srvcs {
		ok, err := srvc.Respond(ctx, msg)
		if err != nil {
			// request processing failed
			log.Printf("[%.8s] Request processing failed: %s\n", rcv, err.Error())
		}
		if ok {
			// request handled
			return true, err
		}
	}
	log.Printf("[%.8s] Unknown request type '%d'...\n", rcv, hdr.Type)
	return false, ErrServiceRequestUnknown

}

// Listen to service responses
func (sl *ServiceList) Listen(ctx context.Context, msg Message) (bool, error) {
	hdr := msg.Header()
	rcv := hdr.Receiver.String()
	log.Printf("[%.8s] Received response: %s\n", rcv, msg)

	for _, srvc := range sl.srvcs {
		ok, err := srvc.Listen(ctx, msg)
		if err != nil {
			// response processing failed
			log.Printf("[%.8s] Response processing failed: %s\n", rcv, err.Error())
			return false, err
		}
		if ok {
			// response successfully handled
			return true, nil
		}
	}
	log.Printf("[%.8s] Unknown response for txId '%d'...\n", rcv, hdr.TxId)
	return false, ErrServiceResponseUnknown
}
