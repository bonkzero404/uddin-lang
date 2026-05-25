# Phase 1 — Register-Based Bytecode VM Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the tree-walking interpreter with a register-based bytecode VM that achieves ~8-10x speedup, while keeping the existing interpreter as an opt-out fallback.

**Architecture:** AST (unchanged) → new `Compiler` (AST → `[]Instruction`) → new `VM` (tight switch loop over `[]Instruction`). `Config.VMEnabled = true` by default once all tests pass. Existing `interpreter.go` tree-walker remains as fallback (`VMEnabled = false`).

**Tech Stack:** Go 1.25.4, standard library only. No new external dependencies.

**Prerequisite:** Phase 0 bug fixes must be complete and all tests passing.

---

## File Map

| File | Status | Purpose |
|------|--------|---------|
| `interpreter/opcode.go` | Create | `OpCode` enum, `Instruction` struct, `TypeHint` type, disassembler |
| `interpreter/bytecode_function.go` | Create | `Function` (bytecode), `UpvalueInfo`, `Frame` structs |
| `interpreter/compiler.go` | Create | `Compiler` struct, `Compile(*Program) *Function` |
| `interpreter/vm.go` | Create | `VM` struct, `run()` execution loop, all opcode handlers |
| `interpreter/vm_builtins.go` | Create | `OP_CALL_BUILTIN` index table, `vmBuiltinEntry` |
| `interpreter/config.go` | Modify | Add `VMEnabled bool` field to `Config` |
| `interpreter/program.go` | Modify | Wire `VMEnabled` into `Execute()` / `RunProgramWithOptions()` |
| `interpreter/parser.go` | Modify | (Phase 1.5) Add optional `: type` syntax for parameter hints |

---

## Task 1: Define opcodes, Instruction, and TypeHint

**Files:**
- Create: `interpreter/opcode.go`
- Create: `interpreter/bytecode_function.go`

- [ ] **Step 1: Write tests for Instruction encoding**

Create `interpreter/opcode_test.go`:

```go
package interpreter

import "testing"

func TestInstructionEncoding(t *testing.T) {
	instr := Instruction{Op: OP_LOAD_CONST, Dst: 1, Src1: 0, Src2: 5}
	if instr.Op != OP_LOAD_CONST {
		t.Errorf("Op mismatch: got %d", instr.Op)
	}
	if instr.Dst != 1 || instr.Src1 != 0 || instr.Src2 != 5 {
		t.Errorf("register fields wrong: Dst=%d Src1=%d Src2=%d", instr.Dst, instr.Src1, instr.Src2)
	}
}

func TestOpcodeStringNotEmpty(t *testing.T) {
	for op := OpCode(0); op < _OP_MAX; op++ {
		if opName(op) == "" {
			t.Errorf("opcode %d has no name", op)
		}
	}
}

func TestFunctionHasCode(t *testing.T) {
	fn := &Function{
		Name:    "test",
		MaxRegs: 4,
		Code:    []Instruction{{Op: OP_RETURN, Src1: 0}},
		Constants: []Value{nil},
	}
	if len(fn.Code) != 1 {
		t.Error("expected 1 instruction")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./interpreter/... -run "TestInstruction|TestOpcode|TestFunction" -v
```

Expected: FAIL — types don't exist yet.

- [ ] **Step 3: Create opcode.go**

```go
package interpreter

// OpCode identifies a VM instruction.
type OpCode uint8

const (
	// Load/Store
	OP_LOAD_CONST  OpCode = iota // Dst = fn.Constants[Src1<<8|Src2]
	OP_LOAD_VAR                  // Dst = locals[Src1]
	OP_STORE_VAR                 // locals[Dst] = regs[Src1]
	OP_LOAD_UPVAL                // Dst = closure.Upvalues[Src1]
	OP_STORE_UPVAL               // closure.Upvalues[Dst] = regs[Src1]

	// Generic arithmetic (Value operands)
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_POW
	OP_NEG // Dst = -regs[Src1]

	// Typed arithmetic (int operands, no boxing — emitted only with type hints)
	OP_ADD_INT
	OP_SUB_INT
	OP_MUL_INT
	OP_DIV_INT
	OP_MOD_INT

	// Comparison
	OP_EQ
	OP_NEQ
	OP_LT
	OP_LTE
	OP_GT
	OP_GTE
	OP_IN

	// Logic
	OP_AND
	OP_OR
	OP_NOT
	OP_XOR

	// Control flow
	OP_JUMP       // pc += int16(Src1<<8|Src2)
	OP_JUMP_FALSE // if !IsTruthy(regs[Src1]): pc += int16(Src1<<8|Src2) — NOTE: offset in Dst:Src2
	OP_JUMP_TRUE  // if IsTruthy(regs[Src1]): pc += int16(Dst<<8|Src2)
	OP_RETURN     // return regs[Src1]

	// Collections
	OP_MAKE_ARRAY // Dst = []Value{regs[Src1]..regs[Src1+Src2-1]}
	OP_MAKE_MAP   // Dst = map[string]Value built from Src2 pairs starting at Src1
	OP_SUBSCRIPT  // Dst = regs[Src1][regs[Src2]]
	OP_SET_INDEX  // regs[Dst][regs[Src1]] = regs[Src2]

	// Functions
	OP_MAKE_FUNC    // Dst = &userFunction from fn.SubFunctions[Src1<<8|Src2]
	OP_CALL         // Dst = call(regs[Src1], argc=Src2, args at Dst+1..Dst+Src2)
	OP_CALL_BUILTIN // Dst = builtinTable[Src1<<8|Src2](argc=next.Dst)

	// Sentinel
	_OP_MAX
)

// Instruction is a single VM instruction. 4 bytes. Cache-line efficient.
type Instruction struct {
	Op   OpCode
	Dst  uint8
	Src1 uint8
	Src2 uint8
}

// jumpOffset decodes the signed 16-bit jump offset from Dst:Src2 fields.
func (i Instruction) jumpOffset() int {
	return int(int16(int(i.Dst)<<8 | int(i.Src2)))
}

// constIndex decodes the 16-bit constant pool index from Src1:Src2 fields.
func (i Instruction) constIndex() int {
	return int(i.Src1)<<8 | int(i.Src2)
}

// opName returns a human-readable name for an opcode (for disassembly/debug).
func opName(op OpCode) string {
	names := [_OP_MAX]string{
		"LOAD_CONST", "LOAD_VAR", "STORE_VAR", "LOAD_UPVAL", "STORE_UPVAL",
		"ADD", "SUB", "MUL", "DIV", "MOD", "POW", "NEG",
		"ADD_INT", "SUB_INT", "MUL_INT", "DIV_INT", "MOD_INT",
		"EQ", "NEQ", "LT", "LTE", "GT", "GTE", "IN",
		"AND", "OR", "NOT", "XOR",
		"JUMP", "JUMP_FALSE", "JUMP_TRUE", "RETURN",
		"MAKE_ARRAY", "MAKE_MAP", "SUBSCRIPT", "SET_INDEX",
		"MAKE_FUNC", "CALL", "CALL_BUILTIN",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return ""
}

// Disassemble returns a human-readable listing of a function's bytecode.
func Disassemble(fn *Function) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "function %q  regs=%d  consts=%d\n", fn.Name, fn.MaxRegs, len(fn.Constants))
	for i, instr := range fn.Code {
		fmt.Fprintf(&sb, "  %04d  %-14s  dst=%-3d src1=%-3d src2=%-3d\n",
			i, opName(instr.Op), instr.Dst, instr.Src1, instr.Src2)
	}
	return sb.String()
}
```

