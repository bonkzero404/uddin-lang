package interpreter

// Type-specific value extractors for better performance
// Optimized with inline hints for better performance
//
//go:inline
func asInt(v Value) (int, bool) {
	if i, ok := v.(int); ok {
		return i, true
	}
	return 0, false
}

//go:inline
func asFloat(v Value) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	return 0, false
}

//go:inline
func asString(v Value) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

//go:inline
func asArray(v Value) (*[]Value, bool) {
	if arr, ok := v.(*[]Value); ok {
		return arr, true
	}
	return nil, false
}

//go:inline
func asMap(v Value) (map[string]Value, bool) {
	if m, ok := v.(map[string]Value); ok {
		return m, true
	}
	return nil, false
}

// Evaluator handles expression evaluation
type Evaluator struct {
	env       *Environment
	stats     *Stats
	interp    *interpreter
	optimizer *ConstantFolder
}

// NewEvaluator creates a new expression evaluator
func NewEvaluator(env *Environment, stats *Stats, interp *interpreter) *Evaluator {
	return &Evaluator{
		env:       env,
		stats:     stats,
		interp:    interp,
		optimizer: NewConstantFolder(GetGlobalExpressionOptimizer()),
	}
}

// Optimized EvaluateExpression with type-specific fast paths
func (e *Evaluator) EvaluateExpression(expr Expression) Value {
	e.stats.Ops++

	// Try constant folding optimization first
	if e.optimizer != nil {
		if result, ok := e.optimizer.FoldConstants(expr); ok {
			return result
		}
	}

	// Fast path for most common expression types
	if literal, ok := expr.(*Literal); ok {
		return literal.Value
	}
	if variable, ok := expr.(*Variable); ok {
		if value, exists := e.env.Lookup(variable.Name); exists {
			return value
		}
		panic(nameError(variable.Position(), "name '%s' is not defined", variable.Name))
	}
	if binary, ok := expr.(*Binary); ok {
		return e.evaluateBinary(binary)
	}
	if call, ok := expr.(*Call); ok {
		return e.evaluateCall(call)
	}

	// Less common types
	switch node := expr.(type) {
	case *Unary:
		return e.evaluateUnary(node)
	case *Ternary:
		return e.evaluateTernary(node)
	case *List:
		return e.evaluateList(node)
	case *Map:
		return e.evaluateMap(node)
	case *Subscript:
		return e.evaluateSubscript(node)
	case *FunctionExpression:
		return &userFunction{
			Name:       "<anonymous>",
			Parameters: node.Parameters,
			Ellipsis:   node.Ellipsis,
			Body:       node.Body,
			Closure:    e.env.vars[len(e.env.vars)-1],
			Memoized:   false,
		}
	default:
		panic(runtimeError(expr.Position(), "unexpected expression type %T", expr))
	}
}

// evaluateBinary evaluates binary operations
func (e *Evaluator) evaluateBinary(node *Binary) Value {
	// Handle logical operators with short-circuit evaluation
	switch node.Operator {
	case AND:
		left := e.EvaluateExpression(node.Left)
		if !IsTruthy(left) {
			return false
		}
		right := e.EvaluateExpression(node.Right)
		return IsTruthy(right)

	case OR:
		left := e.EvaluateExpression(node.Left)
		if IsTruthy(left) {
			return true
		}
		right := e.EvaluateExpression(node.Right)
		return IsTruthy(right)

	case XOR:
		left := e.EvaluateExpression(node.Left)
		right := e.EvaluateExpression(node.Right)
		// XOR returns true if exactly one operand is truthy
		leftTruthy := IsTruthy(left)
		rightTruthy := IsTruthy(right)
		return leftTruthy != rightTruthy
	}

	// Evaluate both operands for other operators
	left := e.EvaluateExpression(node.Left)
	right := e.EvaluateExpression(node.Right)

	switch node.Operator {
	case PLUS:
		return e.evalPlus(node.Position(), left, right)
	case MINUS:
		return e.evalMinus(node.Position(), left, right)
	case TIMES:
		return e.evalTimes(node.Position(), left, right)
	case DIVIDE:
		return e.evalDivide(node.Position(), left, right)
	case MODULO:
		return e.evalModulo(node.Position(), left, right)
	case EQUAL:
		return e.evalEqual(left, right)
	case NOTEQUAL:
		return !e.evalEqual(left, right)
	case LT:
		return e.evalLess(node.Position(), left, right)
	case LTE:
		return !e.evalLess(node.Position(), right, left)
	case GT:
		return e.evalLess(node.Position(), right, left)
	case GTE:
		return !e.evalLess(node.Position(), left, right)
	case IN:
		return e.evalIn(node.Position(), left, right)
	default:
		panic(runtimeError(node.Position(), "unknown binary operator: %v", node.Operator))
	}
}

