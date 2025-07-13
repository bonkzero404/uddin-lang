package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
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

	// String Manipulation Functions (Point 1)
	"replace":        {replaceFunc, "replace"},
	"trim":           {trimFunc, "trim"},
	"starts_with":    {startsWithFunc, "starts_with"},
	"ends_with":      {endsWithFunc, "ends_with"},
	"repeat":         {repeatFunc, "repeat"},
	"reverse_str":    {reverseStrFunc, "reverse_str"},

	// Array Methods (Point 1)
	"map":            {mapFunc, "map"},
	"filter":         {filterFunc, "filter"},
	"reduce":         {reduceFunc, "reduce"},
	"reverse":        {reverseFunc, "reverse"},
	"push":           {pushFunc, "push"},
	"pop":            {popFunc, "pop"},
	"shift":          {shiftFunc, "shift"},
	"unshift":        {unshiftFunc, "unshift"},
	"index_of":       {indexOfFunc, "index_of"},
	"last_index_of":  {lastIndexOfFunc, "last_index_of"},

	// Data Structures (Point 2)
	"set_new":        {setNewFunc, "set_new"},
	"set_add":        {setAddFunc, "set_add"},
	"set_remove":     {setRemoveFunc, "set_remove"},
	"set_has":        {setHasFunc, "set_has"},
	"set_size":       {setSizeFunc, "set_size"},
	"set_to_array":   {setToArrayFunc, "set_to_array"},
	"stack_new":      {stackNewFunc, "stack_new"},
	"stack_push":     {stackPushFunc, "stack_push"},
	"stack_pop":      {stackPopFunc, "stack_pop"},
	"stack_peek":     {stackPeekFunc, "stack_peek"},
	"stack_size":     {stackSizeFunc, "stack_size"},
	"queue_new":      {queueNewFunc, "queue_new"},
	"queue_enqueue":  {queueEnqueueFunc, "queue_enqueue"},
	"queue_dequeue":  {queueDequeueFunc, "queue_dequeue"},
	"queue_front":    {queueFrontFunc, "queue_front"},
	"queue_size":     {queueSizeFunc, "queue_size"},

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

	// HTTP Client Functions
	"http_get":    {httpGetFunc, "http_get"},
	"http_post":   {httpPostFunc, "http_post"},
	"http_put":    {httpPutFunc, "http_put"},
	"http_delete": {httpDeleteFunc, "http_delete"},
	"http_request": {httpRequestFunc, "http_request"},

	// HTTP Server Functions
	"http_server_start": {httpServerStartFunc, "http_server_start"},
	"http_server_stop":  {httpServerStopFunc, "http_server_stop"},
	"http_server_route": {httpServerRouteFunc, "http_server_route"},
	"http_response":     {httpResponseFunc, "http_response"},

	// Network Functions
	"tcp_connect":   {tcpConnectFunc, "tcp_connect"},
	"tcp_listen":    {tcpListenFunc, "tcp_listen"},
	"tcp_accept":    {tcpAcceptFunc, "tcp_accept"},
	"tcp_read":      {tcpReadFunc, "tcp_read"},
	"tcp_write":     {tcpWriteFunc, "tcp_write"},
	"tcp_close":     {tcpCloseFunc, "tcp_close"},
	"udp_connect":   {udpConnectFunc, "udp_connect"},
	"udp_listen":    {udpListenFunc, "udp_listen"},
	"udp_read":      {udpReadFunc, "udp_read"},
	"udp_write":     {udpWriteFunc, "udp_write"},
	"udp_close":     {udpCloseFunc, "udp_close"},
	"net_resolve":   {netResolveFunc, "net_resolve"},
	"net_ping":      {netPingFunc, "net_ping"},
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

// ========================================
// String Manipulation Functions (Point 1)
// ========================================

// replaceFunc implements the replace() built-in function
// replace(str, old, new) -> string
// Example: replace("hello world", "world", "universe") -> "hello universe"
func replaceFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "replace", args, 3)
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
	ensureNumArgs(pos, "trim", args, 1)
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
	ensureNumArgs(pos, "starts_with", args, 2)
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
	ensureNumArgs(pos, "ends_with", args, 2)
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
	ensureNumArgs(pos, "repeat", args, 2)
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
	ensureNumArgs(pos, "reverse_str", args, 1)
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

// ========================================
// Array Methods (Point 1)
// ========================================

