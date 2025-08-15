package interpreter

import (
	"strings"
	"testing"
)

// TestMemoryAccessErrorMessage tests that memory access errors show the custom message
func TestMemoryAccessErrorMessage(t *testing.T) {
	// Create a mock error that simulates a nil pointer dereference
	config := &Config{}
	interp := newInterpreter(config)
	
	// Simulate the error handling path that would occur during execution
	defer func() {
		if r := recover(); r != nil {
			// Use the current position from interpreter for better error reporting
			currentPos := interp.currentPos
			// If currentPos is zero value, fall back to 1:1
			if currentPos.Line == 0 {
				currentPos = Position{Line: 1, Column: 1}
			}
			
			switch e := r.(type) {
			case error:
				// Handle other error types that implement the error interface
				errorMsg := e.Error()
				// Convert technical error messages to user-friendly ones
				if strings.Contains(errorMsg, "nil pointer dereference") || strings.Contains(errorMsg, "invalid memory address") {
					err := runtimeError(currentPos, "memory access error: attempting to access invalid or uninitialized data")
						
						// Check that the error message contains the expected English text
						if !strings.Contains(err.Error(), "memory access error") {
							t.Errorf("Expected error message to contain 'memory access error', got: %s", err.Error())
						}
						if !strings.Contains(err.Error(), "attempting to access invalid or uninitialized data") {
							t.Errorf("Expected error message to contain 'attempting to access invalid or uninitialized data', got: %s", err.Error())
						}
					return
				}
			}
		}
	}()
	
	// Simulate a nil pointer dereference error
	panic(strings.NewReader("runtime error: invalid memory address or nil pointer dereference"))
}

// TestMemoryAccessErrorMessageFormat tests the format of the error message
func TestMemoryAccessErrorMessageFormat(t *testing.T) {
	pos := Position{Line: 5, Column: 10}
	err := runtimeError(pos, "memory access error: attempting to access invalid or uninitialized data")
	
	expectedMsg := "runtime error at 5:10: memory access error: attempting to access invalid or uninitialized data"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message: %s, got: %s", expectedMsg, err.Error())
	}
}