package network

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2023 Bernd Fix  >Y<
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

import (
	"crypto/tls"
	"errors"
	"net"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// POP3Session data structure
type POP3Session struct {
	conn        *textproto.Conn
	c0          net.Conn
	c1          *tls.Conn
	established bool
}

// POP3Connect establishes a session with POP3 mailbox.
//
//   - The connections to the service can be either plain (port 110)
//     or SSL/TLS (port 995)
//
//   - If the server supports STARTTLS and the channel is not already
//     encrypted (via SSL), the application will use the "STLS" command
//     to initiate a channel encryption.
//
// - Connections can be tunneled through any SOCKS5 proxy (like Tor)
func POP3Connect(service, proxy string) (sess *POP3Session, err error) {
	sess = new(POP3Session)
	sess.conn = nil
	sess.c0 = nil
	sess.c1 = nil
	sess.established = false

	defer func() {
		if !sess.established {
			if sess.conn != nil {
				_ = sess.conn.Close()
			}
			if sess.c1 != nil {
				_ = sess.c1.Close()
			}
			if sess.c0 != nil {
				_ = sess.c0.Close()
			}
		}
	}()

	var res []string
	var uSrv *url.URL
	if uSrv, err = url.Parse(service); err != nil {
		return
	}
	if proxy == "" {
		sess.c0, err = net.Dial("tcp", uSrv.Host)
	} else {
		var (
			host, portS string
			port        int64
		)
		host, portS, err = net.SplitHostPort(uSrv.Host)
		if err != nil {
			return nil, err
		}
		port, err = strconv.ParseInt(portS, 10, 32)
		if err != nil {
			return nil, err
		}
		sess.c0, err = Socks5Connect("tcp", host, int(port), proxy)
	}
	if err != nil {
		return
	}
	if sess.c0 == nil {
		err = errors.New("Can't estabish connection to " + uSrv.Host)
		return
	}
	sslConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // intentional
	}
	if uSrv.Scheme == "pops" {
		sess.c1 = tls.Client(sess.c0, sslConfig)
		if err = sess.c1.Handshake(); err != nil {
			return
		}
	}
	if sess.c1 != nil {
		sess.conn = textproto.NewConn(sess.c1)
	} else {
		sess.conn = textproto.NewConn(sess.c0)
	}
	if _, err = sess.Exec("", false); err != nil {
		return
	}
	if sess.c1 == nil {
		var capabilities []string
		if capabilities, err = sess.Exec("CAPA", true); err != nil {
			return
		}
		if success, msg := checkResponse(capabilities[0]); !success {
			err = errors.New(msg)
			return
		}
		for _, s := range capabilities[1:] {
			if s == "STLS" {
				if res, err = sess.Exec("STLS", false); err != nil {
					return
				}
				if success, _ := checkResponse(res[0]); success {
					sess.c1 = tls.Client(sess.c0, sslConfig)
					if err = sess.c1.Handshake(); err != nil {
						return
					}
					sess.conn = textproto.NewConn(sess.c1)
				}
			}
		}
	}
	if res, err = sess.Exec("USER "+uSrv.User.Username(), false); err != nil {
		return
	}
	if success, msg := checkResponse(res[0]); !success {
		return nil, errors.New(msg)
	}
	pw, ok := uSrv.User.Password()
	if !ok {
		return nil, errors.New("missing password")
	}
	if res, err = sess.Exec("PASS "+pw, false); err != nil {
		return
	}
	if success, msg := checkResponse(res[0]); !success {
		return nil, errors.New(msg)
	}
	sess.established = true
	return
}

// Close a POP3 session with the server.
func (sess *POP3Session) Close() (err error) {
	if _, err = sess.Exec("QUIT", false); err != nil {
		return
	}
	if err = sess.conn.Close(); err != nil {
		return
	}
	if err = sess.c0.Close(); err != nil {
		return
	}
	if sess.c1 != nil {
		err = sess.c1.Close()
	}
	return
}

// Exec executes a command on the POP3 server:
// Expected data is assumed to be terminated by a line
// containing a single dot.
func (sess *POP3Session) Exec(cmd string, expectData bool) ([]string, error) {
	if len(cmd) > 0 {
		if err := sess.conn.PrintfLine(cmd); err != nil {
			return nil, err
		}
	}
	if expectData {
		var res []string
		for {
			s, err := sess.conn.ReadLine()
			if err != nil {
				return nil, err
			}
			if s == "." {
				break
			}
			res = append(res, s)
		}
		if len(res) == 0 {
			return nil, errors.New("no response data")
		}
		return res, nil
	}
	res, err := sess.conn.ReadLine()
	if err != nil {
		return nil, err
	}
	return []string{res}, nil
}

// ListUnread returns a list of unread messages.
func (sess *POP3Session) ListUnread() (list []int, err error) {
	var res []string
	if res, err = sess.Exec("LIST", true); err != nil {
		return
	}
	if success, msg := checkResponse(res[0]); !success {
		err = errors.New(msg)
		return
	}
	for _, s := range res[1:] {
		idStr := strings.Split(s, " ")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		list = append(list, id)
	}
	return
}

// Retrieve message# <id> from the server.
func (sess *POP3Session) Retrieve(id int) ([]string, error) {
	if err := sess.c0.SetDeadline(time.Now().Add(30 * time.Minute)); err != nil {
		return nil, err
	}
	res, err := sess.Exec("RETR "+strconv.Itoa(id), true)
	if err != nil {
		return nil, err
	}
	success, msg := checkResponse(res[0])
	if !success {
		return nil, errors.New(msg)
	}
	return res[1:], nil
}

// Delete message# <id> from the server.
func (sess *POP3Session) Delete(id int) error {
	res, err := sess.Exec("DELE "+strconv.Itoa(id), false)
	if err != nil {
		return err
	}
	success, msg := checkResponse(res[0])
	if !success {
		return errors.New(msg)
	}
	return nil
}

// Check the response from the server.
func checkResponse(res string) (success bool, msg string) {
	success = false
	if strings.HasPrefix(res, "+OK") {
		success = true
	}
	msg = ""
	if pos := strings.IndexRune(res, ' '); pos != -1 {
		msg = res[pos+1:]
	}
	return success, msg
}