// mapFunc implements the map() built-in function
// map(array, function) -> array
// Example: map([1, 2, 3], lambda x: x * 2) -> [2, 4, 6]
func mapFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "map", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "map() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "map() requires second argument to be a function"))
	}

	result := make([]Value, len(*arr))
	for i, v := range *arr {
		result[i] = interp.callFunction(pos, fn, []Value{v})
	}
	return Value(&result)
}

// filterFunc implements the filter() built-in function
// filter(array, function) -> array
// Example: filter([1, 2, 3, 4], lambda x: x % 2 == 0) -> [2, 4]
func filterFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "filter", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "filter() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "filter() requires second argument to be a function"))
	}

	result := make([]Value, 0)
	for _, v := range *arr {
		if IsTruthy(interp.callFunction(pos, fn, []Value{v})) {
			result = append(result, v)
		}
	}
	return Value(&result)
}

// reduceFunc implements the reduce() built-in function
// reduce(array, function, initial) -> any
// Example: reduce([1, 2, 3, 4], lambda acc, x: acc + x, 0) -> 10
func reduceFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 3 {
		panic(typeError(pos, "reduce() requires 3 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "reduce() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "reduce() requires second argument to be a function"))
	}

	accumulator := args[2]
	for _, v := range *arr {
		accumulator = interp.callFunction(pos, fn, []Value{accumulator, v})
	}
	return accumulator
}

// reverseFunc implements the reverse() built-in function
// reverse(array) -> null (modifies array in place)
// Example: reverse([1, 2, 3]) -> [3, 2, 1]
func reverseFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "reverse", args, 1)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "reverse() requires an array argument"))
	}

	for i, j := 0, len(*arr)-1; i < j; i, j = i+1, j-1 {
		(*arr)[i], (*arr)[j] = (*arr)[j], (*arr)[i]
	}
	return Value(nil)
}

// pushFunc implements the push() built-in function
// push(array, element) -> null (modifies array in place)
// Example: push([1, 2], 3) -> [1, 2, 3]
func pushFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		panic(typeError(pos, "push() requires at least 2 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "push() requires first argument to be an array"))
	}

	*arr = append(*arr, args[1:]...)
	return Value(nil)
}

// popFunc implements the pop() built-in function
// pop(array) -> any (removes and returns last element)
// Example: pop([1, 2, 3]) -> 3, array becomes [1, 2]
func popFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "pop", args, 1)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "pop() requires an array argument"))
	}

	if len(*arr) == 0 {
		return Value(nil)
	}

	last := (*arr)[len(*arr)-1]
	*arr = (*arr)[:len(*arr)-1]
	return last
}

// shiftFunc implements the shift() built-in function
// shift(array) -> any (removes and returns first element)
// Example: shift([1, 2, 3]) -> 1, array becomes [2, 3]
func shiftFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "shift", args, 1)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "shift() requires an array argument"))
	}

	if len(*arr) == 0 {
		return Value(nil)
	}

	first := (*arr)[0]
	*arr = (*arr)[1:]
	return first
}

// unshiftFunc implements the unshift() built-in function
// unshift(array, element) -> null (adds element to beginning)
// Example: unshift([2, 3], 1) -> [1, 2, 3]
func unshiftFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		panic(typeError(pos, "unshift() requires at least 2 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "unshift() requires first argument to be an array"))
	}

	newArr := make([]Value, 0, len(*arr)+len(args)-1)
	newArr = append(newArr, args[1:]...)
	newArr = append(newArr, *arr...)
	*arr = newArr
	return Value(nil)
}

// indexOfFunc implements the index_of() built-in function
// index_of(array, element) -> int
// Example: index_of([1, 2, 3, 2], 2) -> 1
func indexOfFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "index_of", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "index_of() requires first argument to be an array"))
	}

	for i, v := range *arr {
		if evalEqual(pos, v, args[1]).(bool) {
			return Value(i)
		}
	}
	return Value(-1)
}

// lastIndexOfFunc implements the last_index_of() built-in function
// last_index_of(array, element) -> int
// Example: last_index_of([1, 2, 3, 2], 2) -> 3
func lastIndexOfFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "last_index_of", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "last_index_of() requires first argument to be an array"))
	}

	for i := len(*arr) - 1; i >= 0; i-- {
		if evalEqual(pos, (*arr)[i], args[1]).(bool) {
			return Value(i)
		}
	}
	return Value(-1)
}

// ========================================
// Data Structures (Point 2)
// ========================================

// Set implementation using map[string]Value for uniqueness
type Set struct {
	data map[string]Value
	keys []Value // Keep track of insertion order
}

