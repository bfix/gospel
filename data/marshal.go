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
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

//######################################################################
//
// Serialization of Golang objects of type 'struct{}':
// Field types can be any of these:
//
//    int{8,16,32,64}       -- Signed integer of given size
//    uint{8,16,32,64}      -- Unsigned integer of given size (little-endian)
//    []uint8               -- variable length byte array
//    string                -- variable length string
//    *struct{}, struct{}   -- nested structure
//    []*struct{}, []struct -- list of structures with allowed fields
//
// Integer fields (of size > 1) can be tagged for Big-Endian representation
// by using the tag "order" with a value of "big":
//
//    field1 int64 `order:"big"`
//
// Variable-length slices can be tagged with a "size" tag to help the
// Unmarshal function to figure out the number of slice elements to
// process. The values can be "*" for greedy (as many elements as
// possible before running out of data), "<num>" a decimal number specifying
// the fixed size or two dynamic/variable approaches:
//
// (1) A "<name>" referring to a previous unsigned integer field in the
//     struct object:
//
//     ListSize uint16
//     List     []*Entry `size:"ListSize"`
//
// (2) A "(<name>)" referring to a struct object method, that takes no
//     arguments and returns an unsigned integer for length:
//
//     List     []*Entry `size:"(CalcSize)"`
//
//     The "ValcSize" method works on an incompletely initialized object
//     instance and can only consider data that has already been read.
//
//######################################################################

//======================================================================
// Marshal/unmarshal Golang objects to/from byte arrays.
//======================================================================

// Marshal creates a byte array from a (reference to an) object.
func Marshal(obj interface{}) ([]byte, error) {
	var marshal func(x reflect.Value) ([]byte, error)
	marshal = func(x reflect.Value) ([]byte, error) {
		data := new(bytes.Buffer)
		for i := 0; i < x.NumField(); i++ {
			f := x.Field(i)
			// do not serialize unexported fields
			if !f.CanSet() {
				continue
			}
			ft := x.Type().Field(i)
			switch v := f.Interface().(type) {
			//----------------------------------------------------------
			// Strings
			//----------------------------------------------------------
			case string:
				data.Write([]byte(v))
				data.Write([]byte{0})
			//----------------------------------------------------------
			// Integers
			//----------------------------------------------------------
			case uint8, int8, uint16, int16, uint32, int32, uint64, int64:
				if ft.Tag.Get("order") == "big" {
					binary.Write(data, binary.BigEndian, v)
				} else {
					binary.Write(data, binary.LittleEndian, v)
				}
			//----------------------------------------------------------
			// Byte arrays
			//----------------------------------------------------------
			case []uint8:
				data.Write(v)

			//----------------------------------------------------------
			// Handle other complex types...
			//----------------------------------------------------------
			default:
				switch f.Kind() {
				//------------------------------------------------------
				// Pointers
				//------------------------------------------------------
				case reflect.Ptr:
					e := f.Elem()
					if e.IsValid() {
						sub, err := marshal(f.Elem())
						if err != nil {
							return nil, err
						}
						data.Write(sub)
					}
				//------------------------------------------------------
				// Structs
				//------------------------------------------------------
				case reflect.Struct:
					sub, err := marshal(f)
					if err != nil {
						return nil, err
					}
					data.Write(sub)
				//------------------------------------------------------
				// Slices
				//------------------------------------------------------
				case reflect.Slice:
					for i := 0; i < f.Len(); i++ {
						e := f.Index(i)
						switch e.Kind() {
						//----------------------------------------------
						// Pointer elements
						//----------------------------------------------
						case reflect.Ptr:
							sub, err := marshal(e.Elem())
							if err != nil {
								return nil, err
							}
							data.Write(sub)
						//----------------------------------------------
						// Struct elements
						//----------------------------------------------
						case reflect.Struct:
							sub, err := marshal(e)
							if err != nil {
								return nil, err
							}
							data.Write(sub)
						//----------------------------------------------
						// Intrinsics (strings, integers)
						//----------------------------------------------
						default:
							switch v := e.Interface().(type) {
							case string:
								data.Write([]byte(v))
								data.Write([]byte{0})
							case uint8, int8, uint16, int16, uint32, int32, uint64, int64:
								if ft.Tag.Get("order") == "big" {
									binary.Write(data, binary.BigEndian, v)
								} else {
									binary.Write(data, binary.LittleEndian, v)
								}
							}
						}
					}
				default:
					return nil, fmt.Errorf("Marshal: Unknown field type: %v", f.Type())
				}
			}
		}
		return data.Bytes(), nil
	}
	// process if object is a '*struct{}' or a 'struct{}'
	a := reflect.ValueOf(obj)
	switch a.Kind() {
	case reflect.Ptr:
		e := a.Elem()
		if e.IsValid() {
			return marshal(e)
		}
		return nil, errors.New("Marshal: object is nil")
	case reflect.Struct:
		return marshal(a)
	}
	return nil, errors.New("Marshal: object is not a 'struct{}'")
}

