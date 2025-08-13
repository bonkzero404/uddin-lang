package interpreter

import (
	"sync"
	"unsafe"
)

// TaggedValueType represents the type of a value for tagged union optimization
type TaggedValueType uint8

const (
	TaggedNull TaggedValueType = iota
	TaggedBool
	TaggedInt
	TaggedFloat
	TaggedString
	TaggedArray
	TaggedMap
	TaggedFunction
	TaggedOther
)

// TaggedValue represents a value with its type tag for memory optimization
type TaggedValue struct {
	Type TaggedValueType
	_    [7]byte // Padding for alignment
	Data uintptr // Pointer to actual data or small value storage
}

// TaggedValuePool manages a pool of TaggedValue instances
type TaggedValuePool struct {
	pool sync.Pool
}

// NewTaggedValuePool creates a new tagged value pool
func NewTaggedValuePool() *TaggedValuePool {
	return &TaggedValuePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &TaggedValue{}
			},
		},
	}
}

// Global tagged value pool
var globalTaggedValuePool = NewTaggedValuePool()

// Global string storage for TaggedValue strings
var (
	globalStringStore = make(map[int]string)
	globalStringIndex = 0
	globalStringMutex sync.RWMutex
)

func storeStringGlobally(s string) int {
	globalStringMutex.Lock()
	defer globalStringMutex.Unlock()
	globalStringIndex++
	globalStringStore[globalStringIndex] = s
	return globalStringIndex
}

func getStringGlobally(index int) string {
	globalStringMutex.RLock()
	defer globalStringMutex.RUnlock()
	if s, exists := globalStringStore[index]; exists {
		return s
	}
	return ""
}

// GetTaggedValue gets a TaggedValue from the pool
func (tvp *TaggedValuePool) GetTaggedValue() *TaggedValue {
	return tvp.pool.Get().(*TaggedValue)
}

// PutTaggedValue returns a TaggedValue to the pool
func (tvp *TaggedValuePool) PutTaggedValue(tv *TaggedValue) {
	tv.Type = TaggedNull
	tv.Data = 0
	tvp.pool.Put(tv)
}

// GetGlobalTaggedValuePool returns the global tagged value pool
func GetGlobalTaggedValuePool() *TaggedValuePool {
	return globalTaggedValuePool
}

// CreateTaggedValue creates a tagged value from a regular value
// This function now creates persistent memory allocations to avoid use-after-free bugs
func CreateTaggedValue(value Value) *TaggedValue {
	tv := globalTaggedValuePool.GetTaggedValue()
	
	switch v := value.(type) {
	case nil:
		tv.Type = TaggedNull
		tv.Data = 0
	case bool:
		tv.Type = TaggedBool
		if v {
			tv.Data = 1
		} else {
			tv.Data = 0
		}
	case int:
		tv.Type = TaggedInt
		tv.Data = uintptr(v)
	case float64:
		tv.Type = TaggedFloat
		// Allocate persistent memory for float64
		floatPtr := new(float64)
		*floatPtr = v
		tv.Data = uintptr(unsafe.Pointer(floatPtr))
	case string:
		tv.Type = TaggedString
		// Store string in global storage and use index
		index := storeStringGlobally(v)
		tv.Data = uintptr(index)
	case *[]Value:
		tv.Type = TaggedArray
		tv.Data = uintptr(unsafe.Pointer(v))
	case map[string]Value:
		tv.Type = TaggedMap
		// Allocate persistent memory for map
		mapPtr := new(map[string]Value)
		*mapPtr = v
		tv.Data = uintptr(unsafe.Pointer(mapPtr))
	case functionType:
		tv.Type = TaggedFunction
		// Allocate persistent memory for function
		funcPtr := new(functionType)
		*funcPtr = v
		tv.Data = uintptr(unsafe.Pointer(funcPtr))
	default:
		// Fallback for unknown types
		tv.Type = TaggedOther
		tv.Data = 0
	}
	
	return tv
}

