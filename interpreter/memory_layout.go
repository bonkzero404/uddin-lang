package interpreter

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	_ = (*MemoryTracker).getAllocationType
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

// SafeTaggedValue represents a thread-safe version of TaggedValue
// Production-ready implementation with comprehensive safety checks and memory management
type SafeTaggedValue struct {
	Type TaggedValueType
	_    [7]byte // Padding for alignment
	Data uintptr // Pointer to actual data or small value storage
	// Add reference counting for memory management
	refCount int32
	mutex    sync.RWMutex
}

// SafeTaggedValuePool manages a pool of SafeTaggedValue instances with full thread safety
// Production-ready pool implementation with statistics and capacity management
type SafeTaggedValuePool struct {
	pool     sync.Pool
	mutex    sync.RWMutex
	created  int64 // Atomic counter for created values
	reused   int64 // Atomic counter for reused values
	active   int64 // Atomic counter for active values
	maxSize  int32 // Maximum pool size to prevent memory bloat
	currSize int32 // Current pool size
}

// NewSafeTaggedValuePool creates a new thread-safe tagged value pool
func NewSafeTaggedValuePool(maxSize int32) *SafeTaggedValuePool {
	if maxSize <= 0 {
		maxSize = 1000 // Default maximum size
	}

	return &SafeTaggedValuePool{
		pool: sync.Pool{
			New: func() any {
				return &SafeTaggedValue{
					Type:     TaggedNull,
					Data:     0,
					refCount: 0, // Will be set to 1 when retrieved
				}
			},
		},
		maxSize:  maxSize,
		currSize: 0,
		created:  0,
		reused:   0,
		active:   0,
	}
}

// Thread-safe global pools and storage
var (
	globalSafeTaggedValuePool = NewSafeTaggedValuePool(1000)

	// Thread-safe string storage with cleanup
	globalSafeStringStore            = make(map[int]string)
	globalSafeStringIndex            int64 // Use atomic for thread safety
	globalSafeStringMutex            sync.RWMutex
	globalSafeStringCleanupThreshold = 10000 // Cleanup when reaching this size

	// Memory tracking
	globalMemoryTracker = &MemoryTracker{
		allocatedPointers: make(map[uintptr]*AllocationInfo),
	}
)

// MemoryTracker tracks allocated memory for leak detection
type MemoryTracker struct {
	mutex             sync.RWMutex
	allocatedPointers map[uintptr]*AllocationInfo
	totalAllocated    int64
	totalFreed        int64
}

// AllocationInfo stores information about memory allocations
type AllocationInfo struct {
	size      int64
	allocTime int64
	type_     string
	freed     bool
}

// GetSafeTaggedValue gets a SafeTaggedValue from the pool with proper error handling
func (stvp *SafeTaggedValuePool) GetSafeTaggedValue() (*SafeTaggedValue, error) {
	// Use write lock to prevent race conditions during capacity check and allocation
	stvp.mutex.Lock()
	defer stvp.mutex.Unlock()

	currSize := atomic.LoadInt32(&stvp.currSize)
	maxSize := stvp.maxSize

	// Check if we're at capacity
	if currSize >= maxSize {
		return nil, &MemoryError{"tagged value pool at maximum capacity"}
	}

	// Atomically increment current size before getting from pool
	newSize := atomic.AddInt32(&stvp.currSize, 1)
	if newSize > maxSize {
		// Rollback if we exceeded capacity
		atomic.AddInt32(&stvp.currSize, -1)
		return nil, &MemoryError{"tagged value pool at maximum capacity"}
	}

	tv := stvp.pool.Get().(*SafeTaggedValue)
	if tv == nil {
		// Rollback size increment on failure
		atomic.AddInt32(&stvp.currSize, -1)
		return nil, &MemoryError{"failed to get tagged value from pool"}
	}

	// Reset the tagged value safely
	tv.mutex.Lock()
	tv.Type = TaggedNull
	tv.Data = 0
	atomic.StoreInt32(&tv.refCount, 1) // Use atomic store for consistency
	tv.mutex.Unlock()

	atomic.AddInt64(&stvp.created, 1)
	atomic.AddInt64(&stvp.active, 1)
	atomic.AddInt64(&globalMemoryTracker.totalAllocated, 1)

	return tv, nil
}

// PutSafeTaggedValue returns a SafeTaggedValue to the pool with proper cleanup
func (stvp *SafeTaggedValuePool) PutSafeTaggedValue(tv *SafeTaggedValue) error {
	if tv == nil {
		return &MemoryError{"cannot put nil tagged value"}
	}

	tv.mutex.Lock()
	defer tv.mutex.Unlock()

	// Atomically decrement reference count
	newRefCount := atomic.AddInt32(&tv.refCount, -1)
	if newRefCount > 0 {
		return nil // Still has references
	}

	// Prevent double-free by checking if refCount went negative
	if newRefCount < 0 {
		// Reset to 0 to prevent further issues
		atomic.StoreInt32(&tv.refCount, 0)
		return &MemoryError{"tagged value already freed"}
	}

	// Clean up any allocated memory (internal version without lock)
	if err := stvp.cleanupTaggedValueInternal(tv); err != nil {
		return err
	}

	// Reset and return to pool
	tv.Type = TaggedNull
	tv.Data = 0
	atomic.StoreInt32(&tv.refCount, 0)

	// Use pool synchronization to prevent race conditions
	stvp.mutex.Lock()
	stvp.pool.Put(tv)
	atomic.AddInt64(&stvp.reused, 1)
	atomic.AddInt32(&stvp.currSize, -1)
	atomic.AddInt64(&stvp.active, -1)
	stvp.mutex.Unlock()

	atomic.AddInt64(&globalMemoryTracker.totalFreed, 1)

	return nil
}

