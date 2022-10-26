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

package data

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
// Serialization of Golang objects of the following types:
//
//    int{8,16,32,64}       -- Signed integer of given size
//    uint{8,16,32,64}      -- Unsigned integer of given size (little-endian)
//    []uint8,[]byte        -- variable length byte array (special handling)
//    string                -- variable length string (\0 -terminated)
//    bool                  -- boolean value (single byte, 0=false, true otherwise)
//    *struct{}, struct{}   -- nested structure
//    []*T, []T             -- array of supported types
//
// The serialization can be controlled by field annotations:
//
// ---------------------------------------
// (1) Endianness of integers: tag "order"
// ---------------------------------------
// Integer objects (of size > 1) can be tagged for Big-Endian representation
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
// N.B.: You can't do math in the size expression (except for byte arrays
//       with the greedy expression like "*-16"); you need to out-source
//       the calculation to a method if required.
//
// ------------------------------
// (3) Optional fields: tag "opt"
// ------------------------------
// A field can be marked as optional with the "opt" tag; the tag value
// must be either a (previous) boolean field or a method returning a
// bool value. As with size functions, the method can take a single
// string argument (field name).
//
// ------------------------------
// (4) Initialization method
// ------------------------------
// A struct field can have a "init" tag; the tag value must be name
// of a struct method used to initialize the instance after
// unmarshalling the binary representation.
//
//######################################################################

// Errors
var (
	ErrMarshalNil           = errors.New("object is nil")
	ErrMarshalInvalid       = errors.New("object is invalid")
	ErrMarshalType          = errors.New("invalid object type")
	ErrMarshalNoSize        = errors.New("missing/invalid size tag on field")
	ErrMarshalNoOpt         = errors.New("missing/invalid opt tag on field")
	ErrMarshalSizeMismatch  = errors.New("size mismatch during unmarshal")
	ErrMarshalEmptyIntf     = errors.New("can't handle empty interface")
	ErrMarshalUnknownType   = errors.New("unknown field type")
	ErrMarshalMthdMissing   = errors.New("missing method")
	ErrMarshalFieldRef      = errors.New("field reference invalid")
	ErrMarshalMthdNumArg    = errors.New("method has more than one argument")
	ErrMarshalMthdArgType   = errors.New("method argument not a string")
	ErrMarshalMthdResult    = errors.New("invalid method result")
	ErrMarshalParentMissing = errors.New("parent missing")
)

//======================================================================
// Marshal Golang objects to byte arrays.
//======================================================================

// Marshal creates a byte array from an object.
func Marshal(obj interface{}) ([]byte, error) {
	// Wrapping stream marshaller with buffer.
	wrt := new(bytes.Buffer)
	if err := MarshalStream(wrt, obj); err != nil {
		return nil, err
	}
	return wrt.Bytes(), nil
}

// MarshalStream writes an object to stream
func MarshalStream(wrt io.Writer, obj interface{}) error {
	inst := reflect.ValueOf(obj)
	ctx := _NewMarshalContext(wrt, inst)
	return marshalValue(ctx, inst)
}

// marshal a single value instance
func marshalValue(ctx *_MarshalContext, v reflect.Value) error {
	// try intrinsic types first
	if ok, err := marshalIntrinsic(ctx, v); ok {
		return err
	}
	// try complex types next
	if ok, err := marshalComplex(ctx, v); ok {
		return err
	}
	// custom type
	if ok, err := marshalCustom(ctx, v); ok {
		return err
	}
	// unknown type
	return gerr.New(ErrMarshalUnknownType,
		"marshal: field '%s', type '%v', kind '%v'",
		ctx.string(), v.Type(), v.Kind())
}

