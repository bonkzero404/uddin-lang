---
sidebar_position: 7
---

# Best Practices

Follow these best practices to write clean, maintainable, and efficient Uddin-Lang code.

## Code Organization

### 1. Use Meaningful Names

```uddin
// ❌ Bad
fun calc(x, y):
    return x * y * 0.5
end

// ✅ Good
fun calculateTriangleArea(base, height):
    return base * height * 0.5
end
```

### 2. Keep Functions Small and Focused

```uddin
// ❌ Bad - doing too much
fun processUserData(userData):
    // Validate data
    if userData["email"] == null:
        throw "Email is required"
    end

    // Transform data
    userData["email"] = userData["email"].toLowerCase()

    // Save to database
    database.save(userData)

    // Send email
    emailService.send(userData["email"], "Welcome!")

    return userData
end

// ✅ Good - separated concerns
fun validateUserData(userData):
    if userData["email"] == null:
        throw "Email is required"
    end
    return true
end

fun normalizeUserData(userData):
    userData["email"] = userData["email"].toLowerCase()
    return userData
end

fun saveUser(userData):
    return database.save(userData)
end

fun sendWelcomeEmail(email):
    return emailService.send(email, "Welcome!")
end

fun processUserData(userData):
    validateUserData(userData)
    normalizedData = normalizeUserData(userData)
    savedUser = saveUser(normalizedData)
    sendWelcomeEmail(savedUser["email"])
    return savedUser
end
```

## Variable and Function Naming

### Use Descriptive Names

```uddin
// ❌ Bad
fun calc(p, r, t):
    return p * (1 + r) ** t
end

// ✅ Good
fun calculateCompoundInterest(principal, rate, time):
    return principal * (1 + rate) ** time
end
```

### Follow Naming Conventions

```uddin
// Variables and functions: camelCase
userAge = 25
isActive = true

fun getUserInfo():
    return userDatabase.fetch()
end

// Constants: UPPER_CASE
MAX_RETRY_ATTEMPTS = 3
DEFAULT_TIMEOUT = 30

// File names: kebab-case
// user-service.din
// email-validator.din
```

## Error Handling

### 1. Be Specific with Error Messages

```uddin
// ❌ Bad
fun validateAge(age):
    if age < 0:
        throw "Invalid age"
    end
end

// ✅ Good
fun validateAge(age):
    if age < 0:
        throw "ValidationError: Age cannot be negative, got: " + str(age)
    end
    if age > 150:
        throw "ValidationError: Age seems unrealistic, got: " + str(age)
    end
end
```

### 2. Handle Errors at the Right Level

```uddin
// ✅ Good - Handle at the appropriate level
fun readConfigFile(filename):
    try:
        return file.read(filename)
    catch error:
        if error.contains("FileNotFound"):
            // Provide default config instead of crashing
            return getDefaultConfig()
        else:
            throw error  // Let other errors bubble up
        end
    end
end
```

## Performance Considerations

### 1. Minimize Object Creation in Loops

```uddin
// ❌ Bad - creates objects inside loop
fun processItems(items):
    results = []
    for item in items:
        config = {"timeout": 30, "retries": 3}  // Created every iteration
        result = processItem(item, config)
        results.append(result)
    end
    return results
end

// ✅ Good - create once outside loop
fun processItems(items):
    config = {"timeout": 30, "retries": 3}  // Created once
    results = []
    for item in items:
        result = processItem(item, config)
        results.append(result)
    end
    return results
end
```

### 2. Use Early Returns

```uddin
// ❌ Bad - nested conditions
fun processUser(user):
    if user != null:
        if user["active"]:
            if user["verified"]:
                return doProcessing(user)
            else:
                return "User not verified"
            end
        else:
            return "User not active"
        end
    else:
        return "User is null"
    end
end

// ✅ Good - early returns
fun processUser(user):
    if user == null:
        return "User is null"
    end

    if not user["active"]:
        return "User not active"
    end

    if not user["verified"]:
        return "User not verified"
    end

    return doProcessing(user)
end
```

## Documentation and Comments

### 1. Write Self-Documenting Code

