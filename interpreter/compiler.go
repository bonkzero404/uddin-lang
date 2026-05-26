package interpreter

import "fmt"

// Compiler compiles a parsed AST into bytecode Functions.
type Compiler struct {
	fn      *Function
	locals  map[string]uint8
	nextReg uint8
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
	case *ExpressionStatement:
		_, err := c.compileExpr(s.Expression)
		return err
	case *Return:
		return c.compileReturn(s)
	case *If:
		return c.compileIf(s)
	case *While:
		return c.compileWhile(s)
	case *FunctionDefinition:
		return c.compileFunctionDef(s)
	default:
		return fmt.Errorf("compiler: unsupported statement type %T", stmt)
	}
}

func (c *Compiler) compileAssign(s *Assign) error {
	// Only support simple variable assignment for now (Target must be a *Variable)
	v, ok := s.Target.(*Variable)
	if !ok {
		return fmt.Errorf("compiler: unsupported assignment target type %T", s.Target)
	}
	src, err := c.compileExpr(s.Value)
	if err != nil {
		return err
	}
	dst := c.localReg(v.Name)
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
func (c *Compiler) compileFunctionDef(s *FunctionDefinition) error {
	sub := &Compiler{
		fn:     &Function{Name: s.Name, Params: s.Parameters},
		locals: make(map[string]uint8),
	}
	// Parameters occupy the first registers in the sub-function's window.
	for i, param := range s.Parameters {
		sub.locals[param] = uint8(i)
		sub.nextReg++
	}
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

	// Register the sub-function in the parent and bind it to a local register.
	subIdx := uint16(len(c.fn.SubFunctions))
	c.fn.SubFunctions = append(c.fn.SubFunctions, sub.fn)
	dst := c.localReg(s.Name)
	c.emit(Instruction{Op: OP_MAKE_FUNC, Dst: dst, Src1: uint8(subIdx >> 8), Src2: uint8(subIdx)})
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
	argRegs := make([]uint8, len(e.Arguments))
	for i, arg := range e.Arguments {
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
		c.emit(Instruction{Op: OP_CALL_BUILTIN, Dst: dst,
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

// compileWhile compiles a while loop.
// *While fields: Condition Expression, Body Block
func (c *Compiler) compileWhile(s *While) error {
	loopStart := len(c.fn.Code)

	cond, err := c.compileExpr(s.Condition)
	if err != nil {
		return err
	}
	// Jump past loop body when condition is false.
	exitJump := c.emitJump(OP_JUMP_FALSE, cond)

	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Unconditional back-jump to loop start.
	offset := int16(loopStart - len(c.fn.Code) - 1)
	c.emit(Instruction{
		Op:   OP_JUMP,
		Dst:  uint8(uint16(offset) >> 8),
		Src2: uint8(offset),
	})

	// Patch exit jump to here (past the back-jump).
	c.patchJump(exitJump)
	return nil
}
