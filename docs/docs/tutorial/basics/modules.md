---
sidebar_position: 8
---

# Modules and Import System

Uddin-Lang provides a powerful module system that allows you to organize code into reusable libraries and import functionality from other files.

## Learning Objectives

By the end of this chapter, you will understand:
- How to create reusable modules
- How to import code from other files
- Best practices for organizing modular code
- Creating and using library files
- Module execution behavior

## What are Modules?

Modules in Uddin-Lang are simply `.din` files that contain functions, variables, and code that can be imported and used in other programs. This allows you to:

- **Organize code** into logical units
- **Reuse functionality** across multiple programs
- **Create libraries** of common functions
- **Maintain cleaner** and more manageable codebases

## Creating a Module

Any `.din` file can serve as a module. Here's how to create a simple math library:

```uddin
// math_utils.din
// A simple mathematical utilities module

// Function to calculate square
fun square(n):
    return n * n
end

// Function to calculate cube
fun cube(n):
    return n * n * n
end

// Function to check if number is even
fun isEven(n):
    return n % 2 == 0
end

// Function to check if number is odd
fun isOdd(n):
    return n % 2 != 0
end

// Optional: You can include a main function for testing
fun main():
    print("Math utilities module loaded!")
    print("Testing square(5):", square(5))
    print("Testing cube(3):", cube(3))
end
```

## Importing Modules

### Basic Import Syntax

Use the `import` statement to load code from another file:

```uddin
// main.din
import "math_utils.din"

fun main():
    // Now you can use functions from math_utils.din
    result = square(10)
    print("Square of 10 is:", result)
    
    print("Is 8 even?", isEven(8))
    print("Is 7 odd?", isOdd(7))
end
```

### Import with Relative Paths

You can import files from different directories:

```uddin
// Import from subdirectory
import "libraries/string_utils.din"

// Import from parent directory
import "../shared/common.din"

// Import from specific path
import "utils/math/advanced.din"
```

## Module Execution Behavior

### Important: Main Function Handling

When you import a module, Uddin-Lang has a special behavior:

- **All functions and variables** from the imported file become available
- **The `main()` function is NOT automatically executed** during import
- **Top-level code** (outside functions) IS executed during import

```uddin
// library.din
print("This will run when imported")  // Executes during import

value = 42  // This variable will be available

fun helper():
    return "Helper function"
end

fun main():
    print("This will NOT run during import")  // Does not execute
end
```

```uddin
// main.din
import "library.din"  // Prints "This will run when imported"

fun main():
    print("Value from library:", value)  // Prints: 42
    print(helper())  // Prints: "Helper function"
end
```

## Practical Examples

### Example 1: String Utilities Library

```uddin
// string_utils.din
// String manipulation utilities

fun capitalize_first(text):
    if (len(text) == 0) then:
        return text
    end
    first = upper(text[0:1])
    rest = lower(text[1:])
    return first + rest
end

fun reverse_string(text):
    result = ""
    i = len(text) - 1
    while (i >= 0):
        result = result + text[i:i+1]
        i = i - 1
    end
    return result
end

fun count_words(text):
    words = split(text, " ")
    return len(words)
end

print("String utilities loaded!")
```

```uddin
// text_processor.din
import "string_utils.din"

fun main():
    text = "hello world programming"
    
    print("Original:", text)
    print("Capitalized:", capitalize_first(text))
    print("Reversed:", reverse_string(text))
    print("Word count:", count_words(text))
end
```

### Example 2: Data Processing Library

```uddin
// data_utils.din
// Data processing utilities

fun find_max(numbers):
    if (len(numbers) == 0) then:
        return null
    end
    
    max_val = numbers[0]
    for (num in numbers):
        if (num > max_val) then:
            max_val = num
        end
    end
    return max_val
end

fun find_min(numbers):
    if (len(numbers) == 0) then:
        return null
    end
    
    min_val = numbers[0]
    for (num in numbers):
        if (num < min_val) then:
            min_val = num
        end
    end
    return min_val
end

fun calculate_average(numbers):
    if (len(numbers) == 0) then:
        return 0
    end
    
    total = 0
    for (num in numbers):
        total = total + num
    end
    return total / len(numbers)
end

fun filter_even(numbers):
    result = []
    for (num in numbers):
        if (num % 2 == 0) then:
            result = push(result, num)
        end
    end
    return result
end
```

