package interpreter

import (
	"fmt"
	"os"

	"github.com/coregx/coregex"
)

// isRegexBuiltin returns true for builtins that accept a regex pattern.
func isRegexBuiltin(name string) bool {
	switch name {
	case "is_regex_match", "regex_match", "regex_find", "regex_find_all", "regex_replace", "regex_split":
		return true
	}
	return false
}

// regexPatternArgIndex returns the argument index that holds the pattern for each regex builtin.
// is_regex_match(pattern, str) → 0; all others (text, pattern) → 1.
func regexPatternArgIndex(name string) int {
	if name == "is_regex_match" {
		return 0
	}
	return 1
}

// Compiler compiles a parsed AST into bytecode Functions.
type Compiler struct {
	fn      *Function
	locals  map[string]uint8
	nextReg uint8

	// Loop control: slices of instruction indices that need patching.
	// breakTargets holds indices of OP_JUMP instructions emitted by 'break'.
	// continueTargets holds indices of OP_JUMP instructions emitted by 'continue'.
	// Each loop saves/restores these slices around its body compilation.
	breakTargets    []int
	continueTargets []int

	// selfName is the name of the function currently being compiled, used to
	// resolve recursive self-calls within compileFunctionDef.
	selfName string

	// preSeededVars lists variable names that should be pre-allocated as locals
	// at the start of Compile (before any source statements are compiled).
	// This allows Config.Vars to be visible to the compiler.
	preSeededVars []string

	// regTypes tracks the TypeHint for each register so the compiler can emit
	// typed opcodes (e.g. OP_ADD_INT) when both operands are known-int.
	regTypes map[uint8]TypeHint

	// parent is the enclosing Compiler scope, used for upvalue (free variable) resolution.
	parent *Compiler

	// upvalues maps variable names to their upvalue index in fn.Upvalues.
	// Populated lazily as free variables are encountered during compilation.
	upvalues map[string]uint8
}

// isIntReg reports whether register r is known to hold an int value.
func (c *Compiler) isIntReg(r uint8) bool {
	return c.regTypes != nil && c.regTypes[r] == HintInt
}

// markIntReg records that register r holds an int value.
func (c *Compiler) markIntReg(r uint8) {
	if c.regTypes == nil {
		c.regTypes = make(map[uint8]TypeHint)
	}
	c.regTypes[r] = HintInt
}

// resolveUpvalue searches parent compiler scopes for the variable name.
// If found, records it as an upvalue in fn.Upvalues and returns its index.
// Returns an error if the variable is not found in any enclosing scope.
func (c *Compiler) resolveUpvalue(name string) (uint8, error) {
	if c.parent == nil {
		return 0, fmt.Errorf("not found")
	}
	// Check if already recorded as an upvalue in this compiler.
	if c.upvalues == nil {
		c.upvalues = make(map[string]uint8)
	}
	if idx, ok := c.upvalues[name]; ok {
		return idx, nil
	}
	// Check immediate parent's locals.
	if reg, ok := c.parent.locals[name]; ok {
		idx := uint8(len(c.fn.Upvalues))
		c.fn.Upvalues = append(c.fn.Upvalues, UpvalueInfo{Name: name, IsLocal: true, Index: reg})
		c.upvalues[name] = idx
		return idx, nil
	}
	// Recurse: check parent's upvalues (for deeper nesting).
	if uvIdx, err := c.parent.resolveUpvalue(name); err == nil {
		idx := uint8(len(c.fn.Upvalues))
		c.fn.Upvalues = append(c.fn.Upvalues, UpvalueInfo{Name: name, IsLocal: false, Index: uvIdx})
		c.upvalues[name] = idx
		return idx, nil
	}
	return 0, fmt.Errorf("not found")
}

