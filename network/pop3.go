/*
 * Handle POP3 server to receive messages
 * ======================================
 *
 * - The connections to the service can be either plain (port 110)
 *   or SSL/TLS (port 995)
 * - If the server supports STARTTLS and the channel is not already
 *   encrypted (via SSL), the application will use the "STLS" command
 *   to initiate a channel encryption.
 * - Connections can be tunneled through any SOCKS5 proxy (like Tor)
 *
 * (c) 2013-2014 Bernd Fix    >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package network

///////////////////////////////////////////////////////////////////////
// Import external declarations.

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

///////////////////////////////////////////////////////////////////////
// POP3-related types and methods

// POP3Session data structure
type POP3Session struct {
	conn        *textproto.Conn
	c0          net.Conn
	c1          *tls.Conn
	established bool
}

//---------------------------------------------------------------------

// POP3Connect establishes a session with POP3 mailbox.
func POP3Connect(service, proxy string) (*POP3Session, error) {
	sess := new(POP3Session)
	sess.conn = nil
	sess.c0 = nil
	sess.c1 = nil
	sess.established = false

	defer func() {
		if !sess.established {
			if sess.conn != nil {
				sess.conn.Close()
			}
			if sess.c1 != nil {
				sess.c1.Close()
			}
			if sess.c0 != nil {
				sess.c0.Close()
			}
		}
	}()

	uSrv, err := url.Parse(service)
	if err != nil {
		return nil, err
	}
	if proxy == "" {
		sess.c0, err = net.Dial("tcp", uSrv.Host)
	} else {
		host, port, err := SplitHost(uSrv.Host)
		if err != nil {
			return nil, err
		}
		sess.c0, err = Socks5Connect("tcp", host, port, proxy)
	}
	if err != nil {
		return nil, err
	}
	if sess.c0 == nil {
		return nil, errors.New("Can't estabish connection to " + uSrv.Host)
	}

	sslConfig := &tls.Config{InsecureSkipVerify: true}
	if uSrv.Scheme == "pops" {
		sess.c1 = tls.Client(sess.c0, sslConfig)
		if err = sess.c1.Handshake(); err != nil {
			return nil, err
		}
	}

	if sess.c1 != nil {
		sess.conn = textproto.NewConn(sess.c1)
	} else {
		sess.conn = textproto.NewConn(sess.c0)
	}

	sess.Exec("", false)

	if sess.c1 == nil {
		capabilities, err := sess.Exec("CAPA", true)
		if err != nil {
			return nil, err
		}
		if success, msg := checkResponse(capabilities[0]); !success {
			return nil, errors.New(msg)
		}
		for _, s := range capabilities[1:] {
			if s == "STLS" {
				res, err := sess.Exec("STLS", false)
				if err != nil {
					return nil, err
				}
				if success, _ := checkResponse(res[0]); success {
					sess.c1 = tls.Client(sess.c0, sslConfig)
					if err = sess.c1.Handshake(); err != nil {
						return nil, err
					}
					sess.conn = textproto.NewConn(sess.c1)
				}
			}
		}
	}

	res, err := sess.Exec("USER "+uSrv.User.Username(), false)
	if err != nil {
		return nil, err
	}
	if success, msg := checkResponse(res[0]); !success {
		return nil, errors.New(msg)
	}
	pw, ok := uSrv.User.Password()
	if !ok {
		return nil, errors.New("Missing password")
	}
	if res, err = sess.Exec("PASS "+pw, false); err != nil {
		return nil, err
	}
	if success, msg := checkResponse(res[0]); !success {
		return nil, errors.New(msg)
	}
	sess.established = true
	return sess, nil
}

//---------------------------------------------------------------------

// Close a POP3 session with the server.
func (sess *POP3Session) Close() {
	sess.Exec("QUIT", false)
	sess.conn.Close()
	sess.c0.Close()
	if sess.c1 != nil {
		sess.c1.Close()
	}
}

//---------------------------------------------------------------------

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
			return nil, errors.New("No response data")
		}
		return res, nil
	}
	res, err := sess.conn.ReadLine()
	if err != nil {
		return nil, err
	}
	return []string{res}, nil
}

//---------------------------------------------------------------------

// ListUnread returns a list of unread messages.
func (sess *POP3Session) ListUnread() ([]int, error) {
	res, err := sess.Exec("LIST", true)
	if err != nil {
		return nil, err
	}
	if success, msg := checkResponse(res[0]); !success {
		return nil, errors.New(msg)
	}

	var idList []int
	for _, s := range res[1:] {
		idStr := strings.Split(s, " ")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		idList = append(idList, id)
	}
	return idList, nil
}

//---------------------------------------------------------------------

// Retrieve message# <id> from the server.
func (sess *POP3Session) Retrieve(id int) ([]string, error) {
	sess.c0.SetDeadline(time.Now().Add(30 * time.Minute))
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

//---------------------------------------------------------------------

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

///////////////////////////////////////////////////////////////////////
// Helper methods.

/*
 * Check the response from the server.
 * @param res string - response code from server
 * @return success bool - successful operation?
 * @return msg string - response message
 */
func checkResponse(res string) (success bool, msg string) {
	success = false
	if strings.HasPrefix(res, "+OK") {
		success = true
	}
	msg = ""
	if pos := strings.IndexRune(res, ' '); pos != -1 {
		msg = string(res[pos+1:])
	}
	return success, msg
}
