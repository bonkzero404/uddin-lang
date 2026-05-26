package interpreter

const vmInitialRegCount = 256
const vmMaxFrames = 512

// VM executes compiled bytecode Functions.
type VM struct {
	regs   []Value
	frames []Frame
	interp *interpreter
}

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

// run is the main dispatch loop. It processes one instruction at a time across
// all frames, returning when the call stack is empty.
func (vm *VM) run() Value {
	for len(vm.frames) > 0 {
		frame := &vm.frames[len(vm.frames)-1]
		if frame.pc >= len(frame.fn.Code) {
			// Implicit nil return when execution falls off the end.
			result := vm.doReturn(nil)
			if len(vm.frames) == 0 {
				return result
			}
			continue
		}

		instr := frame.fn.Code[frame.pc]
		frame.pc++

		regs := vm.regs[frame.baseReg:]
		consts := frame.fn.Constants

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

		default:
			panic(runtimeError(Position{}, "VM: unimplemented opcode %s", opName(instr.Op)))
		}
	}
	return nil
}

// doReturn pops the top frame and, for non-top-level calls, writes the return
// value into the caller's register window.
func (vm *VM) doReturn(val Value) Value {
	top := vm.frames[len(vm.frames)-1]
	vm.frames = vm.frames[:len(vm.frames)-1]
	if len(vm.frames) == 0 {
		return val
	}
	callerBase := vm.frames[len(vm.frames)-1].baseReg
	vm.regs[callerBase+top.retReg] = val
	return val
}
