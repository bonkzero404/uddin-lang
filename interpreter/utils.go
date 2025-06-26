package interpreter

import (
	"fmt"
	"io"
	"os"
)

// UtilityFunctions provides helper functions for common interpreter operations

// PrintStats prints execution statistics in a readable format
func PrintStats(stats Stats, output io.Writer) {
	fmt.Fprintf(output, "Execution Statistics:\n")
	fmt.Fprintf(output, "  Operations: %d\n", stats.Ops)
	fmt.Fprintf(output, "  User Calls: %d\n", stats.UserCalls)
	fmt.Fprintf(output, "  Builtin Calls: %d\n", stats.BuiltinCalls)
	fmt.Fprintf(output, "  Total: %d\n", stats.Ops+stats.UserCalls+stats.BuiltinCalls)
}

// FormatError formats an error with position information
func FormatError(err error, filename string) string {
	if err == nil {
		return ""
	}

	// Check if it's one of our custom error types with position
	switch e := err.(type) {
	case TypeError:
		return fmt.Sprintf("%s:%d:%d: Type Error: %s", filename, e.pos.Line, e.pos.Column, e.Message)
	case ValueError:
		return fmt.Sprintf("%s:%d:%d: Value Error: %s", filename, e.pos.Line, e.pos.Column, e.Message)
	case NameError:
		return fmt.Sprintf("%s:%d:%d: Name Error: %s", filename, e.pos.Line, e.pos.Column, e.Message)
	case RuntimeError:
		return fmt.Sprintf("%s:%d:%d: Runtime Error: %s", filename, e.pos.Line, e.pos.Column, e.Message)
	default:
		return fmt.Sprintf("%s: %s", filename, err.Error())
	}
}

// SafeExecute executes code with error recovery
func SafeExecute(program *Program, config *Config) (result string, success bool, stats *Stats) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				result = FormatError(err, "script")
			} else {
				result = fmt.Sprintf("Panic: %v", r)
			}
			success = false
		}
	}()

	stats, err := Execute(program, config)
	if err != nil {
		result = FormatError(err, "script")
		success = false
		return
	}

	success = true
	return
}

// ValidateConfig checks if a configuration is valid and sets defaults
func ValidateConfig(config *Config) *Config {
	if config == nil {
		return DefaultConfig()
	}

	// Set defaults for nil fields
	if config.Vars == nil {
		config.Vars = make(map[string]Value)
	}
	if config.Args == nil {
		config.Args = []string{}
	}
	if config.Stdin == nil && !config.IsUnitTest {
		config.Stdin = os.Stdin
	}
	if config.Stdout == nil && !config.IsUnitTest {
		config.Stdout = os.Stdout
	}
	if config.Exit == nil {
		if config.IsUnitTest {
			config.Exit = func(int) {} // No-op for tests
		} else {
			config.Exit = os.Exit
		}
	}

	return config
}

// IsValidIdentifier checks if a string is a valid identifier
func IsValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Check first character
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// CreateSafeValue creates a safe copy of a value for use in different contexts
func CreateSafeValue(value Value) Value {
	return DeepCopy(value)
}