```uddin
// analyzer.din
import "data_utils.din"

fun main():
    data = [15, 8, 23, 42, 7, 16, 31, 4]
    
    print("Data:", data)
    print("Maximum:", find_max(data))
    print("Minimum:", find_min(data))
    print("Average:", calculate_average(data))
    print("Even numbers:", filter_even(data))
end
```

### Example 3: Configuration Module

```uddin
// config.din
// Application configuration

// Application settings
APP_NAME = "My Uddin App"
VERSION = "1.0.0"
DEBUG_MODE = true

// Database settings
DB_HOST = "localhost"
DB_PORT = 5432
DB_NAME = "myapp_db"

// API settings
API_BASE_URL = "https://api.example.com"
API_TIMEOUT = 30

fun get_app_info():
    return {
        "name": APP_NAME,
        "version": VERSION,
        "debug": DEBUG_MODE
    }
end

fun get_db_config():
    return {
        "host": DB_HOST,
        "port": DB_PORT,
        "database": DB_NAME
    }
end

print("Configuration loaded: " + APP_NAME + " v" + VERSION)
```

```uddin
// app.din
import "config.din"

fun main():
    app_info = get_app_info()
    db_config = get_db_config()
    
    print("Starting", app_info["name"], "version", app_info["version"])
    
    if (app_info["debug"]) then:
        print("Debug mode enabled")
        print("Database:", db_config["host"] + ":" + str(db_config["port"]))
    end
    
    print("API URL:", API_BASE_URL)
end
```

## Module Organization Best Practices

### 1. **Use Descriptive Names**

```uddin
// Good
import "math_utilities.din"
import "string_helpers.din"
import "database_connection.din"

// Avoid
import "utils.din"
import "helpers.din"
import "stuff.din"
```

### 2. **Organize by Functionality**

```
project/
├── main.din
├── config.din
├── utils/
│   ├── math_utils.din
│   ├── string_utils.din
│   └── date_utils.din
├── data/
│   ├── database.din
│   └── file_handler.din
└── api/
    ├── client.din
    └── parser.din
```

### 3. **Include Documentation**

```uddin
// math_utils.din
// Mathematical utility functions for common calculations
// 
// Functions:
//   - square(n): Returns n squared
//   - cube(n): Returns n cubed
//   - factorial(n): Returns n factorial
//   - isPrime(n): Checks if n is prime

fun square(n):
    // Calculate the square of a number
    return n * n
end
```

### 4. **Provide Module Testing**

```uddin
// math_utils.din
fun square(n):
    return n * n
end

fun cube(n):
    return n * n * n
end

// Test the module when run directly
fun main():
    print("Testing math_utils module:")
    print("square(5) =", square(5))  // Should be 25
    print("cube(3) =", cube(3))      // Should be 27
    
    // Test edge cases
    print("square(0) =", square(0))  // Should be 0
    print("cube(-2) =", cube(-2))    // Should be -8
end
```

### 5. **Handle Dependencies**

```uddin
// advanced_math.din
import "basic_math.din"  // Import dependencies first

fun calculate_hypotenuse(a, b):
    // Uses square() from basic_math.din
    return sqrt(square(a) + square(b))
end
```

## Common Patterns

### Library with Initialization

```uddin
// logger.din
LOG_LEVEL = "INFO"
LOG_FILE = "app.log"

fun log_info(message):
    print("[INFO] " + message)
end

fun log_error(message):
    print("[ERROR] " + message)
end

fun set_log_level(level):
    LOG_LEVEL = level
end

// Initialize logger
print("Logger initialized with level: " + LOG_LEVEL)
```

### Utility Collections

```uddin
// validators.din
// Collection of validation functions

fun is_email(email):
    // Simple email validation
    return len(email) > 0 and index_of(email, "@") > 0
end

fun is_phone(phone):
    // Simple phone validation
    return len(phone) >= 10
end

fun is_not_empty(text):
    return len(text) > 0
end

fun is_numeric(text):
    // Check if string contains only numbers
    for (i = 0; i < len(text); i = i + 1):
        char = text[i:i+1]
        if (char < "0" or char > "9") then:
            return false
        end
    end
    return true
end
```

## Error Handling in Modules

### Graceful Error Handling

