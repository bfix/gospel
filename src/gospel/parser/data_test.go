//*********************************************************************
//*   PGMID.        TEST NESTED DATA STRUCTURE IMPLEMENTATION.        *
//*   AUTHOR.       BERND R. FIX   >Y<                                *
//*   DATE WRITTEN. 10/11/01.                                         *
//*   COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.         *
//*                 LICENSED MATERIAL - PROGRAM PROPERTY OF THE       *
//*                 AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.          *
//*   REMARKS.      REVISION HISTORY AT END OF FILE.                  *
//*********************************************************************

package parser

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

//---------------------------------------------------------------------
/**
 * Run test suite for Data implementation.
 * @param t *testing.T - test handler
 */
//---------------------------------------------------------------------
func TestData(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("parser/data Test")
	fmt.Println("********************************************************")

	//=================================================================
	// Basic 
	//=================================================================
	testCase1("Simple=Test", "/Simple", "", "Simple=\"Test\"", t, "111")
	testCase1("Test", "/#1", "", "\"Test\"", t, "111")
	testCase1("\"Test\"", "/#1", "", "", t, "111")
	testCase1("Simple=\"Test\"", "/Simple", "", "", t, "111")
	testCase1("Simple=\"Test mit Spaces\"", "/#1", "/Simple", "", t, "111")
	testCase1("Simple=\"Test,mit=Spaces{}\"", "/Simple", "", "", t, "111")

	testCase1("Simple=Test,", "/Simple", "", "Simple=\"Test\"", t, "111")
	testCase1("Simple=Test,Muster", "/#2", "", "\"Muster\"", t, "111")
	testCase1("Simple=Test", "/#3", "", "", t, "10X")
	testCase1("Simple=Test", "/#-1", "", "", t, "10X")

	testCase1("List={Entry1=Value1,Entry2=Value2}", "/List", "", "List={}", t, "111")
	testCase1("List={Entry1=Value1,Entry2=Value2}", "/List/#1", "/List/Entry1", "Entry1=\"Value1\"", t, "111")
	testCase1("List={Entry1=Value1,{{", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,{{}", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,{{}}", "", "", "", t, "0XX")
	testCase1("List={Entry1=Value1,,{}}", "/List/#2", "", "~", t, "111")
	testCase1("List={}", "/List", "", "", t, "111")
	testCase1("List={{}}", "/List", "", "List={}", t, "111")
	testCase1("List={,,,}", "/List/#1", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#2", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#3", "", "~", t, "111")
	testCase1("List={,,,}", "/List/#4", "", "~", t, "111")
}

///////////////////////////////////////////////////////////////////////
//	private helper methods

var err error

//---------------------------------------------------------------------
/**
 * Parse string into data object.
 * @param s string - data definition
 * @param t *testing.T - test handler
 * @return *Data - parsed data object 
 */
//---------------------------------------------------------------------
func getData(s string, t *testing.T) *Data {

	rdr := bufio.NewReader(strings.NewReader(s))
	d := new(Data)
	err = d.Read(rdr)
	if err != nil {
		return nil
	}
	return d
}

//---------------------------------------------------------------------
/**
 * Run test case.
 * @param data string - data definition 
 * @param access string - lookup key  
 * @param path string - expected path of lookup element  
 * @param elem string - printable lookup element  
 * @param t *testing.T - test handler
 * @param flags string - condition flags  
 */
//---------------------------------------------------------------------
func testCase1(data, access, path, elem string, t *testing.T, flags string) {

	if len(elem) == 0 {
		elem = data
	}
	if len(path) == 0 {
		path = access
	}

	fmt.Printf("<Read>")
	d := getData(data, t)
	if (flags[0] == '1') != (d != nil) {
		fmt.Println(" *Failed*")
		if err != nil {
			fmt.Printf("*** %v\n", err)
		}
		t.Fail()
		return
	}
	if d != nil {

		e := d.Lookup(access)
		p := ""
		if e != nil {
			p = e.GetPath()
		}
		fmt.Printf(", Path: \"%s\"", p)
		if (flags[1] == '1') != (p == path) {
			fmt.Println(" *Failed*")
			t.Fail()
			return
		}
		if e != nil {
			p = e.String()
			fmt.Printf(", Elem: %s", p)
			if (flags[2] == '1') != (p == ("`" + elem + "`")) {
				fmt.Println(" *Failed*")
				t.Fail()
				return
			}
		}
	}
	fmt.Println(" [O.K.]")
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-09 16:03:23  brf
//  First release as free software (GPL3+)
//
//	Revision 1.7  2010-11-27 22:29:32  brf
//	New test layout.
//
