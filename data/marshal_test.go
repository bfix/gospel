package data

import (
	"encoding/hex"
	"fmt"
	"testing"
)

type NestedStruct struct {
	A int64 `order:"big"`
	B int32
}

func (n *NestedStruct) String() string {
	return fmt.Sprintf("%v", *n)
}

type SubStruct struct {
	G int32
}

func (s *SubStruct) String() string {
	return fmt.Sprintf("%v", *s)
}

type MainStruct struct {
	C uint64 `order:"big"`
	D string
	F *SubStruct
	G uint32
	E []*NestedStruct `size:"G"`
}

func TestNested(t *testing.T) {
	r := new(MainStruct)
	r.C = 19031962
	r.D = "Just a test"
	r.E = make([]*NestedStruct, 3)
	r.F = new(SubStruct)
	r.G = 3
	r.F.G = 0x23
	for i := 0; i < 3; i++ {
		n := new(NestedStruct)
		n.A = int64(255 - i)
		n.B = int32(815 * (i + 1))
		r.E[i] = n
	}

	data, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("<<< %v\n", r)
	fmt.Printf("    [%s]\n", hex.EncodeToString(data))

	s := new(MainStruct)
	if err = Unmarshal(s, data); err != nil {
		t.Fatal(err)
	}
	fmt.Printf(">>> %v\n", s)
}
