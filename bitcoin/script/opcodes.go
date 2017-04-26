package script

import (
	"fmt"
	"github.com/bfix/gospel/bitcoin/util"
	"github.com/bfix/gospel/math"
)

const (
	// Opcodes
	Op_FALSE               = 0
	Op_PUSHDATA1           = 76
	Op_PUSHDATA2           = 77
	Op_PUSHDATA4           = 78
	Op_1NEGATE             = 79
	Op_RESERVED            = 80
	Op_TRUE                = 81
	Op_2                   = 82
	Op_3                   = 83
	Op_4                   = 84
	Op_5                   = 85
	Op_6                   = 86
	Op_7                   = 87
	Op_8                   = 88
	Op_9                   = 89
	Op_10                  = 90
	Op_11                  = 91
	Op_12                  = 92
	Op_13                  = 93
	Op_14                  = 94
	Op_15                  = 95
	Op_16                  = 96
	Op_NOP                 = 97
	Op_VER                 = 98
	Op_IF                  = 99
	Op_NOTIF               = 100
	Op_VERIF               = 101
	Op_VERNOTIF            = 102
	Op_ELSE                = 103
	Op_ENDIF               = 104
	Op_VERIFY              = 105
	Op_RETURN              = 106
	Op_TOALTSTACK          = 107
	Op_FROMALTSTACK        = 108
	Op_2DROP               = 109
	Op_2DUP                = 110
	Op_3DUP                = 111
	Op_2OVER               = 112
	Op_2ROT                = 113
	Op_2SWAP               = 114
	Op_IFDUP               = 115
	Op_DEPTH               = 116
	Op_DROP                = 117
	Op_DUP                 = 118
	Op_NIP                 = 119
	Op_OVER                = 120
	Op_PICK                = 121
	Op_ROLL                = 122
	Op_ROT                 = 123
	Op_SWAP                = 124
	Op_TUCK                = 125
	Op_CAT                 = 126
	Op_SUBSTR              = 127
	Op_LEFT                = 128
	Op_RIGHT               = 129
	Op_SIZE                = 130
	Op_INVERT              = 131
	Op_AND                 = 132
	Op_OR                  = 133
	Op_XOR                 = 134
	Op_EQUAL               = 135
	Op_EQUALVERIFY         = 136
	Op_RESERVED1           = 137
	Op_RESERVED2           = 138
	Op_1ADD                = 139
	Op_1SUB                = 140
	Op_2MUL                = 141
	Op_2DIV                = 142
	Op_NEGATE              = 143
	Op_ABS                 = 144
	Op_NOT                 = 145
	Op_0NOTEQUAL           = 146
	Op_ADD                 = 147
	Op_SUB                 = 148
	Op_MUL                 = 149
	Op_DIV                 = 150
	Op_MOD                 = 151
	Op_LSHIFT              = 152
	Op_RSHIFT              = 153
	Op_BOOLAND             = 154
	Op_BOOLOR              = 155
	Op_NUMEQUAL            = 156
	Op_NUMEQUALVERIFY      = 157
	Op_NUMNOTEQUAL         = 158
	Op_LESSTHAN            = 159
	Op_GREATERTHAN         = 160
	Op_LESSTHANOREQUAL     = 161
	Op_GREATERTHANOREQUAL  = 162
	Op_MIN                 = 163
	Op_MAX                 = 164
	Op_WITHIN              = 165
	Op_RIPEMD160           = 166
	Op_SHA1                = 167
	Op_SHA256              = 168
	Op_HASH160             = 169
	Op_HASH256             = 170
	Op_CODESEPARATOR       = 171
	Op_CHECKSIG            = 172
	Op_CHECKSIGVERIFY      = 173
	Op_CHECKMULTISIG       = 174
	Op_CHECKMULTISIGVERIFY = 175
	Op_NOP1                = 176
	Op_CHECKLOCKTIMEVERIFY = 177
	Op_CHECKSEQUENCEVERIFY = 178
	Op_NOP4                = 179
	Op_NOP5                = 180
	Op_NOP6                = 181
	Op_NOP7                = 182
	Op_NOP8                = 183
	Op_NOP9                = 184
	Op_NOP10               = 185
	Op_PUBKEYHASH          = 253
	Op_PUBKEY              = 254
	Op_INVALIDOPCODE       = 255
)

