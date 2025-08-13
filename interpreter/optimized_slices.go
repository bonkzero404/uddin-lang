package interpreter

import (
	"sync"
)

// OptimizedSlicePool provides pooling for slices of different sizes
type OptimizedSlicePool struct {
	smallPool   sync.Pool // 0-16 elements
	mediumPool  sync.Pool // 17-64 elements
	largePool   sync.Pool // 65-256 elements
	xlargePool  sync.Pool // 257+ elements
	stats       SlicePoolStats
	mutex       sync.RWMutex
}

// SlicePoolStats tracks slice pool usage statistics
type SlicePoolStats struct {
	SmallHits   int64
	MediumHits  int64
	LargeHits   int64
	XLargeHits  int64
	Misses      int64
	Allocations int64
}

// NewOptimizedSlicePool creates a new optimized slice pool
func NewOptimizedSlicePool() *OptimizedSlicePool {
	return &OptimizedSlicePool{
		smallPool: sync.Pool{
			New: func() any {
				return make([]Value, 0, 16)
			},
		},
		mediumPool: sync.Pool{
			New: func() any {
				return make([]Value, 0, 64)
			},
		},
		largePool: sync.Pool{
			New: func() any {
				return make([]Value, 0, 256)
			},
		},
		xlargePool: sync.Pool{
			New: func() any {
				return make([]Value, 0, 1024)
			},
		},
	}
}

// GetSlice gets a slice with the specified capacity
func (osp *OptimizedSlicePool) GetSlice(capacity int) []Value {
	osp.mutex.Lock()
	defer osp.mutex.Unlock()
	
	var slice []Value
	
	switch {
	case capacity <= 16:
		slice = osp.smallPool.Get().([]Value)
		osp.stats.SmallHits++
	case capacity <= 64:
		slice = osp.mediumPool.Get().([]Value)
		osp.stats.MediumHits++
	case capacity <= 256:
		slice = osp.largePool.Get().([]Value)
		osp.stats.LargeHits++
	case capacity <= 1024:
		slice = osp.xlargePool.Get().([]Value)
		osp.stats.XLargeHits++
	default:
		// For very large slices, allocate directly
		slice = make([]Value, 0, capacity)
		osp.stats.Misses++
		osp.stats.Allocations++
		return slice
	}
	
	// Ensure the slice has the right capacity
	if cap(slice) < capacity {
		slice = make([]Value, 0, capacity)
		osp.stats.Allocations++
	}
	
	return slice[:0] // Reset length to 0
}

// PutSlice returns a slice to the appropriate pool
func (osp *OptimizedSlicePool) PutSlice(slice []Value) {
	if slice == nil {
		return
	}
	
	capacity := cap(slice)
	slice = slice[:0] // Reset length
	
	switch {
	case capacity <= 16:
		osp.smallPool.Put(slice)
	case capacity <= 64:
		osp.mediumPool.Put(slice)
	case capacity <= 256:
		osp.largePool.Put(slice)
	case capacity <= 1024:
		osp.xlargePool.Put(slice)
	// Don't pool very large slices
	}
}

// GetStats returns pool statistics
func (osp *OptimizedSlicePool) GetStats() SlicePoolStats {
	osp.mutex.RLock()
	defer osp.mutex.RUnlock()
	return osp.stats
}

// ResetStats resets pool statistics
func (osp *OptimizedSlicePool) ResetStats() {
	osp.mutex.Lock()
	defer osp.mutex.Unlock()
	osp.stats = SlicePoolStats{}
}

// Global optimized slice pool
var globalOptimizedSlicePool = NewOptimizedSlicePool()

// GetOptimizedSlice gets a slice from the global pool
func GetOptimizedSlice(capacity int) []Value {
	return globalOptimizedSlicePool.GetSlice(capacity)
}

// PutOptimizedSlice returns a slice to the global pool
func PutOptimizedSlice(slice []Value) {
	globalOptimizedSlicePool.PutSlice(slice)
}

// GetSlicePoolStats returns statistics from the global slice pool
func GetSlicePoolStats() SlicePoolStats {
	return globalOptimizedSlicePool.GetStats()
}

// SmartSliceBuilder provides intelligent slice building with growth strategies
type SmartSliceBuilder struct {
	slice    []Value
	capacity int
	growth   GrowthStrategy
	pool     *OptimizedSlicePool
}

// GrowthStrategy defines how slices should grow
type GrowthStrategy int

const (
	GrowthLinear      GrowthStrategy = iota // Add fixed amount
	GrowthExponential                       // Double capacity
	GrowthFibonacci                         // Fibonacci growth
	GrowthOptimal                           // Optimal growth based on usage patterns
)

