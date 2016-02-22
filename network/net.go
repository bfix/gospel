/*
 * Network helper functions: Send and receive data over socket buffer.
 *
 * (c) 2012 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package network

///////////////////////////////////////////////////////////////////////
// Import external declarations.

import (
	"github.com/bfix/gospel/logger"
	"net"
	"time"
)

///////////////////////////////////////////////////////////////////////
// Global attributes

var delay, _ = time.ParseDuration("1ms")
var retries = 1000
var timeout, _ = time.ParseDuration("100us")

///////////////////////////////////////////////////////////////////////
// Public functions

// SendData sends data over network connection (stream-oriented).
func SendData(conn net.Conn, data []byte, srv string) bool {

	count := len(data) // total length of data
	start := 0         // start position of slice
	retry := 0         // retry counter

	// write data to socket buffer
	for count > 0 {
		// set timeout
		conn.SetDeadline(time.Now().Add(timeout))
		// get (next) chunk to be send
		chunk := data[start : start+count]
		if num, err := conn.Write(chunk); num > 0 {
			// advance slice on partial write
			start += num
			count -= num
			retry = 0
		} else if err != nil {
			// handle error condition
			switch err.(type) {
			case net.Error:
				// network error: retry...
				nerr := err.(net.Error)
				if nerr.Timeout() || nerr.Temporary() {
					retry++
					time.Sleep(delay)
					if retry == retries {
						logger.Printf(logger.ERROR, "[%s] Write failed after retries: %s\n", srv, err.Error())
						return false
					}
				}
			default:
				logger.Printf(logger.INFO, "[%s] Connection closed by peer\n", srv)
				return false
			}
		}
	}
	// report success
	if retry > 0 {
		logger.Printf(logger.INFO, "[%s] %d retries needed to send data.\n", srv, retry)
	}
	return true
}

//---------------------------------------------------------------------

// RecvData receives data over network connection (stream-oriented).
func RecvData(conn net.Conn, data []byte, srv string) (int, bool) {

	for retry := 0; retry < retries; {
		// set timeout
		conn.SetDeadline(time.Now().Add(timeout))
		// read data from socket buffer
		n, err := conn.Read(data)
		if err != nil {
			// handle error condition
			switch err.(type) {
			case net.Error:
				// network error: retry...
				nerr := err.(net.Error)
				if nerr.Timeout() {
					return 0, true
				} else if nerr.Temporary() {
					retry++
					time.Sleep(delay)
					continue
				}
			default:
				logger.Printf(logger.INFO, "[%s] Connection closed by peer\n", srv)
				return 0, false
			}
		}
		// report success
		if retry > 0 {
			logger.Printf(logger.INFO, "[%s] %d retries needed to receive data.\n", srv, retry)
		}
		return n, true
	}
	// retries failed
	logger.Printf(logger.ERROR, "[%s] Read failed after retries...\n", srv)
	return 0, false
}
