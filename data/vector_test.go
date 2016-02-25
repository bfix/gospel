package data

import (
	"testing"
)

func TestVector(t *testing.T) {
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
	}

	vec := NewVector() // ""
	if vec.Len() != 0 {
		t.Fatal("new vector not empty")
	}
	for _, v := range list {
		vec.Add(v)
	}
	if vec.Len() != 5 { // "aa" "bb" "cc" "dd" "ee"
		t.Fatal("length mismatch")
	}
	vec.Insert(-3, "mm") // "mm" nil nil "aa" "bb" "cc" "dd" "ee"
	if vec.At(3).(string) != list[0] {
		t.Fatal("prepending failed")
	}
	if vec.Len() != 8 {
		t.Fatal("length mismatch")
	}
	vec.Insert(10, "nn") // "mm" nil nil "aa" "bb" "cc" "dd" "ee" nil nil "nn"
	if vec.Len() != 11 {
		t.Fatal("size mismatch")
	}
	if vec.Delete(7).(string) != list[4] { // "mm" nil nil "aa" "bb" "cc" "dd" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
		t.Fatal("delete failed")
	}
	if vec.Len() != 10 {
		t.Fatal("length mismatch")
	}
	vec.Insert(5, "pp") // "mm" nil nil "aa" "bb" "pp" "cc" "dd" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
	if vec.Len() != 11 {
		t.Fatal("length mismatch")
	}
}
