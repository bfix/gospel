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
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/bfix/gospel/crypto"
	gerr "github.com/bfix/gospel/errors"
	"github.com/bfix/gospel/logger"
)

// SendMailMessage handles outgoing message to SMTP server.
//
//   - The connections to the service can be either plain (port 25)
//     or SSL/TLS (port 465)
//
//   - If the server supports STARTTLS and the channel is not already
//     encrypted (via SSL), the application will use the "STLS" command
//     to initiate a channel encryption.
//
// - Connections can be tunneled through any SOCKS5 proxy (like Tor)
func SendMailMessage(host, proxy, fromAddr, toAddr string, body []byte) (err error) {
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

	var uSrv *url.URL
	if uSrv, err = url.Parse(host); err != nil {
		return
	}
	if proxy == "" {
		c0, err = net.Dial("tcp", uSrv.Host)
	} else {
		var (
			host, portS string
			port        int64
		)
		host, portS, err = net.SplitHostPort(uSrv.Host)
		if err != nil {
			return
		}
		port, err = strconv.ParseInt(portS, 10, 32)
		if err != nil {
			return
		}
		c0, err = Socks5Connect("tcp", host, int(port), proxy)
	}
	if err != nil {
		return
	}
	if c0 == nil {
		err = errors.New("can't estabish connection to " + uSrv.Host)
		return
	}

	sslConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // intentional
	}
	if uSrv.Scheme == "smtps" {
		c1 = tls.Client(c0, sslConfig)
		if err = c1.Handshake(); err != nil {
			return
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
		return
	}
	pw, _ := uSrv.User.Password()
	auth := smtp.PlainAuth("", uSrv.User.Username(), pw, uSrv.Host)
	if err = cli.Auth(auth); err != nil {
		return
	}
	if err = cli.Mail(fromAddr); err != nil {
		return
	}
	if err = cli.Rcpt(toAddr); err != nil {
		return
	}
	wrt, err := cli.Data()
	if err != nil {
		return
	}
	if _, err = wrt.Write(body); err != nil {
		return
	}
	if err = wrt.Close(); err != nil {
		return
	}
	err = cli.Quit()
	return
}

// MailAttachment is a data structure for data attached to a mail.
type MailAttachment struct {
	Header textproto.MIMEHeader
	Data   []byte
}

// CreateMailMessage creates a (plain) SMTP email with body and
// optional attachments.
func CreateMailMessage(body []byte, att []*MailAttachment) (msg []byte, err error) {
	buf := new(bytes.Buffer)
	wrt := multipart.NewWriter(buf)
	_, err = buf.WriteString(
		"MIME-Version: 1.0\n" +
			"Content-Type: multipart/mixed;\n" +
			" boundary=\"" + wrt.Boundary() + "\"\n\n" +
			"This is a multi-part message in MIME format.\n")
	if err != nil {
		return
	}
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "text/plain; charset=ISO-8859-15")
	hdr.Set("Content-Transfer-Encoding", "utf-8")
	var pw io.Writer
	if pw, err = wrt.CreatePart(hdr); err != nil {
		return
	}
	if _, err = pw.Write(body); err != nil {
		return
	}

	for _, a := range att {
		if pw, err = wrt.CreatePart(a.Header); err != nil {
			return
		}
		if _, err = pw.Write(a.Data); err != nil {
			return
		}
	}
	if err = wrt.Close(); err != nil {
		return
	}
	msg = buf.Bytes()
	return
}

// EncryptMailMessage encrypts a mail with given public key.
func EncryptMailMessage(key, body []byte) (cipher []byte, err error) {
	rdr := bytes.NewBuffer(key)
	var keyring openpgp.EntityList
	if keyring, err = openpgp.ReadArmoredKeyRing(rdr); err != nil {
		return
	}

	out := new(bytes.Buffer)
	var ct, wrt io.WriteCloser
	if ct, err = armor.Encode(out, "PGP MESSAGE", nil); err != nil {
		err = gerr.New(err, "no armorer created")
		return
	}
	if wrt, err = openpgp.Encrypt(ct, []*openpgp.Entity{keyring[0]}, nil, &openpgp.FileHints{IsBinary: true}, nil); err != nil {
		return
	}
	if _, err = wrt.Write(body); err != nil {
		return
	}
	if err = wrt.Close(); err != nil {
		return
	}
	if err = ct.Close(); err != nil {
		return
	}

	tmp := make([]byte, 30)
	if _, err = io.ReadFull(rand.Reader, tmp); err != nil {
		return
	}
	bndry := fmt.Sprintf("%x", tmp)
	msg := new(bytes.Buffer)
	_, err = msg.WriteString(
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
			out.String() + "\n--" + bndry + "--")
	if err != nil {
		return
	}
	cipher = msg.Bytes()
	return
}

