package interpreter

import (
	"io"
	"sort"
)

var globalModuleRegistry = map[string]Module{}

// RegisterModule adds m to the global registry. Called from stdlib init() functions.
func RegisterModule(m Module) {
	globalModuleRegistry[m.Name()] = m
}

// LookupModule returns the module for name, or (nil, false).
func LookupModule(name string) (Module, bool) {
	m, ok := globalModuleRegistry[name]
	return m, ok
}

// knownModuleNames returns sorted module names for error messages.
func knownModuleNames() []string {
	names := make([]string, 0, len(globalModuleRegistry))
	for name := range globalModuleRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// buildNamespaceObject converts a Module's Functions() into a map[string]Value
// where each value is a builtinFunction.
func buildNamespaceObject(mod Module, interp *interpreter) map[string]Value {
	ns := make(map[string]Value, len(mod.Functions()))
	for fname, fn := range mod.Functions() {
		fn, modName, fname := fn, mod.Name(), fname
		ns[fname] = builtinFunction{
			Function: func(interp *interpreter, pos Position, args []Value) Value {
				ctx := &interpreterContext{interp: interp}
				return fn(ctx, pos, args)
			},
			Name: modName + "." + fname,
		}
	}
	return ns
}

// interpreterContext wraps *interpreter to implement ModuleContext.
type interpreterContext struct{ interp *interpreter }

func (c *interpreterContext) LookupVar(name string) (Value, bool) {
	return c.interp.lookup(name)
}
func (c *interpreterContext) SetVar(name string, val Value) {
	c.interp.assign(name, val)
}
func (c *interpreterContext) Stdout() io.Writer {
	return c.interp.stdout
}
func (c *interpreterContext) GetContext(key string) (map[string]any, bool) {
	val, ok := c.interp.lookup("_" + key + "_ctx")
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]any)
	return m, ok
}
func (c *interpreterContext) ReturnValue(val Value, pos Position) {
	panic(returnResult{value: val, pos: pos})
}
func (c *interpreterContext) CDCStore() CDCAccessor {
	return &cdcAccessor{interp: c.interp}
}
func (c *interpreterContext) FactDB() FactAccessor {
	return &factAccessor{interp: c.interp}
}

// cdcAccessor implements CDCAccessor.
type cdcAccessor struct{ interp *interpreter }

func (a *cdcAccessor) Lock()    { a.interp.eventMutex.Lock() }
func (a *cdcAccessor) Unlock()  { a.interp.eventMutex.Unlock() }
func (a *cdcAccessor) RLock()   { a.interp.eventMutex.RLock() }
func (a *cdcAccessor) RUnlock() { a.interp.eventMutex.RUnlock() }
func (a *cdcAccessor) Events() []map[string]any { return a.interp.eventStore }
func (a *cdcAccessor) SetEvents(store []map[string]any) { a.interp.eventStore = store }
func (a *cdcAccessor) AppendEvent(event map[string]any) {
	a.interp.eventStore = append(a.interp.eventStore, event)
}
func (a *cdcAccessor) Patterns() map[string]any { return a.interp.eventPatterns }
func (a *cdcAccessor) SetPattern(name string, pattern map[string]any) {
	a.interp.eventPatterns[name] = pattern
}

// factAccessor implements FactAccessor.
type factAccessor struct{ interp *interpreter }

func (a *factAccessor) Lock()    { a.interp.factMutex.Lock() }
func (a *factAccessor) Unlock()  { a.interp.factMutex.Unlock() }
func (a *factAccessor) RLock()   { a.interp.factMutex.RLock() }
func (a *factAccessor) RUnlock() { a.interp.factMutex.RUnlock() }
func (a *factAccessor) Get(key string) (any, bool) {
	v, ok := a.interp.factDatabase[key]
	return v, ok
}
func (a *factAccessor) Set(key string, val any)  { a.interp.factDatabase[key] = val }
func (a *factAccessor) Delete(key string)        { delete(a.interp.factDatabase, key) }
func (a *factAccessor) All() map[string]any      { return a.interp.factDatabase }
