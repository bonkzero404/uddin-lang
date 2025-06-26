package interpreter

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestImportStatement(t *testing.T) {
	// Create a temporary library file
	libContent := `
    // Test library file
    fun add(a, b):
        return a + b
    end

    fun multiply(a, b):
        return a * b
    end

    test_var = "Hello from library"
    `

	// Write library file
	err := os.WriteFile("test_lib.din", []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}
	defer os.Remove("test_lib.din") // Cleanup

	// Main program that imports the library
	mainProgram := `
    import "test_lib.din"

    fun main():
        result1 = add(5, 3)
        result2 = multiply(4, 6)
        print("add(5, 3) =", result1)
        print("multiply(4, 6) =", result2)
        print("test_var =", test_var)
    end
    `

	var buf bytes.Buffer
	config := &Config{
		Stdout: &buf,
	}

	prog, err := ParseProgram([]byte(mainProgram))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	_, err = Execute(prog, config)
	if err != nil {
		t.Fatalf("Failed to execute program: %v", err)
	}

	output := buf.String()

	// Check expected outputs
	expectedOutputs := []string{
		"add(5, 3) = 8",
		"multiply(4, 6) = 24",
		"test_var = Hello from library",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
		}
	}
}

func TestImportStatementFileNotFound(t *testing.T) {
	// Program that tries to import non-existent file
	program := `
    import "non_existent_file.din"
    `

	var buf bytes.Buffer
	config := &Config{
		Stdout: &buf,
	}

	prog, err := ParseProgram([]byte(program))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	_, err = Execute(prog, config)
	if err == nil {
		t.Error("Expected error when importing non-existent file, but got none")
	}

	// Check that error message mentions the file
	if !strings.Contains(err.Error(), "non_existent_file.din") {
		t.Errorf("Expected error message to mention the file name, got: %v", err)
	}
}

func TestImportStatementSyntaxError(t *testing.T) {
	// Create a library file with actual syntax error
	libContent := `
    // Invalid syntax - missing closing parenthesis
    fun broken_function(
    `

	// Write library file
	err := os.WriteFile("broken_lib.din", []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test library file: %v", err)
	}
	defer os.Remove("broken_lib.din") // Cleanup

	// Program that imports the broken library
	program := `
    import "broken_lib.din"
    `

	var buf bytes.Buffer
	config := &Config{
		Stdout: &buf,
	}

	prog, err := ParseProgram([]byte(program))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	// Execute should return an error due to parsing the imported file
	_, err = Execute(prog, config)
	if err == nil {
		t.Error("Expected error when importing file with syntax error, but got none")
		return
	}

	// Check that the error mentions the broken library file
	errorStr := err.Error()
	if !strings.Contains(errorStr, "broken_lib.din") {
		t.Errorf("Expected error to mention the broken library file, got: %v", err)
	}
}
