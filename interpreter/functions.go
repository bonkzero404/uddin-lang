package interpreter

import (
	"fmt"
	"math"
	"math/rand"
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

	// Basic Math Operations
	"abs":    {absFunc, "abs"},
	"max":    {maxFunc, "max"},
	"min":    {minFunc, "min"},
	"pow":    {powFunc, "pow"},
	"sqrt":   {sqrtFunc, "sqrt"},
	"cbrt":   {cbrtFunc, "cbrt"},

	// Rounding Functions
	"round": {roundFunc, "round"},
	"floor": {floorFunc, "floor"},
	"ceil":  {ceilFunc, "ceil"},
	"trunc": {truncFunc, "trunc"},

	// Trigonometric Functions
	"sin":   {sinFunc, "sin"},
	"cos":   {cosFunc, "cos"},
	"tan":   {tanFunc, "tan"},
	"asin":  {asinFunc, "asin"},
	"acos":  {acosFunc, "acos"},
	"atan":  {atanFunc, "atan"},
	"atan2": {atan2Func, "atan2"},

	// Hyperbolic Functions
	"sinh": {sinhFunc, "sinh"},
	"cosh": {coshFunc, "cosh"},
	"tanh": {tanhFunc, "tanh"},

	// Logarithmic Functions
	"log":   {logFunc, "log"},
	"log10": {log10Func, "log10"},
	"log2":  {log2Func, "log2"},
	"logb":  {logbFunc, "logb"},
	"exp":   {expFunc, "exp"},
	"exp2":  {exp2Func, "exp2"},

	// Statistical Functions
	"sum":      {sumFunc, "sum"},
	"mean":     {meanFunc, "mean"},
	"median":   {medianFunc, "median"},
	"mode":     {modeFunc, "mode"},
	"std_dev":  {stdDevFunc, "std_dev"},
	"variance": {varianceFunc, "variance"},

	// Number Theory Functions
	"gcd":           {gcdFunc, "gcd"},
	"lcm":           {lcmFunc, "lcm"},
	"factorial":     {factorialFunc, "factorial"},
	"fibonacci":     {fibonacciFunc, "fibonacci"},
	"is_prime":      {isPrimeFunc, "is_prime"},
	"prime_factors": {primeFactorsFunc, "prime_factors"},

	// Random Number Functions
	"random":        {randomFunc, "random"},
	"random_int":    {randomIntFunc, "random_int"},
	"random_float":  {randomFloatFunc, "random_float"},
	"random_choice": {randomChoiceFunc, "random_choice"},
	"shuffle":       {shuffleFunc, "shuffle"},
	"seed_random":   {seedRandomFunc, "seed_random"},

	// Utility Functions
	"sign":        {signFunc, "sign"},
	"clamp":       {clampFunc, "clamp"},
	"lerp":        {lerpFunc, "lerp"},
	"degrees":     {degreesFunc, "degrees"},
	"radians":     {radiansFunc, "radians"},
	"is_nan":      {isNanFunc, "is_nan"},
	"is_infinite": {isInfiniteFunc, "is_infinite"},
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
	case []Value:
		// Number of elements in array
		length = len(arg)
	case *[]Value:
		// Number of elements in array (pointer variant)
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
	case []Value:
		// Convert array elements recursively (slice variant)
		strs := make([]string, len(v))
		for i, val := range v {
			strs[i] = toString(val, true)
		}
		s = fmt.Sprintf("[%s]", strings.Join(strs, ", "))
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

// ========================================
// Mathematical Functions Implementation
// ========================================

// Helper function to convert Value to float64
func toFloat64(pos Position, v Value, funcName string) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case float64:
		return val
	default:
		panic(typeError(pos, "%s() requires a number, got %s", funcName, typeName(v)))
	}
}

// Helper function to convert Value to int
func toInt(pos Position, v Value, funcName string) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		panic(typeError(pos, "%s() requires a number, got %s", funcName, typeName(v)))
	}
}

// ========================================
// Basic Math Operations
// ========================================

// absFunc implements the abs() built-in function
// Returns the absolute value of a number
func absFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "abs", args, 1)
	switch val := args[0].(type) {
	case int:
		if val < 0 {
			return Value(-val)
		}
		return Value(val)
	case float64:
		return Value(math.Abs(val))
	default:
		panic(typeError(pos, "abs() requires a number, got %s", typeName(args[0])))
	}
}

