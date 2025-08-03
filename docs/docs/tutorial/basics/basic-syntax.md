---
sidebar_position: 2
---

# Basic Syntax

Learn the fundamental syntax rules and constructs of Uddin Programming Language.

## Learning Objectives

By the end of this chapter, you will understand:
- How to write comments
- Basic program structure
- Statement and expression syntax
- Variable naming conventions
- Code organization principles

## Program Structure

Every Uddin program starts with a `main` function:

```uddin
fun main():
    print("Hello, World!")
end
```

### Key Elements

1. **Function Declaration**: `fun main():`
2. **Function Body**: Indented code block
3. **Function End**: `end` keyword
4. **Statements**: Individual instructions like `print()`

## Comments

Comments document your code and are ignored by the interpreter.

### Single-Line Comments

```uddin
// This is a single-line comment
print("Hello") // Comment at end of line
```

### Comment Best Practices

```uddin
fun main():
    // Good: Explains WHY, not just WHAT
    // Calculate compound interest using daily compounding
    rate = 0.05
    
    // Bad: Just repeats what code does
    // Set rate to 0.05
    // rate = 0.05
    
    // Good: Explains business logic
    // Apply 10% discount for premium customers
    if (customer_type == "premium") then:
        discount = 0.10
    end
end
```

## Statements and Expressions

### Statements
Statements perform actions and typically end with a newline:

```uddin
fun main():
    // Assignment statements
    name = "Alice"
    age = 25
    
    // Function call statements
    print("Hello")
    print(name)
    
    // Control flow statements
    if (age >= 18) then:
        print("Adult")
    end
end
```

### Expressions
Expressions evaluate to values and can be used within statements:

```uddin
fun main():
    // Arithmetic expressions
    result = 5 + 3 * 2    // Evaluates to 11
    
    // String expressions
    greeting = "Hello, " + "World!"  // Evaluates to "Hello, World!"
    
    // Boolean expressions
    is_adult = age >= 18  // Evaluates to true or false
    
    // Function call expressions
    length = len([1, 2, 3])  // Evaluates to 3
    
    // Complex expressions
    final_price = base_price * (1 + tax_rate) - discount
end
```

## Variables

### Variable Assignment

Variables are created when you first assign a value:

```uddin
// Variable assignment
name = "Alice"
age = 25
pi = 3.14159
is_student = true
```

### Variable Naming Rules

#### Valid Names
```uddin
fun main():
    // Letters and underscores
    user_name = "John"
    firstName = "Jane"
    _private_var = 42
    
    // Numbers (but not at start)
    value1 = 10
    item2 = "second"
    
    // Descriptive names
    total_price = 99.99
    is_valid = true
    MAX_ATTEMPTS = 3
end
```

#### Invalid Names
```uddin
fun main():
    // These will cause syntax errors:
    
    // 1name = "Invalid"        // Cannot start with number
    // user-name = "Invalid"    // Cannot contain hyphens
    // fun = "Invalid"          // Cannot use reserved keywords
    // if = "Invalid"           // Cannot use reserved keywords
    // user name = "Invalid"    // Cannot contain spaces
end
```

### Naming Conventions

```uddin
fun main():
    // Use snake_case for variables and functions
    user_name = "Alice"
    total_amount = 100.50
    
    // Use UPPER_CASE for constants
    MAX_RETRIES = 3
    DEFAULT_TIMEOUT = 30
    
    // Use descriptive names
    customer_email = "alice@example.com"  // Good
    // e = "alice@example.com"            // Bad: too short
    
    // Boolean variables should be questions
    is_valid = true
    has_permission = false
    can_edit = true
end
```

## Indentation and Code Structure

Uddin-Lang uses indentation to define code blocks:

### Proper Indentation

```uddin
fun main():
    // Function body is indented
    name = "Alice"
    
    if (name == "Alice") then:
        // If block is further indented
        print("Hello Alice!")
        age = 25
        
        if (age >= 18) then:
            // Nested blocks are indented even more
            print("You are an adult")
        end
    end
    
    // Back to function level indentation
    print("End of program")
end
```

### Indentation Best Practices

```uddin
fun main():
    // Use consistent indentation (4 spaces recommended)
    if (true) then:
        print("Properly indented")
    end
    
    // Group related code with blank lines
    first_name = "John"
    last_name = "Doe"
    full_name = first_name + " " + last_name
    
    // Separate logical sections
    age = 30
    birth_year = 2024 - age
    
    // Final output
    print("Name: " + full_name)
    print("Born in: " + str(birth_year))
end
```

## Reserved Keywords

These words have special meaning and cannot be used as variable names:

