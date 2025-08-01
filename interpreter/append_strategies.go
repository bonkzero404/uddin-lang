package interpreter

import (
	"sync"
)

// FastSlice provides optimized append operations with pre-allocation and pooling
type FastSlice struct {
	data     []any
	capacity int
	length   int
	pool     *sync.Pool
}

// NewFastSlice creates a new FastSlice with initial capacity
func NewFastSlice(initialCapacity int) *FastSlice {
	if initialCapacity < 16 {
		initialCapacity = 16 // minimum capacity
	}

	return &FastSlice{
		data:     make([]any, initialCapacity),
		capacity: initialCapacity,
		length:   0,
		pool: &sync.Pool{
			New: func() any {
				slice := make([]any, initialCapacity)
				return &slice
			},
		},
	}
}

// FastAppend appends elements with optimized memory allocation
func (fs *FastSlice) FastAppend(elements ...any) {
	neededCapacity := fs.length + len(elements)

	// If we need more capacity, grow the slice efficiently
	if neededCapacity > fs.capacity {
		fs.grow(neededCapacity)
	}

	// Copy elements directly
	copy(fs.data[fs.length:], elements)
	fs.length += len(elements)
}

// grow increases the capacity of the slice using an optimized growth strategy
func (fs *FastSlice) grow(minCapacity int) {
	newCapacity := fs.capacity

	// Use exponential growth with a cap to avoid excessive memory usage
	for newCapacity < minCapacity {
		if newCapacity < 1024 {
			newCapacity *= 2 // Double for small slices
		} else {
			newCapacity += newCapacity / 4 // Grow by 25% for larger slices
		}
	}

	// Try to get a slice from the pool first
	var newData []any
	if pooled := fs.pool.Get(); pooled != nil {
		if pooledPtr, ok := pooled.(*[]any); ok {
			pooledSlice := *pooledPtr
			if cap(pooledSlice) >= newCapacity {
				newData = pooledSlice[:newCapacity]
			} else {
				fs.pool.Put(pooled) // Return to pool if not suitable
				newData = make([]any, newCapacity)
			}
		} else {
			newData = make([]any, newCapacity)
		}
	} else {
		newData = make([]any, newCapacity)
	}

	// Copy existing data
	copy(newData, fs.data[:fs.length])

	// Return old slice to pool if it's large enough to be useful
	if cap(fs.data) >= 16 {
		fs.pool.Put(&fs.data)
	}

	fs.data = newData
	fs.capacity = newCapacity
}

// ToSlice returns the current data as a regular slice
func (fs *FastSlice) ToSlice() []any {
	return fs.data[:fs.length]
}

// Len returns the current length
func (fs *FastSlice) Len() int {
	return fs.length
}

// Cap returns the current capacity
func (fs *FastSlice) Cap() int {
	return fs.capacity
}

// Reset clears the slice but keeps the allocated memory
func (fs *FastSlice) Reset() {
	fs.length = 0
}

// FastAppendGlobal provides a global optimized append function
var _ = &sync.Pool{
	New: func() any {
		return make([]any, 0, 64)
	},
}

// HybridAppendStrategy implements a smart append strategy that chooses
// between built-in append and FastSliceAppend based on data size
type HybridAppendStrategy struct {
	threshold int // threshold for switching to FastSliceAppend
}

// NewHybridAppendStrategy creates a new hybrid append strategy
// with the specified threshold (default: 1000)
func NewHybridAppendStrategy(threshold int) *HybridAppendStrategy {
	if threshold <= 0 {
		threshold = 1000 // default threshold
	}
	return &HybridAppendStrategy{
		threshold: threshold,
	}
}

// AppendValues appends values to a slice using the optimal strategy
func (h *HybridAppendStrategy) AppendValues(slice []Value, values ...Value) []Value {
	// For single element appends, always use built-in append (most common case)
	if len(values) == 1 {
		return append(slice, values...)
	}

	totalSize := len(slice) + len(values)
	// For larger batch operations, use pre-allocation
	if totalSize >= h.threshold && cap(slice) < totalSize {
		newSlice := make([]Value, len(slice), totalSize)
		copy(newSlice, slice)
		return append(newSlice, values...)
	}

	// Default case: use built-in append
	return append(slice, values...)
}

