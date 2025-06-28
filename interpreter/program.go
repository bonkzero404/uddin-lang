package interpreter

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// showErrorSource returns a string with the source code and a pointer to the error.
// It formats the error message with a visual indicator pointing to the exact position of the error.
//
// Parameters:
//   - source: The complete source code as a byte array
//   - pos: The position (line and column) where the error occurred
//   - dividerLen: The length of the divider line to use in the error message
//
// Returns:
//   - string: A formatted error message with visual indication of error location
func showErrorSource(source []byte, pos Position, dividerLen int) string {
	errMessage := ""
	errMessage += divider(dividerLen) + "\n"

	// Split source into lines and get the line with the error
	lines := bytes.Split(source, []byte{'\n'})
	errorLine := string(lines[pos.Line-1])

	// Count tabs to adjust the pointer position correctly
	numTabs := strings.Count(errorLine[:pos.Column-1], "\t")

	// Replace tabs with spaces for consistent display
	errMessage += strings.Replace(errorLine, "\t", "    ", -1) + "\n"

	// Add pointer (^) at the error position
	errMessage += strings.Repeat(" ", pos.Column-1) + strings.Repeat("   ", numTabs) + "^" + "\n"
	errMessage += divider(dividerLen) + "\n"

	return errMessage
}

// divider returns a string of dashes of the given length.
// This is used to create visual separation in error messages.
//
// Parameters:
//   - stringLen: The length of the divider line
//
// Returns:
//   - string: A string of dashes with the specified length or an empty string if length is 0
func divider(stringLen int) string {
	// Return empty string for zero length
	if stringLen <= 0 {
		return ""
	}

	// Create a string of dashes with the specified length
	return strings.Repeat("-", stringLen)
}

// writerFunc is a function type that writes to a string.
// It's used to capture output from the interpreter.
type writerFunc func(string)

// Write implements the io.Writer interface for writerFunc.
// This allows writerFunc to be used anywhere an io.Writer is expected.
//
// Parameters:
//   - p: The bytes to write
//
// Returns:
//   - n: The number of bytes written
//   - err: Any error that occurred (always nil in this implementation)
func (wf writerFunc) Write(p []byte) (n int, err error) {
	// Convert bytes to string and call the function
	wf(string(p))

	// Return the number of bytes written and no error
	return len(p), nil
}

// SyntaxAnalyze checks the syntax of the given input source without executing it.
// This function only validates the syntax correctness of the program.
//
// Parameters:
//   - inputSource: The source code to analyze as a string
//
// Returns:
//   - bool: true if syntax is correct, false otherwise
//   - string: A message describing the result (success message or error details)
func SyntaxAnalyze(inputSource string) (bool, string) {
	// Default success message
	console := "All syntax is correct\n"

	// Try to parse the program
	_, err := ParseProgram([]byte(inputSource))

	// Handle parsing errors
	if err != nil {
		errorMessage := fmt.Sprintf("%s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(Error); ok {
			console = showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
		}

		// Add the error message to the console output
		console += errorMessage
		return false, console
	}

	return true, console
}

// RunProgram parses and executes the given program and returns the output.
// This is the main entry point for running code in the interpreter.
//
// Parameters:
//   - inputSource: The source code to run as a string
//
// Returns:
//   - bool: true if execution was successful, false if there was an error
//   - string: The program output or error message
func RunProgram(inputSource string) (bool, string) {
	// Variables to store program output and console messages
	resultProgram := ""
	console := ""

	// Parse the program
	prog, err := ParseProgram([]byte(inputSource))

	// Handle parsing errors
	if err != nil {
		errorMessage := fmt.Sprintf("%s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(Error); ok {
			console += showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
		}

		// Add the error message to the console output
		console += errorMessage
		return false, console
	}

	// Create config to capture output
	config := &Config{
		Stdout: writerFunc(func(s string) {
			resultProgram += s
		}),
	}

	// Execute the program and capture output
	startTime := time.Now()
	stats, err := Execute(prog, config)

	// Handle execution errors
	if err != nil {
		errorMessage := fmt.Sprintf("%s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(ErrorInterpreter); ok {
			console += showErrorSource([]byte(inputSource), e.Position(), len(errorMessage))
		} else if e, ok := err.(Error); ok {
			console += showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
		}

		// Add the error message to the console output
		console += errorMessage
		return false, console
	}

	// Calculate execution time
	elapsedTime := time.Since(startTime)

	// Add execution statistics to the output
	console += resultProgram
	console += fmt.Sprintf("\nTime Program Execution: %v\n", elapsedTime)
	console += fmt.Sprintf("Elapsed Operation: %d Ops (%d/s)\n", stats.Ops, int64(float64(stats.Ops)/elapsedTime.Seconds()))
	console += fmt.Sprintf("Builtin Calls: %d (%d/s)\n", stats.BuiltinCalls, int64(float64(stats.BuiltinCalls)/elapsedTime.Seconds()))
	console += fmt.Sprintf("User Calls: %d (%d/s)\n", stats.UserCalls, int64(float64(stats.UserCalls)/elapsedTime.Seconds()))

	return true, console
}

