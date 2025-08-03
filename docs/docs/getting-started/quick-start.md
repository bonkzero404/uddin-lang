---
sidebar_position: 2
---

# Quick Start

Get up and running with Uddin-Lang in just 5 minutes! This guide will walk you through the essential concepts and syntax.

## Your First Program

Let's start with the classic "Hello, World!" program:

```uddin
// hello.din
fun main():
    print("Hello, World!")
end
```

Run it:

```bash
./uddinlang hello.din
```

Output:

```
Hello, World!
```

## Program Execution: With and Without Main Function

Uddin-Lang supports two ways to structure your programs:

### 1. Programs with Main Function (Recommended)

The `main()` function serves as the entry point of your program. When you run a Uddin-Lang program, the interpreter:

1. First executes all top-level statements (variable declarations, function definitions)
2. Then automatically calls the `main()` function if it exists

```uddin
// program_with_main.din
fun greet(name):
    return "Hello, " + name + "!"
end

fun main():
    message = greet("World")
    print(message)
    print("Program executed successfully!")
end
```

**Benefits of using main():**

-   Clear program structure and entry point
-   Better organization for larger programs
-   Follows common programming conventions
-   Easier to test individual functions

### 2. Programs without Main Function (Script Style)

You can also write programs that execute top-level statements directly, without a `main()` function. This is useful for:

-   Simple scripts
-   Library files
-   Configuration files
-   Quick prototypes

```uddin
// script_without_main.din
// This code executes immediately when the file is run

PI = 3.14159
E = 2.71828

fun calculate_circle_area(radius):
    return PI * radius * radius
end

fun factorial(n):
    if (n <= 1) then:
        return 1
    else:
        return n * factorial(n - 1)
    end
end

// These statements execute immediately
print("Math Library Loaded!")
print("Available constants: PI = " + str(PI) + ", E = " + str(E))
print("Available functions: calculate_circle_area(), factorial()")

// Example usage
area = calculate_circle_area(5)
print("Area of circle with radius 5: " + str(area))

fact_5 = factorial(5)
print("Factorial of 5: " + str(fact_5))
```

**When to use script style:**

-   Creating reusable libraries
-   Writing simple automation scripts
-   Quick calculations or data processing
-   Configuration or setup files

### Key Differences

| Aspect          | With Main Function                | Without Main Function     |
| --------------- | --------------------------------- | ------------------------- |
| **Execution**   | Top-level statements → main()     | Top-level statements only |
| **Structure**   | Organized, clear entry point      | Linear, script-like       |
| **Best for**    | Applications, larger programs     | Libraries, simple scripts |
| **Testing**     | Easy to test functions separately | Code executes immediately |
| **Reusability** | Functions can be imported         | Can be used as library    |

### Import Behavior

When you import a file using `import "filename.din"`, the interpreter:

-   Executes all top-level statements in the imported file
-   **Does NOT** call the `main()` function of the imported file
-   Makes all functions and variables available in your current scope

This allows you to create library files with `main()` for testing, while still being able to import them safely.

## Basic Syntax Overview

### Variables and Data Types

Uddin-Lang uses dynamic typing - no need to declare variable types:

```uddin
fun main():
    // Numbers
    age = 25
    price = 19.99

    // Strings
    name = "Alice"
    message = "Welcome to Uddin-Lang!"

    // Booleans
    is_active = true
    is_complete = false

    // Arrays
    numbers = [1, 2, 3, 4, 5]
    names = ["Alice", "Bob", "Charlie"]

    // Objects (Maps)
    person = {
        "name": "Alice",
        "age": 30,
        "city": "Jakarta"
    }

    print(name)
    print(numbers)
    print(person)
end
```

### Functions

Functions are first-class citizens in Uddin-Lang:

```uddin
fun main():
    // Simple function
    fun greet(name):
        return "Hello, " + name + "!"
    end

    // Function with multiple parameters
    fun add(a, b):
        return a + b
    end

    // Using functions
    message = greet("Developer")
    sum = add(10, 20)

    print(message)  // Hello, Developer!
    print(sum)      // 30
end
```

### Control Flow

#### Conditional Statements

```uddin
fun main():
    age = 18

    if (age >= 18) then:
        print("You are an adult")
    else:
        print("You are a minor")
    end

    // Multiple conditions
    score = 85

    if (score >= 90) then:
        print("Grade: A")
    else if (score >= 80) then:
        print("Grade: B")
    else if (score >= 70) then:
        print("Grade: C")
    else:
        print("Grade: F")
    end
end
```