// AppendStrings appends strings to a slice using the optimal strategy
func (h *HybridAppendStrategy) AppendStrings(slice []string, values ...string) []string {
	totalSize := len(slice) + len(values)

	// For small datasets, use built-in append
	if totalSize < h.threshold {
		return append(slice, values...)
	}

	// For large datasets, use pre-allocation instead of FastSlice conversion
	if cap(slice) < totalSize {
		newSlice := make([]string, len(slice), totalSize)
		copy(newSlice, slice)
		return append(newSlice, values...)
	}
	return append(slice, values...)
}

// AppendInterface appends any values using the optimal strategy
func (h *HybridAppendStrategy) AppendInterface(slice []any, values ...any) []any {
	totalSize := len(slice) + len(values)

	// For small datasets, use built-in append
	if totalSize < h.threshold {
		return append(slice, values...)
	}

	// For large datasets, use pre-allocation instead of FastSlice conversion
	if cap(slice) < totalSize {
		newSlice := make([]any, len(slice), totalSize)
		copy(newSlice, slice)
		return append(newSlice, values...)
	}
	return append(slice, values...)
}

// FastSlice-based wrapper functions for better performance
func FastSliceAppendValue(slice []Value, elements ...Value) []Value {
	fs := NewFastSlice(len(slice) + len(elements))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range elements {
		fs.FastAppend(v)
	}
	result := make([]Value, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(Value)
	}
	return result
}

func FastSliceAppendString(slice []string, elements ...string) []string {
	fs := NewFastSlice(len(slice) + len(elements))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range elements {
		fs.FastAppend(v)
	}
	result := make([]string, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(string)
	}
	return result
}

func FastSliceAppendCallFrame(slice []CallFrame, elements ...CallFrame) []CallFrame {
	fs := NewFastSlice(len(slice) + len(elements))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range elements {
		fs.FastAppend(v)
	}
	result := make([]CallFrame, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(CallFrame)
	}
	return result
}

func FastSliceAppendMapValue(slice []map[string]Value, elements ...map[string]Value) []map[string]Value {
	fs := NewFastSlice(len(slice) + len(elements))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range elements {
		fs.FastAppend(v)
	}
	result := make([]map[string]Value, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(map[string]Value)
	}
	return result
}

func FastSliceAppendBlock(slice Block, items ...Statement) Block {
	fs := NewFastSlice(len(slice) + len(items))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range items {
		fs.FastAppend(v)
	}
	result := make(Block, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(Statement)
	}
	return result
}

func FastSliceAppendExpression(slice []Expression, items ...Expression) []Expression {
	fs := NewFastSlice(len(slice) + len(items))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range items {
		fs.FastAppend(v)
	}
	result := make([]Expression, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(Expression)
	}
	return result
}

func FastSliceAppendMapItem(slice []MapItem, items ...MapItem) []MapItem {
	fs := NewFastSlice(len(slice) + len(items))
	for _, v := range slice {
		fs.FastAppend(v)
	}
	for _, v := range items {
		fs.FastAppend(v)
	}
	result := make([]MapItem, fs.Len())
	for i, v := range fs.ToSlice() {
		result[i] = v.(MapItem)
	}
	return result
}

// Smart append functions that use hybrid strategy

// Global strategy instance to avoid repeated allocations
var globalHybridStrategy = NewHybridAppendStrategy(1000)

// SmartAppend is a convenience function that automatically chooses
// the best append strategy based on data size
func SmartAppend(slice []Value, values ...Value) []Value {
	// Fast path for single element appends (most common case)
	if len(values) == 1 {
		return append(slice, values...)
	}
	return globalHybridStrategy.AppendValues(slice, values...)
}

