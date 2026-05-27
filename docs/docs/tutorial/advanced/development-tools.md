---
sidebar_position: 7
---

# Development Tools and CLI

Uddin-Lang provides various development tools and command line interface options to assist with development, debugging, and code analysis processes.

## Command Line Interface

### Basic Usage

```bash
# Run a Uddin-Lang file
uddin-lang script.din

# Run with performance profiling
uddin-lang --profile script.din

# Analyze syntax without execution
uddin-lang --analyze script.din

# Show help
uddin-lang --help

# Show version
uddin-lang --version

# List available examples
uddin-lang --examples
```

### Available Flags

#### Core Flags

```bash
# Performance profiling
uddin-lang --profile script.din
uddin-lang -p script.din

# Syntax analysis only (no execution)
uddin-lang --analyze script.din
uddin-lang -a script.din

# Memory optimization (stable - production-ready)
uddin-lang --memory-optimize-stable script.din

# Memory optimization (experimental - testing only)
uddin-lang --memory-optimize-experimental script.din

# Show help information
uddin-lang --help
uddin-lang -h

# Show version information
uddin-lang --version
uddin-lang -v

# List example files
uddin-lang --examples
uddin-lang -e
```

#### Transpiler Features

```bash
# Convert Uddin-Lang to JSON AST
uddin-lang --to_json script.din

# Convert JSON AST back to Uddin-Lang
uddin-lang --from_json ast.json
```

## Implemented Features

### Syntax Analysis

The built-in syntax analyzer checks your code for errors without executing it:

```bash
# Analyze syntax only
uddin-lang --analyze mycode.din
```

Example output for syntax errors:

```
------------------------------------------------------------------
        if (x > 10) then
        ^
------------------------------------------------------------------
Syntax Error: expected (, but got if
```

### Performance Profiling

The profiler shows execution timing and performance metrics:

```bash
# Run with profiling enabled
uddin-lang --profile mycode.din
```

Example profiling output:

```
Program executed successfully.
Output: Hello World!

=== Execution Profile ===
Parse time: 2.345ms
Execution time: 1.234ms
Total time: 3.579ms
Memory usage: 1.2MB
======================
```

### Memory Optimization

Uddin-Lang provides two levels of memory optimization to suit different use cases:

#### Stable Memory Optimization (Production-Ready)

Enable production-ready memory optimizations that are thread-safe and compatible with all language features:

```bash
# Run with stable memory optimizations
uddin-lang --memory-optimize-stable mycode.din

# Combine with profiling to measure optimization effects
uddin-lang --memory-optimize-stable --profile mycode.din
```

**✅ Stable Features:**

-   Tagged value types for reduced memory overhead
-   Compact environment structures
-   Cache-friendly data layouts
-   Thread-safe implementation
-   Compatible with concurrent functions

**Best Use Cases:**

-   Production environments
-   Applications using concurrent functions
-   Multi-threaded environments
-   Performance-critical applications requiring stability

#### Experimental Memory Optimization

Enable experimental memory optimizations for maximum performance in sequential workloads:

```bash
# Run with experimental memory optimizations
uddin-lang --memory-optimize-experimental mycode.din

# Combine with profiling to measure optimization effects
uddin-lang --memory-optimize-experimental --profile mycode.din
```

**⚠️ Experimental Features (Additional to stable):**

-   Variable lookup caching
-   Expression memoization
-   **Not compatible** with concurrent functions (`concurrent_map`, `concurrent_filter`, `concurrent_reduce`)
-   **Not thread-safe** - avoid in multi-threaded environments
-   May have stability issues

**Best Use Cases:**

-   Sequential processing workloads only
-   Development and testing environments
-   Performance benchmarking
-   Performance-critical single-threaded applications

#### Performance Comparison

**Memory vs. Speed Trade-offs:**

The choice between default mode and `--memory-optimize-stable` depends on your priorities:

**Default Mode:**

-   ⚡ Fast execution with minimal overhead
-   Lower memory efficiency
-   No memory leak detection
-   Suitable for simple scripts and development

**Stable Memory Optimization:**

