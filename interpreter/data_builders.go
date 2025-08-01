package interpreter

import (
	"strings"
	"sync"
)

// ========================================
// OPTIMIZATION 1: Loop-based Array Concatenation Patterns
// ========================================

// ArrayConcatenator provides optimized array concatenation
type ArrayConcatenator struct {
	result []Value
	cap    int
}

// NewArrayConcatenator creates a new optimized array concatenator
func NewArrayConcatenator(estimatedSize int) *ArrayConcatenator {
	return &ArrayConcatenator{
		result: make([]Value, 0, estimatedSize),
		cap:    estimatedSize,
	}
}

// Append adds values efficiently without repeated allocations
func (ac *ArrayConcatenator) Append(values ...Value) {
	ac.result = SmartAppend(ac.result, values...)
}

// AppendArray adds an entire array efficiently
func (ac *ArrayConcatenator) AppendArray(arr *[]Value) {
	ac.result = SmartAppend(ac.result, *arr...)
}

// Result returns the final concatenated array
func (ac *ArrayConcatenator) Result() *[]Value {
	return &ac.result
}

// Reset clears the concatenator for reuse
func (ac *ArrayConcatenator) Reset() {
	ac.result = ac.result[:0]
}

// ========================================
// OPTIMIZATION 2: String Concatenation in Loops
// ========================================

// StringConcatenator provides optimized string concatenation using sync.Pool
type StringConcatenator struct {
	builder *strings.Builder
	pool    *sync.Pool
}

// Global string builder pool for reuse
var stringBuilderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

// NewStringConcatenator creates a new optimized string concatenator
func NewStringConcatenator() *StringConcatenator {
	builder := stringBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	return &StringConcatenator{
		builder: builder,
		pool:    &stringBuilderPool,
	}
}

// Append adds a string efficiently
func (sc *StringConcatenator) Append(s string) {
	sc.builder.WriteString(s)
}

// AppendRune adds a rune efficiently
func (sc *StringConcatenator) AppendRune(r rune) {
	sc.builder.WriteRune(r)
}

// AppendByte adds a byte efficiently
func (sc *StringConcatenator) AppendByte(b byte) {
	sc.builder.WriteByte(b)
}

// Result returns the final concatenated string and releases the builder
func (sc *StringConcatenator) Result() string {
	result := sc.builder.String()
	sc.pool.Put(sc.builder)
	sc.builder = nil
	return result
}

// ========================================
// OPTIMIZATION 3: Map Creation in Loops
// ========================================

// MapBuilder provides optimized map creation with pre-allocation
type MapBuilder struct {
	data map[string]Value
	pool *sync.Pool
}

// Global map pool for reuse
var mapPool = sync.Pool{
	New: func() any {
		return make(map[string]Value)
	},
}

// NewMapBuilder creates a new optimized map builder
func NewMapBuilder(estimatedSize int) *MapBuilder {
	m := mapPool.Get().(map[string]Value)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return &MapBuilder{
		data: m,
		pool: &mapPool,
	}
}

// Set adds a key-value pair efficiently
func (mb *MapBuilder) Set(key string, value Value) {
	mb.data[key] = value
}

// SetAll adds multiple key-value pairs from another map
func (mb *MapBuilder) SetAll(source map[string]Value) {
	for k, v := range source {
		mb.data[k] = v
	}
}

// Result returns the final map and releases it to the pool when done
func (mb *MapBuilder) Result() map[string]Value {
	result := mb.data
	mb.data = nil
	return result
}

// Release returns the map to the pool for reuse
func (mb *MapBuilder) Release() {
	if mb.data != nil {
		mb.pool.Put(mb.data)
		mb.data = nil
	}
}

// ========================================
// OPTIMIZATION 4: Function Call Stack Pooling
// ========================================

// CallFrame represents a function call frame
type CallFrame struct {
	FunctionName string
	Args         []Value
	Scope        map[string]Value
	ReturnValue  Value
}

// CallStack provides optimized function call stack management
type CallStack struct {
	frames []CallFrame
	pool   *sync.Pool
}

// Global call frame pool
var callFramePool = sync.Pool{
	New: func() any {
		return &CallFrame{
			Args:  make([]Value, 0, 8), // Pre-allocate for 8 args
			Scope: make(map[string]Value),
		}
	},
}

// Global call stack pool
var callStackPool = sync.Pool{
	New: func() any {
		return &CallStack{
			frames: make([]CallFrame, 0, 32), // Pre-allocate for 32 frames
			pool:   &callFramePool,
		}
	},
}

// NewCallStack creates a new optimized call stack
func NewCallStack() *CallStack {
	cs := callStackPool.Get().(*CallStack)
	cs.frames = cs.frames[:0] // Reset length but keep capacity
	return cs
}

// Push adds a new call frame to the stack
func (cs *CallStack) Push(functionName string, args []Value, scope map[string]Value) {
	frame := cs.pool.Get().(*CallFrame)
	frame.FunctionName = functionName
	frame.Args = frame.Args[:0] // Reset length
	frame.Args = SmartAppend(frame.Args, args...)
	frame.Scope = scope
	frame.ReturnValue = nil

	cs.frames = SmartAppendCallFrame(cs.frames, *frame)
}

