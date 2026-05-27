# Phase 2 Performance Design — uddin-lang

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Date:** 2026-05-27
**Status:** Approved
**Scope:** Three sub-phases to make uddin-lang blazingly fast for rule-based evaluation (WAF, CEP, fact rules)

---

## Background & Motivation

Phase 1 delivered a register-based bytecode VM (~30% faster than tree-walker). For the primary use case — rule evaluation engines (WAF, CEP, scoring) — the remaining bottlenecks are not in instruction dispatch but in:

1. **Re-compilation per execution** — every `ExecuteProgram(prog)` re-compiles AST → bytecode, discards result. For a rule called 10,000×/day this is pure waste.
2. **Builtin dispatch chain** — `OP_CALL_BUILTIN` traverses 4 layers before reaching actual work. WAF rules call `match_regex`, `cidr_match`, `len` on every request.
3. **VM dispatch overhead** — after 2A+2B are done, the remaining bottleneck for computation-heavy rules is the interpreter loop itself.

Target after all three sub-phases: **10–200× speedup** vs tree-walker depending on rule type.

---

## Phase 2A — Bytecode Cache + Engine Persistence

### Goal

Eliminate re-compilation overhead. Compile once per `*Program`, reuse forever.

### Architecture

**Engine-level program cache** (`uddin.go`):

```go
type Engine struct {
    config       *interpreter.Config
    programCache sync.Map // *interpreter.Program → *interpreter.Function
}
```

`ExecuteProgram(prog *Program)` path:
1. Check `programCache.Load(prog)` — hit → skip compile, inject vars, execute
2. Miss → `NewCompiler().Compile(prog)` → `programCache.Store(prog, fn)` → execute

Cache key is `*Program` pointer identity. Same parsed program object → same compiled function. Thread-safe via `sync.Map` (Engine used concurrently across goroutines).

**Vars injection per-execution** (unchanged):
Bytecode is cached but `Config.Vars` values (`_waf_ctx`, `request`, etc.) are injected fresh at execution time via existing `preSeededVars` register seeding. Compile once, run with different data — already architected for this.

**Regex constant pre-compilation** (`interpreter/compiler.go`):

When compiler sees a call `match_regex(expr, "literal_pattern")` with a string literal as second argument:
- Compile `*regexp.Regexp` at bytecode compile time via `regexp.MustCompile`
- Store compiled `*regexp.Regexp` in `Function.Constants` pool
- Emit `OP_LOAD_CONST` for the pre-compiled regexp instead of string
- `match_regex` builtin detects `*regexp.Regexp` argument and skips re-compilation

This eliminates regex compilation from the hot path entirely.

**VM instance reset** (`interpreter/vm.go`):

Add `VM.Reset()` method:
```go
func (vm *VM) Reset() {
    vm.frames   = vm.frames[:0]
    vm.tryStack = vm.tryStack[:0]
    // regs intentionally NOT cleared — will be overwritten before use
}
```

`Engine` holds a `sync.Pool` of `*VM` instances (one pool per Engine, not package-level — Engines may have different configs). `ExecuteProgram` borrows from pool, resets, executes, returns to pool. Pool is initialized lazily on first `ExecuteProgram` call.

### Files Changed

| File | Change |
|------|--------|
| `uddin.go` | `programCache sync.Map` in Engine; cache check in `ExecuteProgram` |
| `interpreter/interpreter.go` | `executeVM` accepts pre-compiled `*Function` optionally |
| `interpreter/compiler.go` | Detect literal regex args; pre-compile to `*regexp.Regexp` |
| `interpreter/vm.go` | `Reset()` method; pooled VM support |

### Estimated Speedup

5–20× for typical WAF/CEP rules (dominated by eliminating re-compilation overhead).

---

## Phase 2B — Builtin Fast-Path

### Goal

Eliminate dispatch chain overhead for the most-called builtins. Current chain:

```
OP_CALL_BUILTIN
  └─ vmBuiltinTable[idx].fn(interp, pos, args)       ← closure wrapper
       └─ DispatchOrPanic(name, interp, pos, args)    ← string key
            └─ dispatcher.fns[name](...)              ← map lookup
                 └─ validateArgs(...)                 ← framework validation
                      └─ actual implementation        ← actual work
```

Overhead: ~30–100 ns per call from dispatch chain, not from the function itself.

### Architecture

**New opcode `OP_CALL_BUILTIN_DIRECT`** (`interpreter/opcode.go`):

Identical encoding to `OP_CALL_BUILTIN`. Compiler emits this when calling a builtin registered in the direct table. VM handler calls `vmBuiltinDirectTable[idx]` — a raw function pointer, no wrapper.

**Direct function pointer table** (`interpreter/vm_builtins.go`):

```go
// vmBuiltinDirectTable maps builtin index → direct Go function.
// nil entry means: fall back to vmBuiltinTable (DispatchOrPanic path).
var vmBuiltinDirectTable []func(*interpreter, Position, []Value) Value
```

