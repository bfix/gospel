/*
 * Network-related utility functions
 * =================================
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
	"errors"
	"strconv"
	"strings"
)

///////////////////////////////////////////////////////////////////////
// Public methods

/*
 * Split "host:port" string into components.
 * @param host string - incoming service spec
 * @return addr string - service address
 * @return port int - port address (1-65535)
 * @return error - error instance or nil
 */
func SplitHost(host string) (addr string, port int, err error) {
	idx := strings.Index(host, ":")
	if idx == -1 {
		err = errors.New("Invalid host definition")
		return
	}
	addr = host[:idx]
	port, err = strconv.Atoi(host[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		err = errors.New("Invalid host definition")
	}
	return
}
