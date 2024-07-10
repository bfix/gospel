package bitcoin

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

/*
 * ====================================================================
 * Elliptic curve cryptography (ECDSA) based on curve "Secp256k1"
 * ====================================================================
 * The elliptic curve used for Bitcoin crypto ("Secp256k1") is a
 * short-form Weierstrass curve of the form "E: y² = x³ + 7" over
 * an underlying prime field of order "p". The parameter values used
 * are defined in [http://www.secg.org/collateral/sec2_final.pdf]
 * on page 15.
 */
