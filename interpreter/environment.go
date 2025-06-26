package interpreter

import (
	"io"
	"maps"
)

// Environment manages the execution environment of the interpreter
type Environment struct {
	// vars is a stack of variable scopes, with the most local scope at the end
	vars []map[string]Value
	// args holds command-line arguments for the args() builtin
	args []string
	// stdin is the input stream for the read() builtin
	stdin io.Reader
	// stdout is the output stream for the print() builtin
	stdout io.Writer
	// exit is the function called by the exit() builtin
	exit func(int)
	// inUnitTest indicates if we're running in unit test mode
	inUnitTest bool
}

// NewEnvironment creates a new execution environment
func NewEnvironment(config *Config) *Environment {
	env := &Environment{
		vars:       []map[string]Value{make(map[string]Value)},
		args:       config.Args,
		stdin:      config.Stdin,
		stdout:     config.Stdout,
		exit:       config.Exit,
		inUnitTest: config.IsUnitTest,
	}

	// Set default I/O if not provided
	if env.stdin == nil {
		env.stdin = io.MultiReader() // No input by default
	}
	if env.stdout == nil {
		env.stdout = io.Discard // No output by default
	}
	if env.exit == nil {
		env.exit = func(code int) {
			// Do nothing in default case
		}
	}

	// Initialize with predefined variables
	if config.Vars != nil {
		maps.Copy(env.vars[0], config.Vars)
	}

	// Initialize builtin functions
	for name, builtin := range builtins {
		env.vars[0][name] = builtin
	}

	return env
}

// PushScope creates a new variable scope
func (env *Environment) PushScope() {
	env.vars = append(env.vars, make(map[string]Value))
}

// PopScope removes the most local variable scope
func (env *Environment) PopScope() {
	if len(env.vars) > 1 {
		env.vars = env.vars[:len(env.vars)-1]
	}
}

// Assign sets a variable in the most local scope
func (env *Environment) Assign(name string, value Value) {
	env.vars[len(env.vars)-1][name] = value
}

// AssignOuter sets a variable in the global scope
func (env *Environment) AssignOuter(name string, value Value) {
	env.vars[0][name] = value
}

// Lookup retrieves a variable value, searching from local to global scope
func (env *Environment) Lookup(name string) (Value, bool) {
	// Search from most local to global scope
	for i := len(env.vars) - 1; i >= 0; i-- {
		if value, exists := env.vars[i][name]; exists {
			return value, true
		}
	}
	return nil, false
}

// GetStdin returns the input stream
func (env *Environment) GetStdin() io.Reader {
	return env.stdin
}

// GetStdout returns the output stream
func (env *Environment) GetStdout() io.Writer {
	return env.stdout
}

// GetExit returns the exit function
func (env *Environment) GetExit() func(int) {
	return env.exit
}

// GetArgs returns the command line arguments
func (env *Environment) GetArgs() []string {
	return env.args
}

// IsUnitTest returns whether we're in unit test mode
func (env *Environment) IsUnitTest() bool {
	return env.inUnitTest
}

// ScopeCount returns the number of active scopes
func (env *Environment) ScopeCount() int {
	return len(env.vars)
}
