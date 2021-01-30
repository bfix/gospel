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
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bfix/gospel/crypto/ed25519"
	"github.com/bfix/gospel/math"
	"golang.org/x/crypto/sha3"
)

//----------------------------------------------------------------------
// Tor onion handling (hidden services)
//----------------------------------------------------------------------

var (
	ErrOnionMissingKey     = fmt.Errorf("Missing private key for onion")
	ErrOnionInvalidKeySpec = fmt.Errorf("Invalid provate key specification")
	ErrOnionAddFailed      = fmt.Errorf("Failed to add hidden service")
)

// Onion is a Ed25519-based hidden Tor service.
type Onion struct {
	Prv   *ed25519.PrivateKey // private key
	Flags []string            // hidden service flags
	Ports map[int]string      // port mappings
	srvId string              // generated service id (transient)
}

// ServiceID generates a service identifier for a hidden service
func (o *Onion) ServiceID() string {
	// handle case of unknown key
	if o.Prv == nil {
		return ""
	}
	// check for existing service identifier
	if len(o.srvId) == 0 {
		// generate service identifier
		keyData := o.Prv.Public().Bytes()
		hsh := sha3.New256()
		hsh.Write([]byte(".onion checksum"))
		hsh.Write(keyData)
		hsh.Write([]byte{0x03})
		sum := hsh.Sum(nil)
		sum[2] = 0x03
		o.srvId = base32.StdEncoding.EncodeToString(append(keyData, sum[:3]...))
	}
	return o.srvId
}

// AddOnion instantiates a new hidden service. The service will listen at the
// specified port and will redirect traffic to host:port.
func (c *Control) AddOnion(o *Onion) (string, error) {
	cmd := "ADD_ONION "
	// use private Ed25519 key if available
	if o.Prv != nil {
		kb := make([]byte, 64)
		if _, err := rand.Read(kb); err != nil {
			return "", err
		}
		copy(kb[:32], o.Prv.D.Bytes())
		cmd += fmt.Sprintf("ED25519-V3:%s", base64.StdEncoding.EncodeToString(kb))
	} else {
		cmd += "NEW:ED25519-V3"
	}
	// add flags (optional)
	if len(o.Flags) > 0 {
		cmd += " Flags="
		for i, flag := range o.Flags {
			if i > 0 {
				cmd += ","
			}
			cmd += flag
		}
	}
	// add port mappings
	for listen, tgt := range o.Ports {
		cmd += fmt.Sprintf(" Port=%d,%s", listen, tgt)
	}
	// execute command
	list, err := c.execute(cmd)
	if err != nil {
		return "", err
	}
	// get newly generated key
	if o.Prv == nil {
		kd, ok := list["PrivateKey"]
		if !ok {
			return "", ErrOnionMissingKey
		}
		parts := strings.Split(kd, ":")
		if len(parts) != 2 || parts[0] != "ED25519-V3" {
			return "", ErrOnionInvalidKeySpec
		}
		spec, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", err
		}
		d := math.NewIntFromBytes(spec[:32])
		o.Prv = ed25519.NewPrivateKeyFromD(d)
	}
	// get generated service identifier
	if id, ok := list["ServiceID"]; ok {
		o.srvId = id
		return id, nil
	}
	return "", ErrOnionAddFailed
}

// DelOnion removes a hidden service with given ID
func (c *Control) DelOnion(o *Onion) error {
	cmd := fmt.Sprintf("DEL_ONION %s", o.ServiceID())
	_, err := c.execute(cmd)
	return err
}
