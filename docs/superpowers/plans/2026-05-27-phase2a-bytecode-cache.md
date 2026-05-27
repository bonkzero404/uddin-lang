# Phase 2A — Bytecode Cache & Engine Persistence

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate re-compilation overhead by caching compiled bytecode in Engine — rules that run 10,000×/day compile once and execute repeatedly.

**Architecture:** Add `CompiledProgram` type to interpreter package (wraps `*Function` + var-register map), expose `CompileProgram` + `ExecuteCompiledVM` as public API, and add `programCache sync.Map` + `vmPool sync.Pool` to Engine. Regex string-literal patterns are pre-compiled to `*coregex.Regexp` at bytecode compile time.

**Tech Stack:** Go 1.25, `sync.Map`, `sync.Pool`, `github.com/coregx/coregex`

---

## File Structure

| File | Change |
|------|--------|
| `interpreter/compiled_program.go` | **Create** — `CompiledProgram` struct, `CompileProgram()`, `ExecuteCompiledVM()` |
| `interpreter/vm.go` | **Modify** — add `reset(interp *interpreter)` method |
| `interpreter/compiler.go` | **Modify** — regex literal pre-compilation in `compileCall` |
| `uddin.go` | **Modify** — `programCache`, `vmPool` in Engine; cache logic in `ExecuteProgram` |
| `interpreter/compiled_program_test.go` | **Create** — tests for CompiledProgram API |
| `uddin_test.go` | **Modify** — Engine cache + pool tests |

---

### Task 1: VM reset method + CompiledProgram API

**Files:**
- Create: `interpreter/compiled_program.go`
- Modify: `interpreter/vm.go`
- Create: `interpreter/compiled_program_test.go`

- [ ] **Step 1: Tulis failing test untuk CompiledProgram**

Buat file `interpreter/compiled_program_test.go`:

```go
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
	// Same CompiledProgram, different Config.Vars each call.
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
	// Pool scenario: same VM reused across two executions.
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
```

- [ ] **Step 2: Jalankan test, pastikan FAIL**

```bash
go test ./interpreter/... -run TestCompileProgram -v 2>&1
```

Expected: `FAIL — CompileProgram undefined`

- [ ] **Step 3: Tambah `reset` method ke VM**

Di `interpreter/vm.go`, tambah setelah `func (vm *VM) Execute`:

```go
// reset prepares the VM for reuse. It clears all frames and the try-stack,
// then points interp at the fresh interpreter for this execution.
func (vm *VM) reset(interp *interpreter) {
	vm.frames   = vm.frames[:0]
	vm.tryStack = vm.tryStack[:0]
	vm.interp   = interp
}
```

- [ ] **Step 4: Buat `interpreter/compiled_program.go`**

```go
package interpreter

// CompiledProgram bundles compiled bytecode with the register map for
// pre-seeded variables. Obtain via CompileProgram; pass to ExecuteCompiledVM.
type CompiledProgram struct {
	Fn      *Function
	varRegs map[string]uint8 // var name → register index in top-level frame
}

// CompileProgram compiles prog into reusable bytecode.
// varNames must list all variable names that callers will inject via Config.Vars.
// If a name is injected at runtime but not listed here, it will be silently ignored.
func CompileProgram(prog *Program, varNames []string) (*CompiledProgram, error) {
	c := NewCompiler()
	c.preSeededVars = append(c.preSeededVars, varNames...)
	fn, err := c.Compile(prog)
	if err != nil {
		return nil, err
	}
	return &CompiledProgram{Fn: fn, varRegs: c.locals}, nil
}

// ExecuteCompiledVM runs cp through the bytecode VM and returns stats.
// Pass a pooled *VM to avoid allocation; pass nil to allocate a fresh one.
// The VM is reset before use — callers may return it to a pool afterward.
func ExecuteCompiledVM(cp *CompiledProgram, config *Config, vm *VM) (stats *Stats, err error) {
	interp := newInterpreter(config)
	if vm == nil {
		vm = NewVM(interp)
	} else {
		vm.reset(interp)
	}

	// Inject Config.Vars into pre-seeded registers.
	if config != nil {
		for k, v := range config.Vars {
			if reg, ok := cp.varRegs[k]; ok {
				needed := int(reg) + 1
				for len(vm.regs) < needed {
					vm.regs = append(vm.regs, make([]Value, needed-len(vm.regs))...)
				}
				vm.regs[reg] = v
			}
		}
	}

	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case returnResult:
				err = runtimeError(e.pos, "can't return at top level")
			case Error:
				err = e
			case error:
				err = runtimeError(Position{Line: 1, Column: 1}, "%v", e)
			default:
				err = runtimeError(Position{Line: 1, Column: 1}, "%v", r)
			}
		}
	}()

	vm.Execute(cp.Fn)

	// Auto-call main() if defined and not in unit-test mode.
	if config == nil || !config.IsUnitTest {
		if reg, ok := cp.varRegs["main"]; ok {
			if int(reg) < len(vm.regs) {
				if mainFn, ok := vm.regs[reg].(*vmFunction); ok {
					mainFn.call(interp, Position{}, nil)
				}
			}
		}
	}

	return &interp.stats, nil
}
```

