package interpreter

// vmBuiltinEntry maps a builtin name to its VM-callable function.
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
	registerVMBuiltins()
}

func registerVMBuiltin(name string, fn func(*interpreter, Position, []Value) Value) {
	idx := uint16(len(vmBuiltinTable))
	vmBuiltinTable = append(vmBuiltinTable, vmBuiltinEntry{name: name, fn: fn})
	vmBuiltinIndex[name] = idx
}

func registerVMBuiltins() {
	// Delegate all builtins through globalBuiltinDispatcher.
	// Add the most-used builtins first (order = stable index).
	for _, name := range []string{
		"len", "print", "str", "int", "float", "type",
		"append", "range", "keys", "values", "has", "delete",
		"push", "pop", "first", "last", "slice",
	} {
		n := name // capture loop var
		registerVMBuiltin(n, func(interp *interpreter, pos Position, args []Value) Value {
			return globalBuiltinDispatcher.DispatchOrPanic(n, interp, pos, args)
		})
	}
}
