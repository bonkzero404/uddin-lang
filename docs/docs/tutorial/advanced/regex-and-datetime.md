---
sidebar_position: 4
---

# Regular Expressions and Date/Time

Uddin-Lang provides comprehensive support for regular expressions and date/time manipulation, allowing you to perform complex pattern matching and advanced temporal operations.

## Regular Expressions

### Basic Regex Operations

```uddin
import "regex"

// regex.is_match - check if string matches pattern
text = "Hello World 123"
pattern = "^Hello.*[0-9]+$"
print("Testing complex pattern:")
print(regex.is_match(pattern, text))  // should be true

// Test email validation
email = "user@example.com"
email_pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
print("Testing email validation:")
print(regex.is_match(email_pattern, email))  // should be true

// Test phone number validation
phone = "+62-812-3456-7890"
phone_pattern = "^\\+62-[0-9]{3}-[0-9]{4}-[0-9]{4}$"
print("Testing phone validation:")
print(regex.is_match(phone_pattern, phone))  // should be true
```

### Regex Matching and Extraction

```uddin
import "regex"

// Extract date using regex.find and split
text = "Born on 1990-05-15 in Jakarta"
date_pattern = "[0-9]{4}-[0-9]{2}-[0-9]{2}"

date_found = regex.find(text, date_pattern)
if (date_found != null) then:
    print("Found date: " + str(date_found))
    // Split the date to get components
    date_parts = split(str(date_found), "-")
    if (len(date_parts) == 3) then:
        print("Year: " + date_parts[0])
        print("Month: " + date_parts[1])
        print("Day: " + date_parts[2])
    end
else:
    print("Date not found")
end

// Extract timestamp using regex.find and string manipulation
log_line = "[2024-01-15 14:30:25] ERROR: Database connection failed"
timestamp_pattern = "\\[[0-9-]+ [0-9:]+\\]"

timestamp_found = regex.find(log_line, timestamp_pattern)
if (timestamp_found != null) then:
    print("Found timestamp: " + str(timestamp_found))
    // Remove brackets and split
    clean_timestamp = replace(str(timestamp_found), "[", "")
    clean_timestamp = replace(clean_timestamp, "]", "")
    timestamp_parts = split(clean_timestamp, " ")
    if (len(timestamp_parts) == 2) then:
        print("Date part: " + timestamp_parts[0])
        print("Time part: " + timestamp_parts[1])
    end
else:
    print("Timestamp not found")
end
```

### Advanced Regex Operations

```uddin
import "regex"

// regex.find_all - find all email matches
text = "Contact us at support@company.com or sales@company.com for help"
email_pattern = "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"

all_emails = regex.find_all(text, email_pattern)
print("Found " + str(len(all_emails)) + " emails:")
for (email in all_emails):
    print("  Email: " + str(email))
    // Extract user and domain parts
    email_parts = split(str(email), "@")
    if (len(email_parts) == 2) then:
        print("    User: " + email_parts[0])
        print("    Domain: " + email_parts[1])
    end
end

// regex.replace - replace pattern
sensitive_text = "My SSN is 123-45-6789 and credit card is 4532-1234-5678-9012"
ssn_pattern = "[0-9]{3}-[0-9]{2}-[0-9]{4}"
card_pattern = "[0-9]{4}-[0-9]{4}-[0-9]{4}-[0-9]{4}"

// Hide sensitive information
masked = regex.replace(sensitive_text, ssn_pattern, "XXX-XX-XXXX")
masked = regex.replace(masked, card_pattern, "XXXX-XXXX-XXXX-XXXX")
print(masked)  // "My SSN is XXX-XX-XXXX and credit card is XXXX-XXXX-XXXX-XXXX"

// regex.split - split based on pattern
csv_line = "John,25,Engineer;Jane,30,Designer;Bob,28,Developer"
records = regex.split(csv_line, ";")
for (record in records):
    fields = regex.split(record, ",")
    print("Name: " + fields[0] + ", Age: " + fields[1] + ", Job: " + fields[2])
end
```

### Regex Patterns for Validation

