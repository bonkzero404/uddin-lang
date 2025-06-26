package interpreter

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string // String representation of the expected expression type
	}{
		{"5", "Literal"},
		{"3.14", "Literal"},
		{"\"hello\"", "Literal"},
		{"true", "Literal"},
		{"false", "Literal"},
		{"null", "Literal"},
		{"x", "Variable"},
		{"a + b", "Binary"},
		{"a - b", "Binary"},
		{"a * b", "Binary"},
		{"a / b", "Binary"},
		{"a % b", "Binary"},
		{"a == b", "Binary"},
		{"a != b", "Binary"},
		{"a < b", "Binary"},
		{"a <= b", "Binary"},
		{"a > b", "Binary"},
		{"a >= b", "Binary"},
		{"a and b", "Binary"},
		{"a or b", "Binary"},
		{"not a", "Unary"},
		{"a in b", "Binary"},
		{"a + b * c", "Binary"},
		{"(a + b) * c", "Binary"},
		{"a + (b + c)", "Binary"},
		{"a[0]", "Subscript"},
		{"a[i + 1]", "Subscript"},
		{"f()", "Call"},
		{"f(a, b)", "Call"},
		{"f(a, b + c)", "Call"},
		{"[1, 2, 3]", "List"},
		{"{\"a\": 1, \"b\": 2}", "Map"},
		{"fun(x): return x * x end", "FunctionExpression"},
	}

	for i, test := range tests {
		expr, err := ParseExpression([]byte(test.input))
		if err != nil {
			t.Errorf("Test %d: unexpected error: %v", i, err)
			continue
		}

		// Check the type of expression
		typeName := fmt.Sprintf("%T", expr)
		// Extract just the type name without the package prefix
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			typeName = typeName[idx+1:]
		}
		// Remove pointer prefix if present
		typeName = strings.TrimPrefix(typeName, "*")

		if typeName != test.expected {
			t.Errorf("Test %d: expected type %q, got %q", i, test.expected, typeName)
		}
	}
}

func TestParseProgram(t *testing.T) {
	tests := []struct {
		input    string
		expected int // Expected number of statements
	}{
		{"", 0},
		{"x = 5", 1},
		{"x = 5\ny = 10", 2},
		{"if (x > 5) then: y = 10 end", 1},
		{"while (x > 0): x = x - 1 end", 1},
		{"for (i in range(10)): print(i) end", 1},
		{"fun add(a, b): return a + b end", 1},
		{"try: x = 1/0 catch (err): print(err) end", 1},
		{`
		x = 5
		y = 10
		z = x + y
		print(z)
		`, 4},
	}

	for i, test := range tests {
		prog, err := ParseProgram([]byte(test.input))
		if err != nil {
			t.Errorf("Test %d: unexpected error: %v", i, err)
			continue
		}

		if len(prog.Statements) != test.expected {
			t.Errorf("Test %d: expected %d statements, got %d", i, test.expected, len(prog.Statements))
		}
	}
}

func TestParseIfStatement(t *testing.T) {
	tests := []struct {
		input          string
		expectCondType string
		expectBodyLen  int
		expectElseLen  int
	}{
		{"if (x > 5) then: y = 10 end", "Binary", 1, 0},
		{"if (x > 5) then: y = 10 else: y = 5 end", "Binary", 1, 1},
		{"if (x > 5) then: y = 10 z = 20 end", "Binary", 2, 0},
		{"if (x > 5) then: y = 10 else if (x > 0) then: y = 5 else: y = 0 end", "Binary", 1, 1},
	}

	for i, test := range tests {
		prog, err := ParseProgram([]byte(test.input))
		if err != nil {
			t.Errorf("Test %d: unexpected error: %v", i, err)
			continue
		}

		if len(prog.Statements) != 1 {
			t.Errorf("Test %d: expected 1 statement, got %d", i, len(prog.Statements))
			continue
		}

		ifStmt, ok := prog.Statements[0].(*If)
		if !ok {
			t.Errorf("Test %d: expected If statement, got %T", i, prog.Statements[0])
			continue
		}

		// Check condition type
		condType := fmt.Sprintf("%T", ifStmt.Condition)
		if idx := strings.LastIndex(condType, "."); idx >= 0 {
			condType = condType[idx+1:]
		}
		condType = strings.TrimPrefix(condType, "*")

		if condType != test.expectCondType {
			t.Errorf("Test %d: expected condition type %q, got %q", i, test.expectCondType, condType)
		}

		if len(ifStmt.Body) != test.expectBodyLen {
			t.Errorf("Test %d: expected %d body statements, got %d", i, test.expectBodyLen, len(ifStmt.Body))
		}

		if len(ifStmt.Else) != test.expectElseLen {
			t.Errorf("Test %d: expected %d else statements, got %d", i, test.expectElseLen, len(ifStmt.Else))
		}
	}
}

