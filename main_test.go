package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"uddin-lang/interpreter"
)

func getSampleScriptForTest() string {
	content, err := os.ReadFile("examples/02_functional_demo.din")
	if err != nil {
		return `
		// Fungsi main sederhana untuk test
		fun main():
			print("Test script")
			print(42)
			print(true)
		end
		`
	}
	return string(content)
}

func TestSampleScript(t *testing.T) {
	// Get the sample script
	input := getSampleScriptForTest()

	// Run the program
	success, output := interpreter.RunProgram(input)

	// Check if the program executed successfully
	if !success {
		t.Errorf("Sample script execution failed: %s", output)
	}

	// Check if the output contains expected results
	expectedOutputs := []string{
		"FUNCTIONAL PROGRAMMING DEMO",
		"Factorial of 5 = 120",
		"Fibonacci at 7 = 13",
		"Age: 30 years",
		"Number array: 1, 2, 3, 4, 5",
		"Error successfully caught:",
		"Demo completed!",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it doesn't.\nOutput: %s", expected, output)
		}
	}
}

func TestRunProgramWithCustomOutput(t *testing.T) {
	// Simple test script
	script := `
	fun main():
		print("Hello, World!")
		x = 5 + 3
		print("Result: " + str(x))
	end
	`

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Parse the program
	prog, err := interpreter.ParseProgram([]byte(script))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	// Execute with custom output
	_, err = interpreter.Execute(prog, &interpreter.Config{
		Stdout: &buf,
	})

	if err != nil {
		t.Fatalf("Failed to execute program: %v", err)
	}

	// Check output - normalize line endings and spaces
	output := strings.ReplaceAll(buf.String(), "\n", "")
	output = strings.ReplaceAll(output, "\r", "")
	expected := "Hello, World!Result: 8"

	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestCLIPrintUsage(t *testing.T) {
	// Test that CLI printUsage doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CLI.printUsage() panicked: %v", r)
		}
	}()

	cli := NewCLI([]string{"uddinlang"})
	cli.printUsage()
	// If we get here without panic, the test passes
}

func TestCLIListExamples(t *testing.T) {
	// Test that CLI listExamples doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CLI.listExamples() panicked: %v", r)
		}
	}()

	cli := NewCLI([]string{"uddinlang"})
	err := cli.listExamples()
	if err != nil {
		// Error is expected if examples directory doesn't exist, that's ok
		t.Logf("listExamples returned error (expected if no examples dir): %v", err)
	}
}

func TestGetSampleScriptForTest(t *testing.T) {
	script := getSampleScriptForTest()
	if script == "" {
		t.Error("getSampleScriptForTest should return non-empty string")
	}

	// Should contain either actual script content or fallback content
	if !strings.Contains(script, "main") {
		t.Error("Script should contain 'main' function")
	}
}

// Test more main.go functionality
func TestMainFunctionality(t *testing.T) {
	// Create a temporary test file
	testContent := `
	fun main():
		print("Test from file")
		x = 10 + 5
		print("Result: " + str(x))
	end
	`

	// Test with a simple script content
	success, output := interpreter.RunProgram(testContent)
	if !success {
		t.Errorf("Test script should execute successfully, got: %s", output)
	}

	// Test that output contains expected content
	if !strings.Contains(output, "Test from file") {
		t.Error("Output should contain 'Test from file'")
	}
	if !strings.Contains(output, "Result: 15") {
		t.Error("Output should contain 'Result: 15'")
	}
}

func TestMainWithSyntaxCheck(t *testing.T) {
	// Test syntax analysis with valid code
	validCode := `
	fun main():
		print("Valid syntax")
	end
	`
	success, message := interpreter.SyntaxAnalyze(validCode)
	if !success {
		t.Errorf("Valid syntax should pass analysis, got: %s", message)
	}
	if !strings.Contains(message, "All syntax is correct") {
		t.Errorf("Should contain success message, got: %s", message)
	}

	// Test syntax analysis with invalid code
	invalidCode := `
	this is not valid syntax %%%
	`
	success, message = interpreter.SyntaxAnalyze(invalidCode)
	if success {
		t.Error("Invalid syntax should fail analysis")
	}
	if message == "" {
		t.Error("Should return error message for invalid syntax")
	}
}

func TestMainExecutionStats(t *testing.T) {
	// Test that execution stats are included in output
	script := `
	fun main():
		i = 0
		while (i < 3):
			print(i)
			i = i + 1
		end
	end
	`

	success, output := interpreter.RunProgram(script)
	if !success {
		t.Errorf("Script should execute successfully, got: %s", output)
	}

	// Check that execution statistics are present
	expectedStats := []string{
		"Time Program Execution:",
		"Elapsed Operation:",
		"Builtin Calls:",
		"User Calls:",
	}

	for _, stat := range expectedStats {
		if !strings.Contains(output, stat) {
			t.Errorf("Output should contain '%s', output: %s", stat, output)
		}
	}
}
