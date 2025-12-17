package interpreter

import (
	"strings"
	"sync"
	"sync/atomic"
)

// Pool size thresholds - standardized constants
const (
	SmallArrayThreshold  = 16
	MediumArrayThreshold = 1024
	LargeArrayThreshold  = 8192
	MaxPooledMapSize     = 100
	MaxPooledArrayCap    = 8192
	MaxStringBuilderCap  = 4096
)

// PoolMetrics tracks pool usage statistics
type PoolMetrics struct {
	GetCount      int64
	PutCount      int64
	HitCount      int64
	MissCount     int64
	AvgPooledSize int64
	MaxPooledSize int64
}

// BoundedPool is a sync.Pool with capacity limits to prevent memory bloat
type BoundedPool struct {
	pool     sync.Pool
	mu       sync.Mutex
	maxSize  int
	currSize int
	metrics  PoolMetrics
}

// NewBoundedPool creates a new bounded pool with capacity limits
func NewBoundedPool(maxSize int, newFunc func() interface{}) *BoundedPool {
	if maxSize <= 0 {
		maxSize = 100 // Default max size
	}
	return &BoundedPool{
		pool: sync.Pool{
			New: newFunc,
		},
		maxSize: maxSize,
		metrics: PoolMetrics{},
	}
}

// Get gets an item from the pool, respecting capacity limits
func (bp *BoundedPool) Get() interface{} {
	bp.mu.Lock()
	atomic.AddInt64(&bp.metrics.GetCount, 1)

	var item interface{}
	if bp.currSize <= 0 {
		// Pool is empty, create new item
		bp.mu.Unlock()
		item = bp.pool.New()
		atomic.AddInt64(&bp.metrics.MissCount, 1)
		return item
	}

	bp.currSize--
	bp.mu.Unlock()

	// Try to get from pool
	if item = bp.pool.Get(); item != nil {
		atomic.AddInt64(&bp.metrics.HitCount, 1)
		return item
	}

	// Pool returned nil, create new
	atomic.AddInt64(&bp.metrics.MissCount, 1)
	return bp.pool.New()
}

