package interpreter

// CompiledProgram bundles compiled bytecode with the register map for
// pre-seeded variables. Obtain via CompileProgram; pass to ExecuteCompiledVM.
type CompiledProgram struct {
	Fn      *Function
	varRegs map[string]uint8
}

// CompileProgram compiles prog into reusable bytecode.
// varNames lists all variable names that callers will inject via Config.Vars.
func CompileProgram(prog *Program, varNames []string) (*CompiledProgram, error) {
	c := NewCompiler()
	c.preSeededVars = append(c.preSeededVars, varNames...)
	fn, err := c.Compile(prog)
	if err != nil {
		return nil, err
	}
	return &CompiledProgram{Fn: fn, varRegs: c.locals}, nil
}

// ExecuteCompiledVM runs cp through the bytecode VM and returns stats.
// Pass a pooled *VM to avoid allocation; pass nil to allocate a fresh one.
// The VM is reset before use — callers may return it to a pool afterward.
func ExecuteCompiledVM(cp *CompiledProgram, config *Config, vm *VM) (stats *Stats, err error) {
	interp := newInterpreter(config)
	if vm == nil {
		vm = NewVM(interp)
	} else {
		vm.reset(interp)
	}

	// Inject Config.Vars into pre-seeded registers.
	if config != nil {
		for k, v := range config.Vars {
			if reg, ok := cp.varRegs[k]; ok {
				needed := int(reg) + 1
				for len(vm.regs) < needed {
					vm.regs = append(vm.regs, make([]Value, needed-len(vm.regs))...)
				}
				vm.regs[reg] = v
			}
		}
	}

	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case returnResult:
				err = runtimeError(e.pos, "can't return at top level")
			case Error:
				err = e
			case error:
				err = runtimeError(Position{Line: 1, Column: 1}, "%v", e)
			default:
				err = runtimeError(Position{Line: 1, Column: 1}, "%v", r)
			}
		}
	}()

	vm.Execute(cp.Fn)

	// Auto-call main() if defined and not in unit-test mode.
	if config == nil || !config.IsUnitTest {
		if reg, ok := cp.varRegs["main"]; ok {
			if int(reg) < len(vm.regs) {
				if mainFn, ok := vm.regs[reg].(*vmFunction); ok {
					mainFn.call(interp, Position{}, nil)
				}
			}
		}
	}

	return &interp.stats, nil
}
