package data

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2022 Bernd Fix
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
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

	gerr "github.com/bfix/gospel/errors"
)

//######################################################################
//
// Serialization of Golang objects of type 'struct{}':
// Field types can be any of these:
//
//    int{8,16,32,64}       -- Signed integer of given size
//    uint{8,16,32,64}      -- Unsigned integer of given size (little-endian)
//    []uint8,[]byte        -- variable length byte array
//    string                -- variable length string (\0 -terminated)
//    bool                  -- boolean value (single byte, 0=false, true otherwise)
//    *struct{}, struct{}   -- nested structure
//    []*struct{}, []struct -- list of structures with allowed fields
//
// The serialization can be controlled by field annotations:
//
// ---------------------------------------
// (1) Endianness of integers: tag "order"
// ---------------------------------------
// Integer fields (of size > 1) can be tagged for Big-Endian representation
// by using the tag "order" with a value of "big":
//
//    field1 int64 `order:"big"`
//
// ---------------------------------
// (2) Array/slice sizes: tag "size"
// ---------------------------------
// Variable-length slices can be tagged with a "size" tag to help the
// Unmarshal function to figure out the number of slice elements to
// process. The values can be "*" for greedy (as many elements as
// possible before running out of data), "<num>" a decimal number specifying
// the fixed size or two dynamic/variable approaches:
//
// (a) A "<name>" referring to a previous unsigned integer field in the
//     struct object:
//
//     ListSize uint16
//     List     []*Entry `size:"ListSize"`
//
// (b) A "(<name>)" referring to a struct object method, that takes no
//     or one argument and returns an unsigned integer for length:
//
//     List     []*Entry `size:"(CalcSize)"`
//
//     The "CalcSize" method works on an incompletely initialized object
//     instance and can only consider data that has already been read.
//     Its possible argument is a string (name of the annotated field).
//
// N.B.: You can't do math in the size expression (except for the greedy
//       expression like "*-16"); you need to out-source the calculation
//       to a method if required.
//
// ------------------------------
// (3) Optional fields: tag "opt"
// ------------------------------
// A field can be marked as optional with the "opt" tag; the tag value
// must be either a (previous) boolean field or a method returning a
// bool value. As with size functions, the method can take a single
// string argument (field name).
//
//######################################################################

// Errors
var (
	ErrMarshalNil          = errors.New("object is nil")
	ErrMarshalType         = errors.New("invalid object type")
	ErrMarshalNoSize       = errors.New("missing size tag on field")
	ErrMarshalSizeMismatch = errors.New("size mismatch during unmarshal")
	ErrMarshalEmptyIntf    = errors.New("can't handle empty interface")
	ErrMarshalUnknownType  = errors.New("unknown field type")
	ErrMarshalMthdMissing  = errors.New("missing method")
	ErrMarshalFieldRef     = errors.New("field reference invalid")
	ErrMarshalMthdNumArg   = errors.New("method has more than one argument")
	ErrMarshalMthdArgType  = errors.New("method argument not a string")
	ErrMarshalMthdResult   = errors.New("invalid method result")
)

//======================================================================
// Marshal Golang objects to byte arrays.
//======================================================================

// Marshal creates a byte array from a (reference to an) object.
func Marshal(obj interface{}) ([]byte, error) {
	wrt := new(bytes.Buffer)
	if err := MarshalStream(wrt, obj); err != nil {
		return nil, err
	}
	return wrt.Bytes(), nil
}

