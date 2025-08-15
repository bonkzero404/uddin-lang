package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"strings"
	"sync"
)

// Value represents any runtime value in the language.
// This includes nil, boolean, integer, float, string, array, object, and function values.
type Value any

// interpreter represents the internal state of the interpreter.
// It maintains variable scopes, I/O streams, and execution statistics.
type interpreter struct {
	// vars is a stack of variable scopes, with the most local scope at the end
	vars []map[string]Value
	// mutex protects concurrent access to vars from HTTP handlers
	mutex sync.RWMutex
	// args holds command-line arguments for the args() builtin
	args []string
	// stdin is the input stream for the read() builtin
	stdin io.Reader
	// stdinReader is a persistent bufio.Reader for stdin
	stdinReader *bufio.Reader
	// stdout is the output stream for the print() builtin
	stdout io.Writer
	// exit is the function called by the exit() builtin
	exit func(int)
	// stats tracks execution statistics
	stats      Stats
	inUnitTest bool
	// currentPos tracks the current position being executed for better error reporting
	currentPos Position
	// Optimization caches
	memoCache map[string]Value
	// String builder pool for efficient memory allocation
	stringBuilderPool sync.Pool
	// Array pool for efficient memory allocation
	arrayPool sync.Pool
	// Map pool for efficient memory allocation
	mapPool sync.Pool
	// Variable lookup cache for performance optimization
	variableLookupCache *VariableLookupCache
	// Expression optimizer
	optimizer *ConstantFolder
}

// returnResult is used to handle return statements in functions.
// When a return statement is executed, it's wrapped in this struct and thrown as a panic,
// which is then caught by the function's call handler.
type returnResult struct {
	// value is the value being returned
	value Value
	// pos is the position of the return statement in the source code
	pos Position
}

// binaryEvalFunc is a function type for evaluating binary operations.
// It takes the position of the operation and two operand values, and returns the result.
type binaryEvalFunc func(pos Position, l, r Value) Value

// binaryEvalFuncs maps binary operator tokens to their evaluation functions.
// This allows for a clean dispatch of binary operations during expression evaluation.
var binaryEvalFuncs = map[Token]binaryEvalFunc{
	DIVIDE:   evalDivide,                                                                   // Division operator: /
	EQUAL:    evalEqual,                                                                    // Equality operator: ==
	GT:       func(pos Position, l, r Value) Value { return evalLess(pos, r, l) },          // Greater than: >
	GTE:      func(pos Position, l, r Value) Value { return !evalLess(pos, l, r).(bool) },  // Greater than or equal: >=
	IN:       evalIn,                                                                       // Containment operator: in
	LT:       evalLess,                                                                     // Less than operator: <
	LTE:      func(pos Position, l, r Value) Value { return !evalLess(pos, r, l).(bool) },  // Less than or equal: <=
	MINUS:    evalMinus,                                                                    // Subtraction operator: -
	MODULO:   evalModulo,                                                                   // Modulo operator: %
	NOTEQUAL: func(pos Position, l, r Value) Value { return !evalEqual(pos, l, r).(bool) }, // Inequality operator: !=
	POWER:    evalPower,                                                                    // Power operator: **
	TIMES:    evalTimes,                                                                    // Multiplication operator: *
}

// ensureIntToFloats converts integer or float operands to float64 for arithmetic operations.
// If either operand is not a numeric type (int or float64), it returns error values.
//
// Parameters:
//   - pos: Position in source code for error reporting
//   - l: Left operand value
//   - r: Right operand value
//   - operation: String description of the operation for error messages
//
// Returns:
//   - Two float64 values representing the converted operands, or (0, 0) on error
func ensureIntToFloats(l, r Value) (float64, float64) {
	// Try to convert left operand to float64
	lf, lok := l.(float64)
	if !lok {
		if li, lok := l.(int); lok {
			lf = float64(li) // Convert int to float64
		} else {
			return 0, 0 // Return error values that can be detected by caller
		}
	}

	// Try to convert right operand to float64
	rf, rok := r.(float64)
	if !rok {
		if ri, rok := r.(int); rok {
			rf = float64(ri) // Convert int to float64
		} else {
			return 0, 0 // Return error values
		}
	}

	return lf, rf
}

// evalEqual evaluates equality between two values of any type.
// It implements deep equality for composite types like arrays and maps.
//
// Parameters:
//   - pos: Position in source code for error reporting
//   - l: Left operand value
//   - r: Right operand value
//
// Returns:
//   - A boolean Value indicating whether the values are equal
func evalEqual(pos Position, l, r Value) Value {
	// Try fast evaluation first for simple types
	if result, handled := GetFastEvaluator().FastEvalEqual(l, r); handled {
		return result
	}

	// Fallback to original deep equality logic
	switch l := l.(type) {
	case nil:
		// nil is only equal to nil
		return Value(r == nil)

	case bool:
		// Boolean equality
		if r, ok := r.(bool); ok {
			return Value(l == r)
		}

	case int:
		// Integer equality
		if r, ok := r.(int); ok {
			return Value(l == r)
		}
		// Mixed int/float equality
		if r, ok := r.(float64); ok {
			return Value(float64(l) == r)
		}

	case float64:
		// Float equality
		if r, ok := r.(float64); ok {
			return Value(l == r)
		}
		// Mixed float/int equality
		if r, ok := r.(int); ok {
			return Value(l == float64(r))
		}

	case string:
		// String equality
		if r, ok := r.(string); ok {
			return Value(l == r)
		}

	case *[]Value:
		// Array equality (deep comparison)
		if r, ok := r.(*[]Value); ok {
			// Arrays must have the same length
			if len(*l) != len(*r) {
				return Value(false)
			}
			// Compare each element recursively
			for i, elem := range *l {
				if !evalEqual(pos, elem, (*r)[i]).(bool) {
					return Value(false)
				}
			}
			return Value(true)
		}

	case map[string]Value:
		// Object/map equality (deep comparison)
		if r, ok := r.(map[string]Value); ok {
			// Maps must have the same size
			if len(l) != len(r) {
				return Value(false)
			}
			// Compare each key-value pair recursively
			for k, v := range l {
				if !evalEqual(pos, v, r[k]).(bool) {
					return Value(false)
				}
			}
			return Value(true)
		}

	case functionType:
		// Function equality (identity comparison)
		if r, ok := r.(functionType); ok {
			return Value(l == r)
		}
	}

	// Values of different types are never equal
	return Value(false)
}

