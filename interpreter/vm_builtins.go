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

// vmBuiltinDirectTable is a parallel table to vmBuiltinTable.
// A non-nil entry means the builtin has a direct implementation that skips
// the DispatchOrPanic chain. Index is stable and matches vmBuiltinTable.
var vmBuiltinDirectTable []func(*interpreter, Position, []Value) Value

func init() {
	registerVMBuiltins()
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
	vmBuiltinDirectTable = make([]func(*interpreter, Position, []Value) Value, len(names))

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

// registerVMBuiltinDirect sets a direct (no-dispatch) implementation for name.
// Must be called after registerVMBuiltins() — from init() in vm_builtins.go.
func registerVMBuiltinDirect(name string, fn func(*interpreter, Position, []Value) Value) {
	idx, ok := vmBuiltinIndex[name]
	if !ok {
		panic("registerVMBuiltinDirect: unknown builtin " + name)
	}
	vmBuiltinDirectTable[idx] = fn
}
