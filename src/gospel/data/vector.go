/*
 * Vector: Indexable data class for generic types.
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
// Generic Vector type and implementation.

/*
 * Generic Vector data structure
 */
type Vector struct {
	data [](interface{}) // list of elements
}

//=====================================================================
/*
 * Instantiate a new (empty) Vector object.
 */
func NewVector() *Vector {
	return &Vector{
		data: make([](interface{}), 0),
	}
}

//---------------------------------------------------------------------
/*
 * Get number of elements inthe vector.
 * @return int - number of elements
 */
func (self *Vector) Len() int {
	return len(self.data)
}

//---------------------------------------------------------------------
/*
 * Add element to the end of the vector.
 * @param v interface{} - element to be added
 */
func (self *Vector) Add(v interface{}) {
	self.data = append(self.data, v)
}

//---------------------------------------------------------------------
/*
 * Insert element at given position. Add 'nil' elements if index
 * is beyond the end of the vector.
 * @param i int - insert position
 * @param v interface{} - element to be inserted
 */
func (self *Vector) Insert(i int, v interface{}) {

	if i < 0 {
		// create a prepending slice
		pre := make([](interface{}), -i)
		pre[0] = v
		self.data = append(pre, self.data...)
	} else if i >= len(self.data) {
		// create appending slice
		idx := i - len(self.data) + 1
		app := make([](interface{}), idx)
		app[idx-1] = v
		self.data = append(self.data, app...)
	} else {
		pre := self.data[:i-1]
		app := self.data[i:]
		self.data = append(append(pre, v), app...)
	}
}

//---------------------------------------------------------------------
/*
 * Drop the last element from the vector.
 * @return v interface{} - dropped element
 */
func (self *Vector) Drop() (v interface{}) {
	pos := len(self.data) - 1
	v, self.data = self.data[pos], self.data[:pos]
	return
}

//---------------------------------------------------------------------
/*
 * Delete indexed element from the vector.
 * @param i int - position of element to be deleted
 * @return v interface{} - deleted element
 */
func (self *Vector) Delete(i int) (v interface{}) {
	if i < 0 || i > len(self.data)-1 {
		return nil
	}
	v = self.data[i]
	self.data = append(self.data[:i], self.data[i+1:]...)
	return
}

//---------------------------------------------------------------------
/*
 * Get indexed element from vector.
 * @param i int - position of element
 * @return v interface{} - indexed element
 */
func (self *Vector) At(i int) (v interface{}) {
	if i < 0 || i > len(self.data)-1 {
		return nil
	}
	return self.data[i]
}