// Stack implementation using slice
type Stack struct {
	data []Value
}

// Queue implementation using slice
type Queue struct {
	data []Value
}

// Helper function to convert Value to string key for Set
func valueToKey(v Value) string {
	switch val := v.(type) {
	case string:
		return "s:" + val
	case int:
		return fmt.Sprintf("i:%d", val)
	case float64:
		return fmt.Sprintf("f:%g", val)
	case bool:
		return fmt.Sprintf("b:%t", val)
	case nil:
		return "n:null"
	default:
		return fmt.Sprintf("o:%p", val)
	}
}

// setNewFunc implements the set_new() built-in function
// set_new() -> set
// Example: s = set_new()
func setNewFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_new", args, 0)
	return Value(&Set{
		data: make(map[string]Value),
		keys: make([]Value, 0),
	})
}

// setAddFunc implements the set_add() built-in function
// set_add(set, element) -> null
// Example: set_add(s, "hello")
func setAddFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_add", args, 2)
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_add() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	if _, exists := set.data[key]; !exists {
		set.data[key] = args[1]
		set.keys = append(set.keys, args[1])
	}
	return Value(nil)
}

// setRemoveFunc implements the set_remove() built-in function
// set_remove(set, element) -> bool
// Example: set_remove(s, "hello") -> true if removed, false if not found
func setRemoveFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_remove", args, 2)
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_remove() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	if _, exists := set.data[key]; exists {
		delete(set.data, key)
		// Remove from keys slice
		for i, v := range set.keys {
			if evalEqual(pos, v, args[1]).(bool) {
				set.keys = append(set.keys[:i], set.keys[i+1:]...)
				break
			}
		}
		return Value(true)
	}
	return Value(false)
}

// setHasFunc implements the set_has() built-in function
// set_has(set, element) -> bool
// Example: set_has(s, "hello") -> true if exists
func setHasFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_has", args, 2)
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_has() requires first argument to be a set"))
	}

	key := valueToKey(args[1])
	_, exists := set.data[key]
	return Value(exists)
}

// setSizeFunc implements the set_size() built-in function
// set_size(set) -> int
// Example: set_size(s) -> 3
func setSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_size", args, 1)
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_size() requires a set argument"))
	}

	return Value(len(set.data))
}

// setToArrayFunc implements the set_to_array() built-in function
// set_to_array(set) -> array
// Example: set_to_array(s) -> ["hello", "world"]
func setToArrayFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "set_to_array", args, 1)
	set, ok := args[0].(*Set)
	if !ok {
		panic(typeError(pos, "set_to_array() requires a set argument"))
	}

	result := make([]Value, len(set.keys))
	copy(result, set.keys)
	return Value(&result)
}

// stackNewFunc implements the stack_new() built-in function
// stack_new() -> stack
// Example: s = stack_new()
func stackNewFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "stack_new", args, 0)
	return Value(&Stack{
		data: make([]Value, 0),
	})
}

// stackPushFunc implements the stack_push() built-in function
// stack_push(stack, element) -> null
// Example: stack_push(s, "hello")
func stackPushFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "stack_push", args, 2)
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_push() requires first argument to be a stack"))
	}

	stack.data = append(stack.data, args[1])
	return Value(nil)
}

// stackPopFunc implements the stack_pop() built-in function
// stack_pop(stack) -> any
// Example: stack_pop(s) -> "hello"
func stackPopFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "stack_pop", args, 1)
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_pop() requires a stack argument"))
	}

	if len(stack.data) == 0 {
		return Value(nil)
	}

	last := stack.data[len(stack.data)-1]
	stack.data = stack.data[:len(stack.data)-1]
	return last
}

// stackPeekFunc implements the stack_peek() built-in function
// stack_peek(stack) -> any
// Example: stack_peek(s) -> "hello" (without removing)
func stackPeekFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "stack_peek", args, 1)
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_peek() requires a stack argument"))
	}

	if len(stack.data) == 0 {
		return Value(nil)
	}

	return stack.data[len(stack.data)-1]
}

// stackSizeFunc implements the stack_size() built-in function
// stack_size(stack) -> int
// Example: stack_size(s) -> 3
func stackSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "stack_size", args, 1)
	stack, ok := args[0].(*Stack)
	if !ok {
		panic(typeError(pos, "stack_size() requires a stack argument"))
	}

	return Value(len(stack.data))
}

// queueNewFunc implements the queue_new() built-in function
// queue_new() -> queue
// Example: q = queue_new()
func queueNewFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "queue_new", args, 0)
	return Value(&Queue{
		data: make([]Value, 0),
	})
}

