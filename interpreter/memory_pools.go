package interpreter

import (
	"strings"
	"sync"
)

// ========================================
// Batch Operations for Large Arrays
// ========================================

// BatchArrayProcessor provides optimized batch processing for large arrays
type BatchArrayProcessor struct {
	result    []Value
	batchSize int
	processor func([]Value) []Value
}

// NewBatchArrayProcessor creates a new batch processor
func NewBatchArrayProcessor(estimatedSize, batchSize int, processor func([]Value) []Value) *BatchArrayProcessor {
	return &BatchArrayProcessor{
		result:    make([]Value, 0, estimatedSize),
		batchSize: batchSize,
		processor: processor,
	}
}

// ProcessBatch processes array in batches to reduce memory pressure
func (bap *BatchArrayProcessor) ProcessBatch(input []Value) []Value {
	if len(input) <= bap.batchSize {
		return bap.processor(input)
	}

	for i := 0; i < len(input); i += bap.batchSize {
		end := i + bap.batchSize
		if end > len(input) {
			end = len(input)
		}
		batch := input[i:end]
		processed := bap.processor(batch)
		bap.result = SmartAppend(bap.result, processed...)
	}
	return bap.result
}

// ========================================
// String Interning for Repeated Strings
// ========================================

// StringInterner provides string interning to reduce memory usage
type StringInterner struct {
	cache map[string]string
	mutex sync.RWMutex
}

var globalStringInterner = &StringInterner{
	cache: make(map[string]string),
}

// Intern returns an interned version of the string
func (si *StringInterner) Intern(s string) string {
	si.mutex.RLock()
	if interned, exists := si.cache[s]; exists {
		si.mutex.RUnlock()
		return interned
	}
	si.mutex.RUnlock()

	si.mutex.Lock()
	defer si.mutex.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := si.cache[s]; exists {
		return interned
	}

	si.cache[s] = s
	return s
}

// InternString provides global string interning
func InternString(s string) string {
	return globalStringInterner.Intern(s)
}

// ========================================
// Function Call Optimization
// ========================================

// FunctionCallCache provides caching for pure function calls
type FunctionCallCache struct {
	cache   map[string]Value
	mutex   sync.RWMutex
	maxSize int
}

// NewFunctionCallCache creates a new function call cache
func NewFunctionCallCache(maxSize int) *FunctionCallCache {
	return &FunctionCallCache{
		cache:   make(map[string]Value),
		maxSize: maxSize,
	}
}

// Get retrieves a cached function result
func (fcc *FunctionCallCache) Get(key string) (Value, bool) {
	fcc.mutex.RLock()
	defer fcc.mutex.RUnlock()
	value, exists := fcc.cache[key]
	return value, exists
}

// Set stores a function result in cache
func (fcc *FunctionCallCache) Set(key string, value Value) {
	fcc.mutex.Lock()
	defer fcc.mutex.Unlock()

	// Simple LRU: if cache is full, clear it
	if len(fcc.cache) >= fcc.maxSize {
		fcc.cache = make(map[string]Value)
	}

	fcc.cache[key] = value
}

// ========================================
// Memory Pool for Complex Operations
// ========================================

// ComplexOperationPool provides pooled resources for complex operations
type ComplexOperationPool struct {
	arrayPool  sync.Pool
	mapPool    sync.Pool
	stringPool sync.Pool
}

