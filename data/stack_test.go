package data

import (
	"testing"
)

func TestIntStack(t *testing.T) {
	is := NewIntStack()
	if is.Len() != 0 {
		t.Fatal()
	}
	for i := 0; i < 10; i++ {
		is.Push(i)
		if !is.IsTop(i) {
			t.Fatal()
		}
	}
	if is.Len() != 10 {
		t.Fatal()
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if is.Peek() != j {
			t.Fatal()
		}
		if is.Pop() != j {
			t.Fatal()
		}
		if is.Len() != j {
			t.Fatal()
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
		t.Fatal()
	}
	for _, v := range list {
		ss.Push(v)
	}
	if ss.Len() != 10 {
		t.Fatal()
	}
	for i := 0; i < 10; i++ {
		j := 9 - i
		if ss.Peek().(string) != list[j] {
			t.Fatal()
		}
		if ss.Pop().(string) != list[j] {
			t.Fatal()
		}
		if ss.Len() != j {
			t.Fatal()
		}
	}
}
