package interpreter

import "strconv"

const vmInitialRegCount = 256
const vmMaxFrames = 512

// vmFunction wraps a compiled Function so it satisfies the functionType interface.
type vmFunction struct {
	fn *Function
	vm *VM
}

func (f *vmFunction) call(interp *interpreter, pos Position, args []Value) Value {
	newBase := 0
	if len(f.vm.frames) > 0 {
		top := f.vm.frames[len(f.vm.frames)-1]
		newBase = top.baseReg + int(top.fn.MaxRegs) + 1
	}
	f.vm.pushFrame(f.fn, newBase, 0)
	newRegs := f.vm.regs[newBase:]
	for i, arg := range args {
		if i < int(f.fn.MaxRegs) {
			newRegs[i] = arg
		}
	}
	return f.vm.run()
}

func (f *vmFunction) name() string { return f.fn.Name }

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
	if len(vm.frames) >= vmMaxFrames {
		panic("VM: call stack overflow (max " + strconv.Itoa(vmMaxFrames) + " frames)")
	}
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

		// Generic arithmetic (delegates to package-level eval functions or interpreter method)
		case OP_ADD:
			regs[instr.Dst] = vm.interp.evalPlus(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_SUB:
			regs[instr.Dst] = evalMinus(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_MUL:
			regs[instr.Dst] = evalTimes(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_DIV:
			regs[instr.Dst] = evalDivide(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_MOD:
			regs[instr.Dst] = evalModulo(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_POW:
			regs[instr.Dst] = evalPower(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_NEG:
			regs[instr.Dst] = evalNegative(Position{}, regs[instr.Src1])

		// Typed integer arithmetic (no boxing, no type conversion)
		case OP_ADD_INT:
			regs[instr.Dst] = regs[instr.Src1].(int) + regs[instr.Src2].(int)
		case OP_SUB_INT:
			regs[instr.Dst] = regs[instr.Src1].(int) - regs[instr.Src2].(int)
		case OP_MUL_INT:
			regs[instr.Dst] = regs[instr.Src1].(int) * regs[instr.Src2].(int)
		case OP_DIV_INT:
			r := regs[instr.Src2].(int)
			if r == 0 {
				panic(runtimeError(Position{}, "VM: integer divide by zero"))
			}
			regs[instr.Dst] = regs[instr.Src1].(int) / r
		case OP_MOD_INT:
			r := regs[instr.Src2].(int)
			if r == 0 {
				panic(runtimeError(Position{}, "VM: integer modulo by zero"))
			}
			regs[instr.Dst] = regs[instr.Src1].(int) % r

		// Comparisons
		case OP_EQ:
			regs[instr.Dst] = evalEqual(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_NEQ:
			eq := evalEqual(Position{}, regs[instr.Src1], regs[instr.Src2])
			regs[instr.Dst] = !eq.(bool)
		case OP_LT:
			regs[instr.Dst] = evalLess(Position{}, regs[instr.Src1], regs[instr.Src2])
		case OP_LTE:
			gt := evalLess(Position{}, regs[instr.Src2], regs[instr.Src1])
			regs[instr.Dst] = !gt.(bool)
		case OP_GT:
			regs[instr.Dst] = evalLess(Position{}, regs[instr.Src2], regs[instr.Src1])
		case OP_GTE:
			lt := evalLess(Position{}, regs[instr.Src1], regs[instr.Src2])
			regs[instr.Dst] = !lt.(bool)
		case OP_IN:
			regs[instr.Dst] = evalIn(Position{}, regs[instr.Src1], regs[instr.Src2])

		// Logic
		case OP_AND:
			regs[instr.Dst] = IsTruthy(regs[instr.Src1]) && IsTruthy(regs[instr.Src2])
		case OP_OR:
			regs[instr.Dst] = IsTruthy(regs[instr.Src1]) || IsTruthy(regs[instr.Src2])
		case OP_NOT:
			regs[instr.Dst] = !IsTruthy(regs[instr.Src1])
		case OP_XOR:
			regs[instr.Dst] = IsTruthy(regs[instr.Src1]) != IsTruthy(regs[instr.Src2])

		// Control flow — jumps use signed 16-bit offset in Dst:Src2
		case OP_JUMP:
			frame.pc += instr.jumpOffset()

		case OP_JUMP_FALSE:
			if !IsTruthy(regs[instr.Src1]) {
				frame.pc += instr.jumpOffset()
			}

		case OP_JUMP_TRUE:
			if IsTruthy(regs[instr.Src1]) {
				frame.pc += instr.jumpOffset()
			}

		// Collections
		case OP_MAKE_ARRAY:
			n := int(instr.Src2)
			arr := make([]Value, n)
			for i := 0; i < n; i++ {
				arr[i] = regs[int(instr.Src1)+i]
			}
			regs[instr.Dst] = &arr

		case OP_MAKE_MAP:
			n := int(instr.Src2)
			m := make(map[string]Value, n)
			base := int(instr.Src1)
			for i := 0; i < n; i++ {
				key := regs[base+i*2]
				val := regs[base+i*2+1]
				k, ok := key.(string)
				if !ok {
					panic(runtimeError(Position{}, "VM: map key must be string"))
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
			switch cv := container.(type) {
			case *[]Value:
				idx, ok := key.(int)
				if !ok {
					panic(runtimeError(Position{}, "VM: array index must be integer"))
				}
				if idx < 0 {
					idx = len(*cv) + idx
				}
				if idx < 0 || idx >= len(*cv) {
					panic(runtimeError(Position{}, "VM: array index %d out of bounds (len=%d)", idx, len(*cv)))
				}
				(*cv)[idx] = val
			case map[string]Value:
				k, ok := key.(string)
				if !ok {
					panic(runtimeError(Position{}, "VM: map key must be string"))
				}
				cv[k] = val
			default:
				panic(runtimeError(Position{}, "VM: SET_INDEX on non-container"))
			}

		case OP_MAKE_FUNC:
			subIdx := int(instr.Src1)<<8 | int(instr.Src2)
			subfn := frame.fn.SubFunctions[subIdx]
			regs[instr.Dst] = &vmFunction{fn: subfn, vm: vm}

		case OP_CALL:
			fnVal := regs[instr.Src1]
			argc := int(instr.Src2)
			// Meta-instruction: Dst=argc, Src1:Src2=argBase (mirrors OP_CALL_BUILTIN).
			metaInstr := frame.fn.Code[frame.pc]
			frame.pc++
			argBase := int(metaInstr.Src1)<<8 | int(metaInstr.Src2)

			args := make([]Value, argc)
			for i := 0; i < argc; i++ {
				args[i] = regs[argBase+i]
			}

			switch f := fnVal.(type) {
			case *vmFunction:
				// Inline dispatch: push a new frame. The loop will re-fetch frame/regs/consts.
				newBase := frame.baseReg + int(frame.fn.MaxRegs) + 1
				vm.pushFrame(f.fn, newBase, int(instr.Dst))
				newRegs := vm.regs[newBase:]
				for i, arg := range args {
					if i < int(f.fn.MaxRegs) {
						newRegs[i] = arg
					}
				}
			case functionType:
				regs[instr.Dst] = f.call(vm.interp, Position{}, args)
			default:
				panic(runtimeError(Position{}, "VM: cannot call non-function %T", fnVal))
			}

		case OP_CALL_BUILTIN:
			builtinIdx := int(instr.Src1)<<8 | int(instr.Src2)
			metaInstr := frame.fn.Code[frame.pc]
			frame.pc++
			argc := int(metaInstr.Dst)
			argBase := int(metaInstr.Src1)<<8 | int(metaInstr.Src2)
			args := make([]Value, argc)
			for i := 0; i < argc; i++ {
				args[i] = regs[argBase+i]
			}
			entry := vmBuiltinTable[builtinIdx]
			regs[instr.Dst] = entry.fn(vm.interp, Position{}, args)

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