```uddin
// file_utils.din
fun safe_divide(a, b):
    if (b == 0) then:
        print("Error: Division by zero")
        return null
    end
    return a / b
end

fun validate_input(value, min_val, max_val):
    if (value < min_val or value > max_val) then:
        print("Error: Value must be between", min_val, "and", max_val)
        return false
    end
    return true
end
```

## Advanced Module Techniques

### Conditional Loading

```uddin
// main.din
DEBUG = true

if (DEBUG) then:
    import "debug_utils.din"
end

fun main():
    if (DEBUG) then:
        debug_log("Application started")
    end
    
    print("Hello, World!")
end
```

### Module Factories

```uddin
// connection_factory.din
fun create_database_connection(host, port, database):
    return {
        "host": host,
        "port": port,
        "database": database,
        "connected": false
    }
end

fun connect(connection):
    print("Connecting to", connection["host"] + ":" + str(connection["port"]))
    connection["connected"] = true
    return connection
end
```

## Practice Exercises

### Exercise 1: Create a Calculator Module

Create a `calculator.din` module with basic arithmetic functions:

```uddin
// calculator.din
// TODO: Implement these functions
// - add(a, b)
// - subtract(a, b)
// - multiply(a, b)
// - divide(a, b)
// - power(base, exponent)
// - percentage(value, percent)
```

**Solution**:
```uddin
// calculator.din
fun add(a, b):
    return a + b
end

fun subtract(a, b):
    return a - b
end

fun multiply(a, b):
    return a * b
end

fun divide(a, b):
    if (b == 0) then:
        print("Error: Cannot divide by zero")
        return null
    end
    return a / b
end

fun power(base, exponent):
    return base ** exponent
end

fun percentage(value, percent):
    return (value * percent) / 100
end

fun main():
    print("Calculator module test:")
    print("10 + 5 =", add(10, 5))
    print("10 - 5 =", subtract(10, 5))
    print("10 * 5 =", multiply(10, 5))
    print("10 / 5 =", divide(10, 5))
    print("2^3 =", power(2, 3))
    print("20% of 100 =", percentage(100, 20))
end
```

### Exercise 2: Array Utilities Module

Create an `array_utils.din` module with array manipulation functions:

```uddin
// array_utils.din
// TODO: Implement these functions
// - sum_array(arr)
// - find_duplicates(arr)
// - remove_duplicates(arr)
// - sort_array(arr)
// - reverse_array(arr)
```

**Solution**:
```uddin
// array_utils.din
fun sum_array(arr):
    total = 0
    for (item in arr):
        total = total + item
    end
    return total
end

fun find_duplicates(arr):
    duplicates = []
    for (i = 0; i < len(arr); i = i + 1):
        for (j = i + 1; j < len(arr); j = j + 1):
            if (arr[i] == arr[j] and index_of(duplicates, arr[i]) == -1) then:
                duplicates = push(duplicates, arr[i])
            end
        end
    end
    return duplicates
end

fun remove_duplicates(arr):
    unique = []
    for (item in arr):
        if (index_of(unique, item) == -1) then:
            unique = push(unique, item)
        end
    end
    return unique
end

fun reverse_array(arr):
    result = []
    for (i = len(arr) - 1; i >= 0; i = i - 1):
        result = push(result, arr[i])
    end
    return result
end

fun main():
    test_array = [1, 2, 3, 2, 4, 1, 5]
    print("Original array:", test_array)
    print("Sum:", sum_array(test_array))
    print("Duplicates:", find_duplicates(test_array))
    print("Unique:", remove_duplicates(test_array))
    print("Reversed:", reverse_array(test_array))
end
```

## Summary

In this chapter, you learned:

✅ **Module Creation**: How to create reusable `.din` files  
✅ **Import System**: Using `import` to load external code  
✅ **Module Behavior**: Understanding execution during import  
✅ **Organization**: Best practices for structuring modular code  
✅ **Practical Examples**: Real-world module patterns  
✅ **Error Handling**: Safe module design techniques  

### Key Takeaways

1. **Modularity**: Break code into logical, reusable units
2. **Import Behavior**: Main functions don't execute during import
3. **Organization**: Use clear naming and directory structure
4. **Testing**: Include main functions for module testing
5. **Documentation**: Comment your modules clearly

### Best Practices Summary

- Use descriptive module names
- Organize modules by functionality
- Include documentation and examples
- Handle errors gracefully
- Test modules independently
- Keep modules focused on single responsibilities

---

**Next**: Continue exploring advanced Uddin-Lang features or practice building your own modular applications!