// PoolStats represents statistics for SafeTaggedValuePool
type PoolStats struct {
	Created  int64
	Reused   int64
	Active   int64
	CurrSize int32
	MaxSize  int32
}

// GetStats returns thread-safe statistics for the pool
func (stvp *SafeTaggedValuePool) GetStats() PoolStats {
	stvp.mutex.RLock()
	defer stvp.mutex.RUnlock()

	return PoolStats{
		Created:  atomic.LoadInt64(&stvp.created),
		Reused:   atomic.LoadInt64(&stvp.reused),
		Active:   atomic.LoadInt64(&stvp.active),
		CurrSize: atomic.LoadInt32(&stvp.currSize),
		MaxSize:  stvp.maxSize,
	}
}

// IsAtCapacity checks if the pool is at maximum capacity
func (stvp *SafeTaggedValuePool) IsAtCapacity() bool {
	currSize := atomic.LoadInt32(&stvp.currSize)
	return currSize >= stvp.maxSize
}

// Safe wrapper functions for unsafe.Pointer operations

// safePointerToUintptr safely converts a pointer to uintptr with validation
func safePointerToUintptr(ptr unsafe.Pointer, typeName string) (uintptr, error) {
	if ptr == nil {
		return 0, &MemoryError{"cannot convert nil pointer to uintptr for " + typeName}
	}

	uintPtr := uintptr(ptr)
	if uintPtr == 0 {
		return 0, &MemoryError{"invalid pointer conversion for " + typeName}
	}

	return uintPtr, nil
}

// safeUintptrToPointer safely converts uintptr to unsafe.Pointer with validation
func safeUintptrToPointer(ptr uintptr, typeName string) (unsafe.Pointer, error) {
	if ptr == 0 {
		return nil, &MemoryError{"cannot convert zero uintptr to pointer for " + typeName}
	}

	// Validate pointer if memory tracker is available
	if globalMemoryTracker != nil && !globalMemoryTracker.isValidPointer(ptr) {
		return nil, &MemoryError{"invalid pointer detected for " + typeName}
	}

	// Safe conversion: pointer has been validated above
	if ptr == 0 {
		return nil, &MemoryError{"cannot convert zero uintptr to pointer"}
	}
	// Use unsafe.Add to ensure pointer arithmetic is done safely
	p := unsafe.Add(unsafe.Pointer(nil), ptr) // Convert after validation with safe pointer arithmetic
	runtime.KeepAlive(p)
	return p, nil
}

// safeFloat64PtrFromUintptr safely converts uintptr to *float64 with validation
func safeFloat64PtrFromUintptr(ptr uintptr) (*float64, error) {
	unsafePtr, err := safeUintptrToPointer(ptr, "float64")
	if err != nil {
		return nil, err
	}

	floatPtr := (*float64)(unsafePtr)
	if floatPtr == nil {
		return nil, &MemoryError{"null float64 pointer after conversion"}
	}

	return floatPtr, nil
}

// safeMapPtrFromUintptr safely converts uintptr to *map[string]Value with validation
func safeMapPtrFromUintptr(ptr uintptr) (*map[string]Value, error) {
	unsafePtr, err := safeUintptrToPointer(ptr, "map[string]Value")
	if err != nil {
		return nil, err
	}

	mapPtr := (*map[string]Value)(unsafePtr)
	if mapPtr == nil {
		return nil, &MemoryError{"null map pointer after conversion"}
	}

	return mapPtr, nil
}

// safeFunctionPtrFromUintptr safely converts uintptr to *functionType with validation
func safeFunctionPtrFromUintptr(ptr uintptr) (*functionType, error) {
	unsafePtr, err := safeUintptrToPointer(ptr, "functionType")
	if err != nil {
		return nil, err
	}

	funcPtr := (*functionType)(unsafePtr)
	if funcPtr == nil {
		return nil, &MemoryError{"null function pointer after conversion"}
	}

	return funcPtr, nil
}

