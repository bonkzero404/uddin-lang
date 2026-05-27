package interpreter

import "testing"

func TestInstructionEncoding(t *testing.T) {
	instr := Instruction{Op: OP_LOAD_CONST, Dst: 1, Src1: 0, Src2: 5}
	if instr.Op != OP_LOAD_CONST {
		t.Errorf("Op mismatch: got %d", instr.Op)
	}
	if instr.Dst != 1 || instr.Src1 != 0 || instr.Src2 != 5 {
		t.Errorf("register fields wrong: Dst=%d Src1=%d Src2=%d", instr.Dst, instr.Src1, instr.Src2)
	}
}

func TestOpcodeStringNotEmpty(t *testing.T) {
	for op := OpCode(0); op < _OP_MAX; op++ {
		if opName(op) == "" {
			t.Errorf("opcode %d has no name", op)
		}
	}
}

func TestFunctionHasCode(t *testing.T) {
	fn := &Function{
		Name:      "test",
		MaxRegs:   4,
		Code:      []Instruction{{Op: OP_RETURN, Src1: 0}},
		Constants: []Value{nil},
	}
	if len(fn.Code) != 1 {
		t.Error("expected 1 instruction")
	}
}
