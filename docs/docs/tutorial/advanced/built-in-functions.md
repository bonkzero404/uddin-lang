---
sidebar_position: 1
---

# Built-in Functions

Uddin-Lang provides over 100 built-in functions that can be used directly without importing additional libraries. These functions cover various categories from string manipulation and mathematical operations to networking and data processing, making Uddin-Lang a powerful language for rapid development.

## Core System Functions

These fundamental functions provide essential system operations and program control.

### print(value)
Outputs a value to standard output. Accepts any data type and automatically converts it to a string representation.

```uddin
fun main():
    print("Hello, World!")  // String output
    print(42)               // Number output
    print([1, 2, 3])        // Array output
    print({"name": "John"}) // Object output
end
```

### typeof(value)
Returns the data type of a given value as a string. Useful for type checking and debugging.

```uddin
fun main():
    print(typeof(42))        // "number"
    print(typeof("hello"))   // "string"
    print(typeof([1, 2, 3])) // "array"
    print(typeof({"a": 1}))  // "object"
    print(typeof(true))      // "boolean"
end
```

### exit(code)
Terminates the program with a specified exit code. Use 0 for successful termination, non-zero for errors.

```uddin
fun main():
    error = check_conditions()
    if (error) then:
        print("An error occurred!")
        exit(1)  // Exit with error code
    end

    print("Program completed successfully")
    exit(0)  // Exit successfully
end
```

## Type Conversion Functions

These functions convert values between different data types safely and predictably.

### int(value)
Converts a value to an integer. Handles strings, floats, and booleans intelligently.

```uddin
fun main():
    print(int("123"))     // 123 - string to int
    print(int(45.67))     // 45 - float truncation
    print(int(true))      // 1 - boolean true
    print(int(false))     // 0 - boolean false
    print(int("12.34"))   // 12 - string float truncation
end
```

### float(value)
Converts a value to a floating-point number. Handles strings and integers seamlessly.

```uddin
fun main():
    print(float("123.45", 2)) // 123.45 - string to float
    print(float(42, 2))       // 42.0 - int to float
    print(float("3.14", 2))   // 3.14 - string decimal
end
```

### str(value)
Converts any value to its string representation. Essential for string concatenation and output formatting.

```uddin
fun main():
    print(str(123))        // "123" - number to string
    print(str(45.67))      // "45.67" - float to string
    print(str(true))       // "true" - boolean to string
    print(str([1, 2, 3]))  // "[1, 2, 3]" - array to string
end
```

### char(code)
Converts an ASCII code to its corresponding character. Useful for character manipulation.

```uddin
fun main():
    print(char(65))  // "A" - uppercase A
    print(char(97))  // "a" - lowercase a
    print(char(48))  // "0" - digit zero
end
```

### rune(char)
Converts a character to its ASCII code value. The inverse of the char() function.

```uddin
fun main():
    print(rune("A"))  // 65 - ASCII code for A
    print(rune("a"))  // 97 - ASCII code for a
    print(rune("0"))  // 48 - ASCII code for 0
end
```

## String Manipulation Functions

Uddin-Lang provides comprehensive string manipulation capabilities for text processing and data handling.

### Basic String Operations

Fundamental string operations for common text processing tasks.

```uddin
fun main():
    // join - combines array elements into a string
    words = ["Hello", "World"]
    print(join(words, " "))  // "Hello World"

    // split - divides string into array
    text = "apple,banana,orange"
    fruits = split(text, ",")
    print(fruits)  // ["apple", "banana", "orange"]

    // lower/upper - case conversion
    print(lower("HELLO"))  // "hello"
    print(upper("world"))  // "WORLD"

    // contains - checks for substring
    print(contains("hello world", "world"))  // true

    // find - locates substring position
    print(find("hello world", "world"))  // 6
end
```

### Advanced String Operations

Powerful string manipulation functions for complex text processing.

```uddin
fun main():
    // replace - substitutes substring
    text = "Hello World"
    print(replace(text, "World", "Uddin"))  // "Hello Uddin"

    // trim - removes whitespace
    print(trim("  hello  "))  // "hello"

    // starts_with/ends_with - prefix/suffix checking
    print(starts_with("hello", "he"))  // true
    print(ends_with("world", "ld"))    // true

    // repeat - duplicates string
    print(repeat("ha", 3))  // "hahaha"

    // reverse_str - reverses string
    print(reverse_str("hello"))  // "olleh"

    // substr - extracts substring
    print(substr("hello world", 6, 5))  // "world"

    // str_pad - adds padding
    print(str_pad("42", 5, "0"))  // "00042"
end
```