// marshal intrinsic data type
func marshalIntrinsic(ctx *_MarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	tagOrder := ctx.tag("order")
	switch v := f.Interface().(type) {
	//----------------------------------------------------------
	// Strings
	//----------------------------------------------------------
	case string:
		if _, err = ctx.wrt.Write([]byte(v)); err != nil {
			err = ctx.fail(err)
			break
		}
		if _, err = ctx.wrt.Write([]byte{0}); err != nil {
			err = ctx.fail(err)
		}
	//----------------------------------------------------------
	// Booleans
	//----------------------------------------------------------
	case bool:
		var a byte
		if v {
			a = 1
		}
		if _, err = ctx.wrt.Write([]byte{a}); err != nil {
			err = ctx.fail(err)
		}
	//----------------------------------------------------------
	// Integers
	//----------------------------------------------------------
	case uint8, int8, uint16, int16, uint32, int32, uint64, int64, int:
		if err = writeInt(ctx.wrt, tagOrder, v); err != nil {
			err = ctx.fail(err)
		}
	//----------------------------------------------------------
	// Byte arrays
	//----------------------------------------------------------
	case []uint8:
		if _, err = ctx.parseSize(len(v)); err != nil {
			err = ctx.fail(err)
			break
		}
		if _, err = ctx.wrt.Write(v); err != nil {
			err = ctx.fail(err)
		}
	default:
		ok = false
	}
	return
}

// marshal complex data type
func marshalComplex(ctx *_MarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	switch f.Kind() {
	//------------------------------------------------------
	// Interfaces
	//------------------------------------------------------
	case reflect.Interface:
		e := f.Elem()
		if e.IsValid() {
			if e.Kind() == reflect.Ptr {
				e = e.Elem()
			}
			if e.IsValid() {
				ctx.use(e)
				if err = marshalValue(ctx, e); err != nil {
					return
				}
			}
		}
	//------------------------------------------------------
	// Pointers
	//------------------------------------------------------
	case reflect.Ptr:
		e := f.Elem()
		if e.IsValid() {
			ctx.use(e)
			if err = marshalValue(ctx, e); err != nil {
				return
			}
		}
	//------------------------------------------------------
	// Structs
	//------------------------------------------------------
	case reflect.Struct:
		if err = marshalStruct(ctx, f); err != nil {
			return
		}
	//------------------------------------------------------
	// Slices
	//------------------------------------------------------
	case reflect.Slice:
		var count int
		if count, err = ctx.parseSize(f.Len()); err != nil {
			err = ctx.fail(err)
			return
		}
		// greedy slice: use existing size
		if count < 0 {
			count = f.Len()
		}
		for i := 0; i < count; i++ {
			e := f.Index(i)
			if err = marshalValue(ctx, e); err != nil {
				return
			}
		}
	default:
		ok = false
	}
	return
}

// marshalStruct a single value
func marshalStruct(ctx *_MarshalContext, x reflect.Value) error {
	for i := 0; i < x.NumField(); i++ {
		f := x.Field(i)
		// do not serialize unexported fields
		if !f.CanSet() {
			continue
		}
		// append field name to path
		ft := x.Type().Field(i)
		ctx.push(ft.Name, f, ft.Tag)

		// check for optional field
		used, err := ctx.isUsed()
		if err != nil {
			return ctx.fail(err)
		}
		if used {
			if err := marshalValue(ctx, f); err != nil {
				return err
			}
		}
		// remove field name from path
		ctx.pop()
	}
	return nil
}

// marshal custom data type
func marshalCustom(ctx *_MarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	var v interface{}
	switch f.Kind() {
	case reflect.Int, reflect.Int32:
		v = int32(f.Int())
	case reflect.Int8:
		v = int8(f.Int())
	case reflect.Int16:
		v = int16(f.Int())
	case reflect.Int64:
		v = f.Int()
	case reflect.Uint, reflect.Uint32:
		v = uint32(f.Uint())
	case reflect.Uint8:
		v = uint8(f.Uint())
	case reflect.Uint16:
		v = uint16(f.Uint())
	case reflect.Uint64:
		v = f.Uint()
	case reflect.Bool:
		v = f.Bool()
	case reflect.String:
		v = f.String()
	default:
		ok = false
	}
	if ok {
		e := reflect.ValueOf(v)
		err = marshalValue(ctx, e)
	}
	return

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
	inst := reflect.ValueOf(obj)
	ctx := _NewUnmarshalContext(rdr, pending, inst)
	return unmarshalValue(ctx, inst)
}

// unmarshal a single value instance
func unmarshalValue(ctx *_UnmarshalContext, v reflect.Value) error {
	// try intrinsic types first
	if ok, err := unmarshalIntrinsic(ctx, v); ok {
		return err
	}
	// try complex types next
	if ok, err := unmarshalComplex(ctx, v); ok {
		return err
	}
	// custom types
	if ok, err := unmarshalCustom(ctx, v); ok {
		return err
	}
	// unknown type
	return gerr.New(ErrMarshalUnknownType,
		"unmarshal: field '%s', type '%v', kind '%v'",
		ctx.string(), v.Type(), v.Kind())
}

