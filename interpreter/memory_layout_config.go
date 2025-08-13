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
		TaggedValuePoolSize:           1000,
		MaxStringCacheSize:            10000,
		VariableLookupCacheSize:       500,
	}
}

// ExperimentalMemoryLayoutConfig returns a configuration with all memory layout optimizations enabled
// This is for testing and experimental use only
func ExperimentalMemoryLayoutConfig() *MemoryLayoutConfig {
	return &MemoryLayoutConfig{
		EnableTaggedValues:            true,
		EnableCompactEnvironment:      true,
		EnableCacheFriendlyStructures: true,
		EnableVariableLookupCache:     true,
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