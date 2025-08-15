---
id: library
title: Go Library
sidebar_label: Go Library
---

# UDDIN-LANG Go Library

UDDIN-LANG (Unified Dynamic Decision Interpreter Notation) is available as a Go library that you can easily integrate into your Go applications.

## Installation

```bash
go get github.com/bonkzero404/uddin-lang
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    // Create a new engine
    engine := uddin.New()

    // Execute UDDIN-LANG code
    stats, err := engine.ExecuteString(`
        x = 10
        y = 20
        result = x + y
        print("Result: " + str(result))
    `)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Executed %d operations\n", stats.Total())
}
```

### Evaluating Expressions

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Evaluate a mathematical expression
    result, stats, err := engine.EvaluateString("(5 + 3) * 2")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Result: %v (executed %d operations)\n", result, stats.Total())
    // Output: Result: 16 (executed 5 operations)
}
```

### Setting Variables

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Set individual variables
    engine.SetVariable("name", "John")
    engine.SetVariable("age", 30)

    // Or set multiple variables at once
    engine.SetVariables(map[string]any{
        "city": "New York",
        "active": true,
    })

    // Use the variables in UDDIN-LANG code
    _, err := engine.ExecuteString(`
        print("Name: " + name)
        print("Age: " + str(age))
        print("City: " + city)
        print("Active: " + str(active))
    `)

    if err != nil {
        log.Fatal(err)
    }
}
```

### Custom I/O

```go
package main

import (
    "bytes"
    "fmt"
    "log"
    "strings"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Capture output
    var output bytes.Buffer
    engine.SetStdout(&output)

    // Provide input
    input := strings.NewReader("Hello from input\n")
    engine.SetStdin(input)

    _, err := engine.ExecuteString(`
        print("Enter your name: ")
        name = read()
        print("Hello, " + name)
    `)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Captured output:\n%s", output.String())
}
```

### Working with Functions

```go
package main

import (
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Define and call functions
    _, err := engine.ExecuteString(`
        fun factorial(n):
            if (n <= 1) then:
                return 1
            end
            return n * factorial(n - 1)
        end

        result = factorial(5)
        print("5! = " + str(result))
    `)

    if err != nil {
        log.Fatal(err)
    }
}
```

### Error Handling

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // This will cause a syntax error
    _, err := engine.ExecuteString("x = ")
    if err != nil {
        fmt.Printf("Syntax error: %v\n", err)
    }

    // This will cause a runtime error
    _, err = engine.ExecuteString("print(undefined_variable)")
    if err != nil {
        fmt.Printf("Runtime error: %v\n", err)
    }
}
```

### Parsing and Execution Separation

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Parse the program first
    program, err := engine.ParseProgram([]byte(`
        fun greet(name):
            return "Hello, " + name + "!"
        end

        message = greet("World")
        print(message)
    `))

    if err != nil {
        log.Fatal("Parse error:", err)
    }

    // Execute the parsed program
    stats, err := engine.ExecuteProgram(program)
    if err != nil {
        log.Fatal("Execution error:", err)
    }

    fmt.Printf("Program executed successfully with %d operations\n", stats.Total())
}
```

### Memory Optimization

```go
package main

import (
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Enable memory optimizations for better performance
    engine.EnableMemoryOptimization()

    // Execute memory-intensive operations
    _, err := engine.ExecuteString(`
        // Process large arrays
        data = range(10000)
        sum = 0
        for (item in data):
            sum = sum + item
        end
        print("Sum: " + str(sum))
    `)

    if err != nil {
        log.Fatal(err)
    }
}
```

### Unit Testing Mode

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Enable unit test mode (prevents automatic main() execution)
    engine.SetUnitTestMode(true)

    // Define functions without executing main
    _, err := engine.ExecuteString(`
        fun add(a, b):
            return a + b
        end

        fun main():
            print("This won't be executed automatically")
        end
    `)

    if err != nil {
        log.Fatal(err)
    }

    // Manually test the add function
    result, _, err := engine.EvaluateString("add(5, 3)")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("add(5, 3) = %v\n", result) // Output: add(5, 3) = 8
}
```

## Convenience Functions

For simple use cases, you can use the convenience functions that don't require creating an engine instance:

```go
package main