// MailContent is the result type for parsing mail messages
type MailContent struct {
	Mode    int    // message type (MDOE_XXX)
	From    string // sender email address
	To      string // recipient email address
	Subject string // subject line
	Body    string // message body
	Key     []byte // attached key or signing key (public)
}

// MailUserInfo is a callback function to request user information:
type MailUserInfo func(key int, data string) interface{}

// Get email-related OpenPGP identity.
func getIdentity(getInfo MailUserInfo, key int, data string) *openpgp.Entity {
	var id *openpgp.Entity
	tmp := getInfo(key, data)
	switch ent := tmp.(type) {
	case *openpgp.Entity:
		id = ent
	}
	return id
}

// Parsing-related constants
const (
	ctPLAIN  = "text/plain;"
	ctMPMIX  = "multipart/mixed;"
	ctMPENC  = "multipart/encrypted;"
	ctMPSIGN = "multipart/signed;"

	modePLAIN    = iota // plain text message
	modeSIGN            // signed
	modeUSIGN           // signed, but unverified signature
	modeENC             // encrypted
	modeSIGNENC         // encrypted and signed
	modeUSIGNENC        // encrypted and signed, but unverified signature

	infoIDENTITY = iota
	infoPASSPHRASE
	infoSENDER
)

// ParseMailMessage dissects an incoming mail message
func ParseMailMessage(msg io.Reader, getInfo MailUserInfo) (mc *MailContent, err error) {
	var (
		m                *mail.Message
		fromAddr, toAddr *mail.Address
	)
	if m, err = mail.ReadMessage(msg); err != nil {
		return
	}
	if fromAddr, err = mail.ParseAddress(m.Header.Get("From")); err != nil {
		return
	}
	if toAddr, err = mail.ParseAddress(m.Header.Get("To")); err != nil {
		return
	}
	ct := m.Header.Get("Content-Type")
	if strings.HasPrefix(ct, ctPLAIN) {
		mc = new(MailContent)
		mc.Mode = modePLAIN
		mc.Key = nil
		var data []byte
		if data, err = ioutil.ReadAll(m.Body); err != nil {
			return
		}
		mc.Body = string(data)
	} else if strings.HasPrefix(ct, ctMPMIX) {
		mc, err = ParsePlain(ct, m.Body)
	} else if strings.HasPrefix(ct, ctMPENC) {
		mc, err = ParseEncrypted(ct, fromAddr.Address, getInfo, m.Body)
	} else if strings.HasPrefix(ct, ctMPSIGN) {
		mc, err = ParseSigned(ct, fromAddr.Address, getInfo, m.Body)
	}
	if err != nil {
		return
	}
	if mc == nil {
		err = errors.New("unparsed mail message")
		return
	}
	mc.From = fromAddr.Address
	mc.To = toAddr.Address
	mc.Subject = m.Header.Get("Subject")
	return
}

// ParsePlain disassembles a plain email message.
func ParsePlain(ct string, body io.Reader) (mc *MailContent, err error) {
	mc = new(MailContent)
	mc.Mode = modePLAIN
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		var part *multipart.Part
		if part, err = rdr.NextPart(); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		ct = part.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(ct, "text/plain;"):
			var data []byte
			if data, err = ioutil.ReadAll(part); err != nil {
				return
			}
			mc.Body = string(data)
		case strings.HasPrefix(ct, "application/pgp-keys;"):
			if mc.Key, err = ioutil.ReadAll(part); err != nil {
				return
			}
		default:
			err = errors.New("Unhandled MIME part: " + ct)
			return
		}
	}
}