// evaluateUnary evaluates unary operations
func (e *Evaluator) evaluateUnary(node *Unary) Value {
	operand := e.EvaluateExpression(node.Operand)

	switch node.Operator {
	case MINUS:
		return e.evalNegative(node.Position(), operand)
	case NOT:
		return !IsTruthy(operand)
	default:
		panic(runtimeError(node.Position(), "unknown unary operator: %v", node.Operator))
	}
}

// evaluateTernary evaluates ternary conditional expressions
func (e *Evaluator) evaluateTernary(node *Ternary) Value {
	condition := e.EvaluateExpression(node.Condition)

	if IsTruthy(condition) {
		return e.EvaluateExpression(node.TrueExpr)
	} else {
		return e.EvaluateExpression(node.FalseExpr)
	}
}

// evaluateCall evaluates function calls
func (e *Evaluator) evaluateCall(node *Call) Value {
	function := e.EvaluateExpression(node.Function)

	// Evaluate arguments
	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		args[i] = e.EvaluateExpression(arg)
	}

	// Call the function based on its type
	switch fn := function.(type) {
	case *userFunction:
		return e.callUserFunction(fn, node.Position(), args)
	case builtinFunction:
		// Try builtin dispatcher first for optimized dispatch
		dispatcher := GetGlobalBuiltinDispatcher()
		// Use the actual function name (fn.Name) instead of the formatted name
		if result, dispatched := dispatcher.DispatchBuiltinFunction(fn.Name, e.interp, node.Position(), args); dispatched {
			e.stats.BuiltinCalls++
			return result
		}

		// Fallback to original method if not in dispatcher
		// Create a temporary interpreter for builtin calls
		interp := &interpreter{
			vars:       e.env.vars,
			args:       e.env.args,
			stdin:      e.env.stdin,
			stdout:     e.env.stdout,
			exit:       e.env.exit,
			stats:      *e.stats,
			inUnitTest: e.env.inUnitTest,
		}
		result := fn.call(interp, node.Position(), args)
		*e.stats = interp.stats // Update stats
		return result
	default:
		// Check if it's a function type wrapped in any (from tagged values)
		if interfaceVal, ok := function.(any); ok {
			if userFn, ok := interfaceVal.(*userFunction); ok {
				return e.callUserFunction(userFn, node.Position(), args)
			}
			if builtinFn, ok := interfaceVal.(builtinFunction); ok {
				// Try builtin dispatcher first for optimized dispatch
				dispatcher := GetGlobalBuiltinDispatcher()
				if result, dispatched := dispatcher.DispatchBuiltinFunction(builtinFn.name(), e.interp, node.Position(), args); dispatched {
					e.stats.BuiltinCalls++
					return result
				}

				// Fallback to original method if not in dispatcher
				interp := &interpreter{
					vars:       e.env.vars,
					args:       e.env.args,
					stdin:      e.env.stdin,
					stdout:     e.env.stdout,
					exit:       e.env.exit,
					stats:      *e.stats,
					inUnitTest: e.env.inUnitTest,
				}
				result := builtinFn.call(interp, node.Position(), args)
				*e.stats = interp.stats // Update stats
				return result
			}
		}
		panic(typeError(node.Position(), "can't call non-function type %s", typeName(function)))
	}
}

// evaluateList evaluates list literals
func (e *Evaluator) evaluateList(node *List) Value {
	values := make([]Value, len(node.Values))
	for i, expr := range node.Values {
		values[i] = e.EvaluateExpression(expr)
	}
	return &values
}

// evaluateMap evaluates map literals
func (e *Evaluator) evaluateMap(node *Map) Value {
	result := make(map[string]Value)
	for _, item := range node.Items {
		key := e.EvaluateExpression(item.Key)
		value := e.EvaluateExpression(item.Value)

		keyStr, ok := key.(string)
		if !ok {
			panic(typeError(item.Key.Position(), "map key must be string, got %T", key))
		}

		result[keyStr] = value
	}
	return result
}

// evaluateSubscript evaluates array/map indexing with optimized type checking
func (e *Evaluator) evaluateSubscript(node *Subscript) Value {
	container := e.EvaluateExpression(node.Container)
	index := e.EvaluateExpression(node.Subscript)

	// Use optimized type checking with existing functions
	switch c := container.(type) {
	case *[]Value: // Array is a pointer to slice
		idx, ok := asInt(index)
		if !ok {
			panic(typeError(node.Position(), "array index must be integer, got %T", index))
		}

		// Handle negative indexing (Python-style)
		if idx < 0 {
			idx = len(*c) + idx
		}

		if idx < 0 || idx >= len(*c) {
			panic(valueError(node.Position(), "array index %d out of range [0:%d]", idx, len(*c)))
		}
		return (*c)[idx]

	case map[string]Value:
		key, ok := asString(index)
		if !ok {
			panic(typeError(node.Position(), "map key must be string, got %T", index))
		}
		if value, exists := c[key]; exists {
			return value
		}
		return nil

	case string:
		idx, ok := asInt(index)
		if !ok {
			panic(typeError(node.Position(), "string index must be integer, got %T", index))
		}

		// Handle negative indexing (Python-style)
		if idx < 0 {
			idx = len(c) + idx
		}

		if idx < 0 || idx >= len(c) {
			panic(valueError(node.Position(), "string index %d out of range [0:%d]", idx, len(c)))
		}
		return string(c[idx])

	default:
		panic(typeError(node.Position(), "object is not subscriptable: %T", container))
	}
}

