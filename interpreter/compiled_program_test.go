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
	src := `x = 7
x + 1`
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

	for i := 0; i < 3; i++ {
		cfg := TestConfig()
		// Pollute cfg.Vars with a large map to grow regs on first run.
		if i == 0 {
			cfg.Vars = map[string]Value{"a": 1, "b": 2, "c": 3}
		}
		_, err = ExecuteCompiledVM(cp, cfg, vm)
		if err != nil {
			t.Fatalf("exec %d: %v", i, err)
		}
	}
}

func TestCompileProgram_RegexPrecompiled(t *testing.T) {
	// is_regex_match with literal pattern — pre-compile at compile time.
	// is_regex_match(pattern, str) — pattern is arg[0]
	src := `is_regex_match("^\\d+$", "12345")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile:", err)
	}

	// Verify one constant is a pre-compiled regexp.
	found := false
	for _, c := range cp.Fn.Constants {
		if _, ok := c.(interface{ MatchString(string) bool }); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pre-compiled *coregex.Regexp in constants, found none")
	}

	_, err = ExecuteCompiledVM(cp, TestConfig(), nil)
	if err != nil {
		t.Fatal("execute:", err)
	}
}

func TestCompileProgram_RegexMatchPrecompiled(t *testing.T) {
	// regex_match(text, pattern) — pattern is arg[1]
	src := `regex_match("12345", "^\\d+$")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile:", err)
	}

	found := false
	for _, c := range cp.Fn.Constants {
		if _, ok := c.(interface{ MatchString(string) bool }); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pre-compiled *coregex.Regexp in constants, found none")
	}

	_, err = ExecuteCompiledVM(cp, TestConfig(), nil)
	if err != nil {
		t.Fatal("execute:", err)
	}
}

func TestCompileProgram_RegexPrecompile_InvalidPattern(t *testing.T) {
	// Invalid pattern — should NOT pre-compile (fall back to runtime).
	src := `is_regex_match("[invalid", "text")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	// Compile should succeed (bad pattern = runtime error, not compile error).
	_, err = CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile should succeed even with bad pattern:", err)
	}
}
