---
sidebar_position: 6
---

# Functions

Functions are reusable blocks of code that perform specific tasks. They help organize your code, reduce repetition, and make programs more maintainable.

## Function Basics

### Defining Functions

Use the `fun` keyword to define a function:

```uddin
fun greet():
    print("Hello, World!")
end

fun main():
    greet()  // Call the function
end
```

### Functions with Parameters

Functions can accept input values called parameters:

```uddin
fun greet_person(name):
    print("Hello, " + name + "!")
end

fun main():
    greet_person("Alice")
    greet_person("Bob")
end
```

### Multiple Parameters

```uddin
fun introduce(name, age, city):
    print("Hi, I'm " + name)
    print("I'm " + str(age) + " years old")
    print("I live in " + city)
end

fun main():
    introduce("Alice", 25, "New York")
end
```

## Return Values

### Returning Values

Functions can return values using the `return` statement:

```uddin
fun add(a, b):
    return a + b
end

fun main():
    result = add(5, 3)
    print("5 + 3 = " + str(result))
end
```

### Multiple Return Statements

```uddin
fun get_grade(score):
    if (score >= 90) then:
        return "A"
    else if (score >= 80) then:
        return "B"
    else if (score >= 70) then:
        return "C"
    else if (score >= 60) then:
        return "D"
    else:
        return "F"
    end
end

fun main():
    grade = get_grade(85)
    print("Grade: " + grade)
end
```

### Early Returns

```uddin
fun divide(a, b):
    if (b == 0) then:
        print("Error: Division by zero")
        return null
    end
    
    return a / b
end

fun main():
    result = divide(10, 2)
    if (result != null) then:
        print("Result: " + str(result))
    end
end
```

## Mathematical Functions

### Basic Math Operations

```uddin
fun square(x):
    return x * x
end

fun cube(x):
    return x * x * x
end

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

fun main():
    print("Square of 5: " + str(square(5)))
    print("Cube of 3: " + str(cube(3)))
    print("2^8: " + str(power(2, 8)))
end
```

### Factorial Function

```uddin
fun factorial(n):
    if (n <= 1) then:
        return 1
    end
    
    result = 1
    i = 2
    while (i <= n):
        result = result * i
        i = i + 1
    end
    return result
end

fun main():
    print("5! = " + str(factorial(5)))
    print("0! = " + str(factorial(0)))
end
```

## Recursive Functions

Functions can call themselves, creating recursion:

### Recursive Factorial

```uddin
fun factorial_recursive(n):
    if (n <= 1) then:
        return 1
    else:
        return n * factorial_recursive(n - 1)
    end
end

fun main():
    print("5! = " + str(factorial_recursive(5)))
end
```

### Fibonacci Sequence

```uddin
fun fibonacci(n):
    if (n <= 0) then:
        return 0
    else if (n == 1) then:
        return 1
    else:
        return fibonacci(n - 1) + fibonacci(n - 2)
    end
end

fun main():
    print("Fibonacci sequence:")
    i = 0
    while (i < 10):
        print("F(" + str(i) + ") = " + str(fibonacci(i)))
        i = i + 1
    end
end
```

## Array and String Functions

### Array Processing

```uddin
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

fun main():
    scores = [85, 92, 78, 96, 88]
    
    max_score = find_max(scores)
    avg_score = calculate_average(scores)
    
    print("Highest score: " + str(max_score))
    print("Average score: " + str(avg_score))
end
```

### String Functions

```uddin
fun count_vowels(text):
    vowels = "aeiouAEIOU"
    count = 0
    
    i = 0
    while (i < len(text)):
        char = text[i]
        if (contains(vowels, char)) then:
            count = count + 1
        end
        i = i + 1
    end
    return count
end

fun reverse_string(text):
    result = ""
    i = len(text) - 1
    while (i >= 0):
        result = result + text[i]
        i = i - 1
    end
    return result
end

fun main():
    word = "Hello"
    print("Vowels in '" + word + "': " + str(count_vowels(word)))
    print("Reversed: " + reverse_string(word))
end
```

## Function Scope

### Local Variables

```uddin
fun calculate_area(length, width):
    area = length * width  // Local variable
    return area
end

fun main():
    result = calculate_area(5, 3)
    print("Area: " + str(result))
    // area is not accessible here
end
```