// MarshalStream writes an object instance to stream
func MarshalStream(wrt io.Writer, obj interface{}) error {
	var inst reflect.Value
	path := newPath()
	var marshal func(x reflect.Value) error
	marshal = func(x reflect.Value) error {
		for i := 0; i < x.NumField(); i++ {
			f := x.Field(i)
			// do not serialize unexported fields
			if !f.CanSet() {
				continue
			}
			ft := x.Type().Field(i)
			path.push(ft.Name)

			// collect annotations
			tagSize := ft.Tag.Get("size")
			tagOrder := ft.Tag.Get("order")
			tagOpt := ft.Tag.Get("opt")

			// check for optional field
			used, err := isUsed(tagOpt, ft.Name, x, inst)
			if err != nil {
				return gerr.New(err, "field '%s'", path.string())
			}
			if !used {
				path.pop()
				continue
			}

			switch v := f.Interface().(type) {
			//----------------------------------------------------------
			// Strings
			//----------------------------------------------------------
			case string:
				if _, err := wrt.Write([]byte(v)); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				if _, err := wrt.Write([]byte{0}); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
			//----------------------------------------------------------
			// Booleans
			//----------------------------------------------------------
			case bool:
				var a byte
				if v {
					a = 1
				}
				if _, err := wrt.Write([]byte{a}); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
			//----------------------------------------------------------
			// Integers
			//----------------------------------------------------------
			case uint8, int8, uint16, int16, uint32, int32, uint64, int64, int:
				if err := writeInt(wrt, tagOrder, v); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
			//----------------------------------------------------------
			// Byte arrays
			//----------------------------------------------------------
			case []uint8:
				if _, err := parseSize(tagSize, ft.Name, x, inst, len(v), -1); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				if _, err := wrt.Write(v); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}

			//----------------------------------------------------------
			// Handle other complex types...
			//----------------------------------------------------------
			default:
				switch f.Kind() {
				//------------------------------------------------------
				// Interfaces
				//------------------------------------------------------
				case reflect.Interface:
					e := f.Elem()
					if e.Kind() == reflect.Ptr {
						e = e.Elem()
					}
					if e.IsValid() {
						if err := marshal(e); err != nil {
							return err
						}
					}
				//------------------------------------------------------
				// Pointers
				//------------------------------------------------------
				case reflect.Ptr:
					e := f.Elem()
					if e.IsValid() {
						if err := marshal(e); err != nil {
							return err
						}
					}
				//------------------------------------------------------
				// Structs
				//------------------------------------------------------
				case reflect.Struct:
					if err := marshal(f); err != nil {
						return err
					}
				//------------------------------------------------------
				// Slices
				//------------------------------------------------------
				case reflect.Slice:
					count, err := parseSize(tagSize, ft.Name, x, inst, f.Len(), -1)
					if err != nil {
						return gerr.New(err, "field '%s'", path.string())
					}
					// greedy slice: use existing size
					if count < 0 {
						count = f.Len()
					}
					for i := 0; i < count; i++ {
						e := f.Index(i)
						switch e.Kind() {
						//----------------------------------------------
						// Interface elements
						//----------------------------------------------
						case reflect.Interface:
							e = e.Elem()
							if e.Kind() == reflect.Ptr {
								e = e.Elem()
							}
							if err := marshal(e); err != nil {
								return err
							}
						//----------------------------------------------
						// Pointer elements
						//----------------------------------------------
						case reflect.Ptr:
							if err := marshal(e.Elem()); err != nil {
								return err
							}
						//----------------------------------------------
						// Struct elements
						//----------------------------------------------
						case reflect.Struct:
							if err := marshal(e); err != nil {
								return err
							}
						//----------------------------------------------
						// Intrinsics (strings, integers)
						//----------------------------------------------
						default:
							switch v := e.Interface().(type) {
							case string:
								if _, err := wrt.Write([]byte(v)); err != nil {
									return gerr.New(err, "field '%s'", path.string())
								}
								if _, err := wrt.Write([]byte{0}); err != nil {
									return gerr.New(err, "field '%s'", path.string())
								}
							case bool:
								var a byte
								if v {
									a = 1
								}
								if _, err := wrt.Write([]byte{a}); err != nil {
									return gerr.New(err, "field '%s'", path.string())
								}
							case uint8, int8, uint16, int16, uint32, int32, uint64, int64, int:
								if err := writeInt(wrt, tagOrder, v); err != nil {
									return gerr.New(err, "field '%s'", path.string())
								}
							}
						}
					}
				default:
					return gerr.New(ErrMarshalUnknownType, "field '%s'", path.string())
				}
			}
			path.pop()
		}
		return nil
	}
	// process if object is a '*struct{}', a 'struct{}' or an interface
	inst = reflect.ValueOf(obj)
	switch inst.Kind() {
	case reflect.Interface:
		e := inst.Elem()
		if e.Kind() == reflect.Ptr {
			e = e.Elem()
		}
		return marshal(e)
	case reflect.Ptr:
		e := inst.Elem()
		if e.IsValid() {
			return marshal(e)
		}
		return ErrMarshalNil
	case reflect.Struct:
		return marshal(inst)
	}
	return ErrMarshalType
}

//======================================================================
// Unmarshal Golang objects from byte arrays.
//======================================================================

// Unmarshal reads a byte array to fill an object pointed to by 'obj'.
func Unmarshal(obj interface{}, data []byte) error {
	buf := bytes.NewBuffer(data)
	return UnmarshalStream(buf, obj, len(data))
}

