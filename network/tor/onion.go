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
	ErrOnionKeyExists      = fmt.Errorf("Onion private key already exists")
	ErrOnionMissingKey     = fmt.Errorf("Missing private key for onion")
	ErrOnionInvalidKeySpec = fmt.Errorf("Invalid provate key specification")
	ErrOnionAddFailed      = fmt.Errorf("Failed to add hidden service")
)

// Onion interface for Tor hidden services
type Onion interface {
	SetKey(string) error      // set private key from string representation
	KeySpec() (string, error) // get key specification (for ADD_ONION call)
	AddFlag(string)           // add a flag for the hidden service
	Flags() []string          // hidden service flags
	AddPort(int, string)      // add a port mapping
	Ports() map[int]string    // hidden service port mappings
	ServiceID() string        // get associated service identifier
}

//----------------------------------------------------------------------
// Base onion
//----------------------------------------------------------------------

// onionBase is a base type for onion implementations
type onionBase struct {
	flags []string       // hidden service flags
	ports map[int]string // port mappings
	srvId string         // service identifier
}

// AddFlag adds a flag to the list
func (o *onionBase) AddFlag(flag string) {
	o.flags = append(o.flags, flag)
}

// Flags returns a list of hidden service flags
func (o *onionBase) Flags() []string {
	return o.flags
}

// AddPort adds a port mapping for the hidden service
func (o *onionBase) AddPort(listen int, spec string) {
	o.ports[listen] = spec
}

// Ports returns port mappings for hidden service
func (o *onionBase) Ports() map[int]string {
	return o.ports
}

//----------------------------------------------------------------------
// Ed25519-based onion
//----------------------------------------------------------------------

// OnionEd25519 is a Ed25519-based hidden Tor service.
type OnionEd25519 struct {
	onionBase
	prv *ed25519.PrivateKey // private key
}

// NewOnionEd25519 instantiates a new Ed25519-based hidden service
func NewOnionEd25519(prv *ed25519.PrivateKey) *OnionEd25519 {
	id := ""
	if prv != nil {
		id = ServiceIdEd25519(prv.Public())
	}
	return &OnionEd25519{
		onionBase: onionBase{
			flags: make([]string, 0),
			ports: make(map[int]string),
			srvId: id,
		},
		prv: prv,
	}
}

// ServiceIdEd25519 computes the hidden service identifier from an Ed25519
// public key of a hidden service.
func ServiceIdEd25519(pub *ed25519.PublicKey) string {
	keyData := pub.Bytes()
	hsh := sha3.New256()
	hsh.Write([]byte(".onion checksum"))
	hsh.Write(keyData)
	hsh.Write([]byte{0x03})
	sum := hsh.Sum(nil)
	sum[2] = 0x03
	id := base32.StdEncoding.EncodeToString(append(keyData, sum[:3]...))
	return strings.ToLower(id)
}

// SetKey sets the private key for a hidden service (only if no key is defined
// yet).
func (o *OnionEd25519) SetKey(spec string) error {
	if o.prv != nil {
		return ErrOnionKeyExists
	}
	parts := strings.Split(spec, ":")
	if len(parts) != 2 || parts[0] != "ED25519-V3" {
		return ErrOnionInvalidKeySpec
	}
	kd, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	d := math.NewIntFromBytes(kd[:32])
	o.prv = ed25519.NewPrivateKeyFromD(d)
	return nil
}

// KeySpec returns the key specification used to add hidden services
func (o *OnionEd25519) KeySpec() (string, error) {
	// check for existing key
	if o.prv != nil {
		kb := make([]byte, 64)
		if _, err := rand.Read(kb); err != nil {
			return "", err
		}
		copy(kb[:32], o.prv.D.Bytes())
		return fmt.Sprintf("ED25519-V3:%s", base64.StdEncoding.EncodeToString(kb)), nil
	}
	// create a new key
	return "NEW:ED25519-V3", nil
}

// ServiceID generates a service identifier for a hidden service
func (o *OnionEd25519) ServiceID() string {
	// handle case of unknown key
	if o.prv == nil {
		return ""
	}
	// check for existing service identifier
	if len(o.srvId) == 0 {
		// generate service identifier
		o.srvId = ServiceIdEd25519(o.prv.Public())
	}
	return o.srvId
}

