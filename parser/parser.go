/*
 * Parser: Read/Access/Write nested data structures.
 *
 * (c) 2010 Bernd Fix   >Y<
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

package parser

///////////////////////////////////////////////////////////////////////
// Import external declarations.

import (
	"bufio"
	"errors"
	"github.com/bfix/gospel/data"
	"io"
	"strconv"
	"unicode"
)

///////////////////////////////////////////////////////////////////////
// Define public constants (pair types).

const (
	DONE  = iota // not defined
	ERROR        // signal error to callback
	EMPTY        // empty parameter
	VAR          // generic parameter
	VALUE        // parameter value
	LIST         // parameter list
)

///////////////////////////////////////////////////////////////////////
// Public types 

/*
 * Parameter type declaration.
 */
type Parameter struct {
	Name  string // name of parameter
	Value string // value of parameter (encoded as string)
}

//---------------------------------------------------------------------
/*
 * Reset parameter instance.
 * @this p *Parameter
 */
func (p *Parameter) reset() {
	p.Name = ""
	p.Value = ""
}

//---------------------------------------------------------------------
/*
 * Printable parameter instance.
 * @this p *Parameter
 * @return string - printable parameter
 */
func (p *Parameter) String() string {
	val := p.Value
	if val != "{}" && val != "~" {
		val = "\"" + val + "\""
	}
	if len(p.Name) > 0 {
		return "`" + p.Name + "=" + val + "`"
	}
	return "`" + val + "`"
}

//---------------------------------------------------------------------
/*
 * Callback prototype
 */
type Callback func(mode int, param *Parameter) bool

///////////////////////////////////////////////////////////////////////
// Public methods

/*
 * Parse data definition from reader and pass parameters to callback.
 * @param rdr *bufio.Reader - stream reader
 * @param cb Callback - callback function
 * @return error - error encountered (or nil if successful)
 */
