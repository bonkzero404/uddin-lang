---
sidebar_position: 6
---

# Error Handling and Debugging

Uddin-Lang menyediakan sistem error handling yang robust dengan try-catch blocks dan berbagai teknik debugging untuk membantu Anda mengidentifikasi dan memperbaiki masalah dalam kode.

## Try-Catch Error Handling

### Basic Try-Catch

```uddin
// Basic error handling
try:
    result = 10 / 0  // Division by zero
    print("Result: " + str(result))
catch (error):
    print("Error occurred: " + str(error))
end

print("Program continues...")
```

### Handling Different Types of Errors

```uddin
fun safeFileOperation(filename):
    import "fs"
    import "json"
    try:
        // Attempt to read file
        content = fs.read(filename)

        // Attempt to parse as JSON
        data = json.parse(content)

        // Attempt to access property
        name = data.name

        return {"success": true, "data": data}

    catch (error):
        error_message = str(error)

        if (contains(error_message, "file not found")) then:
            return {"success": false, "error": "File does not exist", "type": "file_error"}
        else if (contains(error_message, "invalid JSON")) then:
            return {"success": false, "error": "Invalid JSON format", "type": "parse_error"}
        else if (contains(error_message, "undefined property")) then:
            return {"success": false, "error": "Missing required property", "type": "data_error"}
        else:
            return {"success": false, "error": error_message, "type": "unknown_error"}
        end
    end
end

// Usage
result = safeFileOperation("config.json")
if (result.success) then:
    print("Data loaded successfully: " + str(result.data))
else:
    print("Error (" + result.type + "): " + result.error)
end
```

### Nested Try-Catch

```uddin
fun processUserData(user_id):
    try:
        // Outer try: database operations
        user_data = getUserFromDatabase(user_id)

        try:
            // Inner try: data validation
            validateUserData(user_data)

            try:
                // Innermost try: business logic
                processed_data = processBusinessLogic(user_data)
                return {"success": true, "data": processed_data}

            catch (business_error):
                print("Business logic error: " + str(business_error))
                return {"success": false, "error": "Processing failed", "stage": "business"}
            end

        catch (validation_error):
            print("Validation error: " + str(validation_error))
            return {"success": false, "error": "Invalid data", "stage": "validation"}
        end

    catch (db_error):
        print("Database error: " + str(db_error))
        return {"success": false, "error": "Database unavailable", "stage": "database"}
    end
end

fun getUserFromDatabase(user_id):
    // Simulate database operation
    if (user_id == "invalid") then:
        // Trigger a division by zero error to simulate database error
        x = 1 / 0
    end
    return {"id": user_id, "name": "John Doe", "email": "john@example.com"}
end

fun validateUserData(data):
    if (data.email == "" or len(data.email) == 0) then:
        // Trigger an array out of bounds error to simulate validation error
        arr = [1, 2, 3]
        x = arr[10]
    end
end

fun processBusinessLogic(data):
    if (data.name == "blocked_user") then:
        // Trigger a type error to simulate business logic error
        x = len(123)
    end
    return {"processed": true, "user": data}
end

fun main():
    print("=== Testing Nested Try-Catch Error Handling ===")
    print()

    // Test case 1: Normal successful flow
    print("Test 1: Valid user processing")
    result1 = processUserData("user123")
    print("Result: " + str(result1))
    print()

    // Test case 2: Database error
    print("Test 2: Database error")
    result2 = processUserData("invalid")
    print("Result: " + str(result2))
    print()

    // Test case 3: Validation error
    print("Test 3: Validation error (no email)")
    // We need to simulate validation error by modifying the flow
    try:
        user_data = {"id": "test", "name": "Test User", "email": ""}
        validateUserData(user_data)
    catch (validation_error):
        print("Validation error caught: " + str(validation_error))
        result3 = {"success": false, "error": "Invalid data", "stage": "validation"}
        print("Result: " + str(result3))
    end
    print()

    // Test case 4: Business logic error
    print("Test 4: Business logic error")
    try:
        user_data = {"id": "blocked", "name": "blocked_user", "email": "blocked@example.com"}
        validateUserData(user_data)
        processBusinessLogic(user_data)
    catch (business_error):
        print("Business logic error caught: " + str(business_error))
        result4 = {"success": false, "error": "Processing failed", "stage": "business"}
        print("Result: " + str(result4))
    end
    print()

    print("=== All tests completed ===")
end
```

## Error Handling Patterns

### Result Pattern

