# Compiler Pipeline Design — uddin-lang

**Date:** 2026-05-25  
**Status:** Approved  
**Scope:** Three-phase plan to bring uddin-lang execution speed from tree-walker to near-native (LLVM JIT)

---

## Background & Motivation

Current architecture is a **tree-walking interpreter**: the AST is traversed directly at runtime. This is correct and maintainable but carries fundamental overhead:

- Every statement re-evaluates the AST node type via `switch e := expr.(type)`
- Every variable access traverses a `[]map[string]Value` scope stack
- Every function call pushes scope maps onto the heap
- `return` is implemented via `panic/recover`, which is slow and semantically fragile

Benchmarks baseline (Apple M1, interpreter package):
- `BenchmarkAdvancedNestedFunctionCalls` — 870 µs/op
- `BenchmarkAdvancedMapOperations` — 897 µs/op
- `BenchmarkAdvancedComplexDataStructures` — 1111 µs/op

Target after all three phases: **30-200x speedup** depending on type-annotation coverage.

---

## Phase 0 — Bug Fixes (Foundation)

Fix seven confirmed bugs before touching the compiler pipeline. All fixes are non-breaking.

### Group A: Data corruption / blocking

**Bug 1 — Global `factDatabase` / `eventStore`** (`functions.go:10-13`)

Package-level globals shared across all `Engine` instances. Concurrent or multi-tenant use silently corrupts both stores.

Fix: Move both to the `interpreter` struct. Initialize per `newInterpreter()` call.

```go
// Before (functions.go)
var factDatabase = make(map[string]any)
var eventStore   = make([]map[string]any, 0)

// After (interpreter struct)
type interpreter struct {
    // ...
    factDatabase map[string]any
    eventStore   []map[string]any
}
```

**Bug 2 — Pool misuse in `evalPlus`** (`interpreter.go:330-338`)

`GetPooledArray` returns a pooled slice. `SmartAppend` may grow it (new backing array). `defer PutPooledArray(result)` returns the original slice header — which no longer owns the data. Pooled memory is corrupted.

Fix: Remove `defer PutPooledArray`. Allocate fresh slice after SmartAppend, do not pool intermediate result.

**Bug 3 — `ensureIntToFloats` false-zero sentinel** (`interpreter.go:100-122`)

Returns `(0, 0)` for both "type error" and "both operands are legitimately zero". `evalDivide` cannot distinguish — `0 / 0.0` panics with a type error instead of a divide-by-zero error.

Fix: Change signature to `(float64, float64, bool)` — return `false` on type error, `true` on success. Callers check the bool.

### Group B: Performance / correctness

**Bug 4 — `safeValueEqual` uses `reflect.DeepEqual`** (`variable_lookup.go:39-48`)

`reflect.DeepEqual` is ~100x slower than direct comparison for scalar types. The variable lookup cache becomes slower than no cache for the common case (int, string, bool variables).

Fix: Replace with explicit type switch:
```go
switch a := a.(type) {
case int:    return a == b.(int)
case float64: return a == b.(float64)
case string: return a == b.(string)
case bool:   return a == b.(bool)
case *[]Value: return a == b.(*[]Value)   // pointer identity
default:     return false
}
```

**Bug 5 — Dead code: 4-way function dispatch** (`interpreter.go:717-780`)

`*userFunction` and `builtinFunction` both implement `functionType`. The first check always matches. The three subsequent type checks are dead code that confuses readers and prevents future refactoring.

Fix: Remove the three redundant type-assertion blocks. Keep only the `functionType` check and the `any`-wrapped fallback.

**Bug 6 — Contradictory stability labels**

`StableMemoizationConfig()` is labeled "production-ready". `userFunction.Memoized` is labeled "EXPERIMENTAL... not thread-safe". `ExpressionOptimizer` is labeled "EXPERIMENTAL" but instantiated in `DefaultConfig`.

Fix: Standardize to three labels across all files: `// STABLE`, `// EXPERIMENTAL(reason)`, `// DEPRECATED(use X instead)`.

**Bug 7 — Tagged value overhead on every assign/lookup**

`IsTaggedValuesEnabled()` wraps every variable assignment in `SafeTaggedValue` (benchmark: 161 ns/op) and unwraps on every lookup. On the hot path of every statement, this is significant overhead when enabled.