// Put returns an item to the pool, respecting capacity limits
func (bp *BoundedPool) Put(item interface{}) {
	if item == nil {
		return
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.currSize >= bp.maxSize {
		// Pool is full, don't pool it - let GC handle
		atomic.AddInt64(&bp.metrics.PutCount, 1)
		return
	}

	bp.currSize++
	atomic.AddInt64(&bp.metrics.PutCount, 1)
	bp.pool.Put(item)
}

// GetMetrics returns current pool metrics
func (bp *BoundedPool) GetMetrics() PoolMetrics {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return PoolMetrics{
		GetCount:      atomic.LoadInt64(&bp.metrics.GetCount),
		PutCount:      atomic.LoadInt64(&bp.metrics.PutCount),
		HitCount:      atomic.LoadInt64(&bp.metrics.HitCount),
		MissCount:     atomic.LoadInt64(&bp.metrics.MissCount),
		AvgPooledSize: bp.metrics.AvgPooledSize,
		MaxPooledSize: bp.metrics.MaxPooledSize,
	}
}

// UnifiedPoolManager provides a single source of truth for all memory pools
type UnifiedPoolManager struct {
	// Array pools by size class
	smallArrayPool  *BoundedPool // capacity <= 16
	mediumArrayPool *BoundedPool // capacity 17-1024
	largeArrayPool  *BoundedPool // capacity 1025-8192

	// Map pool with optimized clearing
	mapPool *BoundedPool

	// String builder pool
	stringBuilderPool *BoundedPool

	// Int slice pool
	intSlicePool *BoundedPool
}

// NewUnifiedPoolManager creates a new unified pool manager
func NewUnifiedPoolManager() *UnifiedPoolManager {
	return &UnifiedPoolManager{
		smallArrayPool: NewBoundedPool(100, func() interface{} {
			arr := make([]Value, 0, SmallArrayThreshold)
			return &arr
		}),
		mediumArrayPool: NewBoundedPool(50, func() interface{} {
			arr := make([]Value, 0, 64)
			return &arr
		}),
		largeArrayPool: NewBoundedPool(20, func() interface{} {
			arr := make([]Value, 0, 2048)
			return &arr
		}),
		mapPool: NewBoundedPool(50, func() interface{} {
			return make(map[string]Value, 16)
		}),
		stringBuilderPool: NewBoundedPool(50, func() interface{} {
			return &strings.Builder{}
		}),
		intSlicePool: NewBoundedPool(50, func() interface{} {
			slice := make([]int, 0, 32)
			return &slice
		}),
	}
}

// GetArray gets a pooled array based on expected size
func (upm *UnifiedPoolManager) GetArray(expectedSize int) []Value {
	var pool *BoundedPool

	switch {
	case expectedSize <= SmallArrayThreshold:
		pool = upm.smallArrayPool
	case expectedSize <= MediumArrayThreshold:
		pool = upm.mediumArrayPool
	case expectedSize <= LargeArrayThreshold:
		pool = upm.largeArrayPool
	default:
		// Very large arrays: don't pool, allocate directly
		return make([]Value, 0, expectedSize)
	}

	if ptr := pool.Get(); ptr != nil {
		if arr, ok := ptr.(*[]Value); ok {
			return (*arr)[:0]
		}
	}

	// Fallback allocation
	if expectedSize <= SmallArrayThreshold {
		return make([]Value, 0, SmallArrayThreshold)
	} else if expectedSize <= MediumArrayThreshold {
		return make([]Value, 0, 64)
	}
	return make([]Value, 0, 2048)
}

// PutArray returns an array to the appropriate pool
func (upm *UnifiedPoolManager) PutArray(arr []Value) {
	if arr == nil {
		return
	}

	capacity := cap(arr)

	// Don't pool arrays that are too large or too small
	if capacity > MaxPooledArrayCap || capacity < 4 {
		return
	}

	var pool *BoundedPool
	switch {
	case capacity <= SmallArrayThreshold:
		pool = upm.smallArrayPool
	case capacity <= MediumArrayThreshold:
		pool = upm.mediumArrayPool
	case capacity <= LargeArrayThreshold:
		pool = upm.largeArrayPool
	default:
		return
	}

	pool.Put(&arr)
}

// GetMap gets a pooled map with optimized clearing (lazy clearing)
func (upm *UnifiedPoolManager) GetMap() map[string]Value {
	if ptr := upm.mapPool.Get(); ptr != nil {
		if m, ok := ptr.(map[string]Value); ok {
			// Lazy clearing: only clear if map has entries and is small
			// For empty or very large maps, let Go's map handle it
			if len(m) > 0 && len(m) <= MaxPooledMapSize {
				// Clear only small maps to avoid expensive O(n) operation
				for k := range m {
					delete(m, k)
				}
			}
			return m
		}
	}
	return make(map[string]Value, 16)
}

// PutMap returns a map to the pool with optimized clearing strategy
func (upm *UnifiedPoolManager) PutMap(m map[string]Value) {
	if m == nil {
		return
	}

	mapSize := len(m)

	// Only pool small maps to avoid memory bloat
	if mapSize > MaxPooledMapSize {
		// Large maps: don't pool, let GC handle
		return
	}

	// Only clear small maps (lazy clearing strategy)
	// Large maps will be GC'd anyway
	if mapSize > 0 && mapSize <= MaxPooledMapSize {
		// Clear small maps before pooling
		for k := range m {
			delete(m, k)
		}
	}

	upm.mapPool.Put(m)
}

// GetStringBuilder gets a pooled string builder
func (upm *UnifiedPoolManager) GetStringBuilder() *strings.Builder {
	if ptr := upm.stringBuilderPool.Get(); ptr != nil {
		if sb, ok := ptr.(*strings.Builder); ok {
			sb.Reset()
			return sb
		}
	}
	return &strings.Builder{}
}

// PutStringBuilder returns a string builder to the pool
func (upm *UnifiedPoolManager) PutStringBuilder(sb *strings.Builder) {
	if sb == nil {
		return
	}

	// Only pool string builders with reasonable capacity
	if sb.Cap() > MaxStringBuilderCap {
		return
	}

	sb.Reset()
	upm.stringBuilderPool.Put(sb)
}

// GetIntSlice gets a pooled int slice
func (upm *UnifiedPoolManager) GetIntSlice() []int {
	if ptr := upm.intSlicePool.Get(); ptr != nil {
		if slice, ok := ptr.(*[]int); ok {
			return (*slice)[:0]
		}
	}
	return make([]int, 0, 32)
}

// PutIntSlice returns an int slice to the pool
func (upm *UnifiedPoolManager) PutIntSlice(slice []int) {
	if slice == nil {
		return
	}

	capacity := cap(slice)
	if capacity > 128 || capacity < 8 {
		return
	}

	upm.intSlicePool.Put(&slice)
}

// GetAllMetrics returns combined metrics from all pools
func (upm *UnifiedPoolManager) GetAllMetrics() map[string]PoolMetrics {
	return map[string]PoolMetrics{
		"smallArray":    upm.smallArrayPool.GetMetrics(),
		"mediumArray":   upm.mediumArrayPool.GetMetrics(),
		"largeArray":    upm.largeArrayPool.GetMetrics(),
		"map":           upm.mapPool.GetMetrics(),
		"stringBuilder": upm.stringBuilderPool.GetMetrics(),
		"intSlice":      upm.intSlicePool.GetMetrics(),
	}
}

// Global unified pool manager - single source of truth
var globalPoolManager = NewUnifiedPoolManager()

// GetPooledArray returns an appropriately sized pooled array (unified API)
func GetPooledArray(expectedSize int) []Value {
	return globalPoolManager.GetArray(expectedSize)
}

// PutPooledArray returns an array to the appropriate pool (unified API)
func PutPooledArray(arr []Value) {
	globalPoolManager.PutArray(arr)
}

// GetPooledMap returns a pooled map (unified API)
func GetPooledMap() map[string]Value {
	return globalPoolManager.GetMap()
}

// PutPooledMap returns a map to the pool (unified API)
func PutPooledMap(m map[string]Value) {
	globalPoolManager.PutMap(m)
}

// GetPooledStringBuilder returns a pooled string builder (unified API)
func GetPooledStringBuilder() *strings.Builder {
	return globalPoolManager.GetStringBuilder()
}

// PutPooledStringBuilder returns a string builder to the pool (unified API)
func PutPooledStringBuilder(sb *strings.Builder) {
	globalPoolManager.PutStringBuilder(sb)
}

// GetPooledIntSlice returns a pooled int slice (unified API)
func GetPooledIntSlice() []int {
	return globalPoolManager.GetIntSlice()
}

// PutPooledIntSlice returns an int slice to the pool (unified API)
func PutPooledIntSlice(slice []int) {
	globalPoolManager.PutIntSlice(slice)
}

// GetPoolMetrics returns metrics for all pools
func GetPoolMetrics() map[string]PoolMetrics {
	return globalPoolManager.GetAllMetrics()
}

// ========================================
// UTILITY FUNCTIONS (moved from pools.go)
// ========================================

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

// GetOptimalArrayCapacity returns optimal capacity based on expected size
// Uses standardized thresholds
func GetOptimalArrayCapacity(expectedSize int) int {
	switch {
	case expectedSize <= SmallArrayThreshold:
		return SmallArrayThreshold
	case expectedSize <= 64:
		return 64
	case expectedSize <= MediumArrayThreshold:
		return MediumArrayThreshold
	case expectedSize <= LargeArrayThreshold:
		return LargeArrayThreshold
	default:
		// For very large arrays, use a more conservative approach
		return expectedSize + (expectedSize / 8) // 12.5% buffer
	}
}

// OptimizedArrayFilter provides efficient array filtering using unified pools
func OptimizedArrayFilter(arr []Value, predicate func(Value) bool) []Value {
	// Estimate result size (assume 50% will pass filter on average)
	estimatedSize := len(arr) / 2
	if estimatedSize == 0 {
		estimatedSize = 1
	}

	// Use unified pool manager
	result := GetPooledArray(estimatedSize)
	defer PutPooledArray(result)

	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}

	// Return a copy to avoid pool interference
	// Caller owns the returned slice, pool is already returned via defer
	final := make([]Value, len(result))
	copy(final, result)
	return final
}

// OptimizedStringJoin provides efficient string joining with pre-allocation
func OptimizedStringJoin(arr []Value, separator string) string {
	if len(arr) == 0 {
		return ""
	}

	if len(arr) == 1 {
		return toString(arr[0], false)
	}

	sb := GetPooledStringBuilder()
	defer PutPooledStringBuilder(sb)

	for i, v := range arr {
		if i > 0 {
			sb.WriteString(separator)
		}
		sb.WriteString(toString(v, false))
	}

	return sb.String()
}

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

	// Process in chunks using unified pool manager
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

// InternString provides global string interning (wrapper for enhanced interner)
// This replaces the simple StringInterner from pools.go
func InternString(s string) string {
	return InternStringEnhanced(s)
}