// maxFunc implements the max() built-in function
// Returns the maximum value from multiple arguments or from an array
func maxFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		panic(typeError(pos, "max() requires at least 1 argument"))
	}

	// If first argument is an array, find max within the array
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "max() cannot be applied to empty array"))
		}
		maxVal := (*arr)[0]
		for i := 1; i < len(*arr); i++ {
			if evalLess(pos, maxVal, (*arr)[i]).(bool) {
				maxVal = (*arr)[i]
			}
		}
		return maxVal
	}

	// Otherwise, find max among all arguments
	maxVal := args[0]
	for i := 1; i < len(args); i++ {
		if evalLess(pos, maxVal, args[i]).(bool) {
			maxVal = args[i]
		}
	}
	return maxVal
}

// minFunc implements the min() built-in function
// Returns the minimum value from multiple arguments or from an array
func minFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		panic(typeError(pos, "min() requires at least 1 argument"))
	}

	// If first argument is an array, find min within the array
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "min() cannot be applied to empty array"))
		}
		minVal := (*arr)[0]
		for i := 1; i < len(*arr); i++ {
			if evalLess(pos, (*arr)[i], minVal).(bool) {
				minVal = (*arr)[i]
			}
		}
		return minVal
	}

	// Otherwise, find min among all arguments
	minVal := args[0]
	for i := 1; i < len(args); i++ {
		if evalLess(pos, args[i], minVal).(bool) {
			minVal = args[i]
		}
	}
	return minVal
}

// powFunc implements the pow() built-in function
// Returns base raised to the power of exponent
func powFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "pow", args, 2)
	base := toFloat64(pos, args[0], "pow")
	exp := toFloat64(pos, args[1], "pow")
	result := math.Pow(base, exp)

	// Return int if result is a whole number and both inputs were ints
	if _, baseIsInt := args[0].(int); baseIsInt {
		if _, expIsInt := args[1].(int); expIsInt {
			if result == math.Trunc(result) {
				return Value(int(result))
			}
		}
	}
	return Value(result)
}

// sqrtFunc implements the sqrt() built-in function
// Returns the square root of a number
func sqrtFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "sqrt", args, 1)
	val := toFloat64(pos, args[0], "sqrt")
	if val < 0 {
		panic(valueError(pos, "sqrt() of negative number"))
	}
	return Value(math.Sqrt(val))
}

// cbrtFunc implements the cbrt() built-in function
// Returns the cube root of a number
func cbrtFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "cbrt", args, 1)
	val := toFloat64(pos, args[0], "cbrt")
	return Value(math.Cbrt(val))
}

// ========================================
// Rounding Functions
// ========================================

// roundFunc implements the round() built-in function
// Rounds to nearest integer or to specified decimal places
func roundFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 1 {
		val := toFloat64(pos, args[0], "round")
		return Value(int(math.Round(val)))
	} else if len(args) == 2 {
		val := toFloat64(pos, args[0], "round")
		places := toInt(pos, args[1], "round")
		if places < 0 {
			panic(valueError(pos, "round() decimal places must not be negative"))
		}
		multiplier := math.Pow(10, float64(places))
		return Value(math.Round(val*multiplier) / multiplier)
	}
	panic(typeError(pos, "round() requires 1 or 2 arguments, got %d", len(args)))
}

// floorFunc implements the floor() built-in function
// Returns the largest integer less than or equal to the number
func floorFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "floor", args, 1)
	val := toFloat64(pos, args[0], "floor")
	return Value(int(math.Floor(val)))
}

// ceilFunc implements the ceil() built-in function
// Returns the smallest integer greater than or equal to the number
func ceilFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "ceil", args, 1)
	val := toFloat64(pos, args[0], "ceil")
	return Value(int(math.Ceil(val)))
}

// truncFunc implements the trunc() built-in function
// Returns the integer part of a number (truncates decimal part)
func truncFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "trunc", args, 1)
	val := toFloat64(pos, args[0], "trunc")
	return Value(int(math.Trunc(val)))
}

// ========================================
// Trigonometric Functions
// ========================================

// sinFunc implements the sin() built-in function
func sinFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "sin", args, 1)
	val := toFloat64(pos, args[0], "sin")
	return Value(math.Sin(val))
}

// cosFunc implements the cos() built-in function
func cosFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "cos", args, 1)
	val := toFloat64(pos, args[0], "cos")
	return Value(math.Cos(val))
}

// tanFunc implements the tan() built-in function
func tanFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tan", args, 1)
	val := toFloat64(pos, args[0], "tan")
	return Value(math.Tan(val))
}

