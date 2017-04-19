package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
// a JSON data structur.
func (r *Response) UnmarshalResult(v interface{}) error {
	data, err := json.Marshal(r.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Error is a Response-related failure code.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Session type
type Session struct {
	Address    string            // server address/name
	User       string            // user name
	Passwd     string            // user password
	ServerCert *x509.Certificate // server certificate for SSL

	client *http.Client
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
		Address:    addr,
		User:       user,
		Passwd:     pw,
		ServerCert: nil,
	}
	s.client = &http.Client{}
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
		Address:    addr,
		User:       user,
		Passwd:     pw,
		ServerCert: scert,
	}
	s.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					return nil
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
	req, err := http.NewRequest("POST", s.Address, strings.NewReader(string(data)))
	req.SetBasicAuth(s.User, s.Passwd)
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

// GetInfo returns an object containing various state info.
func (s *Session) GetInfo() (*Info, error) {
	res, err := s.call("getinfo", nil)
	if err != nil {
		return nil, err
	}
	info := new(Info)
	if err = res.UnmarshalResult(info); err != nil {
		return nil, err
	}
	return info, err
}

// GetDifficulty returns the proof-of-work difficulty as a multiple
// of the minimum difficulty.
func (s *Session) GetDifficulty() (float64, error) {
	res, err := s.call("getdifficulty", nil)
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// KeypoolRefill creates a number of new Bitcoin addresses for later use.
// Remarks: Requires unlocked wallet
func (s *Session) KeypoolRefill() error {
	_, err := s.call("getnewaddress", nil)
	if err != nil {
		return err
	}
	return nil
}

// EstimateFee estimates the transaction fee per kilobyte that needs to be
// paid for a transaction to be included within a certain number of blocks.
func (s *Session) EstimateFee(waitBlocks int) (float64, error) {
	res, err := s.call("estimatefee", []Data{waitBlocks})
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}

// EstimatePriority estimates the priority that a transaction needs in order
// to be included within a certain number of blocks as a free high-priority
// transaction.
func (s *Session) EstimatePriority(waitBlocks int) (float64, error) {
	res, err := s.call("estimatepriority", []Data{waitBlocks})
	if err != nil {
		return -1, err
	}
	return res.Result.(float64), nil
}
