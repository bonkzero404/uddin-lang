package interpreter

import (
	"reflect"
	"sync"
)

// VariableLookupCache provides optimized variable lookup with caching
// EXPERIMENTAL: This cache is part of experimental memory optimization features
// WARNING: May not be thread-safe and could have compatibility issues
type VariableLookupCache struct {
	cache   map[string]*CachedVariable
	mutex   sync.RWMutex
	maxSize int
	hits    int64
	misses  int64
}

// CachedVariable represents a cached variable with its scope level
type CachedVariable struct {
	Value      Value
	ScopeLevel int
	LastAccess int64
}

// NewVariableLookupCache creates a new variable lookup cache
func NewVariableLookupCache(maxSize int) *VariableLookupCache {
	return &VariableLookupCache{
		cache:   make(map[string]*CachedVariable),
		maxSize: maxSize,
	}
}

// Global variable lookup cache
// EXPERIMENTAL: Global instance for experimental variable lookup caching
var globalVariableLookupCache = NewVariableLookupCache(500)

// safeValueEqual safely compares two values, handling uncomparable types
func safeValueEqual(a, b Value) bool {
	defer func() {
		if recover() != nil {
			// If comparison panics, consider them different
		}
	}()

	// Use reflect.DeepEqual for safe comparison
	return reflect.DeepEqual(a, b)
}

// FastLookup performs optimized variable lookup with caching
func (vlc *VariableLookupCache) FastLookup(name string, env *Environment) (Value, bool) {
	vlc.mutex.RLock()
	cached, exists := vlc.cache[name]
	vlc.mutex.RUnlock()

	if exists {
		// Verify the cached value is still valid for the current scope
		if cached.ScopeLevel < len(env.vars) {
			if value, found := env.vars[cached.ScopeLevel][name]; found && safeValueEqual(value, cached.Value) {
				vlc.hits++
				return cached.Value, true
			}
		}
		// Cache is stale, remove it
		vlc.InvalidateVariable(name)
	}

	vlc.misses++

	// Perform regular lookup and cache the result
	for i := len(env.vars) - 1; i >= 0; i-- {
		if value, found := env.vars[i][name]; found {
			vlc.CacheVariable(name, value, i)
			return value, true
		}
	}

	return nil, false
}

// CacheVariable stores a variable in the cache
func (vlc *VariableLookupCache) CacheVariable(name string, value Value, scopeLevel int) {
	vlc.mutex.Lock()
	defer vlc.mutex.Unlock()

	// Check cache size limit
	if len(vlc.cache) >= vlc.maxSize {
		vlc.evictOldEntries()
	}

	vlc.cache[name] = &CachedVariable{
		Value:      value,
		ScopeLevel: scopeLevel,
		LastAccess: vlc.hits + vlc.misses,
	}
}

// InvalidateVariable removes a variable from the cache
func (vlc *VariableLookupCache) InvalidateVariable(name string) {
	vlc.mutex.Lock()
	defer vlc.mutex.Unlock()
	delete(vlc.cache, name)
}

// InvalidateScope removes all variables from a specific scope level
func (vlc *VariableLookupCache) InvalidateScope(scopeLevel int) {
	vlc.mutex.Lock()
	defer vlc.mutex.Unlock()

	for name, cached := range vlc.cache {
		if cached.ScopeLevel >= scopeLevel {
			delete(vlc.cache, name)
		}
	}
}

// evictOldEntries removes least recently used entries
func (vlc *VariableLookupCache) evictOldEntries() {
	// Simple LRU eviction: remove half of the cache
	toRemove := len(vlc.cache) / 2
	oldestAccess := vlc.hits + vlc.misses
	oldestNames := make([]string, 0, toRemove)

	// Find oldest entries
	for name, cached := range vlc.cache {
		if len(oldestNames) < toRemove {
			oldestNames = append(oldestNames, name)
			if cached.LastAccess < oldestAccess {
				oldestAccess = cached.LastAccess
			}
		} else if cached.LastAccess < oldestAccess {
			// Replace one of the oldest
			for i, oldName := range oldestNames {
				if vlc.cache[oldName].LastAccess > cached.LastAccess {
					oldestNames[i] = name
					break
				}
			}
		}
	}

	// Remove oldest entries
	for _, name := range oldestNames {
		delete(vlc.cache, name)
	}
}

