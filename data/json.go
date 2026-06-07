//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-present, Bernd Fix  >Y<
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

package data

import (
	"fmt"
	"path/filepath"
)

// CompareJSON compares to JSON objects and returns a list of differences
// between the two objects.
func CompareJSON(j1, j2 map[string]any) (list []string, err error) {
	// start at root of objects
	return compareAny(j1, j2, "/")
}

// compareAny is the worker for CompareJSON.
// 'j1' is the reference object for messages, so e.g. "missing" means that an
// element of 'j1' is missing in 'j2'. Objects are addressed in a filepath
// style like '/dataset/validation/is_verified'.
func compareAny(j1, j2 any, path string) (list []string, err error) {
	switch x1 := j1.(type) {
	case map[string]any:
		x2, ok := j2.(map[string]any)
		if !ok {
			list = append(list, fmt.Sprintf("%s: different type", path))
			return
		}
		for key, v1 := range x1 {
			curr := filepath.Join(path, key)
			v2, ok := x2[key]
			if !ok {
				// element of 'j1' is missing in 'j2'
				list = append(list, fmt.Sprintf("%s: missing", curr))
				continue
			}
			// compare sub-values
			var t []string
			if t, err = compareAny(v1, v2, curr); err == nil {
				list = append(list, t...)
			}
		}
		for key := range x2 {
			curr := filepath.Join(path, key)
			if _, ok := x1[key]; !ok {
				// element of 'j2' is missing in 'j1'
				list = append(list, fmt.Sprintf("%s: included", curr))
			}
		}
	case nil:
		if j2 != nil {
			list = append(list, fmt.Sprintf("%s: target not <nil>", path))
		}
	case string:
		x2, ok := j2.(string)
		if !ok {
			list = append(list, fmt.Sprintf("%s: different type (string)", path))
			return
		}
		if x1 != x2 {
			list = append(list, fmt.Sprintf("%s: '%s' != '%s'", path, x1, x2))
		}
	case float64:
		x2, ok := j2.(float64)
		if !ok {
			list = append(list, fmt.Sprintf("%s: different type (float64)", path))
			return
		}
		if x1 != x2 {
			list = append(list, fmt.Sprintf("%s: %f != %f", path, x1, x2))
		}
	case bool:
		x2, ok := j2.(bool)
		if !ok {
			list = append(list, fmt.Sprintf("%s: different type (bool)", path))
			return
		}
		if x1 != x2 {
			list = append(list, fmt.Sprintf("%s: %v != %v", path, x1, x2))
		}
	case []any:
		x2, ok := j2.([]any)
		if !ok {
			list = append(list, fmt.Sprintf("%s: different type ([]any)", path))
			return
		}
		if len(x1) != len(x2) {
			list = append(list, fmt.Sprintf("%s: different length", path))
			return
		}
		for i, y1 := range x1 {
			y2 := x2[i]
			var t []string
			if t, err = compareAny(y1, y2, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return
			}
			list = append(list, t...)
		}
	}
	return
}
