package interpreter

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// functionType is the interface for all callable functions in the interpreter
// Both user-defined functions and built-in functions implement this interface
type functionType interface {
	// call executes the function with the given arguments and returns a value
	call(interp *interpreter, pos Position, args []Value) Value

	// name returns a string representation of the function for debugging
	name() string
}

// userFunction represents a function defined by the user in the script
type userFunction struct {
	Name       string           // Function name (can be empty for anonymous functions)
	Parameters []string         // Parameter names
	Ellipsis   bool             // Whether the last parameter is variadic
	Body       Block            // Function body statements
	Closure    map[string]Value // Captured variables from outer scopes
}

// ensureNumArgs checks if the number of arguments matches the required count
// If not, it panics with a type error indicating the mismatch
// Parameters:
//   - pos: Position in source code for error reporting
//   - name: Function name for error message
//   - args: Actual arguments passed
//   - required: Required number of arguments
func ensureNumArgs(pos Position, name string, args []Value, required int) {
	if len(args) != required {
		plural := ""
		if required != 1 {
			plural = "s"
		}
		panic(typeError(pos, "%s() requires %d arg%s, got %d", name, required, plural, len(args)))
	}
}

// call implements the functionType interface for user-defined functions
// It sets up the function's scope, assigns arguments to parameters, and executes the function body
// Parameters:
//   - interp: The interpreter instance
//   - pos: Position in source code for error reporting
//   - args: Arguments passed to the function
//
// Returns the function's return value or nil if no return statement was executed
func (f *userFunction) call(interp *interpreter, pos Position, args []Value) Value {
	// Handle variadic arguments if this is a variadic function
	if f.Ellipsis {
		ellipsisArgs := args[len(f.Parameters)-1:]
		newArgs := make([]Value, 0, len(f.Parameters)+1)
		newArgs = append(newArgs, args[:len(f.Parameters)-1]...)
		args = append(newArgs, Value(&ellipsisArgs))
	}

	// Verify argument count
	ensureNumArgs(pos, f.Name, args, len(f.Parameters))

	// Set up closure scope (captured variables)
	interp.pushScope(f.Closure)
	defer interp.popScope()

	// Set up local function scope
	interp.pushScope(make(map[string]Value))
	defer interp.popScope()

	// Assign arguments to parameters
	for i, arg := range args {
		interp.assign(f.Parameters[i], arg)
	}

	// Track function call statistics
	interp.stats.UserCalls++

	// Execute the function body
	interp.executeBlock(f.Body)
	return Value(nil)
}

// name implements the functionType interface for user-defined functions
// Returns a string representation of the function for debugging
func (f *userFunction) name() string {
	if f.Name == "" {
		return "<fun>" // Anonymous function
	}
	return fmt.Sprintf("<fun %s>", f.Name)
}

// builtinFunction represents a built-in function provided by the interpreter
type builtinFunction struct {
	Function func(interp *interpreter, pos Position, args []Value) Value // Implementation function
	Name     string                                                      // Function name for debugging
}

// call implements the functionType interface for built-in functions
// It tracks statistics and delegates to the actual implementation function
// Parameters:
//   - interp: The interpreter instance
//   - pos: Position in source code for error reporting
//   - args: Arguments passed to the function
//
// Returns the function's return value
func (f builtinFunction) call(interp *interpreter, pos Position, args []Value) Value {
	interp.stats.BuiltinCalls++
	return f.Function(interp, pos, args)
}

// name implements the functionType interface for built-in functions
// Returns a string representation of the function for debugging
func (f builtinFunction) name() string {
	return fmt.Sprintf("<builtin %s>", f.Name)
}

