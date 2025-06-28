package interpreter

// Evaluator handles expression evaluation
type Evaluator struct {
	env   *Environment
	stats *Stats
}

// NewEvaluator creates a new expression evaluator
func NewEvaluator(env *Environment, stats *Stats) *Evaluator {
	return &Evaluator{
		env:   env,
		stats: stats,
	}
}

// EvaluateExpression evaluates an expression and returns its value
func (e *Evaluator) EvaluateExpression(expr Expression) Value {
	e.stats.Ops++

	switch node := expr.(type) {
	case *Literal:
		return node.Value

	case *Variable:
		if value, exists := e.env.Lookup(node.Name); exists {
			return value
		}
		panic(nameError(node.Position(), "name '%s' is not defined", node.Name))

	case *Binary:
		return e.evaluateBinary(node)

	case *Unary:
		return e.evaluateUnary(node)

	case *Ternary:
		return e.evaluateTernary(node)

	case *Call:
		return e.evaluateCall(node)

	case *List:
		return e.evaluateList(node)

	case *Map:
		return e.evaluateMap(node)

	case *Subscript:
		return e.evaluateSubscript(node)

	case *FunctionExpression:
		return &userFunction{
			Name:       "",
			Parameters: node.Parameters,
			Body:       node.Body,
			Closure:    e.env.vars[len(e.env.vars)-1], // Capture current scope
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
		panic(typeError(node.Position(), "object is not callable: %T", function))
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

// evaluateSubscript evaluates array/map indexing
func (e *Evaluator) evaluateSubscript(node *Subscript) Value {
	container := e.EvaluateExpression(node.Container)
	index := e.EvaluateExpression(node.Subscript)

	switch c := container.(type) {
	case *[]Value:  // Array is a pointer to slice
		idx, ok := index.(int)
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
		key, ok := index.(string)
		if !ok {
			panic(typeError(node.Position(), "map key must be string, got %T", index))
		}
		if value, exists := c[key]; exists {
			return value
		}
		return nil

	case string:
		idx, ok := index.(int)
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

	// Check parameter count
	if len(args) != len(fn.Parameters) {
		panic(valueError(pos, "function %s expects %d arguments, got %d",
			fn.Name, len(fn.Parameters), len(args)))
	}

	// Save current scope stack
	oldVars := e.env.vars

	// Set up function's closure as the base scope
	if fn.Closure != nil {
		// Create a new scope with the closure
		e.env.vars = []map[string]Value{fn.Closure}
	} else {
		// Create empty scope
		e.env.vars = []map[string]Value{make(map[string]Value)}
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
			} else {
				panic(r) // Re-panic other errors
			}
		}
	}()

	// Execute function body
	interp := &interpreter{
		vars:       e.env.vars,
		args:       e.env.args,
		stdin:      e.env.stdin,
		stdout:     e.env.stdout,
		exit:       e.env.exit,
		stats:      *e.stats,
		inUnitTest: e.env.inUnitTest,
	}
	interp.executeBlock(fn.Body)
	*e.stats = interp.stats // Update stats

	return result // Will be nil if no explicit return
}

// Arithmetic and comparison methods

// evalPlus handles addition operation
func (e *Evaluator) evalPlus(pos Position, l, r Value) Value {
	return evalPlus(pos, l, r)
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
