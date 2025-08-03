---
sidebar_position: 3
---

# Variables and Data Types

Learn how to work with different types of data in Uddin Programming Language.

## Learning Objectives

By the end of this chapter, you will understand:
- How to create and use variables
- Different data types available in Uddin-Lang
- Type conversion and checking
- Working with dynamic typing
- Best practices for variable usage

## Variables in Uddin-Lang

Uddin-Lang uses **dynamic typing**, meaning you don't need to declare variable types explicitly. The type is determined by the value you assign.

### Variable Creation

```uddin
fun main():
    // Variables are created when first assigned
    name = "Alice"        // String
    age = 25             // Number (integer)
    height = 5.6         // Number (float)
    is_student = true    // Boolean

    print("Name: " + name)
    print("Age: " + str(age))
    print("Height: " + str(height))
    print("Student: " + str(is_student))
end
```

### Variable Reassignment

```uddin
fun main():
    // Variables can change type when reassigned
    value = 42           // Number
    print("Number: " + str(value))

    value = "Hello"      // Now a string
    print("String: " + value)

    value = true         // Now a boolean
    print("Boolean: " + str(value))

    value = [1, 2, 3]    // Now an array
    print("Array: " + str(value))
end
```

## Data Types

### Numbers

Uddin-Lang supports both integers and floating-point numbers:

```uddin
fun main():
    // Integers
    count = 42
    negative = -17
    zero = 0

    // Floating-point numbers
    price = 19.99
    temperature = -5.5
    pi = 3.14159

    // Large numbers
    population = 1000000
    distance = 9.461e15  // Scientific notation

    // Digit separators for better readability
    big_number = 1_000_000        // One million
    phone = 555_123_4567          // Phone number format
    precise_pi = 3.141_592_653    // Pi with separators
    avogadro = 6.022_140_76e23    // Scientific notation with separators

    print("Count: " + str(count))
    print("Price: $" + str(price))
    print("Temperature: " + str(temperature) + "°C")
    print("Big number: " + str(big_number))
    print("Phone: " + str(phone))
end
```

#### Number Operations

```uddin
fun main():
    a = 10
    b = 3

    // Basic arithmetic
    sum = a + b          // 13
    difference = a - b   // 7
    product = a * b      // 30
    quotient = a / b     // 3.333...
    remainder = a % b    // 1
    power = a ** b       // 1000

    print("Sum: " + str(sum))
    print("Difference: " + str(difference))
    print("Product: " + str(product))
    print("Quotient: " + str(quotient))
    print("Remainder: " + str(remainder))
    print("Power: " + str(power))
end
```

#### Digit Separators

For better readability of large numbers, Uddin-Lang supports underscore (`_`) as digit separators:

```uddin
fun main():
    // Using digit separators for readability
    million = 1_000_000
    billion = 1_000_000_000

    // Works with decimals too
    precise_value = 123_456.789_012

    // Scientific notation with separators
    large_scientific = 1.234_567e23
    small_scientific = 9.876_543e-12

    // Practical examples
    salary = 75_000          // $75,000
    phone_number = 555_123_4567
    credit_card = 1234_5678_9012_3456

    print("Million: " + str(million))
    print("Salary: $" + str(salary))
    print("Phone: " + str(phone_number))
end
```

**Important Rules:**
- Separators can only be placed between digits
- Cannot start or end with a separator
- Cannot have consecutive separators (`1__000` is invalid)
- Cannot appear before or after decimal point

### Strings

Strings represent text data:

```uddin
fun main():
    // String creation
    name = "Alice"
    message = "Hello, World!"
    empty = ""

    // String with quotes
    quote = "She said, \"Hello!\""

    // Multi-word strings
    full_name = "Alice Johnson"
    address = "123 Main Street, Jakarta"

    print(name)
    print(message)
    print(quote)
end
```

#### String Operations