// Pop removes and returns the top call frame
func (cs *CallStack) Pop() *CallFrame {
	if len(cs.frames) == 0 {
		return nil
	}

	frame := &cs.frames[len(cs.frames)-1]
	cs.frames = cs.frames[:len(cs.frames)-1]
	return frame
}

// Peek returns the top call frame without removing it
func (cs *CallStack) Peek() *CallFrame {
	if len(cs.frames) == 0 {
		return nil
	}
	return &cs.frames[len(cs.frames)-1]
}

// Depth returns the current call stack depth
func (cs *CallStack) Depth() int {
	return len(cs.frames)
}

// Release returns the call stack to the pool
func (cs *CallStack) Release() {
	// Return all frames to the pool
	for i := range cs.frames {
		cs.pool.Put(&cs.frames[i])
	}
	cs.frames = cs.frames[:0]
	callStackPool.Put(cs)
}

// ========================================
// OPTIMIZATION UTILITIES
// ========================================

// OptimizedJoin provides efficient string joining for arrays
func OptimizedJoin(arr []Value, separator string) string {
	if len(arr) == 0 {
		return ""
	}

	if len(arr) == 1 {
		return toString(arr[0], false)
	}

	concat := NewStringConcatenator()
	defer func() {
		if concat.builder != nil {
			concat.pool.Put(concat.builder)
		}
	}()

	for i, v := range arr {
		if i > 0 {
			concat.Append(separator)
		}
		concat.Append(toString(v, false))
	}

	return concat.Result()
}

// OptimizedArrayCopy creates an optimized copy of an array
func OptimizedArrayCopy(source []Value) []Value {
	if len(source) == 0 {
		return make([]Value, 0)
	}

	result := make([]Value, len(source))
	copy(result, source)
	return result
}

// OptimizedMapCopy creates an optimized copy of a map
func OptimizedMapCopy(source map[string]Value) map[string]Value {
	if len(source) == 0 {
		return make(map[string]Value)
	}

	result := make(map[string]Value, len(source))
	for k, v := range source {
		result[k] = v
	}
	return result
}

// BatchArrayAppend efficiently appends multiple arrays
func BatchArrayAppend(target *[]Value, sources ...*[]Value) {
	totalLen := len(*target)
	for _, src := range sources {
		totalLen += len(*src)
	}

	// Pre-allocate if needed
	if cap(*target) < totalLen {
		newTarget := make([]Value, len(*target), totalLen)
		copy(newTarget, *target)
		*target = newTarget
	}

	// Append all sources
	for _, src := range sources {
		*target = SmartAppend(*target, *src...)
	}
}

// BatchMapMerge efficiently merges multiple maps
func BatchMapMerge(target map[string]Value, sources ...map[string]Value) {
	for _, src := range sources {
		for k, v := range src {
			target[k] = v
		}
	}
}

// ========================================
// POOL ACCESS FUNCTIONS
// ========================================

// GetArrayConcatenator gets an array concatenator from the pool
func GetArrayConcatenator() *ArrayConcatenator {
	return NewArrayConcatenator(16) // Default capacity
}

// PutArrayConcatenator returns an array concatenator to the pool
func PutArrayConcatenator(ac *ArrayConcatenator) {
	ac.Reset()
}

// ToArray returns the result as an array (alias for Result)
func (ac *ArrayConcatenator) ToArray() []Value {
	return ac.result
}

// GetStringConcatenator gets a string concatenator from the pool
func GetStringConcatenator() *StringConcatenator {
	return NewStringConcatenator()
}

// PutStringConcatenator returns a string concatenator to the pool
func PutStringConcatenator(sc *StringConcatenator) {
	if sc.builder != nil {
		sc.pool.Put(sc.builder)
		sc.builder = nil
	}
}

// WriteString writes a string to the concatenator
func (sc *StringConcatenator) WriteString(s string) {
	sc.builder.WriteString(s)
}

// String returns the concatenated string
func (sc *StringConcatenator) String() string {
	return sc.builder.String()
}

// GetMapBuilder gets a map builder from the pool
func GetMapBuilder() *MapBuilder {
	return NewMapBuilder(16) // Default capacity
}

// PutMapBuilder returns a map builder to the pool
func PutMapBuilder(mb *MapBuilder) {
	mb.Release()
}

// ToMap returns the result as a map (alias for Result)
func (mb *MapBuilder) ToMap() map[string]Value {
	return mb.Result()
}

// GetCallStack gets a call stack from the pool
func GetCallStack() *CallStack {
	return NewCallStack()
}

// PutCallStack returns a call stack to the pool
func PutCallStack(cs *CallStack) {
	cs.Release()
}

// PushCall adds a call info to the stack (compatibility version)
func (cs *CallStack) PushCall(callInfo map[string]any) {
	// For compatibility with existing code
	if functionName, ok := callInfo["function"].(string); ok {
		cs.frames = append(cs.frames, CallFrame{
			FunctionName: functionName,
		})
	}
}

// PopCall removes the top call from the stack (compatibility version)
func (cs *CallStack) PopCall() {
	if len(cs.frames) > 0 {
		cs.frames = cs.frames[:len(cs.frames)-1]
	}
}
