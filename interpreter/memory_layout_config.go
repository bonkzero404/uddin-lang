package interpreter

// MemoryLayoutConfig controls the memory layout optimization features
type MemoryLayoutConfig struct {
	// EnableTaggedValues enables the use of TaggedValue instead of Value
	EnableTaggedValues bool
	// EnableCompactEnvironment enables the use of CompactEnvironment instead of regular Environment
	EnableCompactEnvironment bool
	// EnableCacheFriendlyStructures enables the use of CacheFriendlyArray and CacheFriendlyMap
	EnableCacheFriendlyStructures bool
	// EnableVariableLookupCache enables variable lookup caching for performance
	EnableVariableLookupCache bool
	// EnableMemoryLeakDetection enables automatic memory leak detection
	EnableMemoryLeakDetection bool
	// MemoryLeakDetectionInterval sets the interval (in seconds) for periodic memory leak detection
	MemoryLeakDetectionInterval int
	// AutoCleanupMemoryLeaks enables automatic cleanup of detected memory leaks
	AutoCleanupMemoryLeaks bool
	// TaggedValuePoolSize sets the initial size of the TaggedValue pool
	TaggedValuePoolSize int
	// MaxStringCacheSize sets the maximum number of strings to cache globally
	MaxStringCacheSize int
	// VariableLookupCacheSize sets the maximum size of the variable lookup cache
	VariableLookupCacheSize int
}

// DefaultMemoryLayoutConfig returns the default memory layout configuration
// By default, memory layout optimizations are disabled for backward compatibility
func DefaultMemoryLayoutConfig() *MemoryLayoutConfig {
	return &MemoryLayoutConfig{
		EnableTaggedValues:            false,
		EnableCompactEnvironment:      false,
		EnableCacheFriendlyStructures: false,
		EnableVariableLookupCache:     false,
		EnableMemoryLeakDetection:     false,
		MemoryLeakDetectionInterval:   60, // 60 seconds
		AutoCleanupMemoryLeaks:        false,
		TaggedValuePoolSize:           1000,
		MaxStringCacheSize:            10000,
		VariableLookupCacheSize:       500,
	}
}

// StableMemoryLayoutConfig returns production-ready memory layout configuration
// TaggedValues are disabled (161ns overhead per assignment outweighs type-tag benefits).
// CompactEnvironment and CacheFriendlyStructures are enabled — both are thread-safe.
func StableMemoryLayoutConfig() *MemoryLayoutConfig {
	return &MemoryLayoutConfig{
		EnableTaggedValues:            false, // Disabled — 161ns/assign overhead not worth it
		EnableCompactEnvironment:      true,  // Production-ready
		EnableCacheFriendlyStructures: true,  // Production-ready
		EnableVariableLookupCache:     false, // Still experimental
		EnableMemoryLeakDetection:     true,  // Production-ready
		MemoryLeakDetectionInterval:   300,   // 5 minutes for production
		AutoCleanupMemoryLeaks:        true,  // Safe automatic cleanup
		TaggedValuePoolSize:           1000,
		MaxStringCacheSize:            10000,
		VariableLookupCacheSize:       500,
	}
}

// ExperimentalMemoryLayoutConfig returns a configuration with all memory layout optimizations enabled
// EXPERIMENTAL: This is for testing and experimental use only - not recommended for production
// WARNING: Not compatible with concurrent functions and may cause thread safety issues
func ExperimentalMemoryLayoutConfig() *MemoryLayoutConfig {
	return &MemoryLayoutConfig{
		EnableTaggedValues:            true,
		EnableCompactEnvironment:      true,
		EnableCacheFriendlyStructures: true,
		EnableVariableLookupCache:     true, // Experimental feature
		EnableMemoryLeakDetection:     true, // Experimental aggressive detection
		MemoryLeakDetectionInterval:   30,   // 30 seconds for aggressive testing
		AutoCleanupMemoryLeaks:        true, // Experimental automatic cleanup
		TaggedValuePoolSize:           1000,
		MaxStringCacheSize:            10000,
		VariableLookupCacheSize:       500,
	}
}

// Global memory layout configuration
var globalMemoryLayoutConfig = DefaultMemoryLayoutConfig()

// SetGlobalMemoryLayoutConfig sets the global memory layout configuration
func SetGlobalMemoryLayoutConfig(config *MemoryLayoutConfig) {
	globalMemoryLayoutConfig = config
}

// GetGlobalMemoryLayoutConfig returns the global memory layout configuration
func GetGlobalMemoryLayoutConfig() *MemoryLayoutConfig {
	return globalMemoryLayoutConfig
}

// ResetGlobalMemoryLayoutConfig resets the global memory layout configuration to default
// This function should be called in test cleanup to prevent test interference
func ResetGlobalMemoryLayoutConfig() {
	globalMemoryLayoutConfig = DefaultMemoryLayoutConfig()
}

// IsTaggedValuesEnabled returns true if TaggedValue optimization is enabled
func IsTaggedValuesEnabled() bool {
	return globalMemoryLayoutConfig.EnableTaggedValues
}

// IsCompactEnvironmentEnabled returns true if CompactEnvironment optimization is enabled
func IsCompactEnvironmentEnabled() bool {
	return globalMemoryLayoutConfig.EnableCompactEnvironment
}

// IsCacheFriendlyStructuresEnabled returns true if cache-friendly structures are enabled
func IsCacheFriendlyStructuresEnabled() bool {
	return globalMemoryLayoutConfig.EnableCacheFriendlyStructures
}

// IsVariableLookupCacheEnabled returns true if variable lookup cache is enabled
func IsVariableLookupCacheEnabled() bool {
	return globalMemoryLayoutConfig.EnableVariableLookupCache
}

// GetVariableLookupCacheSize returns the configured variable lookup cache size
func GetVariableLookupCacheSize() int {
	return globalMemoryLayoutConfig.VariableLookupCacheSize
}

// IsMemoryLeakDetectionEnabled returns true if memory leak detection is enabled
func IsMemoryLeakDetectionEnabled() bool {
	return globalMemoryLayoutConfig.EnableMemoryLeakDetection
}

// GetMemoryLeakDetectionInterval returns the configured memory leak detection interval in seconds
func GetMemoryLeakDetectionInterval() int {
	return globalMemoryLayoutConfig.MemoryLeakDetectionInterval
}

// IsAutoCleanupMemoryLeaksEnabled returns true if automatic cleanup of memory leaks is enabled
func IsAutoCleanupMemoryLeaksEnabled() bool {
	return globalMemoryLayoutConfig.AutoCleanupMemoryLeaks
}