// GetCacheStats returns cache performance statistics
func (vlc *VariableLookupCache) GetCacheStats() (hits, misses int64, hitRatio float64) {
	vlc.mutex.RLock()
	defer vlc.mutex.RUnlock()

	total := vlc.hits + vlc.misses
	if total > 0 {
		hitRatio = float64(vlc.hits) / float64(total)
	}

	return vlc.hits, vlc.misses, hitRatio
}

// ClearCache clears all cached variables
func (vlc *VariableLookupCache) ClearCache() {
	vlc.mutex.Lock()
	defer vlc.mutex.Unlock()
	vlc.cache = make(map[string]*CachedVariable)
	vlc.hits = 0
	vlc.misses = 0
}

// GetGlobalVariableLookupCache returns the global variable lookup cache
func GetGlobalVariableLookupCache() *VariableLookupCache {
	return globalVariableLookupCache
}

// GlobalVariableCache provides fast access to frequently used global variables
// EXPERIMENTAL: This cache is experimental and part of memory optimization
type GlobalVariableCache struct {
	cache map[string]Value
	mutex sync.RWMutex
}

// NewGlobalVariableCache creates a new global variable cache
func NewGlobalVariableCache() *GlobalVariableCache {
	return &GlobalVariableCache{
		cache: make(map[string]Value),
	}
}

// Global instance
// EXPERIMENTAL: Global cache instance for experimental variable optimization
var globalVarCache = NewGlobalVariableCache()

// GetGlobalVariable retrieves a global variable from cache
func (gvc *GlobalVariableCache) GetGlobalVariable(name string) (Value, bool) {
	gvc.mutex.RLock()
	defer gvc.mutex.RUnlock()
	value, exists := gvc.cache[name]
	return value, exists
}

// SetGlobalVariable caches a global variable
func (gvc *GlobalVariableCache) SetGlobalVariable(name string, value Value) {
	gvc.mutex.Lock()
	defer gvc.mutex.Unlock()
	gvc.cache[name] = value
}

// InvalidateGlobalVariable removes a global variable from cache
func (gvc *GlobalVariableCache) InvalidateGlobalVariable(name string) {
	gvc.mutex.Lock()
	defer gvc.mutex.Unlock()
	delete(gvc.cache, name)
}

// GetGlobalVariableCache returns the global variable cache instance
func GetGlobalVariableCache() *GlobalVariableCache {
	return globalVarCache
}

// ScopeFlattener optimizes deep scope lookups by flattening variable access
// EXPERIMENTAL: This flattener is experimental and part of memory optimization
type ScopeFlattener struct {
	flattenedVars map[string]Value
	mutex         sync.RWMutex
	maxDepth      int
}

// NewScopeFlattener creates a new scope flattener
func NewScopeFlattener(maxDepth int) *ScopeFlattener {
	return &ScopeFlattener{
		flattenedVars: make(map[string]Value),
		maxDepth:      maxDepth,
	}
}

// FlattenScope creates a flattened view of variables for deep scopes
func (sf *ScopeFlattener) FlattenScope(env *Environment) map[string]Value {
	if len(env.vars) <= sf.maxDepth {
		return nil // No need to flatten
	}

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	// Clear previous flattened vars
	for k := range sf.flattenedVars {
		delete(sf.flattenedVars, k)
	}

	// Flatten variables from all scopes
	for i := 0; i < len(env.vars); i++ {
		for name, value := range env.vars[i] {
			if _, exists := sf.flattenedVars[name]; !exists {
				sf.flattenedVars[name] = value
			}
		}
	}

	return sf.flattenedVars
}

// LookupFlattened performs lookup in flattened scope
func (sf *ScopeFlattener) LookupFlattened(name string) (Value, bool) {
	sf.mutex.RLock()
	defer sf.mutex.RUnlock()
	value, exists := sf.flattenedVars[name]
	return value, exists
}