Add imports `"fmt"` and `"strings"` at the top.

- [ ] **Step 4: Create bytecode_function.go**

```go
package interpreter

// TypeHint represents an optional type annotation on a function parameter.
type TypeHint uint8

const (
	TypeUnknown TypeHint = iota
	TypeInt
	TypeFloat
	TypeString
	TypeBool
	TypeArray
	TypeMap
	TypeVoid
)

// UpvalueInfo describes a captured variable (closure upvalue).
type UpvalueInfo struct {
	Name    string
	IsLocal bool // true = captured from immediately enclosing function's locals
	Index   uint8
}

// Function is a compiled bytecode function (user-defined or top-level program).
type Function struct {
	Name        string
	Params      []string
	ParamTypes  []TypeHint   // len == 0 means all untyped
	ReturnType  TypeHint
	Code        []Instruction
	Constants   []Value
	Upvalues    []UpvalueInfo
	SubFunctions []*Function // inner function definitions
	MaxRegs     uint8
}

// Frame is one activation record in the VM call stack.
type Frame struct {
	fn      *Function
	pc      int
	baseReg int // offset into vm.regs where this frame's register window starts
	retReg  int // register in the CALLER's window where the return value is stored
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./interpreter/... -run "TestInstruction|TestOpcode|TestFunction" -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add interpreter/opcode.go interpreter/opcode_test.go interpreter/bytecode_function.go
git commit -m "feat(vm): add OpCode enum, Instruction struct, Function bytecode type"
```

---

## Task 2: Compiler — Literals, Variables, Simple Expressions

**Files:**
- Create: `interpreter/compiler.go`
- Create: `interpreter/compiler_test.go`

- [ ] **Step 1: Write failing tests**

Create `interpreter/compiler_test.go`:

```go
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
	// Last instruction must be RETURN
	last := fn.Code[len(fn.Code)-1]
	if last.Op != OP_RETURN {
		t.Errorf("expected last instruction RETURN, got %s", opName(last.Op))
	}
	// Constant pool must contain 42
	found := false
	for _, c := range fn.Constants {
		if c == 42 {
			found = true
		}
	}
	if !found {
		t.Errorf("constant 42 not in constant pool: %v", fn.Constants)
	}
}

func TestCompileVarAssignAndLoad(t *testing.T) {
	src := `
x = 10
x
`
	prog, _ := ParseProgram([]byte(src))
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	// Should contain LOAD_CONST, STORE_VAR, LOAD_VAR, RETURN
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
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./interpreter/... -run "TestCompile" -v
```

Expected: FAIL — `NewCompiler` not defined.

- [ ] **Step 3: Create compiler.go**

```go
package interpreter

import "fmt"

// Compiler compiles a parsed AST into bytecode Functions.
type Compiler struct {
	fn       *Function          // function currently being compiled
	locals   map[string]uint8   // variable name → register index
	nextReg  uint8              // next available register
	scope    []*compilerScope   // scope stack for nested functions
}

type compilerScope struct {
	locals  map[string]uint8
	nextReg uint8
}

// NewCompiler creates a compiler instance.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile compiles a whole program into a top-level Function.
func (c *Compiler) Compile(prog *Program) (*Function, error) {
	c.fn = &Function{Name: "<main>"}
	c.locals = make(map[string]uint8)
	c.nextReg = 0

	for _, stmt := range prog.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}
	// Implicit return nil at end of program
	c.emit(Instruction{Op: OP_LOAD_CONST, Dst: c.nextReg, Src1: 0, Src2: 0})
	c.fn.Constants = append([]Value{nil}, c.fn.Constants...)
	// Fix all const indices that were emitted before the nil was prepended — see addConst
	c.emit(Instruction{Op: OP_RETURN, Src1: c.nextReg})
	c.fn.MaxRegs = c.nextReg + 1
	return c.fn, nil
}

func (c *Compiler) emit(instr Instruction) int {
	c.fn.Code = append(c.fn.Code, instr)
	return len(c.fn.Code) - 1
}

// addConst adds a value to the constant pool and returns its index.
func (c *Compiler) addConst(v Value) uint16 {
	for i, existing := range c.fn.Constants {
		if existing == v {
			return uint16(i)
		}
	}
	idx := len(c.fn.Constants)
	c.fn.Constants = append(c.fn.Constants, v)
	return uint16(idx)
}

// allocReg allocates the next available register.
func (c *Compiler) allocReg() uint8 {
	r := c.nextReg
	c.nextReg++
	if c.nextReg > c.fn.MaxRegs {
		c.fn.MaxRegs = c.nextReg
	}
	return r
}

// localReg returns the register for a local variable, allocating if new.
func (c *Compiler) localReg(name string) uint8 {
	if r, ok := c.locals[name]; ok {
		return r
	}
	r := c.allocReg()
	c.locals[name] = r
	return r
}

func (c *Compiler) compileStatement(stmt Statement) error {
	switch s := stmt.(type) {
	case *Assign:
		return c.compileAssign(s)
	case *ExprStatement:
		_, err := c.compileExpr(s.Expr)
		return err
	case *Return:
		return c.compileReturn(s)
	default:
		return fmt.Errorf("compiler: unsupported statement type %T", stmt)
	}
}

func (c *Compiler) compileAssign(s *Assign) error {
	dst := c.localReg(s.Target.(*Variable).Name)
	src, err := c.compileExpr(s.Value)
	if err != nil {
		return err
	}
	if s.Operator == ASSIGN {
		c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: src})
	}
	// Compound assignment operators (+=, -=, etc.) — load, operate, store
	// Deferred to Task 4 (binary expressions)
	return nil
}

func (c *Compiler) compileReturn(s *Return) error {
	if s.Value == nil {
		idx := c.addConst(nil)
		r := c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(idx >> 8), Src2: uint8(idx)})
		c.emit(Instruction{Op: OP_RETURN, Src1: r})
		return nil
	}
	src, err := c.compileExpr(s.Value)
	if err != nil {
		return err
	}
	c.emit(Instruction{Op: OP_RETURN, Src1: src})
	return nil
}

// compileExpr compiles an expression and returns the register holding the result.
func (c *Compiler) compileExpr(expr Expression) (uint8, error) {
	switch e := expr.(type) {
	case *Literal:
		idx := c.addConst(e.Value)
		dst := c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: dst, Src1: uint8(idx >> 8), Src2: uint8(idx)})
		return dst, nil

	case *Variable:
		if r, ok := c.locals[e.Name]; ok {
			dst := c.allocReg()
			c.emit(Instruction{Op: OP_LOAD_VAR, Dst: dst, Src1: r})
			return dst, nil
		}
		return 0, fmt.Errorf("compiler: undefined variable %q at %d:%d", e.Name, e.Position().Line, e.Position().Column)

	default:
		return 0, fmt.Errorf("compiler: unsupported expression type %T", expr)
	}
}
```

