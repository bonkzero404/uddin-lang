---
sidebar_position: 8
---

# Best Practices and Style Guide

Comprehensive guide for developing robust, maintainable, and high-performance Uddin-Lang applications with modern programming patterns and practices.

## Code Style Guidelines

### Naming Conventions

#### Variables and Functions

```uddin
// ✓ Good: Use snake_case for consistency with built-in functions
user_name = "John Doe"
user_age = 25
calculate_total = fun(items):
    // Function implementation
end

// ✓ Good: Boolean variables with clear intent
is_authenticated = true
has_permission = false
can_edit_profile = true

// ✓ Good: Descriptive function names
validate_user_input = fun(data):
    // Validation logic
end

// ✗ Avoid: Inconsistent or unclear naming
userName = "John"        // Mixed case inconsistency
usr = "data"            // Cryptic abbreviations
auth = true             // Ambiguous meaning
calc = fun(): end       // Unclear function purpose

// ✓ Good: Consistent naming patterns
max_retry_attempts = 3
user_authentication_token = "abc123xyz"
default_page_size = 10

// ✗ Avoid: Cryptic abbreviations or unclear names
mra = 3                 // What does 'mra' mean?
uat = "abc123xyz"       // Unclear abbreviation
dps = 10               // Ambiguous meaning
```

#### Constants and Configuration

```uddin
// ✓ Good: Use UPPER_SNAKE_CASE for constants
MAX_CONNECTIONS = 100
API_BASE_URL = "https://api.example.com"
DEFAULT_TIMEOUT_MS = 5000
SESSION_DURATION = 86400  // 24 hours in seconds

// ✓ Good: Group related constants in objects
HTTP_STATUS = {
    "OK": 200,
    "CREATED": 201,
    "BAD_REQUEST": 400,
    "UNAUTHORIZED": 401,
    "NOT_FOUND": 404,
    "SERVER_ERROR": 500
}

UI_COLORS = {
    "PRIMARY": "#007bff",
    "SUCCESS": "#28a745",
    "WARNING": "#ffc107",
    "DANGER": "#dc3545",
    "LIGHT": "#f8f9fa",
    "DARK": "#343a40"
}

// ✓ Good: Mathematical and scientific constants
PI = 3.14159265359
GOLDEN_RATIO = 1.61803398875
AVOGADRO_NUMBER = 6.02214076e23
```

#### File and Module Naming

```uddin
// ✓ Good: Descriptive file names with clear purpose
// user_authentication.din
// data_validation_service.din
// email_notification_handler.din
// payment_processor.din
// inventory_management.din

// ✗ Avoid: Generic or unclear names
// utils.din
// helper.din
// stuff.din
// misc.din
// common.din
```

### Code Formatting and Structure

#### Indentation and Spacing

```uddin
// ✓ Good: Consistent 4-space indentation
fun process_user_registration(user_data):
    if (validate_user_input(user_data)) then:
        sanitized_data = sanitize_input(user_data)

        if (check_user_exists(sanitized_data.email)) then:
            return {
                "success": false,
                "error": "User already exists",
                "code": "USER_EXISTS"
            }
        end

        created_user = create_user_account(sanitized_data)
        send_welcome_email(created_user.email)

        return {
            "success": true,
            "user": created_user,
            "message": "Registration successful"
        }
    else:
        return {
            "success": false,
            "error": "Invalid user data",
            "code": "VALIDATION_FAILED"
        }
    end
end

// ✓ Good: Proper spacing around operators and expressions
result = (price * quantity) + (tax_rate * subtotal)
is_valid = (age >= 18) and (has_license == true) and (passed_test == true)
total_cost = base_price + shipping_fee + tax_amount

// ✗ Avoid: Inconsistent or cramped spacing
result=(price*quantity)+(tax_rate*subtotal)
is_valid=(age>=18)and(has_license==true)
total_cost=base_price+shipping_fee+tax_amount
```

#### Line Length and Breaking

```uddin
// ✓ Good: Break long lines for readability (80-100 characters max)
user_profile = {
    "personal_info": {
        "first_name": user.first_name,
        "last_name": user.last_name,
        "email": user.email_address,
        "phone": user.phone_number
    },
    "preferences": {
        "theme": user.preferred_theme,
        "language": user.language_setting,
        "notifications": user.notification_preferences
    },
    "metadata": {
        "created_at": user.registration_date,
        "last_login": user.last_login_timestamp,
        "status": user.account_status
    }
}

// ✓ Good: Break long function calls across multiple lines
search_results = perform_advanced_search(
    query_string,
    search_filters,
    sort_options,
    pagination_params,
    user_context
)

// ✗ Avoid: Extremely long lines that require horizontal scrolling
user_profile = {"personal_info": {"first_name": user.first_name, "last_name": user.last_name, "email": user.email_address, "phone": user.phone_number}, "preferences": {"theme": user.preferred_theme, "language": user.language_setting}}
```

#### Comments and Documentation

```uddin
// ✓ Good: Clear, informative comments that explain WHY, not WHAT
/**
 * Calculate compound interest using the formula: A = P(1 + r/n)^(nt)
 *
 * This function handles various compounding periods and includes validation
 * for edge cases such as zero principal or negative rates.
 *
 * @param principal - Initial amount invested (must be positive)
 * @param annual_rate - Annual interest rate as percentage (e.g., 5.25 for 5.25%)
 * @param compounds_per_year - Number of compounding periods per year
 * @param years - Investment duration in years
 * @returns Object with final amount, interest earned, and calculation details
 */
fun calculate_compound_interest(principal, annual_rate, compounds_per_year, years):
    // Input validation - prevent common calculation errors
    if (principal <= 0) then:
        panic("Principal must be positive")
    end

    if (annual_rate < 0) then:
        panic("Interest rate cannot be negative")
    end

    // Convert percentage to decimal (5.25% -> 0.0525)
    decimal_rate = annual_rate / 100.0

    // Apply compound interest formula
    // A = P(1 + r/n)^(nt)
    compound_factor = pow(1 + (decimal_rate / compounds_per_year),
                         compounds_per_year * years)
    final_amount = principal * compound_factor
    interest_earned = final_amount - principal

    return {
        "final_amount": round(final_amount, 2),
        "interest_earned": round(interest_earned, 2),
        "effective_rate": round(((final_amount / principal) - 1) * 100, 2),
        "calculation_details": {
            "principal": principal,
            "rate": annual_rate,
            "periods": compounds_per_year,
            "years": years
        }
    }
end

// ✓ Good: Inline comments for complex logic
// Use binary search for O(log n) performance on sorted arrays
while (left <= right):
    mid = floor((left + right) / 2)

    if (arr[mid] == target) then:
        return mid  // Found exact match
    else if (arr[mid] < target) then:
        left = mid + 1  // Search right half
    else:
        right = mid - 1  // Search left half
    end
end

// ✗ Avoid: Obvious or redundant comments
i = i + 1          // Increment i by 1
name = "John"      // Set name to John
return true        // Return true
```

## Function Design Principles

### Single Responsibility Principle

```uddin
// ✓ Good: Functions with single, well-defined purposes
fun validate_email_format(email):
    // Only responsible for email format validation
    email_pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    return is_regex_match(email_pattern, trim(email))
end

fun format_phone_number(phone):
    // Only responsible for phone number formatting
    if (typeof(phone) != "string") then:
        return ""
    end

    // Extract only digits
    digits = replace(phone, "[^0-9]", "")

    // Format as (XXX) XXX-XXXX for 10-digit US numbers
    if (len(digits) == 10) then:
        return "(" + substr(digits, 0, 3) + ") " +
               substr(digits, 3, 6) + "-" +
               substr(digits, 6, 10)
    end

    return phone  // Return original if not standard format
end

fun send_notification_email(email, subject, message):
    // Only responsible for sending email notifications
    try:
        email_config = {
            "to": email,
            "subject": subject,
            "body": message,
            "from": "noreply@example.com"
        }

        return send_email(email_config)
    catch (error):
        log_error("Email notification failed", {
            "email": email,
            "error": str(error)
        })
        return false
    end
end

// ✗ Avoid: Functions that do too many things
fun process_user_signup_bad(user_data):
    // This function violates SRP by doing too many things:

    // 1. Validates email format
    if (not is_regex_match("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", user_data.email)) then:
        return {"error": "Invalid email"}
    end

    // 2. Formats phone number
    digits = replace(user_data.phone, "[^0-9]", "")
    if (len(digits) == 10) then:
        user_data.phone = "(" + substr(digits, 0, 3) + ") " + substr(digits, 3, 6) + "-" + substr(digits, 6, 10)
    end

    // 3. Saves to database
    database_save("users", user_data)

    // 4. Sends welcome email
    send_email(user_data.email, "Welcome!", "Thank you for signing up")

    // 5. Updates analytics
    track_event("user_signup", {"email": user_data.email})

    return {"success": true}
end

// ✓ Good: Proper separation of concerns
fun process_user_signup_good(user_data):
    // Coordinate the signup process with clear separation
    validation_result = validate_user_signup_data(user_data)
    if (not validation_result.valid) then:
        return validation_result
    end

    formatted_data = format_user_data(user_data)
    saved_user = save_user_to_database(formatted_data)

    // Handle post-signup tasks asynchronously
    schedule_welcome_email(saved_user.email)
    track_signup_event(saved_user)

    return {
        "success": true,
        "user": saved_user
    }
end
```