//----------------------------------------------------------------------
// RSA1024-based onion
//----------------------------------------------------------------------

// OnionRSA1024 is a RSA1024-based hidden Tor service.
type OnionRSA1024 struct {
	onionBase
	prv *rsa.PrivateKey // private key
}

// NewOnionRSA1024 instantiates a new RSA1024-based hidden service
func NewOnionRSA1024(prv *rsa.PrivateKey) (*OnionRSA1024, error) {
	id := ""
	if prv != nil {
		var err error
		if id, err = ServiceIdRSA1024(prv.Public().(*rsa.PublicKey)); err != nil {
			return nil, err
		}
	}
	return &OnionRSA1024{
		onionBase: onionBase{
			flags: make([]string, 0),
			ports: make(map[int]string),
			srvId: id,
		},
		prv: prv,
	}, nil
}

// ServiceIdRSA1024 computes the hidden service identifier from a RSA-1024
// public key of a hidden service.
func ServiceIdRSA1024(pub *rsa.PublicKey) (string, error) {
	der, err := asn1.Marshal(*pub)
	if err != nil {
		return "", err
	}
	// Onion id is base32(firstHalf(sha1(publicKeyDER)))
	hash := sha1.Sum(der)
	half := hash[:len(hash)/2]
	id := base32.StdEncoding.EncodeToString(half)
	return strings.ToLower(id), nil
}

// SetKey sets the private key for a hidden service (only if no key is defined
// yet).
func (o *OnionRSA1024) SetKey(spec string) error {
	if o.prv != nil {
		return ErrOnionKeyExists
	}
	parts := strings.Split(spec, ":")
	if len(parts) != 2 || parts[0] != "RSA1024" {
		return ErrOnionInvalidKeySpec
	}
	kd, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	o.prv, err = x509.ParsePKCS1PrivateKey(kd)
	return err
}

// KeySpec returns the key specification used to add hidden services
func (o *OnionRSA1024) KeySpec() (string, error) {
	// check for existing key
	if o.prv != nil {
		kd := x509.MarshalPKCS1PrivateKey(o.prv)
		return fmt.Sprintf("RSA1024:%s", base64.StdEncoding.EncodeToString(kd)), nil
	}
	// create a new key
	return "NEW:RSA1024", nil
}

// ServiceID generates a service identifier for a hidden service
func (o *OnionRSA1024) ServiceID() string {
	// handle case of unknown key
	if o.prv == nil {
		return ""
	}
	// check for existing service identifier
	if len(o.srvId) == 0 {
		// generate service identifier
		var err error
		pub, ok := o.prv.Public().(*rsa.PublicKey)
		if !ok {
			return ""
		}
		if o.srvId, err = ServiceIdRSA1024(pub); err != nil {
			return ""
		}
	}
	return o.srvId
}

//----------------------------------------------------------------------
// Ading and deleting hidden service for a running Tor service
//----------------------------------------------------------------------

// AddOnion instantiates a new hidden service. The service will listen at the
// specified port and will redirect traffic to host:port.
func (c *Control) AddOnion(o Onion) (string, error) {
	spec, err := o.KeySpec()
	if err != nil {
		return "", err
	}
	cmd := "ADD_ONION " + spec
	// add flags (optional)
	flags := o.Flags()
	if len(flags) > 0 {
		cmd += " Flags="
		for i, flag := range flags {
			if i > 0 {
				cmd += ","
			}
			cmd += flag
		}
	}
	// add port mappings
	for listen, tgt := range o.Ports() {
		cmd += fmt.Sprintf(" Port=%d,%s", listen, tgt)
	}
	// execute command
	list, err := c.execute(cmd)
	if err != nil {
		return "", err
	}
	// get newly generated key (optional)
	kd, ok := list["PrivateKey"]
	if ok {
		if err = o.SetKey(kd); err != nil {
			return "", err
		}
	}
	// get generated service identifier
	srvId := o.ServiceID()
	if id, ok := list["ServiceID"]; ok {
		if id != srvId {
			return "", ErrOnionAddFailed
		}
	}
	return srvId, nil
}

// DelOnion removes a hidden service with given ID
func (c *Control) DelOnion(o Onion) error {
	cmd := fmt.Sprintf("DEL_ONION %s", o.ServiceID())
	_, err := c.execute(cmd)
	return err
}
