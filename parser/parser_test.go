//*********************************************************************
//*   PGMID.        TEST PARSER IMPLEMENTATION.                       *
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
	"strconv"
	"strings"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	test variables

var (
	currName    string // current name of machine
	currAddress string // current address of machine
	currPort    string // current port in list
	currService string // current name of service
	countName   int    // current number of machine
	countPort   int    // current number of ports
	rc          bool
	pos         int
	res         = [8]string{
		"Name(1)=\"hades\", Address=192.168.23.254, Port(1)=22, Service=\"ssh\"",
		"Name(1)=\"hades\", Address=192.168.23.254, Port(2)=53, Service=\"dns\"",
		"Name(1)=\"hades\", Address=192.168.23.254, Port(3)=1194, Service=\"openvpn\"",
		"Name(2)=\"olymp\", Address=192.168.23.13, Port(1)=22, Service=\"ssh\"",
		"Name(2)=\"olymp\", Address=192.168.23.13, Port(2)=2401, Service=\"cvspserver\"",
		"Name(2)=\"olymp\", Address=192.168.23.13, Port(3)=990, Service=\"ftps\"",
		"Name(3)=\"saturn\", Address=192.168.23.60, Port(1)=22, Service=\"ssh\"",
		"Name(3)=\"saturn\", Address=192.168.23.60, Port(2)=80, Service=\"http\"",
	}
)

///////////////////////////////////////////////////////////////////////
//	public test method

//---------------------------------------------------------------------
/**
 * Test parser implementation.
 * @param t *testing.T - test handler
 */
//---------------------------------------------------------------------
func TestParser(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("parser/parser Test")
	fmt.Println("********************************************************")

	data := GetTestData1()

	rdr := bufio.NewReader(strings.NewReader(data))
	rc = true
	pos = 0
	err := Parser(rdr, callback)
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Fail()

	} else if !rc {
		t.Fail()
	}
}

//---------------------------------------------------------------------
/**
 * Get test data definition with given id for testing.
 * @param id int - test data identifier
 * @return string - test data definition
 */
//---------------------------------------------------------------------

func GetTestData1() string {

	// assemble test data
	data := "# Test data definition\n"
	data += "Service={\n"
	data += "\tName=\">Y< Test Service\",\n"
	data += "\tAddress=213.3.24.206,\n"
	data += "\tPort=2342\n"
	data += "},\n"
	data += "Host=@/Machines/#1,\n"
	data += "Filter=192.168.23.0/24,\n"
	data += "Machines={\n"
	data += "\t{ Name=\"hades\", Address=192.168.23.254, Ports={\n"
	data += "\t\t{ Port=22, Service=\"ssh\" },\n"
	data += "\t\t{ Port=53, Service=\"dns\" },\n"
	data += "\t\t{ Port=1194, Service=\"openvpn\" }\n"
	data += "\t}},\n"
	data += "\t{ Name=\"olymp\", Address=192.168.23.13, {\n"
	data += "\t\t{ Port=22, Service=\"ssh\" },\n"
	data += "\t\t{ Port=2401, Service=\"cvspserver\" },\n"
	data += "\t\t{ Port=990, Service=\"ftps\" }\n"
	data += "\t}},\n"
	data += "\t{ Name=\"saturn\", Address=192.168.23.60, {\n"
	data += "\t\t{ Port=22, Service=\"ssh\" },\n"
	data += "\t\t{ Port=80, Service=\"http\" }\n"
	data += "\t}}\n"
	data += "}"
	return data
}

///////////////////////////////////////////////////////////////////////
//	private test helper method

//---------------------------------------------------------------------
/**
 * Handle callback from parser.
 * @param mode int - parameter mode
 * @param param *Parameter - reference to new parameter
 * @return bool - successful operation?
 */
//---------------------------------------------------------------------

func callback(mode int, param *Parameter) bool {

	// if parameter is specified
	if param != nil {

		// print incoming parameter
		// fmt.Printf ("%d: `%v=%v`\n", mode, param.Name, param.Value)

		if mode != LIST {
			switch param.Name {
			case "Name":
				currName = param.Value
				countName++
				countPort = 0
			case "Address":
				currAddress = param.Value
			case "Port":
				currPort = param.Value
				countPort++
			case "Service":
				{
					msg := "Name(" + strconv.Itoa(countName) + ")=" + currName +
						", Address=" + currAddress +
						", Port(" + strconv.Itoa(countPort) + ")=" + currPort +
						", Service=" + param.Value
					fmt.Println(msg)
					rc = rc && (msg == res[pos])
					pos++
				}
			}
		} else if param.Name == "Machines" {
			countName = 0
		}
	}
	return rc
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-09 16:03:23  brf
//  First release as free software (GPL3+)
//
//	Revision 1.6  2010-11-28 13:14:26  brf
//	New test layout.
//