// unmarshal intrinsic data types
func unmarshalIntrinsic(ctx *_UnmarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	tagOrder := ctx.tag("order")
	switch f.Interface().(type) {
	//----------------------------------------------------------
	// Strings
	//----------------------------------------------------------
	case string:
		s := ""
		b := make([]byte, 1)
		for {
			if _, err = ctx.rdr.Read(b); err != nil {
				err = ctx.fail(err)
				return
			}
			if b[0] == 0 {
				break
			}
			s += string(b)
		}
		f.SetString(s)
		ctx.pending -= len(s) + 1
	//----------------------------------------------------------
	// Booleans
	//----------------------------------------------------------
	case bool:
		b := make([]byte, 1)
		if _, err = ctx.rdr.Read(b); err != nil {
			err = ctx.fail(err)
			return
		}
		var a bool
		if b[0] != 0 {
			a = true
		}
		f.SetBool(a)
		ctx.pending--
	//----------------------------------------------------------
	// Integers
	//----------------------------------------------------------
	case uint8:
		var a uint8
		if err = binary.Read(ctx.rdr, binary.LittleEndian, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetUint(uint64(a))
		ctx.pending--
	case int8:
		var a int8
		if err = binary.Read(ctx.rdr, binary.LittleEndian, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetInt(int64(a))
		ctx.pending--
	case uint16:
		var a uint16
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetUint(uint64(a))
		ctx.pending -= 2
	case int16:
		var a int16
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetInt(int64(a))
		ctx.pending -= 2
	case uint32:
		var a uint32
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetUint(uint64(a))
		ctx.pending -= 4
	case int32, int:
		var a int32
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetInt(int64(a))
		ctx.pending -= 4
	case uint64:
		var a uint64
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetUint(a)
		ctx.pending -= 8
	case int64:
		var a int64
		if err = readInt(ctx.rdr, tagOrder, &a); err != nil {
			err = ctx.fail(err)
			return
		}
		f.SetInt(a)
		ctx.pending -= 8

	//----------------------------------------------------------
	// Byte arrays
	//----------------------------------------------------------
	case []uint8:
		var size int
		if size, err = ctx.parseSize(f.Len()); err != nil {
			err = ctx.fail(err)
			return
		}
		a := make([]byte, size)
		var n int
		if n, err = ctx.rdr.Read(a); err != nil {
			err = ctx.fail(err)
			return
		}
		if n != size {
			err = ctx.fail(ErrMarshalSizeMismatch)
		}
		f.SetBytes(a)
		ctx.pending -= n

	default:
		ok = false
	}
	return
}

// unmarshal data struct
func unmarshalStruct(ctx *_UnmarshalContext, x reflect.Value) error {
	for i := 0; i < x.NumField(); i++ {
		f := x.Field(i)
		ft := x.Type().Field(i)
		// skip unexported fields
		if !f.CanSet() {
			continue
		}
		ctx.push(ft.Name, f, ft.Tag)

		// check for optional field
		used, err := ctx.isUsed()
		if err != nil {
			return ctx.fail(err)
		}
		if used {
			// unmarshal data
			if err := unmarshalValue(ctx, f); err != nil {
				return err
			}
			// check for initialization method
			if init := ft.Tag.Get("init"); len(init) > 0 {
				ret, err := ctx.callFieldMethod(f, init)
				if err != nil {
					return err
				}
				if len(ret) == 1 && ret[0].CanInterface() {
					if err, _ = ret[0].Interface().(error); err != nil {
						err = ctx.fail(err)
						return err
					}
				}
			}
		}
		ctx.pop()
	}
	return nil
}

// unmarshal complex type
func unmarshalComplex(ctx *_UnmarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	switch f.Kind() {
	//------------------------------------------------------
	// Interfaces
	//------------------------------------------------------
	case reflect.Interface:
		e := f.Elem()
		if !e.IsValid() {
			err = ctx.fail(ErrMarshalInvalid)
			return
		}
		ctx.use(e)
		if err = unmarshalValue(ctx, e); err != nil {
			return
		}
	//------------------------------------------------------
	// Pointers
	//------------------------------------------------------
	case reflect.Ptr:
		e := f.Elem()
		if !e.IsValid() {
			if !f.CanSet() {
				err = ctx.fail(ErrMarshalInvalid)
				return
			}
			ep := reflect.New(f.Type().Elem())
			e = ep.Elem()
			f.Set(ep)
		}
		ctx.use(e)
		if err = unmarshalValue(ctx, e); err != nil {
			return
		}
	//------------------------------------------------------
	// Structs
	//------------------------------------------------------
	case reflect.Struct:
		if err = unmarshalStruct(ctx, f); err != nil {
			return
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
		var count int
		if count, err = ctx.parseSize(f.Len()); err != nil {
			err = ctx.fail(err)
			return
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
			if ctx.pending < 1 {
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

			// unmarshal element
			if err = unmarshalValue(ctx, e); err != nil {
				return
			}
		}
	default:
		ok = false
	}
	return
}

// unmarshal custom data type
func unmarshalCustom(ctx *_UnmarshalContext, f reflect.Value) (ok bool, err error) {
	ok = true
	var v interface{}
	switch f.Kind() {
	case reflect.Int, reflect.Int32:
		v = *new(int32)
	case reflect.Int8:
		v = *new(int8)
	case reflect.Int16:
		v = *new(int16)
	case reflect.Int64:
		v = *new(int64)
	case reflect.Uint, reflect.Uint32:
		v = *new(uint32)
	case reflect.Uint8:
		v = *new(uint8)
	case reflect.Uint16:
		v = *new(uint16)
	case reflect.Uint64:
		v = *new(uint64)
	case reflect.Bool:
		v = *new(bool)
	case reflect.String:
		v = *new(string)
	default:
		ok = false
	}
	if ok {
		n := reflect.New(reflect.TypeOf(v)).Elem()
		if err = unmarshalValue(ctx, n); err == nil {
			f.Set(n.Convert(f.Type()))
		}
	}
	return
}

//======================================================================
// Helper types and methods
//======================================================================

//----------------------------------------------------------------------
// Context for marshalling operations:
// Keep track of nested struct fields while traversing objects.
//----------------------------------------------------------------------

// path _Element
type _Element struct {
	name  string            // field name
	value reflect.Value     // field value
	tags  reflect.StructTag // field tag
}

// _Context keeps track of fields in nested data structures.
// The top-level struct is anonymous and labeled "@".
type _Context struct {
	path []*_Element
	num  int
}

// create a new path with top-level reference set
func _NewContext(inst reflect.Value) *_Context {
	p := &_Context{
		path: make([]*_Element, 0),
		num:  0,
	}
	p.push("@", inst, "")
	return p
}

// push (append) next level
func (c *_Context) push(name string, value reflect.Value, tag reflect.StructTag) {
	e := &_Element{name, value, tag}
	c.path = append(c.path, e)
	c.num++
}

// update current value
func (c *_Context) use(v reflect.Value) {
	c.path[c.num-1].value = v
}

// pop (remove) last level
//
//nolint:unparam // skip false-positive
func (c *_Context) pop() (e *_Element) {
	c.num--
	e = c.path[c.num]
	c.path = c.path[:c.num]
	return
}

// get current tags
func (c *_Context) tag(name string) string {
	return c.path[c.num-1].tags.Get(name)
}

// return human-readable path name
func (c *_Context) string() string {
	list := make([]string, len(c.path))
	for i, e := range c.path {
		list[i] = e.name
	}
	return strings.Join(list, ".")
}

// return error instance for current path
func (c *_Context) fail(err error, mode string) error {
	return gerr.New(err, "%s: field '%s'", mode, c.string())
}

// parse number of slice/array elements
func (c *_Context) parseSize(inSize, pending int) (count int, err error) {
	tagSize := c.tag("size")

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
				var off int64
				if off, err = strconv.ParseInt(tagSize[2:], 10, 16); err != nil {
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
		if res, err = c.callMethod(mthName); err != nil {
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
		} else if c.num > 1 {
			err = nil
			// previous field value
			ref := c.path[c.num-2].value.FieldByName(tagSize)
			if !ref.CanUint() {
				err = ErrMarshalFieldRef
				return
			}
			count = int(ref.Uint())
		} else {
			err = ErrMarshalNoSize
			return
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
func (c *_Context) isUsed() (bool, error) {
	used := true
	tagOpt := c.tag("opt")
	if len(tagOpt) > 0 {
		// evaluate condition: must be either variable or function;
		// defaults to false!
		if tagOpt[0] == '(' {
			// method call
			mthName := strings.Trim(tagOpt, "()")
			res, err := c.callMethod(mthName)
			if err != nil {
				return false, err
			}
			if len(res) != 1 {
				return false, ErrMarshalMthdResult
			}
			used = res[0].Bool()
		} else if c.num > 1 {
			ref := c.path[c.num-2].value.FieldByName(tagOpt)
			if ref.Kind() != reflect.Bool {
				return false, ErrMarshalFieldRef
			}
			used = ref.Bool()
		} else {
			return true, ErrMarshalNoOpt
		}
	}
	return used, nil
}

// Get a method from an instance during (un-)marshalling:
// 'inst' refers to the enclosing struct instance that "owns" the field
// being (un-)marshalled. 'name' either refers to the name of a method of
// the instance ("mthname") or a method of a field (or its subfields)
// previously unmarshalled ("field.mthname", "field.sub. ... .mthdname").
// 'field' must be part of the enclosing instance (sibling of the unmarshalled
// field).
func (c *_Context) getMethod(inst reflect.Value, name string) (mth reflect.Value, err error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		return c.getMethod(inst.FieldByName(parts[0]), parts[1])
	}
	if mth = inst.MethodByName(name); !mth.IsValid() {
		if mth = inst.Addr().MethodByName(name); !mth.IsValid() {
			err = ErrMarshalMthdMissing
		}
	}
	return
}

// call a method on instance with no arguments (instance-internal)
func (c *_Context) callFieldMethod(inst reflect.Value, name string) (res []reflect.Value, err error) {
	// get method (try instance and pointer receiver)
	var mth reflect.Value
	if mth = inst.MethodByName(name); !mth.IsValid() {
		if mth = inst.Addr().MethodByName(name); !mth.IsValid() {
			err = ErrMarshalMthdMissing
			return
		}
	}
	// call method
	var args []reflect.Value
	res = mth.Call(args)
	return
}

// call a method (either x.Mthd() or inst.Mthd())
func (c *_Context) callMethod(mthName string) (res []reflect.Value, err error) {
	// find method on current struct first
	if c.num < 2 {
		err = ErrMarshalParentMissing
		return
	}
	fldName := c.path[c.num-1].name
	mth, err := c.getMethod(c.path[c.num-2].value, mthName)
	if err != nil {
		// try to find method in enclosing struct instance
		if mth, err = c.getMethod(c.path[0].value, mthName); err != nil {
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

//----------------------------------------------------------------------
// Context for marshalling operations
//----------------------------------------------------------------------

// context for marshalling
type _MarshalContext struct {
	*_Context
	wrt io.Writer
}

// create a new marshal context
func _NewMarshalContext(wrt io.Writer, inst reflect.Value) *_MarshalContext {
	return &_MarshalContext{
		_Context: _NewContext(inst),
		wrt:      wrt,
	}
}

// fail wrapper for marshalling
func (c *_MarshalContext) fail(err error) error {
	return c._Context.fail(err, "marshal")
}

// parse size for marshal operation
func (c *_MarshalContext) parseSize(inSize int) (count int, err error) {
	return c._Context.parseSize(inSize, -1)
}

//----------------------------------------------------------------------

// context for unmarshalling
type _UnmarshalContext struct {
	*_Context
	rdr     io.Reader
	pending int
}

// create a new unmarshal context
func _NewUnmarshalContext(rdr io.Reader, pending int, inst reflect.Value) *_UnmarshalContext {
	return &_UnmarshalContext{
		_Context: _NewContext(inst),
		rdr:      rdr,
		pending:  pending,
	}
}

// fail wrapper for unmarshalling
func (c *_UnmarshalContext) fail(err error) error {
	return c._Context.fail(err, "unmarshal")
}

// parse size for unmarshal operation
func (c *_UnmarshalContext) parseSize(inSize int) (count int, err error) {
	pending := -1
	switch c.path[c.num-1].value.Interface().(type) {
	case []uint8:
		pending = c.pending
	}
	return c._Context.parseSize(inSize, pending)
}

//----------------------------------------------------------------------
// helper functions
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
