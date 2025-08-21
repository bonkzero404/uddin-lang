package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// importFunc implements the import() built-in function
// Imports and executes code from another Uddin-Lang file
// Parameters:
//   - filename: Path to the Uddin-Lang file to import
//
// Returns true if import was successful, false otherwise
// Example: import("utils.kv")
func importFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "import", args, 1)

	// Get the filename argument
	filename, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "import() requires a string filename"))
	}

	// If the filename doesn't have .din extension, add it
	if !strings.HasSuffix(filename, ".din") {
		filename = filename + ".din"
	}

	// Try different paths to find the file
	var possiblePaths []string

	// 1. Try the exact path provided
	possiblePaths = append(possiblePaths, filename)

	// 2. Try in the current directory
	if !filepath.IsAbs(filename) {
		currentDir, err := os.Getwd()
		if err == nil {
			possiblePaths = append(possiblePaths, filepath.Join(currentDir, filename))
		}
	}

	// 3. Try in the examples directory
	possiblePaths = append(possiblePaths, filepath.Join("examples", filename))

	// 4. Try one directory up + examples
	possiblePaths = append(possiblePaths, filepath.Join("..", "examples", filename))

	// Try each path
	var fileContent []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		fileContent, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	// If we couldn't find the file, return false
	if err != nil {
		fmt.Fprintf(interp.stdout, "Error importing file %s: file not found in any of the search paths\n", filename)
		return Value(false)
	}

	// Parse the imported program
	importedProg, err := ParseProgram(fileContent)
	if err != nil {
		fmt.Fprintf(interp.stdout, "Error parsing imported file %s: %s\n", foundPath, err)
		return Value(false)
	}

	// Execute the imported program
	// We don't want to call the main function of the imported file
	// We just want to execute the top-level statements and define the functions
	for _, statement := range importedProg.Statements {
		// Skip if statement is a main function definition
		if funcDef, ok := statement.(*FunctionDefinition); ok && funcDef.Name == "main" {
			continue
		}

		// Execute the statement
		interp.executeStatement(statement)
	}

	return Value(true)
}

// exitFunc implements the exit() built-in function
// Terminates the program with the specified exit code
// Parameters:
//   - code: Optional integer exit code (defaults to 0)
//
// Returns null (though execution stops)
// Example: exit(1)
func exitFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) > 1 {
		panic(typeError(pos, "exit() requires 0 or 1 args, got %d", len(args)))
	}
	code := 0
	if len(args) > 0 {
		arg, ok := args[0].(int)
		if !ok {
			panic(typeError(pos, "exit() requires an integer, not %s", typeName(args[0])))
		}
		code = arg
	}
	interp.exit(code)
	return Value(nil)
}

// printFunc implements the print() built-in function
// Prints values to standard output followed by a newline
// Parameters:
//   - args: Any number of values to print
//
// Returns null
// Example: print("hello", 42) -> hello 42
func printFunc(interp *interpreter, pos Position, args []Value) Value {
	// Convert all arguments to strings
	strs := make([]string, len(args))
	for i, a := range args {
		strs[i] = toString(a, false)
	}
	// Join strings with space and add newline
	output := strings.Join(strs, " ") + "\n"
	// Write to the interpreter's stdout (this allows capture during testing)
	if interp.stdout != nil {
		interp.stdout.Write([]byte(output))
		// Also write to real stdout for interactive programs
		os.Stdout.WriteString(output)
	} else {
		// Fallback to direct stdout if no stdout is configured
		os.Stdout.WriteString(output)
		// Force flush using syscall
		syscall.Syscall(syscall.SYS_FSYNC, uintptr(os.Stdout.Fd()), 0, 0)
	}
	return Value(nil)
}

// inputFunc implements the input() built-in function
// Reads user input from stdin with an optional prompt
// Parameters:
//   - prompt (optional): string to display before reading input
// Returns the input string (without newline)
// Example: name = input("Enter your name: ")
func inputFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) > 1 {
		panic(typeError(pos, "input() takes at most 1 argument (%d given)", len(args)))
	}

	// Print prompt if provided
	if len(args) == 1 {
		prompt := toString(args[0], false)
		// For input prompts, always write to both configured stdout AND real stdout
		// This ensures prompts are visible to the user even when output is captured
		if interp.stdout != nil {
			interp.stdout.Write([]byte(prompt))
		}
		// Always show prompt to user on real stdout for interactive input
		fmt.Print(prompt)
	}

	// Use bufio.Scanner to read the entire line including spaces
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return Value(scanner.Text())
	}

	// Handle scanner error or EOF
	if err := scanner.Err(); err != nil {
		panic(runtimeError(pos, "error reading input: %s", err))
	}

	// EOF case
	return Value("")
}

// typeofFunc implements the typeof() built-in function
// Returns the type name of a value
// Parameters:
//   - value: Any value
// Returns: string type name
// Example: typeof(42) -> "int"
func typeofFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "typeof", args, 1); err != nil {
		return err
	}
	return Value(typeName(args[0]))
}
