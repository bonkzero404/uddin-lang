package interpreter

import "sync"

// createThreadSafeInterpreter creates a new interpreter instance for concurrent execution
// This function extracts the common thread-safe interpreter creation logic
func createThreadSafeInterpreter(interp *interpreter, pos Position) *interpreter {
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

	return threadInterp
}