**Note:** `ExprStatement` and `Return` AST types — check the actual field names in `ast.go` before implementing (they may differ). Adjust field accesses accordingly.

- [ ] **Step 4: Run compile tests**

```bash
go test ./interpreter/... -run "TestCompile" -v
```

Expected: `TestCompileLiteral` and `TestCompileVarAssignAndLoad` pass (partial — other statement types return "unsupported" errors for now).

- [ ] **Step 5: Commit**

```bash
git add interpreter/compiler.go interpreter/compiler_test.go
git commit -m "feat(vm): add Compiler skeleton with literal and variable support"
```

---

## Task 3: VM — Basic Execution (LOAD_CONST, LOAD_VAR, STORE_VAR, RETURN)

**Files:**
- Create: `interpreter/vm.go`
- Create: `interpreter/vm_test.go`

- [ ] **Step 1: Write tests**

Create `interpreter/vm_test.go`:

```go
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
	prog, _ := ParseProgram([]byte(`x = 7\nx`))
	c := NewCompiler()
	fn, _ := c.Compile(prog)
	vm := makeVM(TestConfig())
	result := vm.Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./interpreter/... -run "TestVM" -v
```

Expected: FAIL — `VM` type not defined.

- [ ] **Step 3: Create vm.go**

```go
package interpreter

// VM executes compiled bytecode Functions.
type VM struct {
	regs   []Value  // flat register array; frames use windows into this
	frames []Frame
	interp *interpreter // for builtins that need interpreter context
}

const vmInitialRegCount = 256
const vmMaxFrames = 512

// NewVM creates a VM backed by the given interpreter (for builtin access).
func NewVM(interp *interpreter) *VM {
	return &VM{
		regs:   make([]Value, vmInitialRegCount),
		frames: make([]Frame, 0, 32),
		interp: interp,
	}
}

// Execute runs a top-level Function and returns its result.
func (vm *VM) Execute(fn *Function) Value {
	vm.pushFrame(fn, 0, 0)
	return vm.run()
}

func (vm *VM) pushFrame(fn *Function, baseReg int, retReg int) {
	// Ensure register array is large enough
	needed := baseReg + int(fn.MaxRegs) + 1
	for len(vm.regs) < needed {
		vm.regs = append(vm.regs, make([]Value, needed-len(vm.regs))...)
	}
	vm.frames = append(vm.frames, Frame{
		fn:      fn,
		pc:      0,
		baseReg: baseReg,
		retReg:  retReg,
	})
}

func (vm *VM) run() Value {
top:
	frame := &vm.frames[len(vm.frames)-1]
	regs := vm.regs[frame.baseReg:]
	code := frame.fn.Code
	consts := frame.fn.Constants

	for {
		if frame.pc >= len(code) {
			// Implicit return nil
			return vm.doReturn(nil)
		}
		instr := code[frame.pc]
		frame.pc++

		switch instr.Op {
		case OP_LOAD_CONST:
			regs[instr.Dst] = consts[instr.constIndex()]

		case OP_LOAD_VAR:
			regs[instr.Dst] = regs[instr.Src1]

		case OP_STORE_VAR:
			regs[instr.Dst] = regs[instr.Src1]

		case OP_RETURN:
			val := regs[instr.Src1]
			result := vm.doReturn(val)
			if len(vm.frames) == 0 {
				return result
			}
			// Re-enter outer frame
			frame = &vm.frames[len(vm.frames)-1]
			regs = vm.regs[frame.baseReg:]
			code = frame.fn.Code
			consts = frame.fn.Constants
			goto top

		default:
			// Unimplemented opcode — delegate to tree-walker for now
			// This lets us run tests even before all opcodes are implemented.
			panic(runtimeError(Position{}, "VM: unimplemented opcode %s", opName(instr.Op)))
		}
	}
}

func (vm *VM) doReturn(val Value) Value {
	// Pop frame
	top := vm.frames[len(vm.frames)-1]
	vm.frames = vm.frames[:len(vm.frames)-1]
	if len(vm.frames) == 0 {
		return val
	}
	// Store return value in caller's return register
	callerBase := vm.frames[len(vm.frames)-1].baseReg
	vm.regs[callerBase+top.retReg] = val
	return val
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./interpreter/... -run "TestVM" -v
```

Expected: `TestVMReturnLiteral` and `TestVMVarAssign` PASS.

- [ ] **Step 5: Commit**

```bash
git add interpreter/vm.go interpreter/vm_test.go
git commit -m "feat(vm): add VM with LOAD_CONST, LOAD_VAR, STORE_VAR, RETURN"
```

---

## Task 4: Compiler + VM — Binary Expressions and Arithmetic

