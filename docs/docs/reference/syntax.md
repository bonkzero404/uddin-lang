---
sidebar_position: 2
---

# Language Reference

Complete syntax reference for Uddin-Lang programming language.

## Basic Syntax

### Comments

```uddin
// Single-line comment

// Multi-line comments are not supported
// Use multiple single-line comments instead
```

### Variables

```uddin
// Variable declaration and assignment
name = "John"
age = 25
pi = 3.14159
isActive = true

// Variables are dynamically typed
value = 42        // Number
value = "hello"   // Now it's a string
value = [1, 2, 3] // Now it's an array
```

## Data Types

### Numbers

```uddin
// Integers
count = 42
negative = -17

// Floating-point numbers
pi = 3.14159
temperature = -273.15

// Scientific notation
large = 1.23e6    // 1,230,000
small = 4.56e-3   // 0.00456

// Digit separators for better readability
big_number = 1_000_000        // One million
phone = 555_123_4567          // Phone number format
pi_precise = 3.141_592_653    // Pi with separators
avogadro = 6.022_140_76e23    // Scientific notation with separators
planck = 6.626_070_15e-34     // Negative exponent with separators
mass_electron = 9.109_383_7e-31  // More reasonable scientific notation
```

#### Digit Separator Rules

- Use underscore (`_`) to separate digits for better readability
- Can only be placed between digits
- Cannot appear at the beginning or end of a number
- Cannot appear consecutively (`1__000` is invalid)
- Cannot appear immediately before or after decimal point
- Can be used in both integer and fractional parts
- Can be used in scientific notation exponents

```uddin
// Valid examples
valid1 = 1_000_000
valid2 = 3.141_592_653
valid3 = 1.23_45e6_789

// Invalid examples (will cause parse errors)
// invalid1 = _1000      // Cannot start with separator
// invalid2 = 1000_      // Cannot end with separator
// invalid3 = 1__000     // Cannot have consecutive separators
// invalid4 = 1_.5       // Cannot appear before decimal point
```

### Strings

```uddin
// String literals
name = "John Doe"
message = 'Hello, World!'  // Single quotes also work

// String concatenation
fullName = firstName + " " + lastName
greeting = "Hello, " + name + "!"

// Escape sequences
text = "Line 1\nLine 2\tTabbed"
quote = "He said, \"Hello!\""

// Multi-line strings
longText = "This is a very long string " +
           "that spans multiple lines " +
           "for better readability"
```

### Booleans

```uddin
isTrue = true
isFalse = false

// Boolean operations
result = true and false   // false
result = true or false    // true
result = not true         // false
```

### Null

```uddin
value = null

// Checking for null
if value == null then
    print("Value is null")
end

if value != null then
    print("Value is not null")
end
```

### Arrays

```uddin
// Array literals
numbers = [1, 2, 3, 4, 5]
fruits = ["apple", "banana", "orange"]
mixed = [1, "hello", true, null]

// Empty array
emptyArray = []

// Nested arrays
matrix = [[1, 2], [3, 4], [5, 6]]

// Array access
firstNumber = numbers[0]    // 1
lastFruit = fruits[2]       // "orange"

// Array modification
numbers[0] = 10
fruits = append(fruits, "grape")
```

### Objects (Dictionaries)

```uddin
// Object literals
person = {
    "name": "John Doe",
    "age": 30,
    "email": "john@example.com"
}

// Nested objects
user = {
    "profile": {
        "name": "Alice",
        "settings": {
            "theme": "dark",
            "notifications": true
        }
    }
}

// Object access
name = person["name"]
email = person["email"]
theme = user["profile"]["settings"]["theme"]

// Object modification
person["age"] = 31
person["phone"] = "+1234567890"
```

## Operators

### Arithmetic Operators

```uddin
a = 10
b = 3

sum = a + b        // 13
difference = a - b // 7
product = a * b    // 30
quotient = a / b   // 3.333...
remainder = a % b  // 1
power = a ** b     // 1000 (10^3)
```

### Comparison Operators

```uddin
a = 10
b = 5

equal = a == b           // false
notEqual = a != b        // true
greater = a > b          // true
greaterEqual = a >= b    // true
less = a < b             // false
lessEqual = a <= b       // false
```

### Logical Operators

```uddin
a = true
b = false

andResult = a and b      // false
orResult = a or b        // true
notResult = not a        // false

// Short-circuit evaluation
result = false and someFunction()  // someFunction() is not called
result = true or someFunction()    // someFunction() is not called
```

### Assignment Operators

```uddin
a = 10

a += 5    // a = a + 5  (15)
a -= 3    // a = a - 3  (12)
a *= 2    // a = a * 2  (24)
a /= 4    // a = a / 4  (6)
a %= 5    // a = a % 5  (1)
```

## Control Flow

### If Statements

