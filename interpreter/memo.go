package interpreter

import (
	"container/list"
	"hash/fnv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// OptimizedMemoCache provides hash-based memoization to reduce string allocation overhead
// EXPERIMENTAL: This memoization cache is experimental and part of memory optimization
// WARNING: May consume significant memory and not suitable for all function types
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
// EXPERIMENTAL: Global memoization cache for experimental function caching
var globalOptimizedMemoCache = NewOptimizedMemoCache(10000)

// ProductionMemoCache provides a production-ready memoization cache with LRU eviction and TTL support
type ProductionMemoCache struct {
	cache    map[uint64]*cacheEntry
	lruList  *list.List
	mutex    sync.RWMutex
	maxSize  int
	ttl      time.Duration
	enabled  bool
}

// cacheEntry represents a single cache entry with metadata
type cacheEntry struct {
	key       uint64
	value     Value
	createdAt time.Time
	lruNode   *list.Element
}

// NewProductionMemoCache creates a new production-ready memo cache
func NewProductionMemoCache(maxSize int, ttl time.Duration) *ProductionMemoCache {
	return &ProductionMemoCache{
		cache:   make(map[uint64]*cacheEntry, maxSize/4),
		lruList: list.New(),
		maxSize: maxSize,
		ttl:     ttl,
		enabled: true,
	}
}

// Get retrieves a value from the cache with TTL check
func (pmc *ProductionMemoCache) Get(key uint64) (Value, bool) {
	if !pmc.enabled {
		return nil, false
	}

	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	entry, exists := pmc.cache[key]
	if !exists {
		return nil, false
	}

	// Check TTL expiration
	if pmc.ttl > 0 && time.Since(entry.createdAt) > pmc.ttl {
		// Remove expired entry
		pmc.removeEntry(entry)
		return nil, false
	}

	// Move to front of LRU list (most recently used)
	pmc.lruList.MoveToFront(entry.lruNode)

	return entry.value, true
}

// Set stores a value in the cache with proper LRU eviction
func (pmc *ProductionMemoCache) Set(key uint64, value Value) {
	if !pmc.enabled {
		return
	}

	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	// Check if key already exists
	if existingEntry, exists := pmc.cache[key]; exists {
		// Update existing entry
		existingEntry.value = value
		existingEntry.createdAt = time.Now()
		pmc.lruList.MoveToFront(existingEntry.lruNode)
		return
	}

	// Evict entries if cache is full
	for len(pmc.cache) >= pmc.maxSize {
		// Remove least recently used entry
		oldest := pmc.lruList.Back()
		if oldest != nil {
			oldEntry := oldest.Value.(*cacheEntry)
			pmc.removeEntry(oldEntry)
		} else {
			break
		}
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		createdAt: time.Now(),
	}

	// Add to front of LRU list
	entry.lruNode = pmc.lruList.PushFront(entry)
	pmc.cache[key] = entry
}

// removeEntry removes an entry from both cache and LRU list
func (pmc *ProductionMemoCache) removeEntry(entry *cacheEntry) {
	delete(pmc.cache, entry.key)
	pmc.lruList.Remove(entry.lruNode)
}

// Clear clears all entries from the cache
func (pmc *ProductionMemoCache) Clear() {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	pmc.cache = make(map[uint64]*cacheEntry, pmc.maxSize/4)
	pmc.lruList = list.New()
}

// Size returns the current number of entries in the cache
func (pmc *ProductionMemoCache) Size() int {
	pmc.mutex.RLock()
	defer pmc.mutex.RUnlock()
	return len(pmc.cache)
}

// SetEnabled enables or disables the cache
func (pmc *ProductionMemoCache) SetEnabled(enabled bool) {
	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()
	pmc.enabled = enabled
}

// IsEnabled returns whether the cache is enabled
func (pmc *ProductionMemoCache) IsEnabled() bool {
	pmc.mutex.RLock()
	defer pmc.mutex.RUnlock()
	return pmc.enabled
}

// CleanupExpired removes all expired entries from the cache
func (pmc *ProductionMemoCache) CleanupExpired() int {
	if pmc.ttl <= 0 {
		return 0
	}

	pmc.mutex.Lock()
	defer pmc.mutex.Unlock()

	expiredKeys := make([]uint64, 0)

	// Find expired entries
	for key, entry := range pmc.cache {
		if time.Since(entry.createdAt) > pmc.ttl {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredKeys {
		if entry, exists := pmc.cache[key]; exists {
			pmc.removeEntry(entry)
		}
	}

	return len(expiredKeys)
}

// Global production memo cache
var globalProductionMemoCache = NewProductionMemoCache(5000, 10*time.Minute)

// GetGlobalProductionMemoCache returns the global production memo cache
func GetGlobalProductionMemoCache() *ProductionMemoCache {
	return globalProductionMemoCache
}

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