func Parser(rdr *bufio.Reader, cb Callback) error {

	state := 1                  // current state in state machine
	skip := true                // skip white spaces (outside string)?
	escaped := false            // last character was escape?
	comment := false            // we are inside commnent
	buf := ""                   // buffer for string assembly
	param := new(Parameter)     // parameter instance
	stack := data.NewIntStack() // stack for nested values
	line, offset := 1, 0        // line/offset in current stream

	// execute state machine
	param.reset()
	for state != 0 {
		//=============================================================
		// read next rune; skip whitespaces and quit if error occurs
		//=============================================================
		r, _, err := rdr.ReadRune()
		offset++
		if err != nil {
			if err == io.EOF {
				// check for pending element
				if stack.Len() > 0 {
					if stack.Peek() == VALUE {
						// yes: named parameter
						param.Value = buf
						cb(VAR, param)
					} else if stack.Peek() == VAR {
						// yes: unnamed parameter
						param.Name = ""
						param.Value = buf
						cb(VAR, param)
					} else if stack.Len() > 1 {
						// signal parser error
						cb(ERROR, nil)
						return mkError("Pre-mature end of data", line, offset)
					}
				}
				// notify end of processing
				cb(DONE, nil)
				return nil
			}
			return err
		}
		// detect start of comment.
		if comment {
			if r == '\n' {
				comment = false
			}
			continue
		} else if !escaped && r == '#' {
			comment = true
			continue
		}
		// handle line breaks
		if r == '\n' {
			line++
			offset = 0
			continue
		}
		// skip whitechars if not within string
		if skip && unicode.IsSpace(r) {
			continue
		}
		if !escaped && r == '"' {
			skip = !skip
		}
		// handle escaped chars
		if escaped {
			escaped = false
		} else if r == '\\' {
			escaped = true
		}

		//=============================================================
		//	execute state logic
		//=============================================================
		switch state {
		//---------------------------------------------------------
		// next parameter: can be named or unnamed, value or list
		//---------------------------------------------------------
		case 1:
			{
				handled := false
				// operations on unnamed lists outside of string values
				if skip && !escaped {
					if r == '{' {
						// check for pending empty parameter
						if stack.IsTop(EMPTY) {
							stack.Pop()
						}
						// start unnamed list 
						param.Name = ""  // unnamed parameter
						rdr.UnreadRune() // putback first character
						state = 3        // read value
						handled = true
					} else if r == '}' {
						// check for pending empty parameter
						if stack.IsTop(EMPTY) {
							stack.Pop()
							param.reset()
							cb(EMPTY, param)
						}
						// end unnamed list
						rdr.UnreadRune() // putback first character
						state = 5        // read delimiter
						handled = true
					} else if r == ',' {
						// end unnamed empty parameter
						if !stack.IsTop(EMPTY) {
							stack.Push(EMPTY)
						}
						param.reset()
						cb(EMPTY, param)
						handled = true
					}
				}
				// check start of name
				if !handled {
					// check for pending empty parameter
					if stack.IsTop(EMPTY) {
						stack.Pop()
					}
					// named parameter; check first character
					if !unicode.IsLetter(r) && r != '"' {
						cb(ERROR, nil)
						return mkError("Invalid parameter name", line, offset)
					}
					// save initial character
					buf += string(r)
					// push VAR mode
					stack.Push(VAR)
					state = 2
				}
			}
		//---------------------------------------------------------
		// parse parameter name			
		//---------------------------------------------------------
		case 2:
			{
				switch r {

				// assignment found?
				case '=':
					// start parsing value
					param.Name = buf
					state = 3

				// delimiter, opening or closing brace?
				case '{', '}', ',':
					// assign to value.
					param.Name = ""
					param.Value = buf
					// notify value complete
					cb(VAR, param)
					param.reset()
					// pop VAR tag
					stack.Pop()
					// restart
					rdr.UnreadRune()
					state = 5

				// else collect char in name buffer
				default:
					buf += string(r)
				}
			}
		//---------------------------------------------------------
		// read generic value
		//---------------------------------------------------------
		case 3:
			{
				// test for value types:
				switch r {
				// begin of new list
				case '{':
					{
						// drop VAR tag from named list.
						if stack.IsTop(VAR) {
							stack.Pop()
						}
						// push LIST mode
						stack.Push(LIST)
						param.Value = "{"
						cb(LIST, param)
						param.reset()
					}
				// empty parameter
				case ',':
					{
						// named empty parameter?
						if stack.IsTop(VAR) {
							// yes: add empty parameter
							stack.Pop()
							param.Value = ""
							cb(VAR, param)
						} else {
							rdr.UnreadRune() // unread character
							state = 1        // start of parameter
						}
					}
				// begin new value or parameter
				default:
					{
						// check for parameter tag on top of stack...
						if stack.IsTop(VAR) {
							// parse value for defined parameter
							stack.Push(VALUE)
							buf = string(r)
							state = 4
						} else {
							// parse new parameter
							buf = string(r)
							state = 1
						}
					}
				}
			}
		//---------------------------------------------------------
		// read parameter value
		//---------------------------------------------------------
		case 4:
			{
				// drop escapes: use escaped character directly.
				if !escaped {
					// check for termination of value
					if skip && (r == '}' || r == ',') {
						// notify value complete
						param.Value = buf
						cb(VAR, param)
						param.reset()

						// pop VALUE and VAR tags
						stack.Pop()
						stack.Pop()

						rdr.UnreadRune() // restore read character
						state = 5        // read/handle delimiter
					} else {
						buf += string(r)
					}
				}
			}
		//---------------------------------------------------------
		// handle delimiter between parameters				
		//---------------------------------------------------------
		case 5:
			{
				switch r {
				// end of list
				case '}':
					{
						// check for correct parent type
						if !stack.IsTop(LIST) {
							cb(ERROR, nil)
							return mkError("Invalid structure '}'", line, offset)
						}
						// notify callback
						param.Value = "}"
						cb(LIST, param)
						param.reset()

						stack.Pop()
					}
				// non-delimiting character: "unread" character
				// and start new parameter expression
				default:
					{
						if stack.Len() > 0 {
							if stack.Peek() != VAR {
								cb(ERROR, nil)
								return mkError("Invalid structure.", line, offset)
							}
							stack.Pop()
						}
						rdr.UnreadRune()
					}
					fallthrough

				// sibling delimiter
				case ',':
					{
						buf = ""  // reset buffer
						state = 1 // restart with new parameter
					}
				}
			}
		}
	}
	// report success.
	return nil
}

///////////////////////////////////////////////////////////////////////
// Private methods

/*
 * generate error message.
 * @param msg string - error message
 * @param line int - current line number
 * @param offset int - offset into line
 * @return os.Error - error object
 */
func mkError(msg string, line int, offset int) error {
	out := msg + " (Line:" + strconv.Itoa(line)
	out += ", Offset:" + strconv.Itoa(offset) + ")"
	return errors.New(out)
}

///////////////////////////////////////////////////////////////////////
//	Revision history:
///////////////////////////////////////////////////////////////////////
//
//	Revision 2.0  2012-01-09 16:03:23  brf
//  First release as free software (GPL3+)
//
//	Revision 1.10  2010-11-28 13:13:44  brf
//	Handle comments.
//
//	Revision 1.9  2010-11-27 22:24:24  brf
//	Handle empty parameters; quoted values and parameter at the end of the definition.
//
//	Revision 1.8  2010-11-06 10:33:26  brf
//	Correctly handle escaped characters in value strings.
//
//	Revision 1.7  2010-11-02 20:19:31  brf
//	Data format adjusted (removed Section definition).
//
//	Revision 1.6  2010-11-02 20:11:02  brf
//	Added description.
//
//	Revision 1.5  2010-11-02 18:58:51  brf
//	Handle multiple top-level parameters.
//
//	Revision 1.4  2010-11-01 20:23:08  brf
//	Corrected handling of named lists.
//
//	Revision 1.3  2010-11-01 20:10:18  brf
//	Merged LIST and ARRAY processing.
//
//	Revision 1.2  2010-10-31 12:55:38  brf
//	Changed package name.
//
//	Revision 1.1  2010-10-31 11:57:33  brf
//	Initial revision.
//