// ToValue converts a tagged value back to a regular value
func (tv *TaggedValue) ToValue() Value {
	switch tv.Type {
	case TaggedNull:
		return nil
	case TaggedBool:
		return tv.Data != 0
	case TaggedInt:
		return int(tv.Data)
	case TaggedFloat:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return float64(0)
		}
		return *(*float64)(unsafe.Pointer(tv.Data))
	case TaggedString:
		// Safety check for zero index
		if tv.Data == 0 {
			return ""
		}
		// Retrieve string from global store using index
		return getStringGlobally(int(tv.Data))
	case TaggedArray:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return &[]Value{}
		}
		return (*[]Value)(unsafe.Pointer(tv.Data))
	case TaggedMap:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return map[string]Value{}
		}
		return *(*map[string]Value)(unsafe.Pointer(tv.Data))
	case TaggedFunction:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return nil
		}
		return *(*functionType)(unsafe.Pointer(tv.Data))
	default:
		return nil
	}
}

// IsSmallValue checks if a value can be stored inline
func (tv *TaggedValue) IsSmallValue() bool {
	return tv.Type == TaggedNull || tv.Type == TaggedBool || tv.Type == TaggedInt
}

// CompactEnvironment represents a memory-optimized environment
type CompactEnvironment struct {
	vars     []CompactScope
	args     []string
	stdin    interface{}
	stdout   interface{}
	exit     func(int)
	unittest bool
}

// CompactScope represents a memory-optimized scope
type CompactScope struct {
	variables map[string]*TaggedValue
	size      int32 // Track size for memory management
	_         [4]byte // Padding for alignment
}

// NewCompactEnvironment creates a new compact environment
func NewCompactEnvironment() *CompactEnvironment {
	return &CompactEnvironment{
		vars: []CompactScope{{
			variables: make(map[string]*TaggedValue),
		}},
	}
}

// PushScope adds a new scope to the environment
func (ce *CompactEnvironment) PushScope() {
	ce.vars = append(ce.vars, CompactScope{
		variables: make(map[string]*TaggedValue),
	})
}

// PopScope removes the top scope from the environment
func (ce *CompactEnvironment) PopScope() {
	if len(ce.vars) > 1 {
		// Clean up tagged values in the scope being removed
		topScope := &ce.vars[len(ce.vars)-1]
		for _, tv := range topScope.variables {
			globalTaggedValuePool.PutTaggedValue(tv)
		}
		ce.vars = ce.vars[:len(ce.vars)-1]
	}
}

// Assign assigns a value to a variable in the current scope
func (ce *CompactEnvironment) Assign(name string, value Value) {
	if len(ce.vars) == 0 {
		ce.PushScope()
	}
	
	currentScope := &ce.vars[len(ce.vars)-1]
	
	// Clean up existing value if it exists
	if existingTV, exists := currentScope.variables[name]; exists {
		globalTaggedValuePool.PutTaggedValue(existingTV)
	}
	
	// Create new tagged value using safe inline approach
	tv := globalTaggedValuePool.GetTaggedValue()
	*tv = ce.createTaggedValueInline(value)
	currentScope.variables[name] = tv
	currentScope.size++
}

// Lookup looks up a variable in the environment
func (ce *CompactEnvironment) Lookup(name string) (Value, bool) {
	// Search from innermost to outermost scope
	for i := len(ce.vars) - 1; i >= 0; i-- {
		if tv, exists := ce.vars[i].variables[name]; exists {
			return tv.ToValue(), true
		}
	}
	return nil, false
}

// GetScopeCount returns the number of scopes
func (ce *CompactEnvironment) GetScopeCount() int {
	return len(ce.vars)
}

// GetScopeSize returns the size of a specific scope
func (ce *CompactEnvironment) GetScopeSize(index int) int32 {
	if index >= 0 && index < len(ce.vars) {
		return ce.vars[index].size
	}
	return 0
}