// evalIn evaluates the 'in' containment operator.
// It checks if the left value is contained in the right value, which can be:
// - a string (substring check)
// - an array (element equality check)
// - a map/object (key existence check)
//
// Parameters:
//   - pos: Position in source code for error reporting
//   - l: Left operand value (what to look for)
//   - r: Right operand value (where to look)
//
// Returns:
//   - A boolean Value indicating whether the left value is contained in the right value
func evalIn(pos Position, l, r Value) Value {
	switch r := r.(type) {
	case string:
		// String containment: check if l is a substring of r
		if l, ok := l.(string); ok {
			return Value(strings.Contains(r, l))
		}
		panic(typeError(pos, "in string requires string on left side"))

	case *[]Value:
		// Array containment: check if l equals any element in r
		for _, v := range *r {
			if evalEqual(pos, l, v).(bool) {
				return Value(true)
			}
		}
		return Value(false)

	case map[string]Value:
		// Object/map containment: check if l is a key in r
		if l, ok := l.(string); ok {
			_, present := r[l]
			return Value(present)
		}
		panic(typeError(pos, "in object requires string on left side"))
	}

	// The 'in' operator only works with strings, arrays, or objects on the right side
	panic(typeError(pos, "in requires string, array, or object on right side"))
}

// evalLess evaluates the less-than comparison operator (<).
// It compares two values of the same type, which must be comparable types:
// - integers
// - floats
// - strings (lexicographical comparison)
// - arrays (lexicographical comparison with length as tiebreaker)
//
// Parameters:
//   - pos: Position in source code for error reporting
//   - l: Left operand value
//   - r: Right operand value
//
// Returns:
//   - A boolean Value indicating whether l < r
func evalLess(pos Position, l, r Value) Value {
	// Try fast evaluation first for simple types
	if result, handled := GetFastEvaluator().FastEvalLess(l, r); handled {
		return result
	}

	// Fallback to original logic for complex types
	switch l := l.(type) {
	case *[]Value:
		// Array lexicographical comparison
		if r, ok := r.(*[]Value); ok {
			// Compare elements pairwise until a difference is found
			for i := 0; i < len(*l) && i < len(*r); i++ {
				if !evalEqual(pos, (*l)[i], (*r)[i]).(bool) {
					return evalLess(pos, (*l)[i], (*r)[i])
				}
			}
			// If all common elements are equal, shorter array is less
			return Value(len(*l) < len(*r))
		}
	}

	// Only certain types can be compared with <
	panic(typeError(pos, "comparison requires two integers or two strings (or arrays of integers or strings)"))
}

// Optimized evalPlus with reduced type assertions and memory optimization
func (interp *interpreter) evalPlus(pos Position, l, r Value) Value {
	// Track operation for performance monitoring
	TrackOperation("plus")

	// Try fast numeric evaluation first to avoid any boxing
	if result, handled := GetFastEvaluator().FastEvalPlus(l, r); handled {
		return result
	}

	// Handle non-numeric cases
	if ls, ok := l.(string); ok {
		if rs, ok := r.(string); ok {
			// Use string interning for repeated strings
			result := ls + rs
			if len(result) < 100 { // Only intern short strings
				result = InternString(result)
			}
			return Value(result)
		}
	} else if larr, ok := l.(*[]Value); ok {
		if rarr, ok := r.(*[]Value); ok {
			// Optimized array concatenation using advanced batch processing for large arrays
			llen, rlen := len(*larr), len(*rarr)
			if llen+rlen > 1000 {
				// Use batch processing for large arrays
				result := globalComplexPool.GetArray()
				result = SmartAppend(result, *larr...)
				result = SmartAppend(result, *rarr...)
				final := make([]Value, len(result))
				copy(final, result)
				globalComplexPool.PutArray(&result)
				return Value(&final)
			} else {
				// Use existing optimized concatenation for smaller arrays
				concat := NewArrayConcatenator(llen + rlen)
				concat.AppendArray(larr)
				concat.AppendArray(rarr)
				return Value(concat.Result())
			}
		}
	} else if lmap, ok := l.(map[string]Value); ok {
		if rmap, ok := r.(map[string]Value); ok {
			// Use optimized map merging
			result := OptimizedMapMerge(lmap, rmap)
			return Value(result)
		}
	}
	panic(typeError(pos, "+ requires two integers, strings, arrays, or objects"))
}

