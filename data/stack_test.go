package data

import (
	"testing"
)

func TestIntStack(t *testing.T) {
	is := NewIntStack()
	if is.Len() != 0 {
		t.Fatal("new stack not empty")
	}
	for i := 0; i < 10; i++ {
		is.Push(i)
		if !is.IsTop(i) {
			t.Fatal("push/istop failed")
		}
	}
	if is.Len() != 10 {
		t.Fatal("length mismatch")
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if is.Peek() != j {
			t.Fatal("peek failed")
		}
		if is.Pop() != j {
			t.Fatal("pop failed")
		}
		if is.Len() != j {
			t.Fatal("length mismatch")
		}
	}
}

func TestStringStack(t *testing.T) {
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
		"ff", "gg", "hh", "ii", "kk",
	}
	ss := NewStack()
	if ss.Len() != 0 {
		t.Fatal("new stack not empty")
	}
	for _, v := range list {
		ss.Push(v)
	}
	if ss.Len() != 10 {
		t.Fatal("length mismatch")
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if ss.Peek().(string) != list[j] {
			t.Fatal("peek failed")
		}
		if ss.Pop().(string) != list[j] {
			t.Fatal("pop failed")
		}
		if ss.Len() != j {
			t.Fatal("length mismatch")
		}
	}
}
