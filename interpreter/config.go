package interpreter

import (
	"io"
	"os"
)

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
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Vars:       make(map[string]Value),
		Args:       []string{},
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Exit:       os.Exit,
		IsUnitTest: false,
	}
}

// TestConfig returns a configuration suitable for testing
func TestConfig() *Config {
	return &Config{
		Vars:       make(map[string]Value),
		Args:       []string{},
		Stdin:      nil,
		Stdout:     io.Discard,
		Exit:       func(int) {}, // No-op exit function for tests
		IsUnitTest: true,
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