// NewCompiler creates a fresh Compiler instance.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile compiles an entire program into a top-level Function.
func (c *Compiler) Compile(prog *Program) (*Function, error) {
	c.fn = &Function{Name: "<main>"}
	c.locals = make(map[string]uint8)
	c.nextReg = 0

	// Pre-allocate registers for any externally-injected variables so that
	// source code can reference them without an explicit assignment first.
	for _, name := range c.preSeededVars {
		c.localReg(name)
	}

	// lastExprReg tracks the register holding the last expression-statement
	// result so the top-level program returns it implicitly (REPL semantics).
	lastExprReg := int(-1)
	for _, stmt := range prog.Statements {
		if es, ok := stmt.(*ExpressionStatement); ok {
			r, err := c.compileExpr(es.Expression)
			if err != nil {
				return nil, err
			}
			lastExprReg = int(r)
		} else {
			if err := c.compileStatement(stmt); err != nil {
				return nil, err
			}
			lastExprReg = -1
		}
	}

	// Implicit return: use last expression value if available, else nil.
	var r uint8
	if lastExprReg >= 0 {
		r = uint8(lastExprReg)
	} else {
		nilIdx := c.addConst(nil)
		r = c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(nilIdx >> 8), Src2: uint8(nilIdx)})
	}
	c.emit(Instruction{Op: OP_RETURN, Src1: r})
	c.fn.MaxRegs = c.nextReg
	return c.fn, nil
}

func (c *Compiler) emit(instr Instruction) int {
	c.fn.Code = append(c.fn.Code, instr)
	return len(c.fn.Code) - 1
}

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

func (c *Compiler) allocReg() uint8 {
	r := c.nextReg
	c.nextReg++
	if c.nextReg > c.fn.MaxRegs {
		c.fn.MaxRegs = c.nextReg
	}
	return r
}

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
	case *OuterAssign:
		return c.compileOuterAssign(s)
	case *ExpressionStatement:
		_, err := c.compileExpr(s.Expression)
		return err
	case *Return:
		return c.compileReturn(s)
	case *If:
		return c.compileIf(s)
	case *While:
		return c.compileWhile(s)
	case *For:
		return c.compileFor(s)
	case *Break:
		return c.compileBreak(s)
	case *Continue:
		return c.compileContinue(s)
	case *FunctionDefinition:
		return c.compileFunctionDef(s)
	case *TryCatch:
		return c.compileTryCatch(s)
	case *Import:
		return c.compileImport(s)
	default:
		return fmt.Errorf("compiler: unsupported statement type %T", stmt)
	}
}

func (c *Compiler) compileAssign(s *Assign) error {
	switch target := s.Target.(type) {
	case *Variable:
		return c.compileVarAssign(target, s.Value, s.Operator)
	case *Subscript:
		return c.compileSubscriptAssign(target, s.Value, s.Operator)
	default:
		return fmt.Errorf("compiler: unsupported assignment target type %T", s.Target)
	}
}

// compileVarAssign handles simple variable assignment (x = expr) and compound
// variable assignment (x += expr, x -= expr, etc.).
func (c *Compiler) compileVarAssign(v *Variable, value Expression, op Token) error {
	src, err := c.compileExpr(value)
	if err != nil {
		return err
	}

	if op == ASSIGN {
		dst := c.localReg(v.Name)
		c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: src})
		return nil
	}

	// Compound assignment: load variable, apply op, store back.
	binOp, err := compoundOpToOpCode(op)
	if err != nil {
		return err
	}

	if varReg, ok := c.locals[v.Name]; ok {
		curReg := c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_VAR, Dst: curReg, Src1: varReg})
		resultReg := c.allocReg()
		c.emit(Instruction{Op: binOp, Dst: resultReg, Src1: curReg, Src2: src})
		c.emit(Instruction{Op: OP_STORE_VAR, Dst: varReg, Src1: resultReg})
		return nil
	}

	// Try upvalue compound assignment.
	if uvIdx, err2 := c.resolveUpvalue(v.Name); err2 == nil {
		curReg := c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_UPVAL, Dst: curReg, Src1: uvIdx})
		resultReg := c.allocReg()
		c.emit(Instruction{Op: binOp, Dst: resultReg, Src1: curReg, Src2: src})
		c.emit(Instruction{Op: OP_STORE_UPVAL, Dst: uvIdx, Src1: resultReg})
		return nil
	}

	return fmt.Errorf("compiler: undefined variable %q in compound assignment", v.Name)
}

