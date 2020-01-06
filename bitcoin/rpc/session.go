package rpc

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

var (
	strictCheck = false
)

// Data is a generic data structure for RPC data (in/out)
type Data interface{}

// Request is a JSON-RPC task to a running Bitcoin server
type Request struct {
	Version string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []Data `json:"params"`
}

// Response is a JSON-encoded reply from a running Bitcoin server
type Response struct {
	Result Data  `json:"result"`
	Error  Error `json:"error"`
}

// UnmarshalResult will unmarshal the Result field to
// a JSON data structure.
func (r *Response) UnmarshalResult(v interface{}) (bool, error) {
	if r.Result == nil {
		return false, nil
	}
	data, err := json.Marshal(r.Result)
	if err != nil {
		return false, err
	}
	if err = json.Unmarshal(data, v); err != nil {
		return false, err
	}
	if strictCheck {
		rc, msg := checkJSON(r.Result, v)
		if !rc {
			return false, errors.New(">>>>>\n" + msg)
		}
		rc, msg = checkJSON(v, r.Result)
		if !rc {
			return false, errors.New("<<<<<\n" + msg)
		}
	}
	return true, nil
}

// Error is a Response-related failure code.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Session type
type Session struct {
	address    string            // server address/name
	user       string            // user name
	passwd     string            // user password
	serverCert *x509.Certificate // server certificate for SSL
	client     *http.Client      // HTTP client instance
}

// NewSession allocates a new Session instance for communication
func NewSession(addr, user, pw string) (*Session, error) {
	if _, err := url.Parse(addr); err != nil {
		return nil, err
	}
	if len(user) == 0 || len(pw) == 0 {
		return nil, errors.New("Missing credentials")
	}
	s := &Session{
		address:    addr,
		user:       user,
		passwd:     pw,
		serverCert: nil,
		client:     &http.Client{},
	}
	return s, nil
}

// NewSessionSSL allocates a new Session instance for communication over SSL
func NewSessionSSL(addr, user, pw string, scert *x509.Certificate) (*Session, error) {
	if _, err := url.Parse(addr); err != nil {
		return nil, err
	}
	if len(user) == 0 || len(pw) == 0 {
		return nil, errors.New("Missing credentials")
	}
	s := &Session{
		address:    addr,
		user:       user,
		passwd:     pw,
		serverCert: scert,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
					VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
						return nil
					},
				},
			},
		},
	}
	return s, nil
}

// Generic call to running server: Handles input parameters and
// returns generic result data.
func (s *Session) call(methodname string, args []Data) (result *Response, err error) {
	request := &Request{
		Version: "1.0",
		ID:      "",
		Method:  methodname,
		Params:  args,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.address, strings.NewReader(string(data)))
	req.SetBasicAuth(s.user, s.passwd)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}
	if response.Error.Code != 0 {
		return nil, errors.New(response.Error.Message)
	}
	return response, nil
}

//======================================================================
// Internal helper methods for strict checking

func getType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "map"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "float64"
	case bool:
		return "bool"
	default:
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			return "[*" + t.Elem().Name() + "]"
		}
		return "[" + t.Name() + "]"
	}
}

func compare(a, b interface{}, depth int, w io.Writer) bool {
	at := getType(a)
	bt := getType(b)
	fmt.Fprintf(w, "%d| %s\n", depth, at)
	if at != bt {
		fmt.Fprintf(w, "Type mismatch: %s != %s\n", at, bt)
		return false
	}
	switch at {
	case "array":
		aa := a.([]interface{})
		ba := b.([]interface{})
		for i, v := range aa {
			fmt.Fprintf(w, "%d| [%d]\n", depth, i)
			if !compare(v, ba[i], depth+1, w) {
				return false
			}
		}
	case "map":
		am := a.(map[string]interface{})
		bm := b.(map[string]interface{})
		for k, v := range am {
			fmt.Fprintf(w, "%d| ['%s']\n", depth, k)
			x, ok := bm[k]
			if !ok {
				fmt.Fprintf(w, "Key: %s=%v (%s)\n", k, v, getType(v))
				return false
			}
			if !compare(v, x, depth+1, w) {
				return false
			}
		}
	case "string":
		as := a.(string)
		bs := b.(string)
		fmt.Fprintf(w, "%d|   ='%s'\n", depth, as)
		return as == bs
	case "int":
		ai := a.(int)
		bi := b.(int)
		fmt.Fprintf(w, "%d|   =%d\n", depth, ai)
		return ai == bi
	case "float64":
		af := a.(float64)
		bf := b.(float64)
		fmt.Fprintf(w, "%d|   =%f\n", depth, af)
		return af == bf
	case "bool":
		ab := a.(bool)
		bb := b.(bool)
		fmt.Fprintf(w, "%d|   =%v\n", depth, ab)
		return ab == bb
	default:
		panic("compare")
	}
	return true
}

func prepare(i interface{}, w io.Writer) (interface{}, bool) {
	if getType(i)[0] != '[' {
		return i, true
	}
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Fprintln(w, "ERROR: "+err.Error())
		return nil, false
	}
	var ii interface{}
	if err = json.Unmarshal(b, &ii); err != nil {
		fmt.Fprintln(w, "ERROR: "+err.Error())
		return nil, false
	}
	return ii, true
}

func checkJSON(a, b interface{}) (bool, string) {
	buf := new(bytes.Buffer)
	am, ok := prepare(a, buf)
	if !ok {
		return false, buf.String()
	}
	bm, ok := prepare(b, buf)
	if !ok {
		return false, buf.String()
	}
	rc := compare(am, bm, 0, buf)
	return rc, buf.String()
}
