package network

import (
	//	"bufio"
	"testing"
)

func TestSocks5(t *testing.T) {

	conn, err := Socks5Connect("tcp", "www.google.com", 80, "127.0.0.1:9050")
	if conn == nil || err != nil {
		t.Fatal("failed to connect to tor proxy")
	}
	/*
		conn.Write([]byte("GET / HTTP/1.0\n\n"))
		rdr := bufio.NewReader(conn)
		for true {
			data, _, err := rdr.ReadLine()
			if err != nil || rdr.Buffered() == 0 {
				break
			}
		}
	*/
	conn.Close()
}