-   💾 **15-30% reduction** in memory consumption
-   🎯 **Better cache performance** for large data structures
-   🔍 **Memory leak detection** for long-running applications
-   ✅ **Thread-safe** for concurrent operations
-   ⚡ Performance depends on workload - faster for memory-constrained scenarios

```bash
# Fast execution for development
uddin-lang script.din

# Production-ready with memory optimizations
uddin-lang --memory-optimize-stable script.din
```

**When Stable Mode Helps:**

Stable mode is most beneficial for:

1. **Memory Safety** over speed (leak detection, bounds checking)
2. **Production Reliability** over performance (comprehensive error handling)
3. **Thread Safety** over efficiency (synchronization mechanisms)
4. **Resource Management** over speed (cleanup routines, garbage collection)

**Memoization Benefits (Subsequent Runs):**

-   Function results are cached with LRU eviction and TTL
-   Repeated computations (like recursive functions) are served from cache
-   Cache persists across program runs within TTL window (default: 10 minutes)
-   Can achieve 5-10x speedup on computation-heavy scripts with repeated patterns

**When to Use Each Mode:**

-   **Default**: Development, testing, benchmarking, one-time scripts, performance-critical single-run applications
-   **Stable**: Production environments, long-running services, repeated script executions, applications with recursive/repetitive computations, memory-constrained systems requiring reliability

**💡 Pro Tip:** The `--memory-optimize-stable` flag is most beneficial for:

-   Scripts that run multiple times within 10 minutes (TTL window)
-   Applications with heavy recursive functions (factorial, fibonacci, etc.)
-   Repeated data processing tasks
-   Production services where the initial startup cost is acceptable for long-term performance gains

### AST Conversion Tools

#### Code to JSON AST

Convert Uddin-Lang source code to JSON Abstract Syntax Tree:

```bash
# Convert .din file to JSON AST
uddin-lang --to_json script.din > script.json
```

Example output:

```json
{
    "Statements": [
        {
            "type": "FunctionDeclaration",
            "name": "main",
            "parameters": [],
            "body": [
                {
                    "type": "CallExpression",
                    "function": "print",
                    "arguments": ["Hello World!"]
                }
            ]
        }
    ]
}
```

#### JSON AST to Code

Convert JSON AST back to Uddin-Lang source code:

```bash
# Convert JSON AST back to .din file
uddin-lang --from_json script.json > regenerated.din
```

### Error Reporting

Uddin-Lang provides detailed error reporting with:

-   **Syntax Errors**: Line and column position with visual indicators
-   **Runtime Errors**: Stack traces and error context
-   **Type Errors**: Type mismatch information
-   **Name Errors**: Undefined variable/function references

Example error output:

```
------------------------------------------------------------------
    result = undefinedFunction(x)
             ^
------------------------------------------------------------------
name error at 5:14: name 'undefinedFunction' not found
```

## Development Workflow Examples

### Basic Development Cycle

```bash
# 1. Check syntax first
uddin-lang --analyze myproject.din

# 2. Run with profiling to check performance
uddin-lang --profile myproject.din

# 3. Test with memory optimization (if compatible)
uddin-lang --memory-optimize-stable --profile myproject.din

# 4. Normal execution
uddin-lang myproject.din
```

### Working with Examples

```bash
# List all available examples
uddin-lang --examples

# Run a specific example
uddin-lang examples/01_hello_world.din

# Analyze an example's syntax
uddin-lang --analyze examples/15_array_methods_demo.din

# Profile an example
uddin-lang --profile examples/10_number_theory_demo.din

# Test memory optimization with sequential examples
uddin-lang --memory-optimize-stable examples/15_array_methods_demo.din

# Compare performance with and without optimization
uddin-lang --profile examples/10_number_theory_demo.din
uddin-lang --memory-optimize-stable --profile examples/10_number_theory_demo.din
```

### AST Analysis Workflow

