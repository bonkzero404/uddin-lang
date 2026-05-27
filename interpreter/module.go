package interpreter

import "io"

// ModuleFunc is the function signature for all stdlib module functions.
type ModuleFunc func(ctx ModuleContext, pos Position, args []Value) Value

// ModuleContext is the public API stdlib modules use to interact with the interpreter.
type ModuleContext interface {
	LookupVar(name string) (Value, bool)
	SetVar(name string, val Value)
	Stdout() io.Writer
	// GetContext retrieves a host-injected map stored as "_<key>_ctx".
	GetContext(key string) (map[string]any, bool)
	// ReturnValue panics with returnResult — used by waf.return() to exit script.
	ReturnValue(val Value, pos Position)
	CDCStore() CDCAccessor
	FactDB() FactAccessor
}

// CDCAccessor wraps the interpreter's eventStore + eventPatterns + eventMutex.
type CDCAccessor interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	Events() []map[string]any
	SetEvents(store []map[string]any)
	AppendEvent(event map[string]any)
	Patterns() map[string]any
	SetPattern(name string, pattern map[string]any)
}

// FactAccessor wraps the interpreter's factDatabase + factMutex.
type FactAccessor interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	Get(key string) (any, bool)
	Set(key string, val any)
	Delete(key string)
	All() map[string]any
}

// Module is implemented by each stdlib package.
type Module interface {
	Name() string
	Functions() map[string]ModuleFunc
}

// TypeErrorf creates a type error suitable for stdlib module functions.
func TypeErrorf(pos Position, format string, args ...any) error {
	return typeError(pos, format, args...)
}

// RuntimeErrorf creates a runtime error suitable for stdlib module functions.
func RuntimeErrorf(pos Position, format string, args ...any) error {
	return runtimeError(pos, format, args...)
}