// queueEnqueueFunc implements the queue_enqueue() built-in function
// queue_enqueue(queue, element) -> null
// Example: queue_enqueue(q, "hello")
func queueEnqueueFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "queue_enqueue", args, 2)
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_enqueue() requires first argument to be a queue"))
	}

	queue.data = append(queue.data, args[1])
	return Value(nil)
}

// queueDequeueFunc implements the queue_dequeue() built-in function
// queue_dequeue(queue) -> any
// Example: queue_dequeue(q) -> "hello"
func queueDequeueFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "queue_dequeue", args, 1)
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_dequeue() requires a queue argument"))
	}

	if len(queue.data) == 0 {
		return Value(nil)
	}

	first := queue.data[0]
	queue.data = queue.data[1:]
	return first
}

// queueFrontFunc implements the queue_front() built-in function
// queue_front(queue) -> any
// Example: queue_front(q) -> "hello" (without removing)
func queueFrontFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "queue_front", args, 1)
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_front() requires a queue argument"))
	}

	if len(queue.data) == 0 {
		return Value(nil)
	}

	return queue.data[0]
}

// queueSizeFunc implements the queue_size() built-in function
// queue_size(queue) -> int
// Example: queue_size(q) -> 3
func queueSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "queue_size", args, 1)
	queue, ok := args[0].(*Queue)
	if !ok {
		panic(typeError(pos, "queue_size() requires a queue argument"))
	}

	return Value(len(queue.data))
}

// HTTP Client Functions

// httpGetFunc implements the http_get() built-in function
// http_get(url) -> object
// Example: response = http_get("https://api.example.com/data")
func httpGetFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "http_get", args, 1)
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_get() requires a string URL"))
	}

	return httpRequest(pos, "GET", url, nil, nil)
}

// httpPostFunc implements the http_post() built-in function
// http_post(url, data) -> object
// Example: response = http_post("https://api.example.com/data", {"name": "John"})
func httpPostFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "http_post() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_post() requires first argument to be a string URL"))
	}

	var headers map[string]Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_post() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, "POST", url, args[1], headers)
}

// httpPutFunc implements the http_put() built-in function
// http_put(url, data) -> object
// Example: response = http_put("https://api.example.com/data/1", {"name": "Jane"})
func httpPutFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "http_put() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_put() requires first argument to be a string URL"))
	}

	var headers map[string]Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_put() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, "PUT", url, args[1], headers)
}

// httpDeleteFunc implements the http_delete() built-in function
// http_delete(url) -> object
// Example: response = http_delete("https://api.example.com/data/1")
func httpDeleteFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "http_delete", args, 1)
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_delete() requires a string URL"))
	}

	return httpRequest(pos, "DELETE", url, nil, nil)
}

// httpRequestFunc implements the http_request() built-in function
// http_request(method, url, data, headers) -> object
// Example: response = http_request("PATCH", "https://api.example.com/data/1", {"status": "active"}, {"Authorization": "Bearer token"})
func httpRequestFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		panic(typeError(pos, "http_request() requires 2 to 4 args, got %d", len(args)))
	}

	method, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_request() requires first argument to be a string (method)"))
	}

	url, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "http_request() requires second argument to be a string (URL)"))
	}

	var data Value
	var headers map[string]Value

	if len(args) >= 3 {
		data = args[2]
	}

	if len(args) == 4 {
		headersMap, ok := args[3].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_request() requires fourth argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, method, url, data, headers)
}

