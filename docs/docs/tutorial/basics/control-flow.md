---
sidebar_position: 5
---

# Control Flow

Control flow statements allow you to control the execution path of your program based on conditions and loops.

## Conditional Statements

### If Statements

The basic `if` statement executes code when a condition is true:

```uddin
fun main():
    age = 18

    if (age >= 18) then:
        print("You are an adult")
    end
end
```

### If-Else Statements

Use `else` to provide an alternative when the condition is false:

```uddin
fun main():
    temperature = 25

    if (temperature > 30) then:
        print("It's hot outside")
    else:
        print("It's not too hot")
    end
end
```

### If-Else If-Else Chain

Chain multiple conditions using `else if`:

```uddin
fun main():
    score = 85

    if (score >= 90) then:
        print("Grade: A")
    else if (score >= 80) then:
        print("Grade: B")
    else if (score >= 70) then:
        print("Grade: C")
    else if (score >= 60) then:
        print("Grade: D")
    else:
        print("Grade: F")
    end
end
```

### Using elif (Alias for else if)

You can use `elif` as a shorter alternative to `else if`:

```uddin
fun main():
    score = 85

    if (score >= 90) then:
        print("Grade: A")
    elif (score >= 80) then:
        print("Grade: B")
    elif (score >= 70) then:
        print("Grade: C")
    elif (score >= 60) then:
        print("Grade: D")
    else:
        print("Grade: F")
    end
end
```

:::tip
`elif` is an alias for `else if` and provides more concise syntax. Both forms are equivalent and can be used interchangeably.
:::

### Nested Conditions

```uddin
fun main():
    age = 25
    has_license = true

    if (age >= 18) then:
        if (has_license) then:
            print("You can drive")
        else:
            print("You need a license to drive")
        end
    else:
        print("You are too young to drive")
    end
end
```

## Ternary Operator

The ternary operator provides a concise way to write conditional expressions. It follows the format: `condition ? true_value : false_value`.

### Basic Ternary Usage

```uddin
fun main():
    age = 20
    status = age >= 18 ? "adult" : "minor"
    print("Status: " + status)  // Output: Status: adult
end
```

### Nested Ternary Operators

You can chain ternary operators for multiple conditions:

```uddin
fun classify_number(n):
    // Determine sign using nested ternary
    sign = n > 0 ? "positive" : n < 0 ? "negative" : "zero"

    // Determine parity
    parity = n % 2 == 0 ? "even" : "odd"

    return {
        "number": n,
        "sign": sign,
        "parity": parity,
        "absolute": n >= 0 ? n : -n
    }
end

fun main():
    result = classify_number(-5)
    print("Number: " + str(result["number"]))     // Number: -5
    print("Sign: " + result["sign"])             // Sign: negative
    print("Parity: " + result["parity"])         // Parity: odd
    print("Absolute: " + str(result["absolute"])) // Absolute: 5
end
```

### Ternary in Function Returns

```uddin
fun get_discount(age, is_student):
    // Calculate discount based on multiple conditions
    student_discount = is_student ? 0.1 : 0.0
    age_discount = age >= 65 ? 0.15 : age <= 12 ? 0.2 : 0.0

    // Return the maximum discount
    return student_discount > age_discount ? student_discount : age_discount
end

fun main():
    discount1 = get_discount(25, true)   // 0.1 (student discount)
    discount2 = get_discount(70, false)  // 0.15 (senior discount)
    discount3 = get_discount(8, false)   // 0.2 (child discount)

    print("Student discount: " + str(discount1 * 100) + "%")
    print("Senior discount: " + str(discount2 * 100) + "%")
    print("Child discount: " + str(discount3 * 100) + "%")
end
```

### Ternary with String Formatting

```uddin
fun format_time(hours, minutes, use_24h):
    // Convert to 12-hour format if needed
    formatted_hours = use_24h ? hours :
                      hours == 0 ? 12 :
                      hours > 12 ? hours - 12 : hours

    // Add AM/PM for 12-hour format
    period = use_24h ? "" : hours >= 12 ? " PM" : " AM"

    // Ensure two-digit formatting
    h_str = formatted_hours < 10 ? "0" + str(formatted_hours) : str(formatted_hours)
    m_str = minutes < 10 ? "0" + str(minutes) : str(minutes)

    return h_str +":" + m_str + period
end

fun main():
    print(format_time(9, 30, true))   // 09:30
    print(format_time(9, 30, false))  // 09:30 AM
    print(format_time(15, 45, false)) // 03:45 PM
    print(format_time(0, 15, false))  // 12:15 AM
end
```

### Ternary in Array Operations

