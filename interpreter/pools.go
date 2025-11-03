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
	arrayPool      sync.Pool
	mapPool        sync.Pool
	stringPool     sync.Pool
	smallArrayPool sync.Pool // For small arrays (capacity <= 16)
	largeArrayPool sync.Pool // For large arrays (capacity > 1024)
	intSlicePool   sync.Pool // For int slices
}

// NewComplexOperationPool creates a new complex operation pool
func NewComplexOperationPool() *ComplexOperationPool {
	return &ComplexOperationPool{
		arrayPool: sync.Pool{
			New: func() any {
				arr := make([]Value, 0, 64)
				return &arr
			},
		},
		smallArrayPool: sync.Pool{
			New: func() any {
				arr := make([]Value, 0, 16)
				return &arr
			},
		},
		largeArrayPool: sync.Pool{
			New: func() any {
				arr := make([]Value, 0, 2048) // For large arrays
				return &arr
			},
		},
		mapPool: sync.Pool{
			New: func() any {
				return make(map[string]Value, 16)
			},
		},
		stringPool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
		intSlicePool: sync.Pool{
			New: func() any {
				slice := make([]int, 0, 32)
				return &slice
			},
		},
	}
}

// GetArray gets a pooled array
func (cop *ComplexOperationPool) GetArray() []Value {
	if ptr := cop.arrayPool.Get(); ptr != nil {
		if arr, ok := ptr.(*[]Value); ok {
			return (*arr)[:0]
		}
	}
	return make([]Value, 0, 32)
}

// GetSmallArray gets a pooled small array (capacity <= 16)
func (cop *ComplexOperationPool) GetSmallArray() []Value {
	if ptr := cop.smallArrayPool.Get(); ptr != nil {
		if arr, ok := ptr.(*[]Value); ok {
			return (*arr)[:0]
		}
	}
	return make([]Value, 0, 8)
}

// GetLargeArray gets a pooled large array (capacity > 1024)
func (cop *ComplexOperationPool) GetLargeArray() []Value {
	if ptr := cop.largeArrayPool.Get(); ptr != nil {
		if arr, ok := ptr.(*[]Value); ok {
			return (*arr)[:0]
		}
	}
	return make([]Value, 0, 2048)
}

// GetIntSlice gets a pooled int slice
func (cop *ComplexOperationPool) GetIntSlice() []int {
	if ptr := cop.intSlicePool.Get(); ptr != nil {
		if slice, ok := ptr.(*[]int); ok {
			return (*slice)[:0]
		}
	}
	return make([]int, 0, 32)
}

// PutArray returns an array to the pool
func (cop *ComplexOperationPool) PutArray(arr *[]Value) {
	// Only pool arrays with reasonable capacity to avoid memory bloat
	if cap(*arr) <= 1024 && cap(*arr) >= 8 {
		cop.arrayPool.Put(arr)
	}
}

// PutSmallArray returns a small array to the pool
func (cop *ComplexOperationPool) PutSmallArray(arr *[]Value) {
	if cap(*arr) <= 16 && cap(*arr) >= 4 {
		cop.smallArrayPool.Put(arr)
	}
}

// PutLargeArray returns a large array to the pool
func (cop *ComplexOperationPool) PutLargeArray(arr *[]Value) {
	// Only pool large arrays with reasonable upper bound to avoid excessive memory usage
	if cap(*arr) > 1024 && cap(*arr) <= 8192 {
		cop.largeArrayPool.Put(arr)
	}
}

