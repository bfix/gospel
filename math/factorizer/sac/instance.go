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
//*    PGMID.        INTERFACE FOR SIEVER/SOLVER INSTANCES.          */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

package sac

// An instance can run concurrently with other instances (of same or
// different type).<p>
type Instance interface {

	// Set instance identification (id and name).<p>
	// @param id int - identifier
	// @param name String - instance name
	Ident(id int, name string)

	// Return instance identifier.<p>
	// @return int - instance identifier
	GetId() int

	// Check if instance is active.<p>
	// @return boolean - instance active?
	IsActive() bool

	// Terminate instance.<p>
	Terminate()

	Run()
}

//********************************************************************/
//*    PGMID.        ABSTRACT BASE CLASS FOR INSTANCES.              */
//*    AUTHOR.       BERND R. FIX   >Y<                              */
//*    DATE WRITTEN. 08/04/07.                                       */
//*    COPYRIGHT.    (C) BY BERND R. FIX. ALL RIGHTS RESERVED.       */
//*                  LICENSED MATERIAL - PROGRAM PROPERTY OF THE     */
//*                  AUTHOR. REFER TO COPYRIGHT INSTRUCTIONS.        */
//*    REMARKS.                                                      */
//********************************************************************/

// An instance can run concurrently with other instances (of same or
// different type).<p>
type InstanceImpl struct {
	active bool   // instance active?
	id     int    // identifier
	name   string // instance name
}

// Set instance identification (id and name).<p>
// @param id int - identifier
// @param name String - instance name
func (i *InstanceImpl) Ident(id int, name string) {
	i.id = id
	i.name = name
}

// Return instance identifier.<p>
// @return int - instance identifier
func (i *InstanceImpl) GetId() int {
	return i.id
}

// Check if instance is active.<p>
// @return boolean - instance active?
func (i *InstanceImpl) IsActive() bool {
	return i.active
}

// Terminate instance.<p>
func (i *InstanceImpl) Terminate() {
	i.active = false
}
