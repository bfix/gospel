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

//********************************************************************/
//*    PGMID.        RELATION QUEUE IMPLEMENTATION.                  */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package sac

import (
	"bytes"
	"fmt"
	"sync"
)

// A queue is a FIFO stack for relations.>p>
// Adding and retrieving relations is thread-safe.<p>
type Queue struct {
	data      []Relation // list of entries
	semaphore sync.Mutex // access lock
	id        int        // identifier
}

// Instanciate a new queue with given identifier.<p>
// @param id int - queue identifier
func NewQueue(id int) *Queue {
	return &Queue{
		data: make([]Relation, 0),
		id:   id,
	}
}

// Put relation into queue (append to end of list).<p>
// @param r Relation - relation to be stored
func (q *Queue) Put(r Relation) {
	q.semaphore.Lock()
	defer q.semaphore.Unlock()

	// append entry at the end of vector.
	q.data = append(q.data, r)
}

// Get relation from queue.<p>
// If the queue is empty, the method waits for a new entry
// to arrive in the queue. A return value of 'null' indicates
// an error/exception.<p>
// @return Relation - retrieved relation (first in list)
func (q *Queue) Get() (r Relation) {

	// is there any entry in the queue?
	if len(q.data) == 0 {
		// wait for an entry.  @@@
	}
	// only one thread at a time...
	q.semaphore.Lock()
	defer q.semaphore.Unlock()

	// someone has stolen an entry ?!
	// (probably due to a call to the 'reset()' method.
	if len(q.data) == 0 {
		// no entry available.
		return nil
	}

	// get the next entry out of the queue.
	r = q.data[0]
	q.data = q.data[1:]
	return
}

// Get number of entries in the queue.<p>
// @return
func (q *Queue) NumEntries() int {
	return len(q.data)
}

// Generate a printable representation of the queue object.<p>
// @return String - queue in printable (and readable) form
func (q *Queue) String() string {
	q.semaphore.Lock()
	defer q.semaphore.Unlock()

	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "[Queue %d]", q.id)
	for n := 0; n < len(q.data); n++ {
		fmt.Fprintf(buf, "\n   %v", q.data[n])
	}
	return buf.String()
}