```uddin
// Basic if statement
if (condition) then:
    // code block
end

// If-else statement
if (age >= 18) then:
    print("Adult")
else:
    print("Minor")
end

// If-else if-else statement
if (score >= 90) then:
    grade = "A"
else if (score >= 80) then:
    grade = "B"
else if (score >= 70) then:
    grade = "C"
else if (score >= 60) then:
    grade = "D"
else:
    grade = "F"
end
```

### While Loops

```uddin
// Basic while loop
counter = 0
while (counter < 5):
    print(counter)
    counter = counter + 1
end

// While loop with break
while (true):
    input = getUserInput()
    if (input == "quit") then:
        break
    end
    processInput(input)
end

// While loop with continue
i = 0
while (i < 10):
    i = i + 1
    if (i % 2 == 0) then:
        continue  // Skip even numbers
    end
    print(i)
end
```

### For Loops

```uddin
// For-in loop with arrays
fruits = ["apple", "banana", "orange"]
for (fruit in fruits):
    print("I like " + fruit)
end

// For-in loop with strings
text = "hello"
for (char in text):
    print(char)
end

// For loop with break and continue
numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
for (num in numbers):
    if (num > 7) then:
        break  // Stop at 8
    end
    if (num % 2 == 0) then:
        continue  // Skip even numbers
    end
    print(num)  // Prints 1, 3, 5, 7
end
```

## Functions

### Function Declaration

```uddin
// Basic function
fun greet():
    print("Hello, World!")
end

// Function with parameters
fun greet(name):
    print("Hello, " + name + "!")
end

// Function with multiple parameters
fun add(a, b):
    return a + b
end

// Function with default behavior for missing parameters
fun greet(name):
    if (name == null) then:
        name = "World"
    end
    print("Hello, " + name + "!")
end
```

### Function Calls

```uddin
// Call function without parameters
greet()

// Call function with parameters
greet("Alice")
result = add(5, 3)

// Store function result
sum = add(10, 20)
print("Sum: " + str(sum))
```

### Return Statements

```uddin
// Function with return value
fun multiply(a, b):
    return a * b
end

// Function with multiple return points
fun divide(a, b):
    if (b == 0) then:
        return null  // Early return for error case
    end
    return a / b
end

// Function with conditional return
fun getGrade(score):
    if (score >= 90) then:
        return "A"
    else if (score >= 80) then:
        return "B"
    else if (score >= 70) then:
        return "C"
    else:
        return "F"
    end
end
```

### Recursive Functions

```uddin
// Factorial function
fun factorial(n):
    if (n <= 1) then:
        return 1
    else:
        return n * factorial(n - 1)
    end
end

// Fibonacci function
fun fibonacci(n):
    if (n <= 1) then:
        return n
    else:
        return fibonacci(n - 1) + fibonacci(n - 2)
    end
end
```

### Anonymous Functions (Lambda)

```uddin
// Anonymous function assigned to variable
square = fun(x):
    return x * x
end

// Anonymous function as parameter
numbers = [1, 2, 3, 4, 5]
squares = map(numbers, fun(x):
    return x * x
end)

// Anonymous function with multiple parameters
add = fun(a, b):
    return a + b
end
result = add(5, 3)
```

## Error Handling

### Try-Catch Blocks

```uddin
// Basic try-catch
try:
    result = riskyOperation()
    print("Success: " + str(result))
catch (error):
    print("Error occurred: " + str(error))
end

// Try-catch with specific error handling
try:
    data = read_file("config.txt")
    config = json_parse(data)
    return config
catch (fileError):
    print("Could not read file: " + str(fileError))
    return getDefaultConfig()
end

// Nested try-catch
try:
    data = fetchDataFromAPI()
    try:
        processedData = processData(data)
        return processedData
    catch (processingError):
        print("Processing failed: " + str(processingError))
        return null
    end
catch (networkError):
    print("Network error: " + str(networkError))
    return getCachedData()
end
```

## Imports and Modules

### Import Statements

```uddin
// Import entire module
import "math_utils"

// Import with alias
import "string_utils" as str
import "date_utils" as date

// Import from subdirectory
import "utils/file_utils" as files
import "models/user" as User
```

### Using Imported Functions

```uddin
import "math_utils" as math
import "string_utils" as str

// Use imported functions
result = math.sqrt(16)
text = str.capitalize("hello world")

// Access imported constants
pi = math.PI
maxLength = str.MAX_STRING_LENGTH
```

## Built-in Functions

### Type Checking

```uddin
// Check variable type
value = 42
if (typeof(value) == "number") then:
    print("It's a number")
end

// Type conversion
numberStr = str(42)        // "42"
number = int("42")         // 42
floatNum = float("3.14")   // 3.14
boolVal = bool("true")     // true
```

### Array Functions