// compileSubscriptAssign handles subscript assignment (a[i] = expr) and compound
// subscript assignment (a[i] += expr, etc.).
func (c *Compiler) compileSubscriptAssign(sub *Subscript, value Expression, op Token) error {
	containerReg, err := c.compileExpr(sub.Container)
	if err != nil {
		return err
	}
	keyReg, err := c.compileExpr(sub.Subscript)
	if err != nil {
		return err
	}
	valReg, err := c.compileExpr(value)
	if err != nil {
		return err
	}

	if op != ASSIGN {
		// Load current element, apply op, use result as the new value.
		curReg := c.allocReg()
		c.emit(Instruction{Op: OP_SUBSCRIPT, Dst: curReg, Src1: containerReg, Src2: keyReg})
		binOp, err := compoundOpToOpCode(op)
		if err != nil {
			return err
		}
		newVal := c.allocReg()
		c.emit(Instruction{Op: binOp, Dst: newVal, Src1: curReg, Src2: valReg})
		valReg = newVal
	}

	// OP_SET_INDEX: regs[Dst] is container, Src1 is key, Src2 is value.
	c.emit(Instruction{Op: OP_SET_INDEX, Dst: containerReg, Src1: keyReg, Src2: valReg})
	return nil
}

// compoundOpToOpCode converts a compound assignment token to the corresponding
// binary opcode.
func compoundOpToOpCode(op Token) (OpCode, error) {
	switch op {
	case PLUSEQUAL:
		return OP_ADD, nil
	case MINUSEQUAL:
		return OP_SUB, nil
	case TIMESEQUAL:
		return OP_MUL, nil
	case DIVIDEEQUAL:
		return OP_DIV, nil
	case MODULOEQUAL:
		return OP_MOD, nil
	default:
		return 0, fmt.Errorf("compiler: unknown compound assignment operator %v", op)
	}
}

// compileOuterAssign handles 'outer x = expr' — in the flat register-based VM
// the outer variable is treated as a regular local. If the variable is known,
// we update its register; otherwise we create a new one.
func (c *Compiler) compileOuterAssign(s *OuterAssign) error {
	src, err := c.compileExpr(s.Value)
	if err != nil {
		return err
	}
	dst := c.localReg(s.Name)
	c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: src})
	return nil
}

func (c *Compiler) compileReturn(s *Return) error {
	if s.Result == nil {
		nilIdx := c.addConst(nil)
		r := c.allocReg()
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(nilIdx >> 8), Src2: uint8(nilIdx)})
		c.emit(Instruction{Op: OP_RETURN, Src1: r})
		return nil
	}
	src, err := c.compileExpr(s.Result)
	if err != nil {
		return err
	}
	c.emit(Instruction{Op: OP_RETURN, Src1: src})
	return nil
}

// tokenToOpCode maps binary operator tokens to their VM opcodes.
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
			// Propagate type hint so typed opcodes can be used downstream.
			if c.isIntReg(r) {
				c.markIntReg(dst)
			}
			return dst, nil
		}
		// Try to resolve as upvalue from an enclosing scope.
		if uvIdx, err2 := c.resolveUpvalue(e.Name); err2 == nil {
			dst := c.allocReg()
			c.emit(Instruction{Op: OP_LOAD_UPVAL, Dst: dst, Src1: uvIdx})
			return dst, nil
		}
		return 0, fmt.Errorf("compiler: undefined variable %q", e.Name)

	case *Binary:
		return c.compileBinary(e)

	case *Unary:
		return c.compileUnary(e)

	case *List:
		startReg := c.nextReg
		count := len(e.Values)
		for _, elem := range e.Values {
			_, err := c.compileExpr(elem)
			if err != nil {
				return 0, err
			}
		}
		dst := c.allocReg()
		c.emit(Instruction{Op: OP_MAKE_ARRAY, Dst: dst, Src1: uint8(startReg), Src2: uint8(count)})
		return dst, nil

	case *Map:
		startReg := c.nextReg
		count := len(e.Items)
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
		c.emit(Instruction{Op: OP_MAKE_MAP, Dst: dst, Src1: uint8(startReg), Src2: uint8(count)})
		return dst, nil

	case *Subscript:
		container, err := c.compileExpr(e.Container)
		if err != nil {
			return 0, err
		}
		key, err := c.compileExpr(e.Subscript)
		if err != nil {
			return 0, err
		}
		dst := c.allocReg()
		c.emit(Instruction{Op: OP_SUBSCRIPT, Dst: dst, Src1: container, Src2: key})
		return dst, nil

	case *Call:
		return c.compileCall(e)

	case *Ternary:
		return c.compileTernary(e)

	case *FunctionExpression:
		return c.compileFunctionExpr(e)

	default:
		return 0, fmt.Errorf("compiler: unsupported expression type %T", expr)
	}
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
	// Promote arithmetic ops to typed variants when both operands are known-int.
	if c.isIntReg(left) && c.isIntReg(right) {
		switch op {
		case OP_ADD:
			op = OP_ADD_INT
		case OP_SUB:
			op = OP_SUB_INT
		case OP_MUL:
			op = OP_MUL_INT
		case OP_DIV:
			op = OP_DIV_INT
		case OP_MOD:
			op = OP_MOD_INT
		}
	}
	dst := c.allocReg()
	c.emit(Instruction{Op: op, Dst: dst, Src1: left, Src2: right})
	// Propagate int type to the destination register for typed arithmetic.
	switch op {
	case OP_ADD_INT, OP_SUB_INT, OP_MUL_INT, OP_DIV_INT, OP_MOD_INT:
		c.markIntReg(dst)
	}
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

