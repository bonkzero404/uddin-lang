package interpreter

import "sort"

// vmBuiltinEntry maps a builtin name to its VM-callable function.
type vmBuiltinEntry struct {
	name string
	fn   func(*interpreter, Position, []Value) Value
}

// vmBuiltinTable is the ordered list of builtins available via OP_CALL_BUILTIN.
// Index must be stable within a process (sorted alphabetically).
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
	// Collect all names from the package-level builtins map (functions.go).
	// builtins is a package-level var, initialized before init() runs.
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)

	vmBuiltinTable = make([]vmBuiltinEntry, 0, len(names))
	vmBuiltinIndex = make(map[string]uint16, len(names))

	for _, name := range names {
		n := name // capture loop var
		vmBuiltinTable = append(vmBuiltinTable, vmBuiltinEntry{
			name: n,
			fn: func(interp *interpreter, pos Position, args []Value) Value {
				return globalBuiltinDispatcher.DispatchOrPanic(n, interp, pos, args)
			},
		})
		vmBuiltinIndex[n] = uint16(len(vmBuiltinTable) - 1)
	}
}