```uddin
// ❌ Bad - unclear what magic numbers mean
fun calculateDiscount(price, customerType):
    if customerType == 1:
        return price * 0.1
    elif customerType == 2:
        return price * 0.15
    else:
        return 0
    end
end

// ✅ Good - clear constants and logic
REGULAR_CUSTOMER = 1
PREMIUM_CUSTOMER = 2
REGULAR_DISCOUNT_RATE = 0.1
PREMIUM_DISCOUNT_RATE = 0.15

fun calculateDiscount(price, customerType):
    if customerType == REGULAR_CUSTOMER:
        return price * REGULAR_DISCOUNT_RATE
    elif customerType == PREMIUM_CUSTOMER:
        return price * PREMIUM_DISCOUNT_RATE
    else:
        return 0  // No discount for unknown customer types
    end
end
```

### 2. Comment Why, Not What

```uddin
// ❌ Bad - commenting the obvious
fun calculateTax(amount):
    // Multiply amount by 0.08
    return amount * 0.08
end

// ✅ Good - explaining business logic
fun calculateTax(amount):
    // Using 8% tax rate as per current tax regulation (2024)
    // This rate applies to all non-exempt items
    TAX_RATE = 0.08
    return amount * TAX_RATE
end
```

## Code Structure

### 1. Group Related Functions

```uddin
// User validation functions
fun validateEmail(email):
    return email.contains("@") and email.contains(".")
end

fun validateAge(age):
    return age >= 0 and age <= 150
end

fun validateUserData(userData):
    validateEmail(userData["email"])
    validateAge(userData["age"])
    return true
end

// User processing functions
fun saveUser(userData):
    return database.save(userData)
end

fun sendWelcomeEmail(user):
    return emailService.send(user["email"], "Welcome!")
end
```

### 2. Use Constants for Magic Numbers

```uddin
// Configuration constants
MAX_LOGIN_ATTEMPTS = 3
SESSION_TIMEOUT_MINUTES = 30
PASSWORD_MIN_LENGTH = 8

// Business logic constants
STANDARD_SHIPPING_COST = 5.99
EXPRESS_SHIPPING_COST = 12.99
FREE_SHIPPING_THRESHOLD = 50.00

fun calculateShippingCost(orderTotal, shippingType):
    if orderTotal >= FREE_SHIPPING_THRESHOLD:
        return 0
    end

    if shippingType == "express":
        return EXPRESS_SHIPPING_COST
    else:
        return STANDARD_SHIPPING_COST
    end
end
```

## Testing Guidelines

### 1. Write Testable Functions

```uddin
// ❌ Bad - hard to test due to dependencies
fun processOrder():
    order = database.getOrder()
    total = calculateTotal(order)
    database.updateOrder(order, total)
    emailService.sendConfirmation(order["email"])
end

// ✅ Good - pure function, easy to test
fun calculateOrderTotal(orderItems, taxRate, shippingCost):
    subtotal = 0
    for item in orderItems:
        subtotal = subtotal + (item["price"] * item["quantity"])
    end

    tax = subtotal * taxRate
    return subtotal + tax + shippingCost
end
```

### 2. Use Descriptive Test Names

```uddin
// Test function examples (pseudo-code)
fun testCalculateOrderTotal_WithMultipleItems_ReturnsCorrectTotal():
    // Test implementation
end

fun testValidateEmail_WithInvalidFormat_ThrowsValidationError():
    // Test implementation
end
```

## Security Best Practices

### 1. Validate All Input

```uddin
fun processUserInput(input):
    // Always validate input before processing
    if input == null or input.trim() == "":
        throw "ValidationError: Input cannot be empty"
    end

    if input.length() > MAX_INPUT_LENGTH:
        throw "ValidationError: Input too long"
    end

    // Sanitize input
    sanitizedInput = sanitize(input)
    return processCleanInput(sanitizedInput)
end
```

### 2. Don't Log Sensitive Information

```uddin
fun loginUser(username, password):
    // ❌ Bad - logging sensitive data
    print("Login attempt: " + username + " with password: " + password)

    // ✅ Good - log without sensitive data
    print("Login attempt for user: " + username)

    if authenticate(username, password):
        print("Login successful for user: " + username)
        return createSession(username)
    else:
        print("Login failed for user: " + username)
        throw "AuthenticationError: Invalid credentials"
    end
end
```