// createTaggedValueInline creates a TaggedValue by value without using the pool
// This avoids use-after-free bugs when storing by value
func (ce *CompactEnvironment) createTaggedValueInline(value Value) TaggedValue {
	var tv TaggedValue
	
	switch v := value.(type) {
	case nil:
		tv.Type = TaggedNull
		tv.Data = 0
	case bool:
		tv.Type = TaggedBool
		if v {
			tv.Data = 1
		} else {
			tv.Data = 0
		}
	case int:
		tv.Type = TaggedInt
		tv.Data = uintptr(v)
	case float64:
		tv.Type = TaggedFloat
		// For inline storage, we need to allocate persistent memory
		floatPtr := new(float64)
		*floatPtr = v
		tv.Data = uintptr(unsafe.Pointer(floatPtr))
	case string:
		tv.Type = TaggedString
		// Store string in global storage and use index
		index := storeStringGlobally(v)
		tv.Data = uintptr(index)
	case *[]Value:
		tv.Type = TaggedArray
		tv.Data = uintptr(unsafe.Pointer(v))
	case map[string]Value:
		tv.Type = TaggedMap
		mapPtr := new(map[string]Value)
		*mapPtr = v
		tv.Data = uintptr(unsafe.Pointer(mapPtr))
	case functionType:
		tv.Type = TaggedFunction
		funcPtr := new(functionType)
		*funcPtr = v
		tv.Data = uintptr(unsafe.Pointer(funcPtr))
	default:
		// Fallback for unknown types
		tv.Type = TaggedOther
		tv.Data = 0
	}
	
	return tv
}

// CacheFriendlyArray represents a cache-friendly array structure
type CacheFriendlyArray struct {
	data     []TaggedValue // Store values inline for better cache locality
	capacity int32
	size     int32
	_        [4]byte // Padding for alignment
}

// NewCacheFriendlyArray creates a new cache-friendly array
func NewCacheFriendlyArray(initialCapacity int) *CacheFriendlyArray {
	return &CacheFriendlyArray{
		data:     make([]TaggedValue, initialCapacity),
		capacity: int32(initialCapacity),
		size:     0,
	}
}

// Append adds a value to the array
func (cfa *CacheFriendlyArray) Append(value Value) {
	if cfa.size >= cfa.capacity {
		// Grow the array
		newCapacity := cfa.capacity * 2
		newData := make([]TaggedValue, newCapacity)
		copy(newData, cfa.data)
		cfa.data = newData
		cfa.capacity = newCapacity
	}
	
	// Create tagged value and store inline
	cfa.data[cfa.size] = cfa.createTaggedValueInline(value)
	cfa.size++
}

// Get retrieves a value from the array
func (cfa *CacheFriendlyArray) Get(index int32) (Value, bool) {
	if index >= 0 && index < cfa.size {
		return cfa.data[index].ToValue(), true
	}
	return nil, false
}

// Set sets a value in the array
func (cfa *CacheFriendlyArray) Set(index int32, value Value) bool {
	if index >= 0 && index < cfa.size {
		cfa.data[index] = cfa.createTaggedValueInline(value)
		return true
	}
	return false
}

// Size returns the current size of the array
func (cfa *CacheFriendlyArray) Size() int32 {
	return cfa.size
}

// ToSlice converts the array to a regular slice
func (cfa *CacheFriendlyArray) ToSlice() []Value {
	result := make([]Value, cfa.size)
	for i := int32(0); i < cfa.size; i++ {
		result[i] = cfa.data[i].ToValue()
	}
	return result
}

