package interpreter

import (
	"fmt"
	"strings"
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
		lines = append(lines, stmt.String())
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
	pos    Position   // Source position
	Target Expression // Left-hand side of the assignment
	Value  Expression // Right-hand side of the assignment
}

func (s *Assign) Position() Position { return s.pos }

// String returns a string representation of the assignment.
func (s *Assign) String() string {
	return fmt.Sprintf("%s = %s", s.Target, s.Value)
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
func (s *If) String() string {
	str := fmt.Sprintf("if %s {\n%s\n}", s.Condition, indent(s.Body.String()))
	if len(s.Else) > 0 {
		str += fmt.Sprintf(" else {\n%s\n}", indent(s.Else.String()))
	}
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
func (s *While) String() string {
	return fmt.Sprintf("while %s {\n%s\n}", s.Condition, indent(s.Body.String()))
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
func (s *For) String() string {
	return fmt.Sprintf("for %s in %s {\n%s\n}", s.Name, s.Iterable, indent(s.Body.String()))
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
	return fmt.Sprintf("try:\n%s\nend catch %s:\n%s\nend",
		indent(s.TryBlock.String()), s.ErrVar, indent(s.CatchBlock.String()))
}

// Return represents a return statement.
type Return struct {
	pos    Position   // Source position
	Result Expression // The value to return
}

func (s *Return) Position() Position { return s.pos }

// String returns a string representation of the return statement.
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
		bodyStr = "\n" + indent(s.Body.String()) + "\n"
	}
	return fmt.Sprintf("fun %s(%s%s) {%s}",
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
func (e *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Operator, e.Right)
}

// Unary represents a unary operation (operator operand).
type Unary struct {
	pos      Position   // Source position
	Operator Token      // Operator token
	Operand  Expression // The operand
}

func (e *Unary) Position() Position { return e.pos }

// String returns a string representation of the unary expression.
func (e *Unary) String() string {
	space := ""
	// Special case for NOT operator to improve readability
	if e.Operator == NOT {
		space = " "
	}
	return fmt.Sprintf("(%s%s%s)", e.Operator, space, e.Operand)
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
		return "nil"
	}
	if s, ok := e.Value.(string); ok {
		return fmt.Sprintf("%q", s)
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
func (e *Map) String() string {
	items := []string{}
	for _, item := range e.Items {
		items = append(items, fmt.Sprintf("%s: %s", item.Key, item.Value))
	}
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
	return fmt.Sprintf("fun(%s%s) {%s}", strings.Join(e.Parameters, ", "), ellipsisStr, bodyStr)
}

// Subscript represents a container subscript expression (container[index]).
type Subscript struct {
	pos       Position   // Source position
	Container Expression // The container to index into (list, map, etc.)
	Subscript Expression // The index expression
}

func (e *Subscript) Position() Position { return e.pos }

// String returns a string representation of the subscript expression.
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

// Ternary represents a ternary conditional expression (condition ? trueExpr : falseExpr).
type Ternary struct {
	pos       Position
	Condition Expression
	TrueExpr  Expression
	FalseExpr Expression
}

func (e *Ternary) Position() Position { return e.pos }

// String returns a string representation of the ternary expression.
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

// Continue represents a continue statement in loops.
type Continue struct {
	pos Position // Source position
}

func (s *Continue) Position() Position { return s.pos }

// String returns a string representation of the continue statement.
func (s *Continue) String() string {
	return "continue"
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
func NewAssign(pos Position, target Expression, value Expression) *Assign {
	return &Assign{pos: pos, Target: target, Value: value}
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
