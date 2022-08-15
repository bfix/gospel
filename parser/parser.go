package parser

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
	"bufio"
	"errors"
	"io"
	"strconv"
	"unicode"

	"github.com/bfix/gospel/data"
)

const (
	// DONE for undefined/success
	DONE = iota
	// ERROR from callback
	ERROR
	// EMPTY parameter
	EMPTY
	// VAR is a generic parameter
	VAR
	// VALUE denotes a parameter value
	VALUE
	// LIST denotes a parameter list
	LIST
)

// Parameter type declaration.
type Parameter struct {
	Name  string // name of parameter
	Value string // value of parameter (encoded as string)
}

// Reset parameter instance.
func (p *Parameter) reset() {
	p.Name = ""
	p.Value = ""
}

// String returns a human-readable parameter instance.
func (p *Parameter) String() string {
	val := p.Value
	if val != "{}" && val != "~" {
		val = "\"" + val + "\""
	}
	if len(p.Name) > 0 {
		return "`" + p.Name + "=" + val + "`"
	}
	return "`" + val + "`"
}

// Callback prototype
type Callback func(mode int, param *Parameter) bool

// Parser reads data definitions from reader and pass parameters
// to callback.
//nolint:gocyclo // complex
func Parser(rdr *bufio.Reader, cb Callback) error {

	state := 1                  // current state in state machine
	skip := true                // skip white spaces (outside string)?
	escaped := false            // last character was escape?
	comment := false            // we are inside commnent
	buf := ""                   // buffer for string assembly
	param := new(Parameter)     // parameter instance
	stack := data.NewIntStack() // stack for nested values
	line, offset := 1, 0        // line/offset in current stream

	// execute state machine
	param.reset()
	for state != 0 {
		//=============================================================
		// read next rune; skip whitespaces and quit if error occurs
		//=============================================================
		r, _, err := rdr.ReadRune()
		offset++
		if err != nil {
			if err == io.EOF {
				// check for pending element
				if stack.Len() > 0 {
					if stack.Peek() == VALUE {
						// yes: named parameter
						param.Value = buf
						cb(VAR, param)
					} else if stack.Peek() == VAR {
						// yes: unnamed parameter
						param.Name = ""
						param.Value = buf
						cb(VAR, param)
					} else if stack.Len() > 1 {
						// signal parser error
						cb(ERROR, nil)
						return mkError("Pre-mature end of data", line, offset)
					}
				}
				// notify end of processing
				cb(DONE, nil)
				return nil
			}
			return err
		}
		// detect start of comment.
		if comment {
			if r == '\n' {
				comment = false
			}
			continue
		} else if !escaped && r == '#' {
			comment = true
			continue
		}
		// handle line breaks
		if r == '\n' {
			line++
			offset = 0
			continue
		}
		// skip whitechars if not within string
		if skip && unicode.IsSpace(r) {
			continue
		}
		if !escaped && r == '"' {
			skip = !skip
		}
		// handle escaped chars
		if escaped {
			escaped = false
		} else if r == '\\' {
			escaped = true
		}

		//=============================================================
		//	execute state logic
		//=============================================================
		switch state {
		//---------------------------------------------------------
		// next parameter: can be named or unnamed, value or list
		//---------------------------------------------------------
		case 1:
			{
				handled := false
				// operations on unnamed lists outside of string values
				if skip && !escaped {
					if r == '{' {
						// check for pending empty parameter
						if stack.IsTop(EMPTY) {
							stack.Pop()
						}
						// start unnamed list
						param.Name = "" // unnamed parameter
						// putback first character
						if err = rdr.UnreadRune(); err != nil {
							return err
						}
						state = 3 // read value
						handled = true
					} else if r == '}' {
						// check for pending empty parameter
						if stack.IsTop(EMPTY) {
							stack.Pop()
							param.reset()
							cb(EMPTY, param)
						}
						// end unnamed list
						// putback first character
						if err = rdr.UnreadRune(); err != nil {
							return err
						}
						state = 5 // read delimiter
						handled = true
					} else if r == ',' {
						// end unnamed empty parameter
						if !stack.IsTop(EMPTY) {
							stack.Push(EMPTY)
						}
						param.reset()
						cb(EMPTY, param)
						handled = true
					}
				}
				// check start of name
				if !handled {
					// check for pending empty parameter
					if stack.IsTop(EMPTY) {
						stack.Pop()
					}
					// named parameter; check first character
					if !unicode.IsLetter(r) && r != '"' {
						cb(ERROR, nil)
						return mkError("Invalid parameter name", line, offset)
					}
					// save initial character
					buf += string(r)
					// push VAR mode
					stack.Push(VAR)
					state = 2
				}
			}
		//---------------------------------------------------------
		// parse parameter name
		//---------------------------------------------------------
		case 2:
			{
				switch r {

				// assignment found?
				case '=':
					// start parsing value
					param.Name = buf
					state = 3

				// delimiter, opening or closing brace?
				case '{', '}', ',':
					// assign to value.
					param.Name = ""
					param.Value = buf
					// notify value complete
					cb(VAR, param)
					param.reset()
					// pop VAR tag
					stack.Pop()
					// restart
					if err = rdr.UnreadRune(); err != nil {
						return err
					}
					state = 5

				// else collect char in name buffer
				default:
					buf += string(r)
				}
			}
		//---------------------------------------------------------
		// read generic value
		//---------------------------------------------------------
		case 3:
			{
				// test for value types:
				switch r {
				// begin of new list
				case '{':
					{
						// drop VAR tag from named list.
						if stack.IsTop(VAR) {
							stack.Pop()
						}
						// push LIST mode
						stack.Push(LIST)
						param.Value = "{"
						cb(LIST, param)
						param.reset()
					}
				// empty parameter
				case ',':
					{
						// named empty parameter?
						if stack.IsTop(VAR) {
							// yes: add empty parameter
							stack.Pop()
							param.Value = ""
							cb(VAR, param)
						} else {
							// unread character
							if err = rdr.UnreadRune(); err != nil {
								return err
							}
							state = 1 // start of parameter
						}
					}
				// begin new value or parameter
				default:
					{
						// check for parameter tag on top of stack...
						if stack.IsTop(VAR) {
							// parse value for defined parameter
							stack.Push(VALUE)
							buf = string(r)
							state = 4
						} else {
							// parse new parameter
							buf = string(r)
							state = 1
						}
					}
				}
			}
		//---------------------------------------------------------
		// read parameter value
		//---------------------------------------------------------
		case 4:
			{
				// drop escapes: use escaped character directly.
				if !escaped {
					// check for termination of value
					if skip && (r == '}' || r == ',') {
						// notify value complete
						param.Value = buf
						cb(VAR, param)
						param.reset()

						// pop VALUE and VAR tags
						stack.Pop()
						stack.Pop()
						// restore read character
						if err = rdr.UnreadRune(); err != nil {
							return err
						}
						state = 5 // read/handle delimiter
					} else {
						buf += string(r)
					}
				}
			}
		//---------------------------------------------------------
		// handle delimiter between parameters
		//---------------------------------------------------------
		case 5:
			{
				switch r {
				// end of list
				case '}':
					{
						// check for correct parent type
						if !stack.IsTop(LIST) {
							cb(ERROR, nil)
							return mkError("Invalid structure '}'", line, offset)
						}
						// notify callback
						param.Value = "}"
						cb(LIST, param)
						param.reset()
						stack.Pop()
					}
				// non-delimiting character: "unread" character
				// and start new parameter expression
				default: // nolint:stylecheck // has fallthrough
					{
						if stack.Len() > 0 {
							if stack.Peek() != VAR {
								cb(ERROR, nil)
								return mkError("Invalid structure.", line, offset)
							}
							stack.Pop()
						}
						if err = rdr.UnreadRune(); err != nil {
							return err
						}
					}
					fallthrough

				// sibling delimiter
				case ',':
					{
						buf = ""  // reset buffer
						state = 1 // restart with new parameter
					}
				}
			}
		}
	}
	// report success.
	return nil
}

// generate error message.
func mkError(msg string, line int, offset int) error {
	out := msg + " (Line:" + strconv.Itoa(line)
	out += ", Offset:" + strconv.Itoa(offset) + ")"
	return errors.New(out)
}