var builtins = map[string]builtinFunction{
	"append":         {appendFunc, "append"},
	"char":           {charFunc, "char"},
	"exit":           {exitFunc, "exit"},
	"find":           {findFunc, "find"},
	"import":         {importFunc, "import"},
	"int":            {intFunc, "int"},
	"float":          {floatFunc, "float"},
	"join":           {joinFunc, "join"},
	"len":            {lenFunc, "len"},
	"lower":          {lowerFunc, "lower"},
	"print":          {printFunc, "print"},
	"range":          {rangeFunc, "range"},
	"rune":           {runeFunc, "rune"},
	"slice":          {sliceFunc, "slice"},
	"sort":           {sortFunc, "sort"},
	"split":          {splitFunc, "split"},
	"str":            {strFunc, "str"},
	"contains":       {containsFunc, "contains"},
	"str_pad":        {strpadFunc, "str_pad"},
	"substr":         {substrFunc, "substr"},
	"upper":          {upperFunc, "upper"},
	"typeof":         {typeofFunc, "typeof"},
	"is_regex_match": {isregexFunc, "is_regex_match"},
	"date_now":       {datenowFunc, "date_now"},
	"date_format":    {dateformatFunc, "date_format"},
}

// appendFunc implements the append() built-in function
// Appends elements to the end of an array and modifies it in place
// Parameters:
//   - list: The array to append to (first argument)
//   - args: Elements to append to the array (remaining arguments)
//
// Returns null
// Example: append([1, 2], 3, 4) -> [1, 2, 3, 4]
func appendFunc(interp *interpreter, pos Position, args []Value) Value {
	// Check if at least one argument is provided
	if len(args) < 1 {
		panic(typeError(pos, "append() requires at least 1 arg, got %d", len(args)))
	}

	// Check if first argument is an array
	if list, ok := args[0].(*[]Value); ok {
		// Append all remaining arguments to the array
		*list = append(*list, args[1:]...)
		return Value(nil)
	}

	// Error if first argument is not an array
	panic(typeError(pos, "append() requires first argument to be list"))
}

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

// charFunc implements the char() built-in function
// Converts a Unicode code point to its corresponding character
// Parameters:
//   - code: Integer Unicode code point
//
// Returns the character as a string
// Example: char(97) -> "a"
func charFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "char", args, 1)
	if code, ok := args[0].(int); ok {
		return string(rune(code))
	}
	panic(typeError(pos, "char() requires an integer, not %s", typeName(args[0])))
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

// findFunc implements the find() built-in function
// Finds the index of a substring in a string or an element in an array
// Parameters:
//   - haystack: String or array to search in
//   - needle: String to find in a string haystack, or any value to find in an array haystack
//
// Returns the index of the first occurrence, or -1 if not found
// Example: find("hello", "ell") -> 1
func findFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "find", args, 2)
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

// intFunc implements the int() built-in function
// Converts a value to an integer
// Parameters:
//   - value: Value to convert (string or int)
//
// Returns the integer value, or null if conversion fails
// Example: int("42") -> 42
func intFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "int", args, 1)
	switch arg := args[0].(type) {
	case int:
		return args[0] // Already an integer
	case string:
		i, err := strconv.Atoi(arg)
		if err != nil {
			return Value(nil) // Return null if conversion fails
		}
		return Value(i)
	default:
		panic(typeError(pos, "int() requires an int or a string"))
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
	ensureNumArgs(pos, "float", args, 2)

	// Helper function to set decimal precision
	setDigit := func(f any, digit int) float64 {
		if digit < 0 {
			panic(valueError(pos, "float() digit must not be negative"))
		}

		if f, ok := f.(float64); ok {
			// Round to specified number of decimal places
			f = math.Round(f*math.Pow(10, float64(digit))) / math.Pow(10, float64(digit))

			// Ensure we keep decimal point even for whole numbers
			if f == math.Trunc(f) {
				f = f + 0.0
			}

			return f
		}

		panic(typeError(pos, "float() requires a float"))
	}

	switch arg := args[0].(type) {
	case float64:
		// Already a float
		f := setDigit(args[0], args[1].(int))
		return Value(f)
	case int:
		// Convert int to float
		f := setDigit(float64(arg), args[1].(int))
		return Value(f)
	case string:
		// Parse string to float
		f, _ := strconv.ParseFloat(arg, 64)
		fl := setDigit(f, args[1].(int))

		return Value(fl)
	default:
		panic(typeError(pos, "float() requires an integer or a string"))
	}
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
	ensureNumArgs(pos, "join", args, 2)
	// Check if first argument is an array
	if list, ok := args[0].(*[]Value); ok {
		// Check if second argument is a string
		if sep, ok := args[1].(string); ok {
			// Convert each array element to string
			strs := make([]string, len(*list))
			for i, v := range *list {
				strs[i] = toString(v, true)
			}
			// Join the strings with the separator
			return Value(strings.Join(strs, sep))
		}
		panic(typeError(pos, "join() requires second argument to be a string"))
	}
	panic(typeError(pos, "join() requires first argument to be an array"))
}

