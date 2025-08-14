# UDDIN-LANG Example Usages

This directory contains comprehensive examples demonstrating various features and use cases of the UDDIN-LANG library. Each example is organized in its own folder with a dedicated `main.go` file.

## Directory Structure

```
example_usages/
├── basic_usage/          # Basic library usage examples
├── ast_conversion/       # AST conversion and manipulation
├── advanced_usage/       # Advanced features and configurations
├── concurrent_usage/     # Concurrent execution patterns
├── web_integration/      # Web service integration examples
└── README.md            # This file
```

## Running Examples

Each example can be run independently. Navigate to the specific folder and run:

```bash
cd <example_folder>
go run main.go
```

### Available Examples

#### 1. Basic Usage (`basic_usage/`)
Demonstrates fundamental UDDIN-LANG operations:
- Simple arithmetic calculations
- Variable assignments
- Function definitions and calls
- Control flow (if/else statements)
- Arrays and loops

```bash
cd basic_usage
go run main.go
```

#### 2. AST Conversion (`ast_conversion/`)
Shows how to work with Abstract Syntax Trees:
- Converting code to JSON AST
- Converting JSON AST back to code
- Round-trip conversions
- Byte array operations

```bash
cd ast_conversion
go run main.go
```

#### 3. Advanced Usage (`advanced_usage/`)
Covers advanced library features:
- Custom variable settings
- File execution
- Expression parsing and evaluation
- Program parsing
- Error handling
- Complex data structures

```bash
cd advanced_usage
go run main.go
```

#### 4. Concurrent Usage (`concurrent_usage/`)
Demonstrates concurrent execution patterns:
- Worker pools
- Concurrent task execution
- Race condition handling
- Performance benchmarking

```bash
cd concurrent_usage
go run main.go
```

#### 5. Web Integration (`web_integration/`)
Shows how to integrate UDDIN-LANG with web services:
- HTTP API endpoints
- JSON request/response handling
- Web server setup (commented out for safety)
- RESTful service examples

```bash
cd web_integration
go run main.go
```

## Dependencies

Each example folder contains its own `go.mod` file with the necessary dependencies. The examples use a local replace directive to reference the main UDDIN-LANG library:

```go
replace github.com/bonkzero404/uddin-lang => ../../
```

## Notes

- All examples are self-contained and can be run independently
- Each example includes detailed comments explaining the functionality
- The web integration example includes commented-out server code for safety
- All UDDIN-LANG syntax follows the correct format with proper loop syntax: `for (item in array):`

## Getting Started

To get started with UDDIN-LANG examples:

1. Clone the repository
2. Navigate to any example folder
3. Run `go mod tidy` (if needed)
4. Execute `go run main.go`

Each example will output detailed information about what it's demonstrating and the results of the operations.