func evalMinus(pos Position, l, r Value) Value {
	// Try fast numeric evaluation first
	if result, handled := GetFastEvaluator().FastEvalMinus(l, r); handled {
		return result
	}

	panic(typeError(pos, "- requires two floats or integers, got %T and %T", l, r))
}

// Optimized evalTimes with reduced type assertions
func evalTimes(pos Position, l, r Value) Value {
	// Track operation for performance monitoring
	TrackOperation("times")

	// Try fast numeric evaluation first for pure numeric operations
	if result, handled := GetFastEvaluator().FastEvalTimes(l, r); handled {
		return result
	}

	// Handle mixed type operations (int * string, int * array, etc.)
	if li, ok := l.(int); ok {
		if rs, ok := r.(string); ok {
			if li < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			// Use optimized string repetition for large strings
			if len(rs)*li > 1000 {
				sb := globalComplexPool.GetStringBuilder()
				for i := 0; i < li; i++ {
					sb.WriteString(rs)
				}
				result := sb.String()
				globalComplexPool.PutStringBuilder(sb)
				return Value(result)
			} else {
				return Value(strings.Repeat(rs, li))
			}
		}
		if rarr, ok := r.(*[]Value); ok {
			if li < 0 {
				panic(valueError(pos, "can't multiply array by a negative number"))
			}
			// Optimized array repetition with batch processing for large arrays
			totalSize := len(*rarr) * li
			if totalSize > 1000 {
				// Use batch processing for large arrays
				result := globalComplexPool.GetArray()
				for i := 0; i < li; i++ {
					result = SmartAppend(result, *rarr...)
				}
				final := make([]Value, len(result))
				copy(final, result)
				globalComplexPool.PutArray(&result)
				return Value(&final)
			} else {
				// Use existing optimized concatenation for smaller arrays
				concat := NewArrayConcatenator(totalSize)
				for i := 0; i < li; i++ {
					concat.AppendArray(rarr)
				}
				return Value(concat.Result())
			}
		}
	} else if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf * rf)
		}
		if ri, ok := r.(int); ok {
			return Value(lf * float64(ri))
		}
	} else if ls, ok := l.(string); ok {
		if ri, ok := r.(int); ok {
			if ri < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			return Value(strings.Repeat(ls, ri))
		}
	} else if larr, ok := l.(*[]Value); ok {
		if ri, ok := r.(int); ok {
			if ri < 0 {
				panic(valueError(pos, "can't multiply array by a negative number"))
			}
			// Optimized array repetition using ArrayConcatenator
			concat := NewArrayConcatenator(len(*larr) * ri)
			for i := 0; i < ri; i++ {
				concat.AppendArray(larr)
			}
			return Value(concat.Result())
		}
	}
	panic(typeError(pos, "* requires two integers or floats, or a string or array and an integer"))
}

func evalDivide(pos Position, l, r Value) Value {
	// Try fast numeric evaluation first
	if result, handled := GetFastEvaluator().FastEvalDivide(l, r); handled {
		return result
	}

	// Fallback to original logic
	li, ri := ensureIntToFloats(l, r)
	if li == 0 && ri == 0 {
		panic(typeError(pos, "/ requires two floats or integers"))
	}
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(li / ri)
}

func evalModulo(pos Position, l, r Value) Value {
	// Try fast numeric evaluation first
	if result, handled := GetFastEvaluator().FastEvalModulo(l, r); handled {
		return result
	}

	// Fallback to original logic
	li, ri := ensureIntToFloats(l, r)
	if li == 0 && ri == 0 {
		panic(typeError(pos, "modulo operator requires two floats or integers"))
	}
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(int(li) % int(ri))
}

// evalPower evaluates the power operation (exponentiation).
// It converts operands to floats and uses math.Pow for calculation.
func evalPower(pos Position, l, r Value) Value {
	// Try fast numeric evaluation first for simple cases
	if result, handled := GetFastEvaluator().FastEvalPower(l, r); handled {
		return result
	}

	// Fallback to original logic with math.Pow
	li, ri := ensureIntToFloats(l, r)
	if li == 0 && ri == 0 {
		panic(typeError(pos, "** requires two floats or integers"))
	}
	result := math.Pow(li, ri)
	return Value(result)
}

// Unary operator evaluation functions
type unaryEvalFunc func(pos Position, v Value) Value

// Map of unary operator evaluation functions
var unaryEvalFuncs = map[Token]unaryEvalFunc{
	NOT:   evalNot,
	MINUS: evalNegative,
}

// Unary operator not evaluation function
func evalNot(pos Position, v Value) Value {
	if v, ok := v.(bool); ok {
		return Value(!v)
	}
	panic(typeError(pos, "not requires a bool"))
}

// Unary operator negative evaluation function
func evalNegative(pos Position, v Value) Value {
	if vi, ok := v.(int); ok {
		return Value(-vi)
	} else if vf, ok := v.(float64); ok {
		return Value(-vf)
	}

	panic(typeError(pos, "unary - requires an integer or float"))
}