```uddin
fun createResult(success, data, error):
    result = {
        "success": success,
        "data": data,
        "error": error
    }

    result["isSuccess"] = fun():
        return result.success
    end

    result["isError"] = fun():
        return result.success == false
    end

    result["getData"] = fun():
        if (result.success) then:
            return result.data
        else:
            // Trigger an error to simulate throw
            x = 1 / 0
        end
    end

    result["getError"] = fun():
        if (result.success == false) then:
            return result.error
        else:
            return null
        end
    end

    return result
end

fun safeDivide(a, b):
    try:
        if (b == 0) then:
            return createResult(false, null, "Division by zero")
        end

        result = a / b
        return createResult(true, result, null)

    catch (error):
        return createResult(false, null, str(error))
    end
end

fun main():
    print("=== Testing Result Pattern with Error Handling ===")
    print()

    // Test case 1: Successful division
    print("Test 1: Successful division (10 / 2)")
    result1 = safeDivide(10, 2)
    if (result1.isSuccess()) then:
        print("Result: " + str(result1.getData()))
    else:
        print("Error: " + result1.getError())
    end
    print()

    // Test case 2: Division by zero
    print("Test 2: Division by zero (10 / 0)")
    result2 = safeDivide(10, 0)
    if (result2.isSuccess()) then:
        print("Result: " + str(result2.getData()))
    else:
        print("Error: " + result2.getError())
    end
    print()

    // Test case 3: Test error state
    print("Test 3: Testing isError() method")
    result3 = safeDivide(5, 0)
    print("Is error: " + str(result3.isError()))
    print("Is success: " + str(result3.isSuccess()))
    print()

    // Test case 4: Test getData on failed result (should trigger error)
    print("Test 4: Attempting to get data from failed result")
    result4 = safeDivide(8, 0)
    try:
        data = result4.getData()
        print("Unexpected success: " + str(data))
    catch (error):
        print("Expected error caught: " + str(error))
    end
    print()

    print("=== All tests completed ===")
end
```

### Error Accumulation

```uddin
fun validateInput(value, type):
    if (type == "email") then:
        // Simple email validation - check if contains @ and .
        return contains(value, "@") and contains(value, ".")
    end
    return true
end

fun trim(str):
    // Simple trim - remove leading/trailing spaces (simplified)
    return str
end

fun validateForm(form_data):
    errors = []

    // Validate each field
    if (form_data.name == "" or len(form_data.name) == 0) then:
        push(errors, "Name is required")
    end

    if (form_data.email == "" or validateInput(form_data.email, "email") == false) then:
        push(errors, "Valid email is required")
    end

    if (form_data.age < 18) then:
        push(errors, "Age must be 18 or older")
    end

    if (form_data.password == "" or len(form_data.password) < 8) then:
        push(errors, "Password must be at least 8 characters")
    end

    return {
        "valid": len(errors) == 0,
        "errors": errors,
        "data": form_data
    }
end

fun main():
    print("=== Form Validation Demo ===")
    print()

    // Test case 1: Invalid form
    print("Test 1: Invalid form data")
    form_data = {
        "name": "",
        "email": "invalid-email",
        "age": 16,
        "password": "123"
    }

    validation = validateForm(form_data)
    if (validation.valid) then:
        print("Form is valid!")
    else:
        print("Form has errors:")
        // Use traditional for loop instead of for...in
        i = 0
        while (i < len(validation.errors)):
            print("  - " + validation.errors[i])
            i = i + 1
        end
    end
    print()

    // Test case 2: Valid form
    print("Test 2: Valid form data")
    valid_form = {
        "name": "John Doe",
        "email": "john@example.com",
        "age": 25,
        "password": "securepassword123"
    }

    validation2 = validateForm(valid_form)
    if (validation2.valid) then:
        print("Form is valid!")
        print("User data: " + str(validation2.data))
    else:
        print("Form has errors:")
        i = 0
        while (i < len(validation2.errors)):
            print("  - " + validation2.errors[i])
            i = i + 1
        end
    end
    print()

    print("=== Demo completed ===")
end
```

## Debugging Techniques

### Debug Logging

