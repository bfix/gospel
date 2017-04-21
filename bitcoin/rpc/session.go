package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
func (r *Response) UnmarshalResult(v interface{}) error {
	data, err := json.Marshal(r.Result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, v); err != nil {
		return err
	}
	if strictCheck {
		rc, msg := checkJSON(r.Result, v)
		if !rc {
			return errors.New(">>>>>\n" + msg)
		}
		rc, msg = checkJSON(v, r.Result)
		if !rc {
			fmt.Println("Result: " + string(data))
			return errors.New("<<<<<\n" + msg)
		}
	}
	return nil
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