// NewSmartSliceBuilder creates a new smart slice builder
func NewSmartSliceBuilder(initialCapacity int, strategy GrowthStrategy) *SmartSliceBuilder {
	return &SmartSliceBuilder{
		slice:    GetOptimizedSlice(initialCapacity),
		capacity: initialCapacity,
		growth:   strategy,
		pool:     globalOptimizedSlicePool,
	}
}

// Append appends a value to the slice, growing if necessary
func (ssb *SmartSliceBuilder) Append(value Value) {
	if len(ssb.slice) >= cap(ssb.slice) {
		ssb.grow()
	}
	ssb.slice = append(ssb.slice, value)
}

// AppendMultiple appends multiple values efficiently
func (ssb *SmartSliceBuilder) AppendMultiple(values ...Value) {
	neededCapacity := len(ssb.slice) + len(values)
	if neededCapacity > cap(ssb.slice) {
		ssb.growToCapacity(neededCapacity)
	}
	ssb.slice = append(ssb.slice, values...)
}

// grow grows the slice according to the growth strategy
func (ssb *SmartSliceBuilder) grow() {
	currentCap := cap(ssb.slice)
	var newCap int
	
	switch ssb.growth {
	case GrowthLinear:
		newCap = currentCap + 16
	case GrowthExponential:
		newCap = currentCap * 2
		if newCap == 0 {
			newCap = 8
		}
	case GrowthFibonacci:
		newCap = fibonacciGrowth(currentCap)
	case GrowthOptimal:
		newCap = optimalGrowth(currentCap)
	default:
		newCap = currentCap * 2
		if newCap == 0 {
			newCap = 8
		}
	}
	
	ssb.growToCapacity(newCap)
}

// growToCapacity grows the slice to the specified capacity
func (ssb *SmartSliceBuilder) growToCapacity(newCapacity int) {
	newSlice := GetOptimizedSlice(newCapacity)
	copy(newSlice, ssb.slice)
	newSlice = newSlice[:len(ssb.slice)]
	
	// Return old slice to pool
	PutOptimizedSlice(ssb.slice)
	
	ssb.slice = newSlice
}

// Slice returns the current slice
func (ssb *SmartSliceBuilder) Slice() []Value {
	return ssb.slice
}

// Len returns the current length
func (ssb *SmartSliceBuilder) Len() int {
	return len(ssb.slice)
}

// Cap returns the current capacity
func (ssb *SmartSliceBuilder) Cap() int {
	return cap(ssb.slice)
}

// Reset resets the builder
func (ssb *SmartSliceBuilder) Reset() {
	PutOptimizedSlice(ssb.slice)
	ssb.slice = GetOptimizedSlice(ssb.capacity)
}

// Release releases the builder's resources
func (ssb *SmartSliceBuilder) Release() {
	PutOptimizedSlice(ssb.slice)
	ssb.slice = nil
}

// fibonacciGrowth implements Fibonacci growth strategy
func fibonacciGrowth(current int) int {
	if current <= 1 {
		return 2
	}
	
	// Find the Fibonacci number greater than current
	a, b := 1, 1
	for b <= current {
		a, b = b, a+b
	}
	return b
}

// optimalGrowth implements optimal growth strategy based on common usage patterns
func optimalGrowth(current int) int {
	switch {
	case current < 16:
		return 16
	case current < 64:
		return 64
	case current < 256:
		return 256
	case current < 1024:
		return 1024
	default:
		// For large slices, grow by 50%
		return current + current/2
	}
}

// SliceBuilderPool provides pooling for slice builders
type SliceBuilderPool struct {
	pool sync.Pool
}

// NewSliceBuilderPool creates a new slice builder pool
func NewSliceBuilderPool() *SliceBuilderPool {
	return &SliceBuilderPool{
		pool: sync.Pool{
			New: func() any {
				return NewSmartSliceBuilder(16, GrowthOptimal)
			},
		},
	}
}

// Get gets a slice builder from the pool
func (sbp *SliceBuilderPool) Get() *SmartSliceBuilder {
	return sbp.pool.Get().(*SmartSliceBuilder)
}

// Put returns a slice builder to the pool
func (sbp *SliceBuilderPool) Put(builder *SmartSliceBuilder) {
	builder.Reset()
	sbp.pool.Put(builder)
}

// Global slice builder pool
var globalSliceBuilderPool = NewSliceBuilderPool()

// GetSliceBuilder gets a slice builder from the global pool
func GetSliceBuilder() *SmartSliceBuilder {
	return globalSliceBuilderPool.Get()
}

// PutSliceBuilder returns a slice builder to the global pool
func PutSliceBuilder(builder *SmartSliceBuilder) {
	globalSliceBuilderPool.Put(builder)
}

// OptimizedSliceOperations provides optimized operations on slices
type OptimizedSliceOperations struct{}

// NewOptimizedSliceOperations creates a new optimized slice operations instance
func NewOptimizedSliceOperations() *OptimizedSliceOperations {
	return &OptimizedSliceOperations{}
}

