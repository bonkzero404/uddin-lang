package interpreter

import (
	"io"
	"os"
	"time"
)

// MemoizationConfig configures memoization behavior
type MemoizationConfig struct {
	// EnableMemoization enables or disables memoization globally
	EnableMemoization bool

	// CacheSize is the maximum number of entries in the memoization cache
	CacheSize int

	// TTL is the time-to-live for cache entries (0 means no expiration)
	TTL time.Duration

	// UseProductionCache determines whether to use ProductionMemoCache (true) or OptimizedMemoCache (false)
	UseProductionCache bool

	// CleanupInterval is the interval for cleaning up expired cache entries
	CleanupInterval time.Duration
}

// DefaultMemoizationConfig returns default memoization configuration
func DefaultMemoizationConfig() *MemoizationConfig {
	return &MemoizationConfig{
		EnableMemoization:   false, // Disabled by default for safety
		CacheSize:          1000,
		TTL:                5 * time.Minute,
		UseProductionCache:  false, // Use experimental cache by default
		CleanupInterval:     1 * time.Minute,
	}
}

// StableMemoizationConfig returns stable memoization configuration for production
func StableMemoizationConfig() *MemoizationConfig {
	return &MemoizationConfig{
		EnableMemoization:   true,
		CacheSize:          5000,
		TTL:                10 * time.Minute,
		UseProductionCache:  true, // Use production cache for stability
		CleanupInterval:     2 * time.Minute,
	}
}

// ExperimentalMemoizationConfig returns experimental memoization configuration
func ExperimentalMemoizationConfig() *MemoizationConfig {
	return &MemoizationConfig{
		EnableMemoization:   true,
		CacheSize:          10000,
		TTL:                15 * time.Minute,
		UseProductionCache:  false, // Use experimental cache for testing
		CleanupInterval:     30 * time.Second,
	}
}

// Config allows you to configure the interpreter's interaction with the
// outside world. This provides a way to customize the environment in which
// the interpreted code runs.
type Config struct {
	// Vars is a map of pre-defined variables to pass into the interpreter.
	// These variables will be available to the interpreted code.
	Vars map[string]Value

	// Args is the list of command-line arguments for the interpreter's args()
	// builtin function. This simulates command-line arguments passed to a program.
	Args []string

	// Stdin is the interpreter's standard input, used by the read() builtin function.
	// If nil, it defaults to os.Stdin.
	Stdin io.Reader

	// Stdout is the interpreter's standard output, used by the print() builtin function.
	// If nil, it defaults to os.Stdout.
	Stdout io.Writer

	// Exit is a function to call when the exit() builtin is called.
	// If nil, it defaults to os.Exit.
	Exit func(int)

	// IsUnitTest menandakan bahwa interpreter sedang berjalan dalam konteks unit test
	// Jika true, fungsi main() tidak akan dijalankan secara otomatis
	IsUnitTest bool

	// DirectOutput controls whether print() writes directly to os.Stdout (true) or to configured Stdout (false)
	// When true, print() bypasses the configured Stdout and writes directly to os.Stdout
	// This is useful for CLI applications where you want immediate output without capturing
	DirectOutput bool

	// MemoryLayout configures memory layout optimizations
	// If nil, defaults to DefaultMemoryLayoutConfig()
	MemoryLayout *MemoryLayoutConfig

	// Memoization configures memoization behavior
	// If nil, defaults to DefaultMemoizationConfig()
	Memoization *MemoizationConfig
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Vars:         make(map[string]Value),
		Args:         []string{},
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Exit:         os.Exit,
		IsUnitTest:   false,
		MemoryLayout: DefaultMemoryLayoutConfig(),
		Memoization:  DefaultMemoizationConfig(),
	}
}

// TestConfig returns a configuration suitable for testing
func TestConfig() *Config {
	return &Config{
		Vars:         make(map[string]Value),
		Args:         []string{},
		Stdin:        nil,
		Stdout:       io.Discard,
		Exit:         func(int) {}, // No-op exit function for tests
		IsUnitTest:   true,
		MemoryLayout: DefaultMemoryLayoutConfig(),
		Memoization:  DefaultMemoizationConfig(),
	}
}

// Stats collects statistics about the interpreter's execution.
// These statistics are returned from Evaluate or Execute calls.
type Stats struct {
	// Ops counts the total number of operations executed
	Ops int
	// UserCalls counts the number of user-defined function calls
	UserCalls int
	// BuiltinCalls counts the number of builtin function calls
	BuiltinCalls int
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{
		Ops:          0,
		UserCalls:    0,
		BuiltinCalls: 0,
	}
}

// Reset resets all statistics to zero
func (s *Stats) Reset() {
	s.Ops = 0
	s.UserCalls = 0
	s.BuiltinCalls = 0
}

// Total returns the total number of operations
func (s *Stats) Total() int {
	return s.Ops + s.UserCalls + s.BuiltinCalls
}

// StableConfig returns a configuration with stable memory layout optimizations enabled
// These optimizations are production-ready and safe for concurrent use
func StableConfig() *Config {
	return &Config{
		Vars:         make(map[string]Value),
		Args:         []string{},
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Exit:         os.Exit,
		IsUnitTest:   false,
		MemoryLayout: StableMemoryLayoutConfig(),
		Memoization:  StableMemoizationConfig(),
	}
}

// ExperimentalConfig returns a configuration with experimental memory layout optimizations enabled
// EXPERIMENTAL: This is experimental and may have stability issues
// WARNING: Not compatible with concurrent functions - use only for sequential processing
// NOTE: This configuration should not be used in production environments
func ExperimentalConfig() *Config {
	return &Config{
		Vars:         make(map[string]Value),
		Args:         []string{},
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Exit:         os.Exit,
		IsUnitTest:   false,
		MemoryLayout: ExperimentalMemoryLayoutConfig(),
		Memoization:  ExperimentalMemoizationConfig(),
	}
}