// Unmarshal reads a byte array to fill an object pointed to by 'obj'.
func Unmarshal(obj interface{}, data []byte) error {
	var inst reflect.Value
	buf := bytes.NewBuffer(data)
	var unmarshal func(x reflect.Value) error
	unmarshal = func(x reflect.Value) error {
		for i := 0; i < x.NumField(); i++ {
			f := x.Field(i)
			// skip unexported fields
			if !f.CanSet() {
				continue
			}
			ft := x.Type().Field(i)
			readInt := func(a interface{}) {
				if ft.Tag.Get("order") == "big" {
					binary.Read(buf, binary.BigEndian, a)
				} else {
					binary.Read(buf, binary.LittleEndian, a)
				}
			}
			switch f.Interface().(type) {
			//----------------------------------------------------------
			// Strings
			//----------------------------------------------------------
			case string:
				s := ""
				b := make([]byte, 1)
				for {
					buf.Read(b)
					if b[0] == 0 {
						break
					}
					s += string(b)
				}
				f.SetString(s)
			//----------------------------------------------------------
			// Integers
			//----------------------------------------------------------
			case uint8:
				var a uint8
				binary.Read(buf, binary.LittleEndian, &a)
				f.SetUint(uint64(a))
			case int8:
				var a int8
				binary.Read(buf, binary.LittleEndian, &a)
				f.SetInt(int64(a))
			case uint16:
				var a uint16
				readInt(&a)
				f.SetUint(uint64(a))
			case int16:
				var a int16
				readInt(&a)
				f.SetInt(int64(a))
			case uint32:
				var a uint32
				readInt(&a)
				f.SetUint(uint64(a))
			case int32:
				var a int32
				readInt(&a)
				f.SetInt(int64(a))
			case uint64:
				var a uint64
				readInt(&a)
				f.SetUint(a)
			case int64:
				var a int64
				readInt(&a)
				f.SetInt(a)
			//----------------------------------------------------------
			// Byte arrays
			//----------------------------------------------------------
			case []uint8:
				size := f.Len()
				if size == 0 {
					sizeTag := ft.Tag.Get("size")
					stl := len(sizeTag)
					if stl == 0 {
						return errors.New("missing size tag on field")
					}
					if sizeTag == "*" {
						size = buf.Len()
						if stl > 1 {
							off, err := strconv.ParseInt(sizeTag[1:], 10, 16)
							if err != nil {
								return err
							}
							size += int(off)
						}
					} else if sizeTag[0] == '(' {
						// method call
						mthName := strings.Trim(sizeTag, "()")
						mth, err := getMethod(x, mthName)
						if err != nil {
							if mth, err = getMethod(inst, mthName); err != nil {
								return err
							}
						}
						res := mth.Call(nil)
						if res == nil {
							return errors.New("missing return value from method for array size")
						}
						size = int(res[0].Uint())
					} else {
						n, err := strconv.ParseInt(sizeTag, 10, 16)
						if err == nil {
							size = int(n)
						} else {
							// previous field value
							size = int(x.FieldByName(sizeTag).Uint())
						}
					}
				}
				a := make([]byte, size)
				n, _ := buf.Read(a)
				if n != size {
					return fmt.Errorf("Unmarshal: size mismatch - have %d, got %d", size, n)
				}
				f.SetBytes(a)

			//----------------------------------------------------------
			// Read more complex types.
			//----------------------------------------------------------
			default:
				switch f.Kind() {
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
					count := f.Len()
					add := false
					if count == 0 {
						add = true
						// process "size" tag for slice
						sizeTag := ft.Tag.Get("size")
						stl := len(sizeTag)
						if stl == 0 {
							return errors.New("missing size tag on field")
						}
						if sizeTag == "*" {
							count = -1
						} else if sizeTag[0] == '(' {
							// method call
							mthName := strings.Trim(sizeTag, "()")
							mth, err := getMethod(x, mthName)
							if err != nil {
								if mth, err = getMethod(inst, mthName); err != nil {
									return err
								}
							}
							res := mth.Call(nil)
							count = int(res[0].Uint())
						} else {
							n, err := strconv.ParseInt(sizeTag, 10, 16)
							if err == nil {
								count = int(n)
							} else {
								// previous field value
								count = int(x.FieldByName(sizeTag).Uint())
							}
						}
					}
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
						if buf.Len() == 0 {
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
								buf.Read(b)
								if b[0] == 0 {
									break
								}
								s += string(b)
							}
							e.SetString(s)
						//----------------------------------------------------------
						// Integers
						//----------------------------------------------------------
						case reflect.Int8:
							var a int8
							binary.Read(buf, binary.LittleEndian, &a)
							e.SetInt(int64(a))
						case reflect.Uint16:
							var a uint16
							readInt(&a)
							e.SetUint(uint64(a))
						case reflect.Int16:
							var a int16
							readInt(&a)
							e.SetInt(int64(a))
						case reflect.Uint32:
							var a uint32
							readInt(&a)
							e.SetUint(uint64(a))
						case reflect.Int32:
							var a int32
							readInt(&a)
							e.SetInt(int64(a))
						case reflect.Uint64:
							var a uint64
							readInt(&a)
							e.SetUint(a)
						case reflect.Int64:
							var a int64
							readInt(&a)
							e.SetInt(a)
						}
					}
				default:
					return fmt.Errorf("Unmarshal: Unknown field type: %v", f.Kind())
				}
			}
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
	return fmt.Errorf("Unmarshal: Unknown (field) type: %v", inst.Type())
}

// Helper method to get a method from an instance
func getMethod(inst reflect.Value, name string) (mth reflect.Value, err error) {
	// method call
	if mth = inst.MethodByName(name); !mth.IsValid() {
		if mth = inst.Addr().MethodByName(name); !mth.IsValid() {
			err = errors.New("missing method for array size")
		}
	}
	return
}
