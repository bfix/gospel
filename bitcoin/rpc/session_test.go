package rpc

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	sess    *Session
	info    *Info
	verbose = false
)

func init() {
	fmt.Printf("init(): ")
	var err error
	rpcaddr := os.Getenv("BTC_HOST")
	user := os.Getenv("BTC_USER")
	passwd := os.Getenv("BTC_PASSWORD")
	if sess, err = NewSession(rpcaddr, user, passwd); err != nil {
		fmt.Println(err.Error())
		sess = nil
		return
	}
	strictCheck = true
	if info, err = sess.GetInfo(); err != nil {
		fmt.Println(err.Error())
		sess = nil
		return
	}
	fmt.Println("Bitcoin JSON-RPC tests initialized!")
}

func dumpObj(fmtStr string, v interface{}) {
	data, err := json.Marshal(v)
	if err == nil {
		fmt.Printf(fmtStr, string(data))
	} else {
		fmt.Printf(fmtStr, "<"+err.Error()+">")
	}
}