// Function type for subscript evaluation
func evalSubscript(pos Position, container, subscript Value) Value {
	switch c := container.(type) {
	case string:
		if s, ok := subscript.(int); ok {
			// Handle negative indexing for strings
			if s < 0 {
				s = len(c) + s
			}
			if s < 0 || s >= len(c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			return Value(string([]byte{c[s]}))
		}
		panic(typeError(pos, "string subscript must be an integer"))
	case *[]Value:
		if s, ok := subscript.(int); ok {
			// Handle negative indexing for arrays
			if s < 0 {
				s = len(*c) + s
			}
			if s < 0 || s >= len(*c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			return (*c)[s]
		}
		panic(typeError(pos, "array subscript must be an integer"))
	case map[string]Value:
		if s, ok := subscript.(string); ok {
			if value, ok := c[s]; ok {
				return value
			}
			panic(valueError(pos, "key not found: %q", s))
		}
		panic(typeError(pos, "object subscript must be a string"))
	default:
		panic(typeError(pos, "can only subscript string, array, or object"))
	}
}

func (interp *interpreter) evalAnd(pos Position, le, re Expression) Value {
	l := interp.evaluate(le)
	if l, ok := l.(bool); ok {
		if !l {
			// Short circuit: don't evaluate right if left false
			return Value(false)
		}
		r := interp.evaluate(re)
		if r, ok := r.(bool); ok {
			return Value(r)
		} else {
			panic(typeError(pos, "and requires two bools"))
		}
	} else {
		panic(typeError(pos, "and requires two bools"))
	}
}

func (interp *interpreter) evalOr(pos Position, le, re Expression) Value {
	l := interp.evaluate(le)
	if l, ok := l.(bool); ok {
		if l {
			// Short circuit: don't evaluate right if left true
			return Value(true)
		}
		r := interp.evaluate(re)
		if r, ok := r.(bool); ok {
			return Value(r)
		} else {
			panic(typeError(pos, "or requires two bools"))
		}
	} else {
		panic(typeError(pos, "or requires two bools"))
	}
}

func (interp *interpreter) evalXor(_ Position, le, re Expression) Value {
	l := interp.evaluate(le)
	r := interp.evaluate(re)

	// Convert to boolean values using IsTruthy for consistency with evaluator
	leftTruthy := IsTruthy(l)
	rightTruthy := IsTruthy(r)

	// XOR returns true if exactly one operand is truthy
	return Value(leftTruthy != rightTruthy)
}

func (interp *interpreter) callFunction(pos Position, f functionType, args []Value) (ret Value) {
	// Track function calls for performance monitoring
	TrackOperation("function_call")

	// Check optimized memoization cache for functions marked with memo
	// EXPERIMENTAL: Memoization is experimental and may consume significant memory
	// Only cache functions that are explicitly memoized and return non-nil values
	if uf, ok := f.(*userFunction); ok && uf.Name != "" && uf.Memoized {
		memoKey := OptimizedMemoKey(uf.Name, args)
		if cached, exists := GetGlobalOptimizedMemoCache().Get(memoKey); exists {
			return cached
		}
	}

	// Handle builtin functions with dispatcher
	if bf, ok := f.(builtinFunction); ok {
		// Try builtin dispatcher first for optimized dispatch
		dispatcher := GetGlobalBuiltinDispatcher()
		if result, dispatched := dispatcher.DispatchBuiltinFunction(bf.Name, interp, pos, args); dispatched {
			interp.stats.BuiltinCalls++
			return result
		}
		// Fallback to original method if not in dispatcher
	}

	defer func() {
		if r := recover(); r != nil {
			if result, ok := r.(returnResult); ok {
				ret = result.value
				// Cache the result for memoization only if function is marked as memoized and result is not nil
				// EXPERIMENTAL: Memoization caching is experimental feature
				if uf, ok := f.(*userFunction); ok && uf.Name != "" && uf.Memoized && ret != nil {
					memoKey := OptimizedMemoKey(uf.Name, args)
					GetGlobalOptimizedMemoCache().Set(memoKey, ret)
				}
			} else {
				panic(r)
			}
		}
	}()

	result := f.call(interp, pos, args)

	// Cache the result for memoization only if function is marked as memoized and result is not nil
	// This prevents caching functions with side effects that return nil
	if uf, ok := f.(*userFunction); ok && uf.Name != "" && uf.Memoized && result != nil {
		memoKey := OptimizedMemoKey(uf.Name, args)
		GetGlobalOptimizedMemoCache().Set(memoKey, result)
	}

	return result
}

func (interp *interpreter) evaluate(expr Expression) Value {
	interp.currentPos = expr.Position()
	interp.stats.Ops++

	// Try constant folding optimization first
	if interp.optimizer != nil {
		if result, ok := interp.optimizer.FoldConstants(expr); ok {
			return result
		}
	}

	switch e := expr.(type) {
	case *Binary:
		if e.Operator == PLUS {
			return interp.evalPlus(e.Position(), interp.evaluate(e.Left), interp.evaluate(e.Right))
		} else if f, ok := binaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Left), interp.evaluate(e.Right))
		} else if e.Operator == AND {
			return interp.evalAnd(e.Position(), e.Left, e.Right)
		} else if e.Operator == OR {
			return interp.evalOr(e.Position(), e.Left, e.Right)
		} else if e.Operator == XOR {
			return interp.evalXor(e.Position(), e.Left, e.Right)
		}
		// Parser should never give us this
		panic(runtimeError(e.Position(), "unknown binary operator %v", e.Operator))
	case *Unary:
		if f, ok := unaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Operand))
		}
		// Parser should never give us this
		panic(runtimeError(e.Position(), "unknown unary operator %v", e.Operator))
	case *Ternary:
		condition := interp.evaluate(e.Condition)
		if IsTruthy(condition) {
			return interp.evaluate(e.TrueExpr)
		} else {
			return interp.evaluate(e.FalseExpr)
		}
	case *Call:
		function := interp.evaluate(e.Function)
		if f, ok := function.(functionType); ok {
			args := []Value{}
			for _, a := range e.Arguments {
				args = SmartAppend(args, interp.evaluate(a))
			}
			if e.Ellipsis {
				iterator := getIterator(e.Arguments[len(args)-1].Position(), args[len(args)-1])
				args = args[:len(args)-1]
				for iterator.HasNext() {
					args = SmartAppend(args, iterator.Value())
				}
			}
			return interp.callFunction(e.Function.Position(), f, args)
		}
		panic(typeError(e.Function.Position(), "can't call non-function type %s", typeName(function)))
	case *Literal:
		return Value(e.Value)
	case *Variable:
		if v, ok := interp.lookup(e.Name); ok {
			return v
		}
		panic(nameError(e.Position(), "name %q not found", e.Name))
	case *List:
		// Use array pool for better memory management
		values := interp.getArray()
		if cap(values) < len(e.Values) {
			values = make([]Value, len(e.Values))
		} else {
			values = values[:len(e.Values)]
		}
		for i, v := range e.Values {
			values[i] = interp.evaluate(v)
		}
		return Value(&values)
	case *Map:
		value := interp.getMap()
		for _, item := range e.Items {
			key := interp.evaluate(item.Key)
			if k, ok := key.(string); ok {
				value[k] = interp.evaluate(item.Value)
			} else {
				panic(typeError(item.Key.Position(), "object key must be string, not %s", typeName(key)))
			}
		}
		return Value(value)
	case *Subscript:
		container := interp.evaluate(e.Container)
		subscript := interp.evaluate(e.Subscript)
		return evalSubscript(e.Subscript.Position(), container, subscript)
	case *FunctionExpression:
		closure := interp.vars[len(interp.vars)-1]
		return &userFunction{"", e.Parameters, e.Ellipsis, e.Body, closure, false}
	default:
		// Parser should never get us here
		panic(runtimeError(Position{}, "unexpected expression type %T", expr))
	}
}

