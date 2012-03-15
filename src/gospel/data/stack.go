/*
 * Stack: FIFO stack data class for generic types and integers.
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

package data

///////////////////////////////////////////////////////////////////////
// Generic Stack type and implementation.

/*
 * Stack for generic data types.
 */
type Stack struct {
	data [](interface{}) // list of stack elements
}

//=====================================================================
/*
 * Instantiate a new generic Stack object.
 * @return *Stack - reference to new instance
 */
func NewStack() *Stack {
	return &Stack{
		data: make([](interface{}), 0),
	}
}

//---------------------------------------------------------------------
/*
 * Pop last entry from stack and return it to caller.
 * @return v interface{} - generic stack entry
 */
func (self *Stack) Pop() (v interface{}) {
	pos := len(self.data) - 1
	v, self.data = self.data[pos], self.data[:pos]
	return
}

//---------------------------------------------------------------------
/*
 * Push generic entry to stack.
 * @param v interface{} - generic stack entry
 */
func (self *Stack) Push(v interface{}) {
	self.data = append(self.data, v)
}

//---------------------------------------------------------------------
/*
 * Return number of elements on stack.
 * @return int - number of elements on stack
 */
func (self *Stack) Len() int {
	return len(self.data)
}

//---------------------------------------------------------------------
/*
 * Peek at the last element pushed to stack without dropping it.
 * @return v interface{} - generic stack entry
 */
func (self *Stack) Peek() (v interface{}) {
	pos := len(self.data) - 1
	if pos < 0 {
		return nil
	}
	return self.data[pos]
}

///////////////////////////////////////////////////////////////////////
// Integer-based Stack type and implementation.

type IntStack struct {
	data []int // list of stack elements
}

//=====================================================================
/*
 * Instantiate a new integer-based Stack object.
 * @return *IntStack - reference to new stack instance
 */
func NewIntStack() *IntStack {
	return &IntStack{
		data: make([]int, 0),
	}
}

//---------------------------------------------------------------------
/*
 * Pop last entry from stack and return it to caller.
 * @return v int - last top stack entry
 */
func (self *IntStack) Pop() (v int) {
	pos := len(self.data) - 1
	v, self.data = self.data[pos], self.data[:pos]
	return
}

//---------------------------------------------------------------------
/*
 * Push entry to stack.
 * @param v int - new top stack entry
 */
func (self *IntStack) Push(v int) {
	self.data = append(self.data, v)
}

//---------------------------------------------------------------------
/*
 * Return number of elements on stack.
 * @return int - number of elements on stack
 */
func (self *IntStack) Len() int {
	return len(self.data)
}

//---------------------------------------------------------------------
/*
 * Peek at the last element pushed to stack without dropping it.
 * @return v int - top stack entry (retained)
 */
func (self *IntStack) Peek() (v int) {
	return self.data[len(self.data)-1]
}

//---------------------------------------------------------------------
/*
 * Compare last element with given value. 
 * @param v int - value to compare to
 * @return bool - do value and top element match?
 */
func (self *IntStack) IsTop(v int) bool {
	pos := len(self.data) - 1
	if pos < 0 {
		return false
	}
	return self.data[pos] == v
}