Populated at `init()` alongside `vmBuiltinTable`. Index is stable (sorted alphabetically, same as `vmBuiltinTable`).

Direct implementations validate args with inline type assertions, not the framework. They panic with `runtimeError` on bad input — same behavior as DispatchOrPanic, different mechanism.

**Target builtins for direct table** (WAF/CEP priority):

| Builtin | Reason |
|---------|--------|
| `match_regex` | Core WAF pattern matching |
| `cidr_match` | IP allowlist/blocklist |
| `len` | Most frequent call overall |
| `typeof` | Type guards in rules |
| `to_string`, `to_int`, `to_float`, `to_bool` | Request data conversion |
| `contains`, `starts_with`, `ends_with` | String matching |
| `upper`, `lower`, `trim` | Input normalization |
| `split`, `join` | Header/path parsing |
| `keys`, `values` | Map iteration |
| `append` | Array building |

Remaining builtins (network, database, file, regex compile) keep the DispatchOrPanic path — they are I/O bound; dispatch overhead is negligible.

**Compiler detection** (`interpreter/compiler.go`):

In `compileBuiltinCall`, after resolving builtin index:
```go
if vmBuiltinDirectTable[idx] != nil {
    c.emit(Instruction{Op: OP_CALL_BUILTIN_DIRECT, ...})
} else {
    c.emit(Instruction{Op: OP_CALL_BUILTIN, ...})
}
```

### Files Changed

| File | Change |
|------|--------|
| `interpreter/opcode.go` | Add `OP_CALL_BUILTIN_DIRECT` |
| `interpreter/vm_builtins.go` | `vmBuiltinDirectTable`; direct implementations for target builtins |
| `interpreter/compiler.go` | Emit `OP_CALL_BUILTIN_DIRECT` when direct entry exists |
| `interpreter/vm.go` | Handler for `OP_CALL_BUILTIN_DIRECT` |

### Estimated Speedup

2–5× additional for builtin-heavy rules. Combined with 2A: **10–40× vs tree-walker**.

---

## Phase 2C — JIT via `tinygo-org/go-llvm`

### Goal

Eliminate VM dispatch overhead for hot functions. After 1000 invocations, compile to native machine code and swap the function pointer. No stop-the-world, no warmup pause visible to caller.

### Tiered Execution Model

```
Calls 1–999:    VM interpreter (2A+2B already optimal)
Call 1000:      Trigger async JIT compile (background goroutine)
Calls 1001+:    Native code via NativeFunc pointer
```

### Dependencies

```
tinygo-org/go-llvm       — CGo bindings to LLVM C API (IR build + JIT execution)
libLLVM ≥ 17             — system library

# macOS ARM64 (Apple M1)
brew install llvm@17
export CGO_LDFLAGS="-L$(brew --prefix llvm@17)/lib"
export CGO_CFLAGS="-I$(brew --prefix llvm@17)/include"
```

CGo is isolated to `jit/runtime.go`. All IR generation is pure Go.

### Components

#### `jit/jit.go` — JIT Controller

```go
type NativeFunc func(args []Value) Value

type JITController struct {
    threshold  int64
    callCounts sync.Map  // *Function → *int64 (atomic)
    compiled   sync.Map  // *Function → NativeFunc
}

func (j *JITController) RecordAndMaybeCompile(fn *Function) {
    ptr, _ := j.callCounts.LoadOrStore(fn, new(int64))
    if atomic.AddInt64(ptr.(*int64), 1) == j.threshold {
        go j.compileAsync(fn)
    }
}
```

`VM.callFunction` checks `compiled` first:
```go
func (vm *VM) callFunction(fn *Function, args []Value) Value {
    if vm.jit != nil {
        if native, ok := vm.jit.compiled.Load(fn); ok {
            return native.(NativeFunc)(args)
        }
        vm.jit.RecordAndMaybeCompile(fn)
    }
    return vm.callVM(fn, args)
}
```

Default threshold: 1000. Configurable via `Config.JITThreshold` (0 = JIT disabled).

#### `jit/codegen.go` — Bytecode → LLVM IR

Pure Go via `tinygo-org/go-llvm`. Two codegen paths:

**Typed path** — functions with full `ParamTypes` annotations:
```llvm
; fun fib(n: int): int  →  all i64, no boxing
define i64 @fib(i64 %n) {
entry:
  %cmp = icmp sle i64 %n, 1
  br i1 %cmp, label %base, label %rec
base:
  ret i64 1
rec:
  %n1 = sub i64 %n, 1
  %r1 = call i64 @fib(i64 %n1)
  %n2 = sub i64 %n, 2
  %r2 = call i64 @fib(i64 %n2)
  %sum = add i64 %r1, %r2
  ret i64 %sum
}
```

**Untyped path** — tagged union `{ i8 tag, i64 data }` with inline type guards.

