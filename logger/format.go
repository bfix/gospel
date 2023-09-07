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

package logger

import (
	"fmt"
	"strings"
	"time"
)

// Formatter function type: convert a log message into a string for output
type Formatter func(msg *logMsg) string

// SimpleFormat creates a plain format for log messages
func SimpleFormat(msg *logMsg) string {
	ts := msg.ts.Format(time.Stamp)
	lvl := getTag(msg.level)
	txt := msg.text

	txt = strings.Trim(txt, "\n")
	return fmt.Sprintf("%s [%s] %s\n", ts, lvl, txt)
}

// ColorFormat uses colors for different log levels
func ColorFormat(msg *logMsg) string {
	col := 34 // light blue for undef`d levels
	switch msg.level {
	case CRITICAL:
		col = 31
	case ERROR:
		col = 31
	case WARN:
		col = 33
	case INFO:
		col = 37
	case DBG:
		col = 90
	}
	txt := SimpleFormat(msg)
	txt = strings.Trim(txt, "\n")
	return fmt.Sprintf("\033[01;%dm%s\033[01;0m\n", col, txt)
}
