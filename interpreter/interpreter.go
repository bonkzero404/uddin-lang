package interpreter

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Value represents any runtime value in the language.
// This includes nil, boolean, integer, float, string, array, object, and function values.
type Value any



// interpreter represents the internal state of the interpreter.
// It maintains variable scopes, I/O streams, and execution statistics.
type interpreter struct {
	// vars is a stack of variable scopes, with the most local scope at the end
	vars []map[string]Value
	// args holds command-line arguments for the args() builtin
	args []string
	// stdin is the input stream for the read() builtin
	stdin io.Reader
	// stdout is the output stream for the print() builtin
	stdout io.Writer
	// exit is the function called by the exit() builtin
	exit func(int)
	// stats tracks execution statistics
	stats Stats
	inUnitTest bool
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
	PLUS:     evalPlus,                                                                     // Addition operator: +
	TIMES:    evalTimes,                                                                    // Multiplication operator: *
}

// ensureIntToFloats converts integer or float operands to float64 for arithmetic operations.
// If either operand is not a numeric type (int or float64), it panics with a type error.
//
// Parameters:
//   - pos: Position in source code for error reporting
//   - l: Left operand value
//   - r: Right operand value
//   - operation: String description of the operation for error messages
//
// Returns:
//   - Two float64 values representing the converted operands
func ensureIntToFloats(pos Position, l, r Value, operation string) (float64, float64) {
	// Try to convert left operand to float64
	lf, lok := l.(float64)
	if !lok {
		if li, lok := l.(int); lok {
			lf = float64(li) // Convert int to float64
		} else {
			panic(typeError(pos, "%s requires two floats or integers", operation))
		}
	}

	// Try to convert right operand to float64
	rf, rok := r.(float64)
	if !rok {
		if ri, rok := r.(int); rok {
			rf = float64(ri) // Convert int to float64
		} else {
			panic(typeError(pos, "%s requires two floats or integers", operation))
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
	switch l := l.(type) {
	case int:
		// Integer comparison
		if r, ok := r.(int); ok {
			return Value(l < r)
		}
		// Mixed int/float comparison
		if r, ok := r.(float64); ok {
			return Value(float64(l) < r)
		}

	case float64:
		// Float comparison
		if r, ok := r.(float64); ok {
			return Value(l < r)
		}
		// Mixed float/int comparison
		if r, ok := r.(int); ok {
			return Value(l < float64(r))
		}

	case string:
		// String lexicographical comparison
		if r, ok := r.(string); ok {
			return Value(l < r)
		}

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

func evalPlus(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case int:
		if r, ok := r.(int); ok {
			return Value(l + r)
		}
		if r, ok := r.(float64); ok {
			return Value(float64(l) + r)
		}
	case float64:
		if r, ok := r.(float64); ok {
			return Value(l + r)
		}
		if r, ok := r.(int); ok {
			return Value(l + float64(r))
		}
	case string:
		if r, ok := r.(string); ok {
			return Value(l + r)
		}
	case *[]Value:
		if r, ok := r.(*[]Value); ok {
			result := make([]Value, 0, len(*l)+len(*r))
			result = append(result, *l...)
			result = append(result, *r...)
			return Value(&result)
		}
	case map[string]Value:
		if r, ok := r.(map[string]Value); ok {
			result := make(map[string]Value)
			for k, v := range l {
				result[k] = v
			}
			for k, v := range r {
				result[k] = v
			}
			return Value(result)
		}
	}
	panic(typeError(pos, "+ requires two integers, strings, arrays, or objects"))
}

func evalMinus(pos Position, l, r Value) Value {
	if li, ok := l.(int); ok {
		if ri, ok := r.(int); ok {
			return Value(li - ri)
		} else if rf, ok := r.(float64); ok {
			return Value(float64(li) - rf)
		}
	} else if lf, ok := l.(float64); ok {
		if rf, ok := r.(float64); ok {
			return Value(lf - rf)
		} else if ri, ok := r.(int); ok {
			return Value(lf - float64(ri))
		}
	}

	panic(typeError(pos, "- requires two floats or integers, got %T and %T", l, r))
}

func evalTimes(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case int:
		switch r := r.(type) {
		case int:
			return Value(l * r)
		case float64:
			return Value(float64(l) * r)
		case string:
			if l < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			return Value(strings.Repeat(r, l))
		case *[]Value:
			lst := make([]Value, 0, len(*r)*l)
			for i := 0; i < l; i++ {
				lst = append(lst, (*r)...)
			}
			return Value(&lst)
		}
	case float64:
		if r, ok := r.(float64); ok {
			return Value(l * r)
		}
		if r, ok := r.(int); ok {
			return Value(l * float64(r))
		}
	case string:
		if r, ok := r.(int); ok {
			if r < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			return Value(strings.Repeat(l, r))
		}
	case *[]Value:
		if r, ok := r.(int); ok {
			if r < 0 {
				panic(valueError(pos, "can't multiply array by a negative number"))
			}
			lst := make([]Value, 0, len(*l)*r)
			for i := 0; i < r; i++ {
				lst = append(lst, (*l)...)
			}
			return Value(&lst)
		}
	}
	panic(typeError(pos, "* requires two integers or floats, or a string or array and an integer"))
}

func evalDivide(pos Position, l, r Value) Value {
	li, ri := ensureIntToFloats(pos, l, r, "/")
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(li / ri)
}

func evalModulo(pos Position, l, r Value) Value {
	li, ri := ensureIntToFloats(pos, l, r, "%")
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(int(li) % int(ri))
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
	defer func() {
		if r := recover(); r != nil {
			if result, ok := r.(returnResult); ok {
				ret = result.value
			} else {
				panic(r)
			}
		}
	}()
	return f.call(interp, pos, args)
}

func (interp *interpreter) evaluate(expr Expression) Value {
	interp.stats.Ops++
	switch e := expr.(type) {
	case *Binary:
		if f, ok := binaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Left), interp.evaluate(e.Right))
		} else if e.Operator == AND {
			return interp.evalAnd(e.Position(), e.Left, e.Right)
		} else if e.Operator == OR {
			return interp.evalOr(e.Position(), e.Left, e.Right)
		} else if e.Operator == XOR {
			return interp.evalXor(e.Position(), e.Left, e.Right)
		}
		// Parser should never give us this
		panic(fmt.Sprintf("unknown binary operator %v", e.Operator))
	case *Unary:
		if f, ok := unaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Operand))
		}
		// Parser should never give us this
		panic(fmt.Sprintf("unknown unary operator %v", e.Operator))
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
				args = append(args, interp.evaluate(a))
			}
			if e.Ellipsis {
				iterator := getIterator(e.Arguments[len(args)-1].Position(), args[len(args)-1])
				args = args[:len(args)-1]
				for iterator.HasNext() {
					args = append(args, iterator.Value())
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
		values := make([]Value, len(e.Values))
		for i, v := range e.Values {
			values[i] = interp.evaluate(v)
		}
		return Value(&values)
	case *Map:
		value := make(map[string]Value)
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
		return &userFunction{"", e.Parameters, e.Ellipsis, e.Body, closure}
	default:
		// Parser should never get us here
		panic(fmt.Sprintf("unexpected expression type %T", expr))
	}
}

