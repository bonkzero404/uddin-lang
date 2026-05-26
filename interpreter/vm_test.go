package interpreter

import "testing"

func makeVM(cfg *Config) *VM {
	interp := newInterpreter(cfg)
	return NewVM(interp)
}

func TestVMReturnLiteral(t *testing.T) {
	prog, _ := ParseProgram([]byte(`42`))
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	vm := makeVM(TestConfig())
	result := vm.Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v (%T)", result, result)
	}
}

func TestVMVarAssign(t *testing.T) {
	prog, _ := ParseProgram([]byte("x = 7\nx"))
	c := NewCompiler()
	fn, _ := c.Compile(prog)
	vm := makeVM(TestConfig())
	result := vm.Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}