// httpRequest is a helper function that performs the actual HTTP request
func httpRequest(pos Position, method, url string, data Value, headers map[string]Value) Value {
	var body io.Reader

	// Prepare request body if data is provided
	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
		case map[string]Value, []Value:
			// Convert to JSON
			jsonData, err := valueToJSON(data)
			if err != nil {
				panic(valueError(pos, "Failed to convert data to JSON: %v", err))
			}
			body = bytes.NewReader(jsonData)
		default:
			// Convert other types to string
			body = strings.NewReader(fmt.Sprintf("%v", data))
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(valueError(pos, "Failed to create HTTP request: %v", err))
	}

	// Set default Content-Type for POST/PUT requests with JSON data
	if (method == "POST" || method == "PUT" || method == "PATCH") && data != nil {
		if _, isMap := data.(map[string]Value); isMap {
			req.Header.Set("Content-Type", "application/json")
		} else if _, isArray := data.([]Value); isArray {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Set custom headers
	if headers != nil {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Perform the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(valueError(pos, "HTTP request failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(valueError(pos, "Failed to read response body: %v", err))
	}

	// Parse response body as JSON if possible
	var bodyValue Value
	var jsonData interface{}
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		bodyValue = jsonToValue(jsonData)
	} else {
		// If not JSON, return as string
		bodyValue = string(respBody)
	}

	// Convert headers to map
	headerMap := make(map[string]Value)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	// Create response object
	response := map[string]Value{
		"status":     resp.StatusCode,
		"status_text": resp.Status,
		"headers":    headerMap,
		"body":       bodyValue,
		"url":        resp.Request.URL.String(),
	}

	return Value(response)
}

// valueToJSON converts a Value to JSON bytes
func valueToJSON(v Value) ([]byte, error) {
	switch val := v.(type) {
	case map[string]Value:
		// Convert map to Go map for JSON marshaling
		goMap := make(map[string]interface{})
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return json.Marshal(goMap)
	case []Value:
		// Convert array to Go slice for JSON marshaling
		goSlice := make([]interface{}, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return json.Marshal(goSlice)
	default:
		return json.Marshal(valueToInterface(v))
	}
}

// valueToInterface converts a Value to interface{} for JSON marshaling
func valueToInterface(v Value) interface{} {
	switch val := v.(type) {
	case map[string]Value:
		goMap := make(map[string]interface{})
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return goMap
	case []Value:
		goSlice := make([]interface{}, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return goSlice
	case nil:
		return nil
	default:
		return val
	}
}

// jsonToValue converts JSON interface{} to Value
func jsonToValue(data interface{}) Value {
	switch val := data.(type) {
	case map[string]interface{}:
		result := make(map[string]Value)
		for k, v := range val {
			result[k] = jsonToValue(v)
		}
		return Value(result)
	case []interface{}:
		result := make([]Value, len(val))
		for i, v := range val {
			result[i] = jsonToValue(v)
		}
		return Value(result)
	case float64:
		// Check if it's actually an integer
		if val == float64(int(val)) {
			return Value(int(val))
		}
		return Value(val)
	case bool:
		return Value(val)
	case string:
		return Value(val)
	case nil:
		return Value(val)
	default:
		return Value(val)
	}
}

// HTTP Server Functions

// Global HTTP server registry to manage multiple servers
var httpServers = make(map[string]*http.Server)
var httpRoutes = make(map[string]map[string]functionType) // serverID -> route -> handler

// httpServerStartFunc implements the http_server_start() built-in function
// http_server_start(port, server_id?) -> server_object
// Example: server = http_server_start(8080, "main")
func httpServerStartFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		panic(typeError(pos, "http_server_start() requires 1 or 2 args, got %d", len(args)))
	}

	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "http_server_start() requires first argument to be an integer (port)"))
	}

	serverID := "default"
	if len(args) == 2 {
		if id, ok := args[1].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_start() requires second argument to be a string (server_id)"))
		}
	}

	// Check if server already exists
	if _, exists := httpServers[serverID]; exists {
		panic(valueError(pos, "HTTP server with ID '%s' already exists", serverID))
	}

	// Initialize routes for this server
	if httpRoutes[serverID] == nil {
		httpRoutes[serverID] = make(map[string]functionType)
	}

	// Create HTTP server
	address := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()

	// Default handler that looks up routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Look for exact route match first
		routeKey := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if handler, exists := httpRoutes[serverID][routeKey]; exists {
			// Create request object for the handler
			reqObj := map[string]Value{
				"method": r.Method,
				"path":   r.URL.Path,
				"query":  r.URL.RawQuery,
				"headers": convertHeaders(r.Header),
				"body":    readRequestBody(r),
			}

			// Create response object
			resObj := map[string]Value{
				"status":  200,
				"headers": make(map[string]Value),
				"body":    "",
				"_writer": w,
			}

			// Call the handler
			handler.call(interp, pos, []Value{reqObj, resObj})
			return
		}

		// No route found
		w.WriteHeader(404)
		w.Write([]byte("404 Not Found"))
	})

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	// Store server
	httpServers[serverID] = server

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Return server object
	serverObj := map[string]Value{
		"type":      "http_server",
		"server_id": serverID,
		"port":      port,
		"address":   address,
		"_server":   server,
	}
	return Value(serverObj)
}