// asinFunc implements the asin() built-in function
func asinFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "asin", args, 1)
	val := toFloat64(pos, args[0], "asin")
	if val < -1 || val > 1 {
		panic(valueError(pos, "asin() input must be between -1 and 1"))
	}
	return Value(math.Asin(val))
}

// acosFunc implements the acos() built-in function
func acosFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "acos", args, 1)
	val := toFloat64(pos, args[0], "acos")
	if val < -1 || val > 1 {
		panic(valueError(pos, "acos() input must be between -1 and 1"))
	}
	return Value(math.Acos(val))
}

// atanFunc implements the atan() built-in function
func atanFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "atan", args, 1)
	val := toFloat64(pos, args[0], "atan")
	return Value(math.Atan(val))
}

// atan2Func implements the atan2() built-in function
func atan2Func(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "atan2", args, 2)
	y := toFloat64(pos, args[0], "atan2")
	x := toFloat64(pos, args[1], "atan2")
	return Value(math.Atan2(y, x))
}

// ========================================
// Hyperbolic Functions
// ========================================

// sinhFunc implements the sinh() built-in function
func sinhFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "sinh", args, 1)
	val := toFloat64(pos, args[0], "sinh")
	return Value(math.Sinh(val))
}

// coshFunc implements the cosh() built-in function
func coshFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "cosh", args, 1)
	val := toFloat64(pos, args[0], "cosh")
	return Value(math.Cosh(val))
}

// tanhFunc implements the tanh() built-in function
func tanhFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tanh", args, 1)
	val := toFloat64(pos, args[0], "tanh")
	return Value(math.Tanh(val))
}

// ========================================
// Logarithmic Functions
// ========================================

// logFunc implements the log() built-in function (natural logarithm)
func logFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "log", args, 1)
	val := toFloat64(pos, args[0], "log")
	if val <= 0 {
		panic(valueError(pos, "log() of non-positive number"))
	}
	return Value(math.Log(val))
}

// log10Func implements the log10() built-in function
func log10Func(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "log10", args, 1)
	val := toFloat64(pos, args[0], "log10")
	if val <= 0 {
		panic(valueError(pos, "log10() of non-positive number"))
	}
	return Value(math.Log10(val))
}

// log2Func implements the log2() built-in function
func log2Func(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "log2", args, 1)
	val := toFloat64(pos, args[0], "log2")
	if val <= 0 {
		panic(valueError(pos, "log2() of non-positive number"))
	}
	return Value(math.Log2(val))
}

// logbFunc implements the logb() built-in function (logarithm with custom base)
func logbFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "logb", args, 2)
	val := toFloat64(pos, args[0], "logb")
	base := toFloat64(pos, args[1], "logb")
	if val <= 0 {
		panic(valueError(pos, "logb() value must be positive"))
	}
	if base <= 0 || base == 1 {
		panic(valueError(pos, "logb() base must be positive and not equal to 1"))
	}
	return Value(math.Log(val) / math.Log(base))
}

// expFunc implements the exp() built-in function (e^x)
func expFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "exp", args, 1)
	val := toFloat64(pos, args[0], "exp")
	return Value(math.Exp(val))
}

// exp2Func implements the exp2() built-in function (2^x)
func exp2Func(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "exp2", args, 1)
	val := toFloat64(pos, args[0], "exp2")
	return Value(math.Exp2(val))
}

// ========================================
// Statistical Functions
// ========================================

// sumFunc implements the sum() built-in function
func sumFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "sum", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		var total float64
		var hasFloat bool

		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
				hasFloat = true
			default:
				panic(typeError(pos, "sum() array must contain only numbers"))
			}
		}

		if hasFloat {
			return Value(total)
		}
		return Value(int(total))
	}
	panic(typeError(pos, "sum() requires an array"))
}

// meanFunc implements the mean() built-in function
func meanFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "mean", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "mean() of empty array"))
		}

		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "mean() array must contain only numbers"))
			}
		}

		return Value(total / float64(len(*arr)))
	}
	panic(typeError(pos, "mean() requires an array"))
}

// medianFunc implements the median() built-in function
func medianFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "median", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "median() of empty array"))
		}

		// Convert to float slice and sort
		nums := make([]float64, len(*arr))
		for i, v := range *arr {
			switch val := v.(type) {
			case int:
				nums[i] = float64(val)
			case float64:
				nums[i] = val
			default:
				panic(typeError(pos, "median() array must contain only numbers"))
			}
		}

		sort.Float64s(nums)
		n := len(nums)

		if n%2 == 0 {
			// Even number of elements - return average of middle two
			return Value((nums[n/2-1] + nums[n/2]) / 2)
		} else {
			// Odd number of elements - return middle element
			return Value(nums[n/2])
		}
	}
	panic(typeError(pos, "median() requires an array"))
}

