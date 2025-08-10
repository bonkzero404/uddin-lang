---
id: examples
title: Code Examples from Repository
sidebar_label: Examples
---

# Uddin Programming Language - Code Examples

This page contains all the practical examples from the main repository's `examples/` folder. Each example demonstrates specific features and capabilities of the Uddin programming language.

## 🚀 Getting Started Examples

### 1. Hello World
**File:** `01_hello_world.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/01_hello_world.din)

The classic first program that introduces basic syntax, variables, and output.

```uddin
// Main program function
fun main():
    print("Welcome to Uddin-Lang!")
    print("Hello, World!")

    // Example variables and basic operations
    name = "Uddin-Lang"
    version = "1.0"

    print("Programming language: " + name)
    print("Version: " + version)

    // Example mathematical operations
    a = 10
    b = 5
    result = a + b

    print("Result of " + str(a) + " + " + str(b) + " = " + str(result))
end
```

**Features demonstrated:**
- Function definitions
- Variable assignments
- String concatenation
- Basic arithmetic
- Type conversion with `str()`

### 1.1. Digit Separators Demo
**File:** `31_digit_separator_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/31_digit_separator_demo.din)

Demonstrates the use of digit separators for improved readability of large numbers.

```uddin
fun main():
    print("=== Digit Separator Demo ===")
    print()
    
    // Basic Integer Examples
    one_million = 1_000_000
    one_billion = 1_000_000_000
    
    print("Basic Integer Examples:")
    print("One million: " + str(one_million))
    print("One billion: " + str(one_billion))
    print()
    
    // Floating Point Examples
    pi_precise = 3.141_592_653_589_793
    large_decimal = 123_456.789_012
    
    print("Floating Point Examples:")
    print("Pi (precise): " + str(pi_precise))
    print("Large decimal: " + str(large_decimal))
    print()
    
    // Scientific Notation Examples
    avogadro = 6.022_140_76e23
    speed_of_light = 2.997_924_58e8
    planck = 6.626_070_15e-34
    
    print("Scientific Notation Examples:")
    print("Avogadro's number: " + str(avogadro))
    print("Speed of light (m/s): " + str(speed_of_light))
    print("Planck constant: " + str(planck))
end
```

**Features demonstrated:**
- Digit separators in integers
- Digit separators in floating-point numbers
- Digit separators in scientific notation
- Improved code readability for large numbers

### 2. Functional Programming Demo
**File:** `02_functional_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/02_functional_demo.din)

Demonstrates functional programming concepts including recursion and higher-order functions.

```uddin
// Function to calculate factorial using recursion
fun factorial(n):
    if (n <= 1) then:
        return 1
    else:
        return n * factorial(n - 1)
    end
end

// Function to calculate Fibonacci sequence
fun fibonacci(n):
    if (n <= 0) then:
        return 0
    else:
        if (n == 1) then:
            return 1
        else:
            return fibonacci(n - 1) + fibonacci(n - 2)
        end
    end
end
```

**Features demonstrated:**
- Recursive functions
- Conditional statements
- Mathematical computations
- Function composition

## 🧮 Mathematical Examples

### 3. Math Library
**File:** `03_math_library.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/03_math_library.din)

Comprehensive mathematical functions and algorithms.

### 4. Simple Calculator
**File:** `04_simple_calculator.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/04_simple_calculator.din)

A basic calculator implementation with various mathematical operations.

```uddin
// Mathematical constants
PI = 3.14159

// Simple power function
fun power(base, exponent):
    if (exponent == 0) then:
        return 1
    end

    result = 1
    i = 0
    while (i < exponent):
        result = result * base
        i = i + 1
    end
    return result
end

// Circle area calculation
fun circleArea(radius):
    return PI * radius * radius
end
```

### 7. Built-in Math Demo
**File:** `07_builtin_math_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/07_builtin_math_demo.din)

Demonstrates built-in mathematical functions.

