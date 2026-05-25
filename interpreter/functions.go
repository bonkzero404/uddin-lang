package interpreter

import (
	"fmt"
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
// EXPERIMENTAL: Memoization feature is experimental and not recommended for production
type userFunction struct {
	Name       string           // Function name (can be empty for anonymous functions)
	Parameters []string         // Parameter names
	Ellipsis   bool             // Whether the last parameter is variadic
	Body       Block            // Function body statements
	Closure    map[string]Value // Captured variables from outer scopes
	// EXPERIMENTAL: Whether this function should use memoization (not production-ready)
	// WARNING: Memoization may consume significant memory and is not thread-safe
	Memoized bool
}

// ensureNumArgs checks if the number of arguments matches the required count
// If not, it panics with a type error indicating the mismatch
// Parameters:
//   - pos: Position in source code for error reporting
//   - name: Function name for error message
//   - args: Actual arguments passed
//   - required: Required number of arguments
func ensureNumArgs(pos Position, name string, args []Value, required int) Value {
	if len(args) != required {
		plural := ""
		if required != 1 {
			plural = "s"
		}
		panic(typeError(pos, "%s() requires %d arg%s, got %d", name, required, plural, len(args)))
	}
	return Value(nil)
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
	// Use optimized call stack for function call tracking
	callStack := GetCallStack()
	defer PutCallStack(callStack)

	// Push function call info
	callStack.PushCall(map[string]any{
		"function":   f.Name,
		"position":   pos,
		"args_count": len(args),
	})
	defer callStack.PopCall()

	// Handle variadic arguments if this is a variadic function
	if f.Ellipsis {
		ellipsisArgs := args[len(f.Parameters)-1:]
		newArgs := make([]Value, 0, len(f.Parameters)+1)
		newArgs = SmartAppend(newArgs, args[:len(f.Parameters)-1]...)
		args = SmartAppend(newArgs, Value(&ellipsisArgs))
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

	// Handle return statements
	var returnValue Value = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				if ret, ok := r.(returnResult); ok {
					returnValue = ret.value
				} else {
					// Re-panic if it's not a return statement
					panic(r)
				}
			}
		}()
		// Execute the function body
		interp.executeBlock(f.Body)
	}()

	return returnValue
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
	// Core System Functions
	"import":  {importFunc, "import"},
	"exit":    {exitFunc, "exit"},
	"print":   {printFunc, "print"},
	"println": {printlnFunc, "println"},
	"input":   {inputFunc, "input"},
	"typeof":  {typeofFunc, "typeof"},

	// Type Conversion Functions
	"int":   {intFunc, "int"},
	"float": {floatFunc, "float"},
	"str":   {strFunc, "str"},
	"char":  {charFunc, "char"},
	"rune":  {runeFunc, "rune"},

	// Basic Array/Collection Functions
	"append": {appendFunc, "append"},
	"len":    {lenFunc, "len"},
	"range":  {rangeFunc, "range"},
	"slice":  {sliceFunc, "slice"},
	"sort":   {sortFunc, "sort"},

	// String Manipulation Functions - Basic
	"join":     {joinFunc, "join"},
	"split":    {splitFunc, "split"},
	"lower":    {lowerFunc, "lower"},
	"upper":    {upperFunc, "upper"},
	"contains": {containsFunc, "contains"},
	"str_pad":  {strpadFunc, "str_pad"},
	"substr":   {substrFunc, "substr"},
	"find":     {findFunc, "find"},

	// String Manipulation Functions - Advanced
	"replace":     {replaceFunc, "replace"},
	"trim":        {trimFunc, "trim"},
	"starts_with": {startsWithFunc, "starts_with"},
	"ends_with":   {endsWithFunc, "ends_with"},
	"repeat":      {repeatFunc, "repeat"},
	"reverse_str": {reverseStrFunc, "reverse_str"},

	// Array/Collection Methods - Advanced
	"map":               {mapFunc, "map"},
	"filter":            {filterFunc, "filter"},
	"reduce":            {reduceFunc, "reduce"},
	"concurrent_map":    {concurrentMapFunc, "concurrent_map"},
	"concurrent_filter": {concurrentFilterFunc, "concurrent_filter"},
	"concurrent_reduce": {concurrentReduceFunc, "concurrent_reduce"},
	"reverse":           {reverseFunc, "reverse"},
	"push":              {pushFunc, "push"},
	"pop":               {popFunc, "pop"},
	"shift":             {shiftFunc, "shift"},
	"unshift":           {unshiftFunc, "unshift"},
	"index_of":          {indexOfFunc, "index_of"},
	"last_index_of":     {lastIndexOfFunc, "last_index_of"},

	// Data Structures - Set Operations
	"set_new":      {setNewFunc, "set_new"},
	"set_add":      {setAddFunc, "set_add"},
	"set_remove":   {setRemoveFunc, "set_remove"},
	"set_has":      {setHasFunc, "set_has"},
	"set_size":     {setSizeFunc, "set_size"},
	"set_to_array": {setToArrayFunc, "set_to_array"},

	// Data Structures - Stack Operations
	"stack_new":   {stackNewFunc, "stack_new"},
	"stack_push":  {stackPushFunc, "stack_push"},
	"stack_pop":   {stackPopFunc, "stack_pop"},
	"stack_peek":  {stackPeekFunc, "stack_peek"},
	"stack_size":  {stackSizeFunc, "stack_size"},
	"stack_empty": {stackEmptyFunc, "stack_empty"},

	// Data Structures - Queue Operations
	"queue_new":     {queueNewFunc, "queue_new"},
	"queue_enqueue": {queueEnqueueFunc, "queue_enqueue"},
	"queue_dequeue": {queueDequeueFunc, "queue_dequeue"},
	"queue_front":   {queueFrontFunc, "queue_front"},
	"queue_size":    {queueSizeFunc, "queue_size"},
	"queue_empty":   {queueEmptyFunc, "queue_empty"},

	// Mathematical Functions - Basic Operations
	"abs":  {absFunc, "abs"},
	"max":  {maxFunc, "max"},
	"min":  {minFunc, "min"},
	"pow":  {powFunc, "pow"},
	"sqrt": {sqrtFunc, "sqrt"},
	"cbrt": {cbrtFunc, "cbrt"},

	// Mathematical Functions - Rounding
	"round": {roundFunc, "round"},
	"floor": {floorFunc, "floor"},
	"ceil":  {ceilFunc, "ceil"},
	"trunc": {truncFunc, "trunc"},

	// Mathematical Functions - Trigonometric
	"sin":   {sinFunc, "sin"},
	"cos":   {cosFunc, "cos"},
	"tan":   {tanFunc, "tan"},
	"asin":  {asinFunc, "asin"},
	"acos":  {acosFunc, "acos"},
	"atan":  {atanFunc, "atan"},
	"atan2": {atan2Func, "atan2"},

	// Mathematical Functions - Hyperbolic
	"sinh": {sinhFunc, "sinh"},
	"cosh": {coshFunc, "cosh"},
	"tanh": {tanhFunc, "tanh"},

	// Mathematical Functions - Logarithmic & Exponential
	"log":   {logFunc, "log"},
	"log10": {log10Func, "log10"},
	"log2":  {log2Func, "log2"},
	"logb":  {logbFunc, "logb"},
	"exp":   {expFunc, "exp"},
	"exp2":  {exp2Func, "exp2"},

	// Mathematical Functions - Statistical
	"sum":      {sumFunc, "sum"},
	"mean":     {meanFunc, "mean"},
	"median":   {medianFunc, "median"},
	"mode":     {modeFunc, "mode"},
	"std_dev":  {stdDevFunc, "std_dev"},
	"variance": {varianceFunc, "variance"},

	// Mathematical Functions - Number Theory
	"gcd":           {gcdFunc, "gcd"},
	"lcm":           {lcmFunc, "lcm"},
	"factorial":     {factorialFunc, "factorial"},
	"fibonacci":     {fibonacciFunc, "fibonacci"},
	"is_prime":      {isPrimeFunc, "is_prime"},
	"prime_factors": {primeFactorsFunc, "prime_factors"},

	// Mathematical Functions - Random Numbers
	"random":        {randomFunc, "random"},
	"random_int":    {randomIntFunc, "random_int"},
	"random_float":  {randomFloatFunc, "random_float"},
	"random_choice": {randomChoiceFunc, "random_choice"},
	"shuffle":       {shuffleFunc, "shuffle"},
	"seed_random":   {seedRandomFunc, "seed_random"},

	// Mathematical Functions - Utility
	"sign":        {signFunc, "sign"},
	"clamp":       {clampFunc, "clamp"},
	"lerp":        {lerpFunc, "lerp"},
	"degrees":     {degreesFunc, "degrees"},
	"radians":     {radiansFunc, "radians"},
	"is_nan":      {isNanFunc, "is_nan"},
	"is_infinite": {isInfiniteFunc, "is_infinite"},

	// Data Serialization - JSON
	"json_parse":     {jsonParseFunc, "json_parse"},
	"json_stringify": {jsonStringifyFunc, "json_stringify"},

	// Data Serialization - XML
	"xml_parse":     {xmlParseFunc, "xml_parse"},
	"xml_stringify": {xmlStringifyFunc, "xml_stringify"},

	// Network Functions - HTTP Client
	"http_get":     {httpGetFunc, "http_get"},
	"http_post":    {httpPostFunc, "http_post"},
	"http_put":     {httpPutFunc, "http_put"},
	"http_delete":  {httpDeleteFunc, "http_delete"},
	"http_request": {httpRequestFunc, "http_request"},

	// Network Functions - HTTP Server
	"http_server_start": {httpServerStartFunc, "http_server_start"},
	"http_server_stop":  {httpServerStopFunc, "http_server_stop"},
	"http_server_route": {httpServerRouteFunc, "http_server_route"},
	"http_response":     {httpResponseFunc, "http_response"},

	// Network Functions - TCP
	"tcp_connect": {tcpConnectFunc, "tcp_connect"},
	"tcp_listen":  {tcpListenFunc, "tcp_listen"},
	"tcp_accept":  {tcpAcceptFunc, "tcp_accept"},
	"tcp_read":    {tcpReadFunc, "tcp_read"},
	"tcp_write":   {tcpWriteFunc, "tcp_write"},
	"tcp_close":   {tcpCloseFunc, "tcp_close"},

	// Network Functions - UDP
	"udp_connect": {udpConnectFunc, "udp_connect"},
	"udp_listen":  {udpListenFunc, "udp_listen"},
	"udp_read":    {udpReadFunc, "udp_read"},
	"udp_write":   {udpWriteFunc, "udp_write"},
	"udp_close":   {udpCloseFunc, "udp_close"},

	// Network Functions - Utilities
	"net_resolve": {netResolveFunc, "net_resolve"},
	"net_ping":    {netPingFunc, "net_ping"},

	// Database Functions
	"db_connect":                  {dbConnectFunc, "db_connect"},
	"db_query":                    {dbQueryFunc, "db_query"},
	"db_execute":                  {dbExecuteFunc, "db_execute"},
	"db_connect_with_pool":        {dbConnectWithPoolFunc, "db_connect_with_pool"},
	"db_configure_pool":           {dbConfigurePoolFunc, "db_configure_pool"},
	"db_execute_batch":            {dbExecuteBatchFunc, "db_execute_batch"},
	"db_execute_async":            {dbExecuteAsyncFunc, "db_execute_async"},
	"db_get_async_status":         {dbGetAsyncStatusFunc, "db_get_async_status"},
	"db_cancel_async":             {dbCancelAsyncFunc, "db_cancel_async"},
	"db_list_async_operations":    {dbListAsyncOperationsFunc, "db_list_async_operations"},
	"db_cleanup_async_operations": {dbCleanupAsyncOperationsFunc, "db_cleanup_async_operations"},
	"stream_tables":               {streamTablesFunc, "stream_tables"},
	"db_stop_stream":              {dbStopStreamFunc, "db_stop_stream"},
	"db_close":                    {dbCloseFunc, "db_close"},

	// Regular Expression Functions
	"is_regex_match": {isregexFunc, "is_regex_match"},
	"regex_match":    {regexMatchFunc, "regex_match"},
	"regex_find":     {regexFindFunc, "regex_find"},
	"regex_find_all": {regexFindAllFunc, "regex_find_all"},
	"regex_replace":  {regexReplaceFunc, "regex_replace"},
	"regex_split":    {regexSplitFunc, "regex_split"},

	// Date/Time Functions - Advanced Operations
	"date_now":        {datenowFunc, "date_now"},
	"time_now":        {timenowFunc, "time_now"},
	"sleep":           {sleepFunc, "sleep"},
	"date_format":     {dateformatFunc, "date_format"},
	"date_parse":      {dateParseFunc, "date_parse"},
	"date_format_new": {dateFormatEnhancedFunc, "date_format_new"},
	"date_add":        {dateAddFunc, "date_add"},
	"date_subtract":   {dateSubtractFunc, "date_subtract"},
	"date_diff":       {dateDiffFunc, "date_diff"},
	"date_between":    {dateBetweenFunc, "date_between"},
	"date_compare":    {dateCompareFunc, "date_compare"},

	// Rule Engine - Fact Database & Working Memory
	"fact_assert":  {factAssertFunc, "fact_assert"},
	"fact_retract": {factRetractFunc, "fact_retract"},
	"fact_query":   {factQueryFunc, "fact_query"},
	"fact_exists":  {factExistsFunc, "fact_exists"},
	"fact_count":   {factCountFunc, "fact_count"},
	"fact_clear":   {factClearFunc, "fact_clear"},
	"fact_get_all": {factGetAllFunc, "fact_get_all"},

	// Rule Engine - Complex Event Processing (CEP)
	"event_emit":           {eventEmitFunc, "event_emit"},
	"event_define_pattern": {eventDefinePatternFunc, "event_define_pattern"},
	"event_get_window":     {eventGetWindowFunc, "event_get_window"},
	"event_clear":          {eventClearFunc, "event_clear"},
	"event_count":          {eventCountFunc, "event_count"},

	// File System Operations
	"read_file":        {readFileFunc, "read_file"},
	"write_file":       {writeFileFunc, "write_file"},
	"file_exists":      {fileExistsFunc, "file_exists"},
	"file_size":        {fileSizeFunc, "file_size"},
	"file_modified":    {fileModifiedFunc, "file_modified"},
	"file_permissions": {filePermissionsFunc, "file_permissions"},
	"mkdir":            {mkdirFunc, "mkdir"},
	"rmdir":            {rmdirFunc, "rmdir"},
	"list_dir":         {listDirFunc, "list_dir"},
	"copy_file":        {copyFileFunc, "copy_file"},
	"move_file":        {moveFileFunc, "move_file"},
	"delete_file":      {deleteFileFunc, "delete_file"},
	"path_join":        {pathJoinFunc, "path_join"},
	"path_dirname":     {pathDirnameFunc, "path_dirname"},
	"path_basename":    {pathBasenameFunc, "path_basename"},
	"path_ext":         {pathExtFunc, "path_ext"},
	"getcwd":           {getcwdFunc, "getcwd"},
	"chdir":            {chdirFunc, "chdir"},

	// Memory functions
	"detect_memory_leaks": {detectMemoryLeaksFunc, "detect_memory_leaks"},
	"force_cleanup_leaks": {forceCleanupLeaksFunc, "force_cleanup_leaks"},
	"get_memory_stats":    {getMemoryStatsFunc, "get_memory_stats"},
}
