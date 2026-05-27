package interpreter

import (
	"sync"
	"testing"
)

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


func BenchmarkCompiledProgram_NoPool(b *testing.B) {
	src := `
fun fib(n: int):
    if (n <= 1) then:
        return n
    end
    return fib(n - 1) + fib(n - 2)
end
fib(20)
`
	prog, _ := ParseProgram([]byte(src))
	cp, _ := CompileProgram(prog, nil)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		_, _ = ExecuteCompiledVM(cp, cfg, nil)
	}
}

func BenchmarkCompiledProgram_WithPool(b *testing.B) {
	src := `
fun fib(n: int):
    if (n <= 1) then:
        return n
    end
    return fib(n - 1) + fib(n - 2)
end
fib(20)
`
	prog, _ := ParseProgram([]byte(src))
	cp, _ := CompileProgram(prog, nil)
	cfg := TestConfig()

	var pool sync.Pool
	pool.New = func() any { return NewVM(nil) }

	b.ResetTimer()
	for b.Loop() {
		vm := pool.Get().(*VM)
		_, _ = ExecuteCompiledVM(cp, cfg, vm)
		pool.Put(vm)
	}
}

func BenchmarkCompiledProgram_RegexRuntimeCompile(b *testing.B) {
	// Pattern is a variable — compiled at runtime each call.
	src := `is_regex_match(pattern, text)`
	prog, _ := ParseProgram([]byte(src))
	cp, _ := CompileProgram(prog, []string{"pattern", "text"})
	cfg := TestConfig()
	cfg.Vars = map[string]Value{"pattern": `^\d+$`, "text": "12345"}
	b.ResetTimer()
	for b.Loop() {
		_, _ = ExecuteCompiledVM(cp, cfg, nil)
	}
}

func BenchmarkCompiledProgram_RegexPrecompile(b *testing.B) {
	// Pattern is a literal — pre-compiled at bytecode compile time.
	src := `is_regex_match("^\\d+$", text)`
	prog, _ := ParseProgram([]byte(src))
	cp, _ := CompileProgram(prog, []string{"text"})
	cfg := TestConfig()
	cfg.Vars = map[string]Value{"text": "12345"}
	b.ResetTimer()
	for b.Loop() {
		_, _ = ExecuteCompiledVM(cp, cfg, nil)
	}
}