Fix: Default `IsTaggedValuesEnabled()` to `false`. Only activate when `ExperimentalConfig()` is explicitly used. Document that experimental config is for benchmarking only.

---

## Phase 1 — Register-Based Bytecode VM

### Rationale: register-based over stack-based

Register-based VMs (Lua 5.x, Dalvik) produce fewer instructions per operation than stack-based VMs (CPython, JVM). More importantly, register assignments map directly to LLVM SSA values in Phase 2 — the transition from VM to JIT is cleaner with register IR.

### Instruction format

4 bytes per instruction. Fits in one CPU word on 32-bit and 64-bit machines. Cache-line efficient.

```
┌──────────┬──────┬──────┬──────┐
│  OpCode  │ Dst  │ Src1 │ Src2 │
│  8 bits  │8 bit │8 bit │8 bit │
└──────────┴──────┴──────┴──────┘

Special encodings:
  LOAD_CONST  → Src1:Src2 = 16-bit constant pool index
  JUMP        → Src1:Src2 = 16-bit signed PC offset
  CALL        → Src1 = function register, Src2 = arg count
```

### Opcode set (initial)

```go
type OpCode uint8

const (
    // Load / store
    OP_LOAD_CONST   // Dst = constants[Src1<<8|Src2]
    OP_LOAD_VAR     // Dst = locals[Src1]
    OP_STORE_VAR    // locals[Dst] = Src1
    OP_LOAD_UPVAL   // Dst = closure.upvalues[Src1]
    OP_STORE_UPVAL  // closure.upvalues[Dst] = Src1

    // Arithmetic (generic — works on any Value)
    OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_MOD, OP_POW, OP_NEG

    // Arithmetic (typed — int only, no boxing, emitted when type hint known)
    OP_ADD_INT, OP_SUB_INT, OP_MUL_INT, OP_DIV_INT, OP_MOD_INT

    // Comparison
    OP_EQ, OP_NEQ, OP_LT, OP_LTE, OP_GT, OP_GTE, OP_IN

    // Logic
    OP_AND, OP_OR, OP_NOT, OP_XOR

    // Control flow
    OP_JUMP         // unconditional
    OP_JUMP_FALSE   // if !regs[Src1]: pc += int16(Src1:Src2)
    OP_JUMP_TRUE
    OP_RETURN       // return regs[Src1]

    // Collections
    OP_MAKE_ARRAY   // Dst = []Value{regs[Src1]..regs[Src1+Src2-1]}
    OP_MAKE_MAP     // Dst = map (Src2 key+value pairs)
    OP_SUBSCRIPT    // Dst = regs[Src1][regs[Src2]]
    OP_SET_INDEX    // regs[Src1][regs[Src2]] = regs[Dst]

    // Functions
    OP_MAKE_FUNC    // Dst = closure(funcIdx, capture current upvalues)
    OP_CALL         // Dst = call(regs[Src1], argc=Src2)
    OP_CALL_BUILTIN // Dst = builtinDispatch(builtinIdx=Src1:Src2, argc=next instruction's Dst)
)
```

### VM data structures

```go
// Function represents a compiled bytecode function
type Function struct {
    Name      string
    Params    []string
    ParamTypes []TypeHint  // nil = untyped
    Code      []Instruction
    Constants []Value
    Upvalues  []UpvalueInfo
    MaxRegs   uint8
}

// VM is the execution engine
type VM struct {
    regs      []Value      // register file (pre-allocated, sliced per frame)
    frames    []Frame      // call stack
    jit       *JITController  // nil until Phase 2
    interp    *interpreter    // for builtins that need Go context
}

type Frame struct {
    fn      *Function
    pc      int
    baseReg int   // register window base in vm.regs
}
```

### VM inner loop

