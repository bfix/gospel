package tor

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

	"github.com/bfix/gospel/v2/crypto/ed25519"
	"github.com/bfix/gospel/v2/math"
	"golang.org/x/crypto/sha3"
)

//======================================================================
// Tor onion handling (hidden services)
//======================================================================

// Error codes
var (
	ErrOnionAlreadyRunning = fmt.Errorf("Onion already running")
	ErrOnionNotRunning     = fmt.Errorf("Onion not running")
	ErrOnionInvalidKey     = fmt.Errorf("Invalid onion key")
	ErrOnionInvalidKeyType = fmt.Errorf("Invalid onion key type")
	ErrOnionKeyInvalidSize = fmt.Errorf("Invalid onion key size")
	ErrOnionKeyExists      = fmt.Errorf("Onion private key already exists")
	ErrOnionMissingKey     = fmt.Errorf("Missing private key for onion")
	ErrOnionInvalidKeySpec = fmt.Errorf("Invalid private key specification")
	ErrOnionAddFailed      = fmt.Errorf("Failed to add hidden service")
)

//----------------------------------------------------------------------
// Onion (hidden service)
//----------------------------------------------------------------------

// Onion is a hidden service implementation on a Tor service
type Onion struct {
	key        any            // private key
	flags      []string       // hidden service flags
	ports      map[int]string // port mappings
	srvID      string         // service identifier
	maxStreams int            // max. number of allowed streams
	userName   string         // user name (basic auth)
	userPasswd string         // user password (basic auth)
	running    bool           // hidden service running?
}

// NewOnion instantiates a new hidden service
func NewOnion(key any) (o *Onion, err error) {
	o = &Onion{
		key:        key,
		flags:      make([]string, 0),
		ports:      make(map[int]string),
		maxStreams: 0,
		userName:   "",
		userPasswd: "",
		running:    false,
	}
	o.srvID, err = o.ServiceID()
	return
}

// ServiceID returns the "onion name" of the hidden service (without the
// trailing ".onion")
func (o *Onion) ServiceID() (id string, err error) {
	switch prv := o.key.(type) {
	case *ed25519.PrivateKey:
		return ServiceID(prv.Public())
	case *rsa.PrivateKey:
		return ServiceID(prv.Public().(*rsa.PublicKey))
	}
	return "", ErrOnionInvalidKey
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

// SetCredentials sets username and password for basic authentication
func (o *Onion) SetCredentials(name, passwd string) {
	o.userName = name
	o.userPasswd = passwd
}

// Start a new hidden service via a Tor service.
func (o *Onion) Start(srv *Service) (err error) {
	// check if hidden service is already active
	if o.running {
		return ErrOnionAlreadyRunning
	}
	// assemble control command to start hidden service
	cmd := "ADD_ONION "

	// assemble private service key blob
	switch prv := o.key.(type) {
	case *ed25519.PrivateKey:
		kd := make([]byte, 64)
		// fill right half with random data
		if _, err = rand.Read(kd[32:]); err != nil {
			return
		}
		// left half is private scalar in little endian order
		data := prv.D.Bytes()
		dl := len(data)
		for i, b := range data {
			kd[dl-i-1] = b
		}
		cmd += fmt.Sprintf("ED25519-V3:%s", base64.StdEncoding.EncodeToString(kd))
	case *rsa.PrivateKey:
		if prv.Size() != 128 {
			return ErrOnionKeyInvalidSize
		}
		kd := x509.MarshalPKCS1PrivateKey(prv)
		cmd += fmt.Sprintf("RSA1024:%s", base64.StdEncoding.EncodeToString(kd))
	case string:
		if prv != "ED25519-V3" && prv != "RSA1024" {
			return ErrOnionInvalidKeyType
		}
		cmd += fmt.Sprintf("NEW:%s", prv)
	default:
		return ErrOnionInvalidKey
	}
	// add flags (optional)
	limitStreams := false
	withAuth := false
	if len(o.flags) > 0 {
		cmd += " Flags="
		for i, flag := range o.flags {
			if i > 0 {
				cmd += ","
			}
			cmd += flag
			switch flag {
			case "BasicAuth":
				if len(o.userName) == 0 {
					o.userName = "none"
				}
				if len(o.userPasswd) == 0 {
					o.userPasswd = "none"
				}
				withAuth = true
			case "MaxStreamsCloseCircuit":
				limitStreams = true
			}
		}
	}
	// add port mappings
	for listen, tgt := range o.ports {
		cmd += fmt.Sprintf(" Port=%d,%s", listen, tgt)
	}
	// set max. number of streams
	if limitStreams {
		cmd += fmt.Sprintf(" MaxStreams=%d", o.maxStreams)
	}
	// set credentials for basic authentication
	if withAuth {
		cmd += fmt.Sprintf(" ClientAuth=%s:%s", o.userName, o.userPasswd)
	}
	// execute command
	list, err := srv.execute(cmd)
	if err != nil {
		return err
	}
	// get newly generated key (optional)
	if keyList, ok := list["PrivateKey"]; ok {
		if len(keyList) != 1 {
			return ErrOnionInvalidKeySpec
		}
		parts := strings.Split(keyList[0], ":")
		var data []byte
		if spec, ok := o.key.(string); ok {
			if len(parts) != 2 || parts[0] != spec {
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
	// get generated service identifier and compare with
	// pre-computed identifier
	srvID, err := o.ServiceID()
	if err != nil {
		return
	}
	if idList, ok := list["ServiceID"]; ok {
		if len(idList) != 1 {
			return ErrOnionAddFailed
		}
		if idList[0] != srvID {
			return fmt.Errorf("ServiceID mismatch: %s != %s", idList[0], srvID)
		}
	}
	o.running = true
	return nil
}

// Stop removes a hidden service from the Tor service
func (o *Onion) Stop(srv *Service) error {
	// check if hidden serice is running
	if !o.running {
		return ErrOnionNotRunning
	}
	// stop hidden service
	id, err := o.ServiceID()
	if err == nil {
		cmd := fmt.Sprintf("DEL_ONION %s", id)
		_, err = srv.execute(cmd)
	}
	return err
}

//----------------------------------------------------------------------
// Helper functions
//----------------------------------------------------------------------

// ServiceID returns the "onion name" of the hidden service (without the
// trailing ".onion") based on a public key
func ServiceID(key any) (id string, err error) {
	switch pub := key.(type) {
	case *ed25519.PublicKey:
		keyData := pub.Bytes()
		hsh := sha3.New256()
		hsh.Write([]byte(".onion checksum"))
		hsh.Write(keyData)
		hsh.Write([]byte{0x03})
		sum := hsh.Sum(nil)
		sum[2] = 0x03
		id = base32.StdEncoding.EncodeToString(append(keyData, sum[:3]...))
	case *rsa.PublicKey:
		var data []byte
		if data, err = asn1.Marshal(*pub); err != nil {
			return
		}
		sum := sha1.Sum(data)
		id = base32.StdEncoding.EncodeToString(sum[:len(sum)/2])
	default:
		err = ErrOnionInvalidKey
	}
	id = strings.ToLower(id)
	return
}