```uddin
fun createLogger(level):
    import "datetime"
    log_levels = {"DEBUG": 0, "INFO": 1, "WARN": 2, "ERROR": 3}

    // Set default level if not provided
    if (level == null) then:
        level = "INFO"
    end

    current_level = log_levels[level]

    return {
        "level": level,

        "debug": fun(message):
            if (log_levels["DEBUG"] >= current_level) then:
                timestamp = datetime.now()
                print("[" + str(timestamp) + "] DEBUG: " + str(message))
            end
        end,

        "info": fun(message):
            if (log_levels["INFO"] >= current_level) then:
                timestamp = datetime.now()
                print("[" + str(timestamp) + "] INFO: " + str(message))
            end
        end,

        "warn": fun(message):
            if (log_levels["WARN"] >= current_level) then:
                timestamp = datetime.now()
                print("[" + str(timestamp) + "] WARN: " + str(message))
            end
        end,

        "error": fun(message):
            if (log_levels["ERROR"] >= current_level) then:
                timestamp = datetime.now()
                print("[" + str(timestamp) + "] ERROR: " + str(message))
            end
        end
    }
end

// Usage
logger = createLogger("DEBUG")

fun complexCalculation(x, y):
    logger.debug("Starting calculation with x=" + str(x) + ", y=" + str(y))

    try:
        if (x < 0 or y < 0) then:
            logger.warn("Negative values detected")
        end

        step1 = x * 2
        logger.debug("Step 1 result: " + str(step1))

        step2 = step1 + y
        logger.debug("Step 2 result: " + str(step2))

        if (step2 == 0) then:
            logger.error("Division by zero in final step")
            // Simulate error by dividing by zero, which will cause panic
            result = 100 / 0
        end

        result = 100 / step2
        logger.info("Calculation completed successfully: " + str(result))

        return result

    catch (error):
        logger.error("Calculation failed: " + str(error))
        return null
    end
end

// Test with debugging
result = complexCalculation(5, 10)
print("Final result: " + str(result))

// Test error case (division by zero)
print("\n=== Testing error case ===")
try:
    error_result = complexCalculation(-5, 10)
    print("Error result: " + str(error_result))
catch (error):
    print("Caught error: " + str(error))
end

// Test different log levels
print("\n=== Testing different log levels ===")
info_logger = createLogger("INFO")
warn_logger = createLogger("WARN")
error_logger = createLogger("ERROR")

print("\n--- DEBUG Logger (should show all) ---")
logger.debug("This is a debug message")
logger.info("This is an info message")
logger.warn("This is a warning message")
logger.error("This is an error message")

print("\n--- INFO Logger (should show INFO, WARN, ERROR) ---")
info_logger.debug("This debug won't show")
info_logger.info("This info will show")
info_logger.warn("This warning will show")
info_logger.error("This error will show")

print("\n--- WARN Logger (should show WARN, ERROR only) ---")
warn_logger.debug("This debug won't show")
warn_logger.info("This info won't show")
warn_logger.warn("This warning will show")
warn_logger.error("This error will show")

print("\n--- ERROR Logger (should show ERROR only) ---")
error_logger.debug("This debug won't show")
error_logger.info("This info won't show")
error_logger.warn("This warning won't show")
error_logger.error("This error will show")
```

### Variable Inspection

```uddin
fun inspect(variable, name):
    print("=== Inspecting " + name + " ===")
    var_type = typeof(variable)
    print("Type: " + var_type)
    print("Value: " + str(variable))

    if (var_type == "array") then:
        print("Length: " + str(len(variable)))
        if (len(variable) > 0) then:
            print("First element: " + str(variable[0]))
            print("Last element: " + str(variable[len(variable) - 1]))
            print("All elements:")
            i = 0
            while (i < len(variable)):
                print("  [" + str(i) + "] " + str(variable[i]))
                i = i + 1
            end
        end
    end

    if (var_type == "object") then:
        print("Object properties:")
        print("  name: " + str(variable.name))
        print("  age: " + str(variable.age))
        print("  hobbies: " + str(variable.hobbies))
    end

    if (var_type == "string") then:
        print("Length: " + str(len(variable)))
        if (len(variable) > 0) then:
            print("First character: " + variable[0])
            print("Last character: " + variable[len(variable) - 1])
        end
    end

    if (var_type == "integer" or var_type == "float") then:
        print("Numeric operations:")
        print("  * 2 = " + str(variable * 2))
        print("  + 10 = " + str(variable + 10))
        print("  Is positive: " + str(variable > 0))
    end

    if (var_type == "boolean") then:
        print("Boolean operations:")
        print("  NOT " + str(variable) + " = " + str(variable == false))
        print("  AND true = " + str(variable and true))
        print("  OR false = " + str(variable or false))
    end

    print("========================")
end

// Usage
print("=== Advanced Variable Inspection Demo ===")
print()

data = {"name": "John Doe", "age": 30, "hobbies": ["reading", "coding", "gaming"]}
inspect(data, "user_data")

inspect(data.hobbies, "hobbies_array")

inspect(data.name, "user_name")

inspect(42, "number")

inspect(-15.5, "negative_float")

inspect(true, "boolean_true")

inspect(false, "boolean_false")

// Test with empty array
empty_array = []
inspect(empty_array, "empty_array")

// Test with complex nested structure
complex_data = {
    "name": "Complex Object",
    "age": 25,
    "hobbies": ["nested", "array", "inside", "object"]
}
inspect(complex_data, "complex_data")

print("=== Demo completed ===")
```