```uddin
fun main():
    first_name = "Alice"
    last_name = "Johnson"

    // String concatenation
    full_name = first_name + " " + last_name
    greeting = "Hello, " + first_name + "!"

    // String length
    name_length = len(full_name)

    // String case conversion
    uppercase = upper(full_name)
    lowercase = lower(full_name)

    print("Full name: " + full_name)
    print("Length: " + str(name_length))
    print("Uppercase: " + uppercase)
    print("Lowercase: " + lowercase)
end
```

#### String Formatting

```uddin
fun main():
    name = "Alice"
    age = 25
    city = "Jakarta"

    // Building formatted strings
    info = "Name: " + name + ", Age: " + str(age) + ", City: " + city

    // Template-like formatting
    template = "Hello, my name is " + name + ". I am " + str(age) + " years old."

    print(info)
    print(template)
end
```

### Booleans

Booleans represent true/false values:

```uddin
fun main():
    // Boolean literals
    is_active = true
    is_complete = false

    // Boolean from comparisons
    age = 25
    is_adult = age >= 18        // true
    is_senior = age >= 65       // false

    // Boolean from functions
    text = "Hello"
    is_empty = len(text) == 0   // false

    print("Active: " + str(is_active))
    print("Adult: " + str(is_adult))
    print("Senior: " + str(is_senior))
    print("Empty: " + str(is_empty))
end
```

#### Boolean Operations

```uddin
fun main():
    age = 25
    has_license = true
    has_car = false

    // Logical AND
    can_drive = age >= 18 and has_license

    // Logical OR
    can_travel = has_car or has_license

    // Logical NOT
    needs_license = not has_license

    print("Can drive: " + str(can_drive))
    print("Can travel: " + str(can_travel))
    print("Needs license: " + str(needs_license))
end
```

### Arrays

Arrays store ordered collections of items:

```uddin
fun main():
    // Array creation
    numbers = [1, 2, 3, 4, 5]
    names = ["Alice", "Bob", "Charlie"]
    mixed = [1, "hello", true, 3.14]
    empty = []

    // Accessing elements (0-based indexing)
    first_number = numbers[0]    // 1
    last_name = names[2]         // "Charlie"

    print("First number: " + str(first_number))
    print("Last name: " + last_name)
    print("Mixed array: " + str(mixed))
end
```

#### Array Operations

```uddin
fun main():
    fruits = ["apple", "banana", "orange"]

    // Array length
    count = len(fruits)          // 3

    // Adding elements
    push(fruits, "grape")

    // Array after adding
    new_count = len(fruits)      // 4

    print("Original count: " + str(count))
    print("New count: " + str(new_count))
    print("Fruits: " + str(fruits))
end
```

#### Working with Array Elements

```uddin
fun main():
    scores = [85, 92, 78, 96, 88]

    // Accessing elements
    first_score = scores[0]
    last_score = scores[4]

    // Iterating through array
    total = 0
    for (score in scores):
        total = total + score
    end

    average = total / len(scores)

    print("First score: " + str(first_score))
    print("Last score: " + str(last_score))
    print("Average: " + str(average))
end
```

### Objects (Maps)

Objects store key-value pairs:

```uddin
fun main():
    // Object creation
    person = {
        "name": "Alice",
        "age": 25,
        "city": "Jakarta",
        "is_student": true
    }

    // Accessing properties
    name = person["name"]
    age = person["age"]

    print("Name: " + name)
    print("Age: " + str(age))
end
```

#### Object Operations

```uddin
fun main():
    user = {
        "username": "alice123",
        "email": "alice@example.com",
        "active": true
    }

    // Adding new properties
    user["last_login"] = "2024-01-15"
    user["login_count"] = 42

    // Updating existing properties
    user["active"] = false

    // Accessing properties
    username = user["username"]
    email = user["email"]
    last_login = user["last_login"]

    print("Username: " + username)
    print("Email: " + email)
    print("Last login: " + last_login)
    print("Active: " + str(user["active"]))
end
```

#### Nested Objects

```uddin
fun main():
    employee = {
        "name": "Alice Johnson",
        "position": "Developer",
        "contact": {
            "email": "alice@company.com",
            "phone": "+62-123-456-7890"
        },
        "skills": ["Python", "JavaScript", "Uddin-Lang"]
    }

    // Accessing nested data
    name = employee["name"]
    email = employee["contact"]["email"]
    first_skill = employee["skills"][0]

    print("Employee: " + name)
    print("Email: " + email)
    print("First skill: " + first_skill)
end
```