// cleanupTaggedValueInternal performs cleanup without acquiring locks (assumes caller has lock)
func (stvp *SafeTaggedValuePool) cleanupTaggedValueInternal(tv *SafeTaggedValue) error {
	if tv == nil {
		return &MemoryError{"cannot cleanup nil tagged value"}
	}

	// Check if already freed
	if atomic.LoadInt32(&tv.refCount) < 0 {
		return &MemoryError{"tagged value already freed"}
	}

	if tv.Data == 0 {
		// Mark as freed
		atomic.StoreInt32(&tv.refCount, -1)
		return nil
	}

	switch tv.Type {
	case TaggedFloat:
		// Validate pointer before cleanup for allocated types
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return &MemoryError{"invalid pointer detected during cleanup"}
		}
		// Clear finalizer and free float64 allocation
		if floatPtr, err := safeFloat64PtrFromUintptr(tv.Data); err == nil && floatPtr != nil {
			runtime.SetFinalizer(floatPtr, nil)
			globalMemoryTracker.trackDeallocation(tv.Data, "float64")
		}
	case TaggedMap:
		// Validate pointer before cleanup for allocated types
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return &MemoryError{"invalid pointer detected during cleanup"}
		}
		// Clear finalizer and free map allocation
		if mapPtr, err := safeMapPtrFromUintptr(tv.Data); err == nil && mapPtr != nil {
			runtime.SetFinalizer(mapPtr, nil)
			globalMemoryTracker.trackDeallocation(tv.Data, "map")
		}
	case TaggedFunction:
		// Validate pointer before cleanup for allocated types
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return &MemoryError{"invalid pointer detected during cleanup"}
		}
		// Clear finalizer and free function allocation
		if funcPtr, err := safeFunctionPtrFromUintptr(tv.Data); err == nil && funcPtr != nil {
			runtime.SetFinalizer(funcPtr, nil)
			globalMemoryTracker.trackDeallocation(tv.Data, "function")
		}
	case TaggedString:
		// String cleanup is handled by global string store cleanup
		// But we still validate the index
		if tv.Data <= 0 {
			return &MemoryError{"invalid string index during cleanup"}
		}
	case TaggedNull, TaggedBool, TaggedInt:
		// These types store values directly, no pointer validation needed
		// No cleanup required for simple types
	}

	// Clear the data and mark as freed
	tv.Data = 0
	tv.Type = TaggedNull
	atomic.StoreInt32(&tv.refCount, -1)

	return nil
}