```uddin
arr = [1, 2, 3, 4, 5]

// Array manipulation
arr = append(arr, 6)           // [1, 2, 3, 4, 5, 6]
length = len(arr)              // 6
first = arr[0]                 // 1
last = arr[len(arr) - 1]       // 6

// Array methods
filtered = filter(arr, fun(x):
    return x > 3
end)  // [4, 5, 6]
mapped = map(arr, fun(x):
    return x * 2
end)       // [2, 4, 6, 8, 10, 12]
sum = reduce(arr, fun(acc, x):
    return acc + x
end, 0)  // 21
```

### String Functions

```uddin
text = "Hello, World!"

// String methods
length = len(text)                    // 13
upper = toUpperCase(text)             // "HELLO, WORLD!"
lower = toLowerCase(text)             // "hello, world!"
trimmed = trim("  hello  ")           // "hello"

// String searching
index = indexOf(text, "World")        // 7
starts = startsWith(text, "Hello")    // true
ends = endsWith(text, "!")            // true

// String manipulation
replaced = replace(text, "World", "Uddin")  // "Hello, Uddin!"
parts = split(text, ", ")                   // ["Hello", "World!"]
joined = join(["Hello", "World"], ", ")     // "Hello, World"
```

## Regular Expressions

```uddin
// Pattern matching
pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
email = "user@example.com"
isValid = regex_match(pattern, email)  // true

// Find matches
text = "Call me at 123-456-7890 or 987-654-3210"
phonePattern = "\\d{3}-\\d{3}-\\d{4}"
firstPhone = regex_find(phonePattern, text)      // "123-456-7890"
allPhones = regex_find_all(phonePattern, text)   // ["123-456-7890", "987-654-3210"]

// Replace with regex
masked = regex_replace(phonePattern, text, "XXX-XXX-XXXX")
// "Call me at XXX-XXX-XXXX or XXX-XXX-XXXX"

// Split with regex
parts = regex_split("[,;\\s]+", "apple, banana; orange cherry")
// ["apple", "banana", "orange", "cherry"]
```

## File I/O

```uddin
// Reading files
content = read_file("data.txt")
if (content != null) then:
    print("File content: " + content)
else:
    print("Could not read file")
end

// Writing files
data = "Hello, World!"
write_file("output.txt", data)

// Appending to files
append_file("log.txt", "New log entry\n")

// File operations
if (file_exists("config.txt")) then:
    size = file_size("config.txt")
    print("Config file size: " + str(size) + " bytes")
end

// Delete file
file_delete("temp.txt")
```

## Date and Time

```uddin
// Current date and time
now = date_now()
print("Current time: " + str(now))

// Format dates
formatted = date_format(now, "2006-01-02 15:04:05")
print("Formatted: " + formatted)

// Parse dates
dateStr = "2023-12-25 10:30:00"
parsedDate = date_parse(dateStr, "2006-01-02 15:04:05")

// Date arithmetic
future = date_add(now, "24h")      // Add 24 hours
past = date_subtract(now, "1h30m") // Subtract 1 hour 30 minutes

// Date comparison
diff = date_diff(future, now)      // "24h0m0s"
```

## HTTP and Networking

```uddin
// HTTP GET request
response = http_get("https://api.example.com/data")
if (response != null) then:
    print("Response: " + response)
end

// HTTP POST request
data = "{\"name\": \"John\", \"age\": 30}"
response = http_post("https://api.example.com/users", data)

// Other HTTP methods
response = http_put(url, data)
response = http_delete(url)
response = http_patch(url, data)
```

## Syntax Rules

### Statement Termination
- Statements are terminated by newlines
- No semicolons required
- Multi-line statements can be continued with `+` operator

### Block Structure
- Blocks are delimited by keywords and `end`
- Indentation is recommended but not required
- Keywords: `if`/`then:`/`end`, `while`/`(`condition`):`/`end`, `for`/`in`/`:`/`end`, `fun`/`end`, `try:`/`catch`/`(`identifier`):`/`end`

### Identifiers
- Must start with letter or underscore
- Can contain letters, numbers, and underscores
- Case-sensitive
- Cannot be reserved keywords

### Reserved Keywords
```
and      break    catch    continue  else     end
false    for      fun      if        import   in
null     not      or       return    then     true
try      while    xor
```

### Operator Precedence (highest to lowest)
1. `**` (exponentiation)
2. `*`, `/`, `%` (multiplication, division, modulo)
3. `+`, `-` (addition, subtraction)
4. `==`, `!=`, `<`, `<=`, `>`, `>=` (comparison)
5. `not` (logical NOT)
6. `and` (logical AND)
7. `or` (logical OR)
8. `=`, `+=`, `-=`, `*=`, `/=`, `%=` (assignment)

This reference covers the complete syntax of Uddin-Lang. For practical examples and usage patterns, see the [Examples](../examples/index.md) section.