### 8. Statistics Analysis
**File:** `08_statistics_analysis.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/08_statistics_analysis.din)

Statistical calculations and data analysis functions.

### 9. Geometry Calculator
**File:** `09_geometry_calculator.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/09_geometry_calculator.din)

Geometric calculations for various shapes.

### 10. Number Theory Demo
**File:** `10_number_theory_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/10_number_theory_demo.din)

Number theory algorithms and mathematical concepts.

## 🔄 Control Flow Examples

### 5. Loop Control Demo
**File:** `05_loop_control_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/05_loop_control_demo.din)

Demonstrates various loop constructs and control flow.

### 11. Ternary Examples
**File:** `11_ternary_examples.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/11_ternary_examples.din)

Ternary operator usage and conditional expressions.

### 12. Logical Operators
**File:** `12_logical_operators.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/12_logical_operators.din)

Logical operations and boolean expressions.

### 13. Assignment Operators
**File:** `13_assignment_operators.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/13_assignment_operators.din)

Various assignment operators and their usage.

## 📝 Data Manipulation Examples

### 14. String Manipulation Demo
**File:** `14_string_manipulation_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/14_string_manipulation_demo.din)

Comprehensive string operations and text processing.

```uddin
fun main():
    // replace() - Replace substring
    text = "Hello World"
    replaced = replace(text, "World", "Universe")
    print("replace('" + text + "', 'World', 'Universe') = '" + replaced + "'")

    // trim() - Remove whitespace
    messy = "   Hello World   "
    trimmed = trim(messy)
    print("trim('" + messy + "') = '" + trimmed + "'")

    // starts_with() and ends_with()
    filename = "document.pdf"
    print("starts_with('" + filename + "', 'doc') = " + str(starts_with(filename, "doc")))
    print("ends_with('" + filename + "', '.pdf') = " + str(ends_with(filename, ".pdf")))
end
```

### 15. Array Methods Demo
**File:** `15_array_methods_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/15_array_methods_demo.din)

Array operations, manipulation, and built-in methods.

### 16. Data Structures Demo
**File:** `16_data_structures_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/16_data_structures_demo.din)

Advanced data structures and their implementations.

## 🌐 Networking Examples

### 17. HTTP Client Demo
**File:** `17_http_client_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/17_http_client_demo.din)

HTTP client operations for web requests.

```uddin
// Demo HTTP GET request
fun try_get_request():
    print("  Making GET request to JSONPlaceholder...")

    // GET request to public API
    response = http_get("https://jsonplaceholder.typicode.com/posts/1")

    print("  Status: " + str(response["status"]))
    print("  Status Text: " + response["status_text"])

    // Display response body
    body = response["body"]
    if (typeof(body) == "object") then:
        print("  Title: " + body.title)
    end
end
```

### 18. Networking Demo
**File:** `18_networking_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/18_networking_demo.din)

Network programming and communication examples.

### 19. Persistent HTTP Server
**File:** `19_persistent_http_server.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/19_persistent_http_server.din)

HTTP server implementation and request handling.

## 📊 Data Format Examples

### 20. JSON Handling Demo
**File:** `20_json_handling_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/20_json_handling_demo.din)

JSON parsing, serialization, and manipulation.

```uddin
fun main():
    print("=== JSON Handling Demo ===")

    // Parse simple JSON object
    json_string = '{"name": "John", "age": 30, "city": "New York"}'
    print("Original JSON string: " + json_string)

    parsed_object = json_parse(json_string)
    print("Parsed object: " + str(parsed_object))
    print("Name: " + parsed_object.name)
    print("Age: " + str(parsed_object.age))
    print("City: " + parsed_object.city)
end
```

### 21. XML Handling Demo
**File:** `21_xml_handling_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/21_xml_handling_demo.din)

XML parsing and processing capabilities.

### 21. Multiline Strings Demo
**File:** `21_multiline_strings_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/21_multiline_strings_demo.din)

Multiline string handling and formatting.

## 🔍 Advanced Pattern Matching

