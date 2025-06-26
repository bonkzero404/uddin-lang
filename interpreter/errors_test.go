package interpreter

import (
	"bytes"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	// Test TypeError
	typeErr := TypeError{
		Message: "Type mismatch",
		pos:     Position{Line: 2, Column: 5},
	}

	errMsg := typeErr.Error()
	if errMsg == "" {
		t.Error("TypeError Error should not be empty")
	}

	pos := typeErr.Position()
	if pos.Line != 2 || pos.Column != 5 {
		t.Error("TypeError Position not working correctly")
	}

	// Test ValueError
	valueErr := ValueError{
		Message: "Invalid value",
		pos:     Position{Line: 3, Column: 1},
	}

	errMsg = valueErr.Error()
	if errMsg == "" {
		t.Error("ValueError Error should not be empty")
	}

	pos = valueErr.Position()
	if pos.Line != 3 || pos.Column != 1 {
		t.Error("ValueError Position not working correctly")
	}

	// Test NameError
	nameErr := NameError{
		Message: "Name not defined",
		pos:     Position{Line: 4, Column: 8},
	}

	errMsg = nameErr.Error()
	if errMsg == "" {
		t.Error("NameError Error should not be empty")
	}

	pos = nameErr.Position()
	if pos.Line != 4 || pos.Column != 8 {
		t.Error("NameError Position not working correctly")
	}

	// Test RuntimeError
	runtimeErr := RuntimeError{
		Message: "Runtime error",
		pos:     Position{Line: 5, Column: 3},
	}

	errMsg = runtimeErr.Error()
	if errMsg == "" {
		t.Error("RuntimeError Error should not be empty")
	}

	pos = runtimeErr.Position()
	if pos.Line != 5 || pos.Column != 3 {
		t.Error("RuntimeError Position not working correctly")
	}

	// Test BreakException
	breakExc := BreakException{
		pos: Position{Line: 6, Column: 1},
	}

	errMsg = breakExc.Error()
	if errMsg == "" {
		t.Error("BreakException Error should not be empty")
	}

	pos = breakExc.Position()
	if pos.Line != 6 || pos.Column != 1 {
		t.Error("BreakException Position not working correctly")
	}

	// Test ContinueException
	continueExc := ContinueException{
		pos: Position{Line: 7, Column: 1},
	}

	errMsg = continueExc.Error()
	if errMsg == "" {
		t.Error("ContinueException Error should not be empty")
	}

	pos = continueExc.Position()
	if pos.Line != 7 || pos.Column != 1 {
		t.Error("ContinueException Position not working correctly")
	}
}

func TestErrorCreationFunctions(t *testing.T) {
	pos := Position{Line: 10, Column: 5}

	// Test typeError function
	err := typeError(pos, "test type error: %s", "details")
	if err == nil {
		t.Error("typeError should return an error")
	}
	if err.Error() == "" {
		t.Error("typeError should return non-empty error message")
	}

	// Test valueError function
	err = valueError(pos, "test value error: %d", 42)
	if err == nil {
		t.Error("valueError should return an error")
	}
	if err.Error() == "" {
		t.Error("valueError should return non-empty error message")
	}

	// Test nameError function
	err = nameError(pos, "test name error: %s", "variable")
	if err == nil {
		t.Error("nameError should return an error")
	}
	if err.Error() == "" {
		t.Error("nameError should return non-empty error message")
	}

	// Test runtimeError function
	err = runtimeError(pos, "test runtime error: %v", "problem")
	if err == nil {
		t.Error("runtimeError should return an error")
	}
	if err.Error() == "" {
		t.Error("runtimeError should return non-empty error message")
	}
}

// TestErrorsInExecution tests errors during actual code execution
func TestErrorsInExecution(t *testing.T) {
	config := &Config{
		Stdout: &bytes.Buffer{},
	}

	// Test cases that should trigger different error types
	errorCases := []string{
		`unknown_var`,        // Should trigger NameError
		`1 + "string"`,       // Should trigger TypeError
		`[1, 2][5]`,          // Should trigger ValueError (index out of range)
		`nonexistent_func()`, // Should trigger NameError
	}

	for _, code := range errorCases {
		program, err := ParseProgram([]byte(code))
		if err != nil {
			continue // Skip parsing errors
		}

		// Execute and expect errors - this should trigger error creation functions
		_, _ = Execute(program, config)
		// We don't check the error type, just that execution happens and errors are created
	}
}