// createTaggedValueInline creates a TaggedValue by value without using the pool
// This avoids use-after-free bugs when storing by value
func (cfa *CacheFriendlyArray) createTaggedValueInline(value Value) TaggedValue {
	var tv TaggedValue
	
	switch v := value.(type) {
	case nil:
		tv.Type = TaggedNull
		tv.Data = 0
	case bool:
		tv.Type = TaggedBool
		if v {
			tv.Data = 1
		} else {
			tv.Data = 0
		}
	case int:
		tv.Type = TaggedInt
		tv.Data = uintptr(v)
	case float64:
		tv.Type = TaggedFloat
		// For inline storage, we need to allocate persistent memory
		floatPtr := new(float64)
		*floatPtr = v
		tv.Data = uintptr(unsafe.Pointer(floatPtr))
	case string:
		tv.Type = TaggedString
		// Store string in global storage and use index
		index := storeStringGlobally(v)
		tv.Data = uintptr(index)
	case *[]Value:
		tv.Type = TaggedArray
		tv.Data = uintptr(unsafe.Pointer(v))
	case map[string]Value:
		tv.Type = TaggedMap
		mapPtr := new(map[string]Value)
		*mapPtr = v
		tv.Data = uintptr(unsafe.Pointer(mapPtr))
	case functionType:
		tv.Type = TaggedFunction
		funcPtr := new(functionType)
		*funcPtr = v
		tv.Data = uintptr(unsafe.Pointer(funcPtr))
	default:
		// Fallback for unknown types
		tv.Type = TaggedOther
		tv.Data = 0
	}
	
	return tv
}

// CacheFriendlyMap represents a cache-friendly map structure
type CacheFriendlyMap struct {
	entries  []MapEntry
	capacity int32
	size     int32
	_        [4]byte // Padding for alignment
}

// MapEntry represents a key-value pair in the cache-friendly map
type MapEntry struct {
	key   string
	value TaggedValue
	hash  uint32
	_     [4]byte // Padding for alignment
}

// NewCacheFriendlyMap creates a new cache-friendly map
func NewCacheFriendlyMap(initialCapacity int) *CacheFriendlyMap {
	return &CacheFriendlyMap{
		entries:  make([]MapEntry, initialCapacity),
		capacity: int32(initialCapacity),
		size:     0,
	}
}