```uddin
// Reserved keywords in Uddin-Lang:
// fun, end, if, else, while, for, in, return
// true, false, null, and, or, not

fun main():
    // These are valid:
    function_name = "my_function"  // 'function' is OK
    end_time = "5:00 PM"          // 'end_time' is OK
    
    // These would be invalid:
    // fun = "invalid"            // 'fun' is reserved
    // if = "invalid"             // 'if' is reserved
    // return = "invalid"         // 'return' is reserved
end
```

## Line Continuation and Formatting

### Long Lines

```uddin
fun main():
    // Break long lines for readability
    very_long_variable_name = "This is a very long string that might " +
                             "extend beyond the comfortable reading width"
    
    // Function calls with many parameters
    result = some_function(
        parameter1,
        parameter2,
        parameter3
    )
    
    // Complex expressions
    final_calculation = (base_value * multiplier) +
                       (additional_value * factor) -
                       discount_amount
end
```

## Code Organization

### Function Organization

```uddin
// Helper functions first
fun calculate_tax(amount, rate):
    return amount * rate
end

fun format_currency(amount):
    return "$" + str(amount)
end

// Main function last
fun main():
    price = 100.00
    tax_rate = 0.08
    
    tax = calculate_tax(price, tax_rate)
    total = price + tax
    
    print("Price: " + format_currency(price))
    print("Tax: " + format_currency(tax))
    print("Total: " + format_currency(total))
end
```

### Logical Grouping

```uddin
fun main():
    // Input data
    customer_name = "Alice Johnson"
    order_items = ["laptop", "mouse", "keyboard"]
    
    // Calculations
    item_count = len(order_items)
    base_price = 1299.99
    tax_rate = 0.08
    
    // Processing
    tax_amount = base_price * tax_rate
    total_price = base_price + tax_amount
    
    // Output
    print("Customer: " + customer_name)
    print("Items: " + str(item_count))
    print("Total: $" + str(total_price))
end
```

## Common Syntax Patterns

### Variable Initialization

```uddin
fun main():
    // Initialize with default values
    count = 0
    total = 0.0
    message = ""
    items = []
    user_data = {}
    
    // Initialize from calculations
    current_year = 2024
    birth_year = 1990
    age = current_year - birth_year
end
```

### Conditional Assignment

```uddin
fun main():
    age = 17
    
    // Simple conditional assignment
    status = (age >= 18) ? "adult" : "minor"
    
    // Multiple conditions
    if (age < 13) then:
        category = "child"
    else if (age < 18) then:
        category = "teenager"
    else:
        category = "adult"
    end
end
```

## Exercises

### Exercise 1: Basic Program Structure

Create a program that displays your personal information:

```uddin
fun main():
    // TODO: Add your information
    // name = ?
    // age = ?
    // city = ?
    
    // TODO: Display the information
end
```

**Solution**:
```uddin
fun main():
    // Personal information
    name = "Alice Johnson"
    age = 25
    city = "Jakarta"
    occupation = "Software Developer"
    
    // Display information
    print("=== Personal Information ===")
    print("Name: " + name)
    print("Age: " + str(age))
    print("City: " + city)
    print("Occupation: " + occupation)
end
```

### Exercise 2: Variable Naming

Fix the variable names in this code:

```uddin
fun main():
    // Fix these variable names:
    // 1user = "John"           // Invalid: starts with number
    // user-name = "John"       // Invalid: contains hyphen
    // if = 25                  // Invalid: reserved keyword
    
    // TODO: Create valid variable names
end
```

**Solution**:
```uddin
fun main():
    // Fixed variable names:
    user1 = "John"              // Fixed: moved number to end
    user_name = "John"          // Fixed: used underscore
    user_age = 25               // Fixed: used descriptive name
    
    print("User: " + user_name)
    print("Age: " + str(user_age))
end
```

### Exercise 3: Code Organization

Organize this messy code:

```uddin
fun main():
name="Alice"
age=25
print("Hello")
city="Jakarta"
print(name)
print(age)
print(city)
end
```

**Solution**:
```uddin
fun main():
    // Personal data
    name = "Alice"
    age = 25
    city = "Jakarta"
    
    // Greetings and information display
    print("Hello!")
    print("Name: " + name)
    print("Age: " + str(age))
    print("City: " + city)
end
```

## Summary

In this chapter, you learned:

✅ **Program Structure**: Every program needs a `main()` function  
✅ **Comments**: Use `//` for documentation and explanations  
✅ **Variables**: Create with assignment, follow naming conventions  
✅ **Indentation**: Use consistent indentation for code blocks  
✅ **Organization**: Group related code and use descriptive names  
✅ **Best Practices**: Write clean, readable, and maintainable code  

### Key Takeaways

1. **Consistency is Key**: Use consistent naming and indentation
2. **Readability Matters**: Write code that others (and future you) can understand
3. **Comments Help**: Document the "why", not just the "what"
4. **Organization Counts**: Group related code logically

---

**Next**: [Variables and Data Types →](variables-and-data-types.md)