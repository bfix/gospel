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
//*    PGMID.        DIRECTOR -- CONTROLS PARALLEL INSTANCES.        */
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
	"sync"

	"github.com/bfix/gospel/math"
)

const (
	// number of parallel siever instances
	NUM_SIEVERS = 20

	// number of parallel solver instances.
	NUM_SOLVERS = 10
)

// The director is the controller for solver and siever instances
// in a parallel implementation.<p>
type Director struct {
	semaphore     sync.Mutex // wait for completion lock
	modulus       *math.Int  // number to be factorized
	factor        *math.Int  // found factor
	sievers       []Siever   // list of siever instances
	activeSievers int        // number of active sievers
	solvers       []Solver   // list of solver instances
	activeSolvers int        // number of active solvers
	sieveSize     *math.Int  // sieve interval size (per siever)
	fb            FactorBase // complete factor base
}

// Siever response: found new relation.<p>
// @param id int - siever identifier
// @param x BigInteger - x value of relation (function arg)
// @param y BigInteger - y value (function result)
// @return boolean - relation accepted?
func (d *Director) Sieved(id int, x, y *math.Int) bool {

	// if a siever terminates, it passes 'null' arguments.
	if x == nil || y == nil {
		d.activeSievers--
		return false
	}
	// create new relation and put it into the first solver instance queue.
	r := new(RelationImpl)
	r.Init(x, y)
	d.solvers[0].Put(r)
	return true
}

// Solver response: Relation r has finished with given result.<p>
// @param id int - solver identifier
// @param r Relation - processed relation
// @param rc Result - processing state
func (d *Director) Handled(id int, r Relation, rc int) {

	switch rc {
	// state REDUCABLE: pass relation to next solver.
	case REDUCABLE:
		// check for end of chain
		if id == NUM_SOLVERS {
			log.Println("FAIL: INCOMPLETE relation at end of chain")
			log.Printf("      %v", r)
			log.Fatal("<Director>")
		}
		// pass relation to next instance
		d.solvers[id].Put(r)

	//---------------------------------------------------------
	// state REDUCED: last instance reduced h of relation
	case REDUCED:
		// if relation is now completely reduced, we
		// can restart processing of the relation
		// with the first instance.
		if r.IsReduced() {
			d.solvers[0].Put(r)
			return
		} else {
			// otherwise we pass it on to next instance
			if id == NUM_SOLVERS {
				log.Println("FAIL: REDUCED relation at end of chain")
				log.Printf("      %v", r)
				log.Fatal("<Director>")
			}
			// pass relation to next instance
			d.solvers[id].Put(r)
		}

	//---------------------------------------------------------
	// state INSERTED:  relation was added to last solver instance
	case INSERTED:
		// do nothing

	//---------------------------------------------------------
	// state DISJUNCT:  relation has disjunct f over fb
	case DISJUNCT:
		// check for solution.
		a_b := r.IsSquared()
		if a_b != nil {
			// try factorization
			t := d.modulus.GCD(a_b)
			if !t.Equals(d.modulus) && !t.Equals(math.ONE) {

				// we found a factor!!
				log.Printf("[Director] Solution found: %v", t)
				d.factor = t

				// release lock on semaphore.
				d.semaphore.Unlock()
			}
			// trivial factor ignored.
			return
		}
		// check for end of chain
		if id == NUM_SOLVERS {
			log.Println("FAIL: DISJUNCT relation at end of chain")
			log.Printf("      %v", r)
			log.Fatal("<Director>")
		}
		// pass relation to next instance
		d.solvers[id].Put(r)
	}
}

// Factorize integer m and return a factor.<p>
// @param m BigInteger - value to be factorized
// @return BigInteger - factor of m
func (d *Director) Factorize(m *math.Int) *math.Int {
	d.modulus = m

	// make sure that (NUM_SIEVERS * sieveSize + R)^2 < 2*M
	// with R = floor(sqrt(M))
	r := m.NthRoot(2, false)
	d.sieveSize = math.TWO.Mul(m).Sub(r).Div(math.NewInt(NUM_SIEVERS))

	// instanciate (complete,global) factor base.
	d.fb = new(FactorBaseImpl)
	d.fb.Init(m)
	B := d.fb.GetNumPrimes()

	// instanciate sievers
	d.sievers = make([]Siever, NUM_SIEVERS)
	for n := 0; n < NUM_SIEVERS; n++ {
		d.sievers[n] = new(SieverImpl)
		d.sievers[n].Init(n+1, m, d.fb, d.Sieved)

		x0 := math.NewInt(int64(n)).Mul(d.sieveSize).Add(math.ONE)
		x1 := math.NewInt(int64(n + 1)).Mul(d.sieveSize)
		d.sievers[n].SetSieveInterval(x0, x1)
	}
	d.activeSievers = NUM_SIEVERS

	// instanciate solvers
	b := B / NUM_SOLVERS
	d.solvers = make([]Solver, NUM_SOLVERS)
	for n := 0; n < NUM_SOLVERS; n++ {
		d.solvers[n] = new(SolverImpl)

		fbNum := b
		if n == NUM_SOLVERS-1 {
			fbNum = B - b*(NUM_SOLVERS-1)
		}
		log.Printf("[subFb] %d - %d", n*b, n*b+fbNum-1)
		sub := d.fb.GetSubBase(n+1, n*b, fbNum)
		d.solvers[n].Init(n+1, m, sub, d.Handled)
	}
	d.activeSolvers = NUM_SOLVERS

	// start threads
	for n := 0; n < NUM_SIEVERS; n++ {
		go d.sievers[n].Run()
	}
	for n := 0; n < NUM_SOLVERS; n++ {
		go d.solvers[n].Run()
	}

	// wait for completion (solution)
	d.waitForCompletion()
	log.Printf("[Director]: Solution found: %v", d.factor)

	// terminate instances.
	for n := 0; n < NUM_SIEVERS; n++ {
		if d.sievers[n].IsActive() {
			d.sievers[n].Terminate()
		}
	}
	for n := 0; n < NUM_SOLVERS; n++ {
		if d.solvers[n].IsActive() {
			d.solvers[n].Terminate()
		}
	}

	// return factor of m
	return d.factor
}

// Wait for solver instances to find solution (semaphore lock)<p>
func (d *Director) waitForCompletion() {
	d.semaphore.Lock()
}