### 22. Rule Engine Regex Demo
**File:** `22_rule_engine_regex_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/22_rule_engine_regex_demo.din)

Regular expressions and pattern matching.

### 23. Rule Engine DateTime Demo
**File:** `23_rule_engine_datetime_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/23_rule_engine_datetime_demo.din)

Date and time processing with rule engines.

### 24. Rule Engine Fact Database Demo
**File:** `24_rule_engine_fact_database_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/24_rule_engine_fact_database_demo.din)

Fact-based rule engine implementation.

### 25. Rule Engine CEP Demo
**File:** `25_rule_engine_cep_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/25_rule_engine_cep_demo.din)

Complex Event Processing (CEP) with rule engines.

### 26. Rule Engine Comprehensive Demo
**File:** `26_rule_engine_comprehensive_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/26_rule_engine_comprehensive_demo.din)

Comprehensive rule engine system demonstration.

## 📁 System Integration

### 6. Import Library Demo
**File:** `06_import_library_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/06_import_library_demo.din)

Module system and library imports.

### 27. Filesystem Operations Demo
**File:** `27_filesystem_operations_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/27_filesystem_operations_demo.din)

File system operations and file handling.

## 🧪 Experimental Features

:::warning Experimental
The following examples demonstrate experimental features that are not production-ready. Use with caution.
:::

### Function Memoization Demo
**File:** `33_memoization_demo.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/33_memoization_demo.din)

Demonstrates automatic function memoization for performance optimization.

```uddin
// Function without memoization
fun fibonacci_normal(n):
    if (n <= 1) then:
        return n
    else:
        return fibonacci_normal(n - 1) + fibonacci_normal(n - 2)
    end
end

// Function with memoization (experimental)
memo fun fibonacci_memo(n):
    if (n <= 1) then:
        return n
    else:
        return fibonacci_memo(n - 1) + fibonacci_memo(n - 2)
    end
end
```

**Features demonstrated:**
- `memo` keyword for automatic caching
- Performance comparison between memoized and non-memoized functions
- Fibonacci sequence optimization

### Memoization Verification Test
**File:** `34_memoization_verification_test.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/34_memoization_verification_test.din)

Verifies that memoization only works on functions marked with the `memo` keyword.

**Features demonstrated:**
- Testing memoization behavior
- Comparing cached vs non-cached function calls
- Verification of experimental feature functionality

## 📚 Utility Libraries

### Math Library
**File:** `math_library.din`

Reusable mathematical library functions.

## 🧪 Testing and Debugging

### XML Debug
**File:** `xml_debug.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/xml_debug.din)

XML debugging and testing utilities.

### XML Attributes Test
**File:** `xml_attributes_test.din` | [📁 View on GitHub](https://github.com/bonkzero404/uddin-lang/blob/main/examples/xml_attributes_test.din)

XML attribute handling and testing.

## Running the Examples

To run any of these examples:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/bonkzero404/uddin-lang.git
   cd uddin-lang
   ```

2. **Navigate to examples:**
   ```bash
   cd examples
   ```

3. **Run an example:**
   ```bash
   uddin 01_hello_world.din
   ```

## Example Categories Summary

| Category | Count | Description |
|----------|-------|-------------|
| **Getting Started** | 2 | Basic syntax and functional programming |
| **Mathematical** | 6 | Math operations, statistics, geometry |
| **Control Flow** | 4 | Loops, conditionals, operators |
| **Data Manipulation** | 3 | Strings, arrays, data structures |
| **Networking** | 3 | HTTP clients, servers, networking |
| **Data Formats** | 3 | JSON, XML, multiline strings |
| **Pattern Matching** | 5 | Regex, rule engines, CEP |
| **System Integration** | 2 | Imports, filesystem operations |
| **Experimental** | 2 | Memoization and performance features |
| **Utilities** | 1 | Reusable libraries |
| **Testing** | 2 | Debug and test utilities |

**Total Examples:** 33 files

---

*All examples are maintained in the main repository and regularly updated. Check the [GitHub repository](https://github.com/bonkzero404/uddin-lang/tree/main/examples) for the latest versions.*
