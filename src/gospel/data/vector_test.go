package data

///////////////////////////////////////////////////////////////////////
// import external declarations

import (
	"fmt"
	"testing"
)

///////////////////////////////////////////////////////////////////////
// Test case for vectors

func TestVector(t *testing.T) {

	fmt.Println("********************************************")
	fmt.Println("data/vector Test")
	fmt.Println("********************************************")
	fmt.Println()

	// test string vector
	fmt.Println("Testing generic vector with strings")
	list := []string{
		"aa", "bb", "cc", "dd", "ee",
		"ff", "gg", "hh", "ii", "kk",
	}

	vec := NewVector() // ""
	if vec.Len() != 0 {
		t.Fail()
	}
	for _, v := range list {
		vec.Add(v)
	}
	if vec.Len() != 10 { // "aa" "bb" "cc" "dd" "ee" "ff" "gg" "hh" "ii" "kk"
		t.Fail()
	}
	vec.Insert(-3, "mm") // "mm" nil nil "aa" "bb" "cc" "dd" "ee" "ff" "gg" "hh" "ii" "kk"
	if vec.At(3).(string) != list[0] {
		t.Fail()
	}
	if vec.Len() != 13 {
		t.Fail()
	}
	vec.Insert(15, "nn") // "mm" nil nil "aa" "bb" "cc" "dd" "ee" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
	if vec.Len() != 16 {
		t.Fail()
	}
	if vec.Delete(7).(string) != list[4] { // "mm" nil nil "aa" "bb" "cc" "dd" "ff" "gg" "hh" "ii" "kk" nil nil "nn"
		t.Fail()
	}
	if vec.Len() != 15 {
		t.Fail()
	}
}