import (
    "fmt"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    // Quick expression evaluation
    result, _, err := uddin.EvaluateString("10 * 5 + 2")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Result: %v\n", result) // Output: Result: 52

    // Quick program execution
    _, err = uddin.ExecuteString(`print("Hello from UDDIN-LANG!")`)
    if err != nil {
        log.Fatal(err)
    }

    // Quick file execution
    _, err = uddin.ExecuteFile("script.din")
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Reference

### Engine Methods

#### Creation and Configuration
- `New() *Engine` - Create a new engine with default configuration
- `NewWithConfig(config *Config) *Engine` - Create a new engine with custom configuration
- `SetStdout(w io.Writer)` - Set standard output
- `SetStdin(r io.Reader)` - Set standard input
- `SetArgs(args []string)` - Set command-line arguments
- `EnableMemoryOptimization()` - Enable memory optimizations
- `SetUnitTestMode(isUnitTest bool)` - Set unit test mode

#### Variable Management
- `SetVariable(name string, value any)` - Set a single variable
- `SetVariables(vars map[string]any)` - Set multiple variables

#### Parsing
- `ParseExpression(source []byte) (Expression, error)` - Parse an expression
- `ParseProgram(source []byte) (*Program, error)` - Parse a program

#### Execution
- `EvaluateExpression(expr Expression) (Value, *Stats, error)` - Evaluate a parsed expression
- `ExecuteProgram(prog *Program) (*Stats, error)` - Execute a parsed program
- `ExecuteString(source string) (*Stats, error)` - Parse and execute source code
- `ExecuteFile(filename string) (*Stats, error)` - Parse and execute a file
- `EvaluateString(source string) (Value, *Stats, error)` - Parse and evaluate an expression

#### AST Conversion
- `ConvertToJSON(source []byte) ([]byte, error)` - Convert UDDIN-LANG source to JSON AST
- `ConvertFromJSON(jsonData []byte) (string, error)` - Convert JSON AST back to UDDIN-LANG source
- `ConvertStringToJSON(source string) ([]byte, error)` - Convert UDDIN-LANG string to JSON AST
- `ConvertJSONToString(jsonData []byte) (string, error)` - Convert JSON AST back to UDDIN-LANG string

### Convenience Functions

- `ParseExpression(source []byte) (Expression, error)` - Parse an expression with default settings
- `ParseProgram(source []byte) (*Program, error)` - Parse a program with default settings
- `ExecuteString(source string) (*Stats, error)` - Execute source code with default settings
- `ExecuteFile(filename string) (*Stats, error)` - Execute a file with default settings
- `EvaluateString(source string) (Value, *Stats, error)` - Evaluate an expression with default settings
- `ConvertToJSON(source []byte) ([]byte, error)` - Convert UDDIN-LANG source to JSON AST
- `ConvertFromJSON(jsonData []byte) (string, error)` - Convert JSON AST back to UDDIN-LANG source
- `ConvertStringToJSON(source string) ([]byte, error)` - Convert UDDIN-LANG string to JSON AST
- `ConvertJSONToString(jsonData []byte) (string, error)` - Convert JSON AST back to UDDIN-LANG string

### Configuration Functions

- `DefaultConfig() *Config` - Get default configuration
- `TestConfig() *Config` - Get configuration suitable for testing
- `PrintStats(stats Stats, output io.Writer)` - Print execution statistics

## AST Conversion Examples

```go
// Convert source code to JSON AST
source := "x = 42\nprint(x)"
jsonData, err := uddin.ConvertStringToJSON(source)
if err != nil {
    log.Fatal(err)
}

// Convert JSON AST back to source code
convertedSource, err := uddin.ConvertJSONToString(jsonData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Original: %s\n", source)
fmt.Printf("Converted: %s\n", convertedSource)

// Round-trip conversion with byte arrays
originalBytes := []byte("fun factorial(n): if (n <= 1) then: return 1 else: return n * factorial(n - 1) end end")
jsonBytes, err := uddin.ConvertToJSON(originalBytes)
if err != nil {
    log.Fatal(err)
}

restoredSource, err := uddin.ConvertFromJSON(jsonBytes)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Restored: %s\n", restoredSource)
```

## Integration Examples

### Web Server Integration

```go
package main

import (
    "encoding/json"
    "net/http"
    "log"

    "github.com/bonkzero404/uddin-lang"
)

type EvalRequest struct {
    Expression string            `json:"expression"`
    Variables  map[string]any `json:"variables,omitempty"`
}

type EvalResponse struct {
    Result any `json:"result"`
    Error  string     `json:"error,omitempty"`
    Stats  *uddin.Stats `json:"stats"`
}

func evalHandler(w http.ResponseWriter, r *http.Request) {
    var req EvalRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    engine := uddin.New()
    engine.SetUnitTestMode(true)

    if req.Variables != nil {
        engine.SetVariables(req.Variables)
    }

    result, stats, err := engine.EvaluateString(req.Expression)

    resp := EvalResponse{
        Result: result,
        Stats:  stats,
    }

    if err != nil {
        resp.Error = err.Error()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func main() {
    http.HandleFunc("/eval", evalHandler)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### CLI Tool Integration

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/bonkzero404/uddin-lang"
)

func main() {
    var (
        file    = flag.String("file", "", "UDDIN-LANG file to execute")
        expr    = flag.String("expr", "", "UDDIN-LANG expression to evaluate")
        profile = flag.Bool("profile", false, "Show execution statistics")
    )
    flag.Parse()

    engine := uddin.New()

    if *file != "" {
        stats, err := engine.ExecuteFile(*file)
        if err != nil {
            log.Fatal(err)
        }

        if *profile {
            uddin.PrintStats(*stats, os.Stderr)
        }
    } else if *expr != "" {
        result, stats, err := engine.EvaluateString(*expr)
        if err != nil {
            log.Fatal(err)
        }

        fmt.Printf("Result: %v\n", result)

        if *profile {
            uddin.PrintStats(*stats, os.Stderr)
        }
    } else {
        fmt.Println("Please specify either -file or -expr")
        flag.Usage()
        os.Exit(1)
    }
}
```

## Performance Tips

1. **Reuse Engine Instances**: Create an engine once and reuse it for multiple executions
2. **Enable Memory Optimization**: Use `EnableMemoryOptimization()` for memory-intensive operations
3. **Parse Once, Execute Many**: Parse programs once and execute them multiple times
4. **Use Unit Test Mode**: Enable unit test mode when you don't need automatic main() execution
5. **Capture Output**: Use custom stdout/stderr to avoid console I/O overhead in production

## Error Handling

The library provides detailed error information including:
- Syntax errors with line and column numbers
- Runtime errors with context
- Type errors with expected vs actual types
- Name errors for undefined variables/functions

All errors implement the standard Go `error` interface and can be handled using normal Go error handling patterns.

## Variable Management

### Setting Variables

The library provides `SetVariable` and `SetVariables` methods to pass data from your Go application to UDDIN-LANG scripts:

```go
engine := uddin.New()

// Set individual variables
engine.SetVariable("username", "alice")
engine.SetVariable("score", 95)
engine.SetVariable("active", true)

// Set multiple variables at once
engine.SetVariables(map[string]any{
    "config": map[string]any{
        "timeout": 30,
        "retries": 3,
    },
    "items": []string{"apple", "banana", "cherry"},
})
```

### Variable Access Pattern

UDDIN-LANG follows a **one-way data flow** pattern:

1. **Go → UDDIN**: Use `SetVariable`/`SetVariables` to pass data to scripts
2. **UDDIN → Go**: Use return values, output capture, or result objects to get data back

```go
// Pass data to script
engine.SetVariable("input", 42)

// Get result back via return value
result, _, err := engine.EvaluateString("input * 2")
fmt.Printf("Result: %v\n", result) // Result: 84

// Or capture output
var output bytes.Buffer
engine.SetStdout(&output)
engine.ExecuteString(`print("Processed: " + str(input * 2))`)
fmt.Printf("Output: %s", output.String())
```

This design promotes clean separation between your Go application logic and UDDIN-LANG scripts, making the integration more predictable and maintainable.
