package script

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/bfix/gospel/math"
)

// Statement is a single script statement.
type Statement struct {
	Opcode byte
	Data   []byte
}

// NewStatement creates a statement with an opcode.
func NewStatement(op byte) *Statement {
	return &Statement{
		Opcode: op,
		Data:   nil,
	}
}

// NewDataStatement creates a data statement.
func NewDataStatement(data []byte) *Statement {
	var op byte
	ld := len(data)
	switch {
	case ld == 0:
		return nil
	case ld < 76:
		op = byte(ld)
	case ld < 256:
		op = 76
	case ld < 65536:
		op = 77
	default:
		op = 78
	}
	return &Statement{
		Opcode: op,
		Data:   data,
	}
}

// String returns the string representation of a statement.
func (s *Statement) String() string {
	if s.Data != nil {
		d := hex.EncodeToString(s.Data)
		if s.Opcode == OpRETURN {
			return "OP_RETURN:" + d
		}
		return d
	}
	return GetOpcode(s.Opcode).Name
}

// Script is an ordered list of statements.
type Script struct {
	Stmts []*Statement
}

// Bytes returns a (flat) binary representation of the script
func (s *Script) Bytes() []byte {
	buf := new(bytes.Buffer)
	for _, s := range s.Stmts {
		_ = buf.WriteByte(s.Opcode)
		if s.Data != nil {
			ld := uint(len(s.Data))
			switch s.Opcode {
			case 76:
				_ = buf.WriteByte(byte(ld))
			case 77:
				_ = binary.Write(buf, binary.LittleEndian, uint16(ld))
			case 78:
				_ = binary.Write(buf, binary.LittleEndian, uint32(ld))
			}
			_, _ = buf.Write(s.Data)
		}
	}
	return buf.Bytes()
}

// GetTemplate returns a template derived from a script. A template only
// contains a sequence of opcodes; it is used to find structural equivalent
// scripts (but with varying data).
func (s *Script) GetTemplate() (tpl []byte, rc int) {
	for _, s := range s.Stmts {
		tpl = append(tpl, s.Opcode)
	}
	return
}

// Decompile returns a human-readable Bitcoin script source from a
// binary script representation.
func (s *Script) Decompile() string {
	var src []string
	for _, stmt := range s.Stmts {
		if stmt.Data != nil {
			if len(stmt.Data) < 5 {
				v := math.NewIntFromBytes(stmt.Data)
				src = append(src, "#"+v.String())
			} else {
				src = append(src, hex.EncodeToString(stmt.Data))
			}
		} else {
			for _, opcode := range OpCodes {
				if opcode.Value == stmt.Opcode {
					src = append(src, opcode.Name)
					break
				}
			}
		}
	}
	return strings.Join(src, " ")
}

// Add a statement at the end of the script.
func (s *Script) Add(stmt *Statement) {
	s.Stmts = append(s.Stmts, stmt)
}

// NewScript creates a new (empty) script.
func NewScript() *Script {
	return &Script{
		Stmts: make([]*Statement, 0),
	}
}

// ParseBin dissects binary scripts into a sequence of statements that
// constitutes a script.
func ParseBin(code []byte) (scr *Script, rc int) {
	var pos, size, length int
	// convert hex representation to script
	scr = NewScript()
	// get variable-length data from statement.
	getData := func(s *Statement, i int) int {
		if pos+i+1 > length {
			return RcExceeds
		}
		buf := bytes.NewReader(code[pos+1 : pos+i+1])

		var n int
		switch i {
		case 1:
			b, _ := buf.ReadByte()
			n = int(b)
		case 2:
			var v uint16
			_ = binary.Read(buf, binary.LittleEndian, &v)
			n = int(v)
		case 4:
			var v uint32
			_ = binary.Read(buf, binary.LittleEndian, &v)
			n = int(v)
		}

		size += n + i
		if pos+size > length {
			return RcExceeds
		}
		s.Data = make([]byte, n)
		copy(s.Data, code[pos+i+1:pos+i+n+1])
		return RcOK
	}
	// split the binary script code into statements
	length = len(code)
	for pos < length {
		size = 1
		op := code[pos]
		s := NewStatement(op)
		if op > 0 && op < 76 {
			n := int(op)
			if pos+n+1 > length {
				return scr, RcExceeds
			}
			s.Data = make([]byte, n)
			copy(s.Data, code[pos+1:pos+n+1])
			size += n
		} else {
			switch op {
			case OpPUSHDATA1:
				if rc = getData(s, 1); rc != RcOK {
					return
				}
			case OpPUSHDATA2:
				if rc = getData(s, 2); rc != RcOK {
					return
				}
			case OpPUSHDATA4:
				if rc = getData(s, 4); rc != RcOK {
					return
				}
			case OpRETURN:
				num := length - pos - 1
				s.Data = make([]byte, num)
				copy(s.Data, code[pos+1:pos+num+1])
				size += num
			}
		}
		pos += size
		scr.Stmts = append(scr.Stmts, s)
	}
	return scr, RcOK
}

// Parse dissects binary scripts into a sequence of statements that
// constitutes a script.
func Parse(hexScript string) (scr *Script, rc int) {
	code, err := hex.DecodeString(hexScript)
	if err != nil {
		return nil, RcParseError
	}
	return ParseBin(code)
}

// Compile compiles a source code into a Bitcoin script.
func Compile(src string) (*Script, error) {
	script := NewScript()
	for _, op := range strings.Split(src, " ") {
		if len(op) == 0 {
			continue
		}
		if strings.HasPrefix(op, "OP_") {
			found := false
			for _, opcode := range OpCodes {
				if opcode.Name == op {
					script.Add(NewStatement(opcode.Value))
					found = true
					break
				}
			}
			if !found {
				return script, fmt.Errorf("unknown opcode '%s'", op)
			}
		} else if strings.HasPrefix(op, "#") {
			v := math.NewIntFromString(op[1:])
			script.Add(NewDataStatement(v.Bytes()))
		} else {
			b, err := hex.DecodeString(op)
			if err != nil {
				return script, err
			}
			script.Add(NewDataStatement(b))
		}
	}
	return script, nil
}
