package parser

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"bufio"
	"github.com/bfix/gospel/data"
	"strconv"
	"strings"
)

// Data represents parsed data
type Data struct {
	Parameter
	data.Vector
	parent *Data
}

// Read data definition from reader and re-built as internal data
// structure.
func (d *Data) Read(rdr *bufio.Reader) error {

	// variable during parsing
	stack := data.NewVector() // tree of data
	curr := d                 // current data reference

	// define callback method (as closure)
	callback := func(mode int, param *Parameter) bool {

		// check for terminating parser...
		if param == nil {
			if mode == ERROR {
				// handle error condition
			} else if mode == DONE {
				// clean-up data
			}
		} else {
			// no: handle parameter
			switch mode {
			// handle list operations
			case LIST:
				{
					// start new sub-list
					if param.Value == "{" {
						// start new list object
						param.Value = "{}"
						next := curr.addToList(param)
						// remember parent object
						stack.Add(curr)
						curr = next
					} else if param.Value == "}" {
						// end sub-list: pop current object from stack
						curr = stack.Drop().(*Data)
					}
				}
			// handle named parameters
			case VAR:
				{
					// trim quoted string
					val := param.Value
					size := len(val)
					if val[0] == '"' && val[size-1] == '"' {
						param.Value = val[1 : size-1]
					}
					// add parameter to current list
					curr.addToList(param)
				}
			// handle empty parameters
			case EMPTY:
				{
					// add parameter to current list
					param.Value = "~"
					curr.addToList(param)
				}
			}
		}
		// report success
		return true
	}

	// start parser
	return Parser(rdr, callback)
}

// Write data structure to stream writer.
func (d *Data) Write(wrt *bufio.Writer) {
	d.writeData(wrt, 0)
}

// Elem accesses the n.th sub-element of nested data structure.
func (d *Data) Elem(n int) *Data {
	// check range of index
	if n < 0 || n > d.Len()-1 {
		return nil
	}
	// return indexed element
	return d.At(n).(*Data)
}

// GetPath returns the access path for current object from root.
func (d *Data) GetPath() string {

	// return root for last element
	if d.parent == nil {
		return "/"
	}
	// prefix path with parent access
	path := d.parent.GetPath()
	if len(path) > 1 {
		path += "/"
	}

	// return named element reference
	if len(d.Name) > 0 {
		return path + d.Name
	}
	// return indexed element reference
	for n := 0; n < d.parent.Len(); n++ {
		if d.parent.At(n) == d {
			return path + "#" + strconv.Itoa(n+1)
		}
	}
	// unknown access (should not happen)
	return path + "?"
}

// Lookup element in nested data structure by a path description.
// allows for automatic reference resolution (link processing)
func (d *Data) Lookup(path string) *Data {

	// leading slash means "start from real root"
	if path[0] == '/' {
		return d.getRoot().Lookup(path[1:])
	}

	// split path into current level reference
	// and follow-up reference
	list := strings.SplitN(path, "/", 2)
	curr := list[0]

	// sanity check
	if len(curr) == 0 {
		return nil
	}

	// get addressed sub-element
	var elem *Data
	if curr[0] == '#' {
		// indexed access
		if idx, err := strconv.Atoi(curr[1:]); err == nil {
			elem = d.Elem(idx - 1)
		}
		if elem == nil {
			return nil
		}
	} else {
		// named access
		found := false
		for idx := 0; idx < d.Len(); idx++ {
			elem = d.Elem(idx)
			// look for matching names
			if elem.Name == curr {
				found = true
				break
			}
		}
		// check if named element was found
		if !found {
			return nil
		}
	}
	// check for final addressing
	if len(list) > 1 {
		next := list[1]
		if len(next) > 0 {
			// recursive lookup
			return elem.Lookup(next)
		}
	}

	// check for reference resolution
	if elem.Value[0] == '@' {
		// get linked path
		link := elem.Value[1:]
		// lookup reference
		return d.Lookup(link)
	}
	// return found element directly
	return elem
}

// Write internal data structure to stream writer.
func (d *Data) writeData(wrt *bufio.Writer, level int) {

	// emit name (if defined)
	if len(d.Name) > 0 {
		wrt.WriteString(d.Name)
		wrt.WriteRune('=')
	}
	// handle value..
	if d.Len() == 0 {
		// .. as direct value
		wrt.WriteRune('"')
		wrt.WriteString(d.Value)
		wrt.WriteRune('"')
	} else {
		// .. as list of data
		if level > 0 {
			wrt.WriteRune('{')
		}
		// handle all list elements...
		count := d.Len()
		for n := 0; n < count; n++ {
			// emit delimiter
			if n > 0 {
				wrt.WriteRune(',')
			}
			// recursively write list element
			s := d.At(n).(*Data)
			s.writeData(wrt, level+1)
		}
		if level > 0 {
			wrt.WriteRune('}')
		}
	}
}

// Find root instance for current object.
func (d *Data) getRoot() *Data {
	p := d
	for p.parent != nil {
		p = p.parent
	}
	return p
}

// Add element to sub-list of this element.
func (d *Data) addToList(param *Parameter) *Data {

	// start new list object
	elem := new(Data)
	elem.Name = param.Name
	elem.Value = param.Value
	// link back to parent
	elem.parent = d

	// add to list of sub-elements
	d.Add(elem)
	return elem
}