// lenFunc implements the len() built-in function
// Returns the length of a string, array, or object
// Parameters:
//   - value: String, array, or object to get the length of
//
// Returns the length as an integer
// Example: len("hello") -> 5
func lenFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "len", args, 1)
	var length int
	switch arg := args[0].(type) {
	case string:
		// Length of string (in bytes, not runes)
		length = len(arg)
	case *[]Value:
		// Number of elements in array
		length = len(*arg)
	case map[string]Value:
		// Number of key-value pairs in object
		length = len(arg)
	default:
		panic(typeError(pos, "len() requires a string, array, or object"))
	}
	return Value(length)
}

// lowerFunc implements the lower() built-in function
// Converts a string to lowercase
// Parameters:
//   - str: String to convert
//
// Returns the lowercase string
// Example: lower("HELLO") -> "hello"
func lowerFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "lower", args, 1)
	if s, ok := args[0].(string); ok {
		return Value(strings.ToLower(s))
	}
	panic(typeError(pos, "lower() requires a string"))
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
	strs := make([]any, len(args))
	for i, a := range args {
		strs[i] = toString(a, false)
	}
	// Print to stdout with a newline
	fmt.Fprintln(interp.stdout, strs...)
	return Value(nil)
}

// rangeFunc implements the range() built-in function
// Creates an array of integers
// Parameters:
//   - If 1 arg: n (upper bound, exclusive) -> [0, 1, ..., n-1]
//   - If 2 args: start, stop -> [start, start+1, ..., stop-1]
//
// Returns an array of integers
// Examples:
//   range(3) -> [0, 1, 2]
//   range(1, 4) -> [1, 2, 3]
func rangeFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 1 {
		// Single argument: range(n) -> [0, 1, ..., n-1]
		if n, ok := args[0].(int); ok {
			if n < 0 {
				panic(valueError(pos, "range() argument must not be negative"))
			}
			nums := make([]Value, n)
			for i := 0; i < n; i++ {
				nums[i] = i
			}
			return Value(&nums)
		}
		panic(typeError(pos, "range() requires an integer"))
	} else if len(args) == 2 {
		// Two arguments: range(start, stop) -> [start, start+1, ..., stop-1]
		start, startOk := args[0].(int)
		stop, stopOk := args[1].(int)

		if !startOk || !stopOk {
			panic(typeError(pos, "range() requires integer arguments"))
		}

		if start > stop {
			// Return empty array if start > stop
			nums := make([]Value, 0)
			return Value(&nums)
		}

		size := stop - start
		nums := make([]Value, size)
		for i := 0; i < size; i++ {
			nums[i] = start + i
		}
		return Value(&nums)
	}

	panic(valueError(pos, "range() requires 1 or 2 arguments, got %d", len(args)))
}

// runeFunc implements the rune() built-in function
// Converts a single-character string to its Unicode code point
// Parameters:
//   - str: Single-character string
//
// Returns the Unicode code point as an integer
// Example: rune("a") -> 97
func runeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "rune", args, 1)
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

// sliceFunc implements the slice() built-in function
// Extracts a portion of a string or array
// Parameters:
//   - value: String or array to slice
//   - start: Starting index (inclusive)
//   - end: Ending index (exclusive)
//
// Returns a new string or array containing the sliced portion
// Example: slice("hello", 1, 3) -> "el"
func sliceFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "slice", args, 3)
	// Check that start and end are integers
	start, sok := args[1].(int)
	end, eok := args[2].(int)
	if !sok || !eok {
		panic(typeError(pos, "slice() requires start and end to be integers"))
	}

	switch s := args[0].(type) {
	case string:
		// Handle string slicing
		if start < 0 || end > len(s) || start > end {
			panic(valueError(pos, "slice() start or end out of bounds"))
		}
		return Value(s[start:end])
	case *[]Value:
		// Handle array slicing
		if start < 0 || end > len(*s) || start > end {
			panic(valueError(pos, "slice() start or end out of bounds"))
		}
		// Create a new array with the sliced elements
		result := make([]Value, end-start)
		copy(result, (*s)[start:end])
		return Value(&result)
	default:
		panic(typeError(pos, "slice() requires first argument to be a str or array"))
	}
}