```uddin
fun main():
    scores = [85, 92, 78, 96, 88]

    // Assign grades using ternary
    grades = []
    for (score in scores):
        grade = score >= 90 ? "A" : score >= 80 ? "B" : score >= 70 ? "C" : "F"
        append(grades, grade)
    end

    print("Scores: " + str(scores))
    print("Grades: " + str(grades))

    // Pass/fail results
    results = []
    for (score in scores):
        result = score >= 70 ? "PASS" : "FAIL"
        append(results, result)
    end

    print("Results: " + str(results))
end
```

### Mathematical Expressions with Ternary

```uddin
fun main():
    x = 10
    y = 7

    print("x = " + str(x) + ", y = " + str(y))
    print("max(x, y) = " + str(x > y ? x : y))           // 10
    print("min(x, y) = " + str(x < y ? x : y))           // 7
    print("|x - y| = " + str(x > y ? x - y : y - x))     // 3
    print("sign(x - y) = " + str(x > y ? 1 : x < y ? -1 : 0)) // 1
end
```

### Complex Ternary Expressions

```uddin
fun calculate_shipping(weight, is_priority, distance):
    // Calculate base rate based on weight
    base_rate = weight <= 1 ? 5.0 : weight <= 5 ? 8.0 : 12.0

    // Apply priority and distance multipliers
    priority_multiplier = is_priority ? 1.5 : 1.0
    distance_rate = distance > 1000 ? 1.2 : distance > 500 ? 1.1 : 1.0

    total = base_rate * priority_multiplier * distance_rate

    return {
        "cost": round(total, 2),
        "delivery_time": is_priority ?
                        distance > 1000 ? "2-3 days" : "1-2 days" :
                        distance > 1000 ? "5-7 days" : "3-5 days"
    }
end

fun main():
    shipping1 = calculate_shipping(0.5, false, 200)  // Light, standard, short
    shipping2 = calculate_shipping(8.0, true, 1500)  // Heavy, priority, long

    print("Standard shipping: $" + str(shipping1["cost"]) + " (" + shipping1["delivery_time"] + ")")
    print("Priority shipping: $" + str(shipping2["cost"]) + " (" + shipping2["delivery_time"] + ")")
end
```

### When to Use Ternary vs If-Else

**Use ternary when:**
- You need a simple conditional assignment
- The expression is short and readable
- You're working with simple true/false conditions

```uddin
// Good use of ternary
status = age >= 18 ? "adult" : "minor"
message = is_logged_in ? "Welcome back!" : "Please log in"
max_value = a > b ? a : b
```

**Use if-else when:**
- You have complex logic or multiple statements
- The condition involves multiple operations
- Readability would be compromised

```uddin
// Better as if-else
if (user_score >= passing_grade) then:
    print("Congratulations! You passed!")
    update_user_status("passed")
    send_certificate(user)
else:
    print("Sorry, you didn't pass this time.")
    schedule_retake(user)
    send_study_materials(user)
end
```

### Best Practices for Ternary Operators

1. **Keep it simple**: Don't nest too many ternary operators
2. **Use parentheses**: For clarity in complex expressions
3. **Consider readability**: If it's hard to read, use if-else instead
4. **Format properly**: Break long ternary expressions across lines

```uddin
// Good: Simple and clear
result = condition ? value1 : value2

// Good: Properly formatted nested ternary
grade = score >= 90 ? "A" :
        score >= 80 ? "B" :
        score >= 70 ? "C" : "F"

// Avoid: Too complex
result = (a > b && c < d) ? (x > y ? complex_calc1(a, b) : complex_calc2(c, d)) : (z > w ? another_calc(x) : default_value)
```

## Loops

### While Loops

Execute code repeatedly while a condition is true:

```uddin
fun main():
    count = 1

    while (count <= 5):
        print("Count: " + str(count))
        count = count + 1
    end

    print("Loop finished")
end
```

### For Loops with Arrays

Iterate through arrays using for loops:

```uddin
fun main():
    numbers = [1, 2, 3, 4, 5]

    for (num in numbers):
        print("Number: " + str(num))
    end
end
```

### For Loops with Range

```uddin
fun main():
    // Print numbers 1 to 10
    i = 1
    while (i <= 10):
        print("Number: " + str(i))
        i = i + 1
    end
end
```

## Loop Control

### Break Statement

Exit a loop early:

```uddin
fun main():
    count = 1

    while (count <= 10):
        if (count == 5) then:
            print("Breaking at 5")
            break
        end

        print("Count: " + str(count))
        count = count + 1
    end
end
```

### Continue Statement

Skip the current iteration:

```uddin
fun main():
    count = 0

    while (count < 10):
        count = count + 1

        if (count % 2 == 0) then:
            continue  // Skip even numbers
        end

        print("Odd number: " + str(count))
    end
end
```

## Practical Examples

### Example 1: Number Guessing Game

