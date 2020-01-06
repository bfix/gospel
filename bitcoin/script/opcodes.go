package script

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2019 Bernd Fix
//
// Gospel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gospel.  If not, see <http://www.gnu.org/licenses/>.
//----------------------------------------------------------------------

import (
	"fmt"

	"github.com/bfix/gospel/bitcoin"
	"github.com/bfix/gospel/math"
)

// Bitcoin script opcodes
const (
	OpFALSE               = 0
	OpPUSHDATA1           = 76
	OpPUSHDATA2           = 77
	OpPUSHDATA4           = 78
	Op1NEGATE             = 79
	OpRESERVED            = 80
	OpTRUE                = 81
	Op2                   = 82
	Op3                   = 83
	Op4                   = 84
	Op5                   = 85
	Op6                   = 86
	Op7                   = 87
	Op8                   = 88
	Op9                   = 89
	Op10                  = 90
	Op11                  = 91
	Op12                  = 92
	Op13                  = 93
	Op14                  = 94
	Op15                  = 95
	Op16                  = 96
	OpNOP                 = 97
	OpVER                 = 98
	OpIF                  = 99
	OpNOTIF               = 100
	OpVERIF               = 101
	OpVERNOTIF            = 102
	OpELSE                = 103
	OpENDIF               = 104
	OpVERIFY              = 105
	OpRETURN              = 106
	OpTOALTSTACK          = 107
	OpFROMALTSTACK        = 108
	Op2DROP               = 109
	Op2DUP                = 110
	Op3DUP                = 111
	Op2OVER               = 112
	Op2ROT                = 113
	Op2SWAP               = 114
	OpIFDUP               = 115
	OpDEPTH               = 116
	OpDROP                = 117
	OpDUP                 = 118
	OpNIP                 = 119
	OpOVER                = 120
	OpPICK                = 121
	OpROLL                = 122
	OpROT                 = 123
	OpSWAP                = 124
	OpTUCK                = 125
	OpCAT                 = 126
	OpSUBSTR              = 127
	OpLEFT                = 128
	OpRIGHT               = 129
	OpSIZE                = 130
	OpINVERT              = 131
	OpAND                 = 132
	OpOR                  = 133
	OpXOR                 = 134
	OpEQUAL               = 135
	OpEQUALVERIFY         = 136
	OpRESERVED1           = 137
	OpRESERVED2           = 138
	Op1ADD                = 139
	Op1SUB                = 140
	Op2MUL                = 141
	Op2DIV                = 142
	OpNEGATE              = 143
	OpABS                 = 144
	OpNOT                 = 145
	Op0NOTEQUAL           = 146
	OpADD                 = 147
	OpSUB                 = 148
	OpMUL                 = 149
	OpDIV                 = 150
	OpMOD                 = 151
	OpLSHIFT              = 152
	OpRSHIFT              = 153
	OpBOOLAND             = 154
	OpBOOLOR              = 155
	OpNUMEQUAL            = 156
	OpNUMEQUALVERIFY      = 157
	OpNUMNOTEQUAL         = 158
	OpLESSTHAN            = 159
	OpGREATERTHAN         = 160
	OpLESSTHANOREQUAL     = 161
	OpGREATERTHANOREQUAL  = 162
	OpMIN                 = 163
	OpMAX                 = 164
	OpWITHIN              = 165
	OpRIPEMD160           = 166
	OpSHA1                = 167
	OpSHA256              = 168
	OpHASH160             = 169
	OpHASH256             = 170
	OpCODESEPARATOR       = 171
	OpCHECKSIG            = 172
	OpCHECKSIGVERIFY      = 173
	OpCHECKMULTISIG       = 174
	OpCHECKMULTISIGVERIFY = 175
	OpNOP1                = 176
	OpCHECKLOCKTIMEVERIFY = 177
	OpCHECKSEQUENCEVERIFY = 178
	OpNOP4                = 179
	OpNOP5                = 180
	OpNOP6                = 181
	OpNOP7                = 182
	OpNOP8                = 183
	OpNOP9                = 184
	OpNOP10               = 185
	OpPUBKEYHASH          = 253
	OpPUBKEY              = 254
	OpINVALIDOPCODE       = 255
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
		{"OP_FALSE", OpFALSE, func(r *R) int {
			return r.stack.Push(0)
		}},
		{"OP_PUSHDATA1", OpPUSHDATA1, func(r *R) int {
			return r.stack.Push(r.script.Stmts[r.pos].Data)
		}},
		{"OP_PUSHDATA2", OpPUSHDATA2, func(r *R) int {
			return r.stack.Push(r.script.Stmts[r.pos].Data)
		}},
		{"OP_PUSHDATA4", OpPUSHDATA4, func(r *R) int {
			return r.stack.Push(r.script.Stmts[r.pos].Data)
		}},
		{"OP_1NEGATE", Op1NEGATE, func(r *R) int {
			return r.stack.Push(-1)
		}},
		{"OP_RESERVED", OpRESERVED, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_TRUE", OpTRUE, func(r *R) int {
			return r.stack.Push(1)
		}},
		{"OP_2", Op2, func(r *R) int {
			return r.stack.Push(2)
		}},
		{"OP_3", Op3, func(r *R) int {
			return r.stack.Push(3)
		}},
		{"OP_4", Op4, func(r *R) int {
			return r.stack.Push(4)
		}},
		{"OP_5", Op5, func(r *R) int {
			return r.stack.Push(5)
		}},
		{"OP_6", Op6, func(r *R) int {
			return r.stack.Push(6)
		}},
		{"OP_7", Op7, func(r *R) int {
			return r.stack.Push(7)
		}},
		{"OP_8", Op8, func(r *R) int {
			return r.stack.Push(8)
		}},
		{"OP_9", Op9, func(r *R) int {
			return r.stack.Push(9)
		}},
		{"OP_10", Op10, func(r *R) int {
			return r.stack.Push(10)
		}},
		{"OP_11", Op11, func(r *R) int {
			return r.stack.Push(11)
		}},
		{"OP_12", Op12, func(r *R) int {
			return r.stack.Push(12)
		}},
		{"OP_13", Op13, func(r *R) int {
			return r.stack.Push(13)
		}},
		{"OP_14", Op14, func(r *R) int {
			return r.stack.Push(14)
		}},
		{"OP_15", Op15, func(r *R) int {
			return r.stack.Push(15)
		}},
		{"OP_16", Op16, func(r *R) int {
			return r.stack.Push(16)
		}},
		{"OP_NOP", OpNOP, func(r *R) int {
			return RcOK
		}},
		{"OP_VER", OpVER, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_IF", OpIF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ONE) {
				s := len(r.script.Stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.script.Stmts[r.pos].Opcode {
					case OpIF:
						depth++
					case OpNOTIF:
						depth++
					case OpELSE:
						if depth == 0 {
							return RcOK
						}
					case OpENDIF:
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
		{"OP_NOTIF", OpNOTIF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				s := len(r.script.Stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.script.Stmts[r.pos].Opcode {
					case OpIF:
						depth++
					case OpNOTIF:
						depth++
					case OpELSE:
						if depth == 0 {
							return RcOK
						}
					case OpENDIF:
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
		{"OP_VERIF", OpVERIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_VERNOTIF", OpVERNOTIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_ELSE", OpELSE, func(r *R) int {
			s := len(r.script.Stmts)
			depth := 0
			for r.pos++; r.pos < s; r.pos++ {
				switch r.script.Stmts[r.pos].Opcode {
				case OpIF:
					depth++
				case OpNOTIF:
					depth++
				case OpELSE:
					if depth == 0 {
						return RcDoubleElse
					}
				case OpENDIF:
					if depth == 0 {
						return RcOK
					}
					depth--
				}
			}
			return RcUnclosedIf
		}},
		{"OP_ENDIF", OpENDIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_VERIFY", OpVERIFY, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				return RcOK
			}
			return RcNotVerified
		}},
		{"OP_RETURN", OpRETURN, func(r *R) int {
			return RcNotVerified
		}},
		{"OP_TOALTSTACK", OpTOALTSTACK, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.altStack.Push(v)
		}},
		{"OP_FROMALTSTACK", OpFROMALTSTACK, func(r *R) int {
			v, rc := r.altStack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_2DROP", Op2DROP, func(r *R) int {
			if _, rc := r.stack.Pop(); rc != RcOK {
				return rc
			}
			_, rc := r.stack.Pop()
			return rc
		}},
		{"OP_2DUP", Op2DUP, func(r *R) int {
			return r.stack.Dup(2)
		}},
		{"OP_3DUP", Op3DUP, func(r *R) int {
			return r.stack.Dup(3)
		}},
		{"OP_2OVER", Op2OVER, func(r *R) int {
			v, rc := r.stack.PeekAt(3)
			if rc != RcOK {
				return rc
			}
			if rc := r.stack.Push(v); rc != RcOK {
				return rc
			}
			v, rc = r.stack.PeekAt(3)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_2ROT", Op2ROT, func(r *R) int {
			v, rc := r.stack.RemoveAt(5)
			if rc != RcOK {
				return rc
			}
			if rc = r.stack.Push(v); rc != RcOK {
				return rc
			}
			v, rc = r.stack.RemoveAt(5)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_2SWAP", Op2SWAP, func(r *R) int {
			v, rc := r.stack.RemoveAt(3)
			if rc != RcOK {
				return rc
			}
			if rc := r.stack.Push(v); rc != RcOK {
				return rc
			}
			v, rc = r.stack.RemoveAt(3)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_IFDUP", OpIFDUP, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ZERO) {
				r.stack.Push(v)
			}
			return RcOK
		}},
		{"OP_DEPTH", OpDEPTH, func(r *R) int {
			return r.stack.Push(r.stack.Len())
		}},
		{"OP_DROP", OpDROP, func(r *R) int {
			_, rc := r.stack.Pop()
			return rc
		}},
		{"OP_DUP", OpDUP, func(r *R) int {
			return r.stack.Dup(1)
		}},
		{"OP_NIP", OpNIP, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if _, rc = r.stack.Pop(); rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_OVER", OpOVER, func(r *R) int {
			v, rc := r.stack.PeekAt(1)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_PICK", OpPICK, func(r *R) int {
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
		{"OP_ROLL", OpROLL, func(r *R) int {
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
		{"OP_ROT", OpROT, func(r *R) int {
			v, rc := r.stack.RemoveAt(2)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OP_SWAP", OpSWAP, func(r *R) int {
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
		{"OP_TUCK", OpTUCK, func(r *R) int {
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
		{"OP_CAT", OpCAT, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_SUBSTR", OpSUBSTR, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_LEFT", OpLEFT, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_RIGHT", OpRIGHT, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_SIZE", OpSIZE, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(len(v.Bytes()))
		}},
		{"OP_INVERT", OpINVERT, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_AND", OpAND, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_OR", OpOR, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_XOR", OpXOR, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_EQUAL", OpEQUAL, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp == 0)
		}},
		{"OP_EQUALVERIFY", OpEQUALVERIFY, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			if cmp == 0 {
				return RcOK
			}
			return RcTxInvalid
		}},
		{"OP_RESERVED1", OpRESERVED1, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_RESERVED2", OpRESERVED2, func(r *R) int {
			return RcTxInvalid
		}},
		{"OP_1ADD", Op1ADD, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v.Add(math.ONE))
		}},
		{"OP_1SUB", Op1SUB, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v.Sub(math.ONE))
		}},
		{"OP_2MUL", Op2MUL, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_2DIV", Op2DIV, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_NEGATE", OpNEGATE, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v.Neg())
		}},
		{"OP_ABS", OpABS, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v.Abs())
		}},
		{"OP_NOT", OpNOT, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v.Equals(math.ZERO))
		}},
		{"OP_0NOTEQUAL", Op0NOTEQUAL, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(!v.Equals(math.ZERO))
		}},
		{"OP_ADD", OpADD, func(r *R) int {
			b, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			a, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(a.Add(b))
		}},
		{"OP_SUB", OpSUB, func(r *R) int {
			b, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			a, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(a.Sub(b))
		}},
		{"OP_MUL", OpMUL, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_DIV", OpDIV, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_MOD", OpMOD, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_LSHIFT", OpLSHIFT, func(r *R) int {
			return RcDisabledOpcode
		}},
		{"OP_RSHIFT", OpRSHIFT, func(r *R) int {
			return RcOK
		}},
		{"OP_BOOLAND", OpBOOLAND, func(r *R) int {
			a, b, _, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(!a.Equals(math.ZERO) && !b.Equals(math.ZERO))
		}},
		{"OP_BOOLOR", OpBOOLOR, func(r *R) int {
			a, b, _, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(!a.Equals(math.ZERO) || !b.Equals(math.ZERO))
		}},
		{"OP_NUMEQUAL", OpNUMEQUAL, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp == 0)
		}},
		{"OP_NUMEQUALVERIFY", OpNUMEQUALVERIFY, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			if cmp == 0 {
				return RcOK
			}
			return RcTxInvalid
		}},
		{"OP_NUMNOTEQUAL", OpNUMNOTEQUAL, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp != 0)
		}},
		{"OP_LESSTHAN", OpLESSTHAN, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp < 0)
		}},
		{"OP_GREATERTHAN", OpGREATERTHAN, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp > 0)
		}},
		{"OP_LESSTHANOREQUAL", OpLESSTHANOREQUAL, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp <= 0)
		}},
		{"OP_GREATERTHANOREQUAL", OpGREATERTHANOREQUAL, func(r *R) int {
			_, _, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(cmp >= 0)
		}},
		{"OP_MIN", OpMIN, func(r *R) int {
			a, b, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			if cmp < 0 {
				return r.stack.Push(a)
			}
			return r.stack.Push(b)
		}},
		{"OP_MAX", OpMAX, func(r *R) int {
			a, b, cmp, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			if cmp > 0 {
				return r.stack.Push(a)
			}
			return r.stack.Push(b)
		}},
		{"OP_WITHIN", OpWITHIN, func(r *R) int {
			a, b, _, rc := r.stack.Compare()
			if rc != RcOK {
				return rc
			}
			i, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(a.Cmp(i) <= 0 && i.Cmp(b) < 0)
		}},
		{"OP_RIPEMD160", OpRIPEMD160, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(bitcoin.RipeMD160(v.Bytes()))
		}},
		{"OP_SHA1", OpSHA1, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(bitcoin.Sha1(v.Bytes()))
		}},
		{"OP_SHA256", OpSHA256, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(bitcoin.Sha256(v.Bytes()))
		}},
		{"OP_HASH160", OpHASH160, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(bitcoin.Hash160(v.Bytes()))
		}},
		{"OP_HASH256", OpHASH256, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(bitcoin.Hash256(v.Bytes()))
		}},
		{"OP_CODESEPARATOR", OpCODESEPARATOR, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKSIG", OpCHECKSIG, func(r *R) int {
			valid, rc := r.CheckSig()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(valid)
		}},
		{"OP_CHECKSIGVERIFY", OpCHECKSIGVERIFY, func(r *R) int {
			valid, rc := r.CheckSig()
			if rc != RcOK {
				return rc
			}
			if valid {
				return RcOK
			}
			return RcInvalidTransfer
		}},
		{"OP_CHECKMULTISIG", OpCHECKMULTISIG, func(r *R) int {
			valid, rc := r.CheckMultiSig()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(valid)
		}},
		{"OP_CHECKMULTISIGVERIFY", OpCHECKMULTISIGVERIFY, func(r *R) int {
			valid, rc := r.CheckMultiSig()
			if rc != RcOK {
				return rc
			}
			if valid {
				return RcOK
			}
			return RcInvalidTransfer
		}},
		{"OP_NOP1", OpNOP1, func(r *R) int {
			return RcOK
		}},
		{"OP_CHECKLOCKTIMEVERIFY", OpCHECKLOCKTIMEVERIFY, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK || v.Sign() == -1 {
				return RcTxInvalid
			}
			vt := uint64(v.Int64())
			var bounds uint64 = 500000000
			if (vt < bounds && r.tx.LockTime > bounds) ||
				(vt > bounds && r.tx.LockTime < bounds) ||
				r.tx.VinSeq[r.tx.VinSlot] == 0xffffffff {
				return RcTxInvalid
			}
			return RcOK
		}},
		{"OP_CHECKSEQUENCEVERIFY", OpCHECKSEQUENCEVERIFY, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK || v.Sign() == -1 {
				return RcTxInvalid
			}
			vt := uint64(v.Int64())
			inSeq := r.tx.VinSeq[r.tx.VinSlot]
			if vt&(1<<31) == 0 {
				if r.tx.Version < 2 ||
					inSeq&(1<<31) != 0 ||
					vt > inSeq {
					return RcTxInvalid
				}
			}
			return RcOK
		}},
		{"OP_NOP4", OpNOP4, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP5", OpNOP5, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP6", OpNOP6, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP7", OpNOP7, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP8", OpNOP8, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP9", OpNOP9, func(r *R) int {
			return RcOK
		}},
		{"OP_NOP10", OpNOP10, func(r *R) int {
			return RcOK
		}},
		{"OP_PUBKEYHASH", OpPUBKEYHASH, func(r *R) int {
			return RcOK
		}},
		{"OP_PUBKEY", OpPUBKEY, func(r *R) int {
			return RcOK
		}},
		{"OP_INVALIDOPCODE", OpINVALIDOPCODE, func(r *R) int {
			return RcInvalidOpcode
		}},
	}
)

// GetOpcode returns a opcode for a given byte value.
func GetOpcode(v byte) *OpCode {
	if v > 0 && v < 76 {
		return &OpCode{
			Name:  fmt.Sprintf("DATA_%d", int(v)),
			Value: v,
			Exec: func(r *R) int {
				return r.stack.Push(r.script.Stmts[r.pos].Data)
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