// modeFunc implements the mode() built-in function
func modeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "mode", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "mode() of empty array"))
		}

		frequency := make(map[string]int)
		valueMap := make(map[string]Value)

		// Count frequencies
		for _, v := range *arr {
			key := toString(v, true)
			frequency[key]++
			valueMap[key] = v
		}

		// Find the most frequent value
		maxCount := 0
		var modeKey string
		for key, count := range frequency {
			if count > maxCount {
				maxCount = count
				modeKey = key
			}
		}

		return valueMap[modeKey]
	}
	panic(typeError(pos, "mode() requires an array"))
}

// stdDevFunc implements the std_dev() built-in function
func stdDevFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "std_dev", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) <= 1 {
			return Value(0.0)
		}

		// Calculate mean
		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "std_dev() array must contain only numbers"))
			}
		}
		mean := total / float64(len(*arr))

		// Calculate variance
		var sumSquaredDiff float64
		for _, v := range *arr {
			val := toFloat64(pos, v, "std_dev")
			diff := val - mean
			sumSquaredDiff += diff * diff
		}
		variance := sumSquaredDiff / float64(len(*arr)-1)

		return Value(math.Sqrt(variance))
	}
	panic(typeError(pos, "std_dev() requires an array"))
}

// varianceFunc implements the variance() built-in function
func varianceFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "variance", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) <= 1 {
			return Value(0.0)
		}

		// Calculate mean
		var total float64
		for _, v := range *arr {
			switch val := v.(type) {
			case int:
				total += float64(val)
			case float64:
				total += val
			default:
				panic(typeError(pos, "variance() array must contain only numbers"))
			}
		}
		mean := total / float64(len(*arr))

		// Calculate variance
		var sumSquaredDiff float64
		for _, v := range *arr {
			val := toFloat64(pos, v, "variance")
			diff := val - mean
			sumSquaredDiff += diff * diff
		}

		return Value(sumSquaredDiff / float64(len(*arr)-1))
	}
	panic(typeError(pos, "variance() requires an array"))
}

// ========================================
// Number Theory Functions
// ========================================

// gcdFunc implements the gcd() built-in function
func gcdFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "gcd", args, 2)
	a := toInt(pos, args[0], "gcd")
	b := toInt(pos, args[1], "gcd")

	a = int(math.Abs(float64(a)))
	b = int(math.Abs(float64(b)))

	for b != 0 {
		a, b = b, a%b
	}
	return Value(a)
}

// lcmFunc implements the lcm() built-in function
func lcmFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "lcm", args, 2)
	a := toInt(pos, args[0], "lcm")
	b := toInt(pos, args[1], "lcm")

	if a == 0 || b == 0 {
		return Value(0)
	}

	// Calculate GCD first
	gcdArgs := []Value{Value(a), Value(b)}
	gcdResult := gcdFunc(interp, pos, gcdArgs)
	gcd := gcdResult.(int)

	return Value(int(math.Abs(float64(a*b))) / gcd)
}

// factorialFunc implements the factorial() built-in function
func factorialFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "factorial", args, 1)
	n := toInt(pos, args[0], "factorial")

	if n < 0 {
		panic(valueError(pos, "factorial() of negative number"))
	}
	if n > 20 {
		panic(valueError(pos, "factorial() argument too large (max 20)"))
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return Value(result)
}

// fibonacciFunc implements the fibonacci() built-in function
func fibonacciFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "fibonacci", args, 1)
	n := toInt(pos, args[0], "fibonacci")

	if n < 0 {
		panic(valueError(pos, "fibonacci() of negative number"))
	}
	if n > 92 {
		panic(valueError(pos, "fibonacci() argument too large (max 92)"))
	}

	if n <= 1 {
		return Value(n)
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return Value(b)
}

// isPrimeFunc implements the is_prime() built-in function
func isPrimeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "is_prime", args, 1)
	n := toInt(pos, args[0], "is_prime")

	if n <= 1 {
		return Value(false)
	}
	if n <= 3 {
		return Value(true)
	}
	if n%2 == 0 || n%3 == 0 {
		return Value(false)
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return Value(false)
		}
	}
	return Value(true)
}

