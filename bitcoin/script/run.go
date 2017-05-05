package script

import (
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"github.com/bfix/gospel/math"
)

// Result codes returned by script functions.
const (
	RcOK = iota
	RcErr
	RcExceeds
	RcParseError
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
)

// Human-readable result codes
var (
	RcString = []string{
		"OK",
		"Generic error",
		"Operation exceeds available data",
		"Parse error",
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
	}
)

// Statement is a single script statement.
type Statement struct {
	Opcode byte
	Data   []byte
}

// String returns the string representation of a statement.
func (s *Statement) String() string {
	if s.Data != nil {
		return hex.EncodeToString(s.Data)
	}
	return GetOpcode(s.Opcode).Name
}

// R is the Bitcoin script runtime environment
type R struct {
	stmts    []*Statement               // list of parsed statements
	pos      int                        // index of current statement
	stack    *Stack                     // stack for script operations
	altStack *Stack                     // alternative stack
	tx       *util.DissectedTransaction // associated dissected transaction
	CbStep   func(stack *Stack, stmt *Statement, rc int)
}

// NewRuntime creates a new script parser and execution runtime.
func NewRuntime() *R {
	return &R{
		stmts:    nil,
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
func (r *R) ExecScript(script []byte, tx *util.DissectedTransaction) (bool, int) {
	if tx.Signable == nil || tx.VinSlot < 0 {
		return false, RcTxNotSignable
	}
	r.tx = tx
	if rc := r.parse(script); rc != RcOK {
		return false, rc
	}
	return r.exec()
}

// GetTemplate returns a template derived from a script. A template only
// contains a sequence of opcodes; it is used to find structural equivalent
// scripts (but with varying data).
func (r *R) GetTemplate(script []byte) (tpl []byte, rc int) {
	if rc = r.parse(script); rc != RcOK {
		return
	}
	for _, s := range r.stmts {
		tpl = append(tpl, s.Opcode)
	}
	return
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
func (r *R) exec() (bool, int) {
	if r.stmts == nil || len(r.stmts) == 0 {
		return false, RcEmptyStack
	}
	r.pos = 0
	size := len(r.stmts)
	for r.pos < size {
		s := r.stmts[r.pos]
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

// parse dissects binary scripts into a sequence of tokens.
func (r *R) parse(code []byte) int {
	var (
		pos    int
		size   int
		length int
	)
	getData := func(s *Statement, i int) int {
		b := make([]byte, i)
		copy(b, code[pos+1:pos+i+1])
		j, err := util.GetUint(b, 0, i)
		if err != nil {
			return RcLengthMismatch
		}
		n := int(j)
		size += n + i
		if pos+size > length {
			return RcExceeds
		}
		s.Data = make([]byte, n)
		copy(s.Data, code[pos+i+1:pos+i+n+1])
		return RcOK
	}
	r.stmts = make([]*Statement, 0)
	length = len(code)
	for pos < length {
		size = 1
		op := code[pos]
		s := &Statement{Opcode: op}
		if op > 0 && op < 76 {
			n := int(op)
			if pos+n+1 > length {
				return RcExceeds
			}
			s.Data = make([]byte, n)
			copy(s.Data, code[pos+1:pos+n+1])
			size += n
		} else {
			switch op {
			case OpPUSHDATA1:
				if rc := getData(s, 1); rc != RcOK {
					return rc
				}
			case OpPUSHDATA2:
				if rc := getData(s, 2); rc != RcOK {
					return rc
				}
			case OpPUSHDATA4:
				if rc := getData(s, 4); rc != RcOK {
					return rc
				}
			}
		}
		pos += size
		r.stmts = append(r.stmts, s)
	}
	return RcOK
}

// checkSig checks the signature of a prepared transaction.
func (r *R) checkSig(pkInt, sigInt *math.Int) (bool, int) {
	if r.tx == nil {
		return false, RcNoTransaction
	}
	// get public key
	pk, err := ecc.PublicKeyFromBytes(pkInt.Bytes())
	if err != nil {
		return false, RcInvalidPubkey
	}
	// get signature and hash type
	sigData := sigInt.Bytes()
	hashType := sigData[len(sigData)-1]
	sigData = sigData[:len(sigData)-1]
	// compute hash of amended transaction
	txSign := append(r.tx.Signable, []byte{hashType, 0, 0, 0}...)
	txHash := util.Hash256(txSign)
	// decode signature from DER data
	sig, err := ecc.NewSignatureFromASN1(sigData)
	if err != nil {
		return false, RcInvalidSignature
	}
	// perform signature verify
	return ecc.Verify(pk, txHash, sig), RcOK
}
