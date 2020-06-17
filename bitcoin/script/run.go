package script

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2020 Bernd Fix
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

import (
	"fmt"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/math"
)

// Result codes returned by script functions.
const (
	RcOK = iota
	RcErr
	RcExceeds
	RcParseError
	RcScriptError
	RcLengthMismatch
	RcEmptyStack
	RcInvalidFinalStack
	RcNotImplemented
	RcInvalidOpcode
	RcReservedOpcode
	RcTxInvalid
	RcTypeMismatch
	RcInvalidStackType
	RcExceedsStack
	RcNoTransaction
	RcUnclosedIf
	RcDoubleElse
	RcInvalidTransaction
	RcInvalidPubkey
	RcInvalidUint
	RcInvalidSignature
	RcInvalidTransfer
	RcNotVerified
	RcDisabledOpcode
	RcTxNotSignable
	RcEmptyScript
)

// Human-readable result codes
var (
	RcString = []string{
		"OK",
		"Generic error",
		"Operation exceeds available data",
		"Parse error",
		"Script error",
		"Length mismatch",
		"Empty stack",
		"Invalid final stack",
		"Not implemented yet",
		"Invalid opcode",
		"Reserved opcode",
		"Invalid transaction",
		"Type mismatch",
		"Invalid stack type",
		"Operation exceeds stack",
		"No transaction available",
		"Unclosed IF",
		"Double ELSE",
		"Invalid transaction",
		"Invalid pubkey",
		"Invalid Uint",
		"Invalid signature",
		"Invalid transfer",
		"Not verified",
		"Disabled opcode",
		"Transaction not signable",
		"Empty script",
	}
)

// R is the Bitcoin script runtime environment
type R struct {
	script   *Script                       // list of parsed statements
	pos      int                           // index of current statement
	stack    *Stack                        // stack for script operations
	altStack *Stack                        // alternative stack
	tx       *bitcoin.DissectedTransaction // associated dissected transaction
	CbStep   func(stack *Stack, stmt *Statement, rc int)
}

// NewRuntime creates a new script parser and execution runtime.
func NewRuntime() *R {
	return &R{
		script:   nil,
		pos:      -1,
		stack:    NewStack(),
		altStack: NewStack(),
		tx:       nil,
		CbStep:   nil,
	}
}

// ExecScript executes a script belonging to a transaction. If no transaction is
// specified, some script opcodes like OpCHECKSIG could not be executed.
// N.B.: To successfully execute 'script' that involves OpCHECKSIG it needs
// to be assembled (concatenated) and cleaned up from the prev.sigScript and
// curr.pkScript (see https://en.bitcoin.it/wiki/OpCHECKSIG); 'tx' is the
// current transaction in dissected format already prepared for signature.
func (r *R) ExecScript(script *Script, tx *bitcoin.DissectedTransaction) (bool, int) {
	if tx.Signable == nil || tx.VinSlot < 0 {
		return false, RcTxNotSignable
	}
	r.tx = tx
	return r.exec(script)
}

// CheckSig performs a OpCHECKSIG operation on the stack (without pushing a
// result onto the stack)
func (r *R) CheckSig() (bool, int) {
	// pop pubkey from stack
	pkInt, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	// pop signature from stack
	sigInt, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	// perform signature verify
	return r.checkSig(pkInt, sigInt)
}

// CheckMultiSig performs a OpCHECKMULTISIG operation on the stack (without
// pushing a result onto the stack).
func (r *R) CheckMultiSig() (bool, int) {
	// pop pubkeys from stack
	nk, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	var keys []*math.Int
	for i := 0; i < int(nk.Int64()); i++ {
		pkInt, rc := r.stack.Pop()
		if rc != RcOK {
			return false, rc
		}
		keys = append(keys, pkInt)
	}
	// pop signatures from stack
	ns, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	var sigs []*math.Int
	for i := 0; i < int(ns.Int64()); i++ {
		sigInt, rc := r.stack.Pop()
		if rc != RcOK {
			return false, rc
		}
		sigs = append(sigs, sigInt)
	}
	// pop extra (due to a bug in the initial implementation)
	if _, rc := r.stack.Pop(); rc != RcOK {
		return false, rc
	}
	// perform signature verifications
	for _, sigInt := range sigs {
		var (
			j     int
			pkInt *math.Int
			valid = false
			rc    int
		)
		for j, pkInt = range keys {
			if pkInt == nil {
				continue
			}
			valid, rc = r.checkSig(pkInt, sigInt)
			if rc != RcOK {
				return false, rc
			}
			if valid {
				break
			}
		}
		if valid {
			keys[j] = nil
		} else {
			return false, RcOK
		}
	}
	return true, RcOK
}

// exec executes a sequence of parsed statement of a script.
func (r *R) exec(script *Script) (bool, int) {
	r.script = script
	if r.script.Stmts == nil || len(r.script.Stmts) == 0 {
		return false, RcEmptyScript
	}
	r.pos = 0
	size := len(r.script.Stmts)
	for r.pos < size {
		s := r.script.Stmts[r.pos]
		opc := GetOpcode(s.Opcode)
		if opc == nil {
			fmt.Printf("Opcode: %v\n", s.Opcode)
			return false, RcInvalidOpcode
		}
		rc := opc.Exec(r)
		if r.CbStep != nil {
			r.CbStep(r.stack, s, rc)
		}
		if rc != RcOK {
			return false, rc
		}
		r.pos++
	}
	if r.stack.Len() == 1 {
		v, rc := r.stack.Pop()
		if rc != RcOK {
			return false, rc
		}
		if v.Equals(math.ONE) {
			return true, RcOK
		}
		return false, RcOK
	}
	return false, RcInvalidFinalStack
}

// checkSig checks the signature of a prepared transaction.
func (r *R) checkSig(pkInt, sigInt *math.Int) (bool, int) {
	if r.tx == nil {
		return false, RcNoTransaction
	}
	// get public key
	pk, err := bitcoin.PublicKeyFromBytes(pkInt.Bytes())
	if err != nil {
		return false, RcInvalidPubkey
	}
	// get signature and hash type
	sigData := sigInt.Bytes()
	hashType := sigData[len(sigData)-1]
	sigData = sigData[:len(sigData)-1]
	// compute hash of amended transaction
	txSign := append(r.tx.Signable, []byte{hashType, 0, 0, 0}...)
	txHash := bitcoin.Hash256(txSign)
	// decode signature from DER data
	sig, err := bitcoin.NewSignatureFromASN1(sigData)
	if err != nil {
		return false, RcInvalidSignature
	}
	// perform signature verify
	return bitcoin.Verify(pk, txHash, sig), RcOK
}
