---
id: bytecode-vm
title: Bytecode VM & Performance
sidebar_label: Bytecode VM
---

# Bytecode VM & Performance

Uddin-Lang uses a **register-based bytecode virtual machine** as its default execution engine. This page explains the VM architecture, available opcodes, and the performance APIs introduced in Phase 2.

## Architecture Overview

```
Source (.din)
    │
    ▼  Tokenizer + Parser
  AST (Abstract Syntax Tree)
    │
    ▼  Compiler
  Bytecode ([]Instruction)
    │
    ▼  VM (register-based)
  Result
```

### Why a Bytecode VM?

The previous tree-walker interpreter recursively traversed the AST on every execution. The bytecode VM compiles the AST once into a compact array of 4-byte instructions, then executes them in a tight loop. This eliminates recursive function call overhead and improves CPU cache locality.

**Typical speedup vs. tree-walker:** 5–15× on numeric/loop-heavy scripts.

### VM Design

- **Register-based** — each frame has a fixed register array (`regs []Value`). No operand stack.
- **4-byte instructions** — each `Instruction` is `{Op uint8, Dst uint8, Src1 uint8, Src2 uint8}`.
- **Frame stack** — function calls push a `vmFrame`; returns pop it.
- **Try stack** — exception handlers use a separate `tryStack` for `OP_TRY`/`OP_END_TRY`.

### Switching Between Engines

```go
// Default: VM enabled
config := uddin.DefaultConfig()
config.VMEnabled = true   // default

// Use tree-walker (legacy)
config.VMEnabled = false
```

---

## Opcode Reference

All opcodes fit in a `uint8`. Instructions are 4 bytes: `Op | Dst | Src1 | Src2`.

### Load / Store

| Opcode | Encoding | Description |
|--------|----------|-------------|
| `LOAD_CONST` | `Dst = Constants[Src1<<8\|Src2]` | Load constant from pool into register |
| `LOAD_VAR` | `Dst = locals[Src1]` | Load local variable into register |
| `STORE_VAR` | `locals[Dst] = regs[Src1]` | Store register into local variable |
| `LOAD_UPVAL` | `Dst = closure.Upvalues[Src1]` | Load captured variable (closure) |
| `STORE_UPVAL` | `closure.Upvalues[Dst] = regs[Src1]` | Store into captured variable |

### Arithmetic — Generic (any Value)

| Opcode | Operation |
|--------|-----------|
| `ADD` | `regs[Dst] = regs[Src1] + regs[Src2]` |
| `SUB` | `regs[Dst] = regs[Src1] - regs[Src2]` |
| `MUL` | `regs[Dst] = regs[Src1] * regs[Src2]` |
| `DIV` | `regs[Dst] = regs[Src1] / regs[Src2]` |
| `MOD` | `regs[Dst] = regs[Src1] % regs[Src2]` |
| `POW` | `regs[Dst] = regs[Src1] ** regs[Src2]` |
| `NEG` | `regs[Dst] = -regs[Src1]` |

### Arithmetic — Typed (int operands, no boxing)

| Opcode | Operation |
|--------|-----------|
| `ADD_INT` | Integer add, skips type assertion overhead |
| `SUB_INT` | Integer subtract |
| `MUL_INT` | Integer multiply |
| `DIV_INT` | Integer divide |
| `MOD_INT` | Integer modulo |

The compiler emits typed variants when both operands are known integers at compile time.

### Comparison

| Opcode | Operation |
|--------|-----------|
| `EQ` | `regs[Dst] = regs[Src1] == regs[Src2]` |
| `NEQ` | `regs[Dst] = regs[Src1] != regs[Src2]` |
| `LT` | Less than |
| `LTE` | Less than or equal |
| `GT` | Greater than |
| `GTE` | Greater than or equal |
| `IN` | Membership test (`in` operator) |

### Logic

| Opcode | Operation |
|--------|-----------|
| `AND` | Logical and (short-circuit) |
| `OR` | Logical or (short-circuit) |
| `NOT` | Logical not |
| `XOR` | Logical exclusive or |

### Control Flow

| Opcode | Encoding | Description |
|--------|----------|-------------|
| `JUMP` | `pc += int16(Dst<<8\|Src2)` | Unconditional relative jump |
| `JUMP_FALSE` | Jump if `!IsTruthy(regs[Src1])` | Conditional branch (false) |
| `JUMP_TRUE` | Jump if `IsTruthy(regs[Src1])` | Conditional branch (true) |
| `RETURN` | Return `regs[Src1]` | Return from current function |

Jump offsets are signed 16-bit integers encoded in `Dst:Src2` fields.

### Collections

| Opcode | Description |
|--------|-------------|
| `MAKE_ARRAY` | Build `[]Value` from `Src2` registers starting at `Src1` |
| `MAKE_MAP` | Build `map[string]Value` from `Src2` key-value pairs starting at `Src1` |
| `SUBSCRIPT` | `regs[Dst] = regs[Src1][regs[Src2]]` — array/map index |
| `SET_INDEX` | `regs[Dst][regs[Src1]] = regs[Src2]` — array/map assignment |

