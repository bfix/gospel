//*********************************************************************
//*   PGMID.        TEST SOCKS5 OPERATIONS.                           *
//*   AUTHOR.       BERND R. FIX   >Y<                                *
//*   DATE WRITTEN. 12/08/08.                                         *
//*   COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.         *
//*                 LICENSED MATERIAL - PROGRAM PROPERTY OF THE       *
//*                 AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.          *
//*********************************************************************

package network

///////////////////////////////////////////////////////////////////////
// Import external declarations

import (
	"bufio"
	"fmt"
	"testing"
)

///////////////////////////////////////////////////////////////////////
//	public test method

//---------------------------------------------------------------------
/**
 * Run test suite for SOCKS5 connections.
 * @param t *testing.T - test handler
 */
//---------------------------------------------------------------------
func TestSocks5(t *testing.T) {

	fmt.Println("********************************************************")
	fmt.Println("network/socks Test")
	fmt.Println("********************************************************")

	conn, err := Socks5Connect("tcp", "www.google.com", 80, "127.0.0.1:9050")
	if conn == nil || err != nil {
		fmt.Println("Error connecting through proxy.")
		t.Fail()
		return
	}

	conn.Write([]byte("GET / HTTP/1.0\n\n"))
	rdr := bufio.NewReader(conn)
	for true {
		data, _, err := rdr.ReadLine()
		if err != nil || rdr.Buffered() == 0 {
			break
		}
		fmt.Println(string(data))
	}

	conn.Close()
}