### Global Variables

```uddin
// Global variable
PI = 3.14159

fun circle_area(radius):
    return PI * radius * radius  // Using global variable
end

fun circle_circumference(radius):
    return 2 * PI * radius  // Using global variable
end

fun main():
    radius = 5
    area = circle_area(radius)
    circumference = circle_circumference(radius)
    
    print("Circle with radius " + str(radius) + ":")
    print("Area: " + str(area))
    print("Circumference: " + str(circumference))
end
```

## Higher-Order Functions

Functions that work with other functions:

### Function as Parameter

```uddin
fun apply_operation(a, b, operation):
    if (operation == "add") then:
        return a + b
    else if (operation == "multiply") then:
        return a * b
    else if (operation == "subtract") then:
        return a - b
    else:
        return 0
    end
end

fun main():
    x = 10
    y = 5
    
    sum = apply_operation(x, y, "add")
    product = apply_operation(x, y, "multiply")
    
    print("Sum: " + str(sum))
    print("Product: " + str(product))
end
```

## Practical Examples

### Example 1: Temperature Converter

```uddin
fun celsius_to_fahrenheit(celsius):
    return (celsius * 9 / 5) + 32
end

fun fahrenheit_to_celsius(fahrenheit):
    return (fahrenheit - 32) * 5 / 9
end

fun kelvin_to_celsius(kelvin):
    return kelvin - 273.15
end

fun celsius_to_kelvin(celsius):
    return celsius + 273.15
end

fun main():
    temp_c = 25
    temp_f = celsius_to_fahrenheit(temp_c)
    temp_k = celsius_to_kelvin(temp_c)
    
    print(str(temp_c) + "°C = " + str(temp_f) + "°F")
    print(str(temp_c) + "°C = " + str(temp_k) + "K")
end
```

### Example 2: Simple Calculator

```uddin
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
        print("Error: Division by zero")
        return null
    end
    return a / b
end

fun calculator(a, b, operation):
    if (operation == "+") then:
        return add(a, b)
    else if (operation == "-") then:
        return subtract(a, b)
    else if (operation == "*") then:
        return multiply(a, b)
    else if (operation == "/") then:
        return divide(a, b)
    else:
        print("Error: Unknown operation")
        return null
    end
end

fun main():
    a = 10
    b = 3
    
    print(str(a) + " + " + str(b) + " = " + str(calculator(a, b, "+")))
    print(str(a) + " - " + str(b) + " = " + str(calculator(a, b, "-")))
    print(str(a) + " * " + str(b) + " = " + str(calculator(a, b, "*")))
    print(str(a) + " / " + str(b) + " = " + str(calculator(a, b, "/")))
end
```

## Best Practices

### 1. Single Responsibility

```uddin
// Good: Each function has one clear purpose
fun calculate_area(length, width):
    return length * width
end

fun print_area(area):
    print("Area: " + str(area))
end

// Avoid: Function doing too many things
fun calculate_and_print_area(length, width):
    area = length * width
    print("Calculating...")
    print("Length: " + str(length))
    print("Width: " + str(width))
    print("Area: " + str(area))
    return area
end
```

### 2. Meaningful Names

```uddin
// Good: Descriptive function names
fun calculate_monthly_payment(principal, rate, months):
    return principal * rate / (1 - power(1 + rate, -months))
end

// Avoid: Unclear names
fun calc(p, r, m):
    return p * r / (1 - power(1 + r, -m))
end
```

### 3. Keep Functions Small

```uddin
// Good: Small, focused function
fun is_even(number):
    return number % 2 == 0
end

// Good: Break complex logic into smaller functions
fun process_numbers(numbers):
    evens = filter_even_numbers(numbers)
    sum = calculate_sum(evens)
    return sum
end
```

### 4. Handle Edge Cases

```uddin
fun safe_divide(a, b):
    if (b == 0) then:
        print("Warning: Division by zero")
        return null
    end
    return a / b
end

fun safe_array_access(array, index):
    if (index < 0 || index >= len(array)) then:
        print("Warning: Index out of bounds")
        return null
    end
    return array[index]
end
```

