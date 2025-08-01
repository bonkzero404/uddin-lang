package interpreter

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
)

// Program represents the root node of the abstract syntax tree (AST).
// It contains a block of statements that make up the program.
type Program struct {
	Statements Block
}

// String returns a string representation of the program.
func (p *Program) String() string {
	return p.Statements.String()
}

// Block represents a sequence of statements.
type Block []Statement

// String returns a string representation of the block.
func (b Block) String() string {
	lines := []string{}
	for _, stmt := range b {
		lines = SmartAppendString(lines, stmt.String())
	}
	return strings.Join(lines, "\n")
}

// Statement is an interface that all statement nodes in the AST must implement.
type Statement interface {
	// Position returns the source code position of the statement.
	Position() Position
	// String returns a string representation of the statement.
	String() string
}

// Assign represents an assignment statement (target = value).
type Assign struct {
	pos      Position   // Source position
	Target   Expression // Left-hand side of the assignment
	Value    Expression // Right-hand side of the assignment
	Operator Token      // Assignment operator (ASSIGN, PLUSEQUAL, etc.)
}

func (s *Assign) Position() Position { return s.pos }

// String returns a string representation of the assignment.
// UnmarshalJSON implements custom JSON unmarshaling for Assign
func (s *Assign) UnmarshalJSON(data []byte) error {
	type Alias Assign
	aux := &struct {
		Target json.RawMessage `json:"Target"`
		Value  json.RawMessage `json:"Value"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	target, err := unmarshalExpression(aux.Target)
	if err != nil {
		return err
	}
	s.Target = target

	value, err := unmarshalExpression(aux.Value)
	if err != nil {
		return err
	}
	s.Value = value

	return nil
}

func (s *Assign) String() string {
	opStr := "="
	switch s.Operator {
	case PLUSEQUAL:
		opStr = "+="
	case MINUSEQUAL:
		opStr = "-="
	case TIMESEQUAL:
		opStr = "*="
	case DIVIDEEQUAL:
		opStr = "/="
	case MODULOEQUAL:
		opStr = "%="
	}
	return fmt.Sprintf("%s %s %s", s.Target, opStr, s.Value)
}

// OuterAssign represents an assignment to an outer scope variable.
type OuterAssign struct {
	pos   Position   // Source position
	Name  string     // Name of the variable
	Value Expression // Value to assign
}

func (s *OuterAssign) Position() Position { return s.pos }

// String returns a string representation of the outer assignment.
func (s *OuterAssign) String() string {
	return fmt.Sprintf("outer %s = %s", s.Name, s.Value)
}

// If represents an if-else conditional statement.
type If struct {
	pos       Position   // Source position
	Condition Expression // The condition to evaluate
	Body      Block      // The block to execute if condition is true
	Else      Block      // The block to execute if condition is false (optional)
}

func (s *If) Position() Position { return s.pos }

// indent adds four spaces to the beginning of each line in the given string.
// Used for pretty-printing blocks of code.
func indent(s string) string {
	input := strings.Split(s, "\n")
	output := []string{}
	for _, line := range input {
		output = append(output, "    "+line)
	}
	return strings.Join(output, "\n")
}

// String returns a string representation of the if statement.
// UnmarshalJSON implements custom JSON unmarshaling for If
func (s *If) UnmarshalJSON(data []byte) error {
	type Alias If
	aux := &struct {
		Condition json.RawMessage `json:"Condition"`
		Body      json.RawMessage `json:"Body"`
		Else      json.RawMessage `json:"Else"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	condition, err := unmarshalExpression(aux.Condition)
	if err != nil {
		return err
	}
	s.Condition = condition

	if len(aux.Body) > 0 {
		var body Block
		if err := json.Unmarshal(aux.Body, &body); err != nil {
			return err
		}
		s.Body = body
	}

	if len(aux.Else) > 0 {
		var elseBlock Block
		if err := json.Unmarshal(aux.Else, &elseBlock); err != nil {
			return err
		}
		s.Else = elseBlock
	}

	return nil
}

func (s *If) String() string {
	str := fmt.Sprintf("if (%s) then:\n%s", s.Condition, indent(s.Body.String()))
	if len(s.Else) > 0 {
		str += fmt.Sprintf("\nelse:\n%s", indent(s.Else.String()))
	}
	str += "\nend"
	return str
}

// While represents a while loop statement.
type While struct {
	pos       Position   // Source position
	Condition Expression // The condition to evaluate
	Body      Block      // The block to execute while condition is true
}

func (s *While) Position() Position { return s.pos }

// String returns a string representation of the while statement.
// UnmarshalJSON implements custom JSON unmarshaling for While
func (s *While) UnmarshalJSON(data []byte) error {
	type Alias While
	aux := &struct {
		Condition json.RawMessage `json:"Condition"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	condition, err := unmarshalExpression(aux.Condition)
	if err != nil {
		return err
	}
	s.Condition = condition

	return nil
}

func (s *While) String() string {
	return fmt.Sprintf("while (%s):\n%s\nend", s.Condition, indent(s.Body.String()))
}

// For represents a for-in loop statement.
type For struct {
	pos      Position   // Source position
	Name     string     // Loop variable name
	Iterable Expression // The collection to iterate over
	Body     Block      // The block to execute for each item
}

func (s *For) Position() Position { return s.pos }

// String returns a string representation of the for statement.
// UnmarshalJSON implements custom JSON unmarshaling for For
func (s *For) UnmarshalJSON(data []byte) error {
	type Alias For
	aux := &struct {
		Iterable json.RawMessage `json:"Iterable"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	iterable, err := unmarshalExpression(aux.Iterable)
	if err != nil {
		return err
	}
	s.Iterable = iterable

	return nil
}

func (s *For) String() string {
	return fmt.Sprintf("for (%s in %s):\n%s\nend", s.Name, s.Iterable, indent(s.Body.String()))
}

// TryCatch represents a try-catch statement for error handling.
type TryCatch struct {
	pos        Position // Source position
	TryBlock   Block    // The block to try executing
	ErrVar     string   // The variable name to hold the caught error
	CatchBlock Block    // The block to execute if an error occurs
}

func (s *TryCatch) Position() Position { return s.pos }

// String returns a string representation of the try-catch statement.
func (s *TryCatch) String() string {
	return fmt.Sprintf("try:\n%s\ncatch (%s):\n%s\nend",
		indent(s.TryBlock.String()), s.ErrVar, indent(s.CatchBlock.String()))
}

// Return represents a return statement.
type Return struct {
	pos    Position   // Source position
	Result Expression // The value to return
}

func (s *Return) Position() Position { return s.pos }

// String returns a string representation of the return statement.
// UnmarshalJSON implements custom JSON unmarshaling for Return
func (s *Return) UnmarshalJSON(data []byte) error {
	type Alias Return
	aux := &struct {
		Result json.RawMessage `json:"Result"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Result) > 0 {
		result, err := unmarshalExpression(aux.Result)
		if err != nil {
			return err
		}
		s.Result = result
	}

	return nil
}

func (s *Return) String() string {
	return fmt.Sprintf("return %s", s.Result)
}

// ExpressionStatement represents a statement that consists of just an expression.
type ExpressionStatement struct {
	pos        Position   // Source position
	Expression Expression // The expression
}

func (s *ExpressionStatement) Position() Position { return s.pos }

// String returns a string representation of the expression statement.
// UnmarshalJSON implements custom JSON unmarshaling for ExpressionStatement
func (s *ExpressionStatement) UnmarshalJSON(data []byte) error {
	type Alias ExpressionStatement
	aux := &struct {
		Expression json.RawMessage `json:"Expression"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	expr, err := unmarshalExpression(aux.Expression)
	if err != nil {
		return err
	}
	s.Expression = expr

	return nil
}

func (s *ExpressionStatement) String() string {
	return s.Expression.String()
}

// FunctionDefinition represents a function declaration statement.
type FunctionDefinition struct {
	pos        Position // Source position
	Name       string   // Function name
	Parameters []string // Parameter names
	Ellipsis   bool     // Whether the function accepts variable arguments
	Body       Block    // Function body
}

func (s *FunctionDefinition) Position() Position { return s.pos }

// String returns a string representation of the function definition.
func (s *FunctionDefinition) String() string {
	ellipsisStr := ""
	if s.Ellipsis {
		ellipsisStr = "..."
	}
	bodyStr := ""
	if len(s.Body) != 0 {
		bodyStr = "\n" + indent(s.Body.String())
	}
	return fmt.Sprintf("fun %s(%s%s):%s\nend",
		s.Name, strings.Join(s.Parameters, ", "), ellipsisStr, bodyStr)
}

// Expression is an interface that all expression nodes in the AST must implement.
type Expression interface {
	// Position returns the source code position of the expression.
	Position() Position
	// String returns a string representation of the expression.
	String() string
}

// Binary represents a binary operation (left operator right).
type Binary struct {
	pos      Position   // Source position
	Left     Expression // Left operand
	Operator Token      // Operator token
	Right    Expression // Right operand
}

func (e *Binary) Position() Position { return e.pos }

// String returns a string representation of the binary expression.
// UnmarshalJSON implements custom JSON unmarshaling for Binary
func (e *Binary) UnmarshalJSON(data []byte) error {
	type Alias Binary
	aux := &struct {
		Left  json.RawMessage `json:"Left"`
		Right json.RawMessage `json:"Right"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	left, err := unmarshalExpression(aux.Left)
	if err != nil {
		return err
	}
	e.Left = left

	right, err := unmarshalExpression(aux.Right)
	if err != nil {
		return err
	}
	e.Right = right

	return nil
}

// getOperatorPrecedence returns the precedence level of an operator
// Higher numbers indicate higher precedence
func getOperatorPrecedence(op Token) int {
	switch op {
	case TIMES, DIVIDE, MODULO:
		return 6 // multiply level
	case PLUS, MINUS:
		return 5 // addition level
	case LT, LTE, GT, GTE, IN:
		return 4 // comparison level
	case EQUAL, NOTEQUAL:
		return 3 // equality level
	case AND:
		return 2 // and level
	case XOR:
		return 1 // xor level
	case OR:
		return 0 // or level (lowest)
	default:
		return 7 // unknown operators get highest precedence
	}
}

// needsParentheses determines if an expression needs parentheses when used as operand
func needsParentheses(expr Expression, parentOp Token, isRightOperand bool) bool {
	binary, ok := expr.(*Binary)
	if !ok {
		return false
	}

	parentPrec := getOperatorPrecedence(parentOp)
	childPrec := getOperatorPrecedence(binary.Operator)

	// Lower precedence needs parentheses
	if childPrec < parentPrec {
		return true
	}

	// Same precedence: need parentheses for proper associativity
	if childPrec == parentPrec {
		// For same precedence, left operand needs parentheses if operators are different
		// This preserves original grouping like (a * b) / c vs a * b / c
		if !isRightOperand && binary.Operator != parentOp {
			return true
		}
		// Right operand always needs parentheses for same precedence
		if isRightOperand {
			return true
		}
	}

	return false
}

func (e *Binary) String() string {
	leftStr := e.Left.String()
	rightStr := e.Right.String()

	// Add parentheses if needed
	if needsParentheses(e.Left, e.Operator, false) {
		leftStr = "(" + leftStr + ")"
	}
	if needsParentheses(e.Right, e.Operator, true) {
		rightStr = "(" + rightStr + ")"
	}

	return fmt.Sprintf("%s %s %s", leftStr, e.Operator, rightStr)
}

// Unary represents a unary operation (operator operand).
type Unary struct {
	pos      Position   // Source position
	Operator Token      // Operator token
	Operand  Expression // The operand
}

func (e *Unary) Position() Position { return e.pos }

// String returns a string representation of the unary expression.
// UnmarshalJSON implements custom JSON unmarshaling for Unary
func (e *Unary) UnmarshalJSON(data []byte) error {
	type Alias Unary
	aux := &struct {
		Operand json.RawMessage `json:"Operand"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	operand, err := unmarshalExpression(aux.Operand)
	if err != nil {
		return err
	}
	e.Operand = operand

	return nil
}

func (e *Unary) String() string {
	space := ""
	// Special case for NOT operator to improve readability
	if e.Operator == NOT {
		space = " "
	}
	return fmt.Sprintf("%s%s%s", e.Operator, space, e.Operand)
}

// Call represents a function call expression.
type Call struct {
	pos       Position     // Source position
	Function  Expression   // The function to call
	Arguments []Expression // Function arguments
	Ellipsis  bool         // Whether to unpack the last argument
}

func (e *Call) Position() Position { return e.pos }

// String returns a string representation of the function call.
// UnmarshalJSON implements custom JSON unmarshaling for Call
func (e *Call) UnmarshalJSON(data []byte) error {
	type Alias Call
	aux := &struct {
		Function  json.RawMessage   `json:"Function"`
		Arguments []json.RawMessage `json:"Arguments"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	function, err := unmarshalExpression(aux.Function)
	if err != nil {
		return err
	}
	e.Function = function

	e.Arguments = make([]Expression, len(aux.Arguments))
	for i, argData := range aux.Arguments {
		arg, err := unmarshalExpression(argData)
		if err != nil {
			return err
		}
		e.Arguments[i] = arg
	}

	return nil
}

func (e *Call) String() string {
	args := []string{}
	for _, arg := range e.Arguments {
		args = append(args, arg.String())
	}
	ellipsisStr := ""
	if e.Ellipsis {
		ellipsisStr = "..."
	}
	return fmt.Sprintf("%s(%s%s)", e.Function, strings.Join(args, ", "), ellipsisStr)
}

// Literal represents a literal value (number, string, boolean, nil).
type Literal struct {
	pos   Position // Source position
	Value any      // The literal value
}

func (e *Literal) Position() Position { return e.pos }

// String returns a string representation of the literal.
func (e *Literal) String() string {
	if e.Value == nil {
		return "null"
	}
	if s, ok := e.Value.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	// Handle integer values first to preserve them as integers
	if i, ok := e.Value.(int); ok {
		return fmt.Sprintf("%d", i)
	}
	// Handle float formatting to preserve decimal format
	if f, ok := e.Value.(float64); ok {
		// Check if it's a whole number that was originally an integer
		if f == float64(int64(f)) {
			// Convert back to integer format for whole numbers
			return fmt.Sprintf("%d", int64(f))
		}
		// For small decimals like 0.000001, use fixed-point notation
		if f > 0 && f < 0.001 {
			return fmt.Sprintf("%.6f", f)
		}
		// For other values, use %g to remove trailing zeros
		return fmt.Sprintf("%g", f)
	}
	return fmt.Sprintf("%v", e.Value)
}

// List represents a list literal [value1, value2, ...].
type List struct {
	pos    Position     // Source position
	Values []Expression // List elements
}

func (e *List) Position() Position { return e.pos }

// String returns a string representation of the list.
// UnmarshalJSON implements custom JSON unmarshaling for List
func (e *List) UnmarshalJSON(data []byte) error {
	type Alias List
	aux := &struct {
		Values []json.RawMessage `json:"Values"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.Values = make([]Expression, len(aux.Values))
	for i, valueData := range aux.Values {
		value, err := unmarshalExpression(valueData)
		if err != nil {
			return err
		}
		e.Values[i] = value
	}

	return nil
}

func (e *List) String() string {
	values := []string{}
	for _, value := range e.Values {
		values = append(values, value.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ", "))
}

// MapItem represents a key-value pair in a map literal.
type MapItem struct {
	Key   Expression // Map key
	Value Expression // Map value
}

// Map represents a map literal {key1: value1, key2: value2, ...}.
type Map struct {
	pos   Position  // Source position
	Items []MapItem // Map entries
}

func (e *Map) Position() Position { return e.pos }

// String returns a string representation of the map.
// UnmarshalJSON implements custom JSON unmarshaling for MapItem
func (m *MapItem) UnmarshalJSON(data []byte) error {
	type Alias MapItem
	aux := &struct {
		Key   json.RawMessage `json:"Key"`
		Value json.RawMessage `json:"Value"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	key, err := unmarshalExpression(aux.Key)
	if err != nil {
		return err
	}
	m.Key = key

	value, err := unmarshalExpression(aux.Value)
	if err != nil {
		return err
	}
	m.Value = value

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Map
func (e *Map) UnmarshalJSON(data []byte) error {
	type Alias Map
	aux := &struct {
		Items []json.RawMessage `json:"Items"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.Items = make([]MapItem, len(aux.Items))
	for i, itemData := range aux.Items {
		var item MapItem
		if err := json.Unmarshal(itemData, &item); err != nil {
			return err
		}
		e.Items[i] = item
	}

	return nil
}

func (e *Map) String() string {
	if len(e.Items) == 0 {
		return "{}"
	}
	items := []string{}
	for _, item := range e.Items {
		items = append(items, fmt.Sprintf("%s: %s", item.Key, item.Value))
	}
	// Always use single-line format for consistency
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}

// FunctionExpression represents an anonymous function expression.
type FunctionExpression struct {
	pos        Position // Source position
	Parameters []string // Parameter names
	Ellipsis   bool     // Whether the function accepts variable arguments
	Body       Block    // Function body
}

func (e *FunctionExpression) Position() Position { return e.pos }

// String returns a string representation of the function expression.
func (e *FunctionExpression) String() string {
	ellipsisStr := ""
	if e.Ellipsis {
		ellipsisStr = "..."
	}
	bodyStr := ""
	if len(e.Body) != 0 {
		bodyStr = "\n" + indent(e.Body.String()) + "\n"
	}
	return fmt.Sprintf("fun(%s%s):%s\nend", strings.Join(e.Parameters, ", "), ellipsisStr, bodyStr)
}

// Subscript represents a container subscript expression (container[index]).
type Subscript struct {
	pos       Position   // Source position
	Container Expression // The container to index into (list, map, etc.)
	Subscript Expression // The index expression
}

func (e *Subscript) Position() Position { return e.pos }

// String returns a string representation of the subscript expression.
// UnmarshalJSON implements custom JSON unmarshaling for Subscript
func (e *Subscript) UnmarshalJSON(data []byte) error {
	type Alias Subscript
	aux := &struct {
		Container json.RawMessage `json:"Container"`
		Subscript json.RawMessage `json:"Subscript"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	container, err := unmarshalExpression(aux.Container)
	if err != nil {
		return err
	}
	e.Container = container

	subscript, err := unmarshalExpression(aux.Subscript)
	if err != nil {
		return err
	}
	e.Subscript = subscript

	return nil
}

func (e *Subscript) String() string {
	return fmt.Sprintf("%s[%s]", e.Container, e.Subscript)
}

// Variable represents a variable reference.
type Variable struct {
	pos  Position // Source position
	Name string   // Variable name
}

func (e *Variable) Position() Position { return e.pos }

// String returns a string representation of the variable.
func (e *Variable) String() string {
	return e.Name
}

// Ternary represents a ternary conditional expression (condition ? true_expr : false_expr).
type Ternary struct {
	pos       Position   // Source position
	Condition Expression // The condition to evaluate
	TrueExpr  Expression // Expression to return if condition is true
	FalseExpr Expression // Expression to return if condition is false
}

func (e *Ternary) Position() Position { return e.pos }

// String returns a string representation of the ternary expression.
// UnmarshalJSON implements custom JSON unmarshaling for Ternary
func (e *Ternary) UnmarshalJSON(data []byte) error {
	type Alias Ternary
	aux := &struct {
		Condition json.RawMessage `json:"Condition"`
		TrueExpr  json.RawMessage `json:"TrueExpr"`
		FalseExpr json.RawMessage `json:"FalseExpr"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	condition, err := unmarshalExpression(aux.Condition)
	if err != nil {
		return err
	}
	e.Condition = condition

	trueExpr, err := unmarshalExpression(aux.TrueExpr)
	if err != nil {
		return err
	}
	e.TrueExpr = trueExpr

	falseExpr, err := unmarshalExpression(aux.FalseExpr)
	if err != nil {
		return err
	}
	e.FalseExpr = falseExpr

	return nil
}

func (e *Ternary) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", e.Condition, e.TrueExpr, e.FalseExpr)
}

// Break represents a break statement in loops.
type Break struct {
	pos Position // Source position
}

func (s *Break) Position() Position { return s.pos }

// String returns a string representation of the break statement.
func (s *Break) String() string {
	return "break"
}

// MarshalJSON implements custom JSON marshaling for Break
func (s *Break) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"Type": "Break",
		"pos":  s.pos,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Break
func (s *Break) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	if pos, ok := temp["pos"]; ok {
		if err := json.Unmarshal(pos, &s.pos); err != nil {
			return err
		}
	}
	return nil
}

// Continue represents a continue statement in loops.
type Continue struct {
	pos Position // Source position
}

func (s *Continue) Position() Position { return s.pos }

// String returns a string representation of the continue statement.
func (s *Continue) String() string {
	return "continue"
}

// MarshalJSON implements custom JSON marshaling for Continue
func (s *Continue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"Type": "Continue",
		"pos":  s.pos,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Continue
func (s *Continue) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	if pos, ok := temp["pos"]; ok {
		if err := json.Unmarshal(pos, &s.pos); err != nil {
			return err
		}
	}
	return nil
}

// Import represents an import statement for importing .din files.
type Import struct {
	pos      Position // Source position
	Filename string   // The filename to import (as a string literal)
}

func (s *Import) Position() Position { return s.pos }

// String returns a string representation of the import statement.
func (s *Import) String() string {
	return fmt.Sprintf("import \"%s\"", s.Filename)
}

// NewAssign creates a new assignment statement
func NewAssign(pos Position, target Expression, value Expression, operator Token) *Assign {
	return &Assign{pos: pos, Target: target, Value: value, Operator: operator}
}

// NewIf creates a new if statement
func NewIf(pos Position, condition Expression, body Block, elseBlock Block) *If {
	return &If{pos: pos, Condition: condition, Body: body, Else: elseBlock}
}

// NewBinary creates a new binary expression
func NewBinary(pos Position, left Expression, operator Token, right Expression) *Binary {
	return &Binary{pos: pos, Left: left, Operator: operator, Right: right}
}

// NewLiteral creates a new literal expression
func NewLiteral(pos Position, value any) *Literal {
	return &Literal{pos: pos, Value: value}
}

// NewVariable creates a new variable reference
func NewVariable(pos Position, name string) *Variable {
	return &Variable{pos: pos, Name: name}
}

// NewCall creates a new function call expression
func NewCall(pos Position, function Expression, args []Expression) *Call {
	return &Call{pos: pos, Function: function, Arguments: args}
}

// NewTernary creates a new ternary expression
func NewTernary(pos Position, condition Expression, trueExpr Expression, falseExpr Expression) *Ternary {
	return &Ternary{pos: pos, Condition: condition, TrueExpr: trueExpr, FalseExpr: falseExpr}
}

// IsExpression checks if a node is an expression
func IsExpression(node any) bool {
	_, ok := node.(Expression)
	return ok
}

// IsStatement checks if a node is a statement
func IsStatement(node any) bool {
	_, ok := node.(Statement)
	return ok
}

// AsExpression safely converts to Expression
func AsExpression(node any) (Expression, bool) {
	expr, ok := node.(Expression)
	return expr, ok
}

// AsStatement safely converts to Statement
func AsStatement(node any) (Statement, bool) {
	stmt, ok := node.(Statement)
	return stmt, ok
}

// UnmarshalJSON implements custom JSON unmarshaling for Block
func (b *Block) UnmarshalJSON(data []byte) error {
	var rawStatements []json.RawMessage
	if err := json.Unmarshal(data, &rawStatements); err != nil {
		return err
	}

	*b = make(Block, len(rawStatements))
	for i, raw := range rawStatements {
		stmt, err := unmarshalStatement(raw)
		if err != nil {
			return err
		}
		(*b)[i] = stmt
	}
	return nil
}

// unmarshalStatement unmarshals a JSON message into a Statement
func unmarshalStatement(data []byte) (Statement, error) {
	// Try to determine the statement type by checking for specific fields
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	// Check for assignment statement (has Target, Value, Operator)
	if _, hasTarget := temp["Target"]; hasTarget {
		if _, hasValue := temp["Value"]; hasValue {
			if _, hasOperator := temp["Operator"]; hasOperator {
				var assign Assign
				if err := json.Unmarshal(data, &assign); err != nil {
					return nil, err
				}
				return &assign, nil
			}
		}
	}

	// Check for expression statement (has Expression)
	if _, hasExpression := temp["Expression"]; hasExpression {
		var exprStmt ExpressionStatement
		if err := json.Unmarshal(data, &exprStmt); err != nil {
			return nil, err
		}
		return &exprStmt, nil
	}

	// Check for if statement (has Condition, Body, and potentially Else)
	if _, hasCondition := temp["Condition"]; hasCondition {
		if _, hasBody := temp["Body"]; hasBody {
			// Check if it has Else field - this distinguishes If from While
			if _, hasElse := temp["Else"]; hasElse {
				var ifStmt If
				if err := json.Unmarshal(data, &ifStmt); err != nil {
					return nil, err
				}
				return &ifStmt, nil
			}
		}
	}

	// Check for while statement (has Condition and Body, but no Else)
	if _, hasCondition := temp["Condition"]; hasCondition {
		if _, hasBody := temp["Body"]; hasBody {
			// Make sure it doesn't have Else field
			if _, hasElse := temp["Else"]; !hasElse {
				var whileStmt While
				if err := json.Unmarshal(data, &whileStmt); err != nil {
					return nil, err
				}
				return &whileStmt, nil
			}
		}
	}

	// Check for for statement (has Name, Iterable)
	if _, hasName := temp["Name"]; hasName {
		if _, hasIterable := temp["Iterable"]; hasIterable {
			var forStmt For
			if err := json.Unmarshal(data, &forStmt); err != nil {
				return nil, err
			}
			return &forStmt, nil
		}
	}

	// Check for function definition (has Name, Parameters)
	if _, hasName := temp["Name"]; hasName {
		if _, hasParameters := temp["Parameters"]; hasParameters {
			var funcDef FunctionDefinition
			if err := json.Unmarshal(data, &funcDef); err != nil {
				return nil, err
			}
			return &funcDef, nil
		}
	}

	// Check for return statement (has Result)
	if _, hasResult := temp["Result"]; hasResult {
		var returnStmt Return
		if err := json.Unmarshal(data, &returnStmt); err != nil {
			return nil, err
		}
		return &returnStmt, nil
	}

	// Check for try-catch statement
	if _, hasTryBlock := temp["TryBlock"]; hasTryBlock {
		var tryCatch TryCatch
		if err := json.Unmarshal(data, &tryCatch); err != nil {
			return nil, err
		}
		return &tryCatch, nil
	}

	// Check for import statement (has Filename)
	if _, hasFilename := temp["Filename"]; hasFilename {
		var importStmt Import
		if err := json.Unmarshal(data, &importStmt); err != nil {
			return nil, err
		}
		return &importStmt, nil
	}

	// Check for break and continue statements using Type field
	if typeField, ok := temp["Type"]; ok {
		var typeStr string
		if err := json.Unmarshal(typeField, &typeStr); err == nil {
			switch typeStr {
			case "Break":
				var breakStmt Break
				if err := json.Unmarshal(data, &breakStmt); err == nil {
					return &breakStmt, nil
				}
			case "Continue":
				var continueStmt Continue
				if err := json.Unmarshal(data, &continueStmt); err == nil {
					return &continueStmt, nil
				}
			}
		}
	}

	// Fallback for old format without Type field
	if len(temp) == 0 || (len(temp) == 1 && temp["pos"] != nil) {
		// This might be a break or continue statement (old format)
		var breakStmt Break
		if err := json.Unmarshal(data, &breakStmt); err == nil {
			return &breakStmt, nil
		}
		var continueStmt Continue
		if err := json.Unmarshal(data, &continueStmt); err == nil {
			return &continueStmt, nil
		}
	}

	return nil, fmt.Errorf("unknown statement type")
}

// unmarshalExpression unmarshals a JSON message into an Expression
func unmarshalExpression(data []byte) (Expression, error) {
	// Try to determine the expression type by checking for specific fields
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	// Check for binary expression (has Left, Operator, Right)
	if _, hasLeft := temp["Left"]; hasLeft {
		if _, hasOperator := temp["Operator"]; hasOperator {
			if _, hasRight := temp["Right"]; hasRight {
				var binary Binary
				if err := json.Unmarshal(data, &binary); err != nil {
					return nil, err
				}
				return &binary, nil
			}
		}
	}

	// Check for unary expression (has Operator, Operand)
	if _, hasOperator := temp["Operator"]; hasOperator {
		if _, hasOperand := temp["Operand"]; hasOperand {
			var unary Unary
			if err := json.Unmarshal(data, &unary); err != nil {
				return nil, err
			}
			return &unary, nil
		}
	}

	// Check for call expression (has Function, Arguments)
	if _, hasFunction := temp["Function"]; hasFunction {
		if _, hasArguments := temp["Arguments"]; hasArguments {
			var call Call
			if err := json.Unmarshal(data, &call); err != nil {
				return nil, err
			}
			return &call, nil
		}
	}

	// Check for literal expression (has Value)
	if _, hasValue := temp["Value"]; hasValue {
		var literal Literal
		if err := json.Unmarshal(data, &literal); err != nil {
			return nil, err
		}
		return &literal, nil
	}

	// Check for variable expression (has Name)
	if _, hasName := temp["Name"]; hasName {
		var variable Variable
		if err := json.Unmarshal(data, &variable); err != nil {
			return nil, err
		}
		return &variable, nil
	}

	// Check for list expression (has Values)
	if _, hasValues := temp["Values"]; hasValues {
		var list List
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		return &list, nil
	}

	// Check for map expression (has Items)
	if _, hasItems := temp["Items"]; hasItems {
		var mapExpr Map
		if err := json.Unmarshal(data, &mapExpr); err != nil {
			return nil, err
		}
		return &mapExpr, nil
	}

	// Check for function expression (has Parameters, Body)
	if _, hasParameters := temp["Parameters"]; hasParameters {
		if _, hasBody := temp["Body"]; hasBody {
			var funcExpr FunctionExpression
			if err := json.Unmarshal(data, &funcExpr); err != nil {
				return nil, err
			}
			return &funcExpr, nil
		}
	}

	// Check for subscript expression (has Container, Subscript)
	if _, hasContainer := temp["Container"]; hasContainer {
		if _, hasSubscript := temp["Subscript"]; hasSubscript {
			var subscript Subscript
			if err := json.Unmarshal(data, &subscript); err != nil {
				return nil, err
			}
			return &subscript, nil
		}
	}

	// Check for ternary expression (has Condition, TrueExpr, FalseExpr)
	if _, hasCondition := temp["Condition"]; hasCondition {
		if _, hasTrueExpr := temp["TrueExpr"]; hasTrueExpr {
			if _, hasFalseExpr := temp["FalseExpr"]; hasFalseExpr {
				var ternary Ternary
				if err := json.Unmarshal(data, &ternary); err != nil {
					return nil, err
				}
				return &ternary, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown expression type")
}
