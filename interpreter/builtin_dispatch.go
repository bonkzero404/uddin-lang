package interpreter

import (
	"sync"
)

// BuiltinFunctionDispatcher provides optimized dispatch for builtin functions
type BuiltinFunctionDispatcher struct {
	dispatchTable map[string]int
	functions     []BuiltinFunctionEntry
	mutex         sync.RWMutex
	callCounts    map[string]int64
}

// BuiltinFunctionEntry represents a builtin function entry
type BuiltinFunctionEntry struct {
	Name     string
	Function func(*interpreter, Position, []Value) Value
	ArgCount int // -1 for variadic
	FastPath bool
}

// NewBuiltinFunctionDispatcher creates a new builtin function dispatcher
func NewBuiltinFunctionDispatcher() *BuiltinFunctionDispatcher {
	return &BuiltinFunctionDispatcher{
		dispatchTable: make(map[string]int),
		functions:     make([]BuiltinFunctionEntry, 0, 200),
		callCounts:    make(map[string]int64),
	}
}

// Global builtin dispatcher
var globalBuiltinDispatcher = NewBuiltinFunctionDispatcher()

// RegisterBuiltinFunction registers a builtin function with the dispatcher
func (bfd *BuiltinFunctionDispatcher) RegisterBuiltinFunction(name string, fn func(*interpreter, Position, []Value) Value, argCount int, fastPath bool) {
	bfd.mutex.Lock()
	defer bfd.mutex.Unlock()
	
	index := len(bfd.functions)
	bfd.dispatchTable[name] = index
	bfd.functions = append(bfd.functions, BuiltinFunctionEntry{
		Name:     name,
		Function: fn,
		ArgCount: argCount,
		FastPath: fastPath,
	})
}

// DispatchBuiltinFunction dispatches a builtin function call
func (bfd *BuiltinFunctionDispatcher) DispatchBuiltinFunction(name string, interp *interpreter, pos Position, args []Value) (Value, bool) {
	bfd.mutex.RLock()
	index, exists := bfd.dispatchTable[name]
	if !exists {
		bfd.mutex.RUnlock()
		return nil, false
	}
	
	entry := bfd.functions[index]
	bfd.mutex.RUnlock()
	
	// Update call count with write lock
	bfd.mutex.Lock()
	bfd.callCounts[name]++
	bfd.mutex.Unlock()
	
	// Validate argument count if specified
	if entry.ArgCount >= 0 && len(args) != entry.ArgCount {
		plural := ""
		if entry.ArgCount != 1 {
			plural = "s"
		}
		panic(typeError(pos, "%s() requires %d arg%s, got %d", name, entry.ArgCount, plural, len(args)))
	}
	
	// Call the function
	result := entry.Function(interp, pos, args)
	return result, true
}

// GetCallCount returns the call count for a specific function
func (bfd *BuiltinFunctionDispatcher) GetCallCount(name string) int64 {
	bfd.mutex.RLock()
	defer bfd.mutex.RUnlock()
	return bfd.callCounts[name]
}

// GetMostCalledFunctions returns the most frequently called functions
func (bfd *BuiltinFunctionDispatcher) GetMostCalledFunctions(limit int) []string {
	bfd.mutex.RLock()
	defer bfd.mutex.RUnlock()
	
	type funcCount struct {
		name  string
		count int64
	}
	
	funcs := make([]funcCount, 0, len(bfd.callCounts))
	for name, count := range bfd.callCounts {
		funcs = append(funcs, funcCount{name, count})
	}
	
	// Simple bubble sort for small datasets
	for i := 0; i < len(funcs)-1; i++ {
		for j := 0; j < len(funcs)-i-1; j++ {
			if funcs[j].count < funcs[j+1].count {
				funcs[j], funcs[j+1] = funcs[j+1], funcs[j]
			}
		}
	}
	
	result := make([]string, 0, limit)
	for i := 0; i < len(funcs) && i < limit; i++ {
		result = append(result, funcs[i].name)
	}
	
	return result
}

// GetGlobalBuiltinDispatcher returns the global builtin dispatcher
func GetGlobalBuiltinDispatcher() *BuiltinFunctionDispatcher {
	return globalBuiltinDispatcher
}

// ArgumentValidationCache caches argument validation results
type ArgumentValidationCache struct {
	cache   map[string]bool
	mutex   sync.RWMutex
	maxSize int
}

// NewArgumentValidationCache creates a new argument validation cache
func NewArgumentValidationCache(maxSize int) *ArgumentValidationCache {
	return &ArgumentValidationCache{
		cache:   make(map[string]bool),
		maxSize: maxSize,
	}
}

// Global argument validation cache
var globalArgValidationCache = NewArgumentValidationCache(1000)