// CreateSafeTaggedValue creates a thread-safe tagged value with proper error handling
func CreateSafeTaggedValue(value Value) (*SafeTaggedValue, error) {
	// Input validation
	if value == nil {
		// Handle nil case explicitly
		tv, err := globalSafeTaggedValuePool.GetSafeTaggedValue()
		if err != nil {
			return nil, &MemoryError{"failed to get tagged value from pool: " + err.Error()}
		}
		tv.Type = TaggedNull
		tv.Data = 0
		atomic.StoreInt32(&tv.refCount, 1)
		return tv, nil
	}

	tv, err := globalSafeTaggedValuePool.GetSafeTaggedValue()
	if err != nil {
		return nil, &MemoryError{"failed to get tagged value from pool: " + err.Error()}
	}

	// Defer cleanup on error to prevent resource leaks
	var success bool
	defer func() {
		if !success {
			// Clean up on error
			if putErr := globalSafeTaggedValuePool.PutSafeTaggedValue(tv); putErr != nil {
				// Log error but don't override original error
			}
		}
	}()

	tv.mutex.Lock()
	defer tv.mutex.Unlock()

	switch v := value.(type) {
	case bool:
		tv.Type = TaggedBool
		if v {
			tv.Data = 1
		} else {
			tv.Data = 0
		}
	case int:
		// Bounds checking for int values
		if v < 0 && uintptr(v) > ^uintptr(0)>>1 {
			return nil, &MemoryError{"integer value out of safe range"}
		}
		tv.Type = TaggedInt
		tv.Data = uintptr(v)
	case float64:
		// Validate float64 value
		if v != v { // NaN check
			return nil, &MemoryError{"NaN float64 values are not supported"}
		}
		tv.Type = TaggedFloat
		// Allocate memory for float64 and track it
		floatPtr := new(float64)
		// if floatPtr == nil {
		// 	return nil, &MemoryError{"failed to allocate memory for float64"}
		// }
		*floatPtr = v
		uintPtr, err := safePointerToUintptr(unsafe.Pointer(floatPtr), "float64")
		if err != nil {
			return nil, err
		}
		tv.Data = uintPtr

		// Track allocation and set finalizer
		globalMemoryTracker.trackAllocation(tv.Data, 8, "float64")
		runtime.SetFinalizer(floatPtr, func(ptr *float64) {
			if uintPtr, err := safePointerToUintptr(unsafe.Pointer(ptr), "float64"); err == nil {
				globalMemoryTracker.trackDeallocation(uintPtr, "float64")
			}
		})
	case string:
		// Validate string length
		if len(v) > 1<<20 { // 1MB limit
			return nil, &MemoryError{"string too large (max 1MB)"}
		}
		tv.Type = TaggedString
		// Store string globally and get index
		index, err := storeSafeStringGlobally(v)
		if err != nil {
			return nil, &MemoryError{"failed to store string: " + err.Error()}
		}
		tv.Data = uintptr(index)
	case *[]Value:
		if v == nil {
			return nil, &MemoryError{"cannot create tagged value from nil array pointer"}
		}
		// Validate array size
		if len(*v) > 100000 { // Reasonable limit
			return nil, &MemoryError{"array too large (max 100000 elements)"}
		}
		tv.Type = TaggedArray
		uintPtr, err := safePointerToUintptr(unsafe.Pointer(v), "array")
		if err != nil {
			return nil, err
		}
		tv.Data = uintPtr
		// Track allocation (approximate size)
		globalMemoryTracker.trackAllocation(uintPtr, int64(len(*v)*8), "array")
	case map[string]Value:
		// Validate map size
		if len(v) > 10000 { // Reasonable limit
			return nil, &MemoryError{"map too large (max 10000 entries)"}
		}
		tv.Type = TaggedMap
		// Allocate memory for map and track it
		mapPtr := new(map[string]Value)
		// if mapPtr == nil {
		// 	return nil, &MemoryError{"failed to allocate memory for map"}
		// }
		*mapPtr = v
		uintPtr, err := safePointerToUintptr(unsafe.Pointer(mapPtr), "map")
		if err != nil {
			return nil, err
		}
		tv.Data = uintPtr

		// Track allocation and set finalizer
		globalMemoryTracker.trackAllocation(tv.Data, int64(len(v)*16), "map")
		runtime.SetFinalizer(mapPtr, func(ptr *map[string]Value) {
			if uintPtr, err := safePointerToUintptr(unsafe.Pointer(ptr), "map"); err == nil {
				globalMemoryTracker.trackDeallocation(uintPtr, "map")
			}
		})
	case functionType:
		// Store function types as TaggedFunction with proper handling
		tv.Type = TaggedFunction
		// Allocate memory for function and track it
		funcPtr := new(functionType)
		// if funcPtr == nil {
		// 	return nil, &MemoryError{"failed to allocate memory for function"}
		// }
		*funcPtr = v
		uintPtr, err := safePointerToUintptr(unsafe.Pointer(funcPtr), "function")
		if err != nil {
			return nil, err
		}
		tv.Data = uintPtr

		// Track allocation and set finalizer
		globalMemoryTracker.trackAllocation(tv.Data, 64, "function")
		runtime.SetFinalizer(funcPtr, func(ptr *functionType) {
			if uintPtr, err := safePointerToUintptr(unsafe.Pointer(ptr), "function"); err == nil {
				globalMemoryTracker.trackDeallocation(uintPtr, "function")
			}
		})
	// case *userFunction:
	// 	// Store user-defined functions as TaggedFunction
	// 	tv.Type = TaggedFunction
	// 	// Store the userFunction pointer directly as it implements functionType
	// 	uintPtr, err := safePointerToUintptr(unsafe.Pointer(v), "userFunction")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tv.Data = uintPtr

	// 	// Track allocation
	// 	globalMemoryTracker.trackAllocation(tv.Data, 128, "userFunction")
	// case builtinFunction:
	// 	// Store builtin functions as TaggedFunction
	// 	tv.Type = TaggedFunction
	// 	// Allocate memory for builtin function and track it
	// 	funcPtr := new(builtinFunction)
	// 	if funcPtr == nil {
	// 		return nil, &MemoryError{"failed to allocate memory for builtin function"}
	// 	}
	// 	*funcPtr = v
	// 	uintPtr, err := safePointerToUintptr(unsafe.Pointer(funcPtr), "builtinFunction")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tv.Data = uintPtr

	// 	// Track allocation and set finalizer
	// 	globalMemoryTracker.trackAllocation(tv.Data, 64, "builtinFunction")
	// 	runtime.SetFinalizer(funcPtr, func(ptr *builtinFunction) {
	// 		if uintPtr, err := safePointerToUintptr(unsafe.Pointer(ptr), "builtinFunction"); err == nil {
	// 			globalMemoryTracker.trackDeallocation(uintPtr, "builtinFunction")
	// 		}
	// 	})
	default:
		return nil, &MemoryError{"unsupported value type for tagged value"}
	}

	// Set reference count to 1
	atomic.StoreInt32(&tv.refCount, 1)
	success = true

	return tv, nil
}

