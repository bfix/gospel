package ed25519

/*
 * --------------------------------------------------------------------
 * Ed25519 implementation inspired by the Python implementation from
 * https://ed25519.cr.yp.to/python/ed25519.py
 *
 * Ed25519 keys (as defined by the RFC 8032) can be used for:
 *   - EdDSA signatures
 *   - ECDSA signatures
 *   - ECDHE key exchange
 *
 * A private key can either defined by a secret seed (see RFC 8032) or
 * by specifying the private scalar 'd'. If the private key is defined
 * by 'd', it can only be used for ECDSA and ECDHE, but not for EdDSA
 * (as the seed which is required for EdDSA signing and verify) is
 * unknown and can't be computed.
 * --------------------------------------------------------------------
 *
 * (c) 2019 Bernd Fix   >Y<
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
