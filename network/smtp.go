/*
 * Send mail messages through SMTP server
 * ======================================
 *
 * - The connections to the service can be either plain (port 25)
 *   or SSL/TLS (port 465)
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
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"code.google.com/p/go.crypto/openpgp/armor"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/bfix/gospel/logger"
	"io"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

///////////////////////////////////////////////////////////////////////
/*
 * Handle outgoing message to SMTP server.
 * @param host string - SMTP URL of service
 * @param proxy string - proxy service
 * @param fromAddr string - sender address
 * @param toAddr string - receiver address
 * @param body []byte - mail body
 * @return error - error instance or nil
 */
func SendMailMessage(host, proxy, fromAddr, toAddr string, body []byte) error {
	var (
		c0  net.Conn
		c1  *tls.Conn
		cli *smtp.Client
	)
	defer func() {
		if cli != nil {
			cli.Close()
		}
		if c1 != nil {
			c1.Close()
		}
		if c0 != nil {
			c0.Close()
		}
	}()

	uSrv, err := url.Parse(host)
	if err != nil {
		return err
	}
	if proxy == "" {
		c0, err = net.Dial("tcp", uSrv.Host)
	} else {
		uPrx, err := url.Parse(proxy)
		if err != nil {
			return err
		}
		idx := strings.Index(uSrv.Host, ":")
		if idx == -1 {
			return errors.New("Invalid host definition")
		}
		host := uSrv.Host[:idx]
		port, err := strconv.Atoi(uSrv.Host[idx+1:])
		if err != nil || port < 1 || port > 65535 {
			return errors.New("Invalid host definition")
		}
		c0, err = Socks5Connect("tcp", host, port, uPrx.Host)
	}
	if err != nil {
		return err
	}
	if c0 == nil {
		return errors.New("Can't estabish connection to " + uSrv.Host)
	}

	sslConfig := &tls.Config{InsecureSkipVerify: true}
	if uSrv.Scheme == "smtps" {
		c1 = tls.Client(c0, sslConfig)
		if err = c1.Handshake(); err != nil {
			return err
		}
		cli, err = smtp.NewClient(c1, uSrv.Host)
	} else {
		cli, err = smtp.NewClient(c0, uSrv.Host)
		if err == nil {
			if ok, _ := cli.Extension("STLS"); ok {
				err = cli.StartTLS(sslConfig)
			}
		}
	}
	if err != nil {
		return err
	}
	pw, _ := uSrv.User.Password()
	auth := smtp.PlainAuth("", uSrv.User.Username(), pw, uSrv.Host)
	if err = cli.Auth(auth); err != nil {
		return err
	}
	if err = cli.Mail(fromAddr); err != nil {
		return err
	}
	if err = cli.Rcpt(toAddr); err != nil {
		return err
	}
	wrt, err := cli.Data()
	if err != nil {
		return err
	}
	wrt.Write(body)
	wrt.Close()
	if err = cli.Quit(); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////
// Message creation functions

type MailAttachment struct {
	Header textproto.MIMEHeader
	Data   []byte
}

//---------------------------------------------------------------------
/*
 * Create (plain) SMTP email with body and optional attachments.
 * @param body []byte - mail body (UTF-8 encoded)
 * @param att []MailAttachment - list of attachments
 * @return []byte - data to be send to SMTP server (or encryption)
 * @return error - error instance or nil
 */
func CreateMailMessage(body []byte, att []*MailAttachment) ([]byte, error) {
	buf := new(bytes.Buffer)
	wrt := multipart.NewWriter(buf)
	buf.WriteString(
		"MIME-Version: 1.0\n" +
			"Content-Type: multipart/mixed;\n" +
			" boundary=\"" + wrt.Boundary() + "\"\n\n" +
			"This is a multi-part message in MIME format.\n")
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "text/plain; charset=ISO-8859-15")
	hdr.Set("Content-Transfer-Encoding", "utf-8")
	pw, err := wrt.CreatePart(hdr)
	if err != nil {
		return nil, err
	}
	pw.Write(body)

	for _, a := range att {
		pw, err = wrt.CreatePart(a.Header)
		if err != nil {
			return nil, err
		}
		pw.Write(a.Data)
	}
	wrt.Close()
	return buf.Bytes(), nil
}

//---------------------------------------------------------------------
/*
 * Encrypt mail with given public key.
 * @param key []byte - public OpenPGP key used for encryption
 * @param body []byte - mail body
 * @return []byte - data to be send to SMTP server
 * @return error - error instance or nil
 */
func EncryptMailMessage(key, body []byte) ([]byte, error) {
	rdr := bytes.NewBuffer(key)
	keyring, err := openpgp.ReadArmoredKeyRing(rdr)
	if err != nil {
		logger.Println(logger.ERROR, err.Error())
		return nil, err
	}

	out := new(bytes.Buffer)
	ct, err := armor.Encode(out, "PGP MESSAGE", nil)
	if err != nil {
		logger.Println(logger.ERROR, "Can't create armorer: "+err.Error())
		return nil, err
	}
	wrt, err := openpgp.Encrypt(ct, []*openpgp.Entity{keyring[0]}, nil, &openpgp.FileHints{IsBinary: true}, nil)
	if err != nil {
		logger.Println(logger.ERROR, err.Error())
		return nil, err
	}
	wrt.Write(body)
	wrt.Close()
	ct.Close()

	tmp := make([]byte, 30)
	_, err = io.ReadFull(rand.Reader, tmp)
	if err != nil {
		logger.Println(logger.ERROR, err.Error())
		return nil, err
	}
	bndry := fmt.Sprintf("%x", tmp)
	msg := new(bytes.Buffer)
	msg.WriteString(
		"MIME-Version: 1.0\n" +
			"Content-Type: multipart/encrypted;\n" +
			" protocol=\"application/pgp-encrypted\";\n" +
			" boundary=\"" + bndry + "\"\n\n" +
			"This is an OpenPGP/MIME encrypted message (RFC 4880 and 3156)\n" +
			"--" + bndry + "\n" +
			"Content-Type: application/pgp-encrypted\n" +
			"Content-Description: PGP/MIME version identification\n\n" +
			"Version: 1\n\n" +
			"--" + bndry + "\n" +
			"Content-Type: application/octet-stream;\n name=\"encrypted.asc\"\n" +
			"Content-Description: OpenPGP encrypted message\n" +
			"Content-Disposition: inline;\n filename=\"encrypted.asc\"\n\n" +
			string(out.Bytes()) + "\n--" + bndry + "--")
	return msg.Bytes(), nil
}
