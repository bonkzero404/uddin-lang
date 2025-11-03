package interpreter

import (
	"sort"
)

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
		// Fast path: no elements to append
		if len(args) == 1 {
			return Value(nil)
		}

		// Optimize memory allocation by pre-calculating capacity
		toAppend := args[1:]
		if cap(*list)-len(*list) < len(toAppend) {
			// Need to grow slice, allocate with exact capacity
			newCap := len(*list) + len(toAppend)
			newList := make([]Value, len(*list), newCap)
			copy(newList, *list)
			*list = SmartAppend(newList, toAppend...)
		} else {
			// Sufficient capacity, direct append
			*list = SmartAppend(*list, toAppend...)
		}
		return Value(nil)
	}

	// Error if first argument is not an array
	panic(typeError(pos, "append() requires first argument to be list"))
}

// lenFunc implements the len() built-in function
// Returns the length of a string, array, or object
// Parameters:
//   - value: String, array, or object to get the length of
//
// Returns the length as an integer
// Example: len("hello") -> 5
func lenFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "len", args, 1); err != Value(nil) {
		return err
	}
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

// rangeFunc implements the range() built-in function
// Creates an array of integers
// Parameters:
//   - If 1 arg: n (upper bound, exclusive) -> [0, 1, ..., n-1]
//   - If 2 args: start, stop -> [start, start+1, ..., stop-1]
//   - If 3 args: start, stop, step -> [start, start+step, ..., stop-1]
//
// Returns an array of integers
// Examples:
//
//	range(3) -> [0, 1, 2]
//	range(1, 4) -> [1, 2, 3]
//	range(0, 10, 2) -> [0, 2, 4, 6, 8]
func rangeFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 1 {
		// Single argument: range(n) -> [0, 1, ..., n-1]
		if n, ok := args[0].(int); ok {
			if n < 0 {
				panic(valueError(pos, "range() argument must not be negative"))
			}
			// Fast path for small ranges
			if n == 0 {
				nums := make([]Value, 0)
				return Value(&nums)
			}
			nums := make([]Value, n)
			// Optimized loop with fewer type conversions
			for i := 0; i < n; i++ {
				nums[i] = Value(i)
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

		if start >= stop {
			// Return empty array if start >= stop
			nums := make([]Value, 0)
			return Value(&nums)
		}

		size := stop - start
		nums := make([]Value, size)
		// Optimized loop with direct assignment
		for i := 0; i < size; i++ {
			nums[i] = Value(start + i)
		}
		return Value(&nums)
	} else if len(args) == 3 {
		// Three arguments: range(start, stop, step) -> [start, start+step, ..., stop-1]
		start, startOk := args[0].(int)
		stop, stopOk := args[1].(int)
		step, stepOk := args[2].(int)

		if !startOk || !stopOk || !stepOk {
			panic(typeError(pos, "range() requires integer arguments"))
		}

		if step == 0 {
			panic(valueError(pos, "range() step argument must not be zero"))
		}

		// Calculate the size of the result array
		var nums []Value
		if step > 0 {
			// Positive step: start < stop
			for i := start; i < stop; i += step {
				nums = append(nums, Value(i))
			}
		} else {
			// Negative step: start > stop
			for i := start; i > stop; i += step {
				nums = append(nums, Value(i))
			}
		}

		return Value(&nums)
	}

	panic(valueError(pos, "range() requires 1, 2, or 3 arguments, got %d", len(args)))
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
	if err := ensureNumArgs(pos, "slice", args, 3); err != Value(nil) {
		return err
	}
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

// mapFunc implements the map() built-in function
// map(array, function) -> array
// Example: map([1, 2, 3], lambda x: x * 2) -> [2, 4, 6]
func mapFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "map", args, 2); err != Value(nil) {
		return Value(err)
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "map() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "map() requires second argument to be a function"))
	}

	// For small arrays, use sequential processing to avoid overhead
	if len(*arr) < 100 {
		result := make([]Value, len(*arr))
		for i, v := range *arr {
			result[i] = interp.callFunction(pos, fn, []Value{v})
		}
		return Value(&result)
	}

	// Use concurrent executor for parallel processing with thread-safe interpreter instances
	executor := GetGlobalConcurrentExecutor()
	mapFunc := func(v Value) Value {
		// Create a new interpreter instance for this goroutine to avoid scope conflicts
		threadInterp := createThreadSafeInterpreter(interp, pos)
		return threadInterp.callFunction(pos, fn, []Value{v})
	}

	result, err := executor.ParallelMapOperation(*arr, mapFunc)
	if err != nil {
		// Fallback to sequential execution on error
		result = make([]Value, len(*arr))
		for i, v := range *arr {
			result[i] = interp.callFunction(pos, fn, []Value{v})
		}
	}

	return Value(&result)
}