// httpServerStopFunc implements the http_server_stop() built-in function
// http_server_stop(server_id?) -> null
// Example: http_server_stop("main")
func httpServerStopFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) > 1 {
		panic(typeError(pos, "http_server_stop() requires 0 or 1 args, got %d", len(args)))
	}

	serverID := "default"
	if len(args) == 1 {
		if id, ok := args[0].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_stop() requires argument to be a string (server_id)"))
		}
	}

	// Find and stop server
	if server, exists := httpServers[serverID]; exists {
		server.Close()
		delete(httpServers, serverID)
		delete(httpRoutes, serverID)
	} else {
		panic(valueError(pos, "HTTP server with ID '%s' not found", serverID))
	}

	return Value(nil)
}

// httpServerRouteFunc implements the http_server_route() built-in function
// http_server_route(method, path, handler, server_id?) -> null
// Example: http_server_route("GET", "/api/users", my_handler, "main")
func httpServerRouteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 3 || len(args) > 4 {
		panic(typeError(pos, "http_server_route() requires 3 or 4 args, got %d", len(args)))
	}

	method, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_server_route() requires first argument to be a string (method)"))
	}

	path, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "http_server_route() requires second argument to be a string (path)"))
	}

	handler, ok := args[2].(functionType)
	if !ok {
		panic(typeError(pos, "http_server_route() requires third argument to be a function (handler)"))
	}

	serverID := "default"
	if len(args) == 4 {
		if id, ok := args[3].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_route() requires fourth argument to be a string (server_id)"))
		}
	}

	// Check if server exists
	if _, exists := httpServers[serverID]; !exists {
		panic(valueError(pos, "HTTP server with ID '%s' not found. Start server first.", serverID))
	}

	// Register route
	routeKey := fmt.Sprintf("%s %s", method, path)
	httpRoutes[serverID][routeKey] = handler

	return Value(nil)
}

// httpResponseFunc implements the http_response() built-in function
// http_response(response_obj, status?, headers?, body?) -> null
// Example: http_response(res, 200, {"Content-Type": "application/json"}, '{"message": "Hello"}')
func httpResponseFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 || len(args) > 4 {
		panic(typeError(pos, "http_response() requires 1 to 4 args, got %d", len(args)))
	}

	resObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "http_response() requires first argument to be a response object"))
	}

	w, ok := resObj["_writer"].(http.ResponseWriter)
	if !ok {
		panic(typeError(pos, "http_response() requires a valid response object"))
	}

	// Set status code
	status := 200
	if len(args) >= 2 {
		if s, ok := args[1].(int); ok {
			status = s
		} else {
			panic(typeError(pos, "http_response() requires second argument to be an integer (status)"))
		}
	}

	// Set headers
	if len(args) >= 3 {
		if headers, ok := args[2].(map[string]Value); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					w.Header().Set(key, strValue)
				}
			}
		} else {
			panic(typeError(pos, "http_response() requires third argument to be an object (headers)"))
		}
	}

	// Set body
	body := ""
	if len(args) >= 4 {
		switch b := args[3].(type) {
		case string:
			body = b
		case map[string]Value, []Value:
			// Convert to JSON
			jsonData, err := valueToJSON(args[3])
			if err != nil {
				panic(valueError(pos, "Failed to convert body to JSON: %v", err))
			}
			body = string(jsonData)
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json")
			}
		default:
			body = fmt.Sprintf("%v", args[3])
		}
	}

	// Write response
	w.WriteHeader(status)
	w.Write([]byte(body))

	return Value(nil)
}

// Helper functions for HTTP server

func convertHeaders(headers http.Header) Value {
	headerMap := make(map[string]Value)
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	return Value(headerMap)
}

func readRequestBody(r *http.Request) Value {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Value("")
	}
	r.Body.Close()
	return Value(string(body))
}

// TCP Connection Functions

// tcpConnectFunc implements the tcp_connect() built-in function
// tcp_connect(host, port) -> connection_object
// Example: conn = tcp_connect("localhost", 8080)
func tcpConnectFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_connect", args, 2)
	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "tcp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "tcp_connect() requires second argument to be an integer (port)"))
	}

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(valueError(pos, "Failed to connect to %s: %v", address, err))
	}

	// Return connection object
	connObj := map[string]Value{
		"type":    "tcp_connection",
		"address": address,
		"_conn":   conn, // Internal connection object
	}
	return Value(connObj)
}

// tcpListenFunc implements the tcp_listen() built-in function
// tcp_listen(port) -> listener_object
// Example: listener = tcp_listen(8080)
func tcpListenFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_listen", args, 1)
	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "tcp_listen() requires an integer (port)"))
	}

	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(valueError(pos, "Failed to listen on port %d: %v", port, err))
	}

	// Return listener object
	listenerObj := map[string]Value{
		"type":      "tcp_listener",
		"address":   address,
		"_listener": listener, // Internal listener object
	}
	return Value(listenerObj)
}