// ValidateArguments checks if arguments are valid for a function
func (avc *ArgumentValidationCache) ValidateArguments(funcName string, argTypes []string) bool {
	cacheKey := funcName + ":"
	for _, argType := range argTypes {
		cacheKey += argType + ","
	}
	
	avc.mutex.RLock()
	result, exists := avc.cache[cacheKey]
	avc.mutex.RUnlock()
	
	if exists {
		return result
	}
	
	// Perform validation (simplified)
	valid := true // In real implementation, this would contain validation logic
	
	avc.mutex.Lock()
	defer avc.mutex.Unlock()
	
	// Check cache size limit
	if len(avc.cache) >= avc.maxSize {
		// Clear half the cache
		count := 0
		for k := range avc.cache {
			delete(avc.cache, k)
			count++
			if count >= avc.maxSize/2 {
				break
			}
		}
	}
	
	avc.cache[cacheKey] = valid
	return valid
}

// GetGlobalArgumentValidationCache returns the global argument validation cache
func GetGlobalArgumentValidationCache() *ArgumentValidationCache {
	return globalArgValidationCache
}

// SpecializedBuiltinFunctions provides fast implementations for commonly used functions
type SpecializedBuiltinFunctions struct{}

// FastLen provides optimized length calculation
func (sbf *SpecializedBuiltinFunctions) FastLen(value Value) (int, bool) {
	switch v := value.(type) {
	case string:
		return len(v), true
	case *[]Value:
		return len(*v), true
	case map[string]Value:
		return len(v), true
	default:
		return 0, false
	}
}

// FastTypeof provides optimized type checking
func (sbf *SpecializedBuiltinFunctions) FastTypeof(value Value) string {
	switch value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case int:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case *[]Value:
		return "array"
	case map[string]Value:
		return "object"
	case functionType:
		return "function"
	default:
		return "unknown"
	}
}

// FastStr provides optimized string conversion
func (sbf *SpecializedBuiltinFunctions) FastStr(value Value) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return toString(v, false)
	case float64:
		return toString(v, false)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		return toString(v, false)
	}
}

// Global specialized builtin functions
var globalSpecializedBuiltins = &SpecializedBuiltinFunctions{}

// GetGlobalSpecializedBuiltins returns the global specialized builtin functions
func GetGlobalSpecializedBuiltins() *SpecializedBuiltinFunctions {
	return globalSpecializedBuiltins
}

// InitializeBuiltinDispatcher initializes the builtin function dispatcher with all builtin functions
func InitializeBuiltinDispatcher() {
	dispatcher := GetGlobalBuiltinDispatcher()
	
	// Core System Functions (commonly used, fast path enabled)
	dispatcher.RegisterBuiltinFunction("len", lenFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("typeof", typeofFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("str", strFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("int", intFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("float", floatFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("print", printFunc, -1, false)
	
	// Type Conversion Functions
	dispatcher.RegisterBuiltinFunction("char", charFunc, 1, false)
	dispatcher.RegisterBuiltinFunction("rune", runeFunc, 1, false)
	
	// Basic Array/Collection Functions
	dispatcher.RegisterBuiltinFunction("append", appendFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("range", rangeFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("slice", sliceFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("sort", sortFunc, 1, false)
	
	// String Manipulation Functions - Basic (commonly used)
	dispatcher.RegisterBuiltinFunction("join", joinFunc, 2, true)
	dispatcher.RegisterBuiltinFunction("split", splitFunc, 2, true)
	dispatcher.RegisterBuiltinFunction("lower", lowerFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("upper", upperFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("contains", containsFunc, 2, true)
	dispatcher.RegisterBuiltinFunction("substr", substrFunc, 3, false)
	
	// Array/Collection Methods - Advanced
	dispatcher.RegisterBuiltinFunction("push", pushFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("pop", popFunc, 1, false)
	dispatcher.RegisterBuiltinFunction("shift", shiftFunc, 1, false)
	dispatcher.RegisterBuiltinFunction("unshift", unshiftFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("reverse", reverseFunc, 1, false)
	
	// Mathematical Functions - Basic (commonly used)
	dispatcher.RegisterBuiltinFunction("abs", absFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("max", maxFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("min", minFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("pow", powFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("sqrt", sqrtFunc, 1, false)
	
	// File I/O Functions
	dispatcher.RegisterBuiltinFunction("read_file", readFileFunc, 1, false)
	dispatcher.RegisterBuiltinFunction("write_file", writeFileFunc, 2, false)
	
	// Functional Programming Functions
	dispatcher.RegisterBuiltinFunction("map", mapFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("filter", filterFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("reduce", reduceFunc, 3, false)
	
	// System Functions
	dispatcher.RegisterBuiltinFunction("import", importFunc, 1, false)
	dispatcher.RegisterBuiltinFunction("exit", exitFunc, -1, false)
	dispatcher.RegisterBuiltinFunction("input", inputFunc, -1, false)
	
	// Concurrent execution functions
	dispatcher.RegisterBuiltinFunction("concurrent_map", concurrentMapFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("concurrent_filter", concurrentFilterFunc, 2, false)
	dispatcher.RegisterBuiltinFunction("concurrent_reduce", concurrentReduceFunc, 3, false)
	dispatcher.RegisterBuiltinFunction("parallel_execute", parallelExecuteFunc, -1, false)
	
	// Note: Additional functions can be registered as needed
	// This covers the most commonly used builtin functions with appropriate fast path settings
}