```go
func (vm *VM) run() Value {
    frame := &vm.frames[len(vm.frames)-1]
    regs  := vm.regs[frame.baseReg:]
    code  := frame.fn.Code

    for {
        instr := code[frame.pc]
        frame.pc++

        switch instr.Op {
        case OP_LOAD_CONST:
            idx := int(instr.Src1)<<8 | int(instr.Src2)
            regs[instr.Dst] = frame.fn.Constants[idx]

        case OP_ADD_INT:
            regs[instr.Dst] = regs[instr.Src1].(int) + regs[instr.Src2].(int)

        case OP_ADD:
            regs[instr.Dst] = genericAdd(regs[instr.Src1], regs[instr.Src2])

        case OP_JUMP_FALSE:
            if !IsTruthy(regs[instr.Src1]) {
                offset := int(int16(int(instr.Src1)<<8 | int(instr.Src2)))
                frame.pc += offset
            }

        case OP_CALL_BUILTIN:
            // fast path: no scope push, direct dispatch
            idx := int(instr.Src1)<<8 | int(instr.Src2)
            // args are in regs[instr.Dst+1 .. instr.Dst+argc]
            regs[instr.Dst] = vm.callBuiltin(idx, ...)

        case OP_RETURN:
            ret := regs[instr.Src1]
            vm.frames = vm.frames[:len(vm.frames)-1]
            if len(vm.frames) == 0 {
                return ret
            }
            // restore caller frame, put ret in caller's Dst register
            // ...
        }
    }
}
```

### New files

| File | Purpose |
|------|---------|
| `interpreter/opcode.go` | `OpCode` enum, `Instruction` struct, disassembler for debugging |
| `interpreter/compiler.go` | `Compiler` struct, `Compile(ast *Program) *Function` |
| `interpreter/vm.go` | `VM` struct, `run()` execution loop |
| `interpreter/vm_builtins.go` | `OP_CALL_BUILTIN` fast dispatch table, maps builtin index → Go func |

### Integration with existing Config

```go
type Config struct {
    // ... all existing fields unchanged ...
    VMEnabled bool  // default true after Phase 1 passes all existing tests
}
```

`VMEnabled = false` → existing `interpreter.go` tree-walker runs. Zero behavioral change for existing users during rollout.

### Optional type hint syntax (Phase 1.5)

Minimal parser change — `type` after `:` in parameter and return positions:

```
fn add(a: int, b: int): int
    return a + b
end

fn process(items: array): void
    // items guaranteed *[]Value — compiler can skip tag check
end
```

Supported primitive type hints: `int`, `float`, `string`, `bool`, `array`, `map`. Unannotated parameters remain `Value` (unchanged behavior).

Compiler emits `OP_ADD_INT` instead of `OP_ADD` when both operands are known `int` at compile time. No runtime cost for unannotated code.

### Estimated speedup Phase 1 alone

| Benchmark | Today | VM target | Speedup |
|-----------|-------|-----------|---------|
| NestedFunctionCalls | 870 µs | ~80 µs | ~10x |
| MapOperations | 897 µs | ~100 µs | ~9x |
| ComplexDataStructures | 1111 µs | ~130 µs | ~8x |
| StringConcatenation | 755 µs | ~90 µs | ~8x |

---

## Phase 2 — LLVM JIT

### Tiered compilation model

Same model as V8 (Ignition→TurboFan) and HotSpot:

- **Tier 0** (cold): VM interpreter — always correct, no compilation overhead
- **Tier 1** (hot): Function called ≥ 1000 times → JIT compile async → swap to native

Compilation is async (background goroutine). First 1000 calls use VM. After that, native code. No pause, no stop-the-world.

### Dependencies

```
github.com/llir/llvm    — pure Go, generate LLVM IR (no CGo)
libLLVM ≥ 17            — system library, linked via CGo for execution engine
```

CGo is isolated to one file (`jit/runtime.go`). IR generation is pure Go.

### Value representation in LLVM IR

For untyped code:

```llvm
; Tagged union: { i8 tag, i64 data }
; Tags: 0=nil, 1=int, 2=float64, 3=string-ptr, 4=bool, 5=array-ptr, 6=map-ptr
%Value = type { i8, i64 }
```

For type-annotated code (`fn add(a: int, b: int): int`):

```llvm
; No boxing — pure scalar arithmetic
define i64 @add(i64 %a, i64 %b) {
  %r = add i64 %a, %b
  ret i64 %r
}
```

### Profiling controller