### Functions

| Opcode | Description |
|--------|-------------|
| `MAKE_FUNC` | Create closure from `fn.SubFunctions[Src1<<8\|Src2]`, capture upvalues |
| `CALL` | Call `regs[Src1]` with `Src2` args; result in `regs[Dst]` |
| `CALL_BUILTIN` | Call builtin via `vmBuiltinTable[Src1<<8\|Src2]` (dispatcher chain) |
| `CALL_BUILTIN_DIRECT` | Call builtin via `vmBuiltinDirectTable[Src1<<8\|Src2]` (direct pointer) |

### Exception Handling

| Opcode | Encoding | Description |
|--------|----------|-------------|
| `TRY` | `Src1=errReg; Dst:Src2=jump offset to catch block` | Push try handler |
| `END_TRY` | — | Pop innermost try handler (no error occurred) |

---

## `CALL_BUILTIN` vs `CALL_BUILTIN_DIRECT`

Uddin-Lang has two builtin call paths:

**`CALL_BUILTIN`** — routes through the 4-layer `DispatchOrPanic` chain:
```
OP_CALL_BUILTIN → vmBuiltinTable[idx].fn → DispatchOrPanic → metadata lookup → actual function
```

**`CALL_BUILTIN_DIRECT`** (Phase 2B) — direct function pointer, no dispatcher:
```
OP_CALL_BUILTIN_DIRECT → vmBuiltinDirectTable[idx](interp, pos, args)
```

The compiler automatically emits `CALL_BUILTIN_DIRECT` for builtins with a direct implementation. Supported builtins:

`len`, `typeof`, `upper`, `lower`, `trim`, `contains`, `starts_with`, `ends_with`, `split`, `join`, `int`, `str`, `append`, `waf_cidr_match`, `waf_path_match`

---

## Disassembling Bytecode

Use `interpreter.Disassemble()` to inspect compiled bytecode:

```go
prog, _ := uddin.ParseProgram([]byte(source))
fn, _ := interpreter.NewCompiler().Compile(prog)
fmt.Println(interpreter.Disassemble(fn))
```

Example output:
```
function "main"  regs=4  consts=2
  0000  LOAD_CONST      dst=0   src1=0   src2=0    ; 1
  0001  LOAD_CONST      dst=1   src1=0   src2=1    ; 2
  0002  ADD_INT         dst=2   src1=0   src2=1
  0003  STORE_VAR       dst=3   src1=2   src2=0
  0004  RETURN          dst=0   src1=3   src2=0
```

---

## Engine Performance API (Phase 2A)

For embedding uddin-lang in Go services with high request throughput, the `Engine` struct provides compile-once, run-many execution.

### `Engine.ExecuteProgram` with Cache

```go
engine := uddin.New(config)

// First call: compiles AST → bytecode, stores in programCache
stats, err := engine.ExecuteProgram(prog)

// Subsequent calls with the same *Program: uses cached bytecode
stats, err = engine.ExecuteProgram(prog)
```

The cache key is the `*Program` pointer. Compile the same source into the same `*Program` object and reuse it across calls.

### VM Pool (`sync.Pool`)

`Engine` maintains a `sync.Pool` of `*VM` instances. Each call to `ExecuteProgram` gets a VM from the pool, resets it (clears frame stack, rebinds interpreter), and returns it after execution. This eliminates per-request VM allocation overhead.

### `CompiledProgram` — Low-Level API

For maximum control, compile and cache manually:

```go
// Compile once, providing variable names for pre-allocation
varNames := []string{"request", "config"}
sort.Strings(varNames)
cp, err := interpreter.CompileProgram(prog, varNames)

// Execute with a new config on each call
stats, err := interpreter.ExecuteCompiledVM(cp, config, nil)

// Or reuse a VM across calls (advanced)
vm := interpreter.NewVM(nil)
stats, err = interpreter.ExecuteCompiledVM(cp, config, vm)
```

### Regex Pre-compilation

String literals passed as regex patterns are compiled to `*coregex.Regexp` at bytecode compile time. At runtime, no string-to-regex compilation occurs:

```din
// Pattern "^\d+$" is pre-compiled once during bytecode compilation
if regex.is_match("^\d+$", input):
    ...
end
```

This is automatic — no code changes needed.

---

## Performance Summary

| Feature | Gain |
|---------|------|
| Bytecode VM vs tree-walker | 5–15× on numeric/loop-heavy code |
| `Engine.programCache` | Eliminates recompile cost on repeated calls |
| `Engine.vmPool` | Eliminates per-call VM allocation |
| `CALL_BUILTIN_DIRECT` | ~20–30% reduction in hot builtin call overhead |
| Regex pre-compilation | Eliminates string→regex parse on every call |