func (interp *interpreter) pushScope(scope map[string]Value) {
	interp.mutex.Lock()
	defer interp.mutex.Unlock()
	interp.vars = SmartAppendMapValue(interp.vars, scope)
	// Invalidate cache when scope changes
	if interp.variableLookupCache != nil {
		interp.variableLookupCache.InvalidateScope(len(interp.vars) - 1)
	}
}

func (interp *interpreter) popScope() {
	interp.mutex.Lock()
	defer interp.mutex.Unlock()
	scopeLevel := len(interp.vars) - 1
	// Invalidate cache for the scope being removed
	if interp.variableLookupCache != nil {
		interp.variableLookupCache.InvalidateScope(scopeLevel)
	}
	interp.vars = interp.vars[:len(interp.vars)-1]
}

func (interp *interpreter) assign(name string, value Value) {
	interp.mutex.Lock()
	defer interp.mutex.Unlock()
	interp.vars[len(interp.vars)-1][name] = value
	// Invalidate cache for this variable since it's been reassigned
	if interp.variableLookupCache != nil {
		interp.variableLookupCache.InvalidateVariable(name)
	}
}

func (interp *interpreter) lookup(name string) (Value, bool) {
	interp.mutex.RLock()
	defer interp.mutex.RUnlock()

	// Use variable lookup cache if enabled
	if interp.variableLookupCache != nil {
		// Check cache first
		if cached, exists := interp.variableLookupCache.cache[name]; exists {
			// Verify the cached value is still valid for the current scope
			if cached.ScopeLevel < len(interp.vars) {
				if value, found := interp.vars[cached.ScopeLevel][name]; found && safeValueEqual(value, cached.Value) {
					interp.variableLookupCache.hits++
					return cached.Value, true
				}
			}
			// Cache is stale, remove it
			interp.variableLookupCache.InvalidateVariable(name)
		}

		interp.variableLookupCache.misses++

		// Perform regular lookup and cache the result
		for i := len(interp.vars) - 1; i >= 0; i-- {
			if value, found := interp.vars[i][name]; found {
				interp.variableLookupCache.CacheVariable(name, value, i)
				return value, true
			}
		}
		return nil, false
	}

	// Fallback to standard lookup
	for i := len(interp.vars) - 1; i >= 0; i-- {
		thisVars := interp.vars[i]
		if v, ok := thisVars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (interp *interpreter) executeBlock(block Block) {
	for _, s := range block {
		interp.executeStatement(s)
	}
}

type iteratorType interface {
	HasNext() bool
	Value() Value
}

type listIterator struct {
	values []Value
	index  int
}

func (li *listIterator) HasNext() bool {
	return li.index < len(li.values)
}

func (li *listIterator) Value() Value {
	v := li.values[li.index]
	li.index++
	return v
}

func getIterator(_ Position, value Value) iteratorType {
	switch iterable := value.(type) {
	case string:
		strs := []Value{}
		for _, r := range iterable {
			strs = SmartAppend(strs, string(r))
		}
		return &listIterator{strs, 0}
	case *[]Value:
		return &listIterator{*iterable, 0}
	case map[string]Value:
		keys := make([]Value, len(iterable))
		i := 0
		for key := range iterable {
			keys[i] = key
			i++
		}
		return &listIterator{keys, 0}
	default:
		// Return nil iterator to indicate error
		return nil
	}
}

func (interp *interpreter) assignSubscript(pos Position, container, subscript, value Value) error {
	switch c := container.(type) {
	case *[]Value:
		if s, ok := subscript.(int); ok {
			if s < 0 || s >= len(*c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			(*c)[s] = value
			return nil
		} else {
			panic(typeError(pos, "array subscript must be an integer"))
		}
	case map[string]Value:
		if s, ok := subscript.(string); ok {
			c[s] = value
			return nil
		} else {
			panic(typeError(pos, "object subscript must be a string"))
		}
	default:
		panic(typeError(pos, "can only assign to subscript of array or object"))
	}
}

// evaluateAssignmentValue evaluates assignment operators like +=, -=, etc.
func (interp *interpreter) evaluateAssignmentValue(operator Token, target any, value Expression) Value {
	rightValue := interp.evaluate(value)

	// For simple assignment, just return the right value
	if operator == ASSIGN {
		return rightValue
	}

	// For compound assignments, get current value and perform operation
	var currentValue Value
	switch t := target.(type) {
	case string: // Variable name
		if v, ok := interp.lookup(t); ok {
			currentValue = v
		} else {
			panic(nameError(value.Position(), "name %q not found", t))
		}
	default:
		panic(fmt.Errorf("unsupported assignment target type"))
	}

	// Perform the compound operation
	switch operator {
	case PLUSEQUAL:
		return interp.evalPlus(value.Position(), currentValue, rightValue)
	case MINUSEQUAL:
		return evalMinus(value.Position(), currentValue, rightValue)
	case TIMESEQUAL:
		return evalTimes(value.Position(), currentValue, rightValue)
	case DIVIDEEQUAL:
		return evalDivide(value.Position(), currentValue, rightValue)
	case MODULOEQUAL:
		return evalModulo(value.Position(), currentValue, rightValue)
	default:
		panic(fmt.Sprintf("unknown assignment operator %v", operator))
	}
}

// evaluateSubscriptAssignmentValue evaluates assignment operators for subscripts
func (interp *interpreter) evaluateSubscriptAssignmentValue(operator Token, container, subscript Value, value Expression) Value {
	rightValue := interp.evaluate(value)

	// For simple assignment, just return the right value
	if operator == ASSIGN {
		return rightValue
	}

	// For compound assignments, get current value and perform operation
	currentValue := evalSubscript(value.Position(), container, subscript)

	// Perform the compound operation
	switch operator {
	case PLUSEQUAL:
		return interp.evalPlus(value.Position(), currentValue, rightValue)
	case MINUSEQUAL:
		return evalMinus(value.Position(), currentValue, rightValue)
	case TIMESEQUAL:
		return evalTimes(value.Position(), currentValue, rightValue)
	case DIVIDEEQUAL:
		return evalDivide(value.Position(), currentValue, rightValue)
	case MODULOEQUAL:
		return evalModulo(value.Position(), currentValue, rightValue)
	default:
		panic(fmt.Errorf("unknown assignment operator %v", operator))
	}
}

func (interp *interpreter) executeStatement(s Statement) {
	interp.stats.Ops++
	// Track current position for better error reporting
	interp.currentPos = s.Position()
	switch s := s.(type) {
	case *Assign:
		switch target := s.Target.(type) {
		case *Variable:
			newValue := interp.evaluateAssignmentValue(s.Operator, target.Name, s.Value)
			interp.assign(target.Name, newValue)
		case *Subscript:
			container := interp.evaluate(target.Container)
			subscript := interp.evaluate(target.Subscript)
			newValue := interp.evaluateSubscriptAssignmentValue(s.Operator, container, subscript, s.Value)
			interp.assignSubscript(target.Subscript.Position(), container, subscript, newValue)
		default:
			// Parser should never get us here
			panic("can only assign to variable or subscript")
		}
	case *If:
		cond := interp.evaluate(s.Condition)
		if c, ok := cond.(bool); ok {
			if c {
				interp.executeBlock(s.Body)
			} else if len(s.Else) > 0 {
				interp.executeBlock(s.Else)
			}
		} else {
			// Check if it's an error first
			if err, ok := cond.(error); ok {
				panic(err)
			}
			panic(typeError(s.Condition.Position(), "if condition must be bool, got %s", typeName(cond)))
		}
	case *While:
		func() {
			defer func() {
				if r := recover(); r != nil {
					if r == "__break__" {
						// Normal break, just exit the loop
						return
					}
					// Re-panic other exceptions
					panic(r)
				}
			}()

			for {
				cond := interp.evaluate(s.Condition)
				if c, ok := cond.(bool); ok {
					if !c {
						break
					}
					func() {
						defer func() {
							if r := recover(); r != nil {
								switch r.(type) {
								case BreakException:
									// Break out of the loop
									panic("__break__")
								case ContinueException:
									// Continue to next iteration
									return
								default:
									// Re-panic other exceptions
									panic(r)
								}
							}
						}()
						interp.executeBlock(s.Body)
					}()
				} else {
					// Check if it's an error first
					if err, ok := cond.(error); ok {
						panic(err)
					}
					panic(typeError(s.Condition.Position(), "while condition must be bool, got %T", cond))
				}
			}
		}()
	case *For:
		func() {
			defer func() {
				if r := recover(); r != nil {
					if r == "__break__" {
						// Normal break, just exit the loop
						return
					}
					// Re-panic other exceptions
					panic(r)
				}
			}()

			iterable := interp.evaluate(s.Iterable)
			iterator := getIterator(s.Iterable.Position(), iterable)
			for iterator.HasNext() {
				interp.assign(s.Name, iterator.Value())
				func() {
					defer func() {
						if r := recover(); r != nil {
							switch r.(type) {
							case BreakException:
								// Break out of the loop
								panic("__break__")
							case ContinueException:
								// Continue to next iteration
								return
							default:
								// Re-panic other exceptions
								panic(r)
							}
						}
					}()
					interp.executeBlock(s.Body)
				}()
			}
		}()
	case *TryCatch:
		// Create a new scope for the try block
		interp.pushScope(make(map[string]Value))

		// Execute the try block and catch any errors
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Pop the try scope
					interp.popScope()

					// Create a new scope for the catch block
					interp.pushScope(make(map[string]Value))

					// Assign the error to the error variable
					var errValue Value
					switch e := r.(type) {
					case Error:
						errValue = e.Error()
					case error:
						errValue = e.Error()
					case returnResult:
						// Re-panic return statements to handle them normally
						panic(r)
					// case BreakException, ContinueException:
					// 	// Re-panic break and continue to be handled by loops
					// 	panic(r)
					default:
						errValue = fmt.Sprintf("%v", r)
					}

					// Assign the error to the catch variable
					interp.assign(s.ErrVar, errValue)

					// Execute the catch block
					interp.executeBlock(s.CatchBlock)

					// Pop the catch scope
					interp.popScope()
				}
			}()

			// Execute the try block
			interp.executeBlock(s.TryBlock)

			// Pop the try scope if no error occurred
			interp.popScope()
		}()
	case *ExpressionStatement:
		interp.evaluate(s.Expression)
	case *FunctionDefinition:
		interp.mutex.RLock()
		closure := interp.vars[len(interp.vars)-1]
		interp.mutex.RUnlock()
		interp.assign(s.Name, &userFunction{s.Name, s.Parameters, s.Ellipsis, s.Body, closure, s.Memoized})
	case *Return:
		result := interp.evaluate(s.Result)
		panic(returnResult{result, s.Position()})
	case *Break:
		panic(BreakException{s.Position()})
	case *Continue:
		panic(ContinueException{s.Position()})
	case *Import:
		interp.executeImport(s)
	default:
		// Parser should never get us here
		panic(fmt.Sprintf("unexpected statement type %T", s))
	}
}