// callUserFunction calls a user-defined function
func (e *Evaluator) callUserFunction(fn *userFunction, pos Position, args []Value) Value {
	e.stats.UserCalls++

	// Check memoization cache for functions marked with memo
	// EXPERIMENTAL: Memoization is experimental and may consume significant memory
	if fn.Name != "" && fn.Memoized {
		memoKey := getMemoKey(fn.Name, args)
		if cached, exists := e.interp.getMemoValue(memoKey); exists {
			return cached
		}
	}

	// Check parameter count
	if len(args) != len(fn.Parameters) {
		panic(valueError(pos, "function %s expects %d arguments, got %d",
			fn.Name, len(fn.Parameters), len(args)))
	}

	// Save current scope stack
	oldVars := e.env.vars

	// Set up function's closure as the base scope
	if fn.Closure != nil {
		// Create a new scope with the closure combined with current scope
		newVars := make([]map[string]Value, len(oldVars)+1)
		newVars[0] = fn.Closure
		copy(newVars[1:], oldVars)
		e.env.vars = newVars
	} else {
		// Create empty scope
		e.env.vars = []map[string]Value{make(map[string]Value)}
	}

	// Ensure the function can find itself for recursion
	if fn.Name != "" {
		e.env.vars[0][fn.Name] = fn
	}

	// Create new scope for function parameters
	e.env.PushScope()
	for i, param := range fn.Parameters {
		e.env.Assign(param, args[i])
	}

	// Execute function body
	var result Value
	defer func() {
		// Restore original scope stack
		e.env.vars = oldVars

		// Handle return values
		if r := recover(); r != nil {
			if ret, ok := r.(returnResult); ok {
				result = ret.value
				// Cache the result for memoized functions
				// EXPERIMENTAL: Storing result in experimental memoization cache
				if fn.Name != "" && fn.Memoized && result != nil {
					memoKey := getMemoKey(fn.Name, args)
					e.interp.setMemoValue(memoKey, result)
				}
			} else {
				panic(r) // Re-panic other errors
			}
		}
	}()

	// Execute function body
	interp := &interpreter{
		vars:                e.env.vars,
		args:                e.env.args,
		stdin:               e.env.stdin,
		stdout:              e.env.stdout,
		exit:                e.env.exit,
		stats:               *e.stats,
		inUnitTest:          e.env.inUnitTest,
		memoCache:           e.interp.memoCache,           // Share memo cache
		productionMemoCache: e.interp.productionMemoCache, // Share production memo cache
	}
	interp.executeBlock(fn.Body)
	*e.stats = interp.stats // Update stats

	// Cache the result for memoized functions
	// EXPERIMENTAL: Storing result in experimental memoization cache
	if fn.Name != "" && fn.Memoized && result != nil {
		memoKey := getMemoKey(fn.Name, args)
		e.interp.setMemoValue(memoKey, result)
	}

	return result // Will be nil if no explicit return
}

// Arithmetic and comparison methods

// evalPlus handles addition operation
func (e *Evaluator) evalPlus(pos Position, l, r Value) Value {
	return e.interp.evalPlus(pos, l, r)
}

// evalMinus handles subtraction operation
func (e *Evaluator) evalMinus(pos Position, l, r Value) Value {
	return evalMinus(pos, l, r)
}

// evalTimes handles multiplication operation
func (e *Evaluator) evalTimes(pos Position, l, r Value) Value {
	return evalTimes(pos, l, r)
}

// evalDivide handles division operation
func (e *Evaluator) evalDivide(pos Position, l, r Value) Value {
	return evalDivide(pos, l, r)
}

// evalModulo handles modulo operation
func (e *Evaluator) evalModulo(pos Position, l, r Value) Value {
	return evalModulo(pos, l, r)
}

// evalEqual handles equality comparison
func (e *Evaluator) evalEqual(l, r Value) bool {
	result := evalEqual(Position{}, l, r)
	if b, ok := result.(bool); ok {
		return b
	}
	return false
}

// evalLess handles less-than comparison
func (e *Evaluator) evalLess(pos Position, l, r Value) bool {
	result := evalLess(pos, l, r)
	if b, ok := result.(bool); ok {
		return b
	}
	return false
}

// evalIn handles 'in' operator
func (e *Evaluator) evalIn(pos Position, l, r Value) bool {
	result := evalIn(pos, l, r)
	if b, ok := result.(bool); ok {
		return b
	}
	return false
}

// evalNegative handles unary minus
func (e *Evaluator) evalNegative(pos Position, v Value) Value {
	return evalNegative(pos, v)
}