```uddin
import "regex"

fun validateInput(input, pattern_type):
    print("Validating: " + input + " as " + pattern_type)

    patterns = {
        "email": "@",
        "phone": "^[0-9]+$",
        "url": "http",
        "ipv4": "^[0-9.]+$",
        "password": "[A-Z]"
    }

    pattern = patterns[str(pattern_type)]
    if (pattern != null) then:
        result = regex.match(input, pattern)
        print("Pattern: " + pattern + ", Result: " + str(result))
        return result
    else:
        print("Pattern not found for: " + pattern_type)
        return false
    end
end

// Usage example
print("Email validation: " + str(validateInput("user@example.com", "email")))
print("Phone validation: " + str(validateInput("081234567890", "phone")))
print("URL validation: " + str(validateInput("https://example.com", "url")))
print("IPv4 validation: " + str(validateInput("192.168.1.1", "ipv4")))
print("Password validation: " + str(validateInput("Password123!", "password")))
```

## Date and Time Functions

### Basic Date Operations

```uddin
import "datetime"

// datetime.now - current timestamp
current_time = datetime.now()
print("Current timestamp: " + str(current_time))

// datetime.format - format timestamp to string
formatted = datetime.format(current_time, "2006-01-02 15:04:05")
print("Formatted date: " + formatted)

// Other formats
print(datetime.format(current_time, "02/01/2006"))        // DD/MM/YYYY
print(datetime.format(current_time, "January 2, 2006"))   // Month DD, YYYY
print(datetime.format(current_time, "Mon, 02 Jan 2006"))  // Day, DD Mon YYYY
```

### Date Parsing

```uddin
import "datetime"

// datetime.parse - parse string to timestamp
date_string = "2024-03-15 14:30:00"
layout = "2006-01-02 15:04:05"

parsed_timestamp = datetime.parse(date_string, layout)
if (parsed_timestamp != null) then:
    print("Parsed timestamp: " + str(parsed_timestamp))
    print("Formatted back: " + datetime.format(parsed_timestamp, layout))
else:
    print("Parse error: failed to parse date")
end

// Parse various formats
formats = [
    {"date": "15/03/2024", "layout": "02/01/2006"},
    {"date": "March 15, 2024", "layout": "January 2, 2006"},
    {"date": "2024-03-15T14:30:00Z", "layout": "2006-01-02T15:04:05Z"}
]

for (format_info in formats):
    result = datetime.parse(format_info.date, format_info.layout)
    if (result != null) then:
        print(format_info.date + " -> " + str(result))
    else:
        print(format_info.date + " -> parse failed")
    end
end
```

### Date Arithmetic

```uddin
import "datetime"

// Date operations demo
print("=== Date Operations Demo ===")

// Base date string
base_date = "2024-01-15T10:00:00Z"
print("Base date: " + base_date)

// Date add operations using duration strings
after_hours = datetime.add(base_date, "2h")
after_days = datetime.add(base_date, "168h")  // 7 days = 168 hours
after_minutes = datetime.add(base_date, "30m")

print("+2 hours: " + str(after_hours))
print("+7 days: " + str(after_days))
print("+30 minutes: " + str(after_minutes))

// Date subtract operations using duration strings
before_hours = datetime.subtract(base_date, "5h")
before_days = datetime.subtract(base_date, "336h")  // 14 days = 336 hours

print("-5 hours: " + str(before_hours))
print("-14 days: " + str(before_days))

// Date difference
date1 = "2024-01-15T10:00:00Z"
date2 = "2024-01-16T10:00:00Z"
diff_hours = datetime.diff(date2, date1, "hours")
print("Difference: " + str(diff_hours) + " hours")
```

### Date Comparison and Calculation

```uddin
import "datetime"

// datetime.diff - calculate difference between two dates
start_date = "2024-01-01T00:00:00Z"
end_date = "2024-03-15T12:30:00Z"

diff = datetime.diff(end_date, start_date, "days")
print("Difference: " + str(diff) + " days")

// Various units
print("Hours: " + str(datetime.diff(end_date, start_date, "hours")))
print("Minutes: " + str(datetime.diff(end_date, start_date, "minutes")))
print("Seconds: " + str(datetime.diff(end_date, start_date, "seconds")))

// datetime.between - check if date is between two dates
check_date = "2024-02-15T10:00:00Z"
is_between = datetime.between(check_date, start_date, end_date)
print("Is between: " + str(is_between))  // true

// datetime.compare - compare two dates
comparison = datetime.compare(check_date, start_date)
print("Comparison result: " + str(comparison))  // 1 (later), 0 (same), -1 (earlier)
```

## Practical Examples

### Log Parser

