package interpreter

import (
	"fmt"
)

// detectMemoryLeaksFunc implements the detect_memory_leaks() built-in function
// detect_memory_leaks() -> array
// Returns an array of detected memory leak pointers
func detectMemoryLeaksFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "detect_memory_leaks", args, 0)

	// Get the global memory tracker and detect leaks
	tracker := GetGlobalMemoryTracker()
	leaks := tracker.detectMemoryLeaks()

	// Convert leaks to array of objects
	result := make([]Value, len(leaks))
	for i, ptr := range leaks {
		// Get allocation info for this pointer
		tracker.mutex.RLock()
		alloc, exists := tracker.allocatedPointers[ptr]
		tracker.mutex.RUnlock()

		leakObj := make(map[string]Value)
		// #nosec G103 - Intentional uintptr to unsafe.Pointer conversion for debugging output
		leakObj["pointer"] = Value(fmt.Sprintf("%v", ptr))
		if exists && alloc != nil {
			leakObj["size"] = Value(alloc.size)
			leakObj["type"] = Value(alloc.type_)
			leakObj["alloc_time"] = Value(alloc.allocTime)
			leakObj["freed"] = Value(alloc.freed)
		} else {
			leakObj["size"] = Value(int64(0))
			leakObj["type"] = Value("unknown")
			leakObj["alloc_time"] = Value(int64(0))
			leakObj["freed"] = Value(false)
		}
		result[i] = Value(leakObj)
	}

	return Value(result)
}

// forceCleanupLeaksFunc implements the force_cleanup_leaks() built-in function
// force_cleanup_leaks() -> bool
// Forces cleanup of detected memory leaks and returns success status
func forceCleanupLeaksFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "force_cleanup_leaks", args, 0)

	// Get the global memory tracker and force cleanup
	tracker := GetGlobalMemoryTracker()
	err := tracker.forceCleanupLeaks()

	return Value(err == nil)
}

// getMemoryStatsFunc implements the get_memory_stats() built-in function
// get_memory_stats() -> object
// Returns memory statistics as an object
func getMemoryStatsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "get_memory_stats", args, 0)

	// Get the global memory tracker stats
	tracker := GetGlobalMemoryTracker()
	stats := tracker.GetMemoryStats()

	// Convert stats to object
	statsObj := make(map[string]Value)
	statsObj["tagged_values_created"] = Value(stats.TaggedValuesCreated)
	statsObj["tagged_values_reused"] = Value(stats.TaggedValuesReused)
	statsObj["cache_hits"] = Value(stats.CacheHits)
	statsObj["cache_misses"] = Value(stats.CacheMisses)

	return Value(statsObj)
}
