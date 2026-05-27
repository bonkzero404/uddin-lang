package interpreter

import "testing"

func TestCompileProgram_Basic(t *testing.T) {
	src := `x = 42
x`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result, err := ExecuteCompiledVM(cp, TestConfig(), nil)
	if err != nil {
		t.Fatal("execute:", err)
	}
	_ = result
}

func TestCompileProgram_VarsInjected(t *testing.T) {
	src := `score`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, []string{"score"})
	if err != nil {
		t.Fatal("compile:", err)
	}
	cfg := TestConfig()
	cfg.Vars = map[string]Value{"score": 99}
	_, err = ExecuteCompiledVM(cp, cfg, nil)
	if err != nil {
		t.Fatal("execute:", err)
	}
}

func TestCompileProgram_ReusedAcrossExecutions(t *testing.T) {
	src := `threshold`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, []string{"threshold"})
	if err != nil {
		t.Fatal("compile:", err)
	}

	for _, want := range []int{10, 20, 30} {
		cfg := TestConfig()
		cfg.Vars = map[string]Value{"threshold": want}
		_, err := ExecuteCompiledVM(cp, cfg, nil)
		if err != nil {
			t.Fatalf("execute with threshold=%d: %v", want, err)
		}
	}
}

func TestCompileProgram_VMReset(t *testing.T) {
	src := `x = 1
x`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile:", err)
	}

	interp := newInterpreter(TestConfig())
	vm := NewVM(interp)

	// First execution
	_, err = ExecuteCompiledVM(cp, TestConfig(), vm)
	if err != nil {
		t.Fatal("first exec:", err)
	}

	// Second execution with same VM (should be reset internally)
	_, err = ExecuteCompiledVM(cp, TestConfig(), vm)
	if err != nil {
		t.Fatal("second exec:", err)
	}
}
