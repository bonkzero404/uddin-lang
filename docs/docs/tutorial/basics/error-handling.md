---
sidebar_position: 6
---

# Error Handling

Learn how to handle errors gracefully in Uddin-Lang.

## Try-Catch Blocks

Uddin-Lang provides robust error handling through try-catch blocks:

```uddin
fun divide(a, b):
    try:
        if b == 0:
            throw "Division by zero error"
        end
        return a / b
    catch error:
        print("Error occurred: " + error)
        return null
    end
end

fun main():
    result = divide(10, 2)
    print("10 / 2 = " + str(result))

    result = divide(10, 0)
    print("Division by zero handled gracefully")
end
```

## Custom Error Types

You can create custom error types for better error handling:

```uddin
fun validateAge(age):
    if age < 0:
        throw "InvalidAgeError: Age cannot be negative"
    end
    if age > 150:
        throw "InvalidAgeError: Age seems unrealistic"
    end
    return true
end

fun processUser(name, age):
    try:
        validateAge(age)
        print("Processing user: " + name + " (age: " + str(age) + ")")
    catch error:
        if error.startsWith("InvalidAgeError"):
            print("Age validation failed: " + error)
        else:
            print("Unexpected error: " + error)
        end
    end
end

fun main():
    processUser("Alice", 25)
    processUser("Bob", -5)
    processUser("Charlie", 200)
end
```

## Error Propagation

Errors can be propagated up the call stack:

```uddin
fun riskyOperation():
    // Simulate a risky operation
    randomValue = random()
    if randomValue < 0.5:
        throw "OperationFailedError: Random failure occurred"
    end
    return "Success"
end

fun executeTask():
    try:
        result = riskyOperation()
        return result
    catch error:
        // Re-throw with additional context
        throw "TaskError: " + error
    end
end

fun main():
    try:
        result = executeTask()
        print("Task completed: " + result)
    catch error:
        print("Task failed: " + error)
    end
end
```

## Best Practices

### 1. Always Handle Expected Errors

```uddin
fun readFile(filename):
    try:
        // File reading logic here
        return fileContent
    catch error:
        if error.contains("FileNotFound"):
            print("File not found: " + filename)
            return ""
        else:
            throw error  // Re-throw unexpected errors
        end
    end
end
```

### 2. Use Specific Error Messages

```uddin
fun validateEmail(email):
    if not email.contains("@"):
        throw "ValidationError: Email must contain @ symbol"
    end
    if not email.contains("."):
        throw "ValidationError: Email must contain a domain"
    end
    return true
end
```

### 3. Clean Up Resources

```uddin
fun processData(filename):
    file = null
    try:
        file = openFile(filename)
        data = file.read()
        return processContent(data)
    catch error:
        print("Error processing file: " + error)
        return null
    finally:
        if file != null:
            file.close()
        end
    end
end
```

## Common Error Patterns

### Null Pointer Handling

```uddin
fun safeAccess(obj, key):
    try:
        return obj[key]
    catch error:
        if error.contains("KeyError"):
            return null
        else:
            throw error
        end
    end
end
```

### Type Validation

```uddin
fun ensureNumber(value):
    try:
        return toNumber(value)
    catch error:
        throw "TypeError: Expected number, got " + type(value)
    end
end
```

### Array Bounds Checking

```uddin
fun safeArrayAccess(arr, index):
    try:
        return arr[index]
    catch error:
        if error.contains("IndexError"):
            print("Index " + str(index) + " is out of bounds")
            return null
        else:
            throw error
        end
    end
end
```