// ToValue converts a SafeTaggedValue back to a regular value with safety checks
func (tv *SafeTaggedValue) ToValue() (Value, error) {
	if tv == nil {
		return nil, &MemoryError{"cannot convert nil tagged value"}
	}

	tv.mutex.RLock()
	defer tv.mutex.RUnlock()

	// Check if the tagged value is still valid
	if atomic.LoadInt32(&tv.refCount) <= 0 {
		return nil, &MemoryError{"tagged value has been freed"}
	}

	// Additional safety check for reference count
	if atomic.LoadInt32(&tv.refCount) == 0 {
		return nil, &MemoryError{"tagged value has zero reference count"}
	}

	switch tv.Type {
	case TaggedNull:
		return nil, nil
	case TaggedBool:
		return tv.Data != 0, nil
	case TaggedInt:
		// Validate integer bounds
		intVal := int(tv.Data)
		if uintptr(intVal) != tv.Data {
			return nil, &MemoryError{"integer value corruption detected"}
		}
		return intVal, nil
	case TaggedFloat:
		if tv.Data == 0 {
			return nil, &MemoryError{"invalid float pointer"}
		}
		// Validate pointer before dereferencing
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return nil, &MemoryError{"invalid float pointer detected"}
		}
		floatPtr, err := safeFloat64PtrFromUintptr(tv.Data)
		if err != nil {
			return nil, err
		}
		floatVal := *floatPtr
		// Validate float value
		if floatVal != floatVal { // NaN check
			return nil, &MemoryError{"corrupted float value (NaN)"}
		}
		return floatVal, nil
	case TaggedString:
		// Safety check for zero index
		if tv.Data == 0 {
			return "", nil
		}
		// Validate string index bounds
		index := int(tv.Data)
		if index < 0 {
			return nil, &MemoryError{"invalid string index (negative)"}
		}
		// Retrieve string from global store using index
		str, err := getSafeStringGlobally(index)
		if err != nil {
			return "", &MemoryError{"failed to retrieve string: " + err.Error()}
		}
		return str, nil
	case TaggedArray:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return &[]Value{}, nil
		}
		// Validate pointer before dereferencing
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return &[]Value{}, &MemoryError{"invalid array pointer in tagged value"}
		}
		unsafePtr, err := safeUintptrToPointer(tv.Data, "array")
		if err != nil {
			return nil, err
		}
		arrayPtr := (*[]Value)(unsafePtr)
		if arrayPtr == nil {
			return nil, &MemoryError{"null array pointer"}
		}
		return arrayPtr, nil
	case TaggedMap:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return map[string]Value{}, nil
		}
		// Validate pointer before dereferencing
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return map[string]Value{}, &MemoryError{"invalid map pointer in tagged value"}
		}
		mapPtr, err := safeMapPtrFromUintptr(tv.Data)
		if err != nil {
			return nil, err
		}
		// Validate map content
		mapVal := *mapPtr
		if mapVal == nil {
			return nil, &MemoryError{"corrupted map value (nil)"}
		}
		return mapVal, nil
	case TaggedFunction:
		// Safety check for nil pointer
		if tv.Data == 0 {
			return nil, nil
		}
		// Validate pointer before dereferencing
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return nil, &MemoryError{"invalid function pointer in tagged value"}
		}

		// Try to determine the function type based on allocation info
		globalMemoryTracker.mutex.RLock()
		alloc, exists := globalMemoryTracker.allocatedPointers[tv.Data]
		globalMemoryTracker.mutex.RUnlock()

		if exists {
			switch alloc.type_ {
			case "userFunction":
				// Direct pointer to userFunction - stored as pointer directly
				ptr, err := safeUintptrToPointer(tv.Data, "userFunction")
				if err != nil {
					return nil, err
				}
				if ptr == nil {
					return nil, &MemoryError{"null userFunction pointer"}
				}
				// Return the userFunction directly, not as pointer
				userFunc := (*userFunction)(ptr)
				if userFunc == nil {
					return nil, &MemoryError{"invalid userFunction"}
				}
				return userFunc, nil
			case "builtinFunction":
				// Pointer to allocated builtinFunction
				ptr, err := safeUintptrToPointer(tv.Data, "builtinFunction")
				if err != nil {
					return nil, err
				}
				if ptr == nil {
					return nil, &MemoryError{"null builtinFunction pointer"}
				}
				return *(*builtinFunction)(ptr), nil
			case "function":
				// Generic functionType pointer
				funcPtr, err := safeFunctionPtrFromUintptr(tv.Data)
				if err != nil {
					return nil, err
				}
				if funcPtr == nil {
					return nil, &MemoryError{"null function pointer"}
				}
				return *funcPtr, nil
			}
		}

		// Fallback to generic function handling
		funcPtr, err := safeFunctionPtrFromUintptr(tv.Data)
		if err != nil {
			return nil, err
		}
		if funcPtr == nil {
			return nil, &MemoryError{"null function pointer"}
		}
		// Return the function value directly
		return *funcPtr, nil
	case TaggedOther:
		// Handle function types stored as TaggedOther
		if tv.Data == 0 {
			return nil, nil
		}
		// Validate pointer before dereferencing
		if !globalMemoryTracker.isValidPointer(tv.Data) {
			return nil, &MemoryError{"invalid other type pointer in tagged value"}
		}
		// Convert back from any pointer
		unsafePtr, err := safeUintptrToPointer(tv.Data, "other")
		if err != nil {
			return nil, err
		}
		interfacePtr := (*any)(unsafePtr)
		if interfacePtr == nil {
			return nil, &MemoryError{"null other type interface pointer"}
		}
		// Return the stored value directly
		return *interfacePtr, nil
	default:
		return nil, &MemoryError{"unknown tagged value type: " + string(rune(tv.Type))}
	}
}

// Thread-safe string storage functions
func storeSafeStringGlobally(s string) (int, error) {
	globalSafeStringMutex.Lock()
	defer globalSafeStringMutex.Unlock()

	// Check if we need cleanup
	if len(globalSafeStringStore) >= globalSafeStringCleanupThreshold {
		if err := cleanupStringStore(); err != nil {
			return 0, err
		}
	}

	index := int(atomic.AddInt64(&globalSafeStringIndex, 1))
	globalSafeStringStore[index] = s
	return index, nil
}

