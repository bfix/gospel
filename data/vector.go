package data

// Vector data structure
type Vector struct {
	data [](interface{}) // list of elements
}

// NewVector instantiates a new (empty) Vector object.
func NewVector() *Vector {
	return &Vector{
		data: make([](interface{}), 0),
	}
}

// Len returns the number of elements in the vector.
func (vec *Vector) Len() int {
	return len(vec.data)
}

// Add element to the end of the vector.
func (vec *Vector) Add(v interface{}) {
	vec.data = append(vec.data, v)
}

// Insert element at given position. Add 'nil' elements if index
// is beyond the end of the vector.
func (vec *Vector) Insert(i int, v interface{}) {

	if i < 0 {
		// create a prepending slice
		pre := make([](interface{}), -i)
		pre[0] = v
		vec.data = append(pre, vec.data...)
	} else if i >= len(vec.data) {
		// create appending slice
		idx := i - len(vec.data) + 1
		app := make([](interface{}), idx)
		app[idx-1] = v
		vec.data = append(vec.data, app...)
	} else {
		pre := vec.data[:i]
		app := vec.data[i:]
		vec.data = append(append(pre, v), app...)
	}
}

// Drop the last element from the vector.
func (vec *Vector) Drop() (v interface{}) {
	pos := len(vec.data) - 1
	v, vec.data = vec.data[pos], vec.data[:pos]
	return
}

// Delete indexed element from the vector.
func (vec *Vector) Delete(i int) (v interface{}) {
	if i < 0 || i > len(vec.data)-1 {
		return nil
	}
	v = vec.data[i]
	vec.data = append(vec.data[:i], vec.data[i+1:]...)
	return
}

// At return the indexed element from vector.
func (vec *Vector) At(i int) (v interface{}) {
	if i < 0 || i > len(vec.data)-1 {
		return nil
	}
	return vec.data[i]
}