Tag values are defined canonically in `jit/codegen.go` and must match `interpreter/memory_layout.go`:
`0=nil, 1=int, 2=float64, 3=string-ptr, 4=bool, 5=array-ptr, 6=map-ptr`
```llvm
; Guard: expect int (tag == 1)
%tag = extractvalue %Value %a, 0
%ok  = icmp eq i8 %tag, 1
br i1 %ok, label %fast, label %deopt   ; predicted: fast

fast:
  %ia = extractvalue %Value %a, 1      ; raw i64
  ; ... typed arithmetic

deopt:
  %res = call %Value @vm_deopt_execute(ptr %frame, i32 %deopt_pc)
  ret %Value %res
```

Tag values: `0=nil, 1=int, 2=float64, 3=string-ptr, 4=bool, 5=array-ptr, 6=map-ptr`

Guards are inline and branch-predicted as taken. Deopt is the cold path.

Builtin calls (`OP_CALL_BUILTIN`, `OP_CALL_BUILTIN_DIRECT`) are emitted as calls to exported Go functions — not inlined into LLVM IR. I/O builtins are already CGo-bound; dispatch overhead is negligible vs I/O cost.

#### `jit/runtime.go` — CGo Execution Engine

```go
// #include <llvm-c/ExecutionEngine.h>
// #cgo LDFLAGS: -lLLVM
import "C"

func ExecuteModule(mod llvm.Module) (NativeFunc, error) {
    // LLVMCreateMCJITCompilerForModule with O2 optimization
    // LLVMGetFunctionAddress for entry point
    // Wrap as NativeFunc
}
```

Only file with CGo. Wraps LLVM MCJIT execution engine.

#### `jit/cache.go` — Content-Addressed Cache

Cache keyed by FNV-64 hash of bytecode content (not pointer) — enables reuse across Engine instances for identical rules:

```go
type CompiledCache struct {
    mu    sync.RWMutex
    cache map[uint64]NativeFunc  // hash(bytecode) → NativeFunc
}
```

Prevents recompiling identical rules loaded from different `*Program` objects.

#### `jit/deopt.go` — Deoptimization Callback

```go
//export vm_deopt_execute
func vm_deopt_execute(framePtr unsafe.Pointer, pc C.int) C.Value {
    // Reconstruct VM frame from pointer
    // Resume execution at pc in VM interpreter
    // Return result to LLVM IR caller
}
```

Called when a type guard fails in JIT-compiled code. Allows native code to fall back to VM for unexpected types without aborting execution.

### Engine Integration

`Engine` gets `JITController` (created once, persists for Engine lifetime):

```go
type Engine struct {
    config       *interpreter.Config
    programCache sync.Map
    jit          *jit.JITController  // nil if JIT disabled
}
```

`Config.JITThreshold int64` — set to 0 to disable JIT. Default: 1000.

### File Structure

```
interpreter/jit/
  jit.go      — JITController, threshold, NativeFunc registry
  codegen.go  — Function (bytecode) → llvm.Module (pure Go)
  runtime.go  — CGo: LLVM MCJIT execution engine
  cache.go    — content-addressed compiled function cache
  deopt.go    — exported C deoptimization callback
```

### Estimated Speedup

| Rule type | After 2A+2B | After 2C |
|-----------|-------------|----------|
| Numeric/scoring (typed) | 15–30× | 100–200× vs tree-walker |
| Numeric/scoring (untyped) | 10–20× | 30–60× |
| WAF rules (regex+cidr dominant) | 10–40× | 10–45× (builtin-bound) |
| Mixed logic + calculation | 12–25× | 40–100× |

---

## Implementation Sequence

```
Phase 2A (2–3 weeks)
  ├── Engine programCache (sync.Map, *Program → *Function)
  ├── executeVM accepts pre-compiled *Function
  ├── Regex literal pre-compilation in compiler
  └── VM Reset() + pooled VM via sync.Pool

Phase 2B (2–3 weeks)
  ├── OP_CALL_BUILTIN_DIRECT opcode
  ├── vmBuiltinDirectTable + direct implementations
  ├── Compiler emits DIRECT opcode where applicable
  └── Benchmark: direct vs DispatchOrPanic path

Phase 2C (2–4 months)
  ├── jit/jit.go — JITController
  ├── jit/codegen.go — bytecode → LLVM IR (typed path first)
  ├── jit/runtime.go — CGo MCJIT wrapper
  ├── jit/cache.go — content-addressed cache
  ├── jit/deopt.go — deopt callback
  ├── VM integration: callFunction checks JIT compiled map
  ├── Engine integration: JITController lifetime
  └── Untyped path with type guards + deopt (after typed path stable)
```

Each sub-phase is independently shippable. 2A and 2B have no new dependencies.
2C requires `libLLVM ≥ 17` and `tinygo-org/go-llvm`.

---

## What Is Not In Scope

- Garbage collector changes
- WASM compilation target
- Bytecode serialization (`.dinc` cache files)
- JIT for builtin I/O functions (network, database, filesystem) — already CGo-bound, dispatch overhead negligible
- Type inference — only explicit optional annotations
- Cross-compilation (separate initiative)