## Array and Collection Functions

### Basic Array Operations

Essential array manipulation functions for data processing and collection management.

```uddin
fun main():
    // len - gets array length
    arr = [1, 2, 3, 4, 5]
    print(len(arr))  // 5

    // append - adds element to array
    new_arr = append(arr, 6)
    print(new_arr)  // [1, 2, 3, 4, 5, 6]

    // slice - extracts array portion
    print(slice(arr, 1, 3))  // [2, 3]

    // sort - arranges elements in order
    numbers = [3, 1, 4, 1, 5]
    sort(numbers)
    print(numbers)  // [1, 1, 3, 4, 5]

    // reverse - reverses array order
    number_reversed = [1, 2, 3]
    reverse(number_reversed)
    print(number_reversed)  // [3, 2, 1]
end
```

### Advanced Array Operations

Powerful functional programming operations for complex data transformations.

```uddin
fun square(x):
    return x * x
end

fun is_even(x):
    return x % 2 == 0
end

fun add(acc, x):
    return acc + x
end

fun main():
    numbers = [1, 2, 3, 4, 5]

    // map - transforms each element
    squares = map(numbers, square)
    print(squares)  // [1, 4, 9, 16, 25]

    // filter - selects elements based on condition
    evens = filter(numbers, is_even)
    print(evens)  // [2, 4]

    // reduce - combines elements into single value
    sum = reduce(numbers, add, 0)
    print(sum)  // 15

    // push/pop - stack operations
    arr = [1, 2, 3]
    push(arr, 4)  // arr becomes [1, 2, 3, 4]
    value = pop(arr)  // value = 4, arr becomes [1, 2, 3]

    // shift/unshift - queue operations
    value = shift(arr)  // value = 1, arr becomes [2, 3]
    unshift(arr, 0)    // arr becomes [0, 2, 3]

    // index_of/last_index_of - finds element indices
    print(index_of([1, 2, 3, 2], 2))      // 1
    print(last_index_of([1, 2, 3, 2], 2)) // 3
end
```

### range - Array Generation

Generates arrays of sequential numbers with flexible start, end, and step parameters.

```uddin
fun main():
    // range(end) - generates from 0 to end-1
    print(range(5))  // [0, 1, 2, 3, 4]

    // range(start, end) - generates from start to end-1
    print(range(2, 7))  // [2, 3, 4, 5, 6]

    // range(start, end, step) - generates with custom step
    print(range(0, 10, 2))  // [0, 2, 4, 6, 8]
    print(range(10, 0, -2)) // [10, 8, 6, 4, 2]
end
```

## Data Structures

Uddin-Lang provides built-in data structures for specialized use cases beyond basic arrays and objects.

### Set (Unique Collection)

Sets store unique values and automatically handle duplicates, perfect for membership testing and deduplication.

```uddin
fun main():
    // Create new set
    my_set = set_new()

    // Add elements
    set_add(my_set, "apple")
    set_add(my_set, "banana")
    set_add(my_set, "apple")  // duplicate ignored

    // Check membership
    print(set_has(my_set, "apple"))   // true
    print(set_has(my_set, "orange"))  // false

    // Get set size
    print(set_size(my_set))  // 2

    // Convert to array
    array = set_to_array(my_set)
    print(array)  // ["apple", "banana"]

    // Remove element
    set_remove(my_set, "apple")
    print(set_size(my_set))  // 1
end
```

### Stack (LIFO - Last In, First Out)

Stacks follow the Last In, First Out principle, ideal for managing function calls, undo operations, and parsing.

```uddin
fun main():
    // Create new stack
    stack = stack_new()

    // Push elements
    stack_push(stack, "first")
    stack_push(stack, "second")
    stack_push(stack, "third")

    // Peek top element without removing
    print(stack_peek(stack))  // "third"

    // Pop element (removes and returns)
    value = stack_pop(stack)  // "third"
    print(stack_size(stack))  // 2

    // Check if empty
    print(stack_empty(stack)) // false
end
```

### Queue (FIFO - First In, First Out)

Queues follow the First In, First Out principle, perfect for task scheduling, breadth-first search, and buffering.

```uddin
fun main():
    // Create new queue
    queue = queue_new()

    // Enqueue elements
    queue_enqueue(queue, "first")
    queue_enqueue(queue, "second")
    queue_enqueue(queue, "third")

    // Peek front element without removing
    print(queue_front(queue))  // "first"

    // Dequeue element (removes and returns)
    value = queue_dequeue(queue)  // "first"
    print(queue_size(queue))      // 2

    // Check if empty
    print(queue_empty(queue))     // false
end
```