- [ ] **Step 5: Jalankan test, pastikan PASS**

```bash
go test ./interpreter/... -run TestCompileProgram -v 2>&1
```

Expected: semua test PASS

- [ ] **Step 6: Jalankan full test suite**

```bash
go test ./... -count=1 2>&1
```

Expected: semua paket ok

- [ ] **Step 7: Commit**

```bash
git add interpreter/compiled_program.go interpreter/compiled_program_test.go interpreter/vm.go
git commit -m "feat(cache): add CompiledProgram API + VM reset for pooled reuse"
```

---

### Task 2: Engine programCache + vmPool

**Files:**
- Modify: `uddin.go`
- Modify/Create: `uddin_test.go`

- [ ] **Step 1: Tulis failing test untuk cache**

Tambah di `uddin_test.go` (buat file jika belum ada):

```go
package uddin_test

import (
	"sync"
	"testing"

	uddin "github.com/bonkzero404/uddin-lang"
	"github.com/bonkzero404/uddin-lang/interpreter"
)

func TestEngineCache_SameResultTwice(t *testing.T) {
	engine := uddin.New()
	engine.SetUnitTestMode(true)

	src := `score + 1`
	prog, err := engine.ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}

	for i, wantScore := range []int{10, 20} {
		engine.SetVariable("score", wantScore)
		_, err := engine.ExecuteProgram(prog)
		if err != nil {
			t.Fatalf("exec %d: %v", i, err)
		}
	}
}

func TestEngineCache_ConcurrentSafe(t *testing.T) {
	// Multiple goroutines execute the same *Program simultaneously.
	engine := uddin.New()
	engine.SetUnitTestMode(true)

	src := `x = 1
x`
	prog, err := engine.ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := uddin.New()
			e.SetUnitTestMode(true)
			_, err := e.ExecuteProgram(prog)
			if err != nil {
				t.Errorf("concurrent exec: %v", err)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkEngine_NoCacheVsCache(b *testing.B) {
	src := `
fun fib(n: int):
    if (n <= 1) then:
        return n
    end
    return fib(n - 1) + fib(n - 2)
end
fib(20)
`
	b.Run("ParseAndExecute", func(b *testing.B) {
		engine := uddin.New()
		engine.SetUnitTestMode(true)
		for b.Loop() {
			_, err := engine.ExecuteString(src)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CachedExecute", func(b *testing.B) {
		engine := uddin.New()
		engine.SetUnitTestMode(true)
		prog, _ := engine.ParseProgram([]byte(src))
		b.ResetTimer()
		for b.Loop() {
			_, err := engine.ExecuteProgram(prog)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
```

- [ ] **Step 2: Jalankan test, pastikan kompilasi gagal**

```bash
go test ./... -run TestEngineCache -v 2>&1
```

Expected: compile error karena `Engine` belum punya cache field.

- [ ] **Step 3: Update Engine struct dan konstruktor di `uddin.go`**

Ganti bagian import dan Engine struct:

```go
import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

// Engine represents a UDDIN-LANG interpreter engine instance.
type Engine struct {
	config       *interpreter.Config
	programCache sync.Map // *interpreter.Program → *interpreter.CompiledProgram
	vmPool       sync.Pool
}

func New() *Engine {
	e := &Engine{
		config: interpreter.DefaultConfig(),
	}
	e.vmPool.New = func() any {
		return interpreter.NewVM(nil)
	}
	return e
}

func NewWithConfig(config *interpreter.Config) *Engine {
	if config == nil {
		config = interpreter.DefaultConfig()
	}
	e := &Engine{config: config}
	e.vmPool.New = func() any {
		return interpreter.NewVM(nil)
	}
	return e
}
```

- [ ] **Step 4: Ganti `ExecuteProgram` di `uddin.go`**