// OpCode describes a Bitcoin script opcode with a symbolic name and a value.
type OpCode struct {
	// Name is the mnemonic name of the opcode.
	Name string
	// Value is the byte code of the opcode.
	Value byte
	// Exec is function that performs the stack operations for the opcode.
	Exec func(r *R) int
}

var (
	// OpCodes is a list of all valid opcodes in a Bitcoin script.
	OpCodes = []*OpCode{
		{"OP_FALSE", Op_FALSE, func(r *R) int {
			return r.stack.Push(0)
		}},
		{"OP_PUSHDATA1", Op_PUSHDATA1, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OP_PUSHDATA2", Op_PUSHDATA2, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OP_PUSHDATA4", Op_PUSHDATA4, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OP_1NEGATE", Op_1NEGATE, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OP_RESERVED", Op_RESERVED, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_TRUE", Op_TRUE, func(r *R) int {
			return r.stack.Push(1)
		}},
		{"OP_2", Op_2, func(r *R) int {
			return r.stack.Push(2)
		}},
		{"OP_3", Op_3, func(r *R) int {
			return r.stack.Push(3)
		}},
		{"OP_4", Op_4, func(r *R) int {
			return r.stack.Push(4)
		}},
		{"OP_5", Op_5, func(r *R) int {
			return r.stack.Push(5)
		}},
		{"OP_6", Op_6, func(r *R) int {
			return r.stack.Push(6)
		}},
		{"OP_7", Op_7, func(r *R) int {
			return r.stack.Push(7)
		}},
		{"OP_8", Op_8, func(r *R) int {
			return r.stack.Push(8)
		}},
		{"OP_9", Op_9, func(r *R) int {
			return r.stack.Push(9)
		}},
		{"OP_10", Op_10, func(r *R) int {
			return r.stack.Push(10)
		}},
		{"OP_11", Op_11, func(r *R) int {
			return r.stack.Push(11)
		}},
		{"OP_12", Op_12, func(r *R) int {
			return r.stack.Push(12)
		}},
		{"OP_13", Op_13, func(r *R) int {
			return r.stack.Push(13)
		}},
		{"OP_14", Op_14, func(r *R) int {
			return r.stack.Push(14)
		}},
		{"OP_15", Op_15, func(r *R) int {
			return r.stack.Push(15)
		}},
		{"OP_16", Op_16, func(r *R) int {
			return r.stack.Push(16)
		}},
		{"OP_NOP", Op_NOP, func(r *R) int {
			return RcOK
		}},
		{"OP_VER", Op_VER, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_IF", Op_IF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ONE) {
				s := len(r.stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.stmts[r.pos].Opcode {
					case Op_IF:
						depth++
					case Op_NOTIF:
						depth++
					case Op_ELSE:
						if depth == 0 {
							return RcOK
						}
					case Op_ENDIF:
						if depth == 0 {
							return RcOK
						}
						depth--
					}
				}
				return RcUnclosedIf
			}
			return RcOK
		}},
		{"OP_NOTIF", Op_NOTIF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				s := len(r.stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.stmts[r.pos].Opcode {
					case Op_IF:
						depth++
					case Op_NOTIF:
						depth++
					case Op_ELSE:
						if depth == 0 {
							return RcOK
						}
					case Op_ENDIF:
						if depth == 0 {
							return RcOK
						}
						depth--
					}
				}
				return RcUnclosedIf
			}
			return RcOK
		}},
		{"OP_VERIF", Op_VERIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_VERNOTIF", Op_VERNOTIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_ELSE", Op_ELSE, func(r *R) int {
			s := len(r.stmts)
			depth := 0
			for r.pos++; r.pos < s; r.pos++ {
				switch r.stmts[r.pos].Opcode {
				case Op_IF:
					depth++
				case Op_NOTIF:
					depth++
				case Op_ELSE:
					if depth == 0 {
						return RcDoubleElse
					}
				case Op_ENDIF:
					if depth == 0 {
						return RcOK
					}
					depth--
				}
			}
			return RcUnclosedIf
		}},
		{"OP_ENDIF", Op_ENDIF, func(r *R) int {
			return RcOK
		}},
		{"OP_VERIFY", Op_VERIFY, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				return RcOK
			}
			return RcTxInvalid
		}},
		{"OP_RETURN", Op_RETURN, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_TOALTSTACK", Op_TOALTSTACK, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.altStack.Push(v)
		}},
		{"OP_FROMALTSTACK", Op_FROMALTSTACK, func(r *R) int {
			v, rc := r.altStack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_2DROP", Op_2DROP, func(r *R) int {
			if _, rc := r.stack.Pop(); rc != RcOK {
				return rc
			}
			_, rc := r.stack.Pop()
			return rc
		}},
		{"OP_2DUP", Op_2DUP, func(r *R) int {
			return r.stack.Dup(2)
		}},
		{"OP_3DUP", Op_3DUP, func(r *R) int {
			return r.stack.Dup(3)
		}},
		{"OP_2OVER", Op_2OVER, func(r *R) int {
			return RcOK
		}},
		{"OP_2ROT", Op_2ROT, func(r *R) int {
			return RcOK
		}},
		{"OP_2SWAP", Op_2SWAP, func(r *R) int {
			return RcOK
		}},
		{"OP_IFDUP", Op_IFDUP, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ZERO) {
				r.stack.Push(v)
			}
			return RcOK
		}},
		{"OP_DEPTH", Op_DEPTH, func(r *R) int {
			return r.stack.Push(r.stack.Len())
		}},
		{"OP_DROP", Op_DROP, func(r *R) int {
			_, rc := r.stack.Pop()
			return rc
		}},
		{"OP_DUP", Op_DUP, func(r *R) int {
			return r.stack.Dup(1)
		}},
		{"OP_NIP", Op_NIP, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if _, rc = r.stack.Pop(); rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_OVER", Op_OVER, func(r *R) int {
			v, rc := r.stack.PeekAt(1)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_PICK", Op_PICK, func(r *R) int {
			n, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v, rc := r.stack.PeekAt(int(n.Int64()))
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_ROLL", Op_ROLL, func(r *R) int {
			n, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v, rc := r.stack.RemoveAt(int(n.Int64()))
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_ROT", Op_ROT, func(r *R) int {
			v, rc := r.stack.RemoveAt(2)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_SWAP", Op_SWAP, func(r *R) int {
			v1, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v2, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if rc = r.stack.Push(v1); rc != RcOK {
				return rc
			}
			return r.stack.Push(v2)
		}},
		{"OP_TUCK", Op_TUCK, func(r *R) int {
			v1, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v2, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if rc = r.stack.Push(v2); rc != RcOK {
				return rc
			}
			if rc = r.stack.Push(v1); rc != RcOK {
				return rc
			}
			return r.stack.Push(v2)
		}},
		{"OP_CAT", Op_CAT, func(r *R) int {
			return RcOK
		}},
		{"OP_SUBSTR", Op_SUBSTR, func(r *R) int {
			return RcOK
		}},
		{"OP_LEFT", Op_LEFT, func(r *R) int {
			return RcOK
		}},
		{"OP_RIGHT", Op_RIGHT, func(r *R) int {
			return RcOK
		}},
		{"OP_SIZE", Op_SIZE, func(r *R) int {
			return RcOK
		}},
		{"OP_INVERT", Op_INVERT, func(r *R) int {
			return RcOK
		}},
		{"OP_AND", Op_AND, func(r *R) int {
			return RcOK
		}},
		{"OP_OR", Op_OR, func(r *R) int {
			return RcOK
		}},
		{"OP_XOR", Op_XOR, func(r *R) int {
			return RcOK
		}},
		{"OP_EQUAL", Op_EQUAL, func(r *R) int {
			v1, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v2, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v1.Equals(v2) {
				return r.stack.Push(1)
			}
			return r.stack.Push(0)
		}},
		{"OP_EQUALVERIFY", Op_EQUALVERIFY, func(r *R) int {
			v1, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			v2, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v1.Equals(v2) {
				return RcOK
			}
			return RcTxInvalid
		}},
		{"OP_RESERVED1", Op_RESERVED1, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_RESERVED2", Op_RESERVED2, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_1ADD", Op_1ADD, func(r *R) int {
			return RcOK
		}},
		{"OP_1SUB", Op_1SUB, func(r *R) int {
			return RcOK
		}},
		{"OP_2MUL", Op_2MUL, func(r *R) int {
			return RcOK
		}},
		{"OP_2DIV", Op_2DIV, func(r *R) int {
			return RcOK
		}},
		{"OP_NEGATE", Op_NEGATE, func(r *R) int {
			return RcOK
		}},
		{"OP_ABS", Op_ABS, func(r *R) int {
			return RcOK
		}},
		{"OP_NOT", Op_NOT, func(r *R) int {
			return RcOK
		}},
		{"OP_0NOTEQUAL", Op_0NOTEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OP_ADD", Op_ADD, func(r *R) int {
			return RcOK
		}},
		{"OP_SUB", Op_SUB, func(r *R) int {
			return RcOK
		}},
		{"OP_MUL", Op_MUL, func(r *R) int {
			return RcOK
		}},
		{"OP_DIV", Op_DIV, func(r *R) int {
			return RcOK
		}},
		{"OP_MOD", Op_MOD, func(r *R) int {
			return RcOK
		}},
		{"OP_LSHIFT", Op_LSHIFT, func(r *R) int {
			return RcOK
		}},
		{"OP_RSHIFT", Op_RSHIFT, func(r *R) int {
			return RcOK
		}},
		{"OP_BOOLAND", Op_BOOLAND, func(r *R) int {
			return RcOK
		}},
		{"OP_BOOLOR", Op_BOOLOR, func(r *R) int {
			return RcOK
		}},
		{"OP_NUMEQUAL", Op_NUMEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OP_NUMEQUALVERIFY", Op_NUMEQUALVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OP_NUMNOTEQUAL", Op_NUMNOTEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OP_LESSTHAN", Op_LESSTHAN, func(r *R) int {
			return RcOK
		}},
		{"OP_GREATERTHAN", Op_GREATERTHAN, func(r *R) int {
			return RcOK
		}},
		{"OP_LESSTHANOREQUAL", Op_LESSTHANOREQUAL, func(r *R) int {
			return RcOK
		}},
		{"OP_GREATERTHANOREQUAL", Op_GREATERTHANOREQUAL, func(r *R) int {
			return RcOK
		}},
		{"OP_MIN", Op_MIN, func(r *R) int {
			return RcOK
		}},
		{"OP_MAX", Op_MAX, func(r *R) int {
			return RcOK
		}},
		{"OP_WITHIN", Op_WITHIN, func(r *R) int {
			return RcOK
		}},
		{"OP_RIPEMD160", Op_RIPEMD160, func(r *R) int {
			return RcOK
		}},
		{"OP_SHA1", Op_SHA1, func(r *R) int {
			return RcOK
		}},
		{"OP_SHA256", Op_SHA256, func(r *R) int {
			return RcOK
		}},
		{"OP_HASH160", Op_HASH160, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(util.Hash160(v.Bytes()))
		}},
		{"OP_HASH256", Op_HASH256, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(util.Hash256(v.Bytes()))
		}},
		{"OP_CODESEPARATOR", Op_CODESEPARATOR, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKSIG", Op_CHECKSIG, func(r *R) int {
			valid, rc := r.CheckSig()
			if rc != RcOK {
				return rc
			}
			if valid {
				r.stack.Push(1)
			} else {
				r.stack.Push(0)
			}
			return RcOK
		}},
		{"OP_CHECKSIGVERIFY", Op_CHECKSIGVERIFY, func(r *R) int {
			valid, rc := r.CheckSig()
			if rc != RcOK {
				return rc
			}
			if valid {
				return RcOK
			}
			return RcInvalidTransfer
		}},
		{"OP_CHECKMULTISIG", Op_CHECKMULTISIG, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKMULTISIGVERIFY", Op_CHECKMULTISIGVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP1", Op_NOP1, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKLOCKTIMEVERIFY", Op_CHECKLOCKTIMEVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKSEQUENCEVERIFY", Op_CHECKSEQUENCEVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP4", Op_NOP4, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP5", Op_NOP5, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP6", Op_NOP6, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP7", Op_NOP7, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP8", Op_NOP8, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP9", Op_NOP9, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP10", Op_NOP10, func(r *R) int {
			return RcOK
		}},
		{"OP_PUBKEYHASH", Op_PUBKEYHASH, func(r *R) int {
			return RcOK
		}},
		{"OP_PUBKEY", Op_PUBKEY, func(r *R) int {
			return RcOK
		}},
		{"OP_INVALIDOPCODE", Op_INVALIDOPCODE, func(r *R) int {
			return RcInvalidOpcode
		}},
	}
)

// GetOpCode returns a opcode for a given byte value.
func GetOpcode(v byte) *OpCode {
	if v > 0 && v < 76 {
		return &OpCode{
			Name:  fmt.Sprintf("DATA_%d", int(v)),
			Value: v,
			Exec: func(r *R) int {
				return r.stack.Push(r.stmts[r.pos].Data)
			},
		}
	}
	for _, op := range OpCodes {
		if op.Value == v {
			return op
		}
	}
	return nil
}