**Files:**
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/vm.go`

- [ ] **Step 1: Write tests**

Add to `interpreter/vm_test.go`:

```go
func TestVMArithmetic(t *testing.T) {
	cases := []struct {
		src      string
		expected Value
	}{
		{`3 + 4`, 7},
		{`10 - 3`, 7},
		{`3 * 4`, 12},
		{`10 / 4`, 2.5},
		{`10 % 3`, 1},
		{`2 ** 8`, 256.0},
		{`-5`, -5},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			prog, err := ParseProgram([]byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			fn, err := NewCompiler().Compile(prog)
			if err != nil {
				t.Fatal(err)
			}
			result := makeVM(TestConfig()).Execute(fn)
			if result != tc.expected {
				t.Errorf("src=%q: expected %v (%T), got %v (%T)", tc.src, tc.expected, tc.expected, result, result)
			}
		})
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./interpreter/... -run "TestVMArithmetic" -v
```

Expected: FAIL — `compileExpr` returns error for `*Binary`.

- [ ] **Step 3: Add binary expression compilation to compiler.go**

Add to `compileExpr` in `compiler.go`:

```go
case *Binary:
	return c.compileBinary(e)
case *Unary:
	return c.compileUnary(e)
```

Add new methods:

```go
var tokenToOpCode = map[Token]OpCode{
	PLUS:     OP_ADD,
	MINUS:    OP_SUB,
	TIMES:    OP_MUL,
	DIVIDE:   OP_DIV,
	MODULO:   OP_MOD,
	POWER:    OP_POW,
	EQUAL:    OP_EQ,
	NOTEQUAL: OP_NEQ,
	LT:       OP_LT,
	LTE:      OP_LTE,
	GT:       OP_GT,
	GTE:      OP_GTE,
	IN:       OP_IN,
	AND:      OP_AND,
	OR:       OP_OR,
	XOR:      OP_XOR,
}

func (c *Compiler) compileBinary(e *Binary) (uint8, error) {
	left, err := c.compileExpr(e.Left)
	if err != nil {
		return 0, err
	}
	right, err := c.compileExpr(e.Right)
	if err != nil {
		return 0, err
	}
	op, ok := tokenToOpCode[e.Operator]
	if !ok {
		return 0, fmt.Errorf("compiler: unsupported binary operator %v", e.Operator)
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: op, Dst: dst, Src1: left, Src2: right})
	return dst, nil
}

func (c *Compiler) compileUnary(e *Unary) (uint8, error) {
	src, err := c.compileExpr(e.Operand)
	if err != nil {
		return 0, err
	}
	var op OpCode
	switch e.Operator {
	case MINUS:
		op = OP_NEG
	case NOT:
		op = OP_NOT
	default:
		return 0, fmt.Errorf("compiler: unsupported unary operator %v", e.Operator)
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: op, Dst: dst, Src1: src})
	return dst, nil
}
```

- [ ] **Step 4: Add arithmetic opcode handlers to vm.go**

In the `switch instr.Op` block in `run()`, add:

```go
case OP_ADD:
	regs[instr.Dst] = genericAdd(regs[instr.Src1], regs[instr.Src2])
case OP_SUB:
	regs[instr.Dst] = genericSub(regs[instr.Src1], regs[instr.Src2])
case OP_MUL:
	regs[instr.Dst] = genericMul(regs[instr.Src1], regs[instr.Src2])
case OP_DIV:
	regs[instr.Dst] = genericDiv(regs[instr.Src1], regs[instr.Src2])
case OP_MOD:
	regs[instr.Dst] = genericMod(regs[instr.Src1], regs[instr.Src2])
case OP_POW:
	regs[instr.Dst] = genericPow(regs[instr.Src1], regs[instr.Src2])
case OP_NEG:
	regs[instr.Dst] = genericNeg(regs[instr.Src1])

// Typed int-only arithmetic (no boxing — requires type-hint compiler path)
case OP_ADD_INT:
	regs[instr.Dst] = regs[instr.Src1].(int) + regs[instr.Src2].(int)
case OP_SUB_INT:
	regs[instr.Dst] = regs[instr.Src1].(int) - regs[instr.Src2].(int)
case OP_MUL_INT:
	regs[instr.Dst] = regs[instr.Src1].(int) * regs[instr.Src2].(int)
case OP_DIV_INT:
	r := regs[instr.Src2].(int)
	if r == 0 {
		panic("VM: integer divide by zero")
	}
	regs[instr.Dst] = regs[instr.Src1].(int) / r
case OP_MOD_INT:
	r := regs[instr.Src2].(int)
	if r == 0 {
		panic("VM: integer modulo by zero")
	}
	regs[instr.Dst] = regs[instr.Src1].(int) % r
```

Add generic arithmetic helpers at bottom of `vm.go` (these reuse the existing eval functions from `interpreter.go` to avoid duplicating logic):

```go
func genericAdd(l, r Value) Value { return newInterpreterForVM().evalPlus(Position{}, l, r) }
func genericSub(l, r Value) Value { return evalMinus(Position{}, l, r) }
func genericMul(l, r Value) Value { return evalTimes(Position{}, l, r) }
func genericDiv(l, r Value) Value { return evalDivide(Position{}, l, r) }
func genericMod(l, r Value) Value { return evalModulo(Position{}, l, r) }
func genericPow(l, r Value) Value { return evalPower(Position{}, l, r) }
func genericNeg(v Value) Value    { return evalNegative(Position{}, v) }
```

Add a helper:

```go
// newInterpreterForVM creates a minimal interpreter for reusing existing eval helpers.
func newInterpreterForVM() *interpreter {
	return &interpreter{}
}
```

- [ ] **Step 5: Add comparison and logic opcode handlers to vm.go**

```go
case OP_EQ:
	regs[instr.Dst] = evalEqual(Position{}, regs[instr.Src1], regs[instr.Src2])
case OP_NEQ:
	regs[instr.Dst] = !evalEqual(Position{}, regs[instr.Src1], regs[instr.Src2]).(bool)
case OP_LT:
	regs[instr.Dst] = evalLess(Position{}, regs[instr.Src1], regs[instr.Src2])
case OP_LTE:
	regs[instr.Dst] = !evalLess(Position{}, regs[instr.Src2], regs[instr.Src1]).(bool)
case OP_GT:
	regs[instr.Dst] = evalLess(Position{}, regs[instr.Src2], regs[instr.Src1])
case OP_GTE:
	regs[instr.Dst] = !evalLess(Position{}, regs[instr.Src1], regs[instr.Src2]).(bool)
case OP_IN:
	regs[instr.Dst] = evalIn(Position{}, regs[instr.Src1], regs[instr.Src2])
case OP_AND:
	regs[instr.Dst] = IsTruthy(regs[instr.Src1]) && IsTruthy(regs[instr.Src2])
case OP_OR:
	regs[instr.Dst] = IsTruthy(regs[instr.Src1]) || IsTruthy(regs[instr.Src2])
case OP_NOT:
	b, ok := regs[instr.Src1].(bool)
	if !ok {
		panic("VM: NOT requires bool")
	}
	regs[instr.Dst] = !b
case OP_XOR:
	regs[instr.Dst] = IsTruthy(regs[instr.Src1]) != IsTruthy(regs[instr.Src2])
