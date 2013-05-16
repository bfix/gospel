package data

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"fmt"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Test case for Stacks

func TestStack(t *testing.T) {

	fmt.Println("********************************************")
	fmt.Println("data/stack Test")
	fmt.Println("********************************************")
	fmt.Println()

	// test integer stack
	fmt.Println("Testing integer stack")
	is := NewIntStack()
	if is.Len() != 0 {
		t.Fail()
	}
	for i := 0; i < 10; i++ {
		is.Push(i)
		if !is.IsTop(i) {
			t.Fail()
		}
	}
	if is.Len() != 10 {
		t.Fail()
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if is.Peek() != j {
			t.Fail()
		}
		if is.Pop() != j {
			t.Fail()
		}
		if is.Len() != j {
			t.Fail()
		}
	}

	// test string stack
	fmt.Println("Testing generic stack with strings")
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
		"ff", "gg", "hh", "ii", "kk",
	}
	ss := NewStack()
	if ss.Len() != 0 {
		t.Fail()
	}
	for _, v := range list {
		ss.Push(v)
	}
	if ss.Len() != 10 {
		t.Fail()
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if ss.Peek().(string) != list[j] {
			t.Fail()
		}
		if ss.Pop().(string) != list[j] {
			t.Fail()
		}
		if ss.Len() != j {
			t.Fail()
		}
	}
}