// SmartAppendString is a convenience function for string slices
func SmartAppendString(slice []string, values ...string) []string {
	// Optimized version without unnecessary conversions
	totalSize := len(slice) + len(values)
	if totalSize < 1000 {
		return append(slice, values...)
	}

	// For large datasets, pre-allocate capacity
	if cap(slice) < totalSize {
		newSlice := make([]string, len(slice), totalSize)
		copy(newSlice, slice)
		return append(newSlice, values...)
	}
	return append(slice, values...)
}

// SmartAppendInterface is a convenience function for any slices
func SmartAppendInterface(slice []any, values ...any) []any {
	// Optimized version without unnecessary conversions
	totalSize := len(slice) + len(values)
	if totalSize < 1000 {
		return append(slice, values...)
	}

	// For large datasets, pre-allocate capacity
	if cap(slice) < totalSize {
		newSlice := make([]any, len(slice), totalSize)
		copy(newSlice, slice)
		return append(newSlice, values...)
	}
	return append(slice, values...)
}

// SmartAppendCallFrame uses hybrid strategy for CallFrame slices
func SmartAppendCallFrame(slice []CallFrame, elements ...CallFrame) []CallFrame {
	if len(slice)+len(elements) < 100 {
		return append(slice, elements...)
	}
	// For larger datasets, use pre-allocation
	newSlice := make([]CallFrame, len(slice), len(slice)+len(elements)+10)
	copy(newSlice, slice)
	return append(newSlice, elements...)
}

// SmartAppendMapValue uses hybrid strategy for map[string]Value slices
func SmartAppendMapValue(slice []map[string]Value, elements ...map[string]Value) []map[string]Value {
	if len(slice)+len(elements) < 50 {
		return append(slice, elements...)
	}
	// For larger datasets, use pre-allocation
	newSlice := make([]map[string]Value, len(slice), len(slice)+len(elements)+10)
	copy(newSlice, slice)
	return append(newSlice, elements...)
}

// SmartAppendBlock uses hybrid strategy for Block (Statement slices)
func SmartAppendBlock(slice Block, elements ...Statement) Block {
	if len(slice)+len(elements) < 500 {
		return append(slice, elements...)
	}
	// For larger datasets, use pre-allocation
	newSlice := make(Block, len(slice), len(slice)+len(elements)+50)
	copy(newSlice, slice)
	return append(newSlice, elements...)
}

// SmartAppendExpression uses hybrid strategy for Expression slices
func SmartAppendExpression(slice []Expression, elements ...Expression) []Expression {
	if len(slice)+len(elements) < 500 {
		return append(slice, elements...)
	}
	// For larger datasets, use pre-allocation
	newSlice := make([]Expression, len(slice), len(slice)+len(elements)+50)
	copy(newSlice, slice)
	return append(newSlice, elements...)
}

// SmartAppendMapItem uses hybrid strategy for MapItem slices
func SmartAppendMapItem(slice []MapItem, elements ...MapItem) []MapItem {
	if len(slice)+len(elements) < 200 {
		return append(slice, elements...)
	}
	// For larger datasets, use pre-allocation
	newSlice := make([]MapItem, len(slice), len(slice)+len(elements)+20)
	copy(newSlice, slice)
	return append(newSlice, elements...)
}

// Utility functions

// BatchAppend appends multiple slices efficiently
func BatchAppend(base []any, slices ...[]any) []any {
	totalLen := len(base)
	for _, s := range slices {
		totalLen += len(s)
	}

	// Pre-allocate with exact size needed
	if cap(base) < totalLen {
		newSlice := make([]any, len(base), totalLen)
		copy(newSlice, base)
		base = newSlice
	}

	// Append all slices
	for _, s := range slices {
		base = append(base, s...)
	}

	return base
}

// GetOptimalStrategy returns the recommended strategy based on expected data size
func GetOptimalStrategy(expectedSize int) string {
	if expectedSize < 100 {
		return "built-in append (very small dataset)"
	} else if expectedSize < 1000 {
		return "built-in append (small dataset)"
	} else if expectedSize < 10000 {
		return "FastSliceAppend (medium dataset)"
	} else {
		return "FastSliceAppend with pre-allocation (large dataset)"
	}
}
