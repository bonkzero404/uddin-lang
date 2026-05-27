package interpreter

import (
	"fmt"
	"strings"
)

// OpCode identifies a VM instruction.
type OpCode uint8

const (
	// Load/Store
	OP_LOAD_CONST  OpCode = iota // Dst = fn.Constants[Src1<<8|Src2]
	OP_LOAD_VAR                  // Dst = locals[Src1]
	OP_STORE_VAR                 // locals[Dst] = regs[Src1]
	OP_LOAD_UPVAL                // Dst = closure.Upvalues[Src1]
	OP_STORE_UPVAL               // closure.Upvalues[Dst] = regs[Src1]

	// Generic arithmetic (Value operands)
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_POW
	OP_NEG // Dst = -regs[Src1]

	// Typed arithmetic (int operands, no boxing)
	OP_ADD_INT
	OP_SUB_INT
	OP_MUL_INT
	OP_DIV_INT
	OP_MOD_INT

	// Comparison
	OP_EQ
	OP_NEQ
	OP_LT
	OP_LTE
	OP_GT
	OP_GTE
	OP_IN

	// Logic
	OP_AND
	OP_OR
	OP_NOT
	OP_XOR

	// Control flow
	OP_JUMP       // pc += int16(Dst<<8|Src2)
	OP_JUMP_FALSE // if !IsTruthy(regs[Src1]): pc += int16(Dst<<8|Src2)
	OP_JUMP_TRUE  // if IsTruthy(regs[Src1]): pc += int16(Dst<<8|Src2)
	OP_RETURN     // return regs[Src1]

	// Collections
	OP_MAKE_ARRAY       // Dst = []Value{regs[Src1]..regs[Src1+Src2-1]}
	OP_MAKE_MAP         // Dst = map[string]Value built from Src2 pairs starting at Src1
	OP_SUBSCRIPT        // Dst = regs[Src1][regs[Src2]]
	OP_SET_INDEX        // regs[Dst][regs[Src1]] = regs[Src2]
	OP_ITER_NORMALIZE   // Dst = normalize Src1 to *[]Value (map→keys, string→chars, array→as-is)

	// Functions
	OP_MAKE_FUNC           // Dst = &vmFunction from fn.SubFunctions[Src1<<8|Src2]
	OP_CALL                // Dst = call(regs[Src1], argc=Src2, args at argBase)
	OP_CALL_BUILTIN        // Dst = builtinTable[Src1<<8|Src2](argc=next.Dst)
	OP_CALL_BUILTIN_DIRECT // Dst = directTable[Src1<<8|Src2](argc=next.Dst) — direct pointer, no DispatchOrPanic

	// Exception handling
	OP_TRY     // Src1=errReg; Dst:Src2=signed jump offset to catch block
	OP_END_TRY // pop the innermost try handler (no error occurred)

	// Module import
	OP_IMPORT_MODULE // Dst=aliasReg; Src1<<8|Src2=constIdx(module name string)

	// Sentinel
	_OP_MAX
)

// Instruction is a single VM instruction. 4 bytes.
type Instruction struct {
	Op   OpCode
	Dst  uint8
	Src1 uint8
	Src2 uint8
}

// jumpOffset decodes the signed 16-bit jump offset from Dst:Src2 fields.
func (i Instruction) jumpOffset() int {
	return int(int16(int(i.Dst)<<8 | int(i.Src2)))
}

// constIndex decodes the 16-bit constant pool index from Src1:Src2 fields.
func (i Instruction) constIndex() int {
	return int(i.Src1)<<8 | int(i.Src2)
}

// opName returns a human-readable name for an opcode.
func opName(op OpCode) string {
	names := [_OP_MAX]string{
		"LOAD_CONST", "LOAD_VAR", "STORE_VAR", "LOAD_UPVAL", "STORE_UPVAL",
		"ADD", "SUB", "MUL", "DIV", "MOD", "POW", "NEG",
		"ADD_INT", "SUB_INT", "MUL_INT", "DIV_INT", "MOD_INT",
		"EQ", "NEQ", "LT", "LTE", "GT", "GTE", "IN",
		"AND", "OR", "NOT", "XOR",
		"JUMP", "JUMP_FALSE", "JUMP_TRUE", "RETURN",
		"MAKE_ARRAY", "MAKE_MAP", "SUBSCRIPT", "SET_INDEX", "ITER_NORMALIZE",
		"MAKE_FUNC", "CALL", "CALL_BUILTIN", "CALL_BUILTIN_DIRECT",
		"TRY", "END_TRY",
		"IMPORT_MODULE",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return ""
}

// Disassemble returns a human-readable listing of a Function's bytecode.
func Disassemble(fn *Function) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "function %q  regs=%d  consts=%d\n", fn.Name, fn.MaxRegs, len(fn.Constants))
	for i, instr := range fn.Code {
		fmt.Fprintf(&sb, "  %04d  %-14s  dst=%-3d src1=%-3d src2=%-3d\n",
			i, opName(instr.Op), instr.Dst, instr.Src1, instr.Src2)
	}
	return sb.String()
}