### Stack Trace Simulation

```uddin
// Global call stack
call_stack = []

// Helper functions for stack management
fun enterFunction(function_name, args):
    import "datetime"
    if (args == null) then:
        args = []
    end

    entry = {
        "function": function_name,
        "args": args,
        "timestamp": datetime.now()
    }
    append(call_stack, entry)
end

fun exitFunction(function_name):
    if (len(call_stack) > 0) then:
        last_entry = call_stack[len(call_stack) - 1]
        if (last_entry.function == function_name) then:
            call_stack = slice(call_stack, 0, len(call_stack) - 1)
        end
    end
end

fun printCallStack():
    print("=== Call Stack ===")
    i = len(call_stack) - 1
    while (i >= 0):
        entry = call_stack[i]
        depth = len(call_stack) - 1 - i
        indent = repeat("  ", depth)
        print(indent + "at " + entry.function + "(" + str(entry.args) + ")")
        i = i - 1
    end
    print("=================")
end

fun clearCallStack():
    // Remove all elements from call_stack using pop
    while (len(call_stack) > 0):
        pop(call_stack)
    end
end

fun functionA(x):
    enterFunction("functionA", [x])

    try:
        print("In functionA with x=" + str(x))
        result = functionB(x * 2)
        exitFunction("functionA")
        return result
    catch (error):
        printCallStack()
        exitFunction("functionA")
        // Re-throw by causing another error
        return 1 / 0  // This will cause a division by zero error
    end
end

fun functionB(y):
    enterFunction("functionB", [y])

    try:
        print("In functionB with y=" + str(y))
        result = functionC(y + 5)
        exitFunction("functionB")
        return result
    catch (error):
        printCallStack()
        exitFunction("functionB")
        // Re-throw by causing another error
        return 1 / 0  // This will cause a division by zero error
    end
end

fun functionC(z):
    enterFunction("functionC", [z])

    try:
        print("In functionC with z=" + str(z))

        if (z > 20) then:
            printCallStack()
            exitFunction("functionC")
            // Simulate error by division by zero
            return 1 / 0
        end

        result = z * 3
        exitFunction("functionC")
        return result

    catch (error):
        printCallStack()
        exitFunction("functionC")
        // Re-throw by causing another error
        return 1 / 0  // This will cause a division by zero error
    end
end

// Test with error
print("=== Stack Trace Simulation Demo ===")
print()

print("Test 1: Normal execution")
clearCallStack()  // Clear any previous state
try:
    result = functionA(5)  // Normal case: 5 * 2 = 10, 10 + 5 = 15 (< 20)
    print("Result: " + str(result))
catch (error):
    print("Caught error: " + str(error))
end

print()
print("Test 2: Error case with stack trace")
clearCallStack()  // Clear stack for clean test
try:
    result = functionA(8)  // Error case: 8 * 2 = 16, 16 + 5 = 21 (> 20)
    print("Result: " + str(result))
catch (error):
    print("Final error caught: " + str(error))
end

print()
print("Test 3: Another error case")
clearCallStack()
try:
    result = functionA(10)  // Error case: 10 * 2 = 20, 20 + 5 = 25 (> 20)
    print("Result: " + str(result))
catch (error):
    print("Final error caught: " + str(error))
end

print()
print("=== Demo completed ===")
```

## Best Practices

### Error Handling Guidelines

1. **Fail Fast**: Detect errors early and handle them immediately
2. **Specific Errors**: Provide specific error messages for different failure modes
3. **Graceful Degradation**: Provide fallback functionality when possible
4. **Resource Cleanup**: Always clean up resources in finally blocks
5. **Error Logging**: Log errors with sufficient context for debugging

### Debugging Guidelines

1. **Consistent Logging**: Use consistent log formats and levels
2. **Meaningful Messages**: Include relevant context in debug messages
3. **Tool Integration**: Integrate with external debugging tools when available

Dengan sistem error handling dan debugging yang komprehensif ini, Anda dapat membangun aplikasi Uddin-Lang yang robust dan mudah di-maintain.