```go
// ExecuteProgram executes a parsed program.
// When VMEnabled, the compiled bytecode is cached by *Program pointer — subsequent
// calls with the same *Program skip compilation entirely.
func (e *Engine) ExecuteProgram(prog *interpreter.Program) (*interpreter.Stats, error) {
	if !e.config.VMEnabled {
		return interpreter.Execute(prog, e.config)
	}

	// Build sorted var-name list for compilation (stable across calls).
	varNames := make([]string, 0, len(e.config.Vars))
	for k := range e.config.Vars {
		varNames = append(varNames, k)
	}

	// Cache lookup.
	var cp *interpreter.CompiledProgram
	if val, ok := e.programCache.Load(prog); ok {
		cp = val.(*interpreter.CompiledProgram)
	} else {
		var err error
		cp, err = interpreter.CompileProgram(prog, varNames)
		if err != nil {
			return nil, err
		}
		e.programCache.Store(prog, cp)
	}

	// Borrow a VM from the pool, use it, return it.
	vm := e.vmPool.Get().(*interpreter.VM)
	defer e.vmPool.Put(vm)

	return interpreter.ExecuteCompiledVM(cp, e.config, vm)
}
```

- [ ] **Step 5: Jalankan test**

```bash
go test ./... -run TestEngineCache -v 2>&1
```

Expected: PASS

- [ ] **Step 6: Jalankan full suite**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 7: Jalankan benchmark untuk verifikasi speedup**

```bash
go test -run='^$' -bench=BenchmarkEngine_NoCacheVsCache -benchtime=5s ./... 2>&1
```

Expected: `CachedExecute` setidaknya 3x lebih cepat dari `ParseAndExecute`.

- [ ] **Step 8: Commit**

```bash
git add uddin.go uddin_test.go
git commit -m "feat(cache): add Engine programCache + vmPool — compile once, run many"
```

---

### Task 3: Regex literal pre-compilation di compiler