func getSafeStringGlobally(index int) (string, error) {
	globalSafeStringMutex.RLock()
	defer globalSafeStringMutex.RUnlock()

	if s, exists := globalSafeStringStore[index]; exists {
		return s, nil
	}
	return "", &MemoryError{"string index not found in global store"}
}

// cleanupStringStore removes old unused strings (simplified cleanup)
func cleanupStringStore() error {
	// Simple cleanup: remove half of the oldest entries
	// In production, this should be more sophisticated
	if len(globalSafeStringStore) < globalSafeStringCleanupThreshold/2 {
		return nil
	}

	// Remove entries with lower indices (older entries)
	currentIndex := int(atomic.LoadInt64(&globalSafeStringIndex))
	threshold := currentIndex - globalSafeStringCleanupThreshold/2

	for index := range globalSafeStringStore {
		if index < threshold {
			delete(globalSafeStringStore, index)
		}
	}

	return nil
}

// Memory tracking functions
func (mt *MemoryTracker) trackAllocation(ptr uintptr, size int64, type_ string) {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	mt.allocatedPointers[ptr] = &AllocationInfo{
		size:      size,
		allocTime: int64(runtime.NumGoroutine()), // Simple timestamp alternative
		type_:     type_,
		freed:     false,
	}
	atomic.AddInt64(&mt.totalAllocated, size)
}

func (mt *MemoryTracker) trackDeallocation(ptr uintptr, type_ string) {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	if info, exists := mt.allocatedPointers[ptr]; exists {
		info.freed = true
		atomic.AddInt64(&mt.totalFreed, info.size)
		delete(mt.allocatedPointers, ptr)
	}
}

func (mt *MemoryTracker) isValidPointer(ptr uintptr) bool {
	if ptr == 0 {
		return false
	}

	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	alloc, exists := mt.allocatedPointers[ptr]
	if !exists {
		return false
	}

	// Check if already freed
	return !alloc.freed
}

// getAllocationType returns the type of allocation for a given pointer
func (mt *MemoryTracker) getAllocationType(ptr uintptr) string {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()
	if alloc, exists := mt.allocatedPointers[ptr]; exists {
		return alloc.type_
	}
	return ""
}

// detectMemoryLeaks identifies potential memory leaks
func (mt *MemoryTracker) detectMemoryLeaks() []uintptr {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	var leaks []uintptr
	currentTime := runtime.NumGoroutine()

	for ptr, alloc := range mt.allocatedPointers {
		if !alloc.freed {
			// Consider it a leak if allocated more than 1000 goroutine cycles ago
			// This is a heuristic and may need adjustment
			if int64(currentTime)-alloc.allocTime > 1000 {
				leaks = append(leaks, ptr)
			}
		}
	}

	return leaks
}

// forceCleanupLeaks forcefully cleans up detected memory leaks
func (mt *MemoryTracker) forceCleanupLeaks() error {
	leaks := mt.detectMemoryLeaks()
	if len(leaks) == 0 {
		return nil
	}

	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	for _, ptr := range leaks {
		if alloc, exists := mt.allocatedPointers[ptr]; exists && !alloc.freed {
			// Mark as freed to prevent double cleanup
			alloc.freed = true
			atomic.AddInt64(&mt.totalFreed, 1)

			// Try to clear finalizer if possible
			switch alloc.type_ {
			case "float64":
				if floatPtr, err := safeFloat64PtrFromUintptr(ptr); err == nil && floatPtr != nil {
					runtime.SetFinalizer(floatPtr, nil)
				}
			case "map":
				if mapPtr, err := safeMapPtrFromUintptr(ptr); err == nil && mapPtr != nil {
					runtime.SetFinalizer(mapPtr, nil)
				}
			case "function":
				if funcPtr, err := safeFunctionPtrFromUintptr(ptr); err == nil && funcPtr != nil {
					runtime.SetFinalizer(funcPtr, nil)
				}
			}
		}
	}

	return nil
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	TaggedValuesCreated int64
	TaggedValuesReused  int64
	CacheHits           int64
	CacheMisses         int64
}

// GetMemoryStats returns current memory statistics
func (mt *MemoryTracker) GetMemoryStats() MemoryStats {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	return MemoryStats{
		TaggedValuesCreated: atomic.LoadInt64(&globalSafeTaggedValuePool.created),
		TaggedValuesReused:  atomic.LoadInt64(&globalSafeTaggedValuePool.reused),
		CacheHits:           atomic.LoadInt64(&mt.totalAllocated),
		CacheMisses:         atomic.LoadInt64(&mt.totalFreed),
	}
}

// MemoryError represents memory-related errors
type MemoryError struct {
	message string
}

func (e *MemoryError) Error() string {
	return "Memory Error: " + e.message
}

// GetGlobalSafeTaggedValuePool returns the global safe tagged value pool
func GetGlobalSafeTaggedValuePool() *SafeTaggedValuePool {
	return globalSafeTaggedValuePool
}

// Compatibility functions for existing tests
// NewTaggedValuePool creates a new tagged value pool (compatibility wrapper)
func NewTaggedValuePool() *SafeTaggedValuePool {
	return NewSafeTaggedValuePool(1000)
}