// NewComplexOperationPool creates a new complex operation pool
func NewComplexOperationPool() *ComplexOperationPool {
	return &ComplexOperationPool{
		arrayPool: sync.Pool{
			New: func() any {
				return make([]Value, 0, 64)
			},
		},
		mapPool: sync.Pool{
			New: func() any {
				return make(map[string]Value)
			},
		},
		stringPool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// GetArray gets a pooled array
func (cop *ComplexOperationPool) GetArray() []Value {
	arr := cop.arrayPool.Get().([]Value)
	return arr[:0]
}

// PutArray returns an array to the pool
func (cop *ComplexOperationPool) PutArray(arr []Value) {
	if cap(arr) <= 1024 {
		cop.arrayPool.Put(arr)
	}
}

// GetMap gets a pooled map
func (cop *ComplexOperationPool) GetMap() map[string]Value {
	m := cop.mapPool.Get().(map[string]Value)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMap returns a map to the pool
func (cop *ComplexOperationPool) PutMap(m map[string]Value) {
	if len(m) <= 256 {
		cop.mapPool.Put(m)
	}
}

// GetStringBuilder gets a pooled string builder
func (cop *ComplexOperationPool) GetStringBuilder() *strings.Builder {
	sb := cop.stringPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
func (cop *ComplexOperationPool) PutStringBuilder(sb *strings.Builder) {
	cop.stringPool.Put(sb)
}

// Global complex operation pool
var globalComplexPool = NewComplexOperationPool()

// ========================================
// Optimized Type Assertions
// ========================================

// FastTypeAssertion provides optimized type assertions with caching
type FastTypeAssertion struct {
	cache map[any]string
	mutex sync.RWMutex
}

var _ = &FastTypeAssertion{
	cache: make(map[any]string),
}

// GetTypeName returns the cached type name
func (fta *FastTypeAssertion) GetTypeName(v any) string {
	fta.mutex.RLock()
	if typeName, exists := fta.cache[v]; exists {
		fta.mutex.RUnlock()
		return typeName
	}
	fta.mutex.RUnlock()

	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	// Double-check after acquiring write lock
	if typeName, exists := fta.cache[v]; exists {
		return typeName
	}

	typeName := typeName(v)
	fta.cache[v] = typeName
	return typeName
}

// ========================================
// UTILITY FUNCTIONS
// ========================================

// OptimizedArrayFilter provides efficient array filtering
func OptimizedArrayFilter(arr []Value, predicate func(Value) bool) []Value {
	result := globalComplexPool.GetArray()
	defer globalComplexPool.PutArray(result)

	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}

	// Return a copy to avoid pool interference
	final := make([]Value, len(result))
	copy(final, result)
	return final
}

// OptimizedArrayMap provides efficient array mapping
func OptimizedArrayMap(arr []Value, mapper func(Value) Value) []Value {
	result := make([]Value, len(arr))
	for i, v := range arr {
		result[i] = mapper(v)
	}
	return result
}

// OptimizedStringJoin provides efficient string joining with pre-allocation
func OptimizedStringJoin(arr []Value, separator string) string {
	if len(arr) == 0 {
		return ""
	}

	if len(arr) == 1 {
		return toString(arr[0], false)
	}

	sb := globalComplexPool.GetStringBuilder()
	defer globalComplexPool.PutStringBuilder(sb)

	for i, v := range arr {
		if i > 0 {
			sb.WriteString(separator)
		}
		sb.WriteString(toString(v, false))
	}

	return sb.String()
}

// OptimizedMapMerge provides efficient map merging
func OptimizedMapMerge(maps ...map[string]Value) map[string]Value {
	totalSize := 0
	for _, m := range maps {
		totalSize += len(m)
	}

	result := make(map[string]Value, totalSize)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// ========================================
// PERFORMANCE MONITORING
// ========================================

// PerformanceMonitor tracks operation performance
type PerformanceMonitor struct {
	operationCounts map[string]int64
	mutex           sync.RWMutex
}

var globalPerfMonitor = &PerformanceMonitor{
	operationCounts: make(map[string]int64),
}

// IncrementOperation increments the count for an operation
func (pm *PerformanceMonitor) IncrementOperation(operation string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.operationCounts[operation]++
}

// GetOperationCount returns the count for an operation
func (pm *PerformanceMonitor) GetOperationCount(operation string) int64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.operationCounts[operation]
}

// GetAllCounts returns all operation counts
func (pm *PerformanceMonitor) GetAllCounts() map[string]int64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]int64)
	for k, v := range pm.operationCounts {
		result[k] = v
	}
	return result
}

// TrackOperation tracks an operation
func TrackOperation(operation string) {
	globalPerfMonitor.IncrementOperation(operation)
}
