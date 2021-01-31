package network

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix
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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/math"
	"golang.org/x/crypto/sha3"
)

//======================================================================
// Tor onion handling (hidden services)
//======================================================================

// Error codes
var (
	ErrOnionAlreadyRunning = fmt.Errorf("Onion already running")
	ErrOnionInvalidKey     = fmt.Errorf("Invalid onion key specification")
	ErrOnionKeyInvalidSize = fmt.Errorf("Invalid onion key size")
	ErrOnionKeyExists      = fmt.Errorf("Onion private key already exists")
	ErrOnionMissingKey     = fmt.Errorf("Missing private key for onion")
	ErrOnionInvalidKeySpec = fmt.Errorf("Invalid provate key specification")
	ErrOnionAddFailed      = fmt.Errorf("Failed to add hidden service")
)

//----------------------------------------------------------------------
// Onion (hidden service)
//----------------------------------------------------------------------

// Onion is a hidden service implementation on a Tor service
type Onion struct {
	key     interface{}    // private key
	flags   []string       // hidden service flags
	ports   map[int]string // port mappings
	srvId   string         // service identifier
	running bool           // hidden service running?
}

// NewOnion instantiates a new hidden service
func NewOnion(key interface{}) (o *Onion, err error) {
	o = &Onion{
		key:     key,
		flags:   make([]string, 0),
		ports:   make(map[int]string),
		running: false,
	}
	o.srvId, err = o.ServiceID()
	return
}

// ServiceID returns the "onion name" of the hidden service (without the
// trailing ".onion")
func (o *Onion) ServiceID() (id string, err error) {
	switch prv := o.key.(type) {
	case *ed25519.PrivateKey:
		keyData := prv.Public().Bytes()
		hsh := sha3.New256()
		hsh.Write([]byte(".onion checksum"))
		hsh.Write(keyData)
		hsh.Write([]byte{0x03})
		sum := hsh.Sum(nil)
		sum[2] = 0x03
		id = base32.StdEncoding.EncodeToString(append(keyData, sum[:3]...))
	case *rsa.PrivateKey:
		pub := prv.Public().(*rsa.PublicKey)
		var data []byte
		if data, err = asn1.Marshal(*pub); err != nil {
			return
		}
		sum := sha1.Sum(data)
		id = base32.StdEncoding.EncodeToString(sum[:len(sum)/2])
	}
	id = strings.ToLower(id)
	return
}

// AddFlag adds service flags
func (o *Onion) AddFlag(flags ...string) {
	for _, flag := range flags {
		o.flags = append(o.flags, flag)
	}
}

// AddPort adds a port mapping for the hidden service
func (o *Onion) AddPort(listen int, spec string) {
	o.ports[listen] = spec
}

// Start a new hidden service via a Tor controller.
func (o *Onion) Start(ctrl *Control) (err error) {
	// check if hidden service is already active
	if o.running {
		return ErrOnionAlreadyRunning
	}
	// assemble control command to start hidden service
	cmd := "ADD_ONION "

	// specify private service key
	switch prv := o.key.(type) {
	case *ed25519.PrivateKey:
		kd := make([]byte, 64)
		if _, err = rand.Read(kd); err != nil {
			return
		}
		copy(kd[:32], prv.D.Bytes())
		cmd += fmt.Sprintf("ED25519-V3:%s", base64.StdEncoding.EncodeToString(kd))
	case *rsa.PrivateKey:
		if prv.Size() != 128 {
			return ErrOnionKeyInvalidSize
		}
		kd := x509.MarshalPKCS1PrivateKey(prv)
		cmd += fmt.Sprintf("RSA1024:%s", base64.StdEncoding.EncodeToString(kd))
	case string:
		if prv != "ED25519-V3" && prv != "RSA1024" {
			return ErrOnionInvalidKey
		}
		cmd += fmt.Sprintf("NEW:%s", prv)
	default:
		return ErrOnionInvalidKey
	}
	// add flags (optional)
	if len(o.flags) > 0 {
		cmd += " Flags="
		for i, flag := range o.flags {
			if i > 0 {
				cmd += ","
			}
			cmd += flag
		}
	}
	// add port mappings
	for listen, tgt := range o.ports {
		cmd += fmt.Sprintf(" Port=%d,%s", listen, tgt)
	}
	// execute command
	list, err := ctrl.execute(cmd)
	if err != nil {
		return err
	}
	// get newly generated key (optional)
	if kd, ok := list["PrivateKey"]; ok {
		parts := strings.Split(kd, ":")
		var data []byte
		if spec, ok := o.key.(string); ok {
			if len(parts) != 2 || parts[1] != spec {
				return ErrOnionInvalidKeySpec
			}
			if data, err = base64.StdEncoding.DecodeString(parts[1]); err != nil {
				return
			}
			switch spec {
			case "ED25519-V3":
				d := math.NewIntFromBytes(data[:32])
				o.key = ed25519.NewPrivateKeyFromD(d)
			case "RSA1024":
				if o.key, err = x509.ParsePKCS1PrivateKey(data); err != nil {
					return
				}
			}
		}
	}
	// get generated service identifier
	srvId, err := o.ServiceID()
	if err != nil {
		return
	}
	if id, ok := list["ServiceID"]; ok {
		if id != srvId {
			return ErrOnionAddFailed
		}
	}
	o.running = true
	return nil
}

// Stop removes a hidden service from the Tor service
func (o *Onion) Stop(ctrl *Control) error {
	id, err := o.ServiceID()
	if err == nil {
		cmd := fmt.Sprintf("DEL_ONION %s", id)
		_, err = ctrl.execute(cmd)
	}
	return err
}
