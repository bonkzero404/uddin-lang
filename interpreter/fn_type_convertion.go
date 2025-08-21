package interpreter

import (
	"math"
	"strconv"
)

// intFunc implements the int() built-in function
// Converts a value to an integer
// Parameters:
//   - value: Value to convert (string, int, float, or boolean)
//
// Returns the integer value, or null if conversion fails
// Example: int("42") -> 42, int(45.67) -> 45
func intFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "int", args, 1); err != Value(nil) {
		return err
	}
	switch arg := args[0].(type) {
	case int:
		return args[0] // Already an integer
	case int64:
		return Value(int(arg)) // Convert int64 to int
	case float64:
		return Value(int(arg)) // Convert float64 to int (truncation)
	case bool:
		if arg {
			return Value(1) // true -> 1
		} else {
			return Value(0) // false -> 0
		}
	case string:
		// Try to parse as integer first
		if i, err := strconv.Atoi(arg); err == nil {
			return Value(i)
		}
		// Try to parse as float and convert to int
		if f, err := strconv.ParseFloat(arg, 64); err == nil {
			return Value(int(f))
		}
		return Value(nil) // Return null if conversion fails
	default:
		panic(typeError(pos, "int() requires an int, int64, float64, bool, or a string"))
	}
}

// floatFunc implements the float() built-in function
// Converts a value to a float with specified precision
// Parameters:
//   - value: Value to convert (float, int, or string)
//   - digits: Number of decimal places to round to
//
// Returns the float value with specified precision
// Example: float(3.14159, 2) -> 3.14
func floatFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "float", args, 2); err != Value(nil) {
		return err
	}

	// Helper function to set decimal precision
	setDigit := func(f any, digit int) (float64, Value) {
		if digit < 0 {
			return 0, Value(valueError(pos, "float() digit must not be negative"))
		}

		if f, ok := f.(float64); ok {
			// Round to specified number of decimal places
			f = math.Round(f*math.Pow(10, float64(digit))) / math.Pow(10, float64(digit))

			// Ensure we keep decimal point even for whole numbers
			if f == math.Trunc(f) {
				f = f + 0.0
			}

			return f, nil
		}

		return 0, Value(typeError(pos, "float() requires a float"))
	}

	switch arg := args[0].(type) {
	case float64:
		// Already a float
		f, err := setDigit(args[0], args[1].(int))
		if err != Value(nil) {
			return err
		}
		return Value(f)
	case int:
		// Convert int to float
		f, err := setDigit(float64(arg), args[1].(int))
		if err != Value(nil) {
			return err
		}
		return Value(f)
	case string:
		// Parse string to float
		f, _ := strconv.ParseFloat(arg, 64)
		fl, err := setDigit(f, args[1].(int))
		if err != Value(nil) {
			return err
		}
		return Value(fl)
	default:
		panic(typeError(pos, "float() requires an integer or a string"))
	}
}

// strFunc implements the str() built-in function
// Converts a value to a string
// Parameters:
//   - value: Value to convert
//
// Returns the string representation of the value
// Example: str(42) -> "42"
func strFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "str", args, 1); err != nil {
		return err
	}
	return Value(toString(args[0], false))
}

// charFunc implements the char() built-in function
// Converts a Unicode code point to its corresponding character
// Parameters:
//   - code: Integer Unicode code point
//
// Returns the character as a string
// Example: char(97) -> "a"
func charFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "char", args, 1); err != Value(nil) {
		return err
	}
	if code, ok := args[0].(int); ok {
		return string(rune(code))
	}
	panic(typeError(pos, "char() requires an integer, not %s", typeName(args[0])))
}

// runeFunc implements the rune() built-in function
// Converts a single-character string to its Unicode code point
// Parameters:
//   - str: Single-character string
//
// Returns the Unicode code point as an integer
// Example: rune("a") -> 97
func runeFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "rune", args, 1); err != Value(nil) {
		return err
	}
	if s, ok := args[0].(string); ok {
		// Convert string to rune array
		runes := []rune(s)
		// Check that string contains exactly one character
		if len(runes) != 1 {
			panic(valueError(pos, "rune() requires a 1-character string"))
		}
		// Return the Unicode code point
		return Value(int(runes[0]))
	}
	panic(typeError(pos, "rune() requires a string"))
}