```bash
# Convert source to AST for analysis
uddin-lang --to_json mycode.din > mycode_ast.json

# Inspect the AST structure
cat mycode_ast.json | jq '.'

# Convert back to source code
uddin-lang --from_json mycode_ast.json > regenerated.din

# Compare original vs regenerated
diff mycode.din regenerated.din
```

## Available Built-in Functions

Based on the actual interpreter implementation, these functions are available for development tools:

### Core Functions

-   `print()` - Output text
-   `str()` - Convert to string
-   `len()` - Get length of arrays/strings/objects
-   `typeof()` - Get type information

### Array Functions

-   `append()` - Add elements to array
-   `push()` - Add element to end of array
-   `pop()` - Remove and return last element
-   `slice()` - Extract array portion

### String Functions

-   `contains()` - Check if string contains substring
-   `split()` - Split string into array
-   `join()` - Join array elements into string
-   `trim()` - Remove whitespace

### Date/Time Functions

Requires `import "datetime"`.

-   `datetime.now()` - Get current timestamp
-   `datetime.diff()` - Calculate time differences

### Type Checking

-   `typeof()` returns: "string", "integer", "float", "boolean", "array", "object"

## Creating Development Tools in Uddin-Lang

### Simple Syntax Checker

```uddin
// syntax_check.din - Basic syntax validation
fun check_parentheses(line):
    open_count = 0
    i = 0
    while (i < len(line)):
        char = line[i]
        if (char == "(") then:
            open_count = open_count + 1
        end
        if (char == ")") then:
            open_count = open_count - 1
        end
        if (open_count < 0) then:
            return false
        end
        i = i + 1
    end
    return open_count == 0
end

fun validate_syntax(code_lines):
    errors = []
    i = 0
    while (i < len(code_lines)):
        line = code_lines[i]
        line_num = i + 1

        // Check balanced parentheses
        if (check_parentheses(line) == false) then:
            error = {
                "line": line_num,
                "message": "Unbalanced parentheses",
                "code": line
            }
            append(errors, error)
        end

        // Check for missing colons in function definitions
        if (contains(line, "fun ") and contains(line, ")") and contains(line, ":") == false) then:
            error = {
                "line": line_num,
                "message": "Missing colon after function declaration",
                "code": line
            }
            append(errors, error)
        end

        i = i + 1
    end

    return errors
end

fun main():
    // Example usage
    test_code = [
        "fun hello(",
        "    print(\"Hello World\")",
        "end"
    ]

    errors = validate_syntax(test_code)
    if (len(errors) == 0) then:
        print("✓ No syntax errors found")
    else:
        print("✗ Syntax errors detected:")
        i = 0
        while (i < len(errors)):
            error = errors[i]
            print("  Line " + str(error.line) + ": " + error.message)
            print("    Code: " + error.code)
            i = i + 1
        end
    end
end
```

### Code Statistics Tool

```uddin
// code_stats.din - Analyze code statistics
fun analyze_code_stats(lines):
    stats = {
        "total_lines": len(lines),
        "code_lines": 0,
        "comment_lines": 0,
        "blank_lines": 0,
        "function_count": 0,
        "functions": []
    }

    i = 0
    while (i < len(lines)):
        line = trim(lines[i])

        if (len(line) == 0) then:
            stats.blank_lines = stats.blank_lines + 1
        else:
            if (contains(line, "//")) then:
                stats.comment_lines = stats.comment_lines + 1
            else:
                stats.code_lines = stats.code_lines + 1
            end

            // Count functions
            if (contains(line, "fun ") and contains(line, ":")) then:
                stats.function_count = stats.function_count + 1
                // Extract function name (simplified)
                fun_start = find(line, "fun ") + 4
                paren_pos = find(line, "(")
                if (paren_pos > fun_start) then:
                    func_name = slice_string(line, fun_start, paren_pos - fun_start)
                    append(stats.functions, trim(func_name))
                end
            end
        end

        i = i + 1
    end

    return stats
end

fun slice_string(text, start, length):
    result = ""
    i = start
    while (i < start + length and i < len(text)):
        result = result + text[i]
        i = i + 1
    end
    return result
end

fun find(text, substr):
    i = 0
    while (i <= len(text) - len(substr)):
        match = true
        j = 0
        while (j < len(substr)):
            if (text[i + j] != substr[j]) then:
                match = false
                break
            end
            j = j + 1
        end
        if (match) then:
            return i
        end
        i = i + 1
    end
    return -1
end

fun print_stats(stats):
    print("=== Code Statistics ===")
    print("Total lines: " + str(stats.total_lines))
    print("Code lines: " + str(stats.code_lines))
    print("Comment lines: " + str(stats.comment_lines))
    print("Blank lines: " + str(stats.blank_lines))
    print("Functions: " + str(stats.function_count))

    if (stats.function_count > 0) then:
        print("\nFunction list:")
        i = 0
        while (i < len(stats.functions)):
            print("  - " + stats.functions[i])
            i = i + 1
        end
    end
    print("=====================")
end

fun main():
    // Example code analysis
    sample_code = [
        "// This is a comment",
        "fun hello():",
        "    print(\"Hello World\")",
        "end",
        "",
        "fun calculate(x, y):",
        "    return x + y",
        "end",
        "",
        "// Another comment"
    ]

    stats = analyze_code_stats(sample_code)
    print_stats(stats)
end
```