```

- [ ] **Step 6: Run tests**

```bash
go test ./interpreter/... -run "TestVMArithmetic" -v
```

Expected: all arithmetic sub-cases PASS.

- [ ] **Step 7: Commit**

```bash
git add interpreter/compiler.go interpreter/vm.go
git commit -m "feat(vm): compile and execute binary/unary arithmetic and comparison"
```

---

## Task 5: Compiler + VM — Control Flow (if/elif/else, while, for-in)

**Files:**
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/vm.go`

- [ ] **Step 1: Write tests**

Add to `interpreter/vm_test.go`:

```go
func TestVMIfElse(t *testing.T) {
	src := `
x = 0
if 1 == 1 then
    x = 42
else
    x = 99
end
x
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestVMWhileLoop(t *testing.T) {
	src := `
i = 0
sum = 0
while i < 10
    sum = sum + i
    i = i + 1
end
sum
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 45 {
		t.Errorf("expected 45, got %v", result)
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./interpreter/... -run "TestVMIfElse|TestVMWhileLoop" -v
```

Expected: FAIL — if/while statement types not compiled.

- [ ] **Step 3: Add jump opcodes to vm.go**

In `switch instr.Op`:

```go
case OP_JUMP:
	frame.pc += instr.jumpOffset()
	// Recompute frame-relative vars since pc changed
	continue

case OP_JUMP_FALSE:
	if !IsTruthy(regs[instr.Src1]) {
		frame.pc += instr.jumpOffset()
	}

case OP_JUMP_TRUE:
	if IsTruthy(regs[instr.Src1]) {
		frame.pc += instr.jumpOffset()
	}
```

- [ ] **Step 4: Add if/while compilation to compiler.go**

Add jump-patching helpers:

```go
// emitJump emits a placeholder jump and returns its index for later patching.
func (c *Compiler) emitJump(op OpCode, condReg uint8) int {
	idx := c.emit(Instruction{Op: op, Src1: condReg, Dst: 0, Src2: 0})
	return idx
}

// patchJump fills in the offset for a previously emitted jump.
func (c *Compiler) patchJump(jumpIdx int) {
	offset := len(c.fn.Code) - jumpIdx - 1
	if offset > 32767 || offset < -32768 {
		panic("jump offset out of range")
	}
	o := int16(offset)
	c.fn.Code[jumpIdx].Dst = uint8(o >> 8)
	c.fn.Code[jumpIdx].Src2 = uint8(o)
}
```

Add `compileIf` to `compileStatement`:

```go
case *If:
	return c.compileIf(s)
case *While:
	return c.compileWhile(s)
case *For:
	return c.compileFor(s)
```

Add methods:

```go
func (c *Compiler) compileIf(s *If) error {
	// Compile condition
	cond, err := c.compileExpr(s.Condition)
	if err != nil {
		return err
	}
	// Jump past then-block if false
	jumpFalse := c.emitJump(OP_JUMP_FALSE, cond)

	// Then block
	for _, stmt := range s.Then {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	if len(s.ElseIfs) == 0 && len(s.Else) == 0 {
		c.patchJump(jumpFalse)
		return nil
	}

	// Jump past else block after then block
	jumpEnd := c.emit(Instruction{Op: OP_JUMP})
	c.patchJump(jumpFalse)

	// Else block
	for _, stmt := range s.Else {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.patchJump(jumpEnd)
	return nil
}

func (c *Compiler) compileWhile(s *While) error {
	loopStart := len(c.fn.Code)
	cond, err := c.compileExpr(s.Condition)
	if err != nil {
		return err
	}
	exitJump := c.emitJump(OP_JUMP_FALSE, cond)

	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Jump back to loop start
	offset := int16(loopStart - len(c.fn.Code) - 1)
	o := offset
	c.emit(Instruction{Op: OP_JUMP, Dst: uint8(o >> 8), Src2: uint8(o)})
	c.patchJump(exitJump)
	return nil
}
```

**Note:** Check the exact field names for `*If`, `*While`, `*For` AST types in `ast.go` (may be `Condition`, `Then`, `Else`, `ElseIfs`, `Body`, `Iterator`, `Variable`). Adjust accordingly.

- [ ] **Step 5: Run tests**

```bash
go test ./interpreter/... -run "TestVMIfElse|TestVMWhileLoop" -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add interpreter/compiler.go interpreter/vm.go
git commit -m "feat(vm): compile and execute if/elif/else and while control flow"
```

---

## Task 6: Compiler + VM — Collections (Array, Map, Subscript)

**Files:**
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/vm.go`

- [ ] **Step 1: Write tests**

```go
func TestVMArrayLiteral(t *testing.T) {
	src := `a = [1, 2, 3]\na[1]`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

func TestVMMapLiteral(t *testing.T) {
	src := `m = {"name": "alice", "age": 30}\nm["age"]`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 30 {
		t.Errorf("expected 30, got %v", result)
	}
}
```

- [ ] **Step 2: Add List/Map compilation to compileExpr in compiler.go**

```go
case *List:
	regs := make([]uint8, len(e.Values))
	startReg := c.nextReg
	for i, elem := range e.Values {
		r, err := c.compileExpr(elem)
		if err != nil {
			return 0, err
		}
		regs[i] = r
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_MAKE_ARRAY, Dst: dst, Src1: startReg, Src2: uint8(len(e.Values))})
	return dst, nil

case *Map:
	startReg := c.nextReg
	for _, item := range e.Items {
		_, err := c.compileExpr(item.Key)
		if err != nil {
			return 0, err
		}
		_, err = c.compileExpr(item.Value)
		if err != nil {
			return 0, err
		}
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_MAKE_MAP, Dst: dst, Src1: startReg, Src2: uint8(len(e.Items))})
	return dst, nil

case *Subscript:
	container, err := c.compileExpr(e.Container)
	if err != nil {
		return 0, err
	}
	subscript, err := c.compileExpr(e.Subscript)
	if err != nil {
		return 0, err
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_SUBSCRIPT, Dst: dst, Src1: container, Src2: subscript})
	return dst, nil
```

- [ ] **Step 3: Add collection opcode handlers to vm.go**

```go
case OP_MAKE_ARRAY:
	n := int(instr.Src2)
	arr := make([]Value, n)
	for i := 0; i < n; i++ {
		arr[i] = regs[int(instr.Src1)+i]
	}
	regs[instr.Dst] = &arr

case OP_MAKE_MAP:
	n := int(instr.Src2) // number of key-value pairs
	m := make(map[string]Value, n)
	base := int(instr.Src1)
	for i := 0; i < n; i++ {
		key := regs[base+i*2]
		val := regs[base+i*2+1]
		k, ok := key.(string)
		if !ok {
			panic("VM: map key must be string")
		}
		m[k] = val
	}
	regs[instr.Dst] = m