// emitJump emits a jump instruction with a placeholder offset (0).
// Returns the index of the emitted instruction for later patching.
func (c *Compiler) emitJump(op OpCode, condReg uint8) int {
	return c.emit(Instruction{Op: op, Src1: condReg, Dst: 0, Src2: 0})
}

// patchJump fills in the signed 16-bit forward jump offset for a previously
// emitted jump. The offset is relative to the instruction after the jump.
func (c *Compiler) patchJump(jumpIdx int) {
	offset := len(c.fn.Code) - jumpIdx - 1
	if offset > 32767 || offset < -32768 {
		panic("jump offset overflow")
	}
	o := int16(offset)
	c.fn.Code[jumpIdx].Dst = uint8(uint16(o) >> 8)
	c.fn.Code[jumpIdx].Src2 = uint8(o)
}

// compileIf compiles an if/else statement.
// *If fields: Condition Expression, Body Block, Else Block
func (c *Compiler) compileIf(s *If) error {
	cond, err := c.compileExpr(s.Condition)
	if err != nil {
		return err
	}
	// Jump past "then" block if condition is false.
	jumpFalse := c.emitJump(OP_JUMP_FALSE, cond)

	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	if len(s.Else) == 0 {
		// No else — patch the false-jump to here.
		c.patchJump(jumpFalse)
		return nil
	}

	// Emit unconditional jump to skip the else block after then-block executes.
	jumpEnd := c.emit(Instruction{Op: OP_JUMP})
	// The false-jump lands at the start of the else block.
	c.patchJump(jumpFalse)

	for _, stmt := range s.Else {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	// The then-path jump lands here, past the else block.
	c.patchJump(jumpEnd)
	return nil
}

// compileFunctionDef compiles a function definition into a sub-Function and emits OP_MAKE_FUNC.
// The function name is pre-registered in the parent's locals BEFORE the body is compiled, so
// that recursive self-calls resolve correctly.
func (c *Compiler) compileFunctionDef(s *FunctionDefinition) error {
	// Pre-allocate the destination register for the function value in the parent scope.
	// This makes the name visible to recursive calls inside the body.
	dst := c.localReg(s.Name)

	// The sub-index will be the next entry in SubFunctions.
	subIdx := uint16(len(c.fn.SubFunctions))

	// Emit a placeholder OP_MAKE_FUNC first.  We will append the sub-function
	// below — this ordering means self-recursive calls compile against the
	// already-allocated dst register, but the real Function pointer is attached
	// after body compilation (appended to SubFunctions at subIdx).
	c.fn.SubFunctions = append(c.fn.SubFunctions, nil) // placeholder
	makeIdx := c.emit(Instruction{Op: OP_MAKE_FUNC, Dst: dst, Src1: uint8(subIdx >> 8), Src2: uint8(subIdx)})
	_ = makeIdx // will be overwritten below (same index, same encoding)

	sub := &Compiler{
		fn:     &Function{Name: s.Name, Params: s.Parameters},
		locals: make(map[string]uint8),
		parent: c,
	}
	// Parameters occupy the first registers in the sub-function's window.
	for i, param := range s.Parameters {
		sub.locals[param] = uint8(i)
		sub.nextReg++
	}
	sub.fn.MaxRegs = sub.nextReg

	// Store param type hints on the Function for Phase 2 JIT use.
	sub.fn.ParamTypes = s.ParamHints

	// Mark int-annotated parameters so the sub-compiler can emit typed opcodes.
	for i, hint := range s.ParamHints {
		if hint == HintInt && i < len(s.Parameters) {
			sub.markIntReg(uint8(i))
		}
	}

	// Allow the sub-compiler to resolve recursive calls to this function.
	// We allocate a register in the sub-function's locals for the function itself
	// (mirrors what the tree-walker does: the function is in its own closure).
	sub.selfName = s.Name
	selfLocal := sub.allocReg()
	sub.locals[s.Name] = selfLocal
	sub.fn.MaxRegs = sub.nextReg

	for _, stmt := range s.Body {
		if err := sub.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Implicit nil return.
	nilIdx := sub.addConst(nil)
	r := sub.allocReg()
	sub.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(nilIdx >> 8), Src2: uint8(nilIdx)})
	sub.emit(Instruction{Op: OP_RETURN, Src1: r})
	sub.fn.MaxRegs = sub.nextReg

	// Fill in the placeholder with the real sub-function.
	c.fn.SubFunctions[subIdx] = sub.fn

	// Also, emit a self-assignment inside the sub-function: load the function
	// object from register 'dst' (parent) ... but since we are in a separate
	// register window, we cannot access the parent's register.
	// Instead we emit an OP_MAKE_FUNC *inside* the sub-function itself so the
	// sub's selfLocal is initialised to the correct vmFunction at call time.
	// Append the sub-function to its OWN SubFunctions list (self-reference).
	selfSubIdx := uint16(len(sub.fn.SubFunctions))
	sub.fn.SubFunctions = append(sub.fn.SubFunctions, sub.fn)
	// Prepend these two instructions to the sub-function body so they run before
	// any user code. We do this by inserting at the start of sub.fn.Code.
	initInstrs := []Instruction{
		{Op: OP_MAKE_FUNC, Dst: selfLocal, Src1: uint8(selfSubIdx >> 8), Src2: uint8(selfSubIdx)},
	}
	sub.fn.Code = append(initInstrs, sub.fn.Code...)

	return nil
}

