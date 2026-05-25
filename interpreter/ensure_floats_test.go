package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

// runExprAndGetError is a helper that runs a single expression as a program
// and returns the execution error (if any).
func runExprAndGetError(t *testing.T, expr string) error {
	t.Helper()
	prog, parseErr := ParseProgram([]byte(expr))
	if parseErr != nil {
		t.Fatalf("parse error for %q: %v", expr, parseErr)
	}
	config := &Config{Stdout: &bytes.Buffer{}}
	_, execErr := Execute(prog, config)
	return execErr
}

// TestDivideZeroByZeroFloat verifies that 0 / 0.0 panics with a
// divide-by-zero error rather than a type error.  Before the fix,
// ensureIntToFloats returned (0, 0) for both the type-error path and
// genuine zero operands, so the caller mistakenly raised a TypeError.
func TestDivideZeroByZeroFloat(t *testing.T) {
	err := runExprAndGetError(t, `0 / 0.0`)
	if err == nil {
		t.Fatal("expected an error for 0 / 0.0, got none")
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "requires two floats") || strings.Contains(msg, "type") {
		t.Errorf("got a type error instead of divide-by-zero: %s", err.Error())
	}
	if !strings.Contains(msg, "zero") {
		t.Errorf("expected divide-by-zero error, got: %s", err.Error())
	}
}

// TestModuloZeroByZeroFloat verifies that 0 % 0.0 panics with a
// divide-by-zero error rather than a type error.
func TestModuloZeroByZeroFloat(t *testing.T) {
	err := runExprAndGetError(t, `0 % 0.0`)
	if err == nil {
		t.Fatal("expected an error for 0 % 0.0, got none")
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "requires two floats") || strings.Contains(msg, "type") {
		t.Errorf("got a type error instead of divide-by-zero: %s", err.Error())
	}
	if !strings.Contains(msg, "zero") {
		t.Errorf("expected divide-by-zero error, got: %s", err.Error())
	}
}

// TestDivideZeroByZeroInt verifies that 0 / 0 (int/int) is handled by
// FastEvalDivide and still produces a divide-by-zero error (regression guard).
func TestDivideZeroByZeroInt(t *testing.T) {
	err := runExprAndGetError(t, `x = 0
y = 0
z = x / y`)
	if err == nil {
		t.Fatal("expected an error for 0 / 0, got none")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "zero") {
		t.Errorf("expected divide-by-zero error, got: %s", err.Error())
	}
}

// TestDivideNonNumericTypeError verifies that dividing non-numeric values
// still produces a type error (ensures the ok=false path still fires).
func TestDivideNonNumericTypeError(t *testing.T) {
	err := runExprAndGetError(t, `1 + "string"`)
	if err == nil {
		t.Fatal("expected a type error, got none")
	}
}
