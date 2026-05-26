package interpreter

// TypeHint represents an optional type annotation on a function parameter.
type TypeHint uint8

const (
	HintUnknown TypeHint = iota
	HintInt
	HintFloat
	HintString
	HintBool
	HintArray
	HintMap
	HintVoid
)

// UpvalueInfo describes a captured variable (closure upvalue).
type UpvalueInfo struct {
	Name    string
	IsLocal bool // true = captured from immediately enclosing function's locals
	Index   uint8
}

// Function is a compiled bytecode function (user-defined or top-level program).
type Function struct {
	Name         string
	Params       []string
	ParamTypes   []TypeHint // len == 0 means all untyped
	ReturnType   TypeHint
	Code         []Instruction
	Constants    []Value
	Upvalues     []UpvalueInfo
	SubFunctions []*Function // inner function definitions
	MaxRegs      uint8
}

// Frame is one activation record in the VM call stack.
type Frame struct {
	fn      *Function
	pc      int
	baseReg int // offset into vm.regs where this frame's register window starts
	retReg  int // register in the CALLER's window where the return value is stored
}
