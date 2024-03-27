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
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// For details on S-expressions, see https://en.wikipedia.org/wiki/S-expression
// This code can handle []byte elements which are the default type for
// canonical S-expressions.

// SExpr is a node in a S-expression tree
type SExpr struct {
	atoms  []any
	parent *SExpr
}

// SExprAt returns the (typed) n.th element of the node
func SExprAt[T any](s *SExpr, n int) (v T, ok bool) {
	if n < 0 || n >= len(s.atoms) {
		ok = false
		return
	}
	v, ok = s.atoms[n].(T)
	return
}

// Find node by tag (first element)
func (n *SExpr) Find(key string) *SExpr {
	// check if the string at position 0 matches
	if v, ok := n.atoms[0].(string); ok && v == key {
		return n
	}
	// try to find in children
	for _, e := range n.atoms {
		if c, ok := e.(*SExpr); ok {
			if x := c.Find(key); x != nil {
				return x
			}
		}
	}
	return nil
}

// Drop node including children
func (n *SExpr) Drop() {
	n.parent = nil
	for _, e := range n.atoms {
		if c, ok := e.(*SExpr); ok {
			c.Drop()
		}
	}
	n.atoms = nil
}

// String returns a human-readable S-expression tree
func (n *SExpr) String() (s string) {
	s = "("
	// dump elements
	last := len(n.atoms) - 1
	for i, e := range n.atoms {
		switch x := e.(type) {
		case string:
			if strings.Contains(x, " ") {
				s += "\"" + x + "\""
			} else {
				s += x
			}
		case int64:
			s += strconv.FormatInt(x, 10)
		case float64:
			s += strconv.FormatFloat(x, 'f', -1, 64)
		case []byte:
			s += "#" + hex.EncodeToString(x) + "#"
		case *SExpr:
			s += x.String()
		}
		if i != last {
			s += " "
		}
	}
	s += ")"
	return
}

// NewSExpr creates an empty node
func NewSExpr() *SExpr {
	return &SExpr{
		atoms: make([]any, 0),
	}
}

//----------------------------------------------------------------------

// ParseCSecpr parses a canonical S-expression into a tree of nodes.
func ParseCSExpr(buf []byte) (root *SExpr, err error) {

	// check for correct S-expr start
	if buf[0] != '(' {
		err = fmt.Errorf("s-expr does not start with '(' but '%c'", buf[0])
		return
	}
	buf = buf[1:]

	// build S-expr tree
	root = NewSExpr()
	curr := root
	var token any
	for len(buf) > 0 {
		token, buf = tokenizeCanonical(buf)
		switch x := token.(type) {
		case error:
			err = x
			return
		case uint8:
			if x == '(' {
				child := NewSExpr()
				child.parent = curr
				curr.atoms = append(curr.atoms, child)
				curr = child
			} else if x == ')' {
				curr = curr.parent
				if curr == nil {
					return
				}
			}
		default:
			curr.atoms = append(curr.atoms, x)
		}
	}
	return
}

// get next token from a canonical s-expression.
func tokenizeCanonical(s []byte) (token any, rem []byte) {
	// start or end of node
	if s[0] == '(' || s[0] == ')' {
		return s[0], s[1:]
	}
	// get length field
	idx := strings.Index(string(s), ":")
	if idx == -1 {
		token = fmt.Errorf("missing length field at '%s...'", s[:6])
		return
	}
	n, err := strconv.Atoi(string(s[:idx]))
	if err != nil {
		token = err
		return
	}
	// classify parsed token
	buf := make([]byte, n)
	copy(buf, []byte(s)[idx+1:idx+1+n])
	token = classifyToken(buf)

	rem = s[idx+n+1:]
	return
}

//----------------------------------------------------------------------

// ParseSExpr parses a text-only S-expression into a tree of nodes.
func ParseSExpr(line string) (root *SExpr, err error) {

	// check for correct S-expr start
	if line[0] != '(' {
		err = fmt.Errorf("s-expr does not start with '(' but '%c'", line[0])
		return
	}
	line = line[1:]

	// build S-expr tree
	root = NewSExpr()
	curr := root
	var token any
	for len(line) > 0 {
		token, line = tokenize(line)
		switch x := token.(type) {
		case error:
			err = x
			return
		case uint8:
			if x == '(' {
				child := NewSExpr()
				child.parent = curr
				curr.atoms = append(curr.atoms, child)
				curr = child
			} else if x == ')' {
				curr = curr.parent
				if curr == nil {
					return
				}
			}
		default:
			curr.atoms = append(curr.atoms, x)
		}
	}
	return
}

// get next token from a textual S-expression.
func tokenize(s string) (token any, rem string) {

	// skip whitespaces
	for unicode.IsSpace(rune(s[0])) {
		s = s[1:]
	}
	// check for start or end of node
	if s[0] == '(' || s[0] == ')' {
		return s[0], s[1:]
	}
	// read token until delimiter
	var end int
	tok := ""
	for end = 0; ; end++ {
		c := rune(s[end])
		if unicode.IsSpace(c) || strings.ContainsRune("()", c) {
			break
		}
		tok += string(s[end])
	}
	switch tok[0] {
	case '#':
		// unwrap hex byte array
		tok = strings.Trim(tok, "#")
		token, _ = hex.DecodeString(tok)
	case '"':
		// unwrap string literal
		token = strings.Trim(tok, "\"")
	default:
		token = classifyToken(tok)
	}
	rem = s[end:]
	return
}

//----------------------------------------------------------------------

// classify
func classifyToken(tok any) any {

	// stringify token
	var s string
	switch x := tok.(type) {
	case []byte:
		s = string(x)
	case string:
		s = x
	}
	// try to classify:
	// (1) check for floating point number
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		if strings.Contains(s, ".") {
			return v
		}
	}
	// (2) check for integer number
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		i1 := strconv.FormatInt(v, 10)
		if len(i1) == len(s) {
			return v
		}
	}
	// (3) check for ASCII string
	isString := true
	for i := 0; i < len(s); i++ {
		if s[i] < 32 || s[i] > 127 {
			isString = false
			break
		}
	}
	if isString {
		return s
	}
	// keep token as is
	return tok
}