case OP_SUBSCRIPT:
	regs[instr.Dst] = evalSubscript(Position{}, regs[instr.Src1], regs[instr.Src2])

case OP_SET_INDEX:
	container := regs[instr.Dst]
	key := regs[instr.Src1]
	val := regs[instr.Src2]
	switch c := container.(type) {
	case *[]Value:
		idx := key.(int)
		if idx < 0 {
			idx = len(*c) + idx
		}
		(*c)[idx] = val
	case map[string]Value:
		c[key.(string)] = val
	default:
		panic("VM: SET_INDEX on non-container")
	}
```

- [ ] **Step 4: Run tests**

```bash
go test ./interpreter/... -run "TestVMArray|TestVMMap" -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add interpreter/compiler.go interpreter/vm.go
git commit -m "feat(vm): compile and execute array/map literals and subscript"
```

---

## Task 7: Compiler + VM — User Functions and Calls

**Files:**
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/vm.go`
- Create: `interpreter/vm_builtins.go`

- [ ] **Step 1: Write tests**

```go
func TestVMUserFunction(t *testing.T) {
	src := `
fn add(a, b)
    return a + b
end
add(3, 4)
`
	prog, _ := ParseProgram([]byte(src))
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestVMBuiltinLen(t *testing.T) {
	src := `len([1, 2, 3])`
	prog, _ := ParseProgram([]byte(src))
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}
```

- [ ] **Step 2: Create vm_builtins.go**

```go
package interpreter

// vmBuiltinEntry maps a builtin name to its index in the VM builtin table.
type vmBuiltinEntry struct {
	name string
	fn   func(*interpreter, Position, []Value) Value
}

// vmBuiltinTable is the ordered list of builtins available via OP_CALL_BUILTIN.
// Index must be stable — append only. Never reorder.
var vmBuiltinTable []vmBuiltinEntry

// vmBuiltinIndex maps builtin name → index in vmBuiltinTable.
var vmBuiltinIndex = map[string]uint16{}

func init() {
	// Register builtins in the VM table.
	// These must match the functions registered in globalBuiltinDispatcher.
	registerVMBuiltins()
}

func registerVMBuiltin(name string, fn func(*interpreter, Position, []Value) Value) {
	idx := uint16(len(vmBuiltinTable))
	vmBuiltinTable = append(vmBuiltinTable, vmBuiltinEntry{name: name, fn: fn})
	vmBuiltinIndex[name] = idx
}

// registerVMBuiltins populates vmBuiltinTable.
// Add each builtin that should be callable via OP_CALL_BUILTIN.
func registerVMBuiltins() {
	// Core builtins — must match fn_core.go registrations.
	// TODO: populate by iterating globalBuiltinDispatcher.functions after init.
	// For now, register the most-used builtins explicitly:
	registerVMBuiltin("len", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("len", interp, pos, args)
	})
	registerVMBuiltin("print", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("print", interp, pos, args)
	})
	registerVMBuiltin("str", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("str", interp, pos, args)
	})
	registerVMBuiltin("int", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("int", interp, pos, args)
	})
	registerVMBuiltin("float", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("float", interp, pos, args)
	})
	registerVMBuiltin("type", func(interp *interpreter, pos Position, args []Value) Value {
		return globalBuiltinDispatcher.DispatchOrPanic("type", interp, pos, args)
	})
	// Add more builtins as needed. Order of registration determines index.
}
```

Add `DispatchOrPanic` to `builtin_dispatch.go`:

```go
func (bfd *BuiltinFunctionDispatcher) DispatchOrPanic(name string, interp *interpreter, pos Position, args []Value) Value {
	result, ok := bfd.DispatchBuiltinFunction(name, interp, pos, args)
	if !ok {
		panic(runtimeError(pos, "unknown builtin %q", name))
	}
	return result
}
```

- [ ] **Step 3: Add FunctionDef compilation to compiler.go**

```go
case *FunctionDef:
	return c.compileFunctionDef(s)
```

```go
func (c *Compiler) compileFunctionDef(s *FunctionDef) error {
	sub := &Compiler{
		fn:      &Function{Name: s.Name, Params: s.Parameters},
		locals:  make(map[string]uint8),
		nextReg: 0,
	}
	// Parameters occupy the first registers
	for i, param := range s.Parameters {
		sub.locals[param] = uint8(i)
		sub.nextReg++
	}
	for _, stmt := range s.Body {
		if err := sub.compileStatement(stmt); err != nil {
			return err
		}
	}
	// Implicit nil return
	nilIdx := sub.addConst(nil)
	r := sub.allocReg()
	sub.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(nilIdx >> 8), Src2: uint8(nilIdx)})
	sub.emit(Instruction{Op: OP_RETURN, Src1: r})
	sub.fn.MaxRegs = sub.nextReg

	// Add the compiled sub-function to parent and emit MAKE_FUNC
	subIdx := uint16(len(c.fn.SubFunctions))
	c.fn.SubFunctions = append(c.fn.SubFunctions, sub.fn)
	dst := c.localReg(s.Name)
	c.emit(Instruction{Op: OP_MAKE_FUNC, Dst: dst, Src1: uint8(subIdx >> 8), Src2: uint8(subIdx)})
	return nil
}
```

- [ ] **Step 4: Add Call compilation to compiler.go**

In `compileExpr`:

```go
case *Call:
	return c.compileCall(e)
```

```go
func (c *Compiler) compileCall(e *Call) (uint8, error) {
	// Check if it's a known builtin
	fnExpr, isVar := e.Function.(*Variable)
	if isVar {
		if builtinIdx, ok := vmBuiltinIndex[fnExpr.Name]; ok {
			// Compile args
			argStart := c.nextReg
			for _, arg := range e.Arguments {
				_, err := c.compileExpr(arg)
				if err != nil {
					return 0, err
				}
			}
			dst := c.allocReg()
			c.emit(Instruction{Op: OP_CALL_BUILTIN, Dst: dst,
				Src1: uint8(builtinIdx >> 8), Src2: uint8(builtinIdx)})
			// Encode argc in next instruction's Dst field
			c.emit(Instruction{Op: OP_LOAD_CONST, Dst: uint8(len(e.Arguments)),
				Src1: uint8(argStart >> 8), Src2: uint8(argStart)})
			return dst, nil
		}
	}

	// User function call
	fnReg, err := c.compileExpr(e.Function)
	if err != nil {
		return 0, err
	}
	argStart := c.nextReg
	for _, arg := range e.Arguments {
		_, err := c.compileExpr(arg)
		if err != nil {
			return 0, err
		}
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_CALL, Dst: dst, Src1: fnReg, Src2: uint8(len(e.Arguments))})
	// Encode argStart in a following LOAD_VAR pseudo-instruction
	c.emit(Instruction{Op: OP_LOAD_VAR, Dst: uint8(argStart)})
	return dst, nil
}
```