// tcpAcceptFunc implements the tcp_accept() built-in function
// tcp_accept(listener) -> connection_object
// Example: conn = tcp_accept(listener)
func tcpAcceptFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_accept", args, 1)
	listenerObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_accept() requires a listener object"))
	}

	listener, ok := listenerObj["_listener"].(net.Listener)
	if !ok {
		panic(typeError(pos, "tcp_accept() requires a valid TCP listener"))
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(valueError(pos, "Failed to accept connection: %v", err))
	}

	// Return connection object
	connObj := map[string]Value{
		"type":    "tcp_connection",
		"address": conn.RemoteAddr().String(),
		"_conn":   conn, // Internal connection object
	}
	return Value(connObj)
}

// tcpReadFunc implements the tcp_read() built-in function
// tcp_read(connection, size) -> string
// Example: data = tcp_read(conn, 1024)
func tcpReadFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_read", args, 2)
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_read() requires a connection object"))
	}

	conn, ok := connObj["_conn"].(net.Conn)
	if !ok {
		panic(typeError(pos, "tcp_read() requires a valid TCP connection"))
	}

	size, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "tcp_read() requires second argument to be an integer (size)"))
	}

	buffer := make([]byte, size)
	n, err := conn.Read(buffer)
	if err != nil {
		panic(valueError(pos, "Failed to read from connection: %v", err))
	}

	return Value(string(buffer[:n]))
}

// tcpWriteFunc implements the tcp_write() built-in function
// tcp_write(connection, data) -> bytes_written
// Example: written = tcp_write(conn, "Hello, World!")
func tcpWriteFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_write", args, 2)
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_write() requires a connection object"))
	}

	conn, ok := connObj["_conn"].(net.Conn)
	if !ok {
		panic(typeError(pos, "tcp_write() requires a valid TCP connection"))
	}

	data, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "tcp_write() requires second argument to be a string (data)"))
	}

	n, err := conn.Write([]byte(data))
	if err != nil {
		panic(valueError(pos, "Failed to write to connection: %v", err))
	}

	return Value(n)
}

// tcpCloseFunc implements the tcp_close() built-in function
// tcp_close(connection_or_listener) -> null
// Example: tcp_close(conn)
func tcpCloseFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "tcp_close", args, 1)
	obj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_close() requires a connection or listener object"))
	}

	// Try to close as connection first
	if conn, ok := obj["_conn"].(net.Conn); ok {
		err := conn.Close()
		if err != nil {
			panic(valueError(pos, "Failed to close connection: %v", err))
		}
		return Value(nil)
	}

	// Try to close as listener
	if listener, ok := obj["_listener"].(net.Listener); ok {
		err := listener.Close()
		if err != nil {
			panic(valueError(pos, "Failed to close listener: %v", err))
		}
		return Value(nil)
	}

	panic(typeError(pos, "tcp_close() requires a valid TCP connection or listener"))
}

// UDP Connection Functions

// udpConnectFunc implements the udp_connect() built-in function
// udp_connect(host, port) -> connection_object
// Example: conn = udp_connect("localhost", 8080)
func udpConnectFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "udp_connect", args, 2)
	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "udp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "udp_connect() requires second argument to be an integer (port)"))
	}

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("udp", address)
	if err != nil {
		panic(valueError(pos, "Failed to connect to %s: %v", address, err))
	}

	// Return connection object
	connObj := map[string]Value{
		"type":    "udp_connection",
		"address": address,
		"_conn":   conn, // Internal connection object
	}
	return Value(connObj)
}

// udpListenFunc implements the udp_listen() built-in function
// udp_listen(port) -> connection_object
// Example: conn = udp_listen(8080)
func udpListenFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "udp_listen", args, 1)
	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "udp_listen() requires an integer (port)"))
	}

	address := fmt.Sprintf(":%d", port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(valueError(pos, "Failed to resolve UDP address %s: %v", address, err))
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(valueError(pos, "Failed to listen on UDP port %d: %v", port, err))
	}

	// Return connection object
	connObj := map[string]Value{
		"type":    "udp_connection",
		"address": address,
		"_conn":   conn, // Internal connection object
	}
	return Value(connObj)
}

