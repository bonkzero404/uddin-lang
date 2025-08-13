package interpreter

import (
	"container/list"
	"strings"
	"sync"
)

// EnhancedStringInterner provides advanced string interning with LRU eviction
type EnhancedStringInterner struct {
	mutex     sync.RWMutex
	cache     map[string]*list.Element
	lruList   *list.List
	maxSize   int
	current   int
	hitCount  int64
	missCount int64
}

// lruEntry represents an entry in the LRU cache
type lruEntry struct {
	key   string
	value string
}

// NewEnhancedStringInterner creates a new enhanced string interner
func NewEnhancedStringInterner(maxSize int) *EnhancedStringInterner {
	return &EnhancedStringInterner{
		cache:   make(map[string]*list.Element, maxSize),
		lruList: list.New(),
		maxSize: maxSize,
	}
}

// Intern interns a string, returning the canonical instance
func (esi *EnhancedStringInterner) Intern(s string) string {
	esi.mutex.Lock()
	defer esi.mutex.Unlock()

	// Check if string is already interned
	if elem, exists := esi.cache[s]; exists {
		// Move to front (most recently used)
		esi.lruList.MoveToFront(elem)
		esi.hitCount++
		return elem.Value.(*lruEntry).value
	}

	esi.missCount++

	// Check if we need to evict
	if esi.current >= esi.maxSize {
		// Remove least recently used
		oldest := esi.lruList.Back()
		if oldest != nil {
			oldEntry := oldest.Value.(*lruEntry)
			delete(esi.cache, oldEntry.key)
			esi.lruList.Remove(oldest)
			esi.current--
		}
	}

	// Add new entry
	entry := &lruEntry{key: s, value: s}
	elem := esi.lruList.PushFront(entry)
	esi.cache[s] = elem
	esi.current++

	return s
}

// Contains checks if a string is already interned
func (esi *EnhancedStringInterner) Contains(s string) bool {
	esi.mutex.RLock()
	defer esi.mutex.RUnlock()

	_, exists := esi.cache[s]
	return exists
}

// Size returns the current cache size
func (esi *EnhancedStringInterner) Size() int {
	esi.mutex.RLock()
	defer esi.mutex.RUnlock()
	return esi.current
}

// Stats returns cache statistics
func (esi *EnhancedStringInterner) Stats() (hits, misses int64, hitRatio float64) {
	esi.mutex.RLock()
	defer esi.mutex.RUnlock()

	total := esi.hitCount + esi.missCount
	if total > 0 {
		hitRatio = float64(esi.hitCount) / float64(total)
	}

	return esi.hitCount, esi.missCount, hitRatio
}

// Clear clears the cache
func (esi *EnhancedStringInterner) Clear() {
	esi.mutex.Lock()
	defer esi.mutex.Unlock()

	esi.cache = make(map[string]*list.Element, esi.maxSize)
	esi.lruList = list.New()
	esi.current = 0
	esi.hitCount = 0
	esi.missCount = 0
}

// Global enhanced string interner
var globalEnhancedStringInterner = NewEnhancedStringInterner(10000)

// InternStringEnhanced interns a string using the global enhanced interner
func InternStringEnhanced(s string) string {
	// Only intern strings that are likely to be repeated
	if len(s) < 2 || len(s) > 256 {
		return s
	}
	return globalEnhancedStringInterner.Intern(s)
}

// GetStringInternerStats returns statistics from the global string interner
func GetStringInternerStats() (hits, misses int64, hitRatio float64) {
	return globalEnhancedStringInterner.Stats()
}

// SmartStringInterner provides context-aware string interning
type SmartStringInterner struct {
	identifiers *EnhancedStringInterner
	literals    *EnhancedStringInterner
	operators   *EnhancedStringInterner
	general     *EnhancedStringInterner
}

// NewSmartStringInterner creates a new smart string interner
func NewSmartStringInterner() *SmartStringInterner {
	return &SmartStringInterner{
		identifiers: NewEnhancedStringInterner(5000), // Variable names, function names
		literals:    NewEnhancedStringInterner(3000), // String literals
		operators:   NewEnhancedStringInterner(100),  // Operators, keywords
		general:     NewEnhancedStringInterner(2000), // Other strings
	}
}

// StringType represents the type of string for smart interning
type StringType int

const (
	StringTypeIdentifier StringType = iota
	StringTypeLiteral
	StringTypeOperator
	StringTypeGeneral
)

// InternByType interns a string based on its type
func (ssi *SmartStringInterner) InternByType(s string, stringType StringType) string {
	switch stringType {
	case StringTypeIdentifier:
		return ssi.identifiers.Intern(s)
	case StringTypeLiteral:
		return ssi.literals.Intern(s)
	case StringTypeOperator:
		return ssi.operators.Intern(s)
	default:
		return ssi.general.Intern(s)
	}
}