### Null Values

Null represents the absence of a value:

```uddin
fun main():
    // Null assignment
    empty_value = null

    // Checking for null
    if (empty_value == null) then:
        print("Value is null")
    end

    // Default values for null
    name = null
    display_name = name != null ? name : "Unknown"

    print("Display name: " + display_name)
end
```

## Type Conversion

### Converting to String

```uddin
fun main():
    // Numbers to string
    age = 25
    price = 19.99

    age_str = str(age)        // "25"
    price_str = str(price)    // "19.99"

    // Boolean to string
    is_active = true
    status_str = str(is_active)  // "true"

    // Array to string
    numbers = [1, 2, 3]
    numbers_str = str(numbers)   // "[1, 2, 3]"

    print("Age: " + age_str)
    print("Price: $" + price_str)
    print("Status: " + status_str)
end
```

### Converting Numbers

```uddin
fun main():
    // String to number (conceptual - actual implementation may vary)
    age_str = "25"
    price_str = "19.99"

    // In practice, you might need specific parsing functions
    // age = int(age_str)     // Convert to integer
    // price = float(price_str) // Convert to float

    // For now, we work with direct assignment
    age = 25
    price = 19.99

    print("Age: " + str(age))
    print("Price: " + str(price))
end
```

## Type Checking

### Checking Types

```uddin
fun main():
    value = "Hello"

    // Check if value is a string by trying string operations
    if (len(value) >= 0) then:  // Strings have length
        print("Value is likely a string")
    end

    number = 42

    // Check if value behaves like a number
    result = number + 0  // Numbers can be added
    print("Number operation result: " + str(result))
end
```

### Type-Safe Operations

```uddin
fun main():
    // Safe string operations
    fun safe_string_length(value):
        // Try to get length, return 0 if not possible
        if (value == null) then:
            return 0
        end
        return len(str(value))  // Convert to string first
    end

    // Test with different types
    text = "Hello"
    number = 123
    empty = null

    print("Text length: " + str(safe_string_length(text)))
    print("Number length: " + str(safe_string_length(number)))
    print("Null length: " + str(safe_string_length(empty)))
end
```

## Variable Scope

### Local Variables

```uddin
fun calculate_area(length, width):
    // Local variables - only exist within this function
    area = length * width
    message = "Area calculated"
    return area
end

fun main():
    // These are local to main function
    room_length = 10
    room_width = 8

    room_area = calculate_area(room_length, room_width)

    print("Room area: " + str(room_area))

    // Cannot access 'area' or 'message' from calculate_area here
end
```

### Global Variables

```uddin
// Global variables (defined outside functions)
APP_NAME = "My Application"
VERSION = "1.0.0"

fun show_info():
    // Can access global variables
    print("App: " + APP_NAME)
    print("Version: " + VERSION)
end

fun main():
    print("Starting " + APP_NAME)
    show_info()
end
```

## Best Practices

### Meaningful Variable Names

```uddin
fun main():
    // Good: Descriptive names
    customer_name = "Alice Johnson"
    order_total = 299.99
    is_premium_customer = true
    items_in_cart = 3

    // Bad: Unclear names
    // n = "Alice Johnson"     // What is 'n'?
    // t = 299.99             // What is 't'?
    // flag = true            // What does this flag mean?
    // x = 3                  // What does 'x' represent?

    print("Customer: " + customer_name)
    print("Total: $" + str(order_total))
end
```

### Consistent Naming

```uddin
fun main():
    // Use consistent naming patterns
    user_name = "alice"
    user_email = "alice@example.com"
    user_age = 25
    user_city = "Jakarta"

    // Group related data in objects
    user = {
        "name": "alice",
        "email": "alice@example.com",
        "age": 25,
        "city": "Jakarta"
    }

    print("User: " + user["name"])
end
```

### Initialize Variables