// ParseEncrypted parses a encrypted (and possibly signed) message.
func ParseEncrypted(ct, addr string, getInfo MailUserInfo, body io.Reader) (mc *MailContent, err error) {
	mc = new(MailContent)
	mc.Mode = modeENC
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		var part *multipart.Part
		// read next mime part
		if part, err = rdr.NextPart(); err != nil {
			// no more parts: we are done
			if err == io.EOF {
				err = nil
			}
			return
		}
		// decode mime part
		ct = part.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(ct, "application/pgp-encrypted"):
			var buf []byte
			if buf, err = ioutil.ReadAll(part); err != nil {
				return
			}
			logger.Printf(logger.DBG, "application/pgp-encrypted: '%s'\n", strings.TrimSpace(string(buf)))

		case strings.HasPrefix(ct, "application/octet-stream;"):
			var rdr *armor.Block
			if rdr, err = armor.Decode(part); err != nil {
				return
			}
			pw := ""
			pwTmp := getInfo(infoPASSPHRASE, "")
			switch pws := pwTmp.(type) {
			case string:
				pw = pws
			}
			prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
				priv := keys[0].PrivateKey
				if priv.Encrypted {
					if err = priv.Decrypt([]byte(pw)); err != nil {
						return nil, err
					}
				}
				buf := new(bytes.Buffer)
				if err = priv.Serialize(buf); err != nil {
					return nil, err
				}
				return buf.Bytes(), nil
			}
			id := getIdentity(getInfo, infoIDENTITY, "")
			var md *openpgp.MessageDetails
			if md, err = openpgp.ReadMessage(rdr.Body, openpgp.EntityList{id}, prompt, nil); err != nil {
				return
			}
			if md.IsSigned {
				mc.Mode = modeSIGNENC
				id := getIdentity(getInfo, infoSENDER, addr)
				if id == nil {
					mc.Mode = modeUSIGNENC
					var content []byte
					if content, err = ioutil.ReadAll(md.UnverifiedBody); err != nil {
						return
					}
					mc.Body = string(content)
					continue
				}
				md.SignedBy = crypto.GetKeyFromIdentity(id, crypto.KeySign)
				md.SignedByKeyId = md.SignedBy.PublicKey.KeyId
				if mc.Key, err = crypto.GetArmoredPublicKey(id); err != nil {
					return
				}
				var content []byte
				if content, err = ioutil.ReadAll(md.UnverifiedBody); err != nil {
					return
				}
				if md.SignatureError != nil {
					err = md.SignatureError
					return
				}
				logger.Println(logger.INFO, "Signature verified OK")

				var m *mail.Message
				if m, err = mail.ReadMessage(bytes.NewBuffer(content)); err != nil {
					return
				}
				ct = m.Header.Get("Content-Type")
				var mc2 *MailContent
				if mc2, err = ParsePlain(ct, m.Body); err != nil {
					return
				}
				mc.Body = mc2.Body
			}
		default:
			err = errors.New("Unhandled MIME part: " + ct)
			return
		}
	}
}

// ParseSigned reads an unencrypted, but signed message.
func ParseSigned(ct, addr string, getInfo MailUserInfo, body io.Reader) (mc *MailContent, err error) {
	mc = new(MailContent)
	mc.Mode = modeSIGN
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		// get next mime part
		var part *multipart.Part
		if part, err = rdr.NextPart(); err != nil {
			// no more parts: we are done
			if err == io.EOF {
				err = nil
			}
			return
		}
		// check content type
		ct = part.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(ct, "text/plain;"):
			var data []byte
			if data, err = ioutil.ReadAll(part); err != nil {
				return
			}
			mc.Body = string(data)
		case strings.HasPrefix(ct, "application/pgp-signature;"):
			id := getIdentity(getInfo, infoSENDER, addr)
			if id == nil {
				mc.Mode = modeUSIGN
				continue
			}
			buf := bytes.NewBufferString(mc.Body)
			if _, err = openpgp.CheckArmoredDetachedSignature(openpgp.EntityList{id}, buf, part, nil); err != nil {
				return
			}
			logger.Println(logger.INFO, "Signature verified OK")
		}
	}
}

// Extract value from string ('... key="value" ...')
func extractValue(s, key string) string {
	idx := strings.Index(s, key)
	skip := idx + len(key) + 2
	if idx < 0 || len(s) < skip {
		return ""
	}
	s = s[skip:]
	idx = strings.IndexRune(s, '"')
	if idx < 0 {
		return ""
	}
	return s[:idx]
}