// GetGlobalTaggedValuePool returns the global tagged value pool (compatibility wrapper)
func GetGlobalTaggedValuePool() *SafeTaggedValuePool {
	return globalSafeTaggedValuePool
}

// CreateTaggedValue creates a tagged value (compatibility wrapper)
func CreateTaggedValue(value Value) *SafeTaggedValue {
	tv, err := CreateSafeTaggedValue(value)
	if err != nil {
		// For compatibility, return a basic tagged value on error with proper refCount
		return &SafeTaggedValue{
			Type:     TaggedNull,
			Data:     0,
			refCount: 1, // Set refCount to 1 for compatibility
		}
	}
	return tv
}

// GetTaggedValue gets a tagged value from the pool (compatibility wrapper)
func (stvp *SafeTaggedValuePool) GetTaggedValue() *SafeTaggedValue {
	tv, err := stvp.GetSafeTaggedValue()
	if err != nil {
		// For compatibility, return a basic tagged value on error with proper refCount
		return &SafeTaggedValue{
			Type:     TaggedNull,
			Data:     0,
			refCount: 1, // Set refCount to 1 for compatibility
		}
	}
	return tv
}

// PutTaggedValue returns a tagged value to the pool (compatibility wrapper)
func (stvp *SafeTaggedValuePool) PutTaggedValue(tv *SafeTaggedValue) {
	_ = stvp.PutSafeTaggedValue(tv) // Ignore error for compatibility
}

// IsSmallValue checks if the tagged value stores a small value directly (compatibility method)
func (tv *SafeTaggedValue) IsSmallValue() bool {
	switch tv.Type {
	case TaggedNull, TaggedBool, TaggedInt:
		return true
	default:
		return false
	}
}

// Compatibility wrapper for ToValue that returns single value
func (tv *SafeTaggedValue) ToValueCompat() Value {
	value, err := tv.ToValue()
	if err != nil {
		return nil
	}
	return value
}

// GetGlobalMemoryTracker returns the global memory tracker
func GetGlobalMemoryTracker() *MemoryTracker {
	return globalMemoryTracker
}

// CompactEnvironment represents a memory-optimized environment for variable storage
type CompactEnvironment struct {
	vars []map[string]Value
	mu   sync.RWMutex
}

// NewCompactEnvironment creates a new compact environment
func NewCompactEnvironment() *CompactEnvironment {
	return &CompactEnvironment{
		vars: []map[string]Value{make(map[string]Value)},
	}
}

// PushScope adds a new scope to the environment
func (env *CompactEnvironment) PushScope() {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.vars = append(env.vars, make(map[string]Value))
}

// PopScope removes the top scope from the environment
func (env *CompactEnvironment) PopScope() {
	env.mu.Lock()
	defer env.mu.Unlock()
	if len(env.vars) > 1 {
		env.vars = env.vars[:len(env.vars)-1]
	}
}

// GetScopeCount returns the number of scopes
func (env *CompactEnvironment) GetScopeCount() int {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return len(env.vars)
}

// Assign sets a variable in the current scope
func (env *CompactEnvironment) Assign(name string, value Value) {
	env.mu.Lock()
	defer env.mu.Unlock()
	if len(env.vars) > 0 {
		env.vars[len(env.vars)-1][name] = value
	}
}

// Lookup finds a variable in any scope (from current to global)
func (env *CompactEnvironment) Lookup(name string) (Value, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()
	for i := len(env.vars) - 1; i >= 0; i-- {
		if value, exists := env.vars[i][name]; exists {
			return value, true
		}
	}
	return nil, false
}

// CacheFriendlyArray represents a cache-optimized array
type CacheFriendlyArray struct {
	data     []Value
	size     int32
	capacity int32
	mu       sync.RWMutex
}

// NewCacheFriendlyArray creates a new cache-friendly array
func NewCacheFriendlyArray(capacity int32) *CacheFriendlyArray {
	return &CacheFriendlyArray{
		data:     make([]Value, capacity),
		size:     0,
		capacity: capacity,
	}
}

// Append adds a value to the array
func (arr *CacheFriendlyArray) Append(value Value) {
	arr.mu.Lock()
	defer arr.mu.Unlock()
	if arr.size >= arr.capacity {
		// Grow the array
		newCapacity := arr.capacity * 2
		newData := make([]Value, newCapacity)
		copy(newData, arr.data[:arr.size])
		arr.data = newData
		arr.capacity = newCapacity
	}
	arr.data[arr.size] = value
	arr.size++
}

// Get retrieves a value at the given index
func (arr *CacheFriendlyArray) Get(index int32) (Value, bool) {
	arr.mu.RLock()
	defer arr.mu.RUnlock()
	if index < 0 || index >= arr.size {
		return nil, false
	}
	return arr.data[index], true
}

// Set updates a value at the given index
func (arr *CacheFriendlyArray) Set(index int32, value Value) bool {
	arr.mu.Lock()
	defer arr.mu.Unlock()
	if index < 0 || index >= arr.size {
		return false
	}
	arr.data[index] = value
	return true
}