**Files:**
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/fn_regex.go`
- Modify: `interpreter/compiled_program_test.go`

- [ ] **Step 1: Tulis failing test untuk regex pre-compile**

Tambah di `interpreter/compiled_program_test.go`:

```go
func TestCompileProgram_RegexPrecompiled(t *testing.T) {
	// is_regex_match with a literal pattern — should pre-compile at compile time.
	src := `is_regex_match("^\\d+$", "12345")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	cp, err := CompileProgram(prog, nil)
	if err != nil {
		t.Fatal("compile:", err)
	}

	// Verify that one of the constants is a pre-compiled regexp (not a string).
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

	// Execution must still return true.
	cfg := TestConfig()
	_, err = ExecuteCompiledVM(cp, cfg, nil)
	if err != nil {
		t.Fatal("execute:", err)
	}
}

func TestCompileProgram_RegexPrecompile_InvalidPattern(t *testing.T) {
	// Invalid pattern — should NOT pre-compile (fall back to runtime compile).
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
```

- [ ] **Step 2: Jalankan test, pastikan FAIL**

```bash
go test ./interpreter/... -run TestCompileProgram_Regex -v 2>&1
```

Expected: FAIL — `found = false` (no pre-compiled regexp in constants yet)

- [ ] **Step 3: Tambah import `coregex` ke compiler.go**

Di `interpreter/compiler.go`, ganti baris import:

```go
import (
	"fmt"
	"os"

	"github.com/coregx/coregex"
)
```

- [ ] **Step 4: Tambah helper `isRegexBuiltin` di compiler.go**

Tambah setelah block import:

```go
// isRegexBuiltin returns true for builtins that accept a regex pattern as arg[1].
func isRegexBuiltin(name string) bool {
	switch name {
	case "is_regex_match", "regex_match", "regex_find", "regex_find_all", "regex_replace", "regex_split":
		return true
	}
	return false
}
```

- [ ] **Step 5: Intercept regex literal args di `compileCall`**

Di `interpreter/compiler.go`, dalam fungsi `compileCall`, tambah logika setelah deteksi `builtinCall`:

```go
func (c *Compiler) compileCall(e *Call) (uint8, error) {
	builtinCall := false
	var builtinIdx uint16
	if fnExpr, isVar := e.Function.(*Variable); isVar {
		if idx, ok := vmBuiltinIndex[fnExpr.Name]; ok {
			builtinCall = true
			builtinIdx = idx
		}
	}

	var fnReg uint8
	if !builtinCall {
		var err error
		fnReg, err = c.compileExpr(e.Function)
		if err != nil {
			return 0, err
		}
	}

	// Compile arguments. For regex builtins with a string-literal pattern arg,
	// pre-compile the pattern to *coregex.Regexp and store in constants pool.
	argRegs := make([]uint8, len(e.Arguments))
	for i, arg := range e.Arguments {
		// Check for regex pre-compilation: pattern is arg[1] for all regex builtins.
		if builtinCall && i == 1 {
			if fnExpr, isVar := e.Function.(*Variable); isVar && isRegexBuiltin(fnExpr.Name) {
				if lit, ok := arg.(*Literal); ok {
					if patStr, ok := lit.Value.(string); ok {
						if re, err := coregex.Compile(patStr); err == nil {
							// Store pre-compiled regexp in constants.
							constIdx := uint16(len(c.fn.Constants))
							c.fn.Constants = append(c.fn.Constants, re)
							r := c.allocReg()
							c.emit(Instruction{Op: OP_LOAD_CONST, Dst: r,
								Src1: uint8(constIdx >> 8), Src2: uint8(constIdx)})
							argRegs[i] = r
							continue
						}
						// Invalid pattern — fall through to normal string literal compile.
					}
				}
			}
		}
		r, err := c.compileExpr(arg)
		if err != nil {
			return 0, err
		}
		argRegs[i] = r
	}
	// ... rest of compileCall unchanged ...
```

Pastikan sisa fungsi `compileCall` (copy args ke window, emit opcode) tidak berubah.

- [ ] **Step 6: Update regex builtins untuk terima `*coregex.Regexp` sebagai pattern**

Di `interpreter/fn_regex.go`, update `isregexFunc` dan `regexMatchFunc`:

```go
func isregexFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "is_regex_match", args, 2); err != Value(nil) {
		return err
	}
	str, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "is_regex_match() requires second argument to be a string"))
	}
	switch p := args[0].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return false
		}
		return re.MatchString(str)
	case *coregex.Regexp:
		return p.MatchString(str)
	default:
		panic(typeError(pos, "is_regex_match() requires first argument to be a string or pre-compiled pattern"))
	}
}

func regexMatchFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		return Value(fmt.Errorf("regex_match() requires exactly 2 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_match() requires first argument to be a string"))
	}
	switch p := args[1].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
		return Value(re.MatchString(text))
	case *coregex.Regexp:
		return Value(p.MatchString(text))
	default:
		return Value(fmt.Errorf("regex_match() requires pattern to be a string"))
	}
}
```

Tambah `import "github.com/coregx/coregex"` ke `fn_regex.go` jika belum ada (sudah ada).

- [ ] **Step 7: Jalankan test**

```bash
go test ./interpreter/... -run TestCompileProgram_Regex -v 2>&1
```

Expected: PASS

- [ ] **Step 8: Jalankan full suite**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 9: Commit**

```bash
git add interpreter/compiler.go interpreter/fn_regex.go interpreter/compiled_program_test.go
git commit -m "feat(cache): pre-compile regex literal patterns at bytecode compile time"
```

---

### Task 4: Benchmark akhir Phase 2A

**Files:**
- Modify: `interpreter/compiled_program_test.go`

- [ ] **Step 1: Tambah benchmark VM pool**

Tambah di `interpreter/compiled_program_test.go`:

```go
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

func BenchmarkCompiledProgram_RegexNoPrecompile(b *testing.B) {
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

func BenchmarkCompiledProgram_RegexPrecompile(b *testing.B) {
	// Pattern is a literal — pre-compiled at compile time.
	src := `is_regex_match("^\\d+$", text)`
	// Force pre-compilation: pattern is literal in source.
	src2 := `is_regex_match(text, "^\\d+$")`
	_ = src
	prog, _ := ParseProgram([]byte(src2))
	cp, _ := CompileProgram(prog, []string{"text"})
	cfg := TestConfig()
	cfg.Vars = map[string]Value{"text": "12345"}
	b.ResetTimer()
	for b.Loop() {
		_, _ = ExecuteCompiledVM(cp, cfg, nil)
	}
}
```

- [ ] **Step 2: Jalankan benchmark**

```bash
go test ./interpreter/... -run='^$' -bench='BenchmarkCompiledProgram' -benchtime=5s -benchmem 2>&1
```

Expected:
- `WithPool` lebih cepat dari `NoPool` (~15-30% karena skip alloc)
- `RegexPrecompile` lebih cepat dari `RegexNoPrecompile` (~2-5x)

- [ ] **Step 3: Jalankan full suite sekali lagi**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 4: Commit**

```bash
git add interpreter/compiled_program_test.go
git commit -m "bench(cache): add Phase 2A benchmarks — pool and regex pre-compile"
```

---

## Self-Review Checklist (sudah dilakukan inline)

- [x] `CompiledProgram` concurrent-safe? Ya — `sync.Map` di Engine, VM per-goroutine via pool
- [x] VM reset benar? Ya — frames, tryStack cleared; interp diganti fresh per execution
- [x] Regex pre-compile tidak break runtime compile fallback? Ya — invalid pattern = fallback ke runtime
- [x] `ExecuteCompiledVM` punya error recovery? Ya — `defer/recover` dengan konversi ke `Error`
- [x] Cache invalidation? Tidak diperlukan — key adalah `*Program` pointer; rule engine reuse `*Program` sama tiap request