// compileCall compiles a function call expression.
// Builtin calls emit OP_CALL_BUILTIN; user-defined calls emit OP_CALL.
// Both use a following meta-instruction to encode argc/argBase.
//
// Each argument is compiled and its result copied (via OP_STORE_VAR) into a
// contiguous argument window starting at argStart, so the callee sees a
// clean sequential block of registers regardless of how many intermediate
// registers each argument expression consumed.
func (c *Compiler) compileCall(e *Call) (uint8, error) {
	// Determine if this is a direct builtin call.
	builtinCall := false
	var builtinIdx uint16
	if fnExpr, isVar := e.Function.(*Variable); isVar {
		if idx, ok := vmBuiltinIndex[fnExpr.Name]; ok {
			builtinCall = true
			builtinIdx = idx
		}
	}

	// For user calls, compile the function expression first so its register
	// does not overlap with the argument window.
	var fnReg uint8
	if !builtinCall {
		var err error
		fnReg, err = c.compileExpr(e.Function)
		if err != nil {
			return 0, err
		}
	}

	// Compile each argument; collect result registers.
	// For regex builtins with a string-literal pattern argument, pre-compile
	// the pattern to *coregex.Regexp and store in constants — eliminates
	// per-call compilation from the hot path.
	argRegs := make([]uint8, len(e.Arguments))
	var regexFnName string
	if builtinCall {
		if fnExpr, isVar := e.Function.(*Variable); isVar && isRegexBuiltin(fnExpr.Name) {
			regexFnName = fnExpr.Name
		}
	}
	for i, arg := range e.Arguments {
		if regexFnName != "" && i == regexPatternArgIndex(regexFnName) {
			if lit, ok := arg.(*Literal); ok {
				if patStr, ok := lit.Value.(string); ok {
					if re, err := coregex.Compile(patStr); err == nil {
						constIdx := uint16(len(c.fn.Constants))
						c.fn.Constants = append(c.fn.Constants, re)
						r := c.allocReg()
						c.emit(Instruction{Op: OP_LOAD_CONST, Dst: r,
							Src1: uint8(constIdx >> 8), Src2: uint8(constIdx)})
						argRegs[i] = r
						continue
					}
					// Invalid pattern — fall through to normal compile.
				}
			}
		}
		r, err := c.compileExpr(arg)
		if err != nil {
			return 0, err
		}
		argRegs[i] = r
	}

	// Widen argStart to uint16 before encoding so that Src1:Src2 carries the
	// full 16-bit base register index. Shifting a uint8 by 8 always yields 0,
	// so the widening must happen before the shift.
	argStart := uint16(c.nextReg)

	// Copy args into a contiguous window starting at argStart.
	for _, r := range argRegs {
		dst := c.allocReg()
		c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: r})
	}

	if builtinCall {
		dst := c.allocReg()
		op := OP_CALL_BUILTIN
		if int(builtinIdx) < len(vmBuiltinDirectTable) && vmBuiltinDirectTable[builtinIdx] != nil {
			op = OP_CALL_BUILTIN_DIRECT
		}
		c.emit(Instruction{Op: op, Dst: dst,
			Src1: uint8(builtinIdx >> 8), Src2: uint8(builtinIdx)})
		// Meta-instruction: argc in Dst, argStart in Src1:Src2.
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: uint8(len(e.Arguments)),
			Src1: uint8(argStart >> 8), Src2: uint8(argStart)})
		return dst, nil
	}

	// User-defined function call.
	// Encoding matches OP_CALL_BUILTIN: Src1:Src2=argBase, Dst=argc in meta-instruction.
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_CALL, Dst: dst, Src1: fnReg, Src2: uint8(len(e.Arguments))})
	// Meta-instruction: argc in Dst, argStart in Src1:Src2 (mirrors OP_CALL_BUILTIN).
	c.emit(Instruction{Op: OP_LOAD_CONST, Dst: uint8(len(e.Arguments)),
		Src1: uint8(argStart >> 8), Src2: uint8(argStart)})
	return dst, nil
}

