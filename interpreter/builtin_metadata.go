package interpreter

// BuiltinMetadata contains metadata for builtin function configuration
// Used to automatically configure dispatcher settings without hardcoded switch statements
type BuiltinMetadata struct {
	ArgCount int  // Number of arguments (-1 for variadic)
	FastPath bool // Whether to use fast path (skip validation)
}

// builtinMetadataMap contains metadata for builtin functions
// Functions not in this map will use defaults: ArgCount=-1 (variadic), FastPath=false
// This is defined as a function to avoid initialization cycles
func getBuiltinMetadataMap() map[string]BuiltinMetadata {
	return map[string]BuiltinMetadata{
		// Core System Functions
		"print":   {ArgCount: -1, FastPath: true},
		"println": {ArgCount: -1, FastPath: true},
		"input":   {ArgCount: 0, FastPath: false},
		"import":  {ArgCount: 1, FastPath: false},
		"exit":    {ArgCount: -1, FastPath: false},
		"typeof":  {ArgCount: 1, FastPath: true},

		// Type Conversion Functions (all take 1 arg, fast path)
		"int":   {ArgCount: 1, FastPath: true},
		"float": {ArgCount: 1, FastPath: true},
		"str":   {ArgCount: 1, FastPath: true},
		"char":  {ArgCount: 1, FastPath: true},
		"rune":  {ArgCount: 1, FastPath: true},

		// Basic Array/Collection Functions
		"append": {ArgCount: -1, FastPath: false}, // Variadic
		"len":    {ArgCount: 1, FastPath: true},
		"range":  {ArgCount: -1, FastPath: false}, // 1-3 args
		"slice":  {ArgCount: 3, FastPath: false},
		"sort":   {ArgCount: 1, FastPath: false}, // May have comparator

		// String Manipulation - Basic (2 args, fast path)
		"join":     {ArgCount: 2, FastPath: true},
		"split":    {ArgCount: 2, FastPath: true},
		"contains": {ArgCount: 2, FastPath: true},
		"lower":    {ArgCount: 1, FastPath: true},
		"upper":    {ArgCount: 1, FastPath: true},
		"substr":   {ArgCount: 3, FastPath: false},
		"find":     {ArgCount: 2, FastPath: false},
		"str_pad":  {ArgCount: 3, FastPath: false},

		// String Manipulation - Advanced
		"replace":     {ArgCount: 3, FastPath: false},
		"trim":        {ArgCount: 1, FastPath: true},
		"starts_with": {ArgCount: 2, FastPath: false},
		"ends_with":   {ArgCount: 2, FastPath: false},
		"repeat":      {ArgCount: 2, FastPath: false},
		"reverse_str": {ArgCount: 1, FastPath: true},

		// Array/Collection Methods - Advanced
		"map":               {ArgCount: 2, FastPath: false}, // array, function
		"filter":            {ArgCount: 2, FastPath: false},
		"reduce":            {ArgCount: 3, FastPath: false},
		"concurrent_map":    {ArgCount: 2, FastPath: false},
		"concurrent_filter": {ArgCount: 2, FastPath: false},
		"concurrent_reduce": {ArgCount: 3, FastPath: false},
		"reverse":           {ArgCount: 1, FastPath: true},
		"push":              {ArgCount: 2, FastPath: true},
		"pop":               {ArgCount: 1, FastPath: true},
		"shift":             {ArgCount: 1, FastPath: true},
		"unshift":           {ArgCount: 2, FastPath: true},
		"index_of":          {ArgCount: 2, FastPath: false},
		"last_index_of":     {ArgCount: 2, FastPath: false},

		// Data Structures - Set Operations
		"set_new":      {ArgCount: 0, FastPath: false},
		"set_add":      {ArgCount: 2, FastPath: true},
		"set_remove":   {ArgCount: 2, FastPath: true},
		"set_has":      {ArgCount: 2, FastPath: true},
		"set_size":     {ArgCount: 1, FastPath: true},
		"set_to_array": {ArgCount: 1, FastPath: false},

		// Data Structures - Stack Operations
		"stack_new":   {ArgCount: 0, FastPath: false},
		"stack_push":  {ArgCount: 2, FastPath: true},
		"stack_pop":   {ArgCount: 1, FastPath: true},
		"stack_peek":  {ArgCount: 1, FastPath: true},
		"stack_size":  {ArgCount: 1, FastPath: true},
		"stack_empty": {ArgCount: 1, FastPath: true},

		// Data Structures - Queue Operations
		"queue_new":     {ArgCount: 0, FastPath: false},
		"queue_enqueue": {ArgCount: 2, FastPath: true},
		"queue_dequeue": {ArgCount: 1, FastPath: true},
		"queue_front":   {ArgCount: 1, FastPath: true},
		"queue_size":    {ArgCount: 1, FastPath: true},
		"queue_empty":   {ArgCount: 1, FastPath: true},

		// Mathematical Functions - Basic Operations (1 arg, fast path)
		"abs":  {ArgCount: 1, FastPath: true},
		"sqrt": {ArgCount: 1, FastPath: true},
		"cbrt": {ArgCount: 1, FastPath: true},
		"pow":  {ArgCount: 2, FastPath: true},   // 2 args
		"max":  {ArgCount: -1, FastPath: false}, // Variadic
		"min":  {ArgCount: -1, FastPath: false}, // Variadic

		// Mathematical Functions - Rounding
		"round": {ArgCount: -1, FastPath: true}, // 1-2 args: value, [decimal_places]
		"floor": {ArgCount: 1, FastPath: true},
		"ceil":  {ArgCount: 1, FastPath: true},
		"trunc": {ArgCount: 1, FastPath: true},

		// Mathematical Functions - Trigonometric
		"sin":   {ArgCount: 1, FastPath: true},
		"cos":   {ArgCount: 1, FastPath: true},
		"tan":   {ArgCount: 1, FastPath: true},
		"asin":  {ArgCount: 1, FastPath: true},
		"acos":  {ArgCount: 1, FastPath: true},
		"atan":  {ArgCount: 1, FastPath: true},
		"atan2": {ArgCount: 2, FastPath: false},

		// Mathematical Functions - Hyperbolic
		"sinh": {ArgCount: 1, FastPath: true},
		"cosh": {ArgCount: 1, FastPath: true},
		"tanh": {ArgCount: 1, FastPath: true},

		// Mathematical Functions - Logarithmic & Exponential
		"log":   {ArgCount: 1, FastPath: true},
		"log10": {ArgCount: 1, FastPath: true},
		"log2":  {ArgCount: 1, FastPath: true},
		"logb":  {ArgCount: 2, FastPath: false},
		"exp":   {ArgCount: 1, FastPath: true},
		"exp2":  {ArgCount: 1, FastPath: true},

		// Mathematical Functions - Statistical (variadic)
		"sum":      {ArgCount: -1, FastPath: false},
		"mean":     {ArgCount: -1, FastPath: false},
		"median":   {ArgCount: -1, FastPath: false},
		"mode":     {ArgCount: -1, FastPath: false},
		"std_dev":  {ArgCount: -1, FastPath: false},
		"variance": {ArgCount: -1, FastPath: false},

		// Mathematical Functions - Number Theory
		"gcd":           {ArgCount: 2, FastPath: false},
		"lcm":           {ArgCount: 2, FastPath: false},
		"factorial":     {ArgCount: 1, FastPath: false},
		"fibonacci":     {ArgCount: 1, FastPath: false},
		"is_prime":      {ArgCount: 1, FastPath: false},
		"prime_factors": {ArgCount: 1, FastPath: false},

		// Mathematical Functions - Random Numbers
		"random":        {ArgCount: 0, FastPath: false},
		"random_int":    {ArgCount: 2, FastPath: false},
		"random_float":  {ArgCount: 2, FastPath: false}, // min, max
		"random_choice": {ArgCount: 1, FastPath: false},
		"shuffle":       {ArgCount: 1, FastPath: false},
		"seed_random":   {ArgCount: 1, FastPath: false},

		// Mathematical Functions - Utility
		"sign":        {ArgCount: 1, FastPath: true},
		"clamp":       {ArgCount: 3, FastPath: false},
		"lerp":        {ArgCount: 3, FastPath: false},
		"degrees":     {ArgCount: 1, FastPath: true},
		"radians":     {ArgCount: 1, FastPath: true},
		"is_nan":      {ArgCount: 1, FastPath: true},
		"is_infinite": {ArgCount: 1, FastPath: true},

		// Data Serialization - XML (JSON moved to stdlib/json)
		"xml_parse":     {ArgCount: 1, FastPath: false},
		"xml_stringify": {ArgCount: 1, FastPath: false},

		// Memory functions
		"detect_memory_leaks": {ArgCount: 0, FastPath: false},
		"force_cleanup_leaks": {ArgCount: 0, FastPath: false},
		"get_memory_stats":    {ArgCount: 0, FastPath: true},
	}
}

// getBuiltinMetadata returns metadata for a builtin function
// Returns default metadata (variadic, no fast path) if not found
func getBuiltinMetadata(name string) BuiltinMetadata {
	metadataMap := getBuiltinMetadataMap()
	if meta, exists := metadataMap[name]; exists {
		return meta
	}
	// Default: variadic, no fast path
	return BuiltinMetadata{ArgCount: -1, FastPath: false}
}