// Stats returns combined statistics
func (ssi *SmartStringInterner) Stats() map[string]interface{} {
	idHits, idMisses, idRatio := ssi.identifiers.Stats()
	litHits, litMisses, litRatio := ssi.literals.Stats()
	opHits, opMisses, opRatio := ssi.operators.Stats()
	genHits, genMisses, genRatio := ssi.general.Stats()

	return map[string]interface{}{
		"identifiers": map[string]interface{}{
			"hits":      idHits,
			"misses":    idMisses,
			"hit_ratio": idRatio,
			"size":      ssi.identifiers.Size(),
		},
		"literals": map[string]interface{}{
			"hits":      litHits,
			"misses":    litMisses,
			"hit_ratio": litRatio,
			"size":      ssi.literals.Size(),
		},
		"operators": map[string]interface{}{
			"hits":      opHits,
			"misses":    opMisses,
			"hit_ratio": opRatio,
			"size":      ssi.operators.Size(),
		},
		"general": map[string]interface{}{
			"hits":      genHits,
			"misses":    genMisses,
			"hit_ratio": genRatio,
			"size":      ssi.general.Size(),
		},
	}
}

// Global smart string interner
var globalSmartStringInterner = NewSmartStringInterner()

// InternIdentifier interns an identifier string
func InternIdentifier(s string) string {
	return globalSmartStringInterner.InternByType(s, StringTypeIdentifier)
}

// InternLiteral interns a literal string
func InternLiteral(s string) string {
	return globalSmartStringInterner.InternByType(s, StringTypeLiteral)
}

// InternOperator interns an operator string
func InternOperator(s string) string {
	return globalSmartStringInterner.InternByType(s, StringTypeOperator)
}

// GetSmartInternerStats returns statistics from the global smart interner
func GetSmartInternerStats() map[string]interface{} {
	return globalSmartStringInterner.Stats()
}

// StringPool provides pooling for string builders and byte slices
type StringPool struct {
	builderPool sync.Pool
	bytePool    sync.Pool
}

// NewStringPool creates a new string pool
func NewStringPool() *StringPool {
	return &StringPool{
		builderPool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
		bytePool: sync.Pool{
			New: func() any {
				return make([]byte, 0, 256)
			},
		},
	}
}

// GetBuilder gets a string builder from the pool
func (sp *StringPool) GetBuilder() *strings.Builder {
	return sp.builderPool.Get().(*strings.Builder)
}

// PutBuilder returns a string builder to the pool
func (sp *StringPool) PutBuilder(sb *strings.Builder) {
	sb.Reset()
	sp.builderPool.Put(sb)
}

// GetBytes gets a byte slice from the pool
func (sp *StringPool) GetBytes() []byte {
	return sp.bytePool.Get().([]byte)[:0]
}

// PutBytes returns a byte slice to the pool
func (sp *StringPool) PutBytes(b []byte) {
	if cap(b) <= 1024 { // Only pool reasonably sized slices
		sp.bytePool.Put(b[:0])
	}
}

// Global string pool
var globalStringPool = NewStringPool()

// GetStringBuilder gets a string builder from the global pool
func GetStringBuilder() *strings.Builder {
	return globalStringPool.GetBuilder()
}

// PutStringBuilder returns a string builder to the global pool
func PutStringBuilder(sb *strings.Builder) {
	globalStringPool.PutBuilder(sb)
}

// GetByteSlice gets a byte slice from the global pool
func GetByteSlice() []byte {
	return globalStringPool.GetBytes()
}

// PutByteSlice returns a byte slice to the global pool
func PutByteSlice(b []byte) {
	globalStringPool.PutBytes(b)
}

// OptimizedStringConcat performs optimized string concatenation
func OptimizedStringConcat(strings ...string) string {
	if len(strings) == 0 {
		return ""
	}
	if len(strings) == 1 {
		return strings[0]
	}

	// Calculate total length
	totalLen := 0
	for _, s := range strings {
		totalLen += len(s)
	}

	// Use pooled byte slice for small concatenations
	if totalLen <= 1024 {
		buf := GetByteSlice()
		defer PutByteSlice(buf)

		if cap(buf) < totalLen {
			buf = make([]byte, 0, totalLen)
		}

		for _, s := range strings {
			buf = append(buf, s...)
		}

		return InternStringEnhanced(string(buf))
	}

	// Use string builder for larger concatenations
	sb := GetStringBuilder()
	defer PutStringBuilder(sb)

	sb.Grow(totalLen)
	for _, s := range strings {
		sb.WriteString(s)
	}

	return InternStringEnhanced(sb.String())
}

// FastStringEqual performs fast string equality check with interning
func FastStringEqual(a, b string) bool {
	// If both strings are interned and have the same pointer, they're equal
	if len(a) == len(b) {
		// For small strings, just compare directly
		if len(a) <= 16 {
			return a == b
		}

		// For larger strings, try interning first
		if globalEnhancedStringInterner.Contains(a) && globalEnhancedStringInterner.Contains(b) {
			internedA := globalEnhancedStringInterner.Intern(a)
			internedB := globalEnhancedStringInterner.Intern(b)
			// Pointer comparison for interned strings
			return &internedA == &internedB
		}
	}

	return a == b
}