### Function Parameters and Return Values

```uddin
// ✓ Good: Clear parameter names and reasonable parameter count
fun create_user_account(name, email, password, preferences = {}):
    return {
        "id": generate_user_id(),
        "name": trim(name),
        "email": lower(trim(email)),
        "password_hash": hash_password(password),
        "preferences": preferences,
        "created_at": date_now(),
        "status": "active"
    }
end

// ✓ Good: Use objects for complex parameter sets
fun create_advanced_user_account(user_config):
    // Validate required fields
    required_fields = ["name", "email", "password"]
    for (field in required_fields):
        if (not user_config[field] or len(trim(str(user_config[field]))) == 0) then:
            panic("Missing required field: " + field)
        end
    end

    // Apply defaults for optional fields
    config = {
        "name": trim(user_config.name),
        "email": lower(trim(user_config.email)),
        "password": user_config.password,
        "role": ("role" in user_config and user_config.role != null) ? user_config.role : "user",
        "preferences": ("preferences" in user_config and user_config.preferences != null) ? user_config.preferences : {},
        "notifications": ("notifications" in user_config and user_config.notifications != null) ? user_config.notifications : true,
        "theme": ("theme" in user_config and user_config.theme != null) ? user_config.theme : "light"
    }

    return {
        "id": generate_user_id(),
        "name": config.name,
        "email": config.email,
        "password_hash": hash_password(config.password),
        "role": config.role,
        "preferences": config.preferences,
        "settings": {
            "notifications": config.notifications,
            "theme": config.theme
        },
        "created_at": date_now(),
        "status": "active"
    }
end

// ✓ Good: Consistent return value structures
fun find_user_by_id(user_id):
    if (not user_id or len(str(user_id)) == 0) then:
        return {
            "success": false,
            "error": "User ID is required",
            "error_code": "MISSING_USER_ID"
        }
    end

    user = database_find("users", user_id)

    if (user) then:
        return {
            "success": true,
            "user": user,
            "timestamp": date_now()
        }
    else:
        return {
            "success": false,
            "error": "User not found",
            "error_code": "USER_NOT_FOUND",
            "user_id": user_id
        }
    end
end

// ✗ Avoid: Too many parameters (consider using an object)
fun create_user_bad(name, email, password, phone, address, city, state, zip, country,
                   birth_date, preferences, notifications, theme, language):
    // Too many parameters make the function hard to use and maintain
end

// ✗ Avoid: Inconsistent return types
fun find_user_inconsistent(user_id):
    user = database_find("users", user_id)

    if (user) then:
        return user          // Returns user object
    else:
        return false         // Returns boolean - inconsistent!
    end
end
```

### Modern Function Patterns

```uddin
// ✓ Good: Higher-order functions for reusability
fun apply_discount(price, discount_function):
    return discount_function(price)
end

// Discount strategies
percentage_discount = fun(percentage):
    return fun(price):
        return price * (1 - percentage / 100)
    end
end

fixed_amount_discount = fun(amount):
    return fun(price):
        return max(0, price - amount)
    end
end

// Usage
original_price = 100.0
discounted_price_1 = apply_discount(original_price, percentage_discount(15))  // 15% off
discounted_price_2 = apply_discount(original_price, fixed_amount_discount(20))  // $20 off

// ✓ Good: Functional array operations for data processing
fun process_sales_data(sales_records):
    // Filter valid sales (amount > 0 and date exists)
    valid_sales = filter(sales_records, fun(sale):
        return sale.amount > 0 and sale.date != null
    end)

    // Transform to include calculated fields
    enriched_sales = map(valid_sales, fun(sale):
        tax_amount = sale.amount * 0.08
        total_with_tax = sale.amount + tax_amount

        return {
            "id": sale.id,
            "amount": sale.amount,
            "tax": round(tax_amount, 2),
            "total": round(total_with_tax, 2),
            "date": sale.date,
            "category": ("category" in sale and sale.category != null) ? sale.category : "general"
        }
    end)

    // Calculate summary statistics
    total_revenue = reduce(enriched_sales, fun(sum, sale):
        return sum + sale.total
    end, 0)

    return {
        "sales": enriched_sales,
        "summary": {
            "count": len(enriched_sales),
            "total_revenue": round(total_revenue, 2),
            "average_sale": round(total_revenue / len(enriched_sales), 2)
        }
    }
end
```

## Error Handling Best Practices

### Comprehensive Error Handling with try-catch

```uddin
// ✓ Good: Specific error handling with detailed context
fun process_payment_transaction(payment_data):
    try:
        // Step 1: Validate payment data
        validation_result = validate_payment_data(payment_data)
        if (not validation_result.valid) then:
            return {
                "success": false,
                "error": "Payment validation failed",
                "details": validation_result.errors,
                "error_code": "VALIDATION_ERROR"
            }
        end

        // Step 2: Process payment with external service
        payment_response = charge_credit_card(payment_data)

        if (payment_response.success) then:
            // Step 3: Record successful transaction
            transaction_record = {
                "transaction_id": payment_response.transaction_id,
                "amount": payment_data.amount,
                "currency": ("currency" in payment_data and payment_data.currency != null) ? payment_data.currency : "USD",
                "customer_id": payment_data.customer_id,
                "timestamp": date_now(),
                "status": "completed"
            }

            save_transaction(transaction_record)
            send_payment_confirmation(payment_data.customer_email, transaction_record)

            return {
                "success": true,
                "transaction": transaction_record,
                "message": "Payment processed successfully"
            }
        else:
            return {
                "success": false,
                "error": "Payment processing failed",
                "details": payment_response.error_message,
                "error_code": "PAYMENT_DECLINED"
            }
        end

    catch (error):
        // Log detailed error information for debugging
        error_context = {
            "error_message": str(error),
            "payment_data": {
                "amount": payment_data.amount,
                "customer_id": payment_data.customer_id,
                "currency": payment_data.currency
            },
            "timestamp": date_now()
        }

        log_error("Payment processing error", error_context)

        return {
            "success": false,
            "error": "Internal payment processing error",
            "error_code": "INTERNAL_ERROR",
            "details": "Please try again later or contact support"
        }
    end
end

// ✓ Good: Retry logic with exponential backoff
fun fetch_data_with_smart_retry(url, max_retries = 3):
    for (attempt in range(1, max_retries + 1)):
        try:
            response = http_get(url)

            // Success - return immediately
            return {
                "success": true,
                "data": response,
                "attempts": attempt
            }

        catch (error):
            error_message = str(error)

            // Don't retry for client errors (4xx)
            if (contains(error_message, "400") or contains(error_message, "404")) then:
                return {
                    "success": false,
                    "error": "Client error - not retrying",
                    "details": error_message,
                    "attempts": attempt
                }
            end

            // Log attempt failure
            print("Attempt " + str(attempt) + " failed: " + error_message)

            // Return error if this was the last attempt
            if (attempt == max_retries) then:
                return {
                    "success": false,
                    "error": "Max retries exceeded",
                    "details": error_message,
                    "attempts": attempt
                }
            end

            // Calculate exponential backoff delay
            delay_ms = pow(2, attempt) * 1000
            print("Retrying in " + str(delay_ms) + "ms...")

            // In a real implementation, you would add sleep/delay here
            // sleep(delay_ms)
        end
    end
end
```

### Input Validation

````uddin
### Input Validation and Sanitization