// sortFunc implements the sort() built-in function
// Sorts an array in place, optionally using a key function
// Parameters:
//   - list: Array to sort
//   - key: Optional function to extract sort key from each element
//
// Returns null (sorts array in place)
// Example: sort([3, 1, 4, 1, 5, 9]) -> [1, 1, 3, 4, 5, 9]
// Example: sort(["world", "hello"], lambda x: len(x)) -> ["hello", "world"]
func sortFunc(interp *interpreter, pos Position, args []Value) Value {
	// Check argument count
	if len(args) != 1 && len(args) != 2 {
		panic(typeError(pos, "sort() requires 1 or 2 args, got %d", len(args)))
	}

	// Check that first argument is an array
	list, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "sort() requires first argument to be a array"))
	}

	// No need to sort arrays with 0 or 1 elements
	if len(*list) <= 1 {
		return Value(nil)
	}

	// Simple sort without key function
	if len(args) == 1 {
		sort.SliceStable(*list, func(i, j int) bool {
			return evalLess(pos, (*list)[i], (*list)[j]).(bool)
		})
	} else {
		// Sort with key function
		keyFunc, ok := args[1].(functionType)
		if !ok {
			panic(typeError(pos, "sort() requires second argument to be a function"))
		}

		// Decorate, sort, undecorate pattern
		// This ensures we only call the key function once per element
		type pair struct {
			value Value // Original value
			key   Value // Sort key
		}

		// Extract keys for each element
		pairs := make([]pair, len(*list))
		for i, v := range *list {
			key := interp.callFunction(pos, keyFunc, []Value{v})
			pairs[i] = pair{v, key}
		}

		// Sort by keys
		sort.SliceStable(pairs, func(i, j int) bool {
			return evalLess(pos, pairs[i].key, pairs[j].key).(bool)
		})

		// Extract sorted values
		values := make([]Value, len(pairs))
		for i, p := range pairs {
			values[i] = p.value
		}
		*list = values
	}

	return Value(nil)
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

// isregexFunc implements the is_regex_match() built-in function
// Checks if a string matches a regular expression pattern
// Parameters:
//   - pattern: Regular expression pattern
//   - str: String to check against the pattern
//
// Returns true if the string matches the pattern, false otherwise
// Example: is_regex_match("^-?\\d+\\.\\d+$", "3.14") -> true
func isregexFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "regex_match", args, 2)

	// Check that first argument is a string (pattern)
	if pattern, ok := args[0].(string); ok {
		// Check that second argument is a string (target)
		if str, ok := args[1].(string); ok {
			// Compile the regular expression
			re, err := regexp.Compile(pattern)
			if err != nil {
				// Return false if pattern is invalid
				return false
			}
			// Check if the string matches the pattern
			return re.MatchString(str)
		}
		panic(typeError(pos, "regex() requires second argument to be a string"))
	}
	panic(typeError(pos, "regex() requires first argument to be a string"))
}

// datenowFunc implements the date_now() built-in function
// Returns the current date and time in RFC3339 format
// Parameters: none
// Returns the current date/time as a string
// Example: date_now() -> "2020-01-02T15:04:05Z"
func datenowFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "date_now", args, 0)
	// Get current time and format as RFC3339
	return Value(time.Now().Format(time.RFC3339))
}