func TestParseWhileStatement(t *testing.T) {
	input := "while (x > 0): x = x - 1 end"
	prog, err := ParseProgram([]byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Statements))
	}

	whileStmt, ok := prog.Statements[0].(*While)
	if !ok {
		t.Fatalf("Expected While statement, got %T", prog.Statements[0])
	}

	// Check condition type
	condType := fmt.Sprintf("%T", whileStmt.Condition)
	if idx := strings.LastIndex(condType, "."); idx >= 0 {
		condType = condType[idx+1:]
	}
	condType = strings.TrimPrefix(condType, "*")

	if condType != "Binary" {
		t.Errorf("Expected condition type Binary, got %q", condType)
	}

	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestParseForStatement(t *testing.T) {
	input := "for (i in range(10)): print(i) end"
	prog, err := ParseProgram([]byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Statements))
	}

	forStmt, ok := prog.Statements[0].(*For)
	if !ok {
		t.Fatalf("Expected For statement, got %T", prog.Statements[0])
	}

	if forStmt.Name != "i" {
		t.Errorf("Expected variable name 'i', got %q", forStmt.Name)
	}

	// Check iterable type
	iterableType := fmt.Sprintf("%T", forStmt.Iterable)
	if idx := strings.LastIndex(iterableType, "."); idx >= 0 {
		iterableType = iterableType[idx+1:]
	}
	iterableType = strings.TrimPrefix(iterableType, "*")

	if iterableType != "Call" {
		t.Errorf("Expected iterable type Call, got %q", iterableType)
	}

	if len(forStmt.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(forStmt.Body))
	}
}

func TestParseFunctionDefinition(t *testing.T) {
	input := "fun add(a, b): return a + b end"
	prog, err := ParseProgram([]byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Statements))
	}

	funDef, ok := prog.Statements[0].(*FunctionDefinition)
	if !ok {
		t.Fatalf("Expected FunctionDefinition, got %T", prog.Statements[0])
	}

	if funDef.Name != "add" {
		t.Errorf("Expected function name 'add', got %q", funDef.Name)
	}

	expectedParams := []string{"a", "b"}
	if !reflect.DeepEqual(funDef.Parameters, expectedParams) {
		t.Errorf("Expected parameters %v, got %v", expectedParams, funDef.Parameters)
	}

	if len(funDef.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(funDef.Body))
	}
}

func TestParseTryCatchStatement(t *testing.T) {
	input := "try: x = 1/0 catch (err): print(err) end"
	prog, err := ParseProgram([]byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prog.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Statements))
	}

	tryCatch, ok := prog.Statements[0].(*TryCatch)
	if !ok {
		t.Fatalf("Expected TryCatch, got %T", prog.Statements[0])
	}

	if tryCatch.ErrVar != "err" {
		t.Errorf("Expected error variable 'err', got %q", tryCatch.ErrVar)
	}

	if len(tryCatch.TryBlock) != 1 {
		t.Errorf("Expected 1 try block statement, got %d", len(tryCatch.TryBlock))
	}

	if len(tryCatch.CatchBlock) != 1 {
		t.Errorf("Expected 1 catch block statement, got %d", len(tryCatch.CatchBlock))
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
		errorMsg    string
	}{
		{"x = ", true, "parse error at 1:5: expected expression, not EOF"},                              // Incomplete assignment
		{"if (x > 5) then: y = 10 end", false, ""},                                                     // Complete if statement - should not error
		{"if (x > 5) y = 10", true, "parse error at 1:12: expected then and not name"},                 // Missing 'then'
		{"while (x > 0) x = x-1", true, "parse error at 1:15: expected { or :, not name"},             // Missing ':'
		{"for (i in) print(i)", true, "parse error at 1:10: expected expression, not )"},              // Invalid for loop
		{"fun add(a, b) return a+b", true, "parse error at 1:15: expected { or :, not return"},        // Missing ':'
		{"fun add(a, b): return a+b end", false, ""},                                                    // Complete function - should not error
		{"x + ", true, "parse error at 1:5: expected expression, not EOF"},                             // Incomplete expression
		{"(x + y", true, "parse error at 1:7: expected ) and not EOF"},                                // Unclosed parenthesis
		{"[1, 2, 3", true, "parse error at 1:9: expected ] and not EOF"},                              // Unclosed bracket
		{"\"a\": 1,", true, "parse error at 1:4: expected expression, not :"},                         // Invalid expression
		{"try:", true, "parse error at 1:5: expected catch and not EOF"},                              // Incomplete try block
		{"try: x = 1 catch:", true, "parse error at 1:17: expected ( and not :"},                      // Missing catch variable
		{"throw x", false, ""},                                                                         // Complete throw - should not error
		{"try: x = 1 catch (err) print(err)", true, "parse error at 1:24: expected { or :, not name"}, // Missing ':' in catch block
		{"try: x = 1 catch (err): print(err) end", false, ""},                                          // Complete try-catch - should not error
		{"if (x > 5) then: y = 10 end", false, ""},                                                     // Complete if - should not error
		{"fun add(a, b): return", true, "parse error at 1:22: expected expression, not EOF"},          // Missing return value
		{"try: x = 1", true, "parse error at 1:11: expected catch and not EOF"},                       // Missing catch block
		{"try: x = 1 catch (err): print(err) end", false, ""},                                          // Complete try-catch - should not error
	}

	for i, test := range tests {
		_, err := ParseProgram([]byte(test.input))

		if test.expectError {
			if err == nil {
				t.Errorf("Test %d: expected error, got nil", i)
			} else if !strings.Contains(err.Error(), test.errorMsg) {
				t.Errorf("Test %d: expected error message containing '%s', got '%s'", i, test.errorMsg, err.Error())
			}
		} else if !test.expectError && err != nil {
			t.Errorf("Test %d: unexpected error: %v", i, err)
		}
	}
}
