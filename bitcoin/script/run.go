package script

import (
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"github.com/bfix/gospel/bitcoin/ecc"
	"github.com/bfix/gospel/bitcoin/util"
	"github.com/bfix/gospel/math"
	"math/big"
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
		"Tx invalid",
		"Type mismatch",
		"Invalid stack type",
		"Operation exceeds stack",
		"No transaction",
		"Unclosed IF",
		"Double ELSE",
		"Invalid transaction",
		"Invalid pubkey",
		"Invalid Uint",
		"Invalid signature",
		"Invalid transfer",
		"Not verified",
		"Disabled opcode",
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
	stmts    []*Statement // list of parsed statements
	pos      int          // index of current statement
	stack    *Stack       // stack for script operations
	altStack *Stack       // alternative stack
	tx       []byte       // associated dissected transaction
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

// Exec executes a script belonging to a transaction. If no transaction is
// specified, some script opcodes like OpCHECKSIG could not be executed.
// N.B.: To successfully execute 'script' that involves OpCHECKSIG it needs
// to be assembled (concatenated) and cleaned up from the prev.sigScript and
// curr.pkScript (see https://en.bitcoin.it/wiki/OpCHECKSIG); 'tx' is the
// current transaction already prepared for signature.
func (r *R) ExecScript(script []byte, tx []byte) (bool, int) {
	r.tx = tx
	if rc := r.parse(script); rc != RcOK {
		return false, rc
	}
	return r.exec()
}

// CheckSig performs a OpCHECKSIG operation on the stack (without pushing a
// result onto the stack)
func (r *R) CheckSig() (bool, int) {
	// pop pubkey from stack
	pkInt, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	pk, err := ecc.PublicKeyFromBytes(pkInt.Bytes())
	if err != nil {
		return false, RcInvalidPubkey
	}
	// pop signature from stack
	sigInt, rc := r.stack.Pop()
	if rc != RcOK {
		return false, rc
	}
	sigData := sigInt.Bytes()
	hashType := sigData[len(sigData)-1]
	sigData = sigData[:len(sigData)-1]
	// compute hash of amended transaction
	txSign := append(r.tx, []byte{hashType, 0, 0, 0}...)
	txHash := util.Hash256(txSign)
	// decode signature from DER data
	var sig struct{ R, S *big.Int }
	_, err = asn1.Unmarshal(sigData, &sig)
	if err != nil {
		return false, RcInvalidSignature
	}
	sigR := math.NewIntFromBig(sig.R)
	sigS := math.NewIntFromBig(sig.S)
	// perform signature verify
	return ecc.Verify(pk, txHash, sigR, sigS), RcOK
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