func (interp *interpreter) pushScope(scope map[string]Value) {
	interp.vars = append(interp.vars, scope)
}

func (interp *interpreter) popScope() {
	interp.vars = interp.vars[:len(interp.vars)-1]
}

func (interp *interpreter) assign(name string, value Value) {
	interp.vars[len(interp.vars)-1][name] = value
}

func (interp *interpreter) lookup(name string) (Value, bool) {
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

func getIterator(pos Position, value Value) iteratorType {
	switch iterable := value.(type) {
	case string:
		strs := []Value{}
		for _, r := range iterable {
			strs = append(strs, string(r))
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
		panic(typeError(pos, "expected iterable (string, array, or object), got %s", typeName(value)))
	}
}

func (interp *interpreter) assignSubscript(pos Position, container, subscript, value Value) {
	switch c := container.(type) {
	case *[]Value:
		if s, ok := subscript.(int); ok {
			if s < 0 || s >= len(*c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			(*c)[s] = value
		} else {
			panic(typeError(pos, "array subscript must be an integer"))
		}
	case map[string]Value:
		if s, ok := subscript.(string); ok {
			c[s] = value
		} else {
			panic(typeError(pos, "object subscript must be a string"))
		}
	default:
		panic(typeError(pos, "can only assign to subscript of array or object"))
	}
}

func (interp *interpreter) executeStatement(s Statement) {
	interp.stats.Ops++
	switch s := s.(type) {
	case *Assign:
		switch target := s.Target.(type) {
		case *Variable:
			interp.assign(target.Name, interp.evaluate(s.Value))
		case *Subscript:
			container := interp.evaluate(target.Container)
			subscript := interp.evaluate(target.Subscript)
			value := interp.evaluate(s.Value)
			interp.assignSubscript(target.Subscript.Position(), container, subscript, value)
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
		closure := interp.vars[len(interp.vars)-1]
		interp.assign(s.Name, &userFunction{s.Name, s.Parameters, s.Ellipsis, s.Body, closure})
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
	interp.pushScope(make(map[string]Value))
	for k, v := range builtins {
		interp.assign(k, v)
	}

	// Add mathematical constants
	interp.assign("PI", Value(math.Pi))
	interp.assign("E", Value(math.E))
	interp.assign("TAU", Value(2 * math.Pi))
	interp.assign("PHI", Value((1 + math.Sqrt(5)) / 2))  // Golden ratio
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
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case Error:
				err = e
			case returnResult:
				err = runtimeError(e.pos, "can't return at top level")
			default:
				err = r.(error)
			}
		}
	}()
	interp := newInterpreter(config)
	interp.execute(prog)
	stats = &interp.stats
	return
}

// executeImport handles importing and executing .din files
func (interp *interpreter) executeImport(s *Import) {
	// Read the file content
	content, err := os.ReadFile(s.Filename)
	if err != nil {
		panic(runtimeError(s.Position(), "failed to import file '%s': %s", s.Filename, err))
	}

	// Parse the imported file
	prog, err := ParseProgram(content)
	if err != nil {
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
	for key, value := range scope {
		newScope[key] = value
	}
	return newScope
}
