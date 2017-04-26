package script

import (
	"fmt"
	"github.com/bfix/gospel/bitcoin/util"
	"github.com/bfix/gospel/math"
)

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
		{"OpFALSE", OpFALSE, func(r *R) int {
			return r.stack.Push(0)
		}},
		{"OpPUSHDATA1", OpPUSHDATA1, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OpPUSHDATA2", OpPUSHDATA2, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"OpPUSHDATA4", OpPUSHDATA4, func(r *R) int {
			return r.stack.Push(r.stmts[r.pos].Data)
		}},
		{"Op1NEGATE", Op1NEGATE, func(r *R) int {
			return r.stack.Push(-1)
		}},
		{"OpRESERVED", OpRESERVED, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpTRUE", OpTRUE, func(r *R) int {
			return r.stack.Push(1)
		}},
		{"Op2", Op2, func(r *R) int {
			return r.stack.Push(2)
		}},
		{"Op3", Op3, func(r *R) int {
			return r.stack.Push(3)
		}},
		{"Op4", Op4, func(r *R) int {
			return r.stack.Push(4)
		}},
		{"Op5", Op5, func(r *R) int {
			return r.stack.Push(5)
		}},
		{"Op6", Op6, func(r *R) int {
			return r.stack.Push(6)
		}},
		{"Op7", Op7, func(r *R) int {
			return r.stack.Push(7)
		}},
		{"Op8", Op8, func(r *R) int {
			return r.stack.Push(8)
		}},
		{"Op9", Op9, func(r *R) int {
			return r.stack.Push(9)
		}},
		{"Op10", Op10, func(r *R) int {
			return r.stack.Push(10)
		}},
		{"Op11", Op11, func(r *R) int {
			return r.stack.Push(11)
		}},
		{"Op12", Op12, func(r *R) int {
			return r.stack.Push(12)
		}},
		{"Op13", Op13, func(r *R) int {
			return r.stack.Push(13)
		}},
		{"Op14", Op14, func(r *R) int {
			return r.stack.Push(14)
		}},
		{"Op15", Op15, func(r *R) int {
			return r.stack.Push(15)
		}},
		{"Op16", Op16, func(r *R) int {
			return r.stack.Push(16)
		}},
		{"OpNOP", OpNOP, func(r *R) int {
			return RcOK
		}},
		{"OpVER", OpVER, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpIF", OpIF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ONE) {
				s := len(r.stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.stmts[r.pos].Opcode {
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
		{"OpNOTIF", OpNOTIF, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				s := len(r.stmts)
				depth := 0
				for r.pos++; r.pos < s; r.pos++ {
					switch r.stmts[r.pos].Opcode {
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
		{"OpVERIF", OpVERIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpVERNOTIF", OpVERNOTIF, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpELSE", OpELSE, func(r *R) int {
			s := len(r.stmts)
			depth := 0
			for r.pos++; r.pos < s; r.pos++ {
				switch r.stmts[r.pos].Opcode {
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
		{"OpENDIF", OpENDIF, func(r *R) int {
			return RcOK
		}},
		{"OpVERIFY", OpVERIFY, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if v.Equals(math.ONE) {
				return RcOK
			}
			return RcTxInvalid
		}},
		{"OpRETURN", OpRETURN, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpTOALTSTACK", OpTOALTSTACK, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.altStack.Push(v)
		}},
		{"OpFROMALTSTACK", OpFROMALTSTACK, func(r *R) int {
			v, rc := r.altStack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"Op2DROP", Op2DROP, func(r *R) int {
			if _, rc := r.stack.Pop(); rc != RcOK {
				return rc
			}
			_, rc := r.stack.Pop()
			return rc
		}},
		{"Op2DUP", Op2DUP, func(r *R) int {
			return r.stack.Dup(2)
		}},
		{"Op3DUP", Op3DUP, func(r *R) int {
			return r.stack.Dup(3)
		}},
		{"Op2OVER", Op2OVER, func(r *R) int {
			return RcOK
		}},
		{"Op2ROT", Op2ROT, func(r *R) int {
			return RcOK
		}},
		{"Op2SWAP", Op2SWAP, func(r *R) int {
			return RcOK
		}},
		{"OpIFDUP", OpIFDUP, func(r *R) int {
			v, rc := r.stack.Peek()
			if rc != RcOK {
				return rc
			}
			if !v.Equals(math.ZERO) {
				r.stack.Push(v)
			}
			return RcOK
		}},
		{"OpDEPTH", OpDEPTH, func(r *R) int {
			return r.stack.Push(r.stack.Len())
		}},
		{"OpDROP", OpDROP, func(r *R) int {
			_, rc := r.stack.Pop()
			return rc
		}},
		{"OpDUP", OpDUP, func(r *R) int {
			return r.stack.Dup(1)
		}},
		{"OpNIP", OpNIP, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			if _, rc = r.stack.Pop(); rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OpOVER", OpOVER, func(r *R) int {
			v, rc := r.stack.PeekAt(1)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OpPICK", OpPICK, func(r *R) int {
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
		{"OpROLL", OpROLL, func(r *R) int {
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
		{"OpROT", OpROT, func(r *R) int {
			v, rc := r.stack.RemoveAt(2)
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(v)
		}},
		{"OpSWAP", OpSWAP, func(r *R) int {
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
		{"OpTUCK", OpTUCK, func(r *R) int {
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
		{"OpCAT", OpCAT, func(r *R) int {
			return RcOK
		}},
		{"OpSUBSTR", OpSUBSTR, func(r *R) int {
			return RcOK
		}},
		{"OpLEFT", OpLEFT, func(r *R) int {
			return RcOK
		}},
		{"OpRIGHT", OpRIGHT, func(r *R) int {
			return RcOK
		}},
		{"OpSIZE", OpSIZE, func(r *R) int {
			return RcOK
		}},
		{"OpINVERT", OpINVERT, func(r *R) int {
			return RcOK
		}},
		{"OpAND", OpAND, func(r *R) int {
			return RcOK
		}},
		{"OpOR", OpOR, func(r *R) int {
			return RcOK
		}},
		{"OpXOR", OpXOR, func(r *R) int {
			return RcOK
		}},
		{"OpEQUAL", OpEQUAL, func(r *R) int {
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
		{"OpEQUALVERIFY", OpEQUALVERIFY, func(r *R) int {
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
		{"OpRESERVED1", OpRESERVED1, func(r *R) int {
			return RcTxInvalid
		}},
		{"OpRESERVED2", OpRESERVED2, func(r *R) int {
			return RcTxInvalid
		}},
		{"Op1ADD", Op1ADD, func(r *R) int {
			return RcOK
		}},
		{"Op1SUB", Op1SUB, func(r *R) int {
			return RcOK
		}},
		{"Op2MUL", Op2MUL, func(r *R) int {
			return RcOK
		}},
		{"Op2DIV", Op2DIV, func(r *R) int {
			return RcOK
		}},
		{"OpNEGATE", OpNEGATE, func(r *R) int {
			return RcOK
		}},
		{"OpABS", OpABS, func(r *R) int {
			return RcOK
		}},
		{"OpNOT", OpNOT, func(r *R) int {
			return RcOK
		}},
		{"Op0NOTEQUAL", Op0NOTEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OpADD", OpADD, func(r *R) int {
			return RcOK
		}},
		{"OpSUB", OpSUB, func(r *R) int {
			return RcOK
		}},
		{"OpMUL", OpMUL, func(r *R) int {
			return RcOK
		}},
		{"OpDIV", OpDIV, func(r *R) int {
			return RcOK
		}},
		{"OpMOD", OpMOD, func(r *R) int {
			return RcOK
		}},
		{"OpLSHIFT", OpLSHIFT, func(r *R) int {
			return RcOK
		}},
		{"OpRSHIFT", OpRSHIFT, func(r *R) int {
			return RcOK
		}},
		{"OpBOOLAND", OpBOOLAND, func(r *R) int {
			return RcOK
		}},
		{"OpBOOLOR", OpBOOLOR, func(r *R) int {
			return RcOK
		}},
		{"OpNUMEQUAL", OpNUMEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OpNUMEQUALVERIFY", OpNUMEQUALVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OpNUMNOTEQUAL", OpNUMNOTEQUAL, func(r *R) int {
			return RcOK
		}},
		{"OpLESSTHAN", OpLESSTHAN, func(r *R) int {
			return RcOK
		}},
		{"OpGREATERTHAN", OpGREATERTHAN, func(r *R) int {
			return RcOK
		}},
		{"OpLESSTHANOREQUAL", OpLESSTHANOREQUAL, func(r *R) int {
			return RcOK
		}},
		{"OpGREATERTHANOREQUAL", OpGREATERTHANOREQUAL, func(r *R) int {
			return RcOK
		}},
		{"OpMIN", OpMIN, func(r *R) int {
			return RcOK
		}},
		{"OpMAX", OpMAX, func(r *R) int {
			return RcOK
		}},
		{"OpWITHIN", OpWITHIN, func(r *R) int {
			return RcOK
		}},
		{"OpRIPEMD160", OpRIPEMD160, func(r *R) int {
			return RcOK
		}},
		{"OpSHA1", OpSHA1, func(r *R) int {
			return RcOK
		}},
		{"OpSHA256", OpSHA256, func(r *R) int {
			return RcOK
		}},
		{"OpHASH160", OpHASH160, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(util.Hash160(v.Bytes()))
		}},
		{"OpHASH256", OpHASH256, func(r *R) int {
			v, rc := r.stack.Pop()
			if rc != RcOK {
				return rc
			}
			return r.stack.Push(util.Hash256(v.Bytes()))
		}},
		{"OpCODESEPARATOR", OpCODESEPARATOR, func(r *R) int {
			return RcOK
		}},
		{"OpCHECKSIG", OpCHECKSIG, func(r *R) int {
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
		{"OpCHECKSIGVERIFY", OpCHECKSIGVERIFY, func(r *R) int {
			valid, rc := r.CheckSig()
			if rc != RcOK {
				return rc
			}
			if valid {
				return RcOK
			}
			return RcInvalidTransfer
		}},
		{"OpCHECKMULTISIG", OpCHECKMULTISIG, func(r *R) int {
			return RcOK
		}},
		{"OpCHECKMULTISIGVERIFY", OpCHECKMULTISIGVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OpNOP1", OpNOP1, func(r *R) int {
			return RcOK
		}},
		{"OpCHECKLOCKTIMEVERIFY", OpCHECKLOCKTIMEVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OpCHECKSEQUENCEVERIFY", OpCHECKSEQUENCEVERIFY, func(r *R) int {
			return RcOK
		}},
		{"OpNOP4", OpNOP4, func(r *R) int {
			return RcOK
		}},
		{"OpNOP5", OpNOP5, func(r *R) int {
			return RcOK
		}},
		{"OpNOP6", OpNOP6, func(r *R) int {
			return RcOK
		}},
		{"OpNOP7", OpNOP7, func(r *R) int {
			return RcOK
		}},
		{"OpNOP8", OpNOP8, func(r *R) int {
			return RcOK
		}},
		{"OpNOP9", OpNOP9, func(r *R) int {
			return RcOK
		}},
		{"OpNOP10", OpNOP10, func(r *R) int {
			return RcOK
		}},
		{"OpPUBKEYHASH", OpPUBKEYHASH, func(r *R) int {
			return RcOK
		}},
		{"OpPUBKEY", OpPUBKEY, func(r *R) int {
			return RcOK
		}},
		{"OpINVALIDOPCODE", OpINVALIDOPCODE, func(r *R) int {
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