// compileWhile compiles a while loop, supporting break and continue.
// *While fields: Condition Expression, Body Block
func (c *Compiler) compileWhile(s *While) error {
	loopStart := len(c.fn.Code)

	cond, err := c.compileExpr(s.Condition)
	if err != nil {
		return err
	}
	// Jump past loop body when condition is false.
	exitJump := c.emitJump(OP_JUMP_FALSE, cond)

	// Save outer break/continue targets and start fresh for this loop.
	savedBreaks := c.breakTargets
	savedContinues := c.continueTargets
	c.breakTargets = nil
	c.continueTargets = nil

	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Patch all 'continue' jumps to point to the back-edge (condition re-eval).
	continueTarget := len(c.fn.Code)
	for _, idx := range c.continueTargets {
		offset := continueTarget - idx - 1
		o := int16(offset)
		c.fn.Code[idx].Dst = uint8(uint16(o) >> 8)
		c.fn.Code[idx].Src2 = uint8(o)
	}

	// Unconditional back-jump to loop start (re-evaluate condition).
	backOffset := int16(loopStart - len(c.fn.Code) - 1)
	c.emit(Instruction{
		Op:   OP_JUMP,
		Dst:  uint8(uint16(backOffset) >> 8),
		Src2: uint8(backOffset),
	})

	// Patch exit jump and all 'break' jumps to here (past the back-jump).
	exitTarget := len(c.fn.Code)
	c.patchJump(exitJump)
	for _, idx := range c.breakTargets {
		offset := exitTarget - idx - 1
		o := int16(offset)
		c.fn.Code[idx].Dst = uint8(uint16(o) >> 8)
		c.fn.Code[idx].Src2 = uint8(o)
	}

	// Restore outer loop's break/continue lists.
	c.breakTargets = savedBreaks
	c.continueTargets = savedContinues
	return nil
}

// compileBreak emits an OP_JUMP with a placeholder offset and records the
// instruction index so the enclosing loop can patch it later.
func (c *Compiler) compileBreak(s *Break) error {
	idx := c.emit(Instruction{Op: OP_JUMP})
	c.breakTargets = append(c.breakTargets, idx)
	return nil
}

// compileContinue emits an OP_JUMP with a placeholder offset and records the
// instruction index so the enclosing loop can patch it to the back-edge.
func (c *Compiler) compileContinue(s *Continue) error {
	idx := c.emit(Instruction{Op: OP_JUMP})
	c.continueTargets = append(c.continueTargets, idx)
	return nil
}

