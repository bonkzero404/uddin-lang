package interpreter

import "sync"

// concurrentMapFunc implements the concurrent_map() built-in function
// concurrent_map(array, function) -> array
// Example: concurrent_map([1, 2, 3, 4], fun(x): return x * 2 end) -> [2, 4, 6, 8]
func concurrentMapFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "concurrent_map", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_map() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_map() requires second argument to be a function"))
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
		// Copy the current interpreter's configuration and global scope
		threadInterp := &interpreter{
			vars:                make([]map[string]Value, len(interp.vars)),
			mutex:               sync.RWMutex{},
			args:                interp.args,
			stdin:               interp.stdin,
			stdinReader:         interp.stdinReader,
			stdout:              interp.stdout,
			exit:                interp.exit,
			stats:               Stats{},
			inUnitTest:          interp.inUnitTest,
			currentPos:          pos,
			memoCache:           make(map[string]Value),
			stringBuilderPool:   sync.Pool{},
			arrayPool:           sync.Pool{},
			mapPool:             sync.Pool{},
			variableLookupCache: nil, // Disable cache for thread safety
			optimizer:           nil, // Disable optimizer for thread safety
		}

		// Deep copy the variable scopes to avoid shared state
		interp.mutex.RLock()
		for i, scope := range interp.vars {
			threadInterp.vars[i] = make(map[string]Value)
			for k, v := range scope {
				threadInterp.vars[i][k] = v
			}
		}
		interp.mutex.RUnlock()

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

// concurrentFilterFunc implements the concurrent_filter() built-in function
// concurrent_filter(array, function) -> array
// Example: concurrent_filter([1, 2, 3, 4], fun(x): return x % 2 == 0 end) -> [2, 4]
func concurrentFilterFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "concurrent_filter", args, 2)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_filter() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_filter() requires second argument to be a function"))
	}

	// For small arrays, use sequential processing to avoid overhead
	if len(*arr) < 100 {
		var filtered []Value
		for _, v := range *arr {
			result := interp.callFunction(pos, fn, []Value{v})
			if b, ok := result.(bool); ok && b {
				filtered = append(filtered, v)
			}
		}
		return Value(&filtered)
	}

	// Use concurrent executor for parallel processing with thread-safe interpreter instances
	executor := GetGlobalConcurrentExecutor()
	filterFunc := func(v Value) bool {
		// Create a new interpreter instance for this goroutine to avoid scope conflicts
		// Copy the current interpreter's configuration and global scope
		threadInterp := &interpreter{
			vars:                make([]map[string]Value, len(interp.vars)),
			mutex:               sync.RWMutex{},
			args:                interp.args,
			stdin:               interp.stdin,
			stdinReader:         interp.stdinReader,
			stdout:              interp.stdout,
			exit:                interp.exit,
			stats:               Stats{},
			inUnitTest:          interp.inUnitTest,
			currentPos:          pos,
			memoCache:           make(map[string]Value),
			stringBuilderPool:   sync.Pool{},
			arrayPool:           sync.Pool{},
			mapPool:             sync.Pool{},
			variableLookupCache: nil, // Disable cache for thread safety
			optimizer:           nil, // Disable optimizer for thread safety
		}

		// Deep copy the variable scopes to avoid shared state
		interp.mutex.RLock()
		for i, scope := range interp.vars {
			threadInterp.vars[i] = make(map[string]Value)
			for k, v := range scope {
				threadInterp.vars[i][k] = v
			}
		}
		interp.mutex.RUnlock()

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

// concurrentReduceFunc implements the concurrent_reduce() built-in function
// concurrent_reduce(array, function, initial) -> value
// Example: concurrent_reduce([1, 2, 3, 4], fun(a, b): return a + b end, 0) -> 10
func concurrentReduceFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "concurrent_reduce", args, 3)
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "concurrent_reduce() requires first argument to be an array"))
	}
	fn, ok := args[1].(functionType)
	if !ok {
		panic(typeError(pos, "concurrent_reduce() requires second argument to be a function"))
	}
	initialValue := args[2]

	// For small arrays, use sequential processing to avoid overhead
	if len(*arr) < 100 {
		result := initialValue
		for _, v := range *arr {
			result = interp.callFunction(pos, fn, []Value{result, v})
		}
		return result
	}

	// Use concurrent executor for parallel processing with thread-safe interpreter instances
	executor := GetGlobalConcurrentExecutor()
	reduceFunc := func(a, b Value) Value {
		// Create a new interpreter instance for this goroutine to avoid scope conflicts
		// Copy the current interpreter's configuration and global scope
		threadInterp := &interpreter{
			vars:                make([]map[string]Value, len(interp.vars)),
			mutex:               sync.RWMutex{},
			args:                interp.args,
			stdin:               interp.stdin,
			stdinReader:         interp.stdinReader,
			stdout:              interp.stdout,
			exit:                interp.exit,
			stats:               Stats{},
			inUnitTest:          interp.inUnitTest,
			currentPos:          pos,
			memoCache:           make(map[string]Value),
			stringBuilderPool:   sync.Pool{},
			arrayPool:           sync.Pool{},
			mapPool:             sync.Pool{},
			variableLookupCache: nil, // Disable cache for thread safety
			optimizer:           nil, // Disable optimizer for thread safety
		}

		// Deep copy the variable scopes to avoid shared state
		interp.mutex.RLock()
		for i, scope := range interp.vars {
			threadInterp.vars[i] = make(map[string]Value)
			for k, v := range scope {
				threadInterp.vars[i][k] = v
			}
		}
		interp.mutex.RUnlock()

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

// parallelExecuteFunc implements the parallel_execute() built-in function
// parallel_execute(function1, function2, ...) -> [result1, result2, ...]
// Example: parallel_execute(fun(): return 1 end, fun(): return 2 end) -> [1, 2]
func parallelExecuteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 0 {
		panic(typeError(pos, "parallel_execute() requires at least one function argument"))
	}

	// Validate all arguments are functions
	functions := make([]functionType, len(args))
	for i, arg := range args {
		fn, ok := arg.(functionType)
		if !ok {
			panic(typeError(pos, "parallel_execute() requires all arguments to be functions"))
		}
		functions[i] = fn
	}

	// Execute functions in parallel
	results := make([]Value, len(functions))
	resultChan := make(chan struct {
		index int
		value Value
		err   error
	}, len(functions))

	// Submit tasks to worker pool
	for i, fn := range functions {
		go func(index int, function functionType) {
			defer func() {
				if r := recover(); r != nil {
					resultChan <- struct {
						index int
						value Value
						err   error
					}{index, Value(nil), typeError(pos, "function execution failed")}
				}
			}()

			result := interp.callFunction(pos, function, []Value{})
			resultChan <- struct {
				index int
				value Value
				err   error
			}{index, result, nil}
		}(i, fn)
	}

	// Collect results
	for i := 0; i < len(functions); i++ {
		result := <-resultChan
		if result.err != nil {
			panic(result.err)
		}
		results[result.index] = result.value
	}

	return Value(&results)
}