#### Loops

```uddin
fun main():
    // For loop with range
    for (i in range(5)):
        print("Count: " + str(i))
    end

    // For loop with array
    fruits = ["apple", "banana", "orange"]
    for (fruit in fruits):
        print("Fruit: " + fruit)
    end

    // While loop
    count = 0
    while (count < 3):
        print("While count: " + str(count))
        count = count + 1
    end
end
```

### Working with Arrays

```uddin
fun main():
    // Creating arrays
    numbers = [1, 2, 3, 4, 5]

    // Accessing elements
    first = numbers[0]
    last = numbers[4]

    print("First: " + str(first))  // First: 1
    print("Last: " + str(last))    // Last: 5

    // Array operations
    push(numbers, 6)               // Add element
    length = len(numbers)          // Get length

    print("Length: " + str(length)) // Length: 6
    print(numbers)                  // [1, 2, 3, 4, 5, 6]
end
```

### Working with Objects

```uddin
fun main():
    // Creating objects
    person = {
        "name": "Alice",
        "age": 30,
        "email": "alice@example.com"
    }

    // Accessing properties
    name = person["name"]
    age = person["age"]

    print("Name: " + name)         // Name: Alice
    print("Age: " + str(age))      // Age: 30

    // Adding/updating properties
    person["city"] = "Jakarta"
    person["age"] = 31

    print(person)
end
```

## Practical Examples

### Example 1: Calculator

```uddin
fun calculator(operation, a, b):
    if (operation == "add") then:
        return a + b
    else if (operation == "subtract") then:
        return a - b
    else if (operation == "multiply") then:
        return a * b
    else if (operation == "divide") then:
        if (b != 0) then:
            return a / b
        else:
            return "Error: Division by zero"
        end
    else:
        return "Error: Unknown operation"
    end
end

fun main():
    // Test the calculator
    result1 = calculator("add", 10, 5)
    result2 = calculator("multiply", 7, 3)
    result3 = calculator("divide", 15, 3)

    print("10 + 5 = " + str(result1))    // 10 + 5 = 15
    print("7 * 3 = " + str(result2))     // 7 * 3 = 21
    print("15 / 3 = " + str(result3))    // 15 / 3 = 5
end
```

### Example 2: Working with Data

```uddin
// Calculate average grade
fun calculate_average(students):
    total = 0
    count = len(students)

    for (student in students):
        total = total + student["grade"]
    end

    return total / count
end

// Find top student
fun find_top_student(students):
    top_student = students[0]

    for (student in students):
        if (student["grade"] > top_student["grade"]) then:
            top_student = student
        end
    end

    return top_student
end

fun main():
    // Student data
    students = [
        {"name": "Alice", "grade": 85},
        {"name": "Bob", "grade": 92},
        {"name": "Charlie", "grade": 78}
    ]

    average = calculate_average(students)
    top = find_top_student(students)

    print("Average grade: " + str(average))
    print("Top student: " + top["name"] + " (" + str(top["grade"]) + ")")
end
```

## Built-in Functions

Uddin-Lang provides many useful built-in functions:

```uddin
fun main():
    // String functions
    text = "Hello, World!"
    print("Length: " + str(len(text)))        // Length: 13
    print("Uppercase: " + upper(text))        // Uppercase: HELLO, WORLD!
    print("Lowercase: " + lower(text))        // Lowercase: hello, world!

    // Math functions
    print("Absolute: " + str(abs(-5)))        // Absolute: 5
    print("Maximum: " + str(max(10, 20)))     // Maximum: 20
    print("Minimum: " + str(min(10, 20)))     // Minimum: 10

    // Type conversion
    number = 42
    text_num = str(number)                    // Convert to string
    print("Number as string: " + text_num)

    // Array functions
    numbers = [3, 1, 4, 1, 5]
    sorted_nums = sort(numbers)
    print("Sorted: " + str(sorted_nums))      // Sorted: [1, 1, 3, 4, 5]
end
```

## Error Handling

Basic error handling patterns:

```uddin
fun safe_divide(a, b):
    if (b == 0) then:
        print("Error: Cannot divide by zero")
        return null
    end
    return a / b
end

fun safe_access(arr, index):
    if (index < 0 or index >= len(arr)) then:
        print("Error: Index out of bounds")
        return null
    end
    return arr[index]
end

fun main():
    // Test error handling
    result = safe_divide(10, 0)  // Error: Cannot divide by zero

    numbers = [1, 2, 3]
    value = safe_access(numbers, 5)  // Error: Index out of bounds
end
```