// compileFor compiles a for-in loop over an array or map, supporting break and continue.
// *For fields: Name string, Iterable Expression, Body Block
//
// Emitted register layout (all temporary):
//
//	iterReg   — holds the iterable (array or map keys)
//	counterReg — integer counter (0-based index)
//	lenReg     — length of the iterable (constant across iterations)
//	elemReg    — current element (loop variable, Name)
//	condReg    — comparison result (counter < len)
func (c *Compiler) compileFor(s *For) error {
	// Evaluate the iterable once.
	iterReg, err := c.compileExpr(s.Iterable)
	if err != nil {
		return err
	}

	// counter = 0
	counterReg := c.allocReg()
	zeroIdx := c.addConst(0)
	c.emit(Instruction{Op: OP_LOAD_CONST, Dst: counterReg, Src1: uint8(zeroIdx >> 8), Src2: uint8(zeroIdx)})

	// len(iterable) — call the builtin len
	lenBuiltinIdx := vmBuiltinIndex["len"]
	// Copy iterable into a contiguous arg window.
	argStart := uint16(c.nextReg)
	argCopy := c.allocReg()
	c.emit(Instruction{Op: OP_STORE_VAR, Dst: argCopy, Src1: iterReg})
	lenReg := c.allocReg()
	c.emit(Instruction{Op: OP_CALL_BUILTIN, Dst: lenReg,
		Src1: uint8(lenBuiltinIdx >> 8), Src2: uint8(lenBuiltinIdx)})
	c.emit(Instruction{Op: OP_LOAD_CONST, Dst: 1 /*argc*/,
		Src1: uint8(argStart >> 8), Src2: uint8(argStart)})

	// Allocate the loop variable register.
	elemReg := c.localReg(s.Name)

	// Loop header: check counter < len.
	loopStart := len(c.fn.Code)
	condReg := c.allocReg()
	c.emit(Instruction{Op: OP_LT, Dst: condReg, Src1: counterReg, Src2: lenReg})
	exitJump := c.emitJump(OP_JUMP_FALSE, condReg)

	// Load iterable[counter] → elemReg.
	c.emit(Instruction{Op: OP_SUBSCRIPT, Dst: elemReg, Src1: iterReg, Src2: counterReg})

	// Save outer break/continue targets.
	savedBreaks := c.breakTargets
	savedContinues := c.continueTargets
	c.breakTargets = nil
	c.continueTargets = nil

	// Compile body.
	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Continue target: increment counter then jump back to header.
	continueTarget := len(c.fn.Code)
	for _, idx := range c.continueTargets {
		offset := continueTarget - idx - 1
		o := int16(offset)
		c.fn.Code[idx].Dst = uint8(uint16(o) >> 8)
		c.fn.Code[idx].Src2 = uint8(o)
	}

	// counter = counter + 1
	oneIdx := c.addConst(1)
	oneReg := c.allocReg()
	c.emit(Instruction{Op: OP_LOAD_CONST, Dst: oneReg, Src1: uint8(oneIdx >> 8), Src2: uint8(oneIdx)})
	c.emit(Instruction{Op: OP_ADD, Dst: counterReg, Src1: counterReg, Src2: oneReg})

	// Jump back to loop header.
	backOffset := int16(loopStart - len(c.fn.Code) - 1)
	c.emit(Instruction{
		Op:   OP_JUMP,
		Dst:  uint8(uint16(backOffset) >> 8),
		Src2: uint8(backOffset),
	})

	// Patch exit and break jumps.
	exitTarget := len(c.fn.Code)
	c.patchJump(exitJump)
	for _, idx := range c.breakTargets {
		offset := exitTarget - idx - 1
		o := int16(offset)
		c.fn.Code[idx].Dst = uint8(uint16(o) >> 8)
		c.fn.Code[idx].Src2 = uint8(o)
	}

	c.breakTargets = savedBreaks
	c.continueTargets = savedContinues
	return nil
}