// PutIntSlice returns an int slice to the pool
func (cop *ComplexOperationPool) PutIntSlice(slice *[]int) {
	if cap(*slice) <= 128 && cap(*slice) >= 8 {
		cop.intSlicePool.Put(slice)
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
func (cop *ComplexOperationPool) PutMap(m *map[string]Value) {
	// Only pool maps with reasonable size to avoid memory bloat
	if len(*m) <= 256 {
		// Clear the map before returning to pool
		for k := range *m {
			delete(*m, k)
		}
		cop.mapPool.Put(*m)
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
	// Only pool string builders with reasonable capacity
	if sb.Cap() <= 4096 {
		sb.Reset()
		cop.stringPool.Put(sb)
	}
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
	// Use appropriate pool based on input array size
	var result []Value
	var poolType int // 0=small, 1=medium, 2=large

	if len(arr) <= 32 {
		result = globalComplexPool.GetSmallArray()
		poolType = 0
	} else if len(arr) > 2048 {
		result = globalComplexPool.GetLargeArray()
		poolType = 2
	} else {
		result = globalComplexPool.GetArray()
		poolType = 1
	}

	defer func() {
		switch poolType {
		case 0:
			globalComplexPool.PutSmallArray(&result)
		case 1:
			globalComplexPool.PutArray(&result)
		case 2:
			globalComplexPool.PutLargeArray(&result)
		}
	}()

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

// ========================================
// MEMORY OPTIMIZATION UTILITIES
// ========================================

// GetPooledArray returns an appropriately sized pooled array
func GetPooledArray(expectedSize int) []Value {
	if expectedSize <= 16 {
		return globalComplexPool.GetSmallArray()
	} else if expectedSize > 1024 {
		return globalComplexPool.GetLargeArray()
	}
	return globalComplexPool.GetArray()
}

// PutPooledArray returns an array to the appropriate pool
func PutPooledArray(arr []Value) {
	if cap(arr) <= 16 {
		globalComplexPool.PutSmallArray(&arr)
	} else if cap(arr) > 1024 {
		globalComplexPool.PutLargeArray(&arr)
	} else {
		globalComplexPool.PutArray(&arr)
	}
}

// GetPooledMap returns a pooled map
func GetPooledMap() map[string]Value {
	return globalComplexPool.GetMap()
}

// PutPooledMap returns a map to the pool
func PutPooledMap(m map[string]Value) {
	globalComplexPool.PutMap(&m)
}

// GetPooledStringBuilder returns a pooled string builder
func GetPooledStringBuilder() *strings.Builder {
	return globalComplexPool.GetStringBuilder()
}

// PutPooledStringBuilder returns a string builder to the pool
func PutPooledStringBuilder(sb *strings.Builder) {
	globalComplexPool.PutStringBuilder(sb)
}

// OptimizedMakeSlice creates a slice with optimized capacity
func OptimizedMakeSlice(length, expectedGrowth int) []Value {
	// Pre-allocate with some buffer to reduce reallocations
	capacity := length + expectedGrowth
	if capacity < 8 {
		capacity = 8
	} else if capacity > 100000 {
		// For very large arrays, use chunked growth to avoid excessive memory usage
		capacity = length + (expectedGrowth / 4) // More conservative growth
		if capacity > 100000 {
			capacity = 100000 // Cap at 100k to prevent memory bloat
		}
	} else if capacity > 1024 {
		capacity = 1024
	}
	return make([]Value, length, capacity)
}

// OptimizedMakeMap creates a map with optimized initial capacity
func OptimizedMakeMap(expectedSize int) map[string]Value {
	if expectedSize <= 0 {
		return make(map[string]Value, 8)
	}
	// Add 25% buffer to reduce hash collisions
	capacity := expectedSize + (expectedSize / 4)
	return make(map[string]Value, capacity)
}

// ========================================
// LARGE ARRAY HANDLING UTILITIES
// ========================================

// ChunkedArrayProcessor processes large arrays in chunks to reduce memory pressure
type ChunkedArrayProcessor struct {
	chunkSize int
}

// NewChunkedArrayProcessor creates a new chunked array processor
func NewChunkedArrayProcessor(chunkSize int) *ChunkedArrayProcessor {
	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size
	}
	return &ChunkedArrayProcessor{chunkSize: chunkSize}
}

// ProcessInChunks processes a large array in chunks with a given processor function
func (cap *ChunkedArrayProcessor) ProcessInChunks(arr []Value, processor func([]Value) []Value) []Value {
	if len(arr) <= cap.chunkSize {
		// Small array, process directly
		return processor(arr)
	}

	// Process in chunks
	result := GetPooledArray(len(arr))
	defer PutPooledArray(result)

	for i := 0; i < len(arr); i += cap.chunkSize {
		end := i + cap.chunkSize
		if end > len(arr) {
			end = len(arr)
		}

		chunk := arr[i:end]
		processed := processor(chunk)
		result = append(result, processed...)
	}

	// Return a copy
	final := make([]Value, len(result))
	copy(final, result)
	return final
}

// OptimizedLargeArrayFilter filters large arrays efficiently using chunking
func OptimizedLargeArrayFilter(arr []Value, predicate func(Value) bool) []Value {
	if len(arr) <= 10000 {
		// Use regular optimized filter for smaller arrays
		return OptimizedArrayFilter(arr, predicate)
	}

	// Use chunked processing for very large arrays
	processor := NewChunkedArrayProcessor(5000)
	return processor.ProcessInChunks(arr, func(chunk []Value) []Value {
		return OptimizedArrayFilter(chunk, predicate)
	})
}

// GetOptimalArrayCapacity returns optimal capacity based on expected size
func GetOptimalArrayCapacity(expectedSize int) int {
	if expectedSize <= 16 {
		return 16
	} else if expectedSize <= 64 {
		return 64
	} else if expectedSize <= 1024 {
		return 1024
	} else if expectedSize <= 8192 {
		return 8192
	} else {
		// For very large arrays, use a more conservative approach
		return expectedSize + (expectedSize / 8) // 12.5% buffer
	}
}
