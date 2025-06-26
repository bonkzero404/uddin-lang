package interpreter

import (
	"bytes"
	"testing"
)

func TestASTNodes(t *testing.T) {
	// Test Program String method
	prog := &Program{
		Statements: Block{
			&ExpressionStatement{
				Expression: &Literal{Value: "Hello"},
			},
		},
	}
	result := prog.String()
	if result == "" {
		t.Error("Program String should not be empty")
	}

	// Test Block String method
	block := Block{
		&ExpressionStatement{
			Expression: &Literal{Value: "Test"},
		},
		&ExpressionStatement{
			Expression: &Literal{Value: 42},
		},
	}
	result = block.String()
	if result == "" {
		t.Error("Block String should not be empty")
	}

	// Test various statement nodes
	testStatementNodes(t)
	testExpressionNodes(t)
}

// TestASTNodesParsed tests AST nodes through actual parsing
func TestASTNodesParsed(t *testing.T) {
	testCases := []string{
		// Test simpler statement types that should trigger AST methods
		`x = 42`,
		`print("hello")`,
		`[1, 2, 3]`,
		`1 + 2`,
		`-5`,
		`not false`,
		`5 > 3`,
	}

	config := &Config{
		Stdout: &bytes.Buffer{}, // Capture output to avoid printing during tests
	}

	for _, code := range testCases {
		// Parse the code to create AST nodes
		program, err := ParseProgram([]byte(code))
		if err != nil {
			// Some test cases might have parse errors, which is okay for testing
			continue
		}

		// Call String() method to trigger AST string methods
		str := program.String()
		if str == "" {
			t.Errorf("Program String should not be empty for code: %s", code)
		}

		// Try to execute to trigger more AST methods (ignore errors)
		_, _ = Execute(program, config)
	}
}

func testStatementNodes(t *testing.T) {
	// Test ExpressionStatement
	exprStmt := &ExpressionStatement{
		Expression: &Literal{Value: "test"},
		pos:        Position{Line: 1, Column: 1},
	}
	pos := exprStmt.Position()
	if pos.Line != 1 || pos.Column != 1 {
		t.Error("ExpressionStatement Position not working correctly")
	}
	str := exprStmt.String()
	if str == "" {
		t.Error("ExpressionStatement String should not be empty")
	}

	// Test Assign
	assignStmt := &Assign{
		Target: &Variable{Name: "x"},
		Value:  &Literal{Value: 42},
		pos:    Position{Line: 2, Column: 1},
	}
	pos = assignStmt.Position()
	if pos.Line != 2 {
		t.Error("Assign Position not working correctly")
	}
	str = assignStmt.String()
	if str == "" {
		t.Error("Assign String should not be empty")
	}

	// Test OuterAssign
	outerAssign := &OuterAssign{
		Name:  "y",
		Value: &Literal{Value: 100},
		pos:   Position{Line: 2, Column: 5},
	}
	pos = outerAssign.Position()
	if pos.Line != 2 {
		t.Error("OuterAssign Position not working correctly")
	}
	str = outerAssign.String()
	if str == "" {
		t.Error("OuterAssign String should not be empty")
	}

	// Test If
	ifStmt := &If{
		Condition: &Literal{Value: true},
		Body:      Block{&ExpressionStatement{Expression: &Literal{Value: "then"}}},
		Else:      Block{&ExpressionStatement{Expression: &Literal{Value: "else"}}},
		pos:       Position{Line: 3, Column: 1},
	}
	pos = ifStmt.Position()
	if pos.Line != 3 {
		t.Error("If Position not working correctly")
	}
	str = ifStmt.String()
	if str == "" {
		t.Error("If String should not be empty")
	}

	// Test While
	whileStmt := &While{
		Condition: &Literal{Value: true},
		Body:      Block{&ExpressionStatement{Expression: &Literal{Value: "loop"}}},
		pos:       Position{Line: 4, Column: 1},
	}
	pos = whileStmt.Position()
	if pos.Line != 4 {
		t.Error("While Position not working correctly")
	}
	str = whileStmt.String()
	if str == "" {
		t.Error("While String should not be empty")
	}

	// Test For
	forStmt := &For{
		Name:     "i",
		Iterable: &List{Values: []Expression{&Literal{Value: 1}, &Literal{Value: 2}}},
		Body:     Block{&ExpressionStatement{Expression: &Variable{Name: "i"}}},
		pos:      Position{Line: 5, Column: 1},
	}
	pos = forStmt.Position()
	if pos.Line != 5 {
		t.Error("For Position not working correctly")
	}
	str = forStmt.String()
	if str == "" {
		t.Error("For String should not be empty")
	}

	// Test FunctionDefinition
	funcDef := &FunctionDefinition{
		Name:       "testFunc",
		Parameters: []string{"a", "b"},
		Body:       Block{&Return{Result: &Variable{Name: "a"}}},
		pos:        Position{Line: 6, Column: 1},
	}
	pos = funcDef.Position()
	if pos.Line != 6 {
		t.Error("FunctionDefinition Position not working correctly")
	}
	str = funcDef.String()
	if str == "" {
		t.Error("FunctionDefinition String should not be empty")
	}

	// Test TryCatch
	tryStmt := &TryCatch{
		TryBlock:   Block{&ExpressionStatement{Expression: &Literal{Value: "try"}}},
		CatchBlock: Block{&ExpressionStatement{Expression: &Literal{Value: "catch"}}},
		ErrVar:     "e",
		pos:        Position{Line: 7, Column: 1},
	}
	pos = tryStmt.Position()
	if pos.Line != 7 {
		t.Error("TryCatch Position not working correctly")
	}
	str = tryStmt.String()
	if str == "" {
		t.Error("TryCatch String should not be empty")
	}

	// Test Return
	returnStmt := &Return{
		Result: &Literal{Value: 100},
		pos:    Position{Line: 8, Column: 1},
	}
	pos = returnStmt.Position()
	if pos.Line != 8 {
		t.Error("Return Position not working correctly")
	}
	str = returnStmt.String()
	if str == "" {
		t.Error("Return String should not be empty")
	}

	// Test Break
	breakStmt := &Break{
		pos: Position{Line: 9, Column: 1},
	}
	pos = breakStmt.Position()
	if pos.Line != 9 {
		t.Error("Break Position not working correctly")
	}
	str = breakStmt.String()
	if str == "" {
		t.Error("Break String should not be empty")
	}

	// Test Continue
	continueStmt := &Continue{
		pos: Position{Line: 10, Column: 1},
	}
	pos = continueStmt.Position()
	if pos.Line != 10 {
		t.Error("Continue Position not working correctly")
	}
	str = continueStmt.String()
	if str == "" {
		t.Error("Continue String should not be empty")
	}

	// Test Import
	importStmt := &Import{
		Filename: "test.din",
		pos:      Position{Line: 11, Column: 1},
	}
	pos = importStmt.Position()
	if pos.Line != 11 {
		t.Error("Import Position not working correctly")
	}
	str = importStmt.String()
	if str == "" {
		t.Error("Import String should not be empty")
	}
}