## Practice Exercises

### Exercise 1: Prime Number Checker

```uddin
fun is_prime(n):
    // Implement prime number checking
    // Return true if n is prime, false otherwise
end

fun main():
    numbers = [2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
    for (num in numbers):
        if (is_prime(num)) then:
            print(str(num) + " is prime")
        else:
            print(str(num) + " is not prime")
        end
    end
end
```

### Exercise 2: Array Statistics

```uddin
fun find_min(numbers):
    // Find minimum value in array
end

fun find_max(numbers):
    // Find maximum value in array
end

fun calculate_sum(numbers):
    // Calculate sum of all numbers
end

fun calculate_average(numbers):
    // Calculate average of all numbers
end

fun main():
    data = [12, 5, 8, 3, 16, 9, 1, 14]
    
    print("Data: " + str(data))
    print("Min: " + str(find_min(data)))
    print("Max: " + str(find_max(data)))
    print("Sum: " + str(calculate_sum(data)))
    print("Average: " + str(calculate_average(data)))
end
```

### Exercise 3: String Utilities

```uddin
fun count_words(text):
    // Count number of words in text
end

fun capitalize_words(text):
    // Capitalize first letter of each word
end

fun remove_spaces(text):
    // Remove all spaces from text
end

fun main():
    sentence = "hello world programming"
    
    print("Original: " + sentence)
    print("Word count: " + str(count_words(sentence)))
    print("Capitalized: " + capitalize_words(sentence))
    print("No spaces: " + remove_spaces(sentence))
end
```

## The Main Function

The `main()` function has special significance in Uddin-Lang - it serves as the entry point for your program.

### How Main Function Works

When you run a Uddin-Lang program, the interpreter:

1. **First** executes all top-level statements (variable declarations, function definitions)
2. **Then** automatically calls the `main()` function if it exists

```uddin
// This executes first
print("Program starting...")

// Function definitions are processed
fun setup():
    print("Setting up...")
end

fun cleanup():
    print("Cleaning up...")
end

// This executes last (automatically)
fun main():
    print("Main function executing")
    setup()
    // Your main program logic here
    cleanup()
end
```

**Output:**
```
Program starting...
Main function executing
Setting up...
Cleaning up...
```

### When to Use Main Function

**✅ Use `main()` for:**
- Applications with clear entry points
- Programs that need organized structure
- Code that will be tested (functions can be tested individually)
- Larger programs with multiple functions

**❌ Skip `main()` for:**
- Simple scripts that execute immediately
- Library files meant to be imported
- Configuration files
- Quick calculations or data processing

### Main Function Best Practices

```uddin
// ✅ Good: Organized main function
fun initialize_app():
    print("Initializing application...")
end

fun process_data():
    // Process your data here
    return "Data processed successfully"
end

fun cleanup_resources():
    print("Cleaning up resources...")
end

fun main():
    initialize_app()
    
    result = process_data()
    print(result)
    
    cleanup_resources()
end
```

```uddin
// ❌ Avoid: Everything in main
fun main():
    // Don't put all your code directly in main
    print("Initializing application...")
    
    // 100+ lines of code here...
    
    print("Cleaning up resources...")
end
```

### Programs Without Main Function

You can also write programs that execute immediately without a `main()` function:

```uddin
// immediate_script.din
print("This script runs immediately!")

name = "Uddin-Lang"
version = "1.0"

print("Language: " + name)
print("Version: " + version)

// Functions can still be defined
fun helper_function():
    return "Helper result"
end

// And called immediately
result = helper_function()
print("Result: " + result)
```

This style is perfect for:
- Quick scripts and automation
- Library files that initialize themselves
- Configuration files
- Data processing scripts

## Summary

Functions are fundamental building blocks that help you:

- **Organize code** into logical, reusable pieces
- **Reduce repetition** by writing code once and using it multiple times
- **Improve readability** by giving meaningful names to code blocks
- **Enable testing** by isolating functionality
- **Support recursion** for elegant solutions to complex problems

Mastering functions is essential for writing clean, maintainable, and efficient Uddin-Lang programs.

---

**Next**: Learn about [Arrays and Objects](arrays-and-objects.md) to work with complex data structures.