// primeFactorsFunc implements the prime_factors() built-in function
func primeFactorsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "prime_factors", args, 1)
	n := toInt(pos, args[0], "prime_factors")

	if n <= 1 {
		factors := make([]Value, 0)
		return Value(&factors)
	}

	factors := make([]Value, 0)

	// Handle factor of 2
	for n%2 == 0 {
		factors = append(factors, Value(2))
		n /= 2
	}

	// Handle odd factors
	for i := 3; i*i <= n; i += 2 {
		for n%i == 0 {
			factors = append(factors, Value(i))
			n /= i
		}
	}

	// If n is still > 1, then it's a prime
	if n > 1 {
		factors = append(factors, Value(n))
	}

	return Value(&factors)
}

// ========================================
// Random Number Functions
// ========================================

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// randomFunc implements the random() built-in function
func randomFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "random", args, 0)
	return Value(rng.Float64())
}

// randomIntFunc implements the random_int() built-in function
func randomIntFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "random_int", args, 2)
	min := toInt(pos, args[0], "random_int")
	max := toInt(pos, args[1], "random_int")

	if min >= max {
		panic(valueError(pos, "random_int() min must be less than max"))
	}

	return Value(rng.Intn(max-min) + min)
}

// randomFloatFunc implements the random_float() built-in function
func randomFloatFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "random_float", args, 2)
	min := toFloat64(pos, args[0], "random_float")
	max := toFloat64(pos, args[1], "random_float")

	if min >= max {
		panic(valueError(pos, "random_float() min must be less than max"))
	}

	return Value(rng.Float64()*(max-min) + min)
}

// randomChoiceFunc implements the random_choice() built-in function
func randomChoiceFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "random_choice", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		if len(*arr) == 0 {
			panic(valueError(pos, "random_choice() of empty array"))
		}

		index := rng.Intn(len(*arr))
		return (*arr)[index]
	}
	panic(typeError(pos, "random_choice() requires an array"))
}

// shuffleFunc implements the shuffle() built-in function
func shuffleFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "shuffle", args, 1)
	if arr, ok := args[0].(*[]Value); ok {
		// Fisher-Yates shuffle
		for i := len(*arr) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
		}
		return Value(nil)
	}
	panic(typeError(pos, "shuffle() requires an array"))
}

// seedRandomFunc implements the seed_random() built-in function
func seedRandomFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "seed_random", args, 1)
	seed := toInt(pos, args[0], "seed_random")
	rng = rand.New(rand.NewSource(int64(seed)))
	return Value(nil)
}

// ========================================
// Utility Functions
// ========================================

// signFunc implements the sign() built-in function
func signFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "sign", args, 1)
	val := toFloat64(pos, args[0], "sign")

	if val > 0 {
		return Value(1)
	} else if val < 0 {
		return Value(-1)
	}
	return Value(0)
}

// clampFunc implements the clamp() built-in function
func clampFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "clamp", args, 3)
	val := toFloat64(pos, args[0], "clamp")
	min := toFloat64(pos, args[1], "clamp")
	max := toFloat64(pos, args[2], "clamp")

	if min > max {
		panic(valueError(pos, "clamp() min must be less than or equal to max"))
	}

	if val < min {
		val = min
	} else if val > max {
		val = max
	}

	// Return int if all inputs were ints
	if _, ok := args[0].(int); ok {
		if _, ok := args[1].(int); ok {
			if _, ok := args[2].(int); ok {
				return Value(int(val))
			}
		}
	}
	return Value(val)
}

// lerpFunc implements the lerp() built-in function (linear interpolation)
func lerpFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "lerp", args, 3)
	a := toFloat64(pos, args[0], "lerp")
	b := toFloat64(pos, args[1], "lerp")
	t := toFloat64(pos, args[2], "lerp")

	return Value(a + t*(b-a))
}

// degreesFunc implements the degrees() built-in function
func degreesFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "degrees", args, 1)
	radians := toFloat64(pos, args[0], "degrees")
	return Value(radians * 180.0 / math.Pi)
}

// radiansFunc implements the radians() built-in function
func radiansFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "radians", args, 1)
	degrees := toFloat64(pos, args[0], "radians")
	return Value(degrees * math.Pi / 180.0)
}

// isNanFunc implements the is_nan() built-in function
func isNanFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "is_nan", args, 1)
	val := toFloat64(pos, args[0], "is_nan")
	return Value(math.IsNaN(val))
}

// isInfiniteFunc implements the is_infinite() built-in function
func isInfiniteFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "is_infinite", args, 1)
	val := toFloat64(pos, args[0], "is_infinite")
	return Value(math.IsInf(val, 0))
}
