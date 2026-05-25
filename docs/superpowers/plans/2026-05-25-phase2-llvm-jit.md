# Phase 2 — LLVM JIT Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a tiered JIT compilation layer on top of the Phase 1 bytecode VM. Functions called ≥1000 times are compiled to native code via LLVM IR, achieving 30-200x speedup over the original tree-walker.

**Architecture:** VM runs functions in Tier 0 (interpreter loop). A profiling controller counts invocations per `*Function`. At threshold, an async goroutine generates LLVM IR via `llir/llvm`, compiles it via libLLVM's ExecutionEngine (CGo), and caches the native function pointer. Subsequent calls bypass the VM and call native code directly. On unexpected type, native code deoptimizes back to VM.

**Tech Stack:** Go 1.25.4, `github.com/llir/llvm` (pure Go IR generation), `libLLVM ≥ 17` (system library, CGo for JIT execution), CGo (one file only).

**Prerequisite:** Phase 1 complete. All existing tests passing with `VMEnabled = true`. Bytecode VM benchmarks confirmed ≥5x vs tree-walker.

**System requirement:** `llvm-17-dev` (or later) installed. On macOS: `brew install llvm`. On Ubuntu: `apt-get install llvm-17-dev`.

---

## File Map

| File | Status | Purpose |
|------|--------|---------|
| `interpreter/jit/jit.go` | Create | `JITController`: profiling counter, compile trigger, native fn registry |
| `interpreter/jit/codegen.go` | Create | `*Function` (bytecode) → `*ir.Module` (LLVM IR) via `llir/llvm` |
| `interpreter/jit/runtime.go` | Create | CGo wrapper: LLVM `ExecutionEngine`, load + execute compiled module |
| `interpreter/jit/cache.go` | Create | Content-hash → `NativeFunc` cache with concurrent-safe access |
| `interpreter/jit/deopt.go` | Create | Deopt callback exported to C, called from JIT code on type mismatch |
| `interpreter/vm.go` | Modify | Check `JITController` before entering VM loop; call native if available |
| `interpreter/config.go` | Modify | Add `JITEnabled bool`, `JITThreshold int64` fields |

---

## Task 1: Dependencies and Build Setup

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add llir/llvm dependency**

```bash
go get github.com/llir/llvm@latest
go mod tidy
```

- [ ] **Step 2: Verify LLVM system library**

```bash
llvm-config --version        # must be ≥ 17
llvm-config --ldflags        # must show lib path
llvm-config --libs           # lists required libs
```

- [ ] **Step 3: Create jit/ package directory**

```bash
mkdir -p interpreter/jit
```

- [ ] **Step 4: Create a minimal build probe**

Create `interpreter/jit/jit.go` with just the package declaration and a build tag:

```go
//go:build cgo

package jit

// Package jit provides LLVM JIT compilation for hot bytecode functions.
// Requires libLLVM ≥ 17. Disabled when CGO_ENABLED=0.
```

- [ ] **Step 5: Verify build succeeds**

```bash
go build ./interpreter/jit/...
go test ./... 2>&1 | tail -5
```

Expected: compiles cleanly.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum interpreter/jit/jit.go
git commit -m "feat(jit): add llir/llvm dependency and jit/ package scaffold"
```

---

## Task 2: JITController — Profiling and Compilation Trigger

**Files:**
- Modify: `interpreter/jit/jit.go`
- Create: `interpreter/jit/jit_test.go`

- [ ] **Step 1: Write test**

Create `interpreter/jit/jit_test.go`:

```go
//go:build cgo

package jit

import (
	"sync/atomic"
	"testing"
)

func TestJITControllerThreshold(t *testing.T) {
	ctrl := NewJITController(10) // threshold = 10

	key := uintptr(0xDEADBEEF) // fake function pointer as key
	compiled := false

	ctrl.SetCompileFunc(func(fnKey uintptr) {
		compiled = true
	})

	for i := 0; i < 9; i++ {
		ctrl.RecordCall(key)
	}
	if compiled {
		t.Error("should not compile before threshold")
	}

	ctrl.RecordCall(key) // 10th call — triggers compile
	// compileAsync runs in goroutine; wait briefly
	ctrl.WaitForPendingCompiles()
	if !compiled {
		t.Error("should have compiled at threshold")
	}
}

