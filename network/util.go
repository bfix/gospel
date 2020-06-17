package network

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
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
	"errors"
	"strconv"
	"strings"
)

// SplitHost dissects a string of fotm "host:port" string into components.
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
