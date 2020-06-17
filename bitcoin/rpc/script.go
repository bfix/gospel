package rpc

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
	"github.com/bfix/gospel/bitcoin/script"
)

// AddWitnessAddress adds a witness address for a script (with pubkey or
// redeemscript known).
func (s *Session) AddWitnessAddress(addr string) (string, error) {
	res, err := s.call("addwitnessaddress", []Data{addr})
	if err != nil {
		return "", err
	}
	return res.Result.(string), nil
}

// DecodeScript decodes a hex-encoded P2SH redeem script.
func (s *Session) DecodeScript(script string) (*DecodedScript, error) {
	res, err := s.call("decodescript", []Data{script})
	if err != nil {
		return nil, err
	}
	ds := new(DecodedScript)
	if ok, err := res.UnmarshalResult(ds); !ok {
		return nil, err
	}
	return ds, nil
}

// DataScript assembles a OpRETURN script with data attached.
func DataScript(data []byte) (scr []byte) {
	scr = append(scr, script.OpRETURN)
	scr = append(scr, PushData(data)...)
	return
}

// PushData creates an opcode instruction to push binary data onto the stack.
func PushData(data []byte) (res []byte) {
	size := len(data)
	switch {
	case size < 76:
		res = append(res, byte(size))
	case size < 256:
		res = append(res, script.OpPUSHDATA1)
		res = append(res, byte(size))
	case size < 65536:
		// size of script
		res = append(res, script.OpPUSHDATA2)
		res = append(res, byte(size&0xFF))
		res = append(res, byte((size>>8)&0xFF))
	default:
		res = append(res, script.OpPUSHDATA4)
		res = append(res, byte(size&0xFF))
		res = append(res, byte((size>>8)&0xFF))
		res = append(res, byte((size>>16)&0xFF))
		res = append(res, byte((size>>24)&0xFF))
	}
	res = append(res, data...)
	return
}
