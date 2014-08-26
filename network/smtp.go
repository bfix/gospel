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
	"github.com/bfix/gospel/crypto"
	"github.com/bfix/gospel/logger"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
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
		host, port, err := SplitHost(uSrv.Host)
		if err != nil {
			return err
		}
		c0, err = Socks5Connect("tcp", host, port, proxy)
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

//---------------------------------------------------------------------
/*
 * Result type for parsing mail messages
 */
type MailContent struct {
	Mode int    // message type (MDOE_XXX)
	From string // sender email address
	Body string // message body
	Key  []byte // attached key or signing key (public)
}

/*
 * Callback function to request user information:
 * @param key int - info type requested (INFO_XXX)
 * @param data string - additional data
 * @return interface{} - return object
 */
type MailUserInfo func(key int, data string) interface{}

//---------------------------------------------------------------------
/*
 * Parsing-related constants
 */
const (
	ct_PLAIN   = "text/plain;"
	ct_MP_MIX  = "multipart/mixed;"
	ct_MP_ENC  = "multipart/encrypted;"
	ct_MP_SIGN = "multipart/signed;"

	MODE_PLAIN = iota
	MODE_SIGN
	MODE_ENC
	MODE_SIGN_ENC
	MODE_USIGN_ENC // unverified signature (key missing)

	INFO_IDENTITY = iota
	INFO_PASSPHRASE
	INFO_SENDER
)

//---------------------------------------------------------------------
/*
 * Parse mail message:
 * @param msg io.Reader - message reader
 * @param getInfo MailUserInfo - callback for info retrieval
 * @return error - error instance or nil
 */
func ParseMailMessage(msg io.Reader, getInfo MailUserInfo) (*MailContent, error) {
	m, err := mail.ReadMessage(msg)
	if err != nil {
		return nil, err
	}
	addr, err := mail.ParseAddress(m.Header.Get("From"))
	if err != nil {
		return nil, err
	}
	ct := m.Header.Get("Content-Type")
	var mc *MailContent = nil
	if strings.HasPrefix(ct, ct_PLAIN) {
		mc = new(MailContent)
		mc.Mode = MODE_PLAIN
		mc.Key = nil
		data, err := ioutil.ReadAll(m.Body)
		if err != nil {
			return nil, err
		}
		mc.Body = string(data)
	} else if strings.HasPrefix(ct, ct_MP_MIX) {
		mc, err = ParsePlain(ct, m.Body)
	} else if strings.HasPrefix(ct, ct_MP_ENC) {
		mc, err = ParseEncrypted(ct, addr.Address, getInfo, m.Body)
	} else if strings.HasPrefix(ct, ct_MP_SIGN) {
		mc, err = ParseSigned(ct, addr.Address, m.Body)
	}
	if err != nil {
		return nil, err
	}
	if mc == nil {
		return nil, errors.New("Unparsed mail message")
	}
	mc.From = addr.Address
	return mc, nil
}

//---------------------------------------------------------------------
/*
 * Parse plain text message.
 * @param ct string - content type string
 * @param body io.Reader - content reader
 * @return *MailContent - parse result
 * @return error - error instance or nil
 */