## Version Information

Current version: **Uddin-Lang v1.0.0**

```bash
# Check version
uddin-lang --version
# Output: Uddin-Lang v1.0.0
```

## File Requirements

-   Source files must have `.din` extension
-   JSON AST files must have `.json` extension for `--from_json`
-   All paths are relative to current working directory

## Best Practices

### Development Workflow

1. **Syntax Check First**: Always run `--analyze` before execution

    ```bash
    uddin-lang --analyze mycode.din
    ```

2. **Profile Critical Code**: Use `--profile` for performance-sensitive applications

    ```bash
    uddin-lang --profile mycode.din
    ```

3. **Test Memory Optimization**: For sequential workloads, compare with and without optimization

    ```bash
    # Test compatibility first
    uddin-lang --analyze mycode.din

    # Compare performance
    uddin-lang --profile mycode.din
    uddin-lang --memory-optimize-stable --profile mycode.din
    ```

4. **Use Examples**: Learn from existing examples
    ```bash
    uddin-lang --examples
    uddin-lang examples/01_hello_world.din
    ```

### Tool Integration

1. **Editor Integration**: Configure your editor to run syntax analysis on save
2. **Build Scripts**: Integrate analysis and profiling into build pipelines
3. **Version Control**: Track performance changes with profiling output
4. **Code Review**: Use AST conversion for detailed code structure analysis

### Error Handling Best Practices

-   Check syntax before committing code
-   Use profiling to identify performance bottlenecks
-   Leverage built-in error reporting for debugging
-   Maintain consistent code style and structure

## Integration Examples

### Shell Scripts

```bash
#!/bin/bash
# check_code.sh - Automated code checking

echo "Checking syntax..."
if uddin-lang --analyze "$1"; then
    echo "✓ Syntax OK"

    echo "Running with profiling..."
    uddin-lang --profile "$1"
else
    echo "✗ Syntax errors found"
    exit 1
fi
```

### Makefile Integration

```makefile
# Makefile for Uddin-Lang project

.PHONY: check run profile clean

SRC_FILES = $(wildcard *.din)
MAIN_FILE = main.din

check:
	@echo "Checking syntax for all .din files..."
	@for file in $(SRC_FILES); do \
		echo "Checking $$file..."; \
		uddin-lang --analyze "$$file" || exit 1; \
	done

run: check
	uddin-lang $(MAIN_FILE)

profile: check
	uddin-lang --profile $(MAIN_FILE)

ast:
	uddin-lang --to_json $(MAIN_FILE) > $(MAIN_FILE:.din=.json)

clean:
	rm -f *.json

examples:
	uddin-lang --examples
```

Dengan menggunakan tools dan CLI yang tersedia, Anda dapat mengembangkan aplikasi Uddin-Lang dengan lebih efisien dan terstruktur. Semua fitur telah disesuaikan dengan implementasi aktual interpreter Uddin-Lang.
