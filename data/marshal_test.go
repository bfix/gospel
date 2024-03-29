package data

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
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
)

//----------------------------------------------------------------------

type NestedStruct struct {
	A int64 `order:"big"`
	B int32
}

func (n *NestedStruct) String() string {
	return fmt.Sprintf("NS(%v)", *n)
}

//----------------------------------------------------------------------

type SubStruct struct {
	G int32
}

func (s *SubStruct) String() string {
	return fmt.Sprintf("SS(%v)", *s)
}

//----------------------------------------------------------------------

type MainStruct struct {
	C uint64 `order:"big"`
	D string
	F *SubStruct
	G uint32
	E []*NestedStruct `size:"G"`
	H []uint32        `size:"5" order:"big"`
}

//----------------------------------------------------------------------

type MthStruct struct {
	A uint16 `order:"big"`
	B []byte `size:"(BSize)"`
}

func (x *MthStruct) BSize() uint {
	if x.A == 1 {
		return 4
	}
	return 16
}

func (x *MthStruct) CSize() uint {
	if x.A == 1 {
		return 9
	}
	return 7
}

//----------------------------------------------------------------------

type EnvelopeStruct struct {
	A string
	B *MthStruct
	C []byte `size:"(B.CSize)"`
}

//----------------------------------------------------------------------

type VarStruct struct {
	A int16
	B []byte `size:"(CalcSize)"`
	C []byte `size:"(CalcSize)"`
	D []byte `size:"*"`
}

func (x *VarStruct) CalcSize(field string) uint {
	fmt.Printf("Handling field '%s'\n", field)
	if x.A > 0 {
		return 3
	} else if x.A < 0 {
		return 5
	} else {
		switch field {
		case "B":
			return 7
		case "C":
			return 9
		}
	}
	return 1
}

//----------------------------------------------------------------------

type OptStruct struct {
	A uint16
	B []byte `opt:"(IsUsed)" size:"A"`
	C bool
	D []byte `opt:"C" size:"23"`
}

func (x *OptStruct) IsUsed() bool {
	return x.A > 10
}

//----------------------------------------------------------------------

type ErrStruct struct {
	A uint32
	B struct {
		C struct {
			D []byte `size:"X"`
		}
	}
}

//----------------------------------------------------------------------

type ImplStruct struct {
	A uint32
}

func (s *ImplStruct) Foo() uint32 {
	return s.A
}

type NilInterface interface {
	Foo() uint32
}

//----------------------------------------------------------------------

type InitStruct struct {
	N   uint16
	A   []uint16 `size:"N"`
	sum uint16
}

func (s *InitStruct) Init() (err error) {
	for _, v := range s.A {
		s.sum += v
	}
	if s.sum%2 != 0 {
		err = errors.New("SUM failed")
	}
	return
}

type InitWrap struct {
	X uint32      `order:"big"`
	Y *InitStruct `init:"Init"`
}

//----------------------------------------------------------------------
// unit tests
//----------------------------------------------------------------------

