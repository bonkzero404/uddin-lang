package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
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
// High-performance print function with zero-copy optimization
func printFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 0 {
		// Fast path for empty print
		writeDirectStdout([]byte("\n"))
		return Value(nil)
	}
	
	// Use string builder from pool for efficient concatenation
	sb := interp.getStringBuilder()
	defer interp.putStringBuilder(sb)
	
	// Build output string efficiently
	for i, arg := range args {
		if i > 0 {
			sb.WriteByte(' ')
		}
		// Direct string conversion without allocation when possible
		switch v := arg.(type) {
		case string:
			sb.WriteString(v)
		case int:
			// Fast integer to string conversion
			writeIntToBuilder(sb, int64(v))
		case int64:
			writeIntToBuilder(sb, v)
		case float64:
			// Fast float to string conversion
			writeFloatToBuilder(sb, v)
		case bool:
			if v {
				sb.WriteString("true")
			} else {
				sb.WriteString("false")
			}
		case nil:
			sb.WriteString("null")
		default:
			// Fallback to toString for complex types
			sb.WriteString(toString(arg, false))
		}
	}
	sb.WriteByte('\n')
	
	// Get string as bytes without copying
	str := sb.String()
	bytes := stringToBytes(str)
	
	// Write output using direct syscall for maximum performance
	if interp.DirectOutput || interp.stdout == nil {
		writeDirectStdout(bytes)
	} else {
		// Use configured stdout for testing
		interp.stdout.Write(bytes)
	}
	
	return Value(nil)
}

// writeDirectStdout performs direct syscall write to stdout
func writeDirectStdout(data []byte) {
	if len(data) == 0 {
		return
	}
	// Direct syscall write - fastest possible I/O
	syscall.Syscall(syscall.SYS_WRITE, uintptr(1), uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
}

// stringToBytes converts string to []byte without copying
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// writeIntToBuilder writes integer to string builder efficiently
func writeIntToBuilder(sb *strings.Builder, n int64) {
	if n == 0 {
		sb.WriteByte('0')
		return
	}
	
	if n < 0 {
		sb.WriteByte('-')
		n = -n
	}
	
	// Fast integer to string conversion
	var buf [20]byte // enough for 64-bit int
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	sb.Write(buf[i:])
}

// writeFloatToBuilder writes float to string builder efficiently
func writeFloatToBuilder(sb *strings.Builder, f float64) {
	// Simple float formatting - can be optimized further
	if f == float64(int64(f)) {
		// Integer value, write as int
		writeIntToBuilder(sb, int64(f))
	} else {
		// Use fallback for now - can implement custom float formatting
		sb.WriteString(fmt.Sprintf("%.6g", f))
	}
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
		// When DirectOutput is enabled, only write to os.Stdout once
		// When DirectOutput is disabled, write to configured stdout and also to real stdout for visibility
		if interp.DirectOutput {
			// DirectOutput mode: write only to os.Stdout
			fmt.Print(prompt)
		} else {
			// Non-DirectOutput mode: write to both configured stdout and real stdout
			if interp.stdout != nil {
				interp.stdout.Write([]byte(prompt))
			}
			// Always show prompt to user on real stdout for interactive input
			fmt.Print(prompt)
		}
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
