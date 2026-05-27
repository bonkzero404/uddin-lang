package interpreter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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
	registerDirectBuiltins()
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

func registerDirectBuiltins() {
	registerVMBuiltinDirect("len", directLen)
	registerVMBuiltinDirect("typeof", directTypeof)
	registerVMBuiltinDirect("upper", directUpper)
	registerVMBuiltinDirect("lower", directLower)
	registerVMBuiltinDirect("trim", directTrim)
	registerVMBuiltinDirect("contains", directContains)
	registerVMBuiltinDirect("starts_with", directStartsWith)
	registerVMBuiltinDirect("ends_with", directEndsWith)
	registerVMBuiltinDirect("split", directSplit)
	registerVMBuiltinDirect("join", directJoin)
	registerVMBuiltinDirect("int", directInt)
	registerVMBuiltinDirect("str", directStr)
}

func directLen(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "len() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case string:
		return Value(len(v))
	case *[]Value:
		return Value(len(*v))
	case []Value:
		return Value(len(v))
	case map[string]Value:
		return Value(len(v))
	default:
		panic(typeError(pos, "len() requires a string, array, or object"))
	}
}

func directTypeof(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "typeof() requires 1 argument, got %d", len(args)))
	}
	return Value(typeName(args[0]))
}

func directUpper(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "upper() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "upper() requires a string"))
	}
	return Value(strings.ToUpper(s))
}

func directLower(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "lower() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "lower() requires a string"))
	}
	return Value(strings.ToLower(s))
}

func directTrim(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "trim() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "trim() requires a string"))
	}
	return Value(strings.TrimSpace(s))
}

func directContains(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "contains() requires 2 arguments, got %d", len(args)))
	}
	switch haystack := args[0].(type) {
	case string:
		needle, ok := args[1].(string)
		if !ok {
			panic(typeError(pos, "contains() on string requires second argument to be a string"))
		}
		return Value(strings.Contains(haystack, needle))
	case *[]Value:
		needle := args[1]
		for _, v := range *haystack {
			if evalEqual(pos, needle, v).(bool) {
				return Value(true)
			}
		}
		return Value(false)
	default:
		panic(typeError(pos, "contains() requires first argument to be a string or array"))
	}
}

func directStartsWith(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "starts_with() requires 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires first argument to be a string"))
	}
	prefix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires second argument to be a string"))
	}
	return Value(strings.HasPrefix(s, prefix))
}

func directEndsWith(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "ends_with() requires 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires first argument to be a string"))
	}
	suffix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires second argument to be a string"))
	}
	return Value(strings.HasSuffix(s, suffix))
}

func directSplit(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(typeError(pos, "split() requires 1 or 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "split() requires first argument to be a string"))
	}
	var parts []string
	if len(args) == 1 || args[1] == nil {
		parts = strings.Fields(s)
	} else {
		sep, ok := args[1].(string)
		if !ok {
			panic(typeError(pos, "split() requires separator to be a string"))
		}
		parts = strings.Split(s, sep)
	}
	result := make([]Value, len(parts))
	for i, p := range parts {
		result[i] = Value(p)
	}
	return Value(&result)
}

func directJoin(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "join() requires 2 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "join() requires first argument to be an array"))
	}
	sep, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "join() requires second argument to be a string"))
	}
	parts := make([]string, len(*arr))
	for i, v := range *arr {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return Value(strings.Join(parts, sep))
}

func directInt(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "int() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case int:
		return args[0]
	case int64:
		return Value(int(v))
	case float64:
		return Value(int(v))
	case bool:
		if v {
			return Value(1)
		}
		return Value(0)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return Value(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return Value(int(f))
		}
		return Value(nil)
	default:
		panic(typeError(pos, "int() requires an int, float, bool, or string"))
	}
}

func directStr(_ *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "str() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case string:
		return args[0]
	case int:
		return Value(fmt.Sprintf("%d", v))
	case int64:
		return Value(fmt.Sprintf("%d", v))
	case float64:
		return Value(fmt.Sprintf("%g", v))
	case bool:
		if v {
			return Value("true")
		}
		return Value("false")
	case nil:
		return Value("null")
	default:
		return Value(fmt.Sprintf("%v", v))
	}
}
