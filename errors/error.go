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

package errors

import "fmt"

// Error is a wrapper for errors produced by (parts of) the Gospel
// implementation where variable error context is required for
// defined errors
type Error struct {
	Err error  // base error (for errors.Is() and errors.As() calls)
	Ctx string // error context
}

// Unwrap error to standard type
func (e *Error) Unwrap() error {
	return e.Err
}

// Error returns a human-readble error description
func (e *Error) Error() string {
	return e.Err.Error() + " [" + e.Ctx + "]"
}

// New creates a new Error instance
func New(err error, format string, args ...interface{}) *Error {
	return &Error{
		Err: err,
		Ctx: fmt.Sprintf(format, args...),
	}
}