```uddin
fun main():
    secret_number = 7
    guess = 0
    attempts = 0
    max_attempts = 3

    print("Guess the number between 1 and 10!")

    while (attempts < max_attempts):
        // In a real program, you would get input from user
        // For demo, we'll simulate guesses
        if (attempts == 0) then:
            guess = 5
        else if (attempts == 1) then:
            guess = 8
        else:
            guess = 7
        end

        attempts = attempts + 1
        print("Attempt " + str(attempts) + ": Guessing " + str(guess))

        if (guess == secret_number) then:
            print("Congratulations! You guessed it!")
            break
        else if (guess < secret_number) then:
            print("Too low!")
        else:
            print("Too high!")
        end
    end

    if (guess != secret_number) then:
        print("Sorry, you ran out of attempts. The number was " + str(secret_number))
    end
end
```

### Example 2: Array Processing

```uddin
fun main():
    numbers = [12, 5, 8, 3, 16, 9, 1, 14]

    // Find maximum number
    max_num = numbers[0]
    for (num in numbers):
        if (num > max_num) then:
            max_num = num
        end
    end
    print("Maximum number: " + str(max_num))

    // Count even numbers
    even_count = 0
    for (num in numbers):
        if (num % 2 == 0) then:
            even_count = even_count + 1
        end
    end
    print("Even numbers count: " + str(even_count))

    // Sum of all numbers
    total = 0
    for (num in numbers):
        total = total + num
    end
    print("Sum of all numbers: " + str(total))
end
```

### Example 3: Menu System

```uddin
fun main():
    running = true
    choice = 1  // Simulate user choices

    while (running):
        print("\n=== Menu ===")
        print("1. Say Hello")
        print("2. Calculate Square")
        print("3. Exit")
        print("Choice: " + str(choice))

        if (choice == 1) then:
            print("Hello, World!")
            choice = 2  // Simulate next choice
        else if (choice == 2) then:
            number = 5
            square = number * number
            print("Square of " + str(number) + " is " + str(square))
            choice = 3  // Simulate next choice
        else if (choice == 3) then:
            print("Goodbye!")
            running = false
        else:
            print("Invalid choice")
            choice = 3  // Force exit
        end
    end
end
```

## Best Practices

### 1. Keep Conditions Simple

```uddin
// Good: Simple, readable condition
if (age >= 18 and has_license) then:
    print("Can drive")
end

// Avoid: Complex nested conditions
if (age >= 18) then:
    if (has_license) then:
        if (has_car) then:
            if (has_insurance) then:
                print("Can drive")
            end
        end
    end
end
```

### 2. Use Meaningful Variable Names

```uddin
// Good: Descriptive names
is_valid_user = (age >= 18 and has_account)
if (is_valid_user) then:
    print("Access granted")
end

// Avoid: Unclear names
x = (a >= 18 and b)
if (x) then:
    print("Access granted")
end
```

### 3. Avoid Infinite Loops

```uddin
// Good: Clear exit condition
count = 0
while (count < 10):
    print(count)
    count = count + 1  // Don't forget to update!
end

// Dangerous: Potential infinite loop
count = 0
while (count < 10):
    print(count)
    // Missing: count = count + 1
end
```

### 4. Use Early Returns When Appropriate

```uddin
fun check_access(age, has_license):
    if (age < 18) then:
        return "Too young"
    end

    if (!has_license) then:
        return "No license"
    end

    return "Access granted"
end
```

## Practice Exercises

### Exercise 1: FizzBuzz
Write a program that prints numbers 1 to 20, but:
- Print "Fizz" for multiples of 3
- Print "Buzz" for multiples of 5
- Print "FizzBuzz" for multiples of both 3 and 5

### Exercise 2: Prime Number Checker
Create a function that checks if a number is prime:

```uddin
fun is_prime(n):
    // Your implementation here
    // Return true if n is prime, false otherwise
end

fun main():
    number = 17
    if (is_prime(number)) then:
        print(str(number) + " is prime")
    else:
        print(str(number) + " is not prime")
    end
end
```

### Exercise 3: Array Search
Implement a linear search function:

```uddin
fun find_element(array, target):
    // Your implementation here
    // Return the index if found, -1 if not found
end

fun main():
    numbers = [10, 25, 3, 8, 15]
    target = 8
    index = find_element(numbers, target)

    if (index != -1) then:
        print("Found " + str(target) + " at index " + str(index))
    else:
        print(str(target) + " not found")
    end
end
```

## Summary

Control flow statements are essential for creating dynamic programs:

- **Conditional statements** (`if`, `else`) make decisions
- **Loops** (`while`, `for`) repeat operations
- **Loop control** (`break`, `continue`) provides fine-grained control
- **Best practices** ensure readable and maintainable code

Mastering control flow allows you to build complex logic and create interactive programs.

---

**Next**: Learn about [Functions](functions.md) to organize and reuse your code effectively.