// dateformatFunc implements the date_format() built-in function
// Formats a date string according to a specified layout
// Parameters:
//   - t: Date string in RFC3339 format
//   - layout: Format string with placeholders (YYYY, MM, DD, etc.)
//
// Returns the formatted date string, or null if parsing fails
// Example: date_format("2020-01-02T15:04:05Z", "YYYY-MM-DD hh:mm:ss") -> "2020-01-02 15:04:05"
func dateformatFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "date_format", args, 2)
	if t, ok := args[0].(string); ok {
		if layout, ok := args[1].(string); ok {
			// Define replacements for date format placeholders
			replacer := strings.NewReplacer(
				"YYYY", "2006", // Year
				"MM", "01", // Month (numeric)
				"DD", "02", // Day
				"hh", "15", // Hour (24-hour)
				"mm", "04", // Minute
				"ss", "05", // Second
				"ee", "Mon", // Weekday (short)
				"EE", "Monday", // Weekday (long)
				"nn", "Jan", // Month (short)
				"NN", "January", // Month (long)
			)

			// Parse the input date string
			parsed, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return Value(nil) // Return null if parsing fails
			}

			// Replace placeholders with Go's time format specifiers
			layout = replacer.Replace(layout)
			return Value(parsed.Format(layout))
		}
		panic(typeError(pos, "date_format() requires second argument to be a string"))
	}
	panic(typeError(pos, "date_format() requires first argument to be a string"))
}

// contains(haystack: string, needle: string) -> bool
// contains(haystack: array, needle: any) -> bool
// Example: contains("hello", "ell") -> true
// Example: contains([1, 2, 3], 2) -> true
func containsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "contains", args, 2)
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
	ensureNumArgs(pos, "str_pad", args, 3)
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
	ensureNumArgs(pos, "substr", args, 3)
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

// toString converts any interpreter Value to its string representation
// This is used by the str() built-in function and for string concatenation
// Parameters:
//   - value: The Value to convert to a string
//   - quoteStr: Whether to quote string values (for display in arrays/objects)
//
// Returns a string representation of the value
func toString(value Value, quoteStr bool) string {
	var s string
	switch v := value.(type) {
	case nil:
		s = "null" // Null value
	case bool:
		if v {
			s = "true"
		} else {
			s = "false"
		}
	case int:
		s = fmt.Sprintf("%d", v) // Integer
	case float64:
		s = fmt.Sprintf("%g", v) // Float
	case string:
		if quoteStr {
			s = fmt.Sprintf("%q", v) // Quoted string for arrays/objects
		} else {
			s = v // Raw string for display
		}
	case *[]Value:
		// Convert array elements recursively
		strs := make([]string, len(*v))
		for i, v := range *v {
			strs[i] = toString(v, true)
		}
		s = fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	case map[string]Value:
		// Convert object key-value pairs recursively
		strs := make([]string, 0, len(v))
		for k, v := range v {
			item := fmt.Sprintf("%q: %s", k, toString(v, true))
			strs = append(strs, item)
		}
		sort.Strings(strs) // Ensure str(output) is consistent
		s = fmt.Sprintf("{%s}", strings.Join(strs, ", "))
	case functionType:
		s = v.name() // Function representation
	default:
		// Interpreter should never give us this
		panic(fmt.Sprintf("str() got unexpected type %T", v))
	}
	return s
}

func strFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "str", args, 1)
	return Value(toString(args[0], false))
}

// typeName returns the type name of a Value as a string
// This is used by the typeof() built-in function
// Parameters:
//   - v: The Value to get the type name of
//
// Returns a string representing the type name
func typeName(v Value) string {
	var t string
	switch v.(type) {
	case nil:
		t = "nullable" // Null value
	case bool:
		t = "boolean" // Boolean value
	case int:
		t = "integer" // Integer value
	case float64:
		t = "float" // Float value
	case string:
		t = "string" // String value
	case *[]Value:
		t = "array" // Array value
	case map[string]Value:
		t = "object" // Map/Object value
	case functionType:
		t = "function" // Function value
	default:
		// Interpreter should never give us this
		panic(fmt.Sprintf("type() got unexpected type %T", v))
	}
	return t
}

func typeofFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "typeof", args, 1)
	return Value(typeName(args[0]))
}

func upperFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "upper", args, 1)
	if s, ok := args[0].(string); ok {
		return Value(strings.ToUpper(s))
	}
	panic(typeError(pos, "upper() requires a string"))
}

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

	// If the filename doesn't have .kv extension, add it
	if !strings.HasSuffix(filename, ".kv") {
		filename = filename + ".kv"
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
