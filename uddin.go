// Package uddin provides a high-level API for the UDDIN-LANG interpreter.
// UDDIN-LANG (Unified Dynamic Decision Interpreter Notation) is a specialized
// rule engine platform that combines programming language expressiveness with
// business decision system simplicity.
package uddin

import (
	"encoding/json"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

// Engine represents a UDDIN-LANG interpreter engine instance.
// It provides a high-level interface for parsing and executing UDDIN-LANG code.
type Engine struct {
	config       *interpreter.Config
	programCache sync.Map // *interpreter.Program → *interpreter.CompiledProgram
	vmPool       sync.Pool
}

// New creates a new UDDIN-LANG engine with default configuration.
func New() *Engine {
	e := &Engine{
		config: interpreter.DefaultConfig(),
	}
	e.vmPool.New = func() any {
		return interpreter.NewVM(nil)
	}
	return e
}

// NewWithConfig creates a new UDDIN-LANG engine with custom configuration.
func NewWithConfig(config *interpreter.Config) *Engine {
	if config == nil {
		config = interpreter.DefaultConfig()
	}
	e := &Engine{config: config}
	e.vmPool.New = func() any {
		return interpreter.NewVM(nil)
	}
	return e
}

// SetStdout sets the standard output for the engine.
func (e *Engine) SetStdout(w io.Writer) {
	e.config.Stdout = w
}

// SetStdin sets the standard input for the engine.
func (e *Engine) SetStdin(r io.Reader) {
	e.config.Stdin = r
}

// SetArgs sets command-line arguments for the engine.
func (e *Engine) SetArgs(args []string) {
	e.config.Args = args
}

// SetVariable sets a variable in the engine's global scope.
func (e *Engine) SetVariable(name string, value any) {
	if e.config.Vars == nil {
		e.config.Vars = make(map[string]interpreter.Value)
	}
	e.config.Vars[name] = value
}

// SetVariables sets multiple variables in the engine's global scope.
func (e *Engine) SetVariables(vars map[string]any) {
	if e.config.Vars == nil {
		e.config.Vars = make(map[string]interpreter.Value)
	}
	for name, value := range vars {
		e.config.Vars[name] = value
	}
}

// EnableMemoryOptimization enables memory layout optimizations.
func (e *Engine) EnableMemoryOptimization() {
	e.config.MemoryLayout = interpreter.DefaultMemoryLayoutConfig()
}

// SetUnitTestMode sets whether the engine is running in unit test mode.
// In unit test mode, the main() function is not automatically executed.
func (e *Engine) SetUnitTestMode(isUnitTest bool) {
	e.config.IsUnitTest = isUnitTest
}

// ParseExpression parses a UDDIN-LANG expression from source code.
func (e *Engine) ParseExpression(source []byte) (interpreter.Expression, error) {
	return interpreter.ParseExpression(source)
}

// ParseProgram parses a UDDIN-LANG program from source code.
func (e *Engine) ParseProgram(source []byte) (*interpreter.Program, error) {
	return interpreter.ParseProgram(source)
}

// EvaluateExpression evaluates a parsed expression and returns the result.
func (e *Engine) EvaluateExpression(expr interpreter.Expression) (interpreter.Value, *interpreter.Stats, error) {
	return interpreter.Evaluate(expr, e.config)
}

// ExecuteProgram executes a parsed program.
// When VMEnabled, compiled bytecode is cached by *Program pointer — subsequent
// calls with the same *Program skip compilation entirely.
func (e *Engine) ExecuteProgram(prog *interpreter.Program) (*interpreter.Stats, error) {
	if !e.config.VMEnabled {
		return interpreter.Execute(prog, e.config)
	}

	// Build sorted var-name list for compilation (stable across calls).
	varNames := make([]string, 0, len(e.config.Vars))
	for k := range e.config.Vars {
		varNames = append(varNames, k)
	}
	sort.Strings(varNames)

	// Cache lookup by *Program pointer identity.
	var cp *interpreter.CompiledProgram
	if val, ok := e.programCache.Load(prog); ok {
		cp = val.(*interpreter.CompiledProgram)
	} else {
		var err error
		cp, err = interpreter.CompileProgram(prog, varNames)
		if err != nil {
			return nil, err
		}
		e.programCache.Store(prog, cp)
	}

	// Borrow a VM from the pool, use it, return it.
	vm := e.vmPool.Get().(*interpreter.VM)
	defer e.vmPool.Put(vm)

	return interpreter.ExecuteCompiledVM(cp, e.config, vm)
}

// ExecuteString parses and executes UDDIN-LANG source code from a string.
func (e *Engine) ExecuteString(source string) (*interpreter.Stats, error) {
	prog, err := e.ParseProgram([]byte(source))
	if err != nil {
		return nil, err
	}
	return e.ExecuteProgram(prog)
}

// ExecuteFile parses and executes UDDIN-LANG source code from a file.
func (e *Engine) ExecuteFile(filename string) (*interpreter.Stats, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return e.ExecuteString(string(content))
}

// EvaluateString parses and evaluates a UDDIN-LANG expression from a string.
func (e *Engine) EvaluateString(source string) (interpreter.Value, *interpreter.Stats, error) {
	expr, err := e.ParseExpression([]byte(source))
	if err != nil {
		return nil, nil, err
	}
	return e.EvaluateExpression(expr)
}

// ConvertToJSON converts UDDIN-LANG source code to JSON AST representation.
func (e *Engine) ConvertToJSON(source []byte) ([]byte, error) {
	program, err := e.ParseProgram(source)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(program, "", "  ")
}

// ConvertFromJSON converts JSON AST representation back to UDDIN-LANG source code.
func (e *Engine) ConvertFromJSON(jsonData []byte) (string, error) {
	var program interpreter.Program
	if err := json.Unmarshal(jsonData, &program); err != nil {
		return "", err
	}
	return program.String(), nil
}

// ConvertStringToJSON converts UDDIN-LANG source code string to JSON AST.
func (e *Engine) ConvertStringToJSON(source string) ([]byte, error) {
	return e.ConvertToJSON([]byte(source))
}

// ConvertJSONToString converts JSON AST back to UDDIN-LANG source code string.
func (e *Engine) ConvertJSONToString(jsonData []byte) (string, error) {
	return e.ConvertFromJSON(jsonData)
}

// Convenience functions for quick usage without creating an engine instance

// ParseExpression is a convenience function to parse an expression with default settings.
func ParseExpression(source []byte) (interpreter.Expression, error) {
	return interpreter.ParseExpression(source)
}

// ParseProgram is a convenience function to parse a program with default settings.
func ParseProgram(source []byte) (*interpreter.Program, error) {
	return interpreter.ParseProgram(source)
}

// ExecuteString is a convenience function to execute source code with default settings.
func ExecuteString(source string) (*interpreter.Stats, error) {
	engine := New()
	return engine.ExecuteString(source)
}

// ExecuteFile is a convenience function to execute a file with default settings.
func ExecuteFile(filename string) (*interpreter.Stats, error) {
	engine := New()
	return engine.ExecuteFile(filename)
}

// EvaluateString is a convenience function to evaluate an expression with default settings.
func EvaluateString(source string) (interpreter.Value, *interpreter.Stats, error) {
	engine := New()
	return engine.EvaluateString(source)
}

// ConvertToJSON is a convenience function that converts source code to JSON AST.
func ConvertToJSON(source []byte) ([]byte, error) {
	engine := New()
	return engine.ConvertToJSON(source)
}

// ConvertFromJSON is a convenience function that converts JSON AST back to source code.
func ConvertFromJSON(jsonData []byte) (string, error) {
	engine := New()
	return engine.ConvertFromJSON(jsonData)
}

// ConvertStringToJSON is a convenience function that converts source code string to JSON AST.
func ConvertStringToJSON(source string) ([]byte, error) {
	engine := New()
	return engine.ConvertStringToJSON(source)
}

// ConvertJSONToString is a convenience function that converts JSON AST back to source code string.
func ConvertJSONToString(jsonData []byte) (string, error) {
	engine := New()
	return engine.ConvertJSONToString(jsonData)
}

// Re-export commonly used types and functions from interpreter package
type (
	// Value represents any runtime value in UDDIN-LANG
	Value = interpreter.Value
	// Expression represents a parsed expression
	Expression = interpreter.Expression
	// Program represents a parsed program
	Program = interpreter.Program
	// Config represents interpreter configuration
	Config = interpreter.Config
	// Stats represents execution statistics
	Stats = interpreter.Stats
	// Error represents interpreter errors
	Error = interpreter.Error
)

// Re-export configuration functions
var (
	// DefaultConfig returns a configuration with sensible defaults
	DefaultConfig = interpreter.DefaultConfig
	// TestConfig returns a configuration suitable for testing
	TestConfig = interpreter.TestConfig
	// PrintStats prints execution statistics
	PrintStats = interpreter.PrintStats
)