// UnmarshalStream reads an object from strean.
func UnmarshalStream(rdr io.Reader, obj interface{}, pending int) error {
	var inst reflect.Value
	path := newPath()
	var unmarshal func(x reflect.Value) error
	unmarshal = func(x reflect.Value) error {
		for i := 0; i < x.NumField(); i++ {
			f := x.Field(i)
			// skip unexported fields
			if !f.CanSet() {
				continue
			}
			ft := x.Type().Field(i)
			path.push(ft.Name)

			// collect annotations
			tagSize := ft.Tag.Get("size")
			tagOrder := ft.Tag.Get("order")
			tagOpt := ft.Tag.Get("opt")

			// check for optional field
			used, err := isUsed(tagOpt, ft.Name, x, inst)
			if err != nil {
				return gerr.New(err, "field '%s'", path.string())
			}
			if !used {
				path.pop()
				continue
			}

			switch f.Interface().(type) {
			//----------------------------------------------------------
			// Strings
			//----------------------------------------------------------
			case string:
				s := ""
				b := make([]byte, 1)
				for {
					if _, err := rdr.Read(b); err != nil {
						return gerr.New(err, "field '%s'", path.string())
					}
					if b[0] == 0 {
						break
					}
					s += string(b)
				}
				f.SetString(s)
				pending -= len(s) + 1
			//----------------------------------------------------------
			// Booleans
			//----------------------------------------------------------
			case bool:
				b := make([]byte, 1)
				if _, err := rdr.Read(b); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				var a bool
				if b[0] != 0 {
					a = true
				}
				f.SetBool(a)
			//----------------------------------------------------------
			// Integers
			//----------------------------------------------------------
			case uint8:
				var a uint8
				if err := binary.Read(rdr, binary.LittleEndian, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetUint(uint64(a))
				pending--
			case int8:
				var a int8
				if err := binary.Read(rdr, binary.LittleEndian, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetInt(int64(a))
				pending--
			case uint16:
				var a uint16
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetUint(uint64(a))
				pending -= 2
			case int16:
				var a int16
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetInt(int64(a))
				pending -= 2
			case uint32:
				var a uint32
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetUint(uint64(a))
				pending -= 4
			case int32, int:
				var a int32
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetInt(int64(a))
				pending -= 4
			case uint64:
				var a uint64
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetUint(a)
				pending -= 8
			case int64:
				var a int64
				if err := readInt(rdr, tagOrder, &a); err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				f.SetInt(a)
				pending -= 8

			//----------------------------------------------------------
			// Byte arrays
			//----------------------------------------------------------
			case []uint8:
				size, err := parseSize(tagSize, ft.Name, x, inst, f.Len(), pending)
				if err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				a := make([]byte, size)
				n, err := rdr.Read(a)
				if err != nil {
					return gerr.New(err, "field '%s'", path.string())
				}
				if n != size {
					return gerr.New(ErrMarshalSizeMismatch, "field '%s'", path.string())
				}
				f.SetBytes(a)
				pending -= n

			//----------------------------------------------------------
			// Read more complex types.
			//----------------------------------------------------------
			default:
				switch f.Kind() {
				//------------------------------------------------------
				// Interfaces
				//------------------------------------------------------
				case reflect.Interface:
					e := f.Elem()
					if !e.IsValid() {
						return gerr.New(ErrMarshalEmptyIntf, "field '%s'", path.string())
					}
					if err := unmarshal(e); err != nil {
						return err
					}
				//------------------------------------------------------
				// Pointers
				//------------------------------------------------------
				case reflect.Ptr:
					e := f.Elem()
					if !e.IsValid() {
						ep := reflect.New(f.Type().Elem())
						e = ep.Elem()
						f.Set(ep)
					}
					if err := unmarshal(e); err != nil {
						return err
					}
				//------------------------------------------------------
				// Structs
				//------------------------------------------------------
				case reflect.Struct:
					if err := unmarshal(f); err != nil {
						return err
					}
				//------------------------------------------------------
				// Slices
				//------------------------------------------------------
				case reflect.Slice:
					// get size of slice: if the size is zero (empty or nil
					// array), use the "size" tag to determine the desired
					// length. The tag value can be "*" for greedy (read
					// until end of buffer), the name of a (previous) integer
					// field containing the length or an integer value.
					count, err := parseSize(tagSize, ft.Name, x, inst, f.Len(), -1)
					if err != nil {
						return gerr.New(err, "field '%s'", path.string())
					}
					add := (count > f.Len())
					// If the element type is a pointer, get the type of the
					// referenced object and remember to use a pointer.
					et := f.Type().Elem()
					isPtr := false
					if et.Kind() == reflect.Ptr {
						isPtr = true
						et = et.Elem()
					}
					// unmarshal slice elements
					for i := 0; i < count || count < 0; i++ {
						// quit on end-of-buffer
						if pending < 1 {
							break
						}
						// address the slice element. If the element does not
						// exist, create a new one and append it to the slice.
						var e reflect.Value
						if add {
							// create and add new element
							ep := reflect.New(et)
							e = ep.Elem()
							if isPtr {
								f.Set(reflect.Append(f, ep))
							} else {
								f.Set(reflect.Append(f, e))
							}
						}
						// use existing element
						e = f.Index(i)

						switch e.Kind() {
						//----------------------------------------------
						// Interface elements
						//----------------------------------------------
						case reflect.Interface:
							e = e.Elem()
							if !e.IsValid() {
								return gerr.New(ErrMarshalEmptyIntf, "field '%s'", path.string())
							}
							if err := unmarshal(e); err != nil {
								return err
							}
						//----------------------------------------------
						// Pointer elements
						//----------------------------------------------
						case reflect.Ptr:
							if err := unmarshal(e.Elem()); err != nil {
								return err
							}
						//----------------------------------------------
						// Struct elements
						//----------------------------------------------
						case reflect.Struct:
							if err := unmarshal(e); err != nil {
								return err
							}
						//----------------------------------------------------------
						// Strings
						//----------------------------------------------------------
						case reflect.String:
							s := ""
							b := make([]byte, 1)
							for {
								if _, err := rdr.Read(b); err != nil {
									return gerr.New(err, "field '%s'", path.string())
								}
								if b[0] == 0 {
									break
								}
								s += string(b)
							}
							e.SetString(s)
							pending -= len(s) + 1
						//----------------------------------------------------------
						// Integers
						//----------------------------------------------------------
						case reflect.Int8:
							var a int8
							if err := binary.Read(rdr, binary.LittleEndian, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetInt(int64(a))
							pending--
						case reflect.Uint16:
							var a uint16
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetUint(uint64(a))
							pending -= 2
						case reflect.Int16:
							var a int16
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetInt(int64(a))
							pending -= 2
						case reflect.Uint32:
							var a uint32
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetUint(uint64(a))
							pending -= 4
						case reflect.Int32, reflect.Int:
							var a int32
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetInt(int64(a))
							pending -= 4
						case reflect.Uint64:
							var a uint64
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetUint(a)
							pending -= 8
						case reflect.Int64:
							var a int64
							if err := readInt(rdr, tagOrder, &a); err != nil {
								return gerr.New(err, "field '%s'", path.string())
							}
							e.SetInt(a)
							pending -= 8
						}
					}
				default:
					return gerr.New(ErrMarshalUnknownType, "field '%s'", path.string())
				}
			}
			path.pop()
		}
		return nil
	}
	// check if object is a '*struct{}'
	inst = reflect.ValueOf(obj)
	if inst.Kind() == reflect.Ptr {
		if e := inst.Elem(); e.Kind() == reflect.Struct {
			return unmarshal(e)
		}
	}
	return ErrMarshalUnknownType
}

//======================================================================
// Helper types and methods
//======================================================================

// path keeps track of field "addresses" nested data structures.
// The top-level struct is anonymous and labeled "@". The following path
// elements are the field names as defined in the struct.
type path struct {
	list []string
}

// create a new path with top-level reference set
func newPath() *path {
	p := &path{
		list: make([]string, 0),
	}
	p.push("@")
	return p
}

// push (append) next level
func (p *path) push(elem string) {
	p.list = append(p.list, elem)
}

// pop (remove) last level
func (p *path) pop() (elem string) {
	num := len(p.list) - 1
	elem = p.list[num]
	p.list = p.list[:num]
	return
}

// return human-readable path name
func (p *path) string() string {
	return strings.Join(p.list, ".")
}

//----------------------------------------------------------------------

// read integer based on given endianess
func readInt(rdr io.Reader, tag string, v interface{}) (err error) {
	if tag == "big" {
		err = binary.Read(rdr, binary.BigEndian, v)
	} else {
		err = binary.Read(rdr, binary.LittleEndian, v)
	}
	return
}

// write integer based on given endianess
func writeInt(wrt io.Writer, tag string, v interface{}) (err error) {
	if tag == "big" {
		err = binary.Write(wrt, binary.BigEndian, v)
	} else {
		err = binary.Write(wrt, binary.LittleEndian, v)
	}
	return
}

//----------------------------------------------------------------------

// Get a method from an instance during (un-)marshalling:
// 'inst' refers to the enclosing struct instance that "owns" the field
// being unmarshalled. 'name' either refers to the name of a method of
// the instance ("mthname") or a method of a field (or its subfields)
// previously unmarshalled ("field.mthname", "field.sub. ... .mthdname").
// 'field' must be part of the enclosing instance (sibling of the unmarshalled
// field).
func getMethod(inst reflect.Value, name string) (mth reflect.Value, err error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		return getMethod(inst.FieldByName(parts[0]), parts[1])
	}
	if mth = inst.MethodByName(name); !mth.IsValid() {
		if mth = inst.Addr().MethodByName(name); !mth.IsValid() {
			err = ErrMarshalMthdMissing
		}
	}
	return
}

// call a method (either x.Mthd() or inst.Mthd())
func callMethod(mthName, fldName string, x, inst reflect.Value) (res []reflect.Value, err error) {
	// find method on current struct first
	mth, err := getMethod(x, mthName)
	if err != nil {
		// try to find method in enclosing struct instance
		if mth, err = getMethod(inst, mthName); err != nil {
			return
		}
	}
	// check for string argument
	var args []reflect.Value
	numArgs := mth.Type().NumIn()
	if numArgs > 1 {
		// invalid number of arguments (none or just one string)
		err = ErrMarshalMthdNumArg
		return
	} else if numArgs == 1 {
		// check for string argument
		arg0 := mth.Type().In(0)
		if arg0.Kind() != reflect.String {
			err = ErrMarshalMthdArgType
			return
		}
		// set argument
		fname := reflect.New(reflect.TypeOf("")).Elem()
		fname.SetString(fldName)
		args = append(args, fname)
	}
	// call method
	res = mth.Call(args)
	return
}

// parse number of slice/array elements
func parseSize(tagSize, fldName string, x, inst reflect.Value, inSize, pending int) (count int, err error) {
	// process "size" tag for slice/array
	lts := len(tagSize)
	if lts == 0 {
		// if no size annotation is found, return the incoming length
		return inSize, nil
	}
	if tagSize == "*" {
		if pending >= 0 {
			count = pending
			if count > 0 && lts > 1 && tagSize[1] == '-' {
				off, err := strconv.ParseInt(tagSize[2:], 10, 16)
				if err != nil {
					return 0, err
				}
				if count > int(off) {
					count -= int(off)
				}
			}
		} else {
			count = -1
		}
	} else if tagSize[0] == '(' {
		// method call
		mthName := strings.Trim(tagSize, "()")
		var res []reflect.Value
		if res, err = callMethod(mthName, fldName, x, inst); err != nil {
			return
		}
		if len(res) != 1 || !res[0].CanUint() {
			err = ErrMarshalMthdResult
			return
		}
		count = int(res[0].Uint())
	} else {
		var n int64
		if n, err = strconv.ParseInt(tagSize, 10, 16); err == nil {
			count = int(n)
		} else {
			err = nil
			// previous field value
			ref := x.FieldByName(tagSize)
			if !ref.CanUint() {
				err = ErrMarshalFieldRef
				return
			}
			count = int(ref.Uint())
		}
	}
	// check actual size for expected size
	if inSize > 0 && count > 0 && inSize != count {
		err = ErrMarshalSizeMismatch
		return
	}
	return
}

// isUsed returns true if an optional field is used
func isUsed(tagOpt, fldName string, x, inst reflect.Value) (bool, error) {
	used := true
	if len(tagOpt) > 0 {
		// evaluate condition: must be either variable or function;
		// defaults to false!
		used = false
		if tagOpt[0] == '(' {
			// method call
			mthName := strings.Trim(tagOpt, "()")
			res, err := callMethod(mthName, fldName, x, inst)
			if err != nil {
				return false, err
			}
			if len(res) != 1 {
				return false, ErrMarshalMthdResult
			}
			used = res[0].Bool()
		} else {
			ref := x.FieldByName(tagOpt)
			if ref.Kind() != reflect.Bool {
				return false, ErrMarshalFieldRef
			}
			used = ref.Bool()
		}
	}
	return used, nil
}