## CLI Development Tools

Uddin-Lang includes powerful development tools:

### Syntax Analysis

```bash
# Check syntax without running
./uddinlang --analyze your_script.din
```

### Performance Profiling

```bash
# Run with performance profiling
./uddinlang --profile your_script.din
```

### View Examples

```bash
# List available examples
./uddinlang --examples

# Run a specific example
./uddinlang examples/01_hello_world.din
```

## Next Steps

Now that you've learned the basics:

1. **[Write Your First Program](your-first-program.md)** - Create a complete application
2. **[Tutorial Series](../tutorial/basics/introduction.md)** - Deep dive into language features
3. **[Language Reference](../reference/syntax.md)** - Complete syntax reference

## Quick Reference Card

```uddin
// Comments
// Single line comment

// Variables
name = "value"
number = 42
array = [1, 2, 3]
object = {"key": "value"}

// Functions
fun function_name(param1, param2):
    return param1 + param2
end

// Conditionals
if (condition) then:
    // code
else if (other_condition) then:
    // code
else:
    // code
end

// Loops
for (item in array):
    // code
end

while (condition):
    // code
end

// Built-ins
print(value)
len(array)
str(number)
push(array, item)
```

## Practical Examples

### Example 1: Calculator Program (With Main)

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
        print("Error: Division by zero!")
        return null
    end
    return a / b
end

fun main():
    print("=== Simple Calculator ===")

    a = 10
    b = 5

    print("Numbers: " + str(a) + " and " + str(b))
    print("Addition: " + str(add(a, b)))
    print("Subtraction: " + str(subtract(a, b)))
    print("Multiplication: " + str(multiply(a, b)))
    print("Division: " + str(divide(a, b)))
end
```

**Output when run:**

```
=== Simple Calculator ===
Numbers: 10 and 5
Addition: 15
Subtraction: 5
Multiplication: 50
Division: 2
```

### Example 2: Math Utilities Library (Without Main)

```uddin
// math_utils.din
// Mathematical constants
PI = 3.14159265359
E = 2.71828182846
GOLDEN_RATIO = 1.61803398875

// Utility functions
fun square(x):
    return x * x
end

fun cube(x):
    return x * x * x
end

fun circle_area(radius):
    return PI * square(radius)
end

fun circle_circumference(radius):
    return 2 * PI * radius
end

// Library initialization message
print("Math Utilities Library loaded successfully!")
print("Available constants: PI, E, GOLDEN_RATIO")
print("Available functions: square(), cube(), circle_area(), circle_circumference()")
print("")

// Demonstrate some calculations
print("=== Demo Calculations ===")
radius = 3
print("Circle with radius " + str(radius) + ":")
print("  Area: " + str(circle_area(radius)))
print("  Circumference: " + str(circle_circumference(radius)))
print("")
print("Powers of 4:")
print("  Square: " + str(square(4)))
print("  Cube: " + str(cube(4)))
```

**Output when run:**

```
Math Utilities Library loaded successfully!
Available constants: PI, E, GOLDEN_RATIO
Available functions: square(), cube(), circle_area(), circle_circumference()

=== Demo Calculations ===
Circle with radius 3:
  Area: 28.274333882
  Circumference: 18.849555922

Powers of 4:
  Square: 16
  Cube: 64
```

### Example 3: Using Library in Main Program

```uddin
// geometry_app.din
import "math_utils.din"

fun calculate_room_area(length, width):
    return length * width
end

fun main():
    print("=== Geometry Application ===")
    print("")

    // Using imported functions
    room_length = 5
    room_width = 4
    room_area = calculate_room_area(room_length, room_width)

    print("Room dimensions: " + str(room_length) + "m x " + str(room_width) + "m")
    print("Room area: " + str(room_area) + " square meters")
    print("")

    // Using imported circle functions
    table_radius = 1.5
    table_area = circle_area(table_radius)

    print("Round table radius: " + str(table_radius) + "m")
    print("Table area: " + str(table_area) + " square meters")
    print("")

    // Calculate remaining space
    remaining_space = room_area - table_area
    print("Remaining floor space: " + str(remaining_space) + " square meters")
end
```

**Key Takeaways:**

1. **Programs with `main()`** are ideal for applications that need a clear entry point
2. **Programs without `main()`** work great as libraries or for immediate script execution
3. **Import behavior** allows you to use library files safely without executing their `main()` function
4. **Both styles** can coexist - libraries can have `main()` for testing while still being importable

---

**Next**: [Your First Program →](your-first-program.md)
