//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2024 Bernd Fix  >Y<
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

package data

import (
	"bytes"
	"strconv"
	"testing"
)

func TestSExpr(t *testing.T) {
	s := `(rsa
		     (n #deadbeaf00bf19036267#)
		     (e #010001#)
		     (protected openpgp-s2k3-ocb-aes
			    ((sha1 #012345678# 10237235) #cafe07feed#)
			    #FF7CCCEDE05769DB8EAB8C7608C603B51A9D1BB9#
		     )
		     (protected-at "20240324T192342")
	      )`

	root, err := ParseSExpr(s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(root.String())
	root.Drop()
}

func TestCSexpr(t *testing.T) {

	s := new(bytes.Buffer)
	s.WriteString("(3:rsa(1:n")
	m := []byte{0x00, 0xbf, 0x19, 0x03, 0x62, 0x23, 0x42, 0x67}
	s.WriteString(strconv.Itoa(len(m)))
	s.WriteRune(':')
	s.Write(m)
	s.WriteString(")(1:e")
	e := []byte{0x01, 0x00, 0x01}
	s.WriteString(strconv.Itoa(len(e)))
	s.WriteRune(':')
	s.Write(e)
	s.WriteString(")(12:protected-at")
	d := "20240324T192342"
	s.WriteString(strconv.Itoa(len(d)))
	s.WriteRune(':')
	s.WriteString(d)
	s.WriteString("))")
	buf := s.Bytes()

	root, err := ParseCSExpr(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(root.String())
	root.Drop()
}
