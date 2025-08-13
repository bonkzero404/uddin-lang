package interpreter

import (
	"sync"
	"unsafe"
)

// ArenaAllocator provides memory arena allocation for AST nodes to reduce heap fragmentation
type ArenaAllocator struct {
	mutex     sync.Mutex
	blocks    []*arenaBlock
	current   *arenaBlock
	blockSize int
}

// arenaBlock represents a single memory block in the arena
type arenaBlock struct {
	data   []byte
	offset int
	size   int
}

// NewArenaAllocator creates a new arena allocator
func NewArenaAllocator(blockSize int) *ArenaAllocator {
	if blockSize < 4096 {
		blockSize = 4096 // Minimum block size
	}
	return &ArenaAllocator{
		blockSize: blockSize,
	}
}

// Allocate allocates memory from the arena
func (aa *ArenaAllocator) Allocate(size int) unsafe.Pointer {
	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	// Align size to 8 bytes for better performance
	alignedSize := (size + 7) &^ 7

	// Check if current block has enough space
	if aa.current == nil || aa.current.offset+alignedSize > aa.current.size {
		// Need a new block
		newBlockSize := aa.blockSize
		if alignedSize > newBlockSize {
			newBlockSize = alignedSize
		}

		newBlock := &arenaBlock{
			data:   make([]byte, newBlockSize),
			offset: 0,
			size:   newBlockSize,
		}

		aa.blocks = append(aa.blocks, newBlock)
		aa.current = newBlock
	}

	// Allocate from current block
	ptr := unsafe.Pointer(&aa.current.data[aa.current.offset])
	aa.current.offset += alignedSize

	return ptr
}

// Reset resets the arena, making all allocated memory available for reuse
func (aa *ArenaAllocator) Reset() {
	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	// Reset all blocks
	for _, block := range aa.blocks {
		block.offset = 0
	}

	// Set current to first block if available
	if len(aa.blocks) > 0 {
		aa.current = aa.blocks[0]
	} else {
		aa.current = nil
	}
}

// Size returns the total allocated size
func (aa *ArenaAllocator) Size() int {
	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	totalSize := 0
	for _, block := range aa.blocks {
		totalSize += block.size
	}
	return totalSize
}

// UsedSize returns the total used size
func (aa *ArenaAllocator) UsedSize() int {
	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	usedSize := 0
	for _, block := range aa.blocks {
		usedSize += block.offset
	}
	return usedSize
}

// Global arena allocator for AST nodes
var globalArenaAllocator = NewArenaAllocator(64 * 1024) // 64KB blocks

// GetGlobalArenaAllocator returns the global arena allocator
func GetGlobalArenaAllocator() *ArenaAllocator {
	return globalArenaAllocator
}

// ArenaPool provides pooling of arena allocators for different contexts
type ArenaPool struct {
	pool sync.Pool
}

// NewArenaPool creates a new arena pool
func NewArenaPool(blockSize int) *ArenaPool {
	return &ArenaPool{
		pool: sync.Pool{
			New: func() any {
				return NewArenaAllocator(blockSize)
			},
		},
	}
}

// Get gets an arena allocator from the pool
func (ap *ArenaPool) Get() *ArenaAllocator {
	return ap.pool.Get().(*ArenaAllocator)
}

// Put returns an arena allocator to the pool
func (ap *ArenaPool) Put(aa *ArenaAllocator) {
	aa.Reset()
	ap.pool.Put(aa)
}

// Global arena pool for expression evaluation
var globalArenaPool = NewArenaPool(32 * 1024) // 32KB blocks

// GetArenaFromPool gets an arena allocator from the global pool
func GetArenaFromPool() *ArenaAllocator {
	return globalArenaPool.Get()
}

// PutArenaToPool returns an arena allocator to the global pool
func PutArenaToPool(aa *ArenaAllocator) {
	globalArenaPool.Put(aa)
}

// ArenaString creates a string using arena allocation
type ArenaString struct {
	data unsafe.Pointer
	len  int
}

// NewArenaString creates a new arena-allocated string
func NewArenaString(s string, arena *ArenaAllocator) *ArenaString {
	if len(s) == 0 {
		return &ArenaString{data: nil, len: 0}
	}

	ptr := arena.Allocate(len(s))
	copy((*[1 << 30]byte)(ptr)[:len(s)], []byte(s))

	return &ArenaString{
		data: ptr,
		len:  len(s),
	}
}

// String returns the string value
func (as *ArenaString) String() string {
	if as.len == 0 {
		return ""
	}
	return string((*[1 << 30]byte)(as.data)[:as.len])
}