```go
// NativeFunc is the calling convention for JIT-compiled functions.
// Matches Go's *[]Value-based Value type; args are passed as a pre-built slice.
type NativeFunc func(args []Value) Value

type JITController struct {
    callCounts sync.Map           // *Function → *int64 (atomic, per-invocation count)
    compiled   sync.Map           // *Function → NativeFunc
    threshold  int64              // default 1000 (total invocations, not unique call sites)
}

func (j *JITController) RecordCall(fn *Function) {
    ptr, _ := j.callCounts.LoadOrStore(fn, new(int64))
    count := atomic.AddInt64(ptr.(*int64), 1)
    if count == j.threshold {
        go j.compileAsync(fn)  // non-blocking
    }
}

func (vm *VM) callFunction(fn *Function, args []Value) Value {
    if vm.jit != nil {
        if native, ok := vm.jit.compiled.Load(fn); ok {
            return native.(NativeFunc)(args)  // call compiled code
        }
        vm.jit.RecordCall(fn)  // increment counter
    }
    return vm.callVM(fn, args)  // VM fallback
}
```

### Deoptimization

JIT-compiled code includes guard checks before typed operations. If a guard fails (unexpected type), control returns to VM:

```llvm
; Guard: expect int (tag == 1)
%tag = extractvalue %Value %a, 0
%ok  = icmp eq i8 %tag, 1
br i1 %ok, label %typed_path, label %deopt

deopt:
  ; call back into Go VM via exported C function
  %res = call %Value @vm_deopt_execute(ptr %frame, i32 %deopt_pc)
  ret %Value %res

typed_path:
  %ia = extractvalue %Value %a, 1  ; raw i64
  ; ... fast typed arithmetic
```

Guards are inline and branch-predicted as taken. Deopt is the cold path.

### New package structure

```
interpreter/
  jit/
    jit.go       — JITController: profiling, compilation trigger, native fn registry
    codegen.go   — Function (bytecode) → *ir.Module via llir/llvm
    runtime.go   — CGo: LLVM ExecutionEngine wrapper, load/execute native fn
    cache.go     — compiled function cache (content hash → NativeFunc)
    deopt.go     — deopt callback exported to LLVM IR as C function
```

### Built-in functions

Built-ins (`fn_network`, `fn_database`, `fn_file_system`, etc.) are **not JIT-compiled**. They already cross CGo boundaries for I/O. The `OP_CALL_BUILTIN` opcode in the VM dispatches them directly — no JIT overhead.

### Estimated speedup Phase 2

| Scenario | vs tree-walker today | vs Rust/Zig |
|----------|---------------------|-------------|
| Untyped numeric loops | ~30-50x | ~2-5x slower |
| Type-annotated numeric functions | ~100-200x | ~1-2x slower |
| Mixed typed/untyped | ~40-80x | ~2-4x slower |
| String-heavy (still boxing) | ~10-20x | ~5-15x slower |
| Built-in I/O calls | ~5x | ~10x slower (CGo overhead dominates) |

---

## Implementation Sequence

```
Phase 0 (1-2 days)
  ├── Fix global factDatabase/eventStore → per-interpreter
  ├── Fix evalPlus pool misuse
  ├── Fix ensureIntToFloats false-zero sentinel
  ├── Fix safeValueEqual → type switch
  ├── Remove dead dispatch code in evaluate()
  ├── Standardize stability labels
  └── Default TaggedValues to false

Phase 1 (4-8 weeks)
  ├── opcode.go — instruction set + disassembler
  ├── compiler.go — AST → bytecode
  ├── vm.go — execution loop
  ├── vm_builtins.go — builtin dispatch table
  ├── Config.VMEnabled flag + fallback
  ├── All existing tests pass with VMEnabled=true
  └── Phase 1.5: optional type hint parser + typed opcodes

Phase 2 (8-16 weeks)
  ├── jit/jit.go — profiling controller
  ├── jit/codegen.go — bytecode → LLVM IR
  ├── jit/runtime.go — CGo LLVM execution engine
  ├── jit/cache.go — compiled function cache
  ├── jit/deopt.go — deoptimization callbacks
  └── Benchmark-driven tuning of JIT threshold
```

Each phase is independently valuable and shippable.

---

## What Is Not In Scope

- Garbage collector changes (Go GC handles heap; VM stack is slice-based)
- Parallel GC or region-based allocation
- Cross-compilation to WASM (separate initiative if needed)
- Bytecode serialization / `.dinc` compiled cache files (future, after Phase 1 stable)
- Type inference (only explicit optional annotations; no Hindley-Milner)