func (interp *interpreter) execute(prog *Program) {
	for _, statement := range prog.Statements {
		interp.executeStatement(statement)
	}

	// Automatically call the main function if it exists
	mainFunc, ok := interp.lookup("main")
	if ok {
		if fn, ok := mainFunc.(*userFunction); ok {
			// Hanya jalankan main() jika kita bukan dalam konteks unit test
			// Unit test biasanya memanggil fungsi secara eksplisit
			if !interp.inUnitTest {
				interp.callFunction(Position{}, fn, []Value{})
			}
		}
	}
}

func newInterpreter(config *Config) *interpreter {
	interp := new(interpreter)

	// Set global memory layout configuration
	if config.MemoryLayout != nil {
		SetGlobalMemoryLayoutConfig(config.MemoryLayout)
	} else {
		SetGlobalMemoryLayoutConfig(DefaultMemoryLayoutConfig())
	}

	// Initialize optimization caches
	interp.memoCache = make(map[string]Value)
	interp.stringBuilderPool = sync.Pool{
		New: func() any {
			return &strings.Builder{}
		},
	}
	interp.arrayPool = sync.Pool{
		New: func() any {
			return make([]Value, 0, 16) // Pre-allocate with capacity 16
		},
	}
	// Initialize variable lookup cache if enabled
	if IsVariableLookupCacheEnabled() {
		interp.variableLookupCache = NewVariableLookupCache(GetVariableLookupCacheSize())
	}
	// Initialize expression optimizer
	interp.optimizer = NewConstantFolder(GetGlobalExpressionOptimizer())

	// Initialize builtin dispatcher
	InitializeBuiltinDispatcher()

	interp.pushScope(make(map[string]Value))
	for k, v := range builtins {
		interp.assign(k, v)
	}

	// Add mathematical constants
	interp.assign("PI", Value(math.Pi))
	interp.assign("E", Value(math.E))
	interp.assign("TAU", Value(2*math.Pi))
	interp.assign("PHI", Value((1+math.Sqrt(5))/2)) // Golden ratio
	interp.assign("LN2", Value(math.Ln2))
	interp.assign("LN10", Value(math.Ln10))
	interp.assign("SQRT2", Value(math.Sqrt2))
	interp.assign("SQRT3", Value(math.Sqrt(3)))

	for k, v := range config.Vars {
		interp.assign(k, v)
	}
	interp.args = config.Args
	interp.stdin = config.Stdin
	if interp.stdin == nil {
		interp.stdin = os.Stdin
	}
	interp.stdinReader = bufio.NewReader(interp.stdin)
	interp.stdout = config.Stdout
	if interp.stdout == nil {
		interp.stdout = os.Stdout
	}
	interp.exit = config.Exit
	if interp.exit == nil {
		interp.exit = os.Exit
	}
	interp.inUnitTest = config.IsUnitTest
	return interp
}