func TestInit(t *testing.T) {
	run := func(v []uint16) (*InitWrap, error) {
		a := new(InitStruct)
		a.A = v
		a.N = uint16(len(v))
		b := &InitWrap{
			X: 23,
			Y: a,
		}
		data, err := Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		s := new(InitWrap)
		return s, Unmarshal(s, data)
	}
	// case 1
	s, err := run([]uint16{1, 2, 3, 4, 5, 6, 7})
	if err != nil {
		t.Fatal(err)
	}
	if s.Y.sum != 28 {
		t.Logf("sum = %d\n", s.Y.sum)
		t.Fatal("invalid sum")
	}
	// case 2
	if _, err = run([]uint16{1, 2, 3, 4, 5, 6, 7, 23}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNilInterface(t *testing.T) {
	var a NilInterface = (*ImplStruct)(nil)
	if err := Unmarshal(a, []byte{1, 2, 3, 4}); err == nil || errors.Unwrap(err) != ErrMarshalInvalid {
		t.Fatal("expected ErrMarshalInvalid error")
	}
}

func TestNested(t *testing.T) {
	r := new(MainStruct)
	r.C = 19031962
	r.D = "Just a test"
	r.E = make([]*NestedStruct, 3)
	r.F = new(SubStruct)
	r.F.G = 0x23
	r.G = 3
	for i := 0; i < 3; i++ {
		n := new(NestedStruct)
		n.A = int64(255 - i)
		n.B = int32(815 * (i + 1))
		r.E[i] = n
	}
	r.H = make([]uint32, 5)
	for i := range r.H {
		r.H[i] = uint32(i * 23)
	}

	data, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("<<< %v\n", r)
	t.Logf("    [%s]\n", hex.EncodeToString(data))

	s := new(MainStruct)
	if err = Unmarshal(s, data); err != nil {
		t.Fatal(err)
	}
	t.Logf(">>> %v\n", s)
}

func TestMethod(t *testing.T) {
	a := new(MthStruct)
	a.A = 1
	a.B = make([]byte, 4)

	ad, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("    [%s]\n", hex.EncodeToString(ad))

	b := new(MthStruct)
	if err = Unmarshal(b, ad); err != nil {
		t.Fatal(err)
	}

	a = new(MthStruct)
	a.A = 2
	a.B = make([]byte, 16)

	if ad, err = Marshal(a); err != nil {
		t.Fatal(err)
	}
	t.Logf("    [%s]\n", hex.EncodeToString(ad))

	b = new(MthStruct)
	if err = Unmarshal(b, ad); err != nil {
		t.Fatal(err)
	}
}

func TestMethod2(t *testing.T) {
	a := &EnvelopeStruct{
		A: "test",
		B: &MthStruct{
			A: 1,
			B: make([]byte, 4),
		},
		C: make([]byte, 9),
	}

	ad, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("    [%s]\n", hex.EncodeToString(ad))

	b := new(EnvelopeStruct)
	if err = Unmarshal(b, ad); err != nil {
		t.Fatal(err)
	}
}

func TestVar(t *testing.T) {
	a := &VarStruct{
		A: 0,
		B: make([]byte, 7),
		C: make([]byte, 9),
		D: make([]byte, 7),
	}
	ad, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("    [%s]\n", hex.EncodeToString(ad))

	b := new(VarStruct)
	if err = Unmarshal(b, ad); err != nil {
		t.Fatal(err)
	}
	bd, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("    [%s]\n", hex.EncodeToString(bd))
	if !bytes.Equal(ad, bd) {
		t.Fatal("serialization mismatch")
	}
}

func TestOptional(t *testing.T) {
	a := &OptStruct{
		A: 25,
		B: make([]byte, 25),
		C: true,
		D: make([]byte, 23),
	}

	test := func(label string, size int) {
		ad, err := Marshal(a)
		if err != nil {
			t.Fatal(label + ": " + err.Error())
		}
		if len(ad) != size {
			t.Fatalf("%s: size mismatch: %d != %d", label, len(ad), size)
		}

		b := new(OptStruct)
		if err = Unmarshal(b, ad); err != nil {
			t.Fatal(label + ": " + err.Error())
		}
		bd, err := Marshal(a)
		if err != nil {
			t.Fatal(label + ": " + err.Error())
		}
		if !bytes.Equal(ad, bd) {
			t.Fatal(label + ": serialization mismatch")
		}
	}

	test("T1", 51)
	a.A = 3
	test("T2", 26)
	a.C = false
	test("T3", 3)
}

func TestErr(t *testing.T) {
	a := new(ErrStruct)
	a.A = 23
	a.B.C.D = make([]byte, 23)

	_, err := Marshal(a)
	if err == nil {
		t.Fatal("expected error")
	}
	t.Log(err)
}