```uddin
fun main():
    // Initialize variables with appropriate default values
    count = 0
    total = 0.0
    message = ""
    items = []
    user_data = {}
    is_ready = false

    // Use meaningful defaults
    status = "pending"
    priority = "normal"
    max_attempts = 3

    print("Initialized with defaults")
end
```

## Exercises

### Exercise 1: Variable Types

Create variables of different types and display their information:

```uddin
fun main():
    // TODO: Create variables of each type
    // - A string with your name
    // - A number with your age
    // - A boolean indicating if you're a student
    // - An array with your hobbies
    // - An object with your contact information

    // TODO: Display all the information
end
```

**Solution**:
```uddin
fun main():
    // Different data types
    name = "Alice Johnson"
    age = 25
    is_student = false
    hobbies = ["reading", "coding", "hiking"]
    contact = {
        "email": "alice@example.com",
        "phone": "+62-123-456-7890",
        "city": "Jakarta"
    }

    // Display information
    print("=== Personal Information ===")
    print("Name: " + name)
    print("Age: " + str(age))
    print("Student: " + str(is_student))
    print("Hobbies: " + str(hobbies))
    print("Email: " + contact["email"])
    print("Phone: " + contact["phone"])
    print("City: " + contact["city"])
end
```

### Exercise 2: Type Conversion

Practice converting between different types:

```uddin
fun main():
    // TODO: Convert these values and display results
    number = 42
    decimal = 3.14159
    flag = true
    items = ["apple", "banana", "orange"]

    // Convert to strings and create messages
end
```

**Solution**:
```uddin
fun main():
    number = 42
    decimal = 3.14159
    flag = true
    items = ["apple", "banana", "orange"]

    // Convert to strings
    number_str = str(number)
    decimal_str = str(decimal)
    flag_str = str(flag)
    items_str = str(items)

    // Create formatted messages
    print("Number as string: '" + number_str + "'")
    print("Decimal as string: '" + decimal_str + "'")
    print("Boolean as string: '" + flag_str + "'")
    print("Array as string: '" + items_str + "'")

    // Use in sentences
    print("The answer is " + number_str)
    print("Pi is approximately " + decimal_str)
    print("The statement is " + flag_str)
end
```

### Exercise 3: Complex Data Structures

Create a complex data structure representing a library:

```uddin
fun main():
    // TODO: Create a library object with:
    // - Library name
    // - Address
    // - Books (array of book objects)
    // - Each book should have: title, author, year, available

    // TODO: Display library information
    // TODO: Show available books
end
```

**Solution**:
```uddin
fun main():
    library = {
        "name": "Jakarta Central Library",
        "address": "Jl. Sudirman No. 123, Jakarta",
        "books": [
            {
                "title": "The Art of Programming",
                "author": "Jane Smith",
                "year": 2020,
                "available": true
            },
            {
                "title": "Data Structures Guide",
                "author": "John Doe",
                "year": 2019,
                "available": false
            },
            {
                "title": "Modern Web Development",
                "author": "Alice Johnson",
                "year": 2021,
                "available": true
            }
        ]
    }

    // Display library information
    print("=== " + library["name"] + " ===")
    print("Address: " + library["address"])
    print("Total books: " + str(len(library["books"])))
    print("")

    // Show available books
    print("Available Books:")
    for (book in library["books"]):
        if (book["available"]) then:
            print("- " + book["title"] + " by " + book["author"] + " (" + str(book["year"]) + ")")
        end
    end
end
```

## Summary

In this chapter, you learned:

✅ **Dynamic Typing**: Variables don't need type declarations
✅ **Basic Types**: Numbers, strings, booleans, arrays, objects, null
✅ **Type Operations**: Converting between types and type-safe operations
✅ **Variable Scope**: Local vs global variables
✅ **Best Practices**: Meaningful names, initialization, consistency

### Key Takeaways

1. **Flexibility**: Dynamic typing makes code flexible but requires careful handling
2. **Type Safety**: Always validate data before operations
3. **Clarity**: Use descriptive variable names and consistent patterns
4. **Structure**: Organize related data using objects and arrays

---

**Next**: [Operators →](operators.md)
