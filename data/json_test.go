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
	"encoding/json"
	"testing"
)

func TestJsonCompare(t *testing.T) {

	var x1 map[string]any
	if err := json.Unmarshal([]byte(testdataJSON[0]), &x1); err != nil {
		t.Fatal(err)
	}
	var x2 map[string]any
	if err := json.Unmarshal([]byte(testdataJSON[1]), &x2); err != nil {
		t.Fatal(err)
	}

	list, err := CompareJSON(x1, x2)
	if err != nil {
		t.Fatal(err)
	}
	for _, msg := range list {
		t.Log(msg)
	}
}

var testdataJSON = []string{`
{
  "metadata": {
    "file_id": "982b-410a",
    "version": 1.0
  },
  "system_active": true,
  "archive_data": null,
  "dataset": {
    "tags": [
      "random",
      "testing"
    ],
    "matrix_dimensions": [
      42,
      3.14159
    ],
    "validation": {
      "is_verified": false,
      "error_log": null
    }
  }
}    
`, `
{
  "metadata": {
    "file_id": "982b-410a-AMENDED",
    "version": 1.0
  },
  "system_active": true,
  "archive_data": null,
  "dataset": {
    "tags": [
      "random",
      "testing",
      "added_tag"
    ],
    "matrix_dimensions": [
      42,
      99.95
    ],
    "validation": {
      "is_verified": true,
      "error_log": null,
      "extra_config": {}
    }
  },
  "comment": "object added."
}
`}