// compileTernary compiles a ternary expression (cond ? trueExpr : falseExpr).
// The result is written to a pre-allocated destination register.
func (c *Compiler) compileTernary(e *Ternary) (uint8, error) {
	cond, err := c.compileExpr(e.Condition)
	if err != nil {
		return 0, err
	}

	// Allocate the destination register before any branch so both branches
	// write into the same slot.
	dst := c.allocReg()

	// Jump to else if condition is false.
	jumpFalse := c.emitJump(OP_JUMP_FALSE, cond)

	// Then-branch: compile trueExpr and move result to dst.
	thenReg, err := c.compileExpr(e.TrueExpr)
	if err != nil {
		return 0, err
	}
	c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: thenReg})

	// Jump over else branch.
	jumpEnd := c.emit(Instruction{Op: OP_JUMP})
	c.patchJump(jumpFalse)

	// Else-branch: compile falseExpr and move result to dst.
	elseReg, err := c.compileExpr(e.FalseExpr)
	if err != nil {
		return 0, err
	}
	c.emit(Instruction{Op: OP_STORE_VAR, Dst: dst, Src1: elseReg})

	c.patchJump(jumpEnd)
	return dst, nil
}

// compileTryCatch compiles a try/catch block.
//
// Emitted bytecode layout:
//
//	OP_TRY      Src1=errReg, Dst:Src2=offset_to_catch_block
//	<try body>
//	OP_END_TRY                    ; no error — pop handler
//	OP_JUMP     Dst:Src2=offset_past_catch ; skip over catch block
//	<catch body>                  ; ← catch block starts here (OP_TRY jumps here on error)
//	                              ; ← skip jump lands here (past catch block)
func (c *Compiler) compileTryCatch(s *TryCatch) error {
	// Allocate (or reuse) the register that will hold the caught error message.
	errReg := c.localReg(s.ErrVar)

	// Emit OP_TRY with a placeholder offset; Src1 carries errReg.
	// emitJump(op, condReg) emits: Instruction{Op: op, Src1: condReg, Dst: 0, Src2: 0}
	tryIdx := c.emitJump(OP_TRY, errReg)

	// Compile try body.
	for _, stmt := range s.TryBlock {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Successful path: pop the handler and jump over the catch block.
	c.emit(Instruction{Op: OP_END_TRY})
	skipIdx := c.emit(Instruction{Op: OP_JUMP})

	// Patch OP_TRY so its offset points here (start of catch block).
	c.patchJump(tryIdx)

	// Compile catch body.
	for _, stmt := range s.CatchBlock {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Patch the skip-jump to here (past catch block).
	c.patchJump(skipIdx)
	return nil
}

// compileFunctionExpr compiles an anonymous function expression (fun(params) body end).
// It creates a sub-function and emits OP_MAKE_FUNC, returning the register holding
// the function value.
func (c *Compiler) compileFunctionExpr(e *FunctionExpression) (uint8, error) {
	sub := &Compiler{
		fn:     &Function{Name: "<anon>", Params: e.Parameters},
		locals: make(map[string]uint8),
		parent: c,
	}
	for i, param := range e.Parameters {
		sub.locals[param] = uint8(i)
		sub.nextReg++
	}
	sub.fn.MaxRegs = sub.nextReg

	for _, stmt := range e.Body {
		if err := sub.compileStatement(stmt); err != nil {
			return 0, err
		}
	}

	// Implicit nil return.
	nilIdx := sub.addConst(nil)
	r := sub.allocReg()
	sub.emit(Instruction{Op: OP_LOAD_CONST, Dst: r, Src1: uint8(nilIdx >> 8), Src2: uint8(nilIdx)})
	sub.emit(Instruction{Op: OP_RETURN, Src1: r})
	sub.fn.MaxRegs = sub.nextReg

	subIdx := uint16(len(c.fn.SubFunctions))
	c.fn.SubFunctions = append(c.fn.SubFunctions, sub.fn)
	dst := c.allocReg()
	c.emit(Instruction{Op: OP_MAKE_FUNC, Dst: dst, Src1: uint8(subIdx >> 8), Src2: uint8(subIdx)})
	return dst, nil
}

// compileImport reads, parses, and inlines an imported .din file into the
// current compiler at compile time. Imported symbols are visible in the same
// scope as if the file were written inline — matching tree-walker semantics.
func (c *Compiler) compileImport(s *Import) error {
	content, err := os.ReadFile(s.Filename)
	if err != nil {
		return fmt.Errorf("import '%s': %w", s.Filename, err)
	}
	prog, err := ParseProgram(content)
	if err != nil {
		return fmt.Errorf("import '%s': parse error: %w", s.Filename, err)
	}
	for _, stmt := range prog.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return fmt.Errorf("import '%s': %w", s.Filename, err)
		}
	}
	return nil
}