func TestJITControllerNativeRegistration(t *testing.T) {
	ctrl := NewJITController(1)

	key := uintptr(0xCAFEBABE)
	called := false
	ctrl.RegisterNative(key, func(args []interface{}) interface{} {
		called = true
		return 42
	})

	native, ok := ctrl.GetNative(key)
	if !ok {
		t.Fatal("expected native func to be registered")
	}
	result := native(nil)
	if result != 42 || !called {
		t.Error("native func not called correctly")
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./interpreter/jit/... -run "TestJITController" -v
```

Expected: FAIL — types not defined.

- [ ] **Step 3: Implement JITController**

Complete `interpreter/jit/jit.go`:

```go
//go:build cgo

package jit

import (
	"sync"
	"sync/atomic"
)

// NativeFunc is the calling convention for JIT-compiled functions.
type NativeFunc func(args []interface{}) interface{}

// JITController manages profiling, compilation triggers, and native function registry.
type JITController struct {
	callCounts   sync.Map          // uintptr(fn ptr) → *int64
	compiled     sync.Map          // uintptr(fn ptr) → NativeFunc
	threshold    int64
	compileFunc  func(uintptr)    // set by the codegen layer
	pending      sync.WaitGroup
	mu           sync.Mutex
}

// NewJITController creates a controller with the given invocation threshold.
func NewJITController(threshold int64) *JITController {
	return &JITController{threshold: threshold}
}

// SetCompileFunc registers the function that will JIT-compile a given fn key.
func (c *JITController) SetCompileFunc(f func(uintptr)) {
	c.mu.Lock()
	c.compileFunc = f
	c.mu.Unlock()
}

// RecordCall increments the invocation count for fnKey.
// When count reaches threshold, triggers async compilation.
func (c *JITController) RecordCall(fnKey uintptr) {
	ptr, _ := c.callCounts.LoadOrStore(fnKey, new(int64))
	count := atomic.AddInt64(ptr.(*int64), 1)
	if count == c.threshold {
		c.mu.Lock()
		cf := c.compileFunc
		c.mu.Unlock()
		if cf != nil {
			c.pending.Add(1)
			go func() {
				defer c.pending.Done()
				cf(fnKey)
			}()
		}
	}
}

// RegisterNative caches a JIT-compiled function for fnKey.
func (c *JITController) RegisterNative(fnKey uintptr, fn NativeFunc) {
	c.compiled.Store(fnKey, fn)
}

// GetNative returns the JIT-compiled function for fnKey, if available.
func (c *JITController) GetNative(fnKey uintptr) (NativeFunc, bool) {
	v, ok := c.compiled.Load(fnKey)
	if !ok {
		return nil, false
	}
	return v.(NativeFunc), true
}

// WaitForPendingCompiles blocks until all async compilations finish. Used in tests.
func (c *JITController) WaitForPendingCompiles() {
	c.pending.Wait()
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./interpreter/jit/... -run "TestJITController" -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interpreter/jit/jit.go interpreter/jit/jit_test.go
git commit -m "feat(jit): JITController with threshold-based async compile trigger"
```

---

## Task 3: Content-Hash Cache

**Files:**
- Create: `interpreter/jit/cache.go`
- Create: `interpreter/jit/cache_test.go`

The cache maps a hash of a `*Function`'s bytecode to a `NativeFunc`. This allows reusing compiled code if the same function source is compiled in multiple interpreter instances.

- [ ] **Step 1: Write test**

```go
//go:build cgo

package jit

import "testing"

func TestCacheStoreAndLoad(t *testing.T) {
	c := NewFunctionCache()
	hash := uint64(0xABCDABCD)
	called := false
	fn := NativeFunc(func(args []interface{}) interface{} {
		called = true
		return 1
	})
	c.Store(hash, fn)
	got, ok := c.Load(hash)
	if !ok {
		t.Fatal("expected cache hit")
	}
	got(nil)
	if !called {
		t.Error("stored function was not called")
	}
}
```

- [ ] **Step 2: Implement cache.go**

```go
//go:build cgo

package jit

import "sync"

// FunctionCache is a concurrent hash → NativeFunc lookup table.
type FunctionCache struct {
	m sync.Map
}

// NewFunctionCache creates an empty cache.
func NewFunctionCache() *FunctionCache {
	return &FunctionCache{}
}

// Store stores a compiled function under the given hash.
func (c *FunctionCache) Store(hash uint64, fn NativeFunc) {
	c.m.Store(hash, fn)
}

// Load returns the compiled function for hash, if present.
func (c *FunctionCache) Load(hash uint64) (NativeFunc, bool) {
	v, ok := c.m.Load(hash)
	if !ok {
		return nil, false
	}
	return v.(NativeFunc), true
}

// HashFunction computes a stable hash of a bytecode Function's instructions.
// Same bytecode → same hash. Collision probability negligible for normal program sizes.
func HashFunction(code []Instruction) uint64 {
	var h uint64 = 14695981039346656037 // FNV-1a offset basis
	for _, instr := range code {
		h ^= uint64(instr.Op)
		h *= 1099511628211
		h ^= uint64(instr.Dst)
		h *= 1099511628211
		h ^= uint64(instr.Src1)
		h *= 1099511628211
		h ^= uint64(instr.Src2)
		h *= 1099511628211
	}
	return h
}
```

Note: `Instruction` type lives in `interpreter` package. The `jit` package imports it as `interpreter.Instruction`. Adjust the import accordingly.

- [ ] **Step 3: Run tests**

```bash
go test ./interpreter/jit/... -run "TestCache" -v
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add interpreter/jit/cache.go interpreter/jit/cache_test.go
git commit -m "feat(jit): FunctionCache with FNV-1a bytecode hash"
```

---

## Task 4: Deoptimization Callback

**Files:**
- Create: `interpreter/jit/deopt.go`

The deopt callback is a C-exported Go function. JIT-compiled code calls it when a type guard fails. It hands control back to the VM to finish executing the function from the deopt point.

- [ ] **Step 1: Create deopt.go**

```go
//go:build cgo

package jit

/*
#include <stdint.h>

// DeoptFrame holds context for resuming the VM after deoptimization.
typedef struct {
    void* fn_ptr;    // pointer to the bytecode Function
    int32_t pc;      // instruction counter at deopt point
} DeoptFrame;

// vm_deopt_execute is called from JIT code on type guard failure.
// Implemented in Go below (exported via //export).
extern void* vm_deopt_execute(DeoptFrame* frame, int argc, void** args);
*/
import "C"
import "unsafe"

// deoptHandler is set by the VM layer to handle deopt requests.
var deoptHandler func(fnPtr unsafe.Pointer, pc int, args []interface{}) interface{}

// SetDeoptHandler registers the VM callback for deoptimization.
func SetDeoptHandler(fn func(fnPtr unsafe.Pointer, pc int, args []interface{}) interface{}) {
	deoptHandler = fn
}

//export vm_deopt_execute
func vm_deopt_execute(frame *C.DeoptFrame, argc C.int, args **C.void) unsafe.Pointer {
	if deoptHandler == nil {
		return nil
	}
	goArgs := make([]interface{}, int(argc))
	// Convert C void** args to Go slice
	argSlice := (*[1 << 20]*C.void)(unsafe.Pointer(args))[:int(argc):int(argc)]
	for i, arg := range argSlice {
		goArgs[i] = unsafe.Pointer(arg)
	}
	result := deoptHandler(unsafe.Pointer(frame.fn_ptr), int(frame.pc), goArgs)
	// Return as void* — caller casts to Value representation
	return unsafe.Pointer(&result)
}
```

- [ ] **Step 2: Build check**

```bash
go build ./interpreter/jit/...
```

Expected: compiles without CGo errors.

- [ ] **Step 3: Commit**

```bash
git add interpreter/jit/deopt.go
git commit -m "feat(jit): deopt callback exported to C for JIT guard failures"
```

---

## Task 5: LLVM IR Code Generation

**Files:**
- Create: `interpreter/jit/codegen.go`
- Create: `interpreter/jit/codegen_test.go`

This is the most complex task. The codegen translates bytecode `*Function` instructions into an `*ir.Module` using `llir/llvm`.

- [ ] **Step 1: Write test**

```go
//go:build cgo

package jit

import (
	"testing"

	"github.com/llir/llvm/ir"
)

func TestCodegenProducesModule(t *testing.T) {
	// Build a simple function: fn add(a, b) return a + b end
	// as bytecode manually
	fn := &Function{
		Name:    "add",
		Params:  []string{"a", "b"},
		MaxRegs: 4,
		Constants: []Value{nil},
		Code: []Instruction{
			{Op: OP_LOAD_VAR, Dst: 2, Src1: 0},         // r2 = a (reg 0)
			{Op: OP_LOAD_VAR, Dst: 3, Src1: 1},         // r3 = b (reg 1)
			{Op: OP_ADD_INT, Dst: 0, Src1: 2, Src2: 3}, // r0 = r2 + r3
			{Op: OP_RETURN, Src1: 0},
		},
	}

	cg := NewCodegen()
	mod, err := cg.GenerateModule(fn)
	if err != nil {
		t.Fatal(err)
	}
	if mod == nil {
		t.Fatal("expected non-nil ir.Module")
	}
	// Module must have at least one function definition
	if len(mod.Funcs) == 0 {
		t.Error("expected at least one function in module")
	}
}
```

Note: `Function`, `Instruction`, `Value`, `OpCode` constants (`OP_LOAD_VAR`, etc.) come from the `interpreter` package — the `jit` package must import them. Adjust package prefix accordingly.

- [ ] **Step 2: Implement codegen.go**

```go
//go:build cgo

package jit

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	irtypes "github.com/llir/llvm/ir/types"
)

// Codegen translates bytecode Functions to LLVM IR modules.
type Codegen struct{}

// NewCodegen creates a code generator.
func NewCodegen() *Codegen { return &Codegen{} }

// boxedValueType is the LLVM IR type for an untyped Value: { i8 tag, i64 data }.
var boxedValueType = irtypes.NewStruct(irtypes.I8, irtypes.I64)

// GenerateModule compiles fn to an LLVM IR module.
// Returns the module and any generation error.
func (cg *Codegen) GenerateModule(fn *Function) (*ir.Module, error) {
	mod := ir.NewModule()

	// Determine function signature.
	// Typed functions (all params are int) use i64 args.
	// Untyped functions use { i8, i64 } (boxed Value) args.
	allInt := len(fn.ParamTypes) == len(fn.Params)
	if allInt {
		for _, pt := range fn.ParamTypes {
			if pt != TypeInt {
				allInt = false
				break
			}
		}
	}

	if allInt && fn.ReturnType == TypeInt {
		return cg.generateTypedIntFunction(mod, fn)
	}
	return cg.generateBoxedFunction(mod, fn)
}

// generateTypedIntFunction generates a pure i64 → i64 function (no boxing).
func (cg *Codegen) generateTypedIntFunction(mod *ir.Module, fn *Function) (*ir.Module, error) {
	// Build parameter list
	params := make([]*ir.Param, len(fn.Params))
	for i, name := range fn.Params {
		params[i] = ir.NewParam(name, irtypes.I64)
	}

	irFn := mod.NewFunc(fn.Name, irtypes.I64, params...)
	entry := irFn.NewBlock("entry")

	// Register map: bytecode reg → LLVM value
	regs := make(map[uint8]ir.Value)
	for i, p := range params {
		regs[uint8(i)] = p
	}

	var err error
	for _, instr := range fn.Code {
		if err = cg.emitTypedInstr(entry, instr, regs, fn, irFn); err != nil {
			return nil, err
		}
	}

	return mod, nil
}

func (cg *Codegen) emitTypedInstr(block *ir.Block, instr Instruction, regs map[uint8]ir.Value, fn *Function, irFn *ir.Func) error {
	switch instr.Op {
	case OP_LOAD_VAR:
		src, ok := regs[instr.Src1]
		if !ok {
			return fmt.Errorf("codegen: LOAD_VAR: register %d not defined", instr.Src1)
		}
		regs[instr.Dst] = src

	case OP_LOAD_CONST:
		idx := int(instr.Src1)<<8 | int(instr.Src2)
		v := fn.Constants[idx]
		switch cv := v.(type) {
		case int:
			regs[instr.Dst] = constant.NewInt(irtypes.I64, int64(cv))
		case float64:
			regs[instr.Dst] = constant.NewFloat(irtypes.Double, cv)
		default:
			return fmt.Errorf("codegen: LOAD_CONST: unsupported constant type %T", v)
		}

	case OP_ADD_INT:
		l, r := regs[instr.Src1], regs[instr.Src2]
		regs[instr.Dst] = block.NewAdd(l, r)

	case OP_SUB_INT:
		l, r := regs[instr.Src1], regs[instr.Src2]
		regs[instr.Dst] = block.NewSub(l, r)

	case OP_MUL_INT:
		l, r := regs[instr.Src1], regs[instr.Src2]
		regs[instr.Dst] = block.NewMul(l, r)

	case OP_DIV_INT:
		l, r := regs[instr.Src1], regs[instr.Src2]
		regs[instr.Dst] = block.NewSDiv(l, r)

	case OP_MOD_INT:
		l, r := regs[instr.Src1], regs[instr.Src2]
		regs[instr.Dst] = block.NewSRem(l, r)

	case OP_RETURN:
		src, ok := regs[instr.Src1]
		if !ok {
			src = constant.NewInt(irtypes.I64, 0)
		}
		block.NewRet(src)

	default:
		return fmt.Errorf("codegen: unsupported opcode %s in typed function", opName(instr.Op))
	}
	return nil
}

// generateBoxedFunction generates a function using the tagged-union Value type.
// Untyped functions use deopt guards for each arithmetic operation.
func (cg *Codegen) generateBoxedFunction(mod *ir.Module, fn *Function) (*ir.Module, error) {
	// This is the full deopt-guarded path.
	// For brevity, a minimal implementation: emit a stub that returns nil (i64 0).
	// Full implementation expands each opcode with tag checks and deopt branches.
	params := make([]*ir.Param, len(fn.Params))
	for i, name := range fn.Params {
		params[i] = ir.NewParam(name, boxedValueType)
	}
	irFn := mod.NewFunc(fn.Name, boxedValueType, params...)
	entry := irFn.NewBlock("entry")
	// Return zero-tagged nil value as stub
	nilVal := constant.NewStruct(boxedValueType,
		constant.NewInt(irtypes.I8, 0),  // tag = nil
		constant.NewInt(irtypes.I64, 0), // data = 0
	)
	entry.NewRet(nilVal)
	_ = enum.IPredEQ // suppress unused import
	return mod, nil
}
```

**Note:** The boxed function implementation is intentionally minimal (stub). Full deopt-guarded codegen for all opcodes is an extension task — implement incrementally per opcode, same TDD pattern as Phase 1. Start with the typed path which is already complete above.

- [ ] **Step 3: Run test**

```bash
go test ./interpreter/jit/... -run "TestCodegenProducesModule" -v
```

Expected: PASS — module generated with at least one function.

- [ ] **Step 4: Commit**

```bash
git add interpreter/jit/codegen.go interpreter/jit/codegen_test.go
git commit -m "feat(jit): LLVM IR codegen for typed int functions via llir/llvm"
```

---

## Task 6: LLVM Execution Engine (CGo Runtime)

**Files:**
- Create: `interpreter/jit/runtime.go`
- Create: `interpreter/jit/runtime_test.go`

This task requires libLLVM to be installed. Uses CGo to call `LLVMCreateExecutionEngineForModule`.

- [ ] **Step 1: Write smoke test**

```go
//go:build cgo

package jit

import "testing"

func TestRuntimeCompileAndRun(t *testing.T) {
	// Build a trivial typed function: fn id(x: int): int return x end
	fn := &Function{
		Name:       "id",
		Params:     []string{"x"},
		ParamTypes: []TypeHint{TypeInt},
		ReturnType: TypeInt,
		MaxRegs:    2,
		Constants:  []Value{nil},
		Code: []Instruction{
			{Op: OP_LOAD_VAR, Dst: 1, Src1: 0},
			{Op: OP_RETURN, Src1: 1},
		},
	}

	cg := NewCodegen()
	mod, err := cg.GenerateModule(fn)
	if err != nil {
		t.Fatal(err)
	}

	rt := NewRuntime()
	native, err := rt.CompileModule(mod, fn.Name)
	if err != nil {
		t.Skipf("LLVM runtime not available: %v", err)
	}

	result := native([]interface{}{int64(42)})
	if result != int64(42) {
		t.Errorf("expected 42, got %v", result)
	}
}
```

- [ ] **Step 2: Create runtime.go**

```go
//go:build cgo

package jit

/*
#cgo CFLAGS: -I/usr/include/llvm-17
#cgo LDFLAGS: -L/usr/lib/llvm-17/lib -lLLVM-17
// Adjust paths for macOS: use output of `llvm-config --cflags --ldflags`

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <string.h>
#include <stdlib.h>

static void initLLVM() {
    LLVMInitializeNativeTarget();
    LLVMInitializeNativeAsmPrinter();
    LLVMLinkInMCJIT();
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/llir/llvm/ir"
)

func init() {
	C.initLLVM()
}

// Runtime wraps the LLVM Execution Engine.
type Runtime struct{}

// NewRuntime creates an LLVM JIT runtime.
func NewRuntime() *Runtime { return &Runtime{} }

// CompileModule JIT-compiles an ir.Module and returns a NativeFunc for fnName.
func (r *Runtime) CompileModule(mod *ir.Module, fnName string) (NativeFunc, error) {
	// Serialize IR to bitcode string
	irStr := mod.String()
	cStr := C.CString(irStr)
	defer C.free(unsafe.Pointer(cStr))

	// Parse IR string into LLVMModuleRef
	var llvmMod C.LLVMModuleRef
	var errMsg *C.char
	buf := C.LLVMCreateMemoryBufferWithMemoryRangeCopy(
		(*C.char)(unsafe.Pointer(cStr)),
		C.size_t(len(irStr)),
		C.CString("<jit>"),
	)
	if C.LLVMParseIRInContext(C.LLVMGetGlobalContext(), buf, &llvmMod, &errMsg) != 0 {
		msg := C.GoString(errMsg)
		C.LLVMDisposeMessage(errMsg)
		return nil, fmt.Errorf("LLVM IR parse error: %s", msg)
	}

	// Create MCJIT execution engine
	var engine C.LLVMExecutionEngineRef
	var options C.struct_LLVMMCJITCompilerOptions
	C.LLVMInitializeMCJITCompilerOptions(&options, C.size_t(unsafe.Sizeof(options)))
	options.OptLevel = 2 // -O2

	if C.LLVMCreateMCJITCompilerForModule(&engine, llvmMod, &options,
		C.size_t(unsafe.Sizeof(options)), &errMsg) != 0 {
		msg := C.GoString(errMsg)
		C.LLVMDisposeMessage(errMsg)
		return nil, fmt.Errorf("LLVM engine creation failed: %s", msg)
	}

	// Get function pointer
	cFnName := C.CString(fnName)
	defer C.free(unsafe.Pointer(cFnName))
	fnPtr := C.LLVMGetFunctionAddress(engine, cFnName)
	if fnPtr == 0 {
		return nil, fmt.Errorf("LLVM: function %q not found in module", fnName)
	}

	// Wrap as NativeFunc.
	// LLVMGetFunctionAddress returns uint64_t (the absolute address of the compiled fn).
	// Convert to a callable Go function via uintptr → unsafe.Pointer → func pointer cast.
	fnAddr := uintptr(fnPtr)
	nativeFn := func(args []interface{}) interface{} {
		type intFn func(int64) int64
		f := *(*intFn)(unsafe.Pointer(&fnAddr))
		if len(args) > 0 {
			return f(args[0].(int64))
		}
		return f(0)
	}
	return nativeFn, nil
}
```

**Note:** The CGo flags (`-I`, `-L`, `-l`) must be adjusted based on the target system. Use `llvm-config --cflags` and `llvm-config --ldflags --libs` to get the correct values. Consider a `Makefile` target or `pkg-config` integration for portability.

- [ ] **Step 3: Run test (requires libLLVM installed)**

```bash
CGO_ENABLED=1 go test ./interpreter/jit/... -run "TestRuntimeCompileAndRun" -v
```

Expected: PASS if libLLVM-17 installed; `t.Skip` if not available.

- [ ] **Step 4: Commit**

```bash
git add interpreter/jit/runtime.go interpreter/jit/runtime_test.go
git commit -m "feat(jit): CGo LLVM ExecutionEngine wrapper for native code compilation"
```

---

## Task 7: Wire JIT into VM Dispatch

**Files:**
- Modify: `interpreter/vm.go`
- Modify: `interpreter/config.go`

- [ ] **Step 1: Add JIT config fields**

In `interpreter/config.go`:

```go
type Config struct {
	// ...
	JITEnabled   bool  // default false during Phase 2 rollout
	JITThreshold int64 // invocations before JIT compile, default 1000
}
```

In `DefaultConfig()`: `JITEnabled: false, JITThreshold: 1000`.

- [ ] **Step 2: Add JITController pointer to VM**

In `interpreter/vm.go`, update `VM` struct:

```go
type VM struct {
	regs   []Value
	frames []Frame
	interp *interpreter
	jit    *jit.JITController // nil when JITEnabled=false
	jitRT  *jit.Runtime
	jitCG  *jit.Codegen
	jitCache *jit.FunctionCache
}
```

- [ ] **Step 3: Initialize JIT in NewVM**

```go
func NewVM(interp *interpreter, config *Config) *VM {
	vm := &VM{
		regs:   make([]Value, vmInitialRegCount),
		frames: make([]Frame, 0, 32),
		interp: interp,
	}
	if config.JITEnabled {
		threshold := config.JITThreshold
		if threshold <= 0 {
			threshold = 1000
		}
		ctrl := jit.NewJITController(threshold)
		cg := jit.NewCodegen()
		rt := jit.NewRuntime()
		cache := jit.NewFunctionCache()

		ctrl.SetCompileFunc(func(fnKey uintptr) {
			fn := (*Function)(unsafe.Pointer(fnKey))
			hash := jit.HashFunction(fn.Code)
			if _, ok := cache.Load(hash); ok {
				ctrl.RegisterNative(fnKey, func(args []interface{}) interface{} {
					native, _ := cache.Load(hash)
					return native(args)
				})
				return
			}
			mod, err := cg.GenerateModule(fn)
			if err != nil {
				return // silent: fall back to VM
			}
			native, err := rt.CompileModule(mod, fn.Name)
			if err != nil {
				return
			}
			cache.Store(hash, native)
			ctrl.RegisterNative(fnKey, native)
		})

		vm.jit = ctrl
		vm.jitRT = rt
		vm.jitCG = cg
		vm.jitCache = cache
	}
	return vm
}
```

- [ ] **Step 4: Check JIT in callFunction**

In `vm.go`, in the `OP_CALL` handler, before pushing a new frame:

```go
case OP_CALL:
	fnVal := regs[instr.Src1]
	argc := int(instr.Src2)
	// ... existing arg loading code ...

	if vf, ok := fnVal.(*vmFunction); ok {
		// Check JIT first
		if vm.jit != nil {
			fnKey := uintptr(unsafe.Pointer(vf.fn))
			if native, ok := vm.jit.GetNative(fnKey); ok {
				// Convert args to interface{}
				iArgs := make([]interface{}, len(args))
				for i, a := range args {
					iArgs[i] = a
				}
				result := native(iArgs)
				regs[instr.Dst] = result
				// skip VM call
				goto nextInstr
			}
			vm.jit.RecordCall(fnKey)
		}
		// ... existing VM frame push ...
	}
nextInstr:
```

- [ ] **Step 5: Write integration test**

```go
func TestJITIntegration(t *testing.T) {
	cfg := TestConfig()
	cfg.VMEnabled = true
	cfg.JITEnabled = true
	cfg.JITThreshold = 5 // trigger fast for tests

	src := `
fn add(a: int, b: int): int
    return a + b
end
// Call enough times to trigger JIT
i = 0
result = 0
while i < 10
    result = add(i, i)
    i = i + 1
end
result
`
	prog, _ := ParseProgram([]byte(src))
	interp := newInterpreter(cfg)
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	vm := NewVM(interp, cfg)
	result := vm.Execute(fn)
	if result != 18 { // add(9, 9) = 18
		t.Errorf("expected 18, got %v", result)
	}
}
```

```bash
CGO_ENABLED=1 go test ./interpreter/... -run "TestJITIntegration" -v
```

Expected: PASS (or Skip if libLLVM unavailable).

- [ ] **Step 6: Run full test suite**

```bash
go test ./... -count=1 2>&1 | tail -10
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add interpreter/vm.go interpreter/config.go
git commit -m "feat(jit): wire JITController into VM callFunction dispatch"
```

---

## Task 8: Benchmark and Tune JIT Threshold

- [ ] **Step 1: Run benchmarks with JIT disabled (VM baseline)**

```bash
go test ./interpreter/... \
  -bench="BenchmarkAdvancedNestedFunctionCalls|BenchmarkAdvancedMapOperations" \
  -benchtime=5s -count=5 2>&1 | tee /tmp/bench_vm.txt
```

- [ ] **Step 2: Run benchmarks with JIT enabled**

Add a benchmark that exercises JIT:

```go
func BenchmarkJITTypedIntLoop(b *testing.B) {
	src := `
fn add(a: int, b: int): int
    return a + b
end
i = 0
sum = 0
while i < 1000
    sum = add(i, i)
    i = i + 1
end
sum
`
	cfg := TestConfig()
	cfg.VMEnabled = true
	cfg.JITEnabled = true
	cfg.JITThreshold = 10

	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)

	b.ResetTimer()
	for b.Loop() {
		interp := newInterpreter(cfg)
		vm := NewVM(interp, cfg)
		vm.Execute(fn)
	}
}
```

```bash
CGO_ENABLED=1 go test ./interpreter/... -bench=BenchmarkJITTypedIntLoop -benchtime=5s -count=5
```

- [ ] **Step 3: Compare and document**

Record: VM-only ns/op vs JIT ns/op. Expected ≥3x improvement for typed integer functions.

- [ ] **Step 4: Tune threshold**

If the improvement is <2x, the compilation overhead exceeds the benefit. Try threshold=100 and threshold=5000. Document the sweet spot in `config.go`:

```go
// JITThreshold is the number of invocations before a function is JIT-compiled.
// Too low: JIT compilation overhead exceeds benefit for short-lived scripts.
// Too high: hot functions never get JIT-compiled.
// Default 1000 is optimal for scripts with >10K total function calls.
JITThreshold int64
```

- [ ] **Step 5: Commit**

```bash
git add interpreter/config.go interpreter/performance_test.go
git commit -m "feat(jit): tune JIT threshold to 1000 based on benchmark results"
```

---

## Phase 2 Complete

Final benchmark across all phases:

```bash
# Phase 0 baseline (VMEnabled=false, JITEnabled=false)
# Phase 1 (VMEnabled=true, JITEnabled=false)
# Phase 2 (VMEnabled=true, JITEnabled=true)

CGO_ENABLED=1 go test ./interpreter/... \
  -bench="BenchmarkAdvancedNestedFunctionCalls|BenchmarkJITTypedIntLoop" \
  -benchtime=5s -count=5
```

Target results:
- `BenchmarkAdvancedNestedFunctionCalls` Phase 1 vs Phase 0: ≥5x improvement
- `BenchmarkJITTypedIntLoop` Phase 2 vs Phase 1: ≥3x improvement for typed int paths
- `BenchmarkJITTypedIntLoop` Phase 2 vs Phase 0 tree-walker: ≥15x total

Record results in `docs/superpowers/specs/2026-05-25-compiler-pipeline-design.md` under a new "Results" section.