// Filter filters a slice efficiently
func (oso *OptimizedSliceOperations) Filter(slice []Value, predicate func(Value) bool) []Value {
	builder := GetSliceBuilder()
	defer PutSliceBuilder(builder)
	
	for _, value := range slice {
		if predicate(value) {
			builder.Append(value)
		}
	}
	
	// Copy result to avoid holding reference to pooled slice
	result := make([]Value, builder.Len())
	copy(result, builder.Slice())
	return result
}

// Map maps a slice efficiently
func (oso *OptimizedSliceOperations) Map(slice []Value, mapper func(Value) Value) []Value {
	builder := GetSliceBuilder()
	defer PutSliceBuilder(builder)
	
	for _, value := range slice {
		builder.Append(mapper(value))
	}
	
	// Copy result to avoid holding reference to pooled slice
	result := make([]Value, builder.Len())
	copy(result, builder.Slice())
	return result
}

// Reduce reduces a slice efficiently
func (oso *OptimizedSliceOperations) Reduce(slice []Value, initial Value, reducer func(Value, Value) Value) Value {
	accumulator := initial
	for _, value := range slice {
		accumulator = reducer(accumulator, value)
	}
	return accumulator
}

// Concat concatenates multiple slices efficiently
func (oso *OptimizedSliceOperations) Concat(slices ...[]Value) []Value {
	// Calculate total length
	totalLen := 0
	for _, slice := range slices {
		totalLen += len(slice)
	}
	
	result := GetOptimizedSlice(totalLen)
	defer PutOptimizedSlice(result)
	
	for _, slice := range slices {
		result = append(result, slice...)
	}
	
	// Copy result to avoid holding reference to pooled slice
	finalResult := make([]Value, len(result))
	copy(finalResult, result)
	return finalResult
}

// Global optimized slice operations
var globalOptimizedSliceOps = NewOptimizedSliceOperations()

// FilterSlice filters a slice using the global optimized operations
func FilterSlice(slice []Value, predicate func(Value) bool) []Value {
	return globalOptimizedSliceOps.Filter(slice, predicate)
}

// MapSlice maps a slice using the global optimized operations
func MapSlice(slice []Value, mapper func(Value) Value) []Value {
	return globalOptimizedSliceOps.Map(slice, mapper)
}

// ReduceSlice reduces a slice using the global optimized operations
func ReduceSlice(slice []Value, initial Value, reducer func(Value, Value) Value) Value {
	return globalOptimizedSliceOps.Reduce(slice, initial, reducer)
}

// ConcatSlices concatenates multiple slices using the global optimized operations
func ConcatSlices(slices ...[]Value) []Value {
	return globalOptimizedSliceOps.Concat(slices...)
}

// SliceMetrics provides metrics about slice usage
type SliceMetrics struct {
	TotalAllocations int64
	TotalDeallocations int64
	CurrentActive int64
	PeakActive int64
	AverageSize float64
	mutex sync.RWMutex
}

// Global slice metrics
var globalSliceMetrics = &SliceMetrics{}

// TrackSliceAllocation tracks a slice allocation
func TrackSliceAllocation(size int) {
	globalSliceMetrics.mutex.Lock()
	defer globalSliceMetrics.mutex.Unlock()
	
	globalSliceMetrics.TotalAllocations++
	globalSliceMetrics.CurrentActive++
	
	if globalSliceMetrics.CurrentActive > globalSliceMetrics.PeakActive {
		globalSliceMetrics.PeakActive = globalSliceMetrics.CurrentActive
	}
	
	// Update average size
	if globalSliceMetrics.TotalAllocations > 0 {
		globalSliceMetrics.AverageSize = (globalSliceMetrics.AverageSize*float64(globalSliceMetrics.TotalAllocations-1) + float64(size)) / float64(globalSliceMetrics.TotalAllocations)
	}
}

// TrackSliceDeallocation tracks a slice deallocation
func TrackSliceDeallocation() {
	globalSliceMetrics.mutex.Lock()
	defer globalSliceMetrics.mutex.Unlock()
	
	globalSliceMetrics.TotalDeallocations++
	if globalSliceMetrics.CurrentActive > 0 {
		globalSliceMetrics.CurrentActive--
	}
}

// GetSliceMetrics returns current slice metrics
func GetSliceMetrics() SliceMetrics {
	globalSliceMetrics.mutex.RLock()
	defer globalSliceMetrics.mutex.RUnlock()
	return *globalSliceMetrics
}

// ResetSliceMetrics resets slice metrics
func ResetSliceMetrics() {
	globalSliceMetrics.mutex.Lock()
	defer globalSliceMetrics.mutex.Unlock()
	*globalSliceMetrics = SliceMetrics{}
}