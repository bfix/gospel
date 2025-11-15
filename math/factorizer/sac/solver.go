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
//*    PGMID.        SOLVER INTERFACE.                               */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package sac

import (
	"log"

	"github.com/bfix/gospel/math"
)

const (
	REDUCABLE = iota
	REDUCED
	INSERTED
	DISJUNCT
)

// Response -- relation handled with result state.<p>
// @param id int - solver identifier
// @param r Relation - processed relation
// @param rc Result - result / processing state
type SolverCallback func(id int, r Relation, rc int)

// Generic interface for solver instances.<p>
type Solver interface {
	Instance

	// Put relation into input queue for solver.<p>
	// This method is called by another thread (Director).<p>
	// @param r Relation - new relation for solver
	// @return boolean - delivery successful?
	Put(r Relation) bool

	// Initialize solver instance.<p>
	// @param id int - solver identifier
	// @param m BigInteger - number to be factorized
	// @param fb Factorbase - factor base to be used
	// @param cb SolverCallback - callback for requests/responses
	// @return boolean - successful operation?
	Init(id int, m *math.Int, fb FactorBase, cb SolverCallback) bool

	// Return number of pending relations in input queue.<p>
	// @return int - number of pending relations
	NumPending() int
}

//********************************************************************/
//*    PGMID.        IMPLEMENTATION OF SOLVER INTERFACE.             */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

// A solver instance handle relations for a section of the factor base.<p>
// A transformed relation is returned together with state information.<p>
type SolverImpl struct {
	InstanceImpl

	queue *Queue         // input queue for solver
	cb    SolverCallback // request/response handler

	list     []Relation // list of relations (size = #primes in fb)
	inserted int        // inserted relations

	modulus *math.Int  // number to be factorized
	fb      FactorBase // sub factor base of solver
}

// Initialize solver instance.<p>
// @param id int - solver identifier
// @param m BigInteger - number to be factorized
// @param fb Factorbase - factor base to be used
// @param cb SolverCallback - callback for requests/responses
// @return boolean - successful operation?
func (s *SolverImpl) Init(id int, m *math.Int, fb FactorBase, cb SolverCallback) bool {
	s.Ident(id, "solver")
	s.cb = cb
	s.fb = fb
	s.modulus = m

	s.queue = NewQueue(id)
	s.list = make([]Relation, s.fb.GetNumPrimes())
	s.inserted = 0
	return true
}

// Put relation into input queue for solver.<p>
// This method is called by another thread (Director).<p>
// @param r Relation - new relation for solver
// @return boolean - delivery successful?
func (s *SolverImpl) Put(r Relation) bool {
	s.queue.Put(r)
	return true
}

// Return number of pending relations in input queue.<p>
// @return int - number of pending relations
func (s *SolverImpl) NumPending() int {
	return s.queue.NumEntries()
}

// Start the solver process.<p>
func (s *SolverImpl) Run() {

	log.Println("Starting...")
	s.active = true
	for s.active {
		// get next relation.
		r := s.queue.Get()
		if r == nil {
			break
		}

		// if a relation is not completely reduced (h > 1),
		// try to reduce it against our factor base.
		if !r.IsReduced() {
			// can the relation be reduced over factor base?
			if r.IsReducable(s.fb) {
				// yes: reduce relation and return it
				r.Normalize(s.fb, s.modulus)
				s.cb(s.id, r, REDUCED)
			} else {
				// some other solver must finish reduction.
				s.cb(s.id, r, REDUCABLE)
			}
			// proceed with next relation
			continue
		}

		// loop forever (shouldn't happen...)
		pos := -1
		for {

			// if yf is not covered by our factor base,
			// we return this relation for further processing
			if !r.IsCovered(s.fb) {
				s.cb(s.id, r, DISJUNCT)
				break
			}

			// insert the relation if there is an empty slot in the list
			pos = r.FirstPrimeIndex(s.fb, pos)
			if s.list[pos] == nil {
				// store in empty slot
				s.list[pos] = r
				s.inserted++
				log.Printf("Relation #%d of %d", s.inserted, s.fb.GetNumPrimes())
				s.cb(s.id, nil, INSERTED)
				break
			}

			// Next multiply the input relation with the list entry.
			// This removes the smallest odd prime power and the
			// transformed relation is checked again.
			r.Multiply(s.list[pos], s.fb, s.modulus)
		}
	}
	log.Println("Terminating...")
}