// simpleHash computes a simple hash for a string
func simpleHash(s string) uint32 {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// Set sets a key-value pair in the map
func (cfm *CacheFriendlyMap) Set(key string, value Value) {
	hash := simpleHash(key)
	
	// Look for existing key
	for i := int32(0); i < cfm.size; i++ {
		if cfm.entries[i].key == key {
			// Update existing entry
			cfm.entries[i].value = cfm.createTaggedValueInline(value)
			return
		}
	}
	
	// Add new entry
	if cfm.size >= cfm.capacity {
		// Grow the map
		newCapacity := cfm.capacity * 2
		newEntries := make([]MapEntry, newCapacity)
		copy(newEntries, cfm.entries)
		cfm.entries = newEntries
		cfm.capacity = newCapacity
	}
	
	cfm.entries[cfm.size] = MapEntry{
		key:   key,
		value: cfm.createTaggedValueInline(value),
		hash:  hash,
	}
	cfm.size++
}

// Get retrieves a value from the map
func (cfm *CacheFriendlyMap) Get(key string) (Value, bool) {
	for i := int32(0); i < cfm.size; i++ {
		if cfm.entries[i].key == key {
			return cfm.entries[i].value.ToValue(), true
		}
	}
	return nil, false
}

// Delete removes a key-value pair from the map
func (cfm *CacheFriendlyMap) Delete(key string) bool {
	for i := int32(0); i < cfm.size; i++ {
		if cfm.entries[i].key == key {
			// Move last entry to this position
			cfm.entries[i] = cfm.entries[cfm.size-1]
			cfm.size--
			return true
		}
	}
	return false
}

// Size returns the current size of the map
func (cfm *CacheFriendlyMap) Size() int32 {
	return cfm.size
}

// createTaggedValueInline creates a TaggedValue by value without using the pool
// This avoids use-after-free bugs when storing by value
func (cfm *CacheFriendlyMap) createTaggedValueInline(value Value) TaggedValue {
	var tv TaggedValue
	
	switch v := value.(type) {
	case nil:
		tv.Type = TaggedNull
		tv.Data = 0
	case bool:
		tv.Type = TaggedBool
		if v {
			tv.Data = 1
		} else {
			tv.Data = 0
		}
	case int:
		tv.Type = TaggedInt
		tv.Data = uintptr(v)
	case float64:
		tv.Type = TaggedFloat
		// For inline storage, we need to allocate persistent memory
		floatPtr := new(float64)
		*floatPtr = v
		tv.Data = uintptr(unsafe.Pointer(floatPtr))
	case string:
		tv.Type = TaggedString
		// For strings, we'll use a different approach - store in a global map
		// and use the Data field as an index
		index := storeStringGlobally(v)
		tv.Data = uintptr(index)
	case *[]Value:
		tv.Type = TaggedArray
		tv.Data = uintptr(unsafe.Pointer(v))
	case map[string]Value:
		tv.Type = TaggedMap
		mapPtr := new(map[string]Value)
		*mapPtr = v
		tv.Data = uintptr(unsafe.Pointer(mapPtr))
	case functionType:
		tv.Type = TaggedFunction
		funcPtr := new(functionType)
		*funcPtr = v
		tv.Data = uintptr(unsafe.Pointer(funcPtr))
	default:
		// Fallback for unknown types
		tv.Type = TaggedOther
		tv.Data = 0
	}
	
	return tv
}

// ToMap converts the cache-friendly map to a regular map
func (cfm *CacheFriendlyMap) ToMap() map[string]Value {
	result := make(map[string]Value)
	for i := int32(0); i < cfm.size; i++ {
		entry := &cfm.entries[i]
		result[entry.key] = entry.value.ToValue()
	}
	return result
}

// MemoryLayoutOptimizer provides memory layout optimization utilities
type MemoryLayoutOptimizer struct {
	taggedValuePool *TaggedValuePool
	stats           MemoryStats
	mutex           sync.RWMutex
}

// MemoryStats tracks memory optimization statistics
type MemoryStats struct {
	TaggedValuesCreated int64
	TaggedValuesReused  int64
	CacheHits           int64
	CacheMisses         int64
}

// NewMemoryLayoutOptimizer creates a new memory layout optimizer
func NewMemoryLayoutOptimizer() *MemoryLayoutOptimizer {
	return &MemoryLayoutOptimizer{
		taggedValuePool: NewTaggedValuePool(),
	}
}

// Global memory layout optimizer
var globalMemoryLayoutOptimizer = NewMemoryLayoutOptimizer()

// GetGlobalMemoryLayoutOptimizer returns the global memory layout optimizer
func GetGlobalMemoryLayoutOptimizer() *MemoryLayoutOptimizer {
	return globalMemoryLayoutOptimizer
}

// GetMemoryStats returns current memory optimization statistics
func (mlo *MemoryLayoutOptimizer) GetMemoryStats() MemoryStats {
	mlo.mutex.RLock()
	defer mlo.mutex.RUnlock()
	return mlo.stats
}

// IncrementTaggedValuesCreated increments the tagged values created counter
func (mlo *MemoryLayoutOptimizer) IncrementTaggedValuesCreated() {
	mlo.mutex.Lock()
	mlo.stats.TaggedValuesCreated++
	mlo.mutex.Unlock()
}

// IncrementTaggedValuesReused increments the tagged values reused counter
func (mlo *MemoryLayoutOptimizer) IncrementTaggedValuesReused() {
	mlo.mutex.Lock()
	mlo.stats.TaggedValuesReused++
	mlo.mutex.Unlock()
}

// IncrementCacheHits increments the cache hits counter
func (mlo *MemoryLayoutOptimizer) IncrementCacheHits() {
	mlo.mutex.Lock()
	mlo.stats.CacheHits++
	mlo.mutex.Unlock()
}

// IncrementCacheMisses increments the cache misses counter
func (mlo *MemoryLayoutOptimizer) IncrementCacheMisses() {
	mlo.mutex.Lock()
	mlo.stats.CacheMisses++
	mlo.mutex.Unlock()
}

// ResetStats resets all memory optimization statistics
func (mlo *MemoryLayoutOptimizer) ResetStats() {
	mlo.mutex.Lock()
	mlo.stats = MemoryStats{}
	mlo.mutex.Unlock()
}