// filterFunc implements the filter() built-in function
// filter(array, function) -> array
// Example: filter([1, 2, 3, 4], lambda x: x % 2 == 0) -> [2, 4]
func filterFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "filter", args, 2); err != Value(nil) {
		return Value(err)
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "filter() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "filter() requires second argument to be a function"))
	}

	// Note: Concurrent execution is available but disabled for thread safety
	// TODO: Implement thread-safe interpreter access for concurrent execution
	// For now, using sequential execution for all array sizes
	if len(*arr) < 100 {
		result := make([]Value, 0)
		for _, v := range *arr {
			if IsTruthy(interp.callFunction(pos, fn, []Value{v})) {
				result = append(result, v)
			}
		}
		return Value(&result)
	}

	// Use concurrent executor for parallel processing with thread-safe interpreter instances
	executor := GetGlobalConcurrentExecutor()
	filterFunc := func(v Value) bool {
		threadInterp := createThreadSafeInterpreter(interp, pos)
		result := threadInterp.callFunction(pos, fn, []Value{v})
		if b, ok := result.(bool); ok {
			return b
		}
		return false
	}

	result, err := executor.ParallelFilterOperation(*arr, filterFunc)
	if err != nil {
		// Fallback to sequential execution on error
		var filtered []Value
		for _, v := range *arr {
			result := interp.callFunction(pos, fn, []Value{v})
			if b, ok := result.(bool); ok && b {
				filtered = append(filtered, v)
			}
		}
		result = filtered
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

	initialValue := args[2]

	// Note: Concurrent execution is available but disabled for thread safety
	// TODO: Implement thread-safe interpreter access for concurrent execution
	// For now, using sequential execution for all array sizes
	if len(*arr) < 100 {
		accumulator := initialValue
		for _, v := range *arr {
			accumulator = interp.callFunction(pos, fn, []Value{accumulator, v})
		}
		return accumulator
	}

	// Use concurrent executor for parallel processing with thread-safe interpreter instances
	executor := GetGlobalConcurrentExecutor()
	reduceFunc := func(a, b Value) Value {
		threadInterp := createThreadSafeInterpreter(interp, pos)
		return threadInterp.callFunction(pos, fn, []Value{a, b})
	}

	result, err := executor.ParallelReduceOperation(*arr, reduceFunc, initialValue)
	if err != nil {
		// Fallback to sequential execution on error
		result = initialValue
		for _, v := range *arr {
			result = interp.callFunction(pos, fn, []Value{result, v})
		}
	}

	return result
}

// reverseFunc implements the reverse() built-in function
// reverse(array) -> null (modifies array in place)
// Example: reverse([1, 2, 3]) -> [3, 2, 1]
func reverseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "reverse", args, 1); err != Value(nil) {
		return err
	}
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

	// Optimize memory allocation by pre-calculating capacity
	toPush := args[1:]
	if cap(*arr)-len(*arr) < len(toPush) {
		// Need to grow slice, allocate with exact capacity
		newCap := len(*arr) + len(toPush)
		newArr := make([]Value, len(*arr), newCap)
		copy(newArr, *arr)
		*arr = append(newArr, toPush...)
	} else {
		// Sufficient capacity, direct append
		*arr = append(*arr, toPush...)
	}
	return Value(nil)
}

// popFunc implements the pop() built-in function
// pop(array) -> any (removes and returns last element)
// Example: pop([1, 2, 3]) -> 3, array becomes [1, 2]
func popFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "pop", args, 1); err != Value(nil) {
		return err
	}
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
	if err := ensureNumArgs(pos, "shift", args, 1); err != Value(nil) {
		return err
	}
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
	if err := ensureNumArgs(pos, "index_of", args, 2); err != Value(nil) {
		return err
	}
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
	if err := ensureNumArgs(pos, "last_index_of", args, 2); err != Value(nil) {
		return err
	}
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

// concurrentMapFunc implements concurrent_map(array, function)
func concurrentMapFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "concurrent_map", args, 2); err != Value(nil) {
		return Value(err)
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_map() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_map() requires second argument to be a function"))
	}

	// Use concurrent executor for all sizes
	executor := GetGlobalConcurrentExecutor()
	mapFunc := func(v Value) Value {
		threadInterp := createThreadSafeInterpreter(interp, pos)
		return threadInterp.callFunction(pos, fn, []Value{v})
	}
	result, err := executor.ParallelMapOperation(*arr, mapFunc)
	if err != nil {
		// Fallback sequentially on error
		result = make([]Value, len(*arr))
		for i, v := range *arr {
			result[i] = interp.callFunction(pos, fn, []Value{v})
		}
	}
	return Value(&result)
}

// concurrentFilterFunc implements concurrent_filter(array, function)
func concurrentFilterFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "concurrent_filter", args, 2); err != Value(nil) {
		return Value(err)
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_filter() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_filter() requires second argument to be a function"))
	}

	executor := GetGlobalConcurrentExecutor()
	filterFunc := func(v Value) bool {
		threadInterp := createThreadSafeInterpreter(interp, pos)
		res := threadInterp.callFunction(pos, fn, []Value{v})
		if b, ok := res.(bool); ok {
			return b
		}
		return false
	}
	result, err := executor.ParallelFilterOperation(*arr, filterFunc)
	if err != nil {
		// Fallback sequential
		var filtered []Value
		for _, v := range *arr {
			res := interp.callFunction(pos, fn, []Value{v})
			if b, ok := res.(bool); ok && b {
				filtered = append(filtered, v)
			}
		}
		result = filtered
	}
	return Value(&result)
}

// concurrentReduceFunc implements concurrent_reduce(array, function, initial)
func concurrentReduceFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 3 {
		panic(typeError(pos, "concurrent_reduce() requires 3 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_reduce() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_reduce() requires second argument to be a function"))
	}
	initialValue := args[2]

	executor := GetGlobalConcurrentExecutor()
	reduceFunc := func(a, b Value) Value {
		threadInterp := createThreadSafeInterpreter(interp, pos)
		return threadInterp.callFunction(pos, fn, []Value{a, b})
	}
	result, err := executor.ParallelReduceOperation(*arr, reduceFunc, initialValue)
	if err != nil {
		// Fallback sequential
		result = initialValue
		for _, v := range *arr {
			result = interp.callFunction(pos, fn, []Value{result, v})
		}
	}
	return result
}