// Evaluate takes a parsed Expression and interpreter config and evaluates the
// expression, returning the Value of the expression, interpreter statistics,
// and an error which is nil on success or an interpreter.Error if there's an
// error.
func Evaluate(expr Expression, config *Config) (v Value, stats *Stats, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert to interpreter.Error or re-panic
			err = r.(Error)
		}
	}()
	interp := newInterpreter(config)
	v = interp.evaluate(expr)
	stats = &interp.stats
	return
}

// Execute takes a parsed Program and interpreter config and interprets the
// program. Return interpreter statistics, and an error which is nil on
// success or an interpreter.Error if there's an error.
func Execute(prog *Program, config *Config) (stats *Stats, err error) {
	interp := newInterpreter(config)
	defer func() {
		if r := recover(); r != nil {
			// Use the current position from interpreter for better error reporting
			currentPos := interp.currentPos
			// If currentPos is zero value, fall back to 1:1
			if currentPos.Line == 0 {
				currentPos = Position{Line: 1, Column: 1}
			}

			switch e := r.(type) {
			case BreakException:
				err = runtimeError(e.pos, "break statement not within a loop")
			case ContinueException:
				err = runtimeError(e.pos, "continue statement not within a loop")
			case ErrorInterpreter:
				// Handle all errors that implement ErrorInterpreter (NameError, ValueError, TypeError, etc.)
				err = e
			case Error:
				err = e
			case returnResult:
				err = runtimeError(e.pos, "can't return at top level")
			case error:
				// Handle other error types that implement the error interface
				// but are not specifically handled above
				errorMsg := r.(error).Error()
				// Convert technical error messages to user-friendly ones
				if strings.Contains(errorMsg, "nil pointer dereference") || strings.Contains(errorMsg, "invalid memory address") {
					err = runtimeError(currentPos, "an error occurred while accessing data - this might be caused by undefined variables or invalid operations")
				} else if strings.Contains(errorMsg, "index out of range") {
					err = runtimeError(currentPos, "array or string index is out of bounds")
				} else if strings.Contains(errorMsg, "slice bounds out of range") {
					err = runtimeError(currentPos, "slice operation is out of bounds")
				} else {
					err = runtimeError(currentPos, "unexpected runtime error: %s", errorMsg)
				}
			default:
				// Handle non-error panic values
				err = runtimeError(currentPos, "unexpected runtime error occurred")
			}
		}
	}()
	interp.execute(prog)
	stats = &interp.stats
	return
}

