package interpreter

import (
	"hash/fnv"
	"strings"
	"sync"
	"unsafe"
)

// OptimizedMemoCache provides hash-based memoization to reduce string allocation overhead
type OptimizedMemoCache struct {
	cache   map[uint64]Value
	mutex   sync.RWMutex
	maxSize int
	size    int
}

// NewOptimizedMemoCache creates a new optimized memo cache
func NewOptimizedMemoCache(maxSize int) *OptimizedMemoCache {
	return &OptimizedMemoCache{
		cache:   make(map[uint64]Value, maxSize/4), // Pre-allocate with reasonable size
		maxSize: maxSize,
		size:    0,
	}
}

// Global optimized memo cache
var globalOptimizedMemoCache = NewOptimizedMemoCache(10000)

// FastHash computes a fast hash for memoization keys
// Uses FNV-1a hash which is faster than string concatenation
func FastHash(funcName string, args []Value) uint64 {
	h := fnv.New64a()

	// Hash function name
	h.Write([]byte(funcName))

	// Hash each argument based on its type
	for _, arg := range args {
		switch v := arg.(type) {
		case nil:
			h.Write([]byte{0}) // null marker
		case bool:
			if v {
				h.Write([]byte{1})
			} else {
				h.Write([]byte{2})
			}
		case int:
			// Convert int to bytes without allocation
			bytes := (*[8]byte)(unsafe.Pointer(&v))[:8]
			h.Write(bytes)
		case float64:
			// Convert float64 to bytes without allocation
			bytes := (*[8]byte)(unsafe.Pointer(&v))[:8]
			h.Write(bytes)
		case string:
			// Hash string directly
			h.Write([]byte(v))
		default:
			// For complex types, use a simple type marker
			// This is a trade-off between accuracy and performance
			h.Write([]byte{255}) // complex type marker
		}
	}

	return h.Sum64()
}

// Get retrieves a cached value by hash key
func (omc *OptimizedMemoCache) Get(key uint64) (Value, bool) {
	omc.mutex.RLock()
	defer omc.mutex.RUnlock()

	value, exists := omc.cache[key]
	return value, exists
}

// Set stores a value with hash key
func (omc *OptimizedMemoCache) Set(key uint64, value Value) {
	omc.mutex.Lock()
	defer omc.mutex.Unlock()

	// Check if key already exists
	if _, exists := omc.cache[key]; !exists {
		// If cache is full, implement simple LRU-like eviction
		if omc.size >= omc.maxSize {
			// Remove approximately 25% of entries to make room
			toRemove := omc.maxSize / 4
			count := 0
			for k := range omc.cache {
				delete(omc.cache, k)
				count++
				if count >= toRemove {
					break
				}
			}
			omc.size -= toRemove
		}
		omc.size++
	}

	omc.cache[key] = value
}

// Clear clears the cache
func (omc *OptimizedMemoCache) Clear() {
	omc.mutex.Lock()
	defer omc.mutex.Unlock()

	omc.cache = make(map[uint64]Value, omc.maxSize/4)
	omc.size = 0
}

// Size returns the current cache size
func (omc *OptimizedMemoCache) Size() int {
	omc.mutex.RLock()
	defer omc.mutex.RUnlock()
	return omc.size
}

// GetGlobalOptimizedMemoCache returns the global optimized memo cache
func GetGlobalOptimizedMemoCache() *OptimizedMemoCache {
	return globalOptimizedMemoCache
}

// OptimizedMemoKey generates an optimized memo key using hash
func OptimizedMemoKey(funcName string, args []Value) uint64 {
	return FastHash(funcName, args)
}

// PooledStringBuilder provides a pooled string builder for temporary string operations
type PooledStringBuilder struct {
	builders sync.Pool
}

// NewPooledStringBuilder creates a new pooled string builder
func NewPooledStringBuilder() *PooledStringBuilder {
	return &PooledStringBuilder{
		builders: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// Get gets a string builder from the pool
func (psb *PooledStringBuilder) Get() *strings.Builder {
	return psb.builders.Get().(*strings.Builder)
}

// Put returns a string builder to the pool
func (psb *PooledStringBuilder) Put(sb *strings.Builder) {
	sb.Reset()
	psb.builders.Put(sb)
}

// Global pooled string builder
var globalPooledStringBuilder = NewPooledStringBuilder()

// GetOptimizedStringBuilder gets a string builder from the global pool
func GetOptimizedStringBuilder() *strings.Builder {
	return globalPooledStringBuilder.Get()
}

// PutOptimizedStringBuilder returns a string builder to the global pool
func PutOptimizedStringBuilder(sb *strings.Builder) {
	globalPooledStringBuilder.Put(sb)
}

// ValuePool provides pooling for common Value types to reduce allocations
type ValuePool struct {
	intPool   sync.Pool
	floatPool sync.Pool
	boolPool  sync.Pool
}

// NewValuePool creates a new value pool
func NewValuePool() *ValuePool {
	return &ValuePool{
		intPool: sync.Pool{
			New: func() any {
				v := 0
				return &v
			},
		},
		floatPool: sync.Pool{
			New: func() any {
				v := 0.0
				return &v
			},
		},
		boolPool: sync.Pool{
			New: func() any {
				v := false
				return &v
			},
		},
	}
}

// Global value pool
var globalValuePool = NewValuePool()

// GetPooledInt gets a pooled int pointer
func GetPooledInt() *int {
	return globalValuePool.intPool.Get().(*int)
}

// PutPooledInt returns an int pointer to the pool
func PutPooledInt(v *int) {
	*v = 0
	globalValuePool.intPool.Put(v)
}

// GetPooledFloat gets a pooled float64 pointer
func GetPooledFloat() *float64 {
	return globalValuePool.floatPool.Get().(*float64)
}

// PutPooledFloat returns a float64 pointer to the pool
func PutPooledFloat(v *float64) {
	*v = 0.0
	globalValuePool.floatPool.Put(v)
}

// GetPooledBool gets a pooled bool pointer
func GetPooledBool() *bool {
	return globalValuePool.boolPool.Get().(*bool)
}

// PutPooledBool returns a bool pointer to the pool
func PutPooledBool(v *bool) {
	*v = false
	globalValuePool.boolPool.Put(v)
}
