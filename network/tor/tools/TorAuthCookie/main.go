package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	// "/run/tor/control.authcookie" base64-encoded
	enc := "L3J1bi90b3IvY29udHJvbC5hdXRoY29va2ll"
	fname, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file, err := os.Open(string(fname))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(hex.EncodeToString(data))
	os.Exit(0)
}
