package interpreter

import "testing"

func TestCompileLiteral(t *testing.T) {
	src := `42`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	if len(fn.Code) == 0 {
		t.Fatal("expected at least one instruction")
	}
	last := fn.Code[len(fn.Code)-1]
	if last.Op != OP_RETURN {
		t.Errorf("expected last instruction RETURN, got %s", opName(last.Op))
	}
	found := false
	for _, cv := range fn.Constants {
		if cv == 42 {
			found = true
		}
	}
	if !found {
		t.Errorf("constant 42 not in constant pool: %v", fn.Constants)
	}
}

func TestCompileVarAssignAndLoad(t *testing.T) {
	src := "x = 10\nx"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	ops := make([]OpCode, len(fn.Code))
	for i, instr := range fn.Code {
		ops[i] = instr.Op
	}
	hasStore := false
	hasLoad := false
	for _, op := range ops {
		if op == OP_STORE_VAR {
			hasStore = true
		}
		if op == OP_LOAD_VAR {
			hasLoad = true
		}
	}
	if !hasStore {
		t.Error("expected STORE_VAR instruction")
	}
	if !hasLoad {
		t.Error("expected LOAD_VAR instruction")
	}
}