func ParsePlain(ct string, body io.Reader) (*MailContent, error) {
	mc := new(MailContent)
	mc.Mode = MODE_PLAIN
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		if part, err := rdr.NextPart(); err == nil {
			ct = part.Header.Get("Content-Type")
			switch {
			case strings.HasPrefix(ct, "text/plain;"):
				data, err := ioutil.ReadAll(part)
				if err != nil {
					return nil, err
				}
				mc.Body = string(data)
			case strings.HasPrefix(ct, "application/pgp-keys;"):
				mc.Key, err = ioutil.ReadAll(part)
				if err != nil {
					return nil, err
				}
			default:
				return nil, errors.New("Unhandled MIME part: " + ct)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}
	return mc, nil
}

//---------------------------------------------------------------------
/*
 * Parse encrypted (and possibly signed) message.
 * @param ct string - content type string
 * @param addr string - sender address
 * @param getInfo MailUserInfo - callback for info retrieval
 * @param body io.Reader - content reader
 * @return *MailContent - parse result
 * @return error - error instance or nil
 */
func ParseEncrypted(ct, addr string, getInfo MailUserInfo, body io.Reader) (*MailContent, error) {
	mc := new(MailContent)
	mc.Mode = MODE_ENC
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		if part, err := rdr.NextPart(); err == nil {
			ct = part.Header.Get("Content-Type")
			switch {
			case strings.HasPrefix(ct, "application/pgp-encrypted"):
				buf, err := ioutil.ReadAll(part)
				if err != nil {
					return nil, err
				}
				logger.Printf(logger.DBG, "application/pgp-encrypted: '%s'\n", strings.TrimSpace(string(buf)))
				continue
			case strings.HasPrefix(ct, "application/octet-stream;"):
				rdr, err := armor.Decode(part)
				if err != nil {
					return nil, err
				}
				pw := getInfo(INFO_PASSPHRASE, "").(string)
				prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
					priv := keys[0].PrivateKey
					if priv.Encrypted {
						priv.Decrypt([]byte(pw))
					}
					buf := new(bytes.Buffer)
					priv.Serialize(buf)
					return buf.Bytes(), nil
				}
				identity := getInfo(INFO_IDENTITY, "").(*openpgp.Entity)
				md, err := openpgp.ReadMessage(rdr.Body, openpgp.EntityList{identity}, prompt, nil)
				if err != nil {
					return nil, err
				}
				if md.IsSigned {
					mc.Mode = MODE_SIGN_ENC
					id := getInfo(INFO_SENDER, addr).(*openpgp.Entity)
					if id == nil {
						mc.Mode = MODE_USIGN_ENC
						content, err := ioutil.ReadAll(md.UnverifiedBody)
						if err != nil {
							return nil, err
						}
						mc.Body = string(content)
						continue
					}
					md.SignedBy = crypto.GetKeyFromIdentity(id, crypto.KEY_SIGN)
					md.SignedByKeyId = md.SignedBy.PublicKey.KeyId
					mc.Key, err = crypto.GetArmoredPublicKey(id)
					if err != nil {
						return nil, err
					}
					content, err := ioutil.ReadAll(md.UnverifiedBody)
					if err != nil {
						return nil, err
					}
					if md.SignatureError != nil {
						return nil, md.SignatureError
					}
					logger.Println(logger.INFO, "Signature verified OK")

					m, err := mail.ReadMessage(bytes.NewBuffer(content))
					if err != nil {
						return nil, err
					}
					ct = m.Header.Get("Content-Type")
					mc2, err := ParsePlain(ct, m.Body)
					if err != nil {
						return nil, err
					}
					mc.Body = mc2.Body
				}
			default:
				return nil, errors.New("Unhandled MIME part: " + ct)
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}
	return mc, nil
}

//---------------------------------------------------------------------
/*
 * Parse signed unencrypted message.
 * @param ct string - content type string
 * @param addr string - sender address
 * @param body io.Reader - content reader
 * @return *MailContent - parse result
 * @return error - error instance or nil
 */
func ParseSigned(ct, addr string, body io.Reader) (*MailContent, error) {
	mc := new(MailContent)
	mc.Mode = MODE_ENC
	boundary := extractValue(ct, "boundary")
	rdr := multipart.NewReader(body, boundary)
	for {
		if part, err := rdr.NextPart(); err == nil {
			ct = part.Header.Get("Content-Type")
			switch {
			case strings.HasPrefix(ct, "text/plain;"):
				data, err := ioutil.ReadAll(part)
				if err != nil {
					return nil, err
				}
				mc.Body = string(data)
			case strings.HasPrefix(ct, "application/pgp-signature;"):
				data, err := ioutil.ReadAll(part)
				if err != nil {
					return nil, err
				}
				fmt.Println("Signature: " + string(data))
			}
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}
	return mc, nil
}

//---------------------------------------------------------------------
/*
 * Extract value from string ('... key="value" ...')
 * @param s string - input string
 * @param key string - name of key
 * @param string - value string (or empty)
 */
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