// ArenaSlice creates a slice using arena allocation
type ArenaSlice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// NewArenaSlice creates a new arena-allocated slice
func NewArenaSlice(capacity int, arena *ArenaAllocator) *ArenaSlice {
	if capacity == 0 {
		return &ArenaSlice{data: nil, len: 0, cap: 0}
	}

	size := capacity * int(unsafe.Sizeof(Value(nil)))
	ptr := arena.Allocate(size)

	return &ArenaSlice{
		data: ptr,
		len:  0,
		cap:  capacity,
	}
}

// Append appends a value to the arena slice
func (as *ArenaSlice) Append(value Value) {
	if as.len >= as.cap {
		panic("arena slice capacity exceeded")
	}

	slice := (*[1 << 30]Value)(as.data)[:as.cap]
	slice[as.len] = value
	as.len++
}

// Slice returns the slice as []Value
func (as *ArenaSlice) Slice() []Value {
	if as.len == 0 {
		return nil
	}
	return (*[1 << 30]Value)(as.data)[:as.len]
}

// ExpressionArena provides arena allocation specifically for expression evaluation
type ExpressionArena struct {
	arena      *ArenaAllocator
	strings    []*ArenaString
	slices     []*ArenaSlice
	tempValues []Value
}

// NewExpressionArena creates a new expression arena
func NewExpressionArena() *ExpressionArena {
	return &ExpressionArena{
		arena: GetArenaFromPool(),
	}
}

// AllocateString allocates a string in the arena
func (ea *ExpressionArena) AllocateString(s string) string {
	as := NewArenaString(s, ea.arena)
	ea.strings = append(ea.strings, as)
	return as.String()
}

// AllocateSlice allocates a slice in the arena
func (ea *ExpressionArena) AllocateSlice(capacity int) []Value {
	as := NewArenaSlice(capacity, ea.arena)
	ea.slices = append(ea.slices, as)
	return as.Slice()
}

// AddTempValue adds a temporary value for cleanup
func (ea *ExpressionArena) AddTempValue(value Value) {
	ea.tempValues = append(ea.tempValues, value)
}

// Reset resets the expression arena
func (ea *ExpressionArena) Reset() {
	ea.strings = ea.strings[:0]
	ea.slices = ea.slices[:0]
	ea.tempValues = ea.tempValues[:0]
	ea.arena.Reset()
}

// Release releases the expression arena back to the pool
func (ea *ExpressionArena) Release() {
	ea.Reset()
	PutArenaToPool(ea.arena)
}

// ExpressionArenaPool provides pooling for expression arenas
type ExpressionArenaPool struct {
	pool sync.Pool
}

// NewExpressionArenaPool creates a new expression arena pool
func NewExpressionArenaPool() *ExpressionArenaPool {
	return &ExpressionArenaPool{
		pool: sync.Pool{
			New: func() any {
				return NewExpressionArena()
			},
		},
	}
}

// Get gets an expression arena from the pool
func (eap *ExpressionArenaPool) Get() *ExpressionArena {
	return eap.pool.Get().(*ExpressionArena)
}

// Put returns an expression arena to the pool
func (eap *ExpressionArenaPool) Put(ea *ExpressionArena) {
	ea.Reset()
	eap.pool.Put(ea)
}

// Global expression arena pool
var globalExpressionArenaPool = NewExpressionArenaPool()

// GetExpressionArena gets an expression arena from the global pool
func GetExpressionArena() *ExpressionArena {
	return globalExpressionArenaPool.Get()
}

// PutExpressionArena returns an expression arena to the global pool
func PutExpressionArena(ea *ExpressionArena) {
	globalExpressionArenaPool.Put(ea)
}

// WithArena executes a function with an expression arena
func WithArena(fn func(*ExpressionArena) Value) Value {
	ea := GetExpressionArena()
	defer PutExpressionArena(ea)
	return fn(ea)
}

// ArenaStats provides statistics about arena usage
type ArenaStats struct {
	TotalAllocated int
	TotalUsed      int
	BlockCount     int
	Utilization    float64
}

// GetArenaStats returns statistics about the global arena allocator
func GetArenaStats() ArenaStats {
	aa := GetGlobalArenaAllocator()
	totalAllocated := aa.Size()
	totalUsed := aa.UsedSize()
	blockCount := len(aa.blocks)

	utilization := 0.0
	if totalAllocated > 0 {
		utilization = float64(totalUsed) / float64(totalAllocated)
	}

	return ArenaStats{
		TotalAllocated: totalAllocated,
		TotalUsed:      totalUsed,
		BlockCount:     blockCount,
		Utilization:    utilization,
	}
}