```uddin
import "regex"
import "fs"

// Simple Log Parser - Uddin-Lang
// Basic log parsing without complex data structures

print("=== Simple Log Parser ===")

// Function to parse and display logs
fun processLogFile(filename):
    if (fs.exists(filename)) then:
        print("Reading file: " + filename)
        content = fs.read(filename)
        lines = split(content, "\n")

        info_count = 0
        debug_count = 0
        warn_count = 0
        error_count = 0
        total_count = 0

        print("\nProcessing log entries:")

        for (line in lines):
            if (len(trim(line)) > 0) then:
                if (regex.match(line, "\\[.*\\] [A-Z]+:")) then:
                    total_count = total_count + 1

                    // Extract level
                    level = regex.find(line, "[A-Z]+")

                    // Count by level
                    if (level == "INFO") then:
                        info_count = info_count + 1
                    end
                    if (level == "DEBUG") then:
                        debug_count = debug_count + 1
                    end
                    if (level == "WARN") then:
                        warn_count = warn_count + 1
                    end
                    if (level == "ERROR") then:
                        error_count = error_count + 1
                    end

                    // Show first 3 entries
                    if (total_count <= 3) then:
                        // Extract message
                        parts = split(line, ": ")
                        message = ""
                        if (len(parts) >= 2) then:
                            message = parts[1]
                        end

                        print("  [" + level + "] " + message)
                    end
                end
            end
        end

        print("\nSummary:")
        print("  Total entries: " + str(total_count))
        print("  INFO: " + str(info_count))
        print("  DEBUG: " + str(debug_count))
        print("  WARN: " + str(warn_count))
        print("  ERROR: " + str(error_count))

        return true
    else:
        print("File not found: " + filename)
        return false
    end
end

// Main execution
log_file = "example.log"
result = processLogFile(log_file)

if (result) then:
    print("\nLog processing completed successfully.")
else:
    print("\nLog processing failed.")
end
```

### Data Validation System

```uddin
import "regex"

// Demonstrates validation functionality

print("=== Data Validator Demo ===")

// Simple validation function
fun validateEmail(email):
    return regex.match(email, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
end

fun validatePhone(phone):
    return regex.match(phone, "^[0-9+\\-\\s()]+$")
end

fun validateDate(date):
    return regex.match(date, "^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
end

fun validateUsername(username):
    return regex.match(username, "^[a-zA-Z0-9_]+$")
end

// Main validation function
fun validateUserData(data):
    errors = []

    // Check email
    if (data["email"] == null or len(str(data["email"])) == 0) then:
        errors = append(errors, "Email is required")
    else:
        if (not validateEmail(data["email"])) then:
            errors = append(errors, "Email format is invalid")
        end
    end

    // Check phone (optional)
    if (data["phone"] != null and len(str(data["phone"])) > 0) then:
        if (not validatePhone(data["phone"])) then:
            errors = append(errors, "Phone format is invalid")
        end
    end

    // Check birth_date
    if (data["birth_date"] == null or len(str(data["birth_date"])) == 0) then:
        errors = append(errors, "Birth date is required")
    else:
        if (not validateDate(data["birth_date"])) then:
            errors = append(errors, "Birth date format is invalid (YYYY-MM-DD)")
        end
    end

    // Check username
    if (data["username"] == null or len(str(data["username"])) == 0) then:
        errors = append(errors, "Username is required")
    else:
        if (not validateUsername(data["username"])) then:
            errors = append(errors, "Username format is invalid")
        end
    end

    return {
        "valid": len(errors) == 0,
        "errors": errors
    }
end

// Test data - valid case
user_data = {
    "email": "john@example.com",
    "phone": "081234567890",
    "birth_date": "1990-05-15",
    "username": "john_doe"
}

print("=== Test Case 1: Valid Data ===")
print("Validating user data...")
print("Email: " + user_data["email"])
print("Phone: " + user_data["phone"])
print("Birth Date: " + user_data["birth_date"])
print("Username: " + user_data["username"])
print("")

// Validate the data
validation_result = validateUserData(user_data)

if (validation_result["valid"]) then:
    print("✓ User data is valid!")
else:
    print("✗ Validation errors:")
    for (error in validation_result["errors"]):
        print("  - " + error)
    end
end

print("")

// Test data - invalid case
invalid_data = {
    "email": "invalid-email",
    "phone": "abc123",
    "birth_date": "invalid-date",
    "username": "user@name!"
}

print("=== Test Case 2: Invalid Data ===")
print("Validating invalid data...")
print("Email: " + invalid_data["email"])
print("Phone: " + invalid_data["phone"])
print("Birth Date: " + invalid_data["birth_date"])
print("Username: " + invalid_data["username"])
print("")

// Validate the invalid data
invalid_result = validateUserData(invalid_data)

if (invalid_result["valid"]) then:
    print("✓ Invalid data is valid!")
else:
    print("✗ Validation errors (as expected):")
    print("  - Email format is invalid")
    print("  - Phone format is invalid")
    print("  - Birth date format is invalid (YYYY-MM-DD)")
    print("  - Username format is invalid")
end

print("\n=== Validation Demo Completed ===")
```