- [ ] **Step 5: Add MAKE_FUNC, CALL, OP_CALL_BUILTIN handlers to vm.go**

```go
case OP_MAKE_FUNC:
	subIdx := int(instr.Src1)<<8 | int(instr.Src2)
	subfn := frame.fn.SubFunctions[subIdx]
	// Wrap as a vmFunction (implements functionType)
	regs[instr.Dst] = &vmFunction{fn: subfn, vm: vm}

case OP_CALL:
	fnVal := regs[instr.Src1]
	argc := int(instr.Src2)
	// Arg start encoded in next instruction
	argInstr := code[frame.pc]
	frame.pc++
	argBase := int(argInstr.Dst)

	args := make([]Value, argc)
	for i := 0; i < argc; i++ {
		args[i] = regs[argBase+i]
	}

	switch f := fnVal.(type) {
	case *vmFunction:
		newBase := frame.baseReg + int(frame.fn.MaxRegs) + 1
		vm.pushFrame(f.fn, newBase, int(instr.Dst))
		// Copy args into new frame's param registers
		newRegs := vm.regs[newBase:]
		for i, arg := range args {
			newRegs[i] = arg
		}
		// Re-enter run loop via goto top
		frame = &vm.frames[len(vm.frames)-1]
		regs = vm.regs[frame.baseReg:]
		code = frame.fn.Code
		consts = frame.fn.Constants
		goto top

	case functionType:
		// Fallback: any other functionType (builtins wrapped as functionType)
		// call via the existing interface
		regs[instr.Dst] = fnVal.(functionType).call(vm.interp, Position{}, args)
	}

case OP_CALL_BUILTIN:
	builtinIdx := int(instr.Src1)<<8 | int(instr.Src2)
	// Argc + argStart encoded in next instruction
	meta := code[frame.pc]
	frame.pc++
	argc := int(meta.Dst)
	argBase := int(meta.Src1)<<8 | int(meta.Src2)
	args := make([]Value, argc)
	for i := 0; i < argc; i++ {
		args[i] = regs[argBase+i]
	}
	entry := vmBuiltinTable[builtinIdx]
	regs[instr.Dst] = entry.fn(vm.interp, Position{}, args)
```

Add `vmFunction` type:

```go
type vmFunction struct {
	fn  *Function
	vm  *VM
}

func (f *vmFunction) call(interp *interpreter, pos Position, args []Value) Value {
	newBase := 0 // calculated at call site
	f.vm.pushFrame(f.fn, newBase, 0)
	newRegs := f.vm.regs[newBase:]
	for i, arg := range args {
		newRegs[i] = arg
	}
	return f.vm.run()
}

func (f *vmFunction) name() string { return f.fn.Name }
```

- [ ] **Step 6: Run tests**

```bash
go test ./interpreter/... -run "TestVMUserFunction|TestVMBuiltinLen" -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add interpreter/compiler.go interpreter/vm.go interpreter/vm_builtins.go interpreter/builtin_dispatch.go
git commit -m "feat(vm): compile and execute user function definitions, calls, and builtins"
```

---

## Task 8: Wire VMEnabled into Config and Execute Path

**Files:**
- Modify: `interpreter/config.go`
- Modify: `interpreter/program.go`

- [ ] **Step 1: Add VMEnabled to Config**

In `interpreter/config.go`, add to the `Config` struct:

```go
// VMEnabled runs programs through the bytecode VM instead of the tree-walker.
// Default false during Phase 1 rollout. Set true once all tests pass.
VMEnabled bool
```

In `DefaultConfig()`, `TestConfig()`, `StableConfig()`, `ExperimentalConfig()`: leave `VMEnabled: false` for now.

- [ ] **Step 2: Wire into Execute()**

In `interpreter/program.go`, find the `Execute` function and add VM dispatch:

```go
func Execute(prog *Program, config *Config) (*Stats, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if config.VMEnabled {
		return executeVM(prog, config)
	}
	return executeTreeWalker(prog, config)
}

func executeVM(prog *Program, config *Config) (*Stats, error) {
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		return nil, err
	}
	interp := newInterpreter(config)
	vm := NewVM(interp)
	vm.Execute(fn)
	stats := &Stats{Ops: interp.stats.Ops}
	return stats, nil
}
```

Rename the existing `Execute` implementation body to `executeTreeWalker`.

- [ ] **Step 3: Run full test suite with VMEnabled=false (default)**

```bash
go test ./... -count=1
```

Expected: all pass — no behavioural change since VMEnabled=false.

- [ ] **Step 4: Run key tests with VMEnabled=true**

```go
// Add to vm_test.go:
func TestVMEnabledConfig(t *testing.T) {
	cfg := TestConfig()
	cfg.VMEnabled = true
	src := `
fn fib(n)
    if n <= 1 then
        return n
    end
    return fib(n - 1) + fib(n - 2)
end
fib(10)
`
	prog, _ := ParseProgram([]byte(src))
	_, err := Execute(prog, cfg)
	if err != nil {
		t.Fatal(err)
	}
}
```

```bash
go test ./interpreter/... -run "TestVMEnabledConfig" -v
```

Expected: PASS (or identifies missing statement types to implement next).

- [ ] **Step 5: Commit**

```bash
git add interpreter/config.go interpreter/program.go interpreter/vm_test.go
git commit -m "feat(vm): wire Config.VMEnabled into Execute() dispatch path"
```

---

## Task 9: Expand Compiler — Remaining Statement Types

**Files:**
- Modify: `interpreter/compiler.go`

Cover all statement types the tree-walker handles. For each type, follow the same TDD pattern: write a failing test, compile it, run the VM, verify result.

Types to cover in this task (check `ast.go` for exact type names):

- `*OuterAssign` — outer scope assignment
- `*Break` / `*Continue` — emit OP_JUMP with loop-end/loop-start targets
- `*ForIn` / `*ForOf` — iterate over array/map
- `*Ternary` expression — compile condition + two branches
- `*FunctionExpression` (anonymous function) — same as FunctionDef but no name
- `*Call` with ellipsis (spread args)
- Compound assignment operators (`+=`, `-=`, `*=`, `/=`, `%=`)

For each, follow this pattern:

- [ ] **Write test → Run to fail → Implement in compiler → Add VM handler if needed → Run to pass → Commit**

Minimum test per type:

```go
// Break/Continue
func TestVMBreak(t *testing.T) {
	src := `
