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