```uddin
// ✓ Good: Comprehensive input validation with proper error messages
fun validate_user_registration(user_data):
    errors = []

    // Validate name
    if (not ("name" in user_data) or user_data.name == null or len(trim(user_data.name)) == 0) then:
        append(errors, "Full name is required")
    else if (len(user_data.name) > 100) then:
        append(errors, "Name cannot exceed 100 characters")
    else if (not is_regex_match("^[a-zA-Z\s\-\.]+$", user_data.name)) then:
        append(errors, "Name contains invalid characters")
    end

    // Validate email
    if (not ("email" in user_data) or user_data.email == null or len(trim(user_data.email)) == 0) then:
        append(errors, "Email address is required")
    else if (not validate_email_format(user_data.email)) then:
        append(errors, "Invalid email format")
    else if (len(user_data.email) > 254) then:
        append(errors, "Email address too long")
    end

    // Validate password strength
    if (not ("password" in user_data) or user_data.password == null or len(user_data.password) == 0) then:
        append(errors, "Password is required")
    else:
        password_validation = validate_password_strength(user_data.password)
        if (not password_validation.valid) then:
            errors = append(errors, password_validation.errors)
        end
    end

    // Validate phone number (optional)
    if (("phone" in user_data) and user_data.phone != null and len(trim(user_data.phone)) > 0) then:
        if (not validate_phone_format(user_data.phone)) then:
            append(errors, "Invalid phone number format")
        end
    end

    // Validate age
    if (("age" in user_data) and user_data.age != null) then:
        if (typeof(user_data.age) != "number" or user_data.age < 13 or user_data.age > 120) then:
            append(errors, "Age must be between 13 and 120")
        end
    end

    return {
        "valid": len(errors) == 0,
        "errors": errors
    }
end

fun validate_password_strength(password):
    errors = []

    if (len(password) < 8) then:
        append(errors, "Password must be at least 8 characters long")
    end

    if (not is_regex_match(".*[A-Z].*", password)) then:
        append(errors, "Password must contain at least one uppercase letter")
    end

    if (not is_regex_match(".*[a-z].*", password)) then:
        append(errors, "Password must contain at least one lowercase letter")
    end

    if (not is_regex_match(".*[0-9].*", password)) then:
        append(errors, "Password must contain at least one number")
    end

    if (not is_regex_match(".*[!@#$%^&*()_+\-=\[\]{};':"\|,.<>\?].*", password)) then:
        append(errors, "Password must contain at least one special character")
    end

    // Check for common weak passwords
    weak_passwords = ["password", "123456", "qwerty", "admin", "letmein"]
    if (contains(weak_passwords, lower(password))) then:
        append(errors, "Password is too common - please choose a more secure password")
    end

    return {
        "valid": len(errors) == 0,
        "errors": errors
    }
end

// ✓ Good: Input sanitization to prevent security issues
fun sanitize_user_input(input, input_type = "general"):
    if (typeof(input) != "string") then:
        return ""
    end

    // Basic sanitization
    sanitized = trim(input)

    // Type-specific sanitization
    if (input_type == "html") then:
        // Escape HTML special characters
        sanitized = replace(sanitized, "&", "&amp;")
        sanitized = replace(sanitized, "<", "&lt;")
        sanitized = replace(sanitized, ">", "&gt;")
        sanitized = replace(sanitized, """, "&quot;")
        sanitized = replace(sanitized, "'", "&#x27;")
    else if (input_type == "sql") then:
        // Escape SQL special characters
        sanitized = replace(sanitized, "'", "''")
        sanitized = replace(sanitized, "", "")
    else if (input_type == "filename") then:
        // Remove dangerous characters for filenames
        sanitized = replace(sanitized, "[<>:"/\|?*]", "")
        sanitized = replace(sanitized, "\.\.+", "")  // Remove path traversal
    else if (input_type == "alphanumeric") then:
        // Keep only alphanumeric characters and spaces
        sanitized = replace(sanitized, "[^a-zA-Z0-9\s]", "")
    end

    // General security measures
    // Remove control characters
    sanitized = replace(sanitized, "[\x00-\x1F\x7F]", "")

    // Limit length to prevent buffer overflow attacks
    max_length = 10000
    if (len(sanitized) > max_length) then:
        sanitized = substr(sanitized, 0, max_length)
    end

    return sanitized
end

// ✓ Good: Parameterized query-like approach for data access
fun safe_database_query(query_template, parameters):
    // This simulates parameterized queries to prevent SQL injection
    safe_query = query_template

    for (i in range(len(parameters))):
        param = parameters[i]
        placeholder = "$" + str(i + 1)

        // Properly escape and quote parameters based on type
        if (typeof(param) == "string") then:
            escaped_param = "'" + replace(param, "'", "''") + "'"
            safe_query = replace(safe_query, placeholder, escaped_param)
        else if (typeof(param) == "number") then:
            safe_query = replace(safe_query, placeholder, str(param))
        else if (param == null) then:
            safe_query = replace(safe_query, placeholder, "NULL")
        else:
            // Convert other types to strings and escape
            escaped_param = "'" + replace(str(param), "'", "''") + "'"
            safe_query = replace(safe_query, placeholder, escaped_param)
        end
    end

    return safe_query
end

// Usage example
user_id = 123
user_name = "O'Connor"  // Contains potentially dangerous single quote

// ✗ Dangerous: Direct string concatenation
bad_query = "SELECT * FROM users WHERE id = " + str(user_id) + " AND name = '" + user_name + "'"

// ✓ Safe: Parameterized approach
safe_query = safe_database_query(
    "SELECT * FROM users WHERE id = $1 AND name = $2",
    [user_id, user_name]
)
````

## Performance Best Practices

### Efficient Data Structures and Algorithms

```uddin
// ✓ Good: Use appropriate data structures for the task
fun create_fast_user_lookup(users):
    // Use object/map for O(1) lookups instead of array O(n) search
    user_index = {}
    email_index = {}

    for (user in users):
        user_index[str(user.id)] = user
        email_index[lower(user.email)] = user
    end

    return {
        "by_id": user_index,
        "by_email": email_index
    }
end

fun find_user_optimized(user_lookup, id = null, email = null):
    if (id) then:
        return user_lookup.by_id[str(id)] or null
    else if (email) then:
        return user_lookup.by_email[lower(email)] or null
    end
    return null
end

// ✓ Good: Efficient set operations using Uddin-Lang's built-in Set
fun find_common_interests(user1_interests, user2_interests):
    // Convert to sets for efficient intersection
    set1 = set_new()
    set2 = set_new()

    for (interest in user1_interests):
        set_add(set1, interest)
    end

    for (interest in user2_interests):
        set_add(set2, interest)
    end

    // Find intersection
    common = []
    for (interest in user1_interests):
        if (set_has(set2, interest)) then:
            append(common, interest)
        end
    end

    return common
end

// ✓ Good: Batch processing for better performance
fun process_user_batch(user_ids, batch_size = 100):
    results = []
    total_batches = ceil(len(user_ids) / batch_size)

    for (batch_num in range(total_batches)):
        start_idx = batch_num * batch_size
        end_idx = min(start_idx + batch_size, len(user_ids))

        batch_ids = slice(user_ids, start_idx, end_idx)

        // Process batch instead of individual items
        batch_users = database_find_multiple("users", batch_ids)

        for (user in batch_users):
            processed_user = process_single_user(user)
            append(results, processed_user)
        end

        // Optional: Progress reporting for long operations
        if (total_batches > 5) then:
            progress = round((batch_num + 1) / total_batches * 100, 1)
            print("Processing batch " + str(batch_num + 1) + "/" + str(total_batches) +
                  " (" + str(progress) + "%)")
        end
    end

    return results
end

// ✗ Avoid: Inefficient nested loops (O(n²) complexity)
fun find_duplicates_slow(items):
    duplicates = []

    for (i in range(len(items))):
        for (j in range(i + 1, len(items))):  // O(n²) time complexity
            if (items[i] == items[j]) then:
                append(duplicates, items[i])
            end
        end
    end

    return duplicates
end

// ✓ Good: Use sets for O(n) duplicate detection
fun find_duplicates_fast(items):
    seen = set_new()
    duplicates = set_new()

    for (item in items):
        if (set_has(seen, item)) then:
            set_add(duplicates, item)
        else:
            set_add(seen, item)
        end
    end

    return set_to_array(duplicates)