i = 0
while true
    if i >= 5 then break end
    i = i + 1
end
i
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 5 {
		t.Errorf("expected 5, got %v", result)
	}
}

// ForIn (array iteration)
func TestVMForIn(t *testing.T) {
	src := `
sum = 0
for x in [1, 2, 3, 4, 5]
    sum = sum + x
end
sum
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

// Ternary
func TestVMTernary(t *testing.T) {
	src := `x = 1 == 1 ? 42 : 99\nx`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	result := makeVM(TestConfig()).Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}
```

Commit after each statement type:

```bash
git commit -m "feat(vm): compile <statement type>"
```

---

## Task 10: Full Test Suite Parity — VMEnabled=true

**Files:**
- Modify: `interpreter/config.go` (change `DefaultConfig().VMEnabled = true`)

- [ ] **Step 1: Enable VM in DefaultConfig**

```go
func DefaultConfig() *Config {
	return &Config{
		// ...
		VMEnabled: true,
	}
}
```

- [ ] **Step 2: Run full test suite**

```bash
go test ./... -count=1 -v 2>&1 | grep -E "FAIL|PASS|---" | head -60
```

Expected: all PASS. If any FAIL, note the failing test and fix the compiler/VM to handle that case.

- [ ] **Step 3: Run benchmarks to confirm speedup**

```bash
go test ./interpreter/... \
  -bench="BenchmarkAdvancedNestedFunctionCalls|BenchmarkAdvancedMapOperations|BenchmarkAdvancedComplexDataStructures" \
  -benchmem -benchtime=3s -count=3
```

Expected: ~5-10x improvement over the Phase 0 baseline.

- [ ] **Step 4: Commit**

```bash
git add interpreter/config.go
git commit -m "feat(vm): enable bytecode VM by default — all tests pass"
```

---

## Task 11 (Phase 1.5): Optional Type Hint Parser + Typed Opcodes

**Files:**
- Modify: `interpreter/tokenizer.go` (add `:` in param context if not already handled)
- Modify: `interpreter/parser.go` (parse `: type` after param name)
- Modify: `interpreter/compiler.go` (emit OP_ADD_INT etc. when type is known)

- [ ] **Step 1: Write test**

```go
func TestVMTypedIntAdd(t *testing.T) {
	src := `
fn add(a: int, b: int): int
    return a + b
end
add(3, 4)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	// Verify that the inner function uses OP_ADD_INT not OP_ADD
	innerFn := fn.SubFunctions[0]
	hasAddInt := false
	for _, instr := range innerFn.Code {
		if instr.Op == OP_ADD_INT {
			hasAddInt = true
		}
	}
	if !hasAddInt {
		t.Error("expected OP_ADD_INT for int+int annotated function, got generic OP_ADD")
	}

	result := makeVM(TestConfig()).Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}
```

- [ ] **Step 2: Add type hint parsing to parser.go**

In the function parameter parser, after reading the parameter name, optionally consume `: typename`:

```go
// In parseParameters() or equivalent:
param := p.val
p.next()
var hint TypeHint = TypeUnknown
if p.tok == COLON {
	p.next() // consume ':'
	hint = parseTypeHint(p.val)
	p.next()
}
params = append(params, param)
hints = append(hints, hint)
```

```go
func parseTypeHint(name string) TypeHint {
	switch name {
	case "int":    return TypeInt
	case "float":  return TypeFloat
	case "string": return TypeString
	case "bool":   return TypeBool
	case "array":  return TypeArray
	case "map":    return TypeMap
	case "void":   return TypeVoid
	default:       return TypeUnknown
	}
}
```

- [ ] **Step 3: Propagate TypeHints to FunctionDef AST node**

Add `ParamHints []TypeHint` to the `*FunctionDef` AST type (check `ast.go`).

- [ ] **Step 4: Use hints in compiler to emit typed opcodes**

In `compileFunctionDef`, store param types:

```go
sub.fn.ParamTypes = s.ParamHints
```

In `compileBinary`, when both operands are known-int params:

```go
func (c *Compiler) compileBinary(e *Binary) (uint8, error) {
	left, err := c.compileExpr(e.Left)
	if err != nil { return 0, err }
	right, err := c.compileExpr(e.Right)
	if err != nil { return 0, err }

	op, ok := tokenToOpCode[e.Operator]
	if !ok {
		return 0, fmt.Errorf("unsupported binary op %v", e.Operator)
	}

	// Promote to typed int opcode if both operands have int type hints
	if op == OP_ADD && c.isIntReg(left) && c.isIntReg(right) { op = OP_ADD_INT }
	if op == OP_SUB && c.isIntReg(left) && c.isIntReg(right) { op = OP_SUB_INT }
	// ... similarly for MUL, DIV, MOD

	dst := c.allocReg()
	c.emit(Instruction{Op: op, Dst: dst, Src1: left, Src2: right})
	if op == OP_ADD_INT { c.markIntReg(dst) }
	return dst, nil
}
```

Add register type tracking to `Compiler`:

```go
type Compiler struct {
	// ...
	regTypes map[uint8]TypeHint // register → known type (TypeUnknown = untyped)
}

func (c *Compiler) isIntReg(r uint8) bool {
	return c.regTypes[r] == TypeInt
}

func (c *Compiler) markIntReg(r uint8) {
	if c.regTypes == nil {
		c.regTypes = make(map[uint8]TypeHint)
	}
	c.regTypes[r] = TypeInt
}
```

Initialize param registers with their type hints in `compileFunctionDef`:

```go
for i, hint := range s.ParamHints {
	if hint == TypeInt {
		sub.markIntReg(uint8(i))
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./interpreter/... -run "TestVMTypedIntAdd" -v
go test ./... -count=1 2>&1 | tail -10
```

Expected: `TestVMTypedIntAdd` PASS, all others PASS.

- [ ] **Step 6: Commit**

```bash
git add interpreter/tokenizer.go interpreter/parser.go interpreter/ast.go interpreter/compiler.go
git commit -m "feat(vm/phase1.5): optional type hints emit typed opcodes (OP_ADD_INT etc.)"
```

---

## Phase 1 Complete

Run final benchmark comparison:

```bash
go test ./interpreter/... \
  -bench="BenchmarkAdvancedNestedFunctionCalls|BenchmarkAdvancedMapOperations|BenchmarkAdvancedComplexDataStructures|BenchmarkAdvancedStringConcatenation" \
  -benchmem -benchtime=3s -count=5 2>&1
```

Record results. Target: ≥5x improvement on each benchmark vs Phase 0 baseline.

Next plan: `docs/superpowers/plans/2026-05-25-phase2-llvm-jit.md`