// executeImport handles importing and executing .din files
func (interp *interpreter) executeImport(s *Import) {
	// Read the file content
	content, err := os.ReadFile(s.Filename)
	if err != nil {
		// Convert to error instead of panic
		panic(runtimeError(s.Position(), "failed to import file '%s': %s", s.Filename, err))
	}

	// Parse the imported file
	prog, err := ParseProgram(content)
	if err != nil {
		// Convert to error instead of panic
		panic(runtimeError(s.Position(), "failed to parse imported file '%s': %s", s.Filename, err))
	}

	// Execute the imported program
	// Note: This will execute in the current scope, so variables and functions
	// from the imported file will be available in the current context
	for _, statement := range prog.Statements {
		interp.executeStatement(statement)
	}
}

// Helper functions for better error handling and debugging

// WrapError wraps an error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// IsUserDefinedFunction checks if a value is a user-defined function
func IsUserDefinedFunction(value Value) bool {
	_, ok := value.(*userFunction)
	return ok
}

// IsBuiltinFunction checks if a value is a builtin function
func IsBuiltinFunction(value Value) bool {
	_, ok := value.(builtinFunction)
	return ok
}

// IsFunction checks if a value is any kind of function
func IsFunction(value Value) bool {
	return IsUserDefinedFunction(value) || IsBuiltinFunction(value)
}

// CreateEmptyScope creates an empty variable scope
func CreateEmptyScope() map[string]Value {
	return make(map[string]Value)
}

// CopyScope creates a copy of a variable scope
func CopyScope(scope map[string]Value) map[string]Value {
	newScope := make(map[string]Value)
	maps.Copy(newScope, scope)
	return newScope
}

// getStringBuilder gets a string builder from the pool
func (interp *interpreter) getStringBuilder() *strings.Builder {
	return interp.stringBuilderPool.Get().(*strings.Builder)
}

// putStringBuilder returns a string builder to the pool
func (interp *interpreter) putStringBuilder(sb *strings.Builder) {
	sb.Reset()
	interp.stringBuilderPool.Put(sb)
}

// getArray gets an array from the pool
func (interp *interpreter) getArray() []Value {
	arr := interp.arrayPool.Get().([]Value)
	return arr[:0] // Reset length but keep capacity
}

// getMemoKey generates a memoization key for function calls
// EXPERIMENTAL: This function is part of experimental memoization feature
func getMemoKey(funcName string, args []Value) string {
	var sb strings.Builder
	sb.WriteString(funcName)
	sb.WriteString(":")
	for i, arg := range args {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%v", arg))
	}
	return sb.String()
}

// Map pool helpers
func (interp *interpreter) getMap() map[string]Value {
	if m := interp.mapPool.Get(); m != nil {
		return m.(map[string]Value)
	}
	return make(map[string]Value)
}