## File I/O Functions

Simple and efficient file operations for reading and writing data.

```uddin
fun main():
    // Read entire file content
    content = read_file("data.txt")
    print(content)

    // Write data to file
    data = "Hello, World!\nThis is Uddin-Lang!"
    write_file("output.txt", data)

    // Check if file exists
    if (file_exists("config.json")) then:
        config = read_file("config.json")
        print("Config loaded: " + config)
    else:
        print("Config file not found")
    end
end
```

## Best Practices

Follow these guidelines to write efficient and maintainable code with built-in functions:

### 1. Choose the Right Function
Select the most appropriate function for your specific use case:

```uddin
fun main():
    // Good: Use specific functions for clarity
    text = "Hello World"
    if (starts_with(text, "Hello")) then:
        print("Greeting detected")
    end

    // Avoid: Using generic functions when specific ones exist
    if (substr(text, 0, 5) == "Hello") then:
        print("Greeting detected")
    end
end
```

### 2. Handle Data Types Properly
Ensure you're using the correct data types for function parameters:

```uddin
fun main():
    // Good: Proper type handling
    user_input = "42"
    number = int(user_input)
    result = sqrt(float(number, 0))

    // Avoid: Assuming types without conversion
    // result = sqrt(user_input)  // This would cause an error
end
```

### 3. Use Error Handling
Implement proper error handling for operations that might fail:

```uddin
fun main():
    try:
        content = read_file("important_data.txt")
        data = json_parse(content)
        print("Data loaded successfully")
    catch (error):
        print("Error loading data: " + str(error))
        // Provide fallback behavior
        data = {"default": true}
    end
end
```

### 4. Consider Performance
For large datasets, choose efficient algorithms and functions:

```uddin
fun main():
    large_array = range(100000)

    // Good: Use built-in functions for better performance
    result = filter(large_array, is_even)

    // Avoid: Manual loops when built-ins are available
    // result = []
    // for item in large_array:
    //     if is_even(item):
    //         append(result, item)
    //     end
    // end
end
```

### 5. Combine Functions Effectively
Chain functions together for powerful data processing:

```uddin
fun process_text(text):
    return upper(trim(text))
end

fun main():
    raw_data = ["  hello  ", "  world  ", "  uddin  "]

    // Chain operations for clean, readable code
    processed = map(raw_data, process_text)
    result = join(processed, " ")
    print(result)  // "HELLO WORLD UDDIN"
end
```

By following these practices, you'll write more robust, efficient, and maintainable Uddin-Lang programs that leverage the full power of the built-in function library.

## Practical Examples

Here are real-world examples demonstrating how to combine multiple built-in functions effectively:

### Text Processing Pipeline

```uddin
fun is_not_empty(line):
    return len(trim(line)) > 0
end

fun to_lowercase(line):
    return lower(trim(line))
end

fun process_text_file(filename):
    try:
        content = read_file(filename)
        lines = split(content, "\n")

        // Filter out empty lines
        non_empty = filter(lines, is_not_empty)

        // Convert to lowercase and sort
        processed = map(non_empty, to_lowercase)
        sorted_lines = sort(processed)
        result = join(sorted_lines, "\n")

        output_filename = "processed_" + filename
        write_file(output_filename, result)
        print("File processed successfully: " + output_filename)

        return true
    catch (error):
        print("Error processing file: " + str(error))
        return false
    end
end

fun main():
    success = process_text_file("data.txt")
    if (success) then:
        print("Text processing completed")
    else:
        print("Text processing failed")
    end
end
```

### Data Analysis Example

```uddin
fun is_positive(num):
    return num > 0
end

fun square(num):
    return num * num
end

fun add(a, b):
    return a + b
end

fun analyze_numbers(data):
    // Filter positive numbers
    positives = filter(data, is_positive)

    // Calculate squares
    squares = map(positives, square)

    // Calculate statistics
    total = reduce(squares, add, 0)
    count = len(squares)
    average = total / count

    return {
        "count": count,
        "total": total,
        "average": average,
        "max": max(squares),
        "min": min(squares)
    }
end

fun main():
    numbers = [-2, 5, -1, 8, 3, -4, 7]
    stats = analyze_numbers(numbers)

    print("Analysis Results:")
    print("Positive numbers processed: " + str(stats["count"]))
    print("Sum of squares: " + str(stats["total"]))
    print("Average: " + str(stats["average"]))
end
```

This comprehensive guide covers the most frequently used built-in functions. For mathematical functions, networking capabilities, and other advanced features, refer to the subsequent chapters in this documentation.