end
```

### Memory Management and Resource Optimization

```uddin
// ✓ Good: Efficient file processing with streaming approach
fun process_large_file_efficiently(filename, chunk_size = 1000):
    if (not file_exists(filename)) then:
        panic("File not found: " + filename)
    end

    file_size = file_size(filename)
    print("Processing file: " + filename + " (" + str(file_size) + " bytes)")

    try:
        // Read and process file in chunks to avoid memory issues
        file_content = read_file(filename)
        lines = split(file_content, "
")

        processed_count = 0
        batch_buffer = []

        for (line in lines):
            if (len(trim(line)) > 0) then:
                processed_line = process_single_line(line)
                append(batch_buffer, processed_line)

                // Process in batches to manage memory
                if (len(batch_buffer) >= chunk_size) then:
                    save_processed_batch(batch_buffer)
                    batch_buffer = []  // Clear buffer to free memory
                    processed_count = processed_count + chunk_size

                    // Progress reporting
                    progress = round(processed_count / len(lines) * 100, 1)
                    print("Processed " + str(processed_count) + "/" + str(len(lines)) +
                          " lines (" + str(progress) + "%)")
                end
            end
        end

        // Process remaining lines
        if (len(batch_buffer) > 0) then:
            save_processed_batch(batch_buffer)
        end

        print("File processing completed successfully")

    catch (error):
        print("Error processing file: " + str(error))
        panic("File processing failed")
    end
end

// ✓ Good: Cache implementation with memory management
fun create_smart_cache(max_size = 1000, ttl_seconds = 3600):
    cache_data = {}
    cache_timestamps = {}
    cache_access_count = {}

    return {
        "get": fun(key):
            current_time = date_now()

            // Check if key exists and is not expired
            if (cache_data[key] and
                (current_time - cache_timestamps[key]) < (ttl_seconds * 1000)) then:

                // Update access count for LRU tracking
                cache_access_count[key] = (cache_access_count[key] or 0) + 1
                return cache_data[key]
            else:
                // Remove expired entry by rebuilding cache objects
                if (cache_data[key]) then:
                    new_cache_data = {}
                    new_cache_timestamps = {}
                    new_cache_access_count = {}

                    for (cache_key in cache_data):
                        if (cache_key != key) then:
                            new_cache_data[cache_key] = cache_data[cache_key]
                            new_cache_timestamps[cache_key] = cache_timestamps[cache_key]
                            new_cache_access_count[cache_key] = cache_access_count[cache_key]
                        end
                    end

                    cache_data = new_cache_data
                    cache_timestamps = new_cache_timestamps
                    cache_access_count = new_cache_access_count
                end
                return null
            end
        end,

        "set": fun(key, value):
            current_time = date_now()

            // Evict old entries if cache is full
            if (len(cache_data) >= max_size and not cache_data[key]) then:
                this.evict_least_recently_used()
            end

            cache_data[key] = value
            cache_timestamps[key] = current_time
            cache_access_count[key] = 1
        end,

        "evict_least_recently_used": fun():
            if (len(cache_data) == 0) then:
                return
            end

            // Find key with lowest access count
            lru_key = null
            min_access = null

            for (key in cache_data):
                access_count = cache_access_count[key] or 0
                if (lru_key == null or access_count < min_access) then:
                    lru_key = key
                    min_access = access_count
                end
            end

            // Remove LRU entry by rebuilding cache objects
            if (lru_key) then:
                new_cache_data = {}
                new_cache_timestamps = {}
                new_cache_access_count = {}

                for (cache_key in cache_data):
                    if (cache_key != lru_key) then:
                        new_cache_data[cache_key] = cache_data[cache_key]
                        new_cache_timestamps[cache_key] = cache_timestamps[cache_key]
                        new_cache_access_count[cache_key] = cache_access_count[cache_key]
                    end
                end

                cache_data = new_cache_data
                cache_timestamps = new_cache_timestamps
                cache_access_count = new_cache_access_count
            end
        end,

        "clear": fun():
            cache_data = {}
            cache_timestamps = {}
            cache_access_count = {}
        end,

        "size": fun():
            return len(cache_data)
        end
    }
end

// Usage example of smart cache
global_cache = create_smart_cache(500, 1800)  // 500 items, 30 min TTL

fun expensive_computation(input):
    cache_key = "computation_" + str(input)

    // Try to get from cache first
    cached_result = global_cache.get(cache_key)
    if (cached_result) then:
        return cached_result
    end

    // Perform expensive computation
    result = {
        "input": input,
        "output": pow(input, 3) + (input * 2),
        "computed_at": date_now()
    }

    // Store in cache for future use
    global_cache.set(cache_key, result)

    return result
end
```

### Functional Programming Optimizations

```uddin
// ✓ Good: Efficient functional programming patterns
fun process_sales_data_efficiently(sales):
    // Chain operations efficiently without creating intermediate arrays
    return sales
        |> filter(fun(sale): return sale.amount > 0 and sale.status == "completed" end)
        |> map(fun(sale):
            tax = sale.amount * 0.08
            return {
                "id": sale.id,
                "net_amount": sale.amount,
                "tax_amount": round(tax, 2),
                "total_amount": round(sale.amount + tax, 2),
                "date": sale.date,
                "category": ("category" in sale and sale.category != null) ? sale.category : "general"
            }
        end)
        |> sort_by(fun(sale): return sale.total_amount end)
end

// Note: The |> operator is conceptual here - in actual Uddin-Lang, chain the operations:
fun process_sales_data_efficiently_actual(sales):
    // Filter valid sales
    valid_sales = filter(sales, fun(sale):
        return sale.amount > 0 and sale.status == "completed"
    end)

    // Transform to add calculated fields
    enriched_sales = map(valid_sales, fun(sale):
        tax = sale.amount * 0.08
        return {
            "id": sale.id,
            "net_amount": sale.amount,
            "tax_amount": round(tax, 2),
            "total_amount": round(sale.amount + tax, 2),
            "date": sale.date,
            "category": ("category" in sale and sale.category != null) ? sale.category : "general"
        }
    end)

    // Sort by total amount
    sort(enriched_sales)  // Sorts in place for better performance

    return enriched_sales
end

// ✓ Good: Efficient reduce operations
fun calculate_statistics(numbers):
    if (len(numbers) == 0) then:
        return {"error": "Cannot calculate statistics for empty array"}
    end

    // Calculate multiple statistics in a single pass
    stats = reduce(numbers, fun(acc, num):
        return {
            "sum": acc.sum + num,
            "count": acc.count + 1,
            "min": min(acc.min, num),
            "max": max(acc.max, num),
            "sum_squares": acc.sum_squares + (num * num)
        }
    end, {
        "sum": 0,
        "count": 0,
        "min": numbers[0],
        "max": numbers[0],
        "sum_squares": 0
    })

    // Calculate derived statistics
    mean_value = stats.sum / stats.count
    variance = (stats.sum_squares / stats.count) - (mean_value * mean_value)

    return {
        "count": stats.count,
        "sum": stats.sum,
        "min": stats.min,
        "max": stats.max,
        "mean": round(mean_value, 2),
        "variance": round(variance, 2),
        "std_dev": round(sqrt(variance), 2)
    }
end
```uddin

// ✓ Good: Sanitize user input
function sanitizeUserInput(input) {
if (typeof(input) != "string") {
return ""
}

    // Remove potentially dangerous characters
    sanitized = input
    sanitized = regex_replace(sanitized, `[<>"'&]`, "")

    // Trim whitespace
    sanitized = trim(sanitized)

    // Limit length
    if (len(sanitized) > 1000) {
        sanitized = substring(sanitized, 0, 1000)
    }

    return sanitized

}
```

## Performance Best Practices

### Efficient Data Structures

```uddin
// ✓ Good: Use appropriate data structures
function createUserIndex(users) {
    // Use object for O(1) lookups instead of array search
    user_index = {}

    for (user in users) {
        user_index[user.id] = user
    }

    return user_index
}

function findUserById(user_index, user_id) {
    // O(1) lookup instead of O(n) array search
    return user_index[user_id] || null
}

// ✓ Good: Batch operations when possible
function processMultipleUsers(user_ids, operation) {
    results = []

    // Batch database queries instead of individual calls
    users = database_find_multiple("users", user_ids)

    for (user in users) {
        try {
            result = operation(user)
            results = append(results, {"user_id": user.id, "success": true, "result": result})
        } catch (error) {
            results = append(results, {"user_id": user.id, "success": false, "error": str(error)})
        }
    }

    return results
}

// ✗ Avoid: Inefficient nested loops
function findCommonElementsBad(list1, list2) {
    common = []

    for (item1 in list1) {
        for (item2 in list2) {  // O(n²) complexity
            if (item1 == item2) {
                common = append(common, item1)
            }
        }
    }

    return common
}

// ✓ Good: Use sets for better performance
function findCommonElementsGood(list1, list2) {
    set1 = set_create()
    for (item in list1) {
        set_add(set1, item)
    }

    common = []
    for (item in list2) {
        if (set_contains(set1, item)) {
            common = append(common, item)
        }
    }

    return common
}
````

### Memory Management

```uddin
// ✓ Good: Clean up resources
function processLargeFile(filename) {
    file_handle = null
    temp_data = null

    try {
        file_handle = file_open(filename)
        temp_data = []

        while (!file_eof(file_handle)) {
            line = file_read_line(file_handle)
            processed_line = processLine(line)
            temp_data = append(temp_data, processed_line)

            // Process in chunks to avoid memory issues
            if (len(temp_data) >= 1000) {
                saveChunk(temp_data)
                temp_data = []  // Clear processed data
            }
        }

        // Process remaining data
        if (len(temp_data) > 0) {
            saveChunk(temp_data)
        }

    } catch (error) {
        throw "File processing error: " + str(error)

    } finally {
        // Always clean up resources
        if (file_handle) {
            file_close(file_handle)
        }
        temp_data = null
    }
}

// ✓ Good: Avoid memory leaks in long-running processes
function createDataProcessor() {
    cache = {}
    cache_size = 0
    max_cache_size = 1000

    return {
        "process": function(data) {
            cache_key = generateCacheKey(data)

            // Check cache first
            if (cache[cache_key]) {
                return cache[cache_key]
            }

            // Process data
            result = expensiveOperation(data)

            // Add to cache with size management
            if (cache_size >= max_cache_size) {
                this.clearOldestCacheEntries()
            }

            cache[cache_key] = result
            cache_size = cache_size + 1

            return result
        },

        "clearOldestCacheEntries": function() {
            // Simple cache eviction - remove half of entries
            keys_to_remove = []
            count = 0

            for (key in cache) {
                keys_to_remove = append(keys_to_remove, key)
                count = count + 1

                if (count >= cache_size / 2) {
                    break
                }
            }

            for (key in keys_to_remove) {
                // Remove key by rebuilding cache object
                new_cache = {}
                for (cache_key in cache) {
                    if (cache_key != key) {
                        new_cache[cache_key] = cache[cache_key]
                    }
                }
                cache = new_cache
                cache_size = cache_size - 1
            }
        }
    }
}
```

## Code Organization

### Module Structure

```uddin
// ✓ Good: Well-organized module structure
// File: user_service.din

// Constants at the top
MAX_LOGIN_ATTEMPTS = 3
SESSION_TIMEOUT = 3600000  // 1 hour in milliseconds

// Helper functions
function generateUserId() {
    return "user_" + str(date_now()) + "_" + str(floor(random() * 10000))
}

function hashPassword(password) {
    // Implementation for password hashing
    return "hashed_" + password  // Simplified for example
}

// Main service functions
function createUser(user_data) {
    // User creation logic
}

function authenticateUser(email, password) {
    // Authentication logic
}

function updateUserProfile(user_id, updates) {
    // Profile update logic
}

// Export main functions (if module system supports it)
if (typeof(module) != "undefined") {
    module.exports = {
        "createUser": createUser,
        "authenticateUser": authenticateUser,
        "updateUserProfile": updateUserProfile
    }
}
```

### Configuration Management

```uddin
// ✓ Good: Centralized configuration
// File: config.din

function createConfig() {
    // Default configuration
    default_config = {
        "database": {
            "host": "localhost",
            "port": 5432,
            "name": "myapp",
            "timeout": 30000
        },
        "api": {
            "base_url": "https://api.example.com",
            "timeout": 10000,
            "retry_attempts": 3
        },
        "security": {
            "jwt_secret": "default_secret",
            "session_timeout": 3600000,
            "max_login_attempts": 3
        },
        "logging": {
            "level": "INFO",
            "file": "app.log",
            "max_size": 10485760  // 10MB
        }
    }

    // Load environment-specific overrides
    env = getEnvironment()
    if (env == "production") {
        default_config.logging.level = "ERROR"
        default_config.security.jwt_secret = getEnvVar("JWT_SECRET")
    } else if (env == "development") {
        default_config.logging.level = "DEBUG"
        default_config.database.host = "localhost"
    }

    return default_config
}

function getEnvironment() {
    // Determine current environment
    return getEnvVar("NODE_ENV") || "development"
}

function getEnvVar(name) {
    // Get environment variable (implementation depends on system)
    return ""  // Simplified for example
}

// Global configuration instance
CONFIG = createConfig()
```

## Testing Best Practices

### Unit Testing

```uddin
// ✓ Good: Comprehensive unit tests
// File: user_service_test.din

import "user_service.din"
import "test_framework.din"

function testCreateUser() {
    // Test valid user creation
    valid_user_data = {
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
    }

    result = createUser(valid_user_data)

    assert(result.success == true, "User creation should succeed")
    assert(result.user.name == "John Doe", "User name should be set correctly")
    assert(result.user.email == "john@example.com", "User email should be set correctly")
    assert(result.user.id != null, "User should have an ID")

    print("✓ testCreateUser passed")
}

function testCreateUserInvalidData() {
    // Test with missing required fields
    invalid_user_data = {
        "name": "",  // Empty name
        "email": "invalid-email",  // Invalid email format
        "age": -5  // Invalid age
    }

    result = createUser(invalid_user_data)

    assert(result.success == false, "User creation should fail with invalid data")
    assert(len(result.errors) > 0, "Should return validation errors")

    print("✓ testCreateUserInvalidData passed")
}

function testAuthenticateUser() {
    // Setup: Create a test user
    user_data = {
        "name": "Test User",
        "email": "test@example.com",
        "password": "secure123"
    }

    create_result = createUser(user_data)
    assert(create_result.success == true, "Test user creation should succeed")

    // Test successful authentication
    auth_result = authenticateUser("test@example.com", "secure123")
    assert(auth_result.success == true, "Authentication should succeed")
    assert(auth_result.user.email == "test@example.com", "Should return correct user")

    // Test failed authentication
    failed_auth = authenticateUser("test@example.com", "wrongpassword")
    assert(failed_auth.success == false, "Authentication should fail with wrong password")

    print("✓ testAuthenticateUser passed")
}

// Test runner
function runAllTests() {
    print("Running user service tests...")

    try {
        testCreateUser()
        testCreateUserInvalidData()
        testAuthenticateUser()

        print("\n✅ All tests passed!")

    } catch (error) {
        print("\n❌ Test failed: " + str(error))
    }
}

// Run tests if this file is executed directly
if (typeof(main) != "undefined") {
    runAllTests()
}
```

### Integration Testing

```uddin
// ✓ Good: Integration test example
// File: api_integration_test.din

function testUserRegistrationFlow() {
    print("Testing complete user registration flow...")

    // Step 1: Register new user
    registration_data = {
        "name": "Integration Test User",
        "email": "integration@test.com",
        "password": "testpass123"
    }

    register_response = http_post("/api/register", registration_data)
    assert(register_response.status == 201, "Registration should return 201")
    assert(register_response.data.user.email == registration_data.email, "Should return user data")

    user_id = register_response.data.user.id

    // Step 2: Verify user can login
    login_data = {
        "email": registration_data.email,
        "password": registration_data.password
    }

    login_response = http_post("/api/login", login_data)
    assert(login_response.status == 200, "Login should succeed")
    assert(login_response.data.token != null, "Should return auth token")

    auth_token = login_response.data.token

    // Step 3: Test authenticated endpoint
    headers = {"Authorization": "Bearer " + auth_token}
    profile_response = http_get("/api/profile", headers)

    assert(profile_response.status == 200, "Profile access should succeed")
    assert(profile_response.data.user.id == user_id, "Should return correct user profile")

    // Step 4: Cleanup - delete test user
    delete_response = http_delete("/api/users/" + user_id, headers)
    assert(delete_response.status == 200, "User deletion should succeed")

    print("✓ User registration flow test passed")
}

function testErrorHandling() {
    print("Testing API error handling...")

    // Test invalid registration data
    invalid_data = {
        "name": "",  // Empty name
        "email": "invalid-email",  // Invalid email
        "password": "123"  // Too short password
    }

    response = http_post("/api/register", invalid_data)
    assert(response.status == 400, "Should return 400 for invalid data")
    assert(len(response.data.errors) > 0, "Should return validation errors")

    // Test non-existent endpoint
    not_found_response = http_get("/api/nonexistent")
    assert(not_found_response.status == 404, "Should return 404 for non-existent endpoint")

    print("✓ Error handling test passed")
}
```

## Security Best Practices

### Input Sanitization and Validation

```uddin
// ✓ Good: Comprehensive input validation
function validateAndSanitizeInput(input, type) {
    if (typeof(input) != "string") {
        return {"valid": false, "error": "Input must be a string"}
    }

    // Basic sanitization
    sanitized = trim(input)

    switch (type) {
        case "email":
            // Email validation
            email_pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
            if (!regex_match(sanitized, email_pattern)) {
                return {"valid": false, "error": "Invalid email format"}
            }
            break

        case "phone":
            // Phone number validation
            phone_pattern = "^[+]?[1-9]?[0-9]{7,15}$"
            clean_phone = regex_replace(sanitized, "[^0-9+]", "")
            if (!regex_match(clean_phone, phone_pattern)) {
                return {"valid": false, "error": "Invalid phone number format"}
            }
            sanitized = clean_phone
            break

        case "name":
            // Name validation
            if (len(sanitized) < 2 || len(sanitized) > 50) {
                return {"valid": false, "error": "Name must be 2-50 characters"}
            }
            // Remove potentially dangerous characters
            sanitized = regex_replace(sanitized, "[<>\"'&{}]", "")
            break

        case "password":
            // Password validation
            if (len(sanitized) < 8) {
                return {"valid": false, "error": "Password must be at least 8 characters"}
            }
            if (!regex_match(sanitized, ".*[A-Z].*")) {
                return {"valid": false, "error": "Password must contain uppercase letter"}
            }
            if (!regex_match(sanitized, ".*[a-z].*")) {
                return {"valid": false, "error": "Password must contain lowercase letter"}
            }
            if (!regex_match(sanitized, ".*[0-9].*")) {
                return {"valid": false, "error": "Password must contain number"}
            }
            break
    }

    return {"valid": true, "value": sanitized}
}

// ✓ Good: SQL injection prevention
function safeDbQuery(query_template, parameters) {
    // Use parameterized queries instead of string concatenation
    safe_query = query_template

    for (i = 0; i < len(parameters); i++) {
        param = parameters[i]

        // Escape special characters
        if (typeof(param) == "string") {
            escaped_param = regex_replace(param, "['\"\\]", "\\$0")
            safe_query = regex_replace(safe_query, "\\$" + str(i + 1), "'" + escaped_param + "'")
        } else {
            safe_query = regex_replace(safe_query, "\\$" + str(i + 1), str(param))
        }
    }

    return safe_query
}

// Usage example
user_id = 123
user_name = "John O'Connor"  // Contains single quote

// ✗ Dangerous: Direct string concatenation
bad_query = "SELECT * FROM users WHERE id = " + str(user_id) + " AND name = '" + user_name + "'"

// ✓ Safe: Parameterized query
safe_query = safeDbQuery("SELECT * FROM users WHERE id = $1 AND name = $2", [user_id, user_name])
```

### Authentication and Authorization

```uddin
// ✓ Good: Secure session management
function createSecureSession(user_id) {
    session_data = {
        "user_id": user_id,
        "created_at": date_now(),
        "expires_at": date_now() + SESSION_TIMEOUT,
        "ip_address": getCurrentIP(),
        "user_agent": getCurrentUserAgent()
    }

    // Generate secure session token
    session_token = generateSecureToken()

    // Store session data securely
    storeSession(session_token, session_data)

    return session_token
}

function validateSession(session_token) {
    if (!session_token || len(session_token) == 0) {
        return {"valid": false, "error": "No session token provided"}
    }

    session_data = getSession(session_token)
    if (!session_data) {
        return {"valid": false, "error": "Invalid session token"}
    }

    // Check if session has expired
    if (date_now() > session_data.expires_at) {
        deleteSession(session_token)
        return {"valid": false, "error": "Session expired"}
    }

    // Optional: Check IP address for additional security
    current_ip = getCurrentIP()
    if (session_data.ip_address != current_ip) {
        deleteSession(session_token)
        return {"valid": false, "error": "Session security violation"}
    }

    return {"valid": true, "user_id": session_data.user_id}
}

function generateSecureToken() {
    // Generate cryptographically secure random token
    chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
    token = ""

    for (i = 0; i < 32; i++) {
        random_index = floor(random() * len(chars))
        token = token + chars[random_index]
    }

    return token + "_" + str(date_now())
}
```

## Documentation Standards

### Code Documentation

```uddin
/**
 * Calculates the monthly payment for a loan using the standard amortization formula
 *
 * @param principal - The loan amount (must be positive number)
 * @param annual_rate - The annual interest rate as percentage (e.g., 5.5 for 5.5%)
 * @param years - The loan term in years (must be positive integer)
 * @returns Object containing monthly payment and total interest
 * @throws Error if any parameter is invalid
 *
 * @example
 * result = calculateLoanPayment(200000, 4.5, 30)
 * print("Monthly payment: $" + str(result.monthly_payment))
 * print("Total interest: $" + str(result.total_interest))
 */
function calculateLoanPayment(principal, annual_rate, years) {
    // Input validation
    if (principal <= 0) {
        throw "Principal must be positive"
    }
    if (annual_rate < 0) {
        throw "Interest rate cannot be negative"
    }
    if (years <= 0 || years != floor(years)) {
        throw "Years must be positive integer"
    }

    // Convert annual rate to monthly decimal rate
    monthly_rate = (annual_rate / 100) / 12

    // Calculate total number of payments
    total_payments = years * 12

    // Handle zero interest rate case
    if (monthly_rate == 0) {
        monthly_payment = principal / total_payments
        total_interest = 0
    } else {
        // Apply standard amortization formula: M = P * [r(1+r)^n] / [(1+r)^n - 1]
        rate_factor = pow(1 + monthly_rate, total_payments)
        monthly_payment = principal * (monthly_rate * rate_factor) / (rate_factor - 1)
        total_interest = (monthly_payment * total_payments) - principal
    }

    return {
        "monthly_payment": round(monthly_payment, 2),
        "total_interest": round(total_interest, 2),
        "total_amount": round(monthly_payment * total_payments, 2),
        "principal": principal,
        "annual_rate": annual_rate,
        "years": years
    }
}
```

### API Documentation

```uddin
/**
 * User Management API
 *
 * This module provides functions for managing user accounts, including
 * registration, authentication, profile management, and account operations.
 *
 * @version 1.2.0
 * @author Development Team
 * @since 2024-01-01
 */

/**
 * Registers a new user account
 *
 * @endpoint POST /api/users/register
 * @param user_data - User registration information
 * @param user_data.name - Full name (2-100 characters)
 * @param user_data.email - Valid email address
 * @param user_data.password - Password (min 8 chars, must include upper, lower, number)
 * @param user_data.phone - Optional phone number
 * @returns Success response with user data or error response
 *
 * @example_request
 * {
 *   "name": "John Doe",
 *   "email": "john@example.com",
 *   "password": "SecurePass123",
 *   "phone": "+1234567890"
 * }
 *
 * @example_response_success
 * {
 *   "success": true,
 *   "user": {
 *     "id": "user_123",
 *     "name": "John Doe",
 *     "email": "john@example.com",
 *     "created_at": "2024-01-15T10:30:00Z"
 *   },
 *   "token": "eyJhbGciOiJIUzI1NiIs..."
 * }
 *
 * @example_response_error
 * {
 *   "success": false,
 *   "error": "Validation failed",
 *   "details": [
 *     "Email already exists",
 *     "Password too weak"
 *   ]
 * }
 */
fun registerUser(user_data):
    // Implementation here
end
```

## Advanced Programming Patterns

### Data Structure Mastery

```uddin
// ✓ Good: Leverage Uddin-Lang's advanced data structures for complex algorithms
fun implement_advanced_task_scheduler():
    // Use Queue for FIFO task processing
    task_queue = queue_new()

    // Use Stack for undo functionality
    undo_stack = stack_new()

    // Use Set for tracking unique processed items
    processed_items = set_new()

    // Priority queue simulation using sorted array
    priority_tasks = []

    return {
        "add_task": fun(task, priority = 0):
            task_with_priority = {
                "task": task,
                "priority": priority,
                "created_at": date_now()
            }

            if (priority > 0) then:
                // Add to priority queue
                append(priority_tasks, task_with_priority)
                sort(priority_tasks)  // Sort by priority
            else:
                // Add to normal queue
                queue_enqueue(task_queue, task)
            end
        end,

        "process_next_task": fun():
            // Process priority tasks first
            if (len(priority_tasks) > 0) then:
                high_priority_task = priority_tasks[0]
                priority_tasks = slice(priority_tasks, 1, len(priority_tasks))
                return this.execute_task(high_priority_task.task)
            end

            // Process normal queue
            if (queue_size(task_queue) == 0) then:
                return {"status": "no_tasks", "message": "No tasks available"}
            end

            current_task = queue_dequeue(task_queue)

                // Check for duplicates using Set
                if (set_has(processed_items, str(current_task.id))) then:
                    return this.process_next_task()  // Skip and process next
                end            return this.execute_task(current_task)
        end,

        "execute_task": fun(task):
            try:
                result = {
                    "task_id": task.id,
                    "result": task.execute(),
                    "executed_at": date_now(),
                    "status": "completed"
                }

                // Track completion
                set_add(processed_items, str(task.id))

                // Save for undo functionality
                stack_push(undo_stack, {
                    "task": task,
                    "result": result,
                    "timestamp": date_now()
                })                return result

            catch (error):
                error_result = {
                    "task_id": task.id,
                    "error": str(error),
                    "executed_at": date_now(),
                    "status": "failed"
                }

                // Still track as processed to avoid infinite retries
                set_add(processed_items, str(task.id))

                return error_result
            end
        end,

        "undo_last_task": fun():
            if (stack_size(undo_stack) == 0) then:
                return {"error": "No tasks to undo"}
            end

            last_operation = stack_pop(undo_stack)

            // Remove from processed set
            set_remove(processed_items, str(last_operation.task.id))

            // Attempt to reverse the operation
            try:
                if (last_operation.task.undo) then:
                    undo_result = last_operation.task.undo(last_operation.result)
                    return {
                        "undone": last_operation.task.id,
                        "undo_result": undo_result
                    }
                else:
                    return {
                        "undone": last_operation.task.id,
                        "message": "Task does not support undo"
                    }
                end
            catch (error):
                return {
                    "error": "Undo failed: " + str(error),
                    "task_id": last_operation.task.id
                }
            end
        end,        "get_comprehensive_stats": fun():
            return {
                "pending_normal_tasks": queue_size(task_queue),
                "pending_priority_tasks": len(priority_tasks),
                "completed_tasks": set_size(processed_items),
                "undo_available": stack_size(undo_stack),
                "total_pending": queue_size(task_queue) + len(priority_tasks)
            }
        end
    }
end

// ✓ Good: Advanced functional composition with error handling
fun create_robust_data_pipeline():
    // Define validation stages with error recovery
    validation_stages = [
        {
            "name": "required_fields",
            "validator": fun(data):
                required = ["id", "name", "email"]
                missing = filter(required, fun(field): return not (field in data) or data[field] == null end)
                if (len(missing) > 0) then:
                    return {"valid": false, "errors": ["Missing fields: " + join(missing, ", ")]}
                end
                return {"valid": true, "data": data}
            end
        },
        {
            "name": "data_types",
            "validator": fun(data):
                errors = []
                if (typeof(data.id) != "number") then:
                    append(errors, "ID must be a number")
                end
                if (typeof(data.name) != "string" or len(trim(data.name)) == 0) then:
                    append(errors, "Name must be a non-empty string")
                end
                if (not is_regex_match(data.email, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")) then:
                    append(errors, "Invalid email format")
                end

                return len(errors) == 0 ?
                    {"valid": true, "data": data} :
                    {"valid": false, "errors": errors}
            end
        },
        {
            "name": "business_rules",
            "validator": fun(data):
                errors = []
                if (("age" in data) and data.age != null and (data.age < 13 or data.age > 120)) then:
                    append(errors, "Age must be between 13 and 120")
                end
                if (("status" in data) and data.status != null and not contains(["active", "inactive", "pending"], data.status)) then:
                    append(errors, "Invalid status value")
                end

                return len(errors) == 0 ?
                    {"valid": true, "data": data} :
                    {"valid": false, "errors": errors}
            end
        }
    ]

    // Define transformation stages
    transformation_stages = [
        {
            "name": "normalize_text",
            "transformer": fun(data):
                normalized = data
                normalized.name = trim(data.name)
                normalized.email = lower(trim(data.email))
                if (data.address) then:
                    normalized.address = trim(data.address)
                end
                return normalized
            end
        },
        {
            "name": "calculate_derived_fields",
            "transformer": fun(data):
                enhanced = data
                enhanced.email_domain = split(data.email, "@")[1]
                enhanced.name_length = len(data.name)
                enhanced.processed_at = date_now()

                // Calculate age category
                if (data.age) then:
                    if (data.age < 18) then:
                        enhanced.age_category = "minor"
                    else if (data.age < 65) then:
                        enhanced.age_category = "adult"
                    else:
                        enhanced.age_category = "senior"
                    end
                end

                return enhanced
            end
        },
        {
            "name": "apply_business_logic",
            "transformer": fun(data):
                final_data = data

                // Determine user tier based on various factors
                tier_score = 0
                if (contains(data.email_domain, "enterprise")) then:
                    tier_score = tier_score + 30
                end
                if (data.age and data.age >= 25) then:
                    tier_score = tier_score + 20
                end
                if (data.verified) then:
                    tier_score = tier_score + 25
                end

                if (tier_score >= 50) then:
                    final_data.user_tier = "premium"
                else if (tier_score >= 25) then:
                    final_data.user_tier = "standard"
                else:
                    final_data.user_tier = "basic"
                end

                return final_data
            end
        }
    ]

    return {
        "process": fun(input_data):
            processing_context = {
                "input": input_data,
                "stage_results": {},
                "errors": [],
                "warnings": []
            }

            try:
                // Validation pipeline
                current_data = input_data
                for (stage in validation_stages):
                    stage_result = stage.validator(current_data)
                    processing_context.stage_results[stage.name] = stage_result

                    if (not stage_result.valid) then:
                        processing_context.errors = append(processing_context.errors, {
                            "stage": stage.name,
                            "errors": stage_result.errors
                        })

                        return {
                            "success": false,
                            "stage_failed": stage.name,
                            "errors": processing_context.errors,
                            "context": processing_context
                        }
                    end

                    current_data = stage_result.data
                end

                // Transformation pipeline
                for (stage in transformation_stages):
                    try:
                        transformed_data = stage.transformer(current_data)
                        processing_context.stage_results[stage.name] = {
                            "success": true,
                            "data": transformed_data
                        }
                        current_data = transformed_data
                    catch (error):
                        processing_context.warnings = append(processing_context.warnings, {
                            "stage": stage.name,
                            "warning": "Transformation failed: " + str(error),
                            "data_used": current_data
                        })
                    end
                end

                return {
                    "success": true,
                    "data": current_data,
                    "context": processing_context,
                    "processed_at": date_now()
                }

            catch (error):
                return {
                    "success": false,
                    "error": "Pipeline processing failed: " + str(error),
                    "context": processing_context
                }
            end
        end,

        "process_batch": fun(input_array, batch_size = 50):
            results = {
                "successful": [],
                "failed": [],
                "total_processed": 0,
                "batch_stats": {}
            }

            total_batches = ceil(len(input_array) / batch_size)

            for (batch_num in range(total_batches)):
                start_idx = batch_num * batch_size
                end_idx = min(start_idx + batch_size, len(input_array))
                batch_data = slice(input_array, start_idx, end_idx)

                batch_results = {
                    "successful": 0,
                    "failed": 0,
                    "errors": []
                }

                for (item in batch_data):
                    result = this.process(item)
                    results.total_processed = results.total_processed + 1

                    if (result.success) then:
                        append(results.successful, result)
                        batch_results.successful = batch_results.successful + 1
                    else:
                        append(results.failed, result)
                        batch_results.failed = batch_results.failed + 1
                        append(batch_results.errors, ("error" in result and result.error != null) ? result.error : "Unknown error")
                    end
                end

                results.batch_stats["batch_" + str(batch_num + 1)] = batch_results

                // Progress reporting for large batches
                if (total_batches > 5) then:
                    progress = round((batch_num + 1) / total_batches * 100, 1)
                    print("Processed batch " + str(batch_num + 1) + "/" + str(total_batches) +
                          " (" + str(progress) + "%)")
                end
            end

            return results
        end
    }
end
```

### Modern API and Service Patterns

```uddin
// ✓ Good: Production-ready service with comprehensive features
fun create_enterprise_user_service():
    // Service configuration
    config = {
        "max_login_attempts": 5,
        "lockout_duration": 15 * 60 * 1000,  // 15 minutes
        "session_timeout": 24 * 60 * 60 * 1000,  // 24 hours
        "password_reset_expiry": 60 * 60 * 1000  // 1 hour
    }

    // In-memory stores (in production, use proper databases)
    user_store = {}
    session_store = {}
    failed_attempts = {}

    return {
        "create_user": fun(user_data):
            try:
                // Comprehensive validation
                validation = validate_user_registration(user_data)
                if (not validation.valid) then:
                    return {
                        "success": false,
                        "error": "Validation failed",
                        "details": validation.errors,
                        "error_code": "VALIDATION_ERROR"
                    }
                end

                // Check for existing user
                if (user_store[lower(user_data.email)]) then:
                    return {
                        "success": false,
                        "error": "User already exists",
                        "error_code": "USER_EXISTS"
                    }
                end

                // Create user with enhanced security
                user_id = "uid_" + str(date_now()) + "_" + str(len(user_data.email))
                salt = "salt_" + str(date_now())
                password_hash = "hash_" + user_data.password + "_" + salt

                user = {
                    "id": user_id,
                    "email": lower(trim(user_data.email)),
                    "name": user_data.name,
                    "password_hash": password_hash,
                    "salt": salt,
                    "created_at": date_now(),
                    "last_login": null,
                    "status": "active",
                    "email_verified": false,
                    "profile": {
                        "first_name": "",
                        "last_name": "",
                        "phone": "",
                        "preferences": {
                            "theme": "light",
                            "notifications": true,
                            "language": "en"
                        }
                    },
                    "security": {
                        "failed_login_attempts": 0,
                        "locked_until": null,
                        "password_changed_at": date_now(),
                        "two_factor_enabled": false
                    }
                }

                user_store[user.email] = user

                // Schedule welcome email and verification
                print("Scheduling email verification for: " + user.email)
                print("Scheduling welcome email for: " + user.name)

                // Log user creation for audit
                print("[SECURITY] USER_CREATED: " + user_id + " (" + user.email + ")")                return {
                    "success": true,
                    "user": {
                        "id": user.id,
                        "email": user.email,
                        "name": user.name,
                        "created_at": user.created_at,
                        "status": user.status
                    },
                    "message": "User created successfully. Please check your email for verification."
                }

            catch (error):
                log_error("User creation failed", {
                    "error": str(error),
                    "email": user_data.email,
                    "timestamp": date_now()
                })

                return {
                    "success": false,
                    "error": "Internal server error",
                    "error_code": "INTERNAL_ERROR"
                }
            end
        end,

        "authenticate_user": fun(email, password, additional_context = {}):
            try:
                email = lower(trim(email))

                // Check for missing credentials
                if (not email or not password) then:
                    return {
                        "success": false,
                        "error": "Email and password are required",
                        "error_code": "MISSING_CREDENTIALS"
                    }
                end

                // Find user
                user = user_store[email]
                if (not user) then:
                    // Log failed attempt even for non-existent users
                    log_security_event("LOGIN_FAILED", {
                        "email": email,
                        "reason": "user_not_found",
                        "ip_address": get_current_ip()
                    })

                    return {
                        "success": false,
                        "error": "Invalid credentials",
                        "error_code": "INVALID_CREDENTIALS"
                    }
                end

                // Check if account is locked
                if (user.security.locked_until and date_now() < user.security.locked_until) then:
                    return {
                        "success": false,
                        "error": "Account temporarily locked due to too many failed attempts",
                        "error_code": "ACCOUNT_LOCKED",
                        "locked_until": user.security.locked_until
                    }
                end

                // Verify password
                if (not verify_password_with_salt(password, user.password_hash, user.salt)) then:
                    // Increment failed attempts
                    user.security.failed_login_attempts = user.security.failed_login_attempts + 1

                    // Lock account if too many failures
                    if (user.security.failed_login_attempts >= config.max_login_attempts) then:
                        user.security.locked_until = date_now() + config.lockout_duration
                    end

                    log_security_event("LOGIN_FAILED", {
                        "user_id": user.id,
                        "email": email,
                        "reason": "invalid_password",
                        "attempts": user.security.failed_login_attempts,
                        "ip_address": get_current_ip()
                    })

                    return {
                        "success": false,
                        "error": "Invalid credentials",
                        "error_code": "INVALID_CREDENTIALS"
                    }
                end

                // Successful authentication

                // Reset failed attempts
                user.security.failed_login_attempts = 0
                user.security.locked_until = null
                user.last_login = date_now()

                // Generate session
                session_data = {
                    "user_id": user.id,
                    "email": user.email,
                    "created_at": date_now(),
                    "expires_at": date_now() + config.session_timeout,
                    "ip_address": get_current_ip(),
                    "user_agent": get_current_user_agent(),
                    "additional_context": additional_context
                }

                session_token = generate_secure_session_token(user.id)
                session_store[session_token] = session_data

                log_security_event("LOGIN_SUCCESS", {
                    "user_id": user.id,
                    "email": email,
                    "session_token": session_token,
                    "ip_address": get_current_ip()
                })

                return {
                    "success": true,
                    "user": {
                        "id": user.id,
                        "email": user.email,
                        "name": user.name,
                        "last_login": user.last_login,
                        "profile": user.profile
                    },
                    "session": {
                        "token": session_token,
                        "expires_at": session_data.expires_at
                    }
                }

            catch (error):
                log_error("Authentication error", {
                    "error": str(error),
                    "email": email
                })

                return {
                    "success": false,
                    "error": "Authentication service unavailable",
                    "error_code": "SERVICE_ERROR"
                }
            end
        end,

        "validate_session": fun(session_token):
            if (not session_token) then:
                return {"valid": false, "error": "No session token provided"}
            end

            session = session_store[session_token]
            if (not session) then:
                return {"valid": false, "error": "Invalid session token"}
            end

            // Check expiration
            if (date_now() > session.expires_at) then:
                // Remove expired session by rebuilding session store
                new_session_store = {}
                for (token in session_store):
                    if (token != session_token) then:
                        new_session_store[token] = session_store[token]
                    end
                end
                session_store = new_session_store
                return {"valid": false, "error": "Session expired"}
            end

            // Optional: Extend session on activity
            session.expires_at = date_now() + config.session_timeout

            return {
                "valid": true,
                "user_id": session.user_id,
                "session": session
            }
        end,

        "logout": fun(session_token):
            if (session_store[session_token]) then:
                session = session_store[session_token]

                log_security_event("LOGOUT", {
                    "user_id": session.user_id,
                    "session_token": session_token,
                    "ip_address": get_current_ip()
                })

                // Remove session by rebuilding session store
                new_session_store = {}
                for (token in session_store):
                    if (token != session_token) then:
                        new_session_store[token] = session_store[token]
                    end
                end
                session_store = new_session_store
                return {"success": true, "message": "Logged out successfully"}
            end

            return {"success": false, "error": "Invalid session"}
        end,

        "get_user_profile": fun(user_id):
            user = find_user_by_id(user_id)
            if (not user) then:
                return {"success": false, "error": "User not found"}
            end

            return {
                "success": true,
                "user": {
                    "id": user.id,
                    "email": user.email,
                    "name": user.name,
                    "profile": user.profile,
                    "created_at": user.created_at,
                    "last_login": user.last_login,
                    "status": user.status
                }
            }
        end,

        "update_user_profile": fun(user_id, updates):
            user = find_user_by_id(user_id)
            if (not user) then:
                return {"success": false, "error": "User not found"}
            end

            // Validate updates
            allowed_fields = ["first_name", "last_name", "phone", "preferences"]
            for (field in updates):
                if (not contains(allowed_fields, field)) then:
                    return {
                        "success": false,
                        "error": "Field '" + field + "' is not allowed to be updated"
                    }
                end
            end

            // Apply updates
            for (field in updates):
                if (field == "preferences") then:
                    // Merge preferences instead of replacing
                    for (pref in updates[field]):
                        user.profile.preferences[pref] = updates[field][pref]
                    end
                else:
                    user.profile[field] = sanitize_user_input(updates[field], "alphanumeric")
                end
            end

            user_store[user.email] = user

            return {
                "success": true,
                "user": {
                    "id": user.id,
                    "profile": user.profile
                }
            }
        end
    }
end

// Helper functions for the service
fun generate_secure_id():
    timestamp = str(date_now())
    random_part = generate_random_string(16)
    return "uid_" + timestamp + "_" + random_part
end

fun generate_secure_session_token(user_id):
    timestamp = str(date_now())
    random_part = generate_random_string(32)
    return "sess_" + str(user_id) + "_" + timestamp + "_" + random_part
end

fun hash_password_with_salt(password, salt):
    // In production, use proper hashing like bcrypt
    return "sha256_" + password + "_" + salt + "_hashed"
end

fun verify_password_with_salt(password, hash, salt):
    expected_hash = hash_password_with_salt(password, salt)
    return hash == expected_hash
end

fun find_user_by_id(user_id):
    for (email in user_store):
        user = user_store[email]
        if (user.id == user_id) then:
            return user
        end
    end
    return null
end

fun log_security_event(event_type, details):
    log_entry = {
        "event_type": event_type,
        "timestamp": date_now(),
        "details": details
    }

    // In production, send to security monitoring system
    print("[SECURITY] " + event_type + ": " + json_stringify(details))
end

fun get_current_ip():
    // In production, get actual client IP
    return "127.0.0.1"
end

fun get_current_user_agent():
    // In production, get actual user agent
    return "UddinLang-Client/1.0"
end

fun schedule_email_verification(email, user_id):
    // In production, queue email job
    print("Email verification scheduled for: " + email)
end

fun schedule_welcome_email(email, name):
    // In production, queue welcome email
    print("Welcome email scheduled for: " + name + " (" + email + ")")
end
```

## Conclusion

This comprehensive best practices guide covers all aspects of professional Uddin-Lang development, from basic style conventions to advanced architectural patterns. By following these practices, you will create:

-   **Maintainable Code**: Clear structure, consistent naming, and comprehensive documentation
-   **Secure Applications**: Proper input validation, authentication, and error handling
-   **High-Performance Systems**: Efficient algorithms, memory management, and caching strategies
-   **Robust Services**: Comprehensive error handling, logging, and testing
-   **Scalable Architecture**: Modular design, separation of concerns, and advanced patterns

Remember that best practices evolve with the language and your project requirements. Regularly review and update your coding standards as Uddin-Lang continues to grow and your applications become more sophisticated.

### Key Takeaways

1. **Consistency is King**: Establish and maintain consistent coding standards across your project
2. **Security First**: Always validate inputs, handle errors gracefully, and implement proper authentication
3. **Performance Matters**: Use appropriate data structures and algorithms for your use case
4. **Test Everything**: Comprehensive testing prevents bugs and enables confident refactoring
5. **Document Extensively**: Good documentation saves time and enables team collaboration
6. **Learn Continuously**: Stay updated with Uddin-Lang features and modern programming practices

With these practices in place, your Uddin-Lang applications will be production-ready, maintainable, and built to last.