// udpReadFunc implements the udp_read() built-in function
// udp_read(connection, size) -> {"data": string, "address": string}
// Example: result = udp_read(conn, 1024)
func udpReadFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "udp_read", args, 2)
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_read() requires a connection object"))
	}

	size, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "udp_read() requires second argument to be an integer (size)"))
	}

	buffer := make([]byte, size)
	var n int
	var addr net.Addr
	var err error

	// Handle both UDP connection types
	if udpConn, ok := connObj["_conn"].(*net.UDPConn); ok {
		n, addr, err = udpConn.ReadFromUDP(buffer)
	} else if conn, ok := connObj["_conn"].(net.Conn); ok {
		n, err = conn.Read(buffer)
		addr = conn.RemoteAddr()
	} else {
		panic(typeError(pos, "udp_read() requires a valid UDP connection"))
	}

	if err != nil {
		panic(valueError(pos, "Failed to read from UDP connection: %v", err))
	}

	result := map[string]Value{
		"data":    string(buffer[:n]),
		"address": addr.String(),
	}
	return Value(result)
}

// udpWriteFunc implements the udp_write() built-in function
// udp_write(connection, data, address?) -> bytes_written
// Example: written = udp_write(conn, "Hello, World!", "192.168.1.100:8080")
func udpWriteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "udp_write() requires 2 or 3 args, got %d", len(args)))
	}

	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_write() requires a connection object"))
	}

	data, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "udp_write() requires second argument to be a string (data)"))
	}

	var n int
	var err error

	if len(args) == 3 {
		// Write to specific address
		addrStr, ok := args[2].(string)
		if !ok {
			panic(typeError(pos, "udp_write() requires third argument to be a string (address)"))
		}

		udpConn, ok := connObj["_conn"].(*net.UDPConn)
		if !ok {
			panic(typeError(pos, "udp_write() with address requires a UDP listener connection"))
		}

		addr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			panic(valueError(pos, "Failed to resolve UDP address %s: %v", addrStr, err))
		}

		n, err = udpConn.WriteToUDP([]byte(data), addr)
	} else {
		// Write to connected address
		conn, ok := connObj["_conn"].(net.Conn)
		if !ok {
			panic(typeError(pos, "udp_write() requires a valid UDP connection"))
		}

		n, err = conn.Write([]byte(data))
	}

	if err != nil {
		panic(valueError(pos, "Failed to write to UDP connection: %v", err))
	}

	return Value(n)
}

// udpCloseFunc implements the udp_close() built-in function
// udp_close(connection) -> null
// Example: udp_close(conn)
func udpCloseFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "udp_close", args, 1)
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_close() requires a connection object"))
	}

	conn, ok := connObj["_conn"].(net.Conn)
	if !ok {
		panic(typeError(pos, "udp_close() requires a valid UDP connection"))
	}

	err := conn.Close()
	if err != nil {
		panic(valueError(pos, "Failed to close UDP connection: %v", err))
	}

	return Value(nil)
}

// Network Utility Functions

// netResolveFunc implements the net_resolve() built-in function
// net_resolve(hostname) -> array of IP addresses
// Example: ips = net_resolve("google.com")
func netResolveFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "net_resolve", args, 1)
	hostname, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "net_resolve() requires a string (hostname)"))
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		panic(valueError(pos, "Failed to resolve hostname %s: %v", hostname, err))
	}

	// Convert IPs to string array
	ipStrings := make([]Value, len(ips))
	for i, ip := range ips {
		ipStrings[i] = Value(ip.String())
	}

	return Value(&ipStrings)
}

// netPingFunc implements the net_ping() built-in function
// net_ping(host, port, timeout_ms?) -> {"success": bool, "time_ms": int, "error": string?}
// Example: result = net_ping("google.com", 80, 5000)
func netPingFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "net_ping() requires 2 or 3 args, got %d", len(args)))
	}

	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "net_ping() requires first argument to be a string (host)"))
	}

	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "net_ping() requires second argument to be an integer (port)"))
	}

	timeoutMs := 5000 // Default 5 seconds
	if len(args) == 3 {
		if timeout, ok := args[2].(int); ok {
			timeoutMs = timeout
		} else {
			panic(typeError(pos, "net_ping() requires third argument to be an integer (timeout_ms)"))
		}
	}

	address := fmt.Sprintf("%s:%d", host, port)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	elapsed := time.Since(start)

	result := map[string]Value{
		"time_ms": int(elapsed.Nanoseconds() / 1000000),
	}

	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
	} else {
		result["success"] = true
		conn.Close()
	}

	return Value(result)
}