// RunProgramOptions defines options for running a program
type RunProgramOptions struct {
	ShowProfiling bool // Whether to show execution profiling information
}

// RunProgramWithOptions parses and executes the given program with custom options.
// This function allows you to control whether profiling information is displayed.
//
// Parameters:
//   - inputSource: The source code to run as a string
//   - options: Options for controlling execution behavior
//
// Returns:
//   - bool: true if execution was successful, false if there was an error
//   - string: The program output or error message
func RunProgramWithOptions(inputSource string, options *RunProgramOptions) (bool, string) {
	// Variables to store program output and console messages
	resultProgram := ""
	console := ""

	// Parse the program
	prog, err := ParseProgram([]byte(inputSource))

	// Handle parsing errors
	if err != nil {
		errorMessage := fmt.Sprintf("%s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(Error); ok {
			console += showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
		}

		// Add the error message to the console output
		console += errorMessage
		return false, console
	}

	// Create config to capture output
	config := &Config{
		Stdout: writerFunc(func(s string) {
			resultProgram += s
		}),
	}

	// Execute the program and capture output
	startTime := time.Now()
	stats, err := Execute(prog, config)

	// Handle execution errors
	if err != nil {
		errorMessage := fmt.Sprintf("%s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(ErrorInterpreter); ok {
			console += showErrorSource([]byte(inputSource), e.Position(), len(errorMessage))
		} else if e, ok := err.(Error); ok {
			console += showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
		}

		// Add the error message to the console output
		console += errorMessage
		return false, console
	}

	// Calculate execution time
	elapsedTime := time.Since(startTime)

	// Add program output
	console += resultProgram

	// Optionally add execution statistics
	if options != nil && options.ShowProfiling {
		console += fmt.Sprintf("\nTime Program Execution: %v\n", elapsedTime)
		console += fmt.Sprintf("Elapsed Operation: %d Ops (%d/s)\n", stats.Ops, int64(float64(stats.Ops)/elapsedTime.Seconds()))
		console += fmt.Sprintf("Builtin Calls: %d (%d/s)\n", stats.BuiltinCalls, int64(float64(stats.BuiltinCalls)/elapsedTime.Seconds()))
		console += fmt.Sprintf("User Calls: %d (%d/s)\n", stats.UserCalls, int64(float64(stats.UserCalls)/elapsedTime.Seconds()))
	}

	return true, console
}

// AnalyzeSyntax performs syntax analysis on the given source code without executing it.
// This function only checks for syntax errors and returns detailed error information if found.
//
// Parameters:
//   - inputSource: The source code to analyze as a string
//
// Returns:
//   - bool: true if syntax is valid, false if there are syntax errors
//   - string: Success message or detailed error information
func AnalyzeSyntax(inputSource string) (bool, string) {
	// Parse the program to check syntax
	prog, err := ParseProgram([]byte(inputSource))

	// Handle parsing errors
	if err != nil {
		errorMessage := fmt.Sprintf("Syntax Error: %s", err)

		// If it's an interpreter error with position information, show the source context
		if e, ok := err.(Error); ok {
			errorContext := showErrorSource([]byte(inputSource), e.Position, len(errorMessage))
			return false, errorContext + errorMessage
		}

		return false, errorMessage
	}

	// Perform basic validation on the parsed program
	if err := ValidateProgram(prog); err != nil {
		return false, fmt.Sprintf("Program Validation Error: %s", err)
	}

	// If we reach here, syntax is valid
	return true, "✓ Syntax analysis passed - No syntax errors found\n"
}

// Enhanced error formatting and reporting functions

// FormatExecutionError formats execution errors with better context
func FormatExecutionError(err error, source []byte, filename string) string {
	if err == nil {
		return ""
	}

	// Check if it's one of our custom error types with position
	switch e := err.(type) {
	case TypeError:
		return formatErrorWithSource(e.Error(), e.Position(), source, filename, "Type Error")
	case ValueError:
		return formatErrorWithSource(e.Error(), e.Position(), source, filename, "Value Error")
	case NameError:
		return formatErrorWithSource(e.Error(), e.Position(), source, filename, "Name Error")
	case RuntimeError:
		return formatErrorWithSource(e.Error(), e.Position(), source, filename, "Runtime Error")
	default:
		return fmt.Sprintf("%s: %s", filename, err.Error())
	}
}

// formatErrorWithSource formats an error with source code context
func formatErrorWithSource(message string, pos Position, source []byte, filename, errorType string) string {
	result := fmt.Sprintf("%s:%d:%d: %s: %s\n", filename, pos.Line, pos.Column, errorType, message)

	// Add source context if available
	if len(source) > 0 {
		result += showErrorSource(source, pos, 50)
	}

	return result
}

// ValidateProgram performs basic validation on a program before execution
func ValidateProgram(program *Program) error {
	if program == nil {
		return fmt.Errorf("program is nil")
	}
	if program.Statements == nil {
		return fmt.Errorf("program has no statements")
	}
	return nil
}

// CreateDefaultInterpreter creates an interpreter with default settings
func CreateDefaultInterpreter() *interpreter {
	config := &Config{
		Vars:       make(map[string]Value),
		Args:       []string{},
		IsUnitTest: false,
	}
	return newInterpreter(config)
}