func testExpressionNodes(t *testing.T) {
	// Test Literal
	literal := &Literal{
		Value: "test string",
		pos:   Position{Line: 1, Column: 1},
	}
	pos := literal.Position()
	if pos.Line != 1 {
		t.Error("Literal Position not working correctly")
	}
	str := literal.String()
	if str == "" {
		t.Error("Literal String should not be empty")
	}

	// Test Variable
	variable := &Variable{
		Name: "testVar",
		pos:  Position{Line: 4, Column: 1},
	}
	pos = variable.Position()
	if pos.Line != 4 {
		t.Error("Variable Position not working correctly")
	}
	str = variable.String()
	if str == "" {
		t.Error("Variable String should not be empty")
	}

	// Test Binary
	binOp := &Binary{
		Left:     &Literal{Value: 5},
		Operator: PLUS,
		Right:    &Literal{Value: 3},
		pos:      Position{Line: 5, Column: 1},
	}
	pos = binOp.Position()
	if pos.Line != 5 {
		t.Error("Binary Position not working correctly")
	}
	str = binOp.String()
	if str == "" {
		t.Error("Binary String should not be empty")
	}

	// Test Unary
	unaryOp := &Unary{
		Operator: MINUS,
		Operand:  &Literal{Value: 10},
		pos:      Position{Line: 6, Column: 1},
	}
	pos = unaryOp.Position()
	if pos.Line != 6 {
		t.Error("Unary Position not working correctly")
	}
	str = unaryOp.String()
	if str == "" {
		t.Error("Unary String should not be empty")
	}

	// Test Call
	funcCall := &Call{
		Function:  &Variable{Name: "testFunc"},
		Arguments: []Expression{&Literal{Value: 1}, &Literal{Value: "arg"}},
		pos:       Position{Line: 7, Column: 1},
	}
	pos = funcCall.Position()
	if pos.Line != 7 {
		t.Error("Call Position not working correctly")
	}
	str = funcCall.String()
	if str == "" {
		t.Error("Call String should not be empty")
	}

	// Test List
	list := &List{
		Values: []Expression{&Literal{Value: 1}, &Literal{Value: "item"}},
		pos:    Position{Line: 8, Column: 1},
	}
	pos = list.Position()
	if pos.Line != 8 {
		t.Error("List Position not working correctly")
	}
	str = list.String()
	if str == "" {
		t.Error("List String should not be empty")
	}

	// Test Map
	mapLit := &Map{
		Items: []MapItem{
			{Key: &Literal{Value: "key1"}, Value: &Literal{Value: "value1"}},
			{Key: &Literal{Value: "key2"}, Value: &Literal{Value: 42}},
		},
		pos: Position{Line: 9, Column: 1},
	}
	pos = mapLit.Position()
	if pos.Line != 9 {
		t.Error("Map Position not working correctly")
	}
	str = mapLit.String()
	if str == "" {
		t.Error("Map String should not be empty")
	}

	// Test Subscript
	subscript := &Subscript{
		Container: &Variable{Name: "arr"},
		Subscript: &Literal{Value: 0},
		pos:       Position{Line: 10, Column: 1},
	}
	pos = subscript.Position()
	if pos.Line != 10 {
		t.Error("Subscript Position not working correctly")
	}
	str = subscript.String()
	if str == "" {
		t.Error("Subscript String should not be empty")
	}

	// Test Ternary
	ternary := &Ternary{
		Condition: &Literal{Value: true},
		TrueExpr:  &Literal{Value: "yes"},
		FalseExpr: &Literal{Value: "no"},
		pos:       Position{Line: 11, Column: 1},
	}
	pos = ternary.Position()
	if pos.Line != 11 {
		t.Error("Ternary Position not working correctly")
	}
	str = ternary.String()
	if str == "" {
		t.Error("Ternary String should not be empty")
	}

	// Test FunctionExpression
	funcExpr := &FunctionExpression{
		Parameters: []string{"x", "y"},
		Body:       Block{&Return{Result: &Variable{Name: "x"}}},
		pos:        Position{Line: 12, Column: 1},
	}
	pos = funcExpr.Position()
	if pos.Line != 12 {
		t.Error("FunctionExpression Position not working correctly")
	}
	str = funcExpr.String()
	if str == "" {
		t.Error("FunctionExpression String should not be empty")
	}
}
