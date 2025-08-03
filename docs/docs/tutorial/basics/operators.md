---
sidebar_position: 4
---

# Operators

Uddin-Lang provides a comprehensive set of operators for mathematical calculations, logical operations, and comparisons.

## Arithmetic Operators

### Basic Arithmetic

```uddin
fun main():
    a = 10
    b = 3
    
    // Basic operations
    sum = a + b        // Addition: 13
    diff = a - b       // Subtraction: 7
    product = a * b    // Multiplication: 30
    quotient = a / b   // Division: 3.333...
    remainder = a % b  // Modulo: 1
    
    print("Sum: " + str(sum))
    print("Difference: " + str(diff))
    print("Product: " + str(product))
    print("Quotient: " + str(quotient))
    print("Remainder: " + str(remainder))
end
```

### Assignment Operators

Uddin-Lang supports compound assignment operators:

```uddin
fun main():
    x = 10
    
    x += 5    // x = x + 5, result: 15
    x -= 3    // x = x - 3, result: 12
    x *= 2    // x = x * 2, result: 24
    x /= 4    // x = x / 4, result: 6
    x %= 5    // x = x % 5, result: 1
    
    print("Final value: " + str(x))
end
```

## Comparison Operators

```uddin
fun main():
    a = 10
    b = 20
    
    // Comparison operations
    equal = (a == b)        // false
    not_equal = (a != b)    // true
    less_than = (a < b)     // true
    greater_than = (a > b)  // false
    less_equal = (a <= b)   // true
    greater_equal = (a >= b) // false
    
    // Containment operator
    numbers = [1, 2, 3, 4, 5]
    contains_three = 3 in numbers  // true
    contains_ten = 10 in numbers   // false
    
    print("Equal: " + str(equal))
    print("Not equal: " + str(not_equal))
    print("Contains 3: " + str(contains_three))
    print("Contains 10: " + str(contains_ten))
end
```

## Logical Operators

```uddin
fun main():
    x = true
    y = false
    
    // Logical operations
    and_result = x and y     // false
    or_result = x or y       // true
    not_result = not x       // false
    xor_result = x xor y     // true (exclusive or)
    
    print("AND: " + str(and_result))
    print("OR: " + str(or_result))
    print("NOT: " + str(not_result))
    print("XOR: " + str(xor_result))
end
```

### XOR (Exclusive OR) Operator

The `xor` operator returns `true` when exactly one operand is `true`, but not both:

```uddin
fun main():
    // XOR truth table
    print("XOR Truth Table:")
    print("false xor false =", false xor false)  // false
    print("false xor true  =", false xor true)   // true
    print("true  xor false =", true xor false)   // true
    print("true  xor true  =", true xor true)    // false
    
    // Practical example: toggle switch
    light_on = false
    switch_pressed = true
    light_on = light_on xor switch_pressed  // toggles light
    print("Light is now: " + (light_on ? "ON" : "OFF"))
end
```

## String Operations

```uddin
fun main():
    first_name = "John"
    last_name = "Doe"
    
    // String concatenation
    full_name = first_name + " " + last_name
    print("Full name: " + full_name)
    
    // String comparison
    same_name = (first_name == "John")  // true
    print("Same name: " + str(same_name))
    
    // String containment
    contains_oh = "oh" in first_name  // true
    print("Contains 'oh': " + str(contains_oh))
end
```

## Operator Precedence

Operators are evaluated in the following order (highest to lowest precedence):

1. **Parentheses**: `()`, **Array access**: `[]`, **Property access**: `.`
2. **Unary operators**: `not`, `-` (negation)
3. **Multiplication and Division**: `*`, `/`, `%`
4. **Addition and Subtraction**: `+`, `-`
5. **Comparison**: `<`, `<=`, `>`, `>=`, `in`
6. **Equality**: `==`, `!=`
7. **Ternary**: `? :`
8. **Logical AND**: `and`
9. **Logical XOR**: `xor`
10. **Logical OR**: `or`
11. **Assignment**: `=`, `+=`, `-=`, `*=`, `/=`, `%=`

### Precedence Examples

```uddin
fun main():
    // Without parentheses
    result1 = 2 + 3 * 4     // 14 (not 20)
    
    // With parentheses
    result2 = (2 + 3) * 4   // 20
    
    // Complex expression
    result3 = 10 + 5 * 2 - 3 / 3  // 10 + 10 - 1 = 19
    
    print("Result 1: " + str(result1))
    print("Result 2: " + str(result2))
    print("Result 3: " + str(result3))
end
```

## Ternary Operator

Uddin-Lang supports the ternary conditional operator:

```uddin
fun main():
    age = 18
    
    // Ternary operator: condition ? value_if_true : value_if_false
    status = (age >= 18) ? "Adult" : "Minor"
    print("Status: " + status)
    
    // Nested ternary
    grade = 85
    letter = (grade >= 90) ? "A" : (grade >= 80) ? "B" : "C"
    print("Grade: " + letter)
end
```

## Type Conversion in Operations

```uddin
fun main():
    // Automatic type conversion
    number = 42
    text = "The answer is: " + str(number)
    print(text)
    
    // Mixed number operations
    integer = 10
    decimal = 3.5
    result = integer + decimal  // 13.5
    print("Mixed result: " + str(result))
end
```

## Best Practices

### 1. Use Parentheses for Clarity

```uddin
// Good: Clear intention
result = (a + b) * (c - d)

// Avoid: Relies on precedence knowledge
result = a + b * c - d
```

### 2. Consistent Spacing

```uddin
// Good: Consistent spacing
x = a + b
y = c * d

// Avoid: Inconsistent spacing
x=a+b
y = c*d
```

### 3. Meaningful Variable Names

```uddin
// Good: Descriptive names
total_price = base_price + tax

// Avoid: Unclear names
t = p + x
```

## Practice Exercises

### Exercise 1: Calculator
Create a simple calculator that performs basic operations:

```uddin
fun main():
    a = 15
    b = 4
    
    // Implement all basic operations
    sum = a + b
    diff = a - b
    product = a * b
    quotient = a / b
    remainder = a % b
    
    print("Sum: " + str(sum))
    print("Difference: " + str(diff))
    print("Product: " + str(product))
    print("Quotient: " + str(quotient))
    print("Remainder: " + str(remainder))
end
```

### Exercise 2: Grade Calculator
Calculate letter grades based on numeric scores:

```uddin
fun main():
    score = 87
    
    // Use ternary operators to assign letter grades
    // A: 90+, B: 80-89, C: 70-79, D: 60-69, F: <60
    letter_grade = (score >= 90) ? "A" : 
                   (score >= 80) ? "B" : 
                   (score >= 70) ? "C" : 
                   (score >= 60) ? "D" : "F"
    
    print("Score: " + str(score))
    print("Grade: " + letter_grade)
end
```

### Exercise 3: Logical Conditions
Create a program that checks multiple conditions:

```uddin
fun main():
    age = 25
    has_license = true
    has_car = false
    
    // Determine if person can drive
    // Must be 18+, have license, and have car
    can_drive = (age >= 18) and has_license and has_car
    print("Can drive: " + str(can_drive))
end
```

## Summary

Operators in Uddin-Lang provide powerful ways to:
- Perform mathematical calculations
- Compare values and make decisions
- Combine logical conditions
- Manipulate strings and data

Understanding operator precedence and using parentheses for clarity will help you write more reliable and readable code.

---

**Next**: Learn about [Control Flow](control-flow.md) to make your programs more dynamic and responsive.