package interpreter

import (
	"strings"
)

// stringsToList converts a Go string slice to a Value array for the interpreter
// This is a helper function used by several built-in functions that work with strings
// Parameters:
//   - strings: A slice of Go strings
//
// Returns a Value representing an array of strings in the interpreter
// Example: stringsToList(["hello", "world"]) -> ["hello", "world"]
func stringsToList(strings []string) Value {
	values := make([]Value, len(strings))
	for i, s := range strings {
		values[i] = s
	}
	return Value(&values)
}

// joinFunc implements the join() built-in function
// Joins elements of an array into a string with a specified separator
// Parameters:
//   - list: Array of values to join
//   - sep: String separator to place between elements
//
// Returns the joined string
// Example: join(["hello", "world"], ", ") -> "hello, world"
func joinFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "join", args, 2); err != Value(nil) {
		return err
	}
	// Check if first argument is an array
	if list, ok := args[0].(*[]Value); ok {
		// Check if second argument is a string
		if sep, ok := args[1].(string); ok {
			// Fast path for empty arrays
			if len(*list) == 0 {
				return Value("")
			}
			// Fast path for single element
			if len(*list) == 1 {
				return Value(toString((*list)[0], true))
			}
			// Use string builder pool for better performance
			builder := interp.getStringBuilder()
			defer interp.putStringBuilder(builder)
			for i, v := range *list {
				if i > 0 {
					builder.WriteString(sep)
				}
				builder.WriteString(toString(v, true))
			}
			return Value(builder.String())
		}
		panic(typeError(pos, "join() requires second argument to be a string"))
	}
	panic(typeError(pos, "join() requires first argument to be an array"))
}

// splitFunc implements the split() built-in function
// Splits a string into an array of substrings
// Parameters:
//   - str: String to split
//   - sep: Optional separator string (if omitted, splits on whitespace)
//
// Returns an array of substrings
// Example: split("hello,world", ",") -> ["hello", "world"]
func splitFunc(interp *interpreter, pos Position, args []Value) Value {
	// Check argument count
	if len(args) != 1 && len(args) != 2 {
		panic(typeError(pos, "split() requires 1 or 2 args, got %d", len(args)))
	}

	// Check that first argument is a string
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "split() requires first argument to be a string"))
	}

	// Split the string
	var parts []string
	if len(args) == 1 || args[1] == nil {
		// Split on whitespace if no separator provided
		parts = strings.Fields(str)
	} else if sep, ok := args[1].(string); ok {
		// Split on the provided separator
		parts = strings.Split(str, sep)
	} else {
		panic(typeError(pos, "split() requires separator to be a str or null"))
	}

	// Convert string slice to Value array
	return stringsToList(parts)
}

// lowerFunc implements the lower() built-in function
// Converts a string to lowercase
// Parameters:
//   - str: String to convert
//
// Returns the lowercase string
// Example: lower("HELLO") -> "hello"
func lowerFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "lower", args, 1); err != Value(nil) {
		return err
	}
	if s, ok := args[0].(string); ok {
		return Value(strings.ToLower(s))
	}
	panic(typeError(pos, "lower() requires a string"))
}

// upperFunc implements the upper() built-in function
// Converts a string to uppercase
// Parameters:
//   - str: String to convert
//
// Returns the uppercase string
// Example: upper("hello") -> "HELLO"
func upperFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "upper", args, 1); err != nil {
		return err
	}
	if s, ok := args[0].(string); ok {
		return Value(strings.ToUpper(s))
	}
	panic(typeError(pos, "upper() requires a string"))
}

// contains(haystack: string, needle: string) -> bool
// contains(haystack: array, needle: any) -> bool
// Example: contains("hello", "ell") -> true
// Example: contains([1, 2, 3], 2) -> true
func containsFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "contains", args, 2); err != Value(nil) {
		return err
	}
	switch haystack := args[0].(type) {
	case string:
		if needle, ok := args[1].(string); ok {
			return Value(strings.Contains(haystack, needle))
		}
		panic(typeError(pos, "contains() on str requires second argument to be a string"))
	case *[]Value:
		needle := args[1]
		for _, v := range *haystack {
			if evalEqual(pos, needle, v).(bool) {
				return Value(true)
			}
		}
		return Value(false)
	default:
		panic(typeError(pos, "contains() requires first argument to be a string or array"))
	}
}

// str_pad(s: str, pad_len: int, pad_str: string) -> string
// Example: str_pad("hello", 10, " ") -> "hello     "
func strpadFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "str_pad", args, 3); err != Value(nil) {
		return err
	}
	if s, ok := args[0].(string); ok {
		if padLen, ok := args[1].(int); ok {
			if padStr, ok := args[2].(string); ok {
				return Value(s + strings.Repeat(padStr, padLen))
			}
			panic(typeError(pos, "str_pad() requires third argument to be a string"))
		}
		panic(typeError(pos, "str_pad() requires second argument to be an integer"))
	}
	panic(typeError(pos, "str_pad() requires first argument to be a string"))
}

// substr(s: str, start: int, end: int) -> string
// Example: substr("hello", 1, 3) -> "ell"
func substrFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "substr", args, 3); err != Value(nil) {
		return err
	}
	if s, ok := args[0].(string); ok {
		if start, ok := args[1].(int); ok {
			if end, ok := args[2].(int); ok {
				return Value(s[start:end])
			}
			panic(typeError(pos, "substr() requires third argument to be an integer"))
		}
		panic(typeError(pos, "substr() requires second argument to be an integer"))
	}
	panic(typeError(pos, "substr() requires first argument to be a string"))
}

// findFunc implements the find() built-in function
// Finds the index of a substring in a string or an element in an array
// Parameters:
//   - haystack: String or array to search in
//   - needle: String to find in a string haystack, or any value to find in an array haystack
//
// Returns the index of the first occurrence, or -1 if not found
// Example: find("hello", "ell") -> 1
func findFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "find", args, 2); err != Value(nil) {
		return err
	}
	switch haystack := args[0].(type) {
	case string:
		if needle, ok := args[1].(string); ok {
			return Value(strings.Index(haystack, needle))
		}
		panic(typeError(pos, "find() on string requires second argument to be a string"))
	case *[]Value:
		needle := args[1]
		for i, v := range *haystack {
			if evalEqual(pos, needle, v).(bool) {
				return Value(i)
			}
		}
		return Value(-1)
	default:
		panic(typeError(pos, "find() requires first argument to be a string or array"))
	}
}

// ========================================
// String Manipulation Functions (Point 1)
// ========================================

// replaceFunc implements the replace() built-in function
// replace(str, old, new) -> string
// Example: replace("hello world", "world", "universe") -> "hello universe"
func replaceFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "replace", args, 3); err != nil {
		return err
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "replace() requires first argument to be a string"))
	}
	old, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "replace() requires second argument to be a string"))
	}
	new, ok := args[2].(string)
	if !ok {
		panic(typeError(pos, "replace() requires third argument to be a string"))
	}
	return Value(strings.ReplaceAll(str, old, new))
}

// trimFunc implements the trim() built-in function
// trim(str) -> string
// Example: trim("  hello world  ") -> "hello world"
func trimFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "trim", args, 1); err != nil {
		return err
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "trim() requires a string argument"))
	}
	return Value(strings.TrimSpace(str))
}

// startsWithFunc implements the starts_with() built-in function
// starts_with(str, prefix) -> bool
// Example: starts_with("hello world", "hello") -> true
func startsWithFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "starts_with", args, 2); err != nil {
		return err
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires first argument to be a string"))
	}
	prefix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires second argument to be a string"))
	}
	return Value(strings.HasPrefix(str, prefix))
}

// endsWithFunc implements the ends_with() built-in function
// ends_with(str, suffix) -> bool
// Example: ends_with("hello world", "world") -> true
func endsWithFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "ends_with", args, 2); err != nil {
		return err
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires first argument to be a string"))
	}
	suffix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires second argument to be a string"))
	}
	return Value(strings.HasSuffix(str, suffix))
}

// repeatFunc implements the repeat() built-in function
// repeat(str, count) -> string
// Example: repeat("hello", 3) -> "hellohellohello"
func repeatFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "repeat", args, 2); err != nil {
		return Value(err)
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "repeat() requires first argument to be a string"))
	}
	count, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "repeat() requires second argument to be an integer"))
	}
	if count < 0 {
		panic(valueError(pos, "repeat() count cannot be negative"))
	}
	return Value(strings.Repeat(str, count))
}

// reverseStrFunc implements the reverse_str() built-in function
// reverse_str(str) -> string
// Example: reverse_str("hello") -> "olleh"
func reverseStrFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "reverse_str", args, 1); err != nil {
		return Value(err)
	}
	str, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "reverse_str() requires a string argument"))
	}
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return Value(string(runes))
}