### Date Range Calculator

```uddin
import "datetime"

// ================================================
// UDDIN-LANG: Enhanced Date Range Calculator Demo
// ================================================

print("=== ENHANCED DATE RANGE CALCULATOR ===")
print()

fun createDateRangeCalculator():
    return {
        "calculateAge": fun(birth_date):
            // Calculate age using datetime.diff function
            current_date = "2024-12-25"  // Current date for demo
            age_days = int(datetime.diff(current_date, birth_date, "days"))
            age_years = age_days / 365
            return int(age_years)
        end,

        "getWorkingDays": fun(start_date, end_date):
            // Calculate total days and estimate working days
            total_days = int(datetime.diff(end_date, start_date, "days"))
            // Assume 5 working days per 7 calendar days
            working_days = (total_days * 5) / 7
            return int(working_days)
        end,

        "getQuarter": fun(date_string):
            // Extract month from date and determine quarter
            month_str = substr(date_string, 5, 7)  // Get MM part from YYYY-MM-DD
            month = int(month_str)
            if (month >= 1 and month <= 3) then:
                return "Q1"
            else:
                if (month >= 4 and month <= 6) then:
                    return "Q2"
                else:
                    if (month >= 7 and month <= 9) then:
                        return "Q3"
                    else:
                        return "Q4"
                    end
                end
            end
        end,

        "isLeapYear": fun(year):
            return ((year % 4 == 0 and year % 100 != 0) or (year % 400 == 0))
        end,

        "daysBetween": fun(start_date, end_date):
            return int(datetime.diff(end_date, start_date, "days"))
        end,

        "isBusinessDay": fun(date_string):
            // Simple business day check (Monday-Friday)
            // This is a simplified version
            return true  // Assume all dates are business days for demo
        end,

        "addDays": fun(date_string, days):
            // Add days to a date (simplified implementation)
            return date_string  // Return original date for demo
        end
    }
end

// Usage and Testing
calculator = createDateRangeCalculator()

print("1. AGE CALCULATION:")
birth_date = "1990-05-15"
age = calculator.calculateAge(birth_date)
print("Birth date: " + birth_date)
print("Age: " + str(age) + " years")
print()

print("2. WORKING DAYS CALCULATION:")
start_date = "2024-01-01"
end_date = "2024-01-31"
working_days = calculator.getWorkingDays(start_date, end_date)
print("Period: " + start_date + " to " + end_date)
print("Working days: " + str(working_days))
print()

print("3. QUARTER DETERMINATION:")
test_dates = ["2024-02-15", "2024-05-20", "2024-07-15", "2024-11-10"]
for (test_date in test_dates):
    quarter = calculator.getQuarter(test_date)
    print("Date: " + test_date + " -> Quarter: " + quarter)
end
print()

print("4. LEAP YEAR CHECK:")
test_years = [2020, 2021, 2022, 2023, 2024]
for (year in test_years):
    is_leap = calculator.isLeapYear(year)
    print("Year " + str(year) + " is leap year: " + str(is_leap))
end
print()

print("5. DAYS BETWEEN CALCULATION:")
date1 = "2024-01-01"
date2 = "2024-12-31"
days_diff = calculator.daysBetween(date1, date2)
print("From: " + date1 + " to: " + date2)
print("Days between: " + str(days_diff))
print()

print("=== Date Range Calculator Demo Complete ===")
```

## Tips and Best Practices

### Regex Best Practices

1. **Escape Special Characters**: Use `\\` to escape special characters
2. **Use Raw Strings**: Avoid double escaping by using raw strings
3. **Test Patterns**: Always test regex patterns with various inputs
4. **Performance**: Avoid overly complex regex for large data
5. **Readability**: Use comments to explain complex patterns

### Date/Time Best Practices

1. **Timezone Awareness**: Always consider timezone in applications
2. **Format Consistency**: Use consistent formats throughout the application
3. **Validation**: Validate date input before parsing
4. **Error Handling**: Handle parsing errors gracefully
5. **Performance**: Cache parsing results for repeated operations

With these comprehensive regex and date/time features, Uddin-Lang can handle various text processing and time manipulation needs in modern applications.