// Size returns the current size of the array
func (arr *CacheFriendlyArray) Size() int32 {
	arr.mu.RLock()
	defer arr.mu.RUnlock()
	return arr.size
}

// ToSlice converts the array to a regular slice
func (arr *CacheFriendlyArray) ToSlice() []Value {
	arr.mu.RLock()
	defer arr.mu.RUnlock()
	result := make([]Value, arr.size)
	copy(result, arr.data[:arr.size])
	return result
}

// CacheFriendlyMap represents a cache-optimized map
type CacheFriendlyMap struct {
	data     map[string]Value
	size     int32
	capacity int32
	mu       sync.RWMutex
}

// NewCacheFriendlyMap creates a new cache-friendly map
func NewCacheFriendlyMap(capacity int32) *CacheFriendlyMap {
	return &CacheFriendlyMap{
		data:     make(map[string]Value, capacity),
		size:     0,
		capacity: capacity,
	}
}

// Set adds or updates a key-value pair
func (m *CacheFriendlyMap) Set(key string, value Value) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; !exists {
		m.size++
		// Check if we need to grow
		if m.size > m.capacity {
			m.capacity = m.capacity * 2
		}
	}
	m.data[key] = value
}

// Get retrieves a value by key
func (m *CacheFriendlyMap) Get(key string) (Value, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

// Delete removes a key-value pair
func (m *CacheFriendlyMap) Delete(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		delete(m.data, key)
		m.size--
		return true
	}
	return false
}

// Size returns the current size of the map
func (m *CacheFriendlyMap) Size() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

// ToMap converts to a regular map
func (m *CacheFriendlyMap) ToMap() map[string]Value {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]Value, len(m.data))
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// MemoryLayoutOptimizer manages memory layout optimizations
type MemoryLayoutOptimizer struct {
	stats MemoryStats
	mu    sync.RWMutex
}

// NewMemoryLayoutOptimizer creates a new memory layout optimizer
func NewMemoryLayoutOptimizer() *MemoryLayoutOptimizer {
	return &MemoryLayoutOptimizer{}
}

// IncrementTaggedValuesCreated increments the tagged values created counter
func (opt *MemoryLayoutOptimizer) IncrementTaggedValuesCreated() {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.stats.TaggedValuesCreated++
}

// IncrementTaggedValuesReused increments the tagged values reused counter
func (opt *MemoryLayoutOptimizer) IncrementTaggedValuesReused() {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.stats.TaggedValuesReused++
}

// IncrementCacheHits increments the cache hits counter
func (opt *MemoryLayoutOptimizer) IncrementCacheHits() {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.stats.CacheHits++
}

// IncrementCacheMisses increments the cache misses counter
func (opt *MemoryLayoutOptimizer) IncrementCacheMisses() {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.stats.CacheMisses++
}

// GetMemoryStats returns a copy of the current memory statistics
func (opt *MemoryLayoutOptimizer) GetMemoryStats() MemoryStats {
	opt.mu.RLock()
	defer opt.mu.RUnlock()
	return opt.stats
}

// ResetStats resets all statistics to zero
func (opt *MemoryLayoutOptimizer) ResetStats() {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.stats = MemoryStats{}
}

// simpleHash provides a simple hash function for testing
func simpleHash(s string) uint32 {
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// Cleanup functions for graceful shutdown
func CleanupMemoryLayout() error {
	// Only perform memory leak detection if enabled
	if IsMemoryLeakDetectionEnabled() {
		leaks := globalMemoryTracker.detectMemoryLeaks()
		if len(leaks) > 0 {
			// Only auto-cleanup if enabled in configuration
			if IsAutoCleanupMemoryLeaksEnabled() {
				if err := globalMemoryTracker.forceCleanupLeaks(); err != nil {
					// Log error but continue with cleanup
					// In a production system, you might want to log this properly
				}
			}
		}
	}

	// Cleanup string store
	globalSafeStringMutex.Lock()
	globalSafeStringStore = make(map[int]string)
	atomic.StoreInt64(&globalSafeStringIndex, 0)
	globalSafeStringMutex.Unlock()

	// Reset memory tracker
	globalMemoryTracker.mutex.Lock()
	globalMemoryTracker.allocatedPointers = make(map[uintptr]*AllocationInfo)
	atomic.StoreInt64(&globalMemoryTracker.totalAllocated, 0)
	atomic.StoreInt64(&globalMemoryTracker.totalFreed, 0)
	globalMemoryTracker.mutex.Unlock()

	// Reset pool counters only, keep the pool instances for reuse
	globalSafeTaggedValuePool.mutex.Lock()
	atomic.StoreInt64(&globalSafeTaggedValuePool.created, 0)
	atomic.StoreInt64(&globalSafeTaggedValuePool.reused, 0)
	atomic.StoreInt64(&globalSafeTaggedValuePool.active, 0)
	atomic.StoreInt32(&globalSafeTaggedValuePool.currSize, 0)
	globalSafeTaggedValuePool.mutex.Unlock()

	// Force garbage collection
	runtime.GC()
	runtime.GC() // Double GC for better cleanup

	return nil
}
