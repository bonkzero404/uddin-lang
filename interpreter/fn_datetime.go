package interpreter

import (
	"fmt"
	"strings"
	"time"
)

// datenowFunc implements the date_now() built-in function
// Returns the current date and time in RFC3339 format
// Parameters: none
// Returns the current date/time as a string
// Example: date_now() -> "2020-01-02T15:04:05Z"
func datenowFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "date_now", args, 0); err != Value(nil) {
		return err
	}
	// Get current time and format as RFC3339
	return Value(time.Now().Format(time.RFC3339))
}

// timenowFunc implements the time_now() built-in function
// Returns the current Unix timestamp in milliseconds
// Parameters: none
// Returns the current timestamp as an integer (milliseconds since Unix epoch)
// Example: time_now() -> 1640995445123
func timenowFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "time_now", args, 0); err != Value(nil) {
		return err
	}
	// Get current time in milliseconds since Unix epoch
	// Convert to int to match Uddin-Lang's integer type
	return Value(int(time.Now().UnixMilli()))
}

// sleepFunc implements the sleep() built-in function
// Pauses execution for the specified number of milliseconds
// Parameters:
//   - milliseconds: The number of milliseconds to sleep (int)
//
// Returns null
// Example: sleep(1000) -> sleeps for 1 second
func sleepFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "sleep() requires exactly 1 argument, got %d", len(args)))
	}

	milliseconds, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "sleep() requires an integer argument (milliseconds)"))
	}

	if milliseconds < 0 {
		panic(typeError(pos, "sleep() requires a non-negative integer"))
	}

	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	return Value(nil)
}

// dateformatFunc implements the date_format() built-in function
// Formats a date string according to a specified layout
// Parameters:
//   - t: Date string in RFC3339 format
//   - layout: Format string with placeholders (YYYY, MM, DD, etc.)
//
// Returns the formatted date string, or null if parsing fails
// Example: date_format("2020-01-02T15:04:05Z", "YYYY-MM-DD hh:mm:ss") -> "2020-01-02 15:04:05"
func dateformatFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "date_format", args, 2); err != Value(nil) {
		return err
	}
	if t, ok := args[0].(string); ok {
		if layout, ok := args[1].(string); ok {
			// Define replacements for date format placeholders
			replacer := strings.NewReplacer(
				"YYYY", "2006", // Year
				"MM", "01", // Month (numeric)
				"DD", "02", // Day
				"hh", "15", // Hour (24-hour)
				"mm", "04", // Minute
				"ss", "05", // Second
				"ee", "Mon", // Weekday (short)
				"EE", "Monday", // Weekday (long)
				"nn", "Jan", // Month (short)
				"NN", "January", // Month (long)
			)

			// Parse the input date string
			parsed, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return Value(nil) // Return null if parsing fails
			}

			// Replace placeholders with Go's time format specifiers
			layout = replacer.Replace(layout)
			return Value(parsed.Format(layout))
		}
		panic(typeError(pos, "date_format() requires second argument to be a string"))
	}
	panic(typeError(pos, "date_format() requires first argument to be a string"))
}

// dateParseFunc parses a date string according to the specified format
// Usage: date_parse("2023-12-25", "2006-01-02")
func dateParseFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		return Value(fmt.Errorf("date_parse() requires exactly 2 arguments, got %d", len(args)))
	}

	dateStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_parse() requires first argument to be a string"))
	}

	format, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("date_parse() requires second argument to be a string"))
	}

	// Parse the date
	parsedTime, err := time.Parse(format, dateStr)
	if err != nil {
		return Value(fmt.Errorf("date_parse() failed to parse date: %v", err))
	}

	// Return as ISO 8601 format string
	return Value(parsedTime.Format(time.RFC3339))
}

// dateFormatEnhancedFunc formats a date string from one format to another
// Usage: date_format_new("2023-12-25T10:30:00Z", "2006-01-02T15:04:05Z07:00", "January 2, 2006 3:04 PM")
func dateFormatEnhancedFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 3 {
		return Value(fmt.Errorf("date_format_new() requires exactly 3 arguments, got %d", len(args)))
	}

	dateStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_format_new() requires first argument to be a string"))
	}

	inputFormat, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("date_format_new() requires second argument to be a string"))
	}

	outputFormat, ok := args[2].(string)
	if !ok {
		return Value(fmt.Errorf("date_format_new() requires third argument to be a string"))
	}

	// Parse the input date
	parsedTime, err := time.Parse(inputFormat, dateStr)
	if err != nil {
		return Value(fmt.Errorf("date_format_new() failed to parse date: %v", err))
	}

	// Format to output format
	return Value(parsedTime.Format(outputFormat))
}

// dateAddFunc adds time duration to a date
// Usage: date_add("2023-12-25T10:30:00Z", "24h") or date_add("2023-12-25T10:30:00Z", "hours", 24)
func dateAddFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("date_add requires at least 2 arguments"))
	}

	dateStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_add: first argument must be a date string"))
	}

	// Parse the date (try multiple formats)
	var parsedTime time.Time
	var err error
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}

	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_add: failed to parse date: %v", err))
	}

	if len(args) == 2 {
		// Duration string format
		durationStr, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("date_add: second argument must be a duration string"))
		}

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return Value(fmt.Errorf("date_add: invalid duration format: %v", err))
		}

		result := parsedTime.Add(duration)
		return Value(result.Format(time.RFC3339))
	} else if len(args) == 3 {
		// Unit and amount format
		unit, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("date_add: second argument must be a unit string"))
		}

		var amount int64
		switch v := args[2].(type) {
		case int64:
			amount = v
		case float64:
			amount = int64(v)
		default:
			return Value(fmt.Errorf("date_add: third argument must be a number"))
		}

		var result time.Time
		switch strings.ToLower(unit) {
		case "years", "year":
			result = parsedTime.AddDate(int(amount), 0, 0)
		case "months", "month":
			result = parsedTime.AddDate(0, int(amount), 0)
		case "days", "day":
			result = parsedTime.AddDate(0, 0, int(amount))
		case "hours", "hour":
			result = parsedTime.Add(time.Duration(amount) * time.Hour)
		case "minutes", "minute":
			result = parsedTime.Add(time.Duration(amount) * time.Minute)
		case "seconds", "second":
			result = parsedTime.Add(time.Duration(amount) * time.Second)
		default:
			return Value(fmt.Errorf("date_add: unsupported unit '%s'", unit))
		}

		return Value(result.Format(time.RFC3339))
	}

	return Value(fmt.Errorf("date_add: invalid number of arguments"))
}

// dateSubtractFunc subtracts time duration from a date
// Usage: date_subtract("2023-12-25T10:30:00Z", "24h") or date_subtract("2023-12-25T10:30:00Z", "hours", 24)
func dateSubtractFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("date_subtract requires at least 2 arguments"))
	}

	dateStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_subtract: first argument must be a date string"))
	}

	// Parse the date (try multiple formats)
	var parsedTime time.Time
	var err error
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}

	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_subtract: failed to parse date: %v", err))
	}

	if len(args) == 2 {
		// Duration string format
		durationStr, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("date_subtract: second argument must be a duration string"))
		}

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return Value(fmt.Errorf("date_subtract: invalid duration format: %v", err))
		}

		result := parsedTime.Add(-duration)
		return Value(result.Format(time.RFC3339))
	} else if len(args) == 3 {
		// Unit and amount format
		unit, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("date_subtract: second argument must be a unit string"))
		}

		var amount int64
		switch v := args[2].(type) {
		case int64:
			amount = v
		case float64:
			amount = int64(v)
		default:
			return Value(fmt.Errorf("date_subtract: third argument must be a number"))
		}

		var result time.Time
		switch strings.ToLower(unit) {
		case "years", "year":
			result = parsedTime.AddDate(-int(amount), 0, 0)
		case "months", "month":
			result = parsedTime.AddDate(0, -int(amount), 0)
		case "days", "day":
			result = parsedTime.AddDate(0, 0, -int(amount))
		case "hours", "hour":
			result = parsedTime.Add(-time.Duration(amount) * time.Hour)
		case "minutes", "minute":
			result = parsedTime.Add(-time.Duration(amount) * time.Minute)
		case "seconds", "second":
			result = parsedTime.Add(-time.Duration(amount) * time.Second)
		default:
			return Value(fmt.Errorf("date_subtract: unsupported unit '%s'", unit))
		}

		return Value(result.Format(time.RFC3339))
	}

	return Value(fmt.Errorf("date_subtract: invalid number of arguments"))
}

// dateDiffFunc calculates the difference between two dates
// Usage: date_diff("2023-12-25T10:30:00Z", "2023-12-24T10:30:00Z", "hours")
func dateDiffFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("date_diff requires at least 2 arguments"))
	}

	date1Str, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_diff: first argument must be a date string"))
	}

	date2Str, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("date_diff: second argument must be a date string"))
	}

	// Parse both dates
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}

	var date1, date2 time.Time
	var err error

	for _, format := range formats {
		date1, err = time.Parse(format, date1Str)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_diff: failed to parse first date: %v", err))
	}

	for _, format := range formats {
		date2, err = time.Parse(format, date2Str)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_diff: failed to parse second date: %v", err))
	}

	diff := date1.Sub(date2)

	if len(args) == 2 {
		// Return difference in seconds as default
		return Value(int64(diff.Seconds()))
	}

	// Return difference in specified unit
	unit, ok := args[2].(string)
	if !ok {
		return Value(fmt.Errorf("date_diff: third argument must be a unit string"))
	}

	switch strings.ToLower(unit) {
	case "nanoseconds", "nanosecond":
		return Value(int64(diff.Nanoseconds()))
	case "microseconds", "microsecond":
		return Value(int64(diff.Nanoseconds() / 1000))
	case "milliseconds", "millisecond":
		return Value(int64(diff.Nanoseconds() / 1000000))
	case "seconds", "second":
		return Value(int64(diff.Seconds()))
	case "minutes", "minute":
		return Value(int64(diff.Minutes()))
	case "hours", "hour":
		return Value(int64(diff.Hours()))
	case "days", "day":
		return Value(int64(diff.Hours() / 24))
	default:
		return Value(fmt.Errorf("date_diff: unsupported unit '%s'", unit))
	}
}

// dateBetweenFunc checks if a date is between two other dates
// Usage: date_between("2023-12-25", "2023-12-24", "2023-12-26")
func dateBetweenFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 3 {
		return Value(fmt.Errorf("date_between requires 3 arguments (date, startDate, endDate)"))
	}

	dateStr, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_between: first argument must be a date string"))
	}

	startDateStr, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("date_between: second argument must be a date string"))
	}

	endDateStr, ok := args[2].(string)
	if !ok {
		return Value(fmt.Errorf("date_between: third argument must be a date string"))
	}

	// Parse all dates
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}

	var date, startDate, endDate time.Time
	var err error

	for _, format := range formats {
		date, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_between: failed to parse date: %v", err))
	}

	for _, format := range formats {
		startDate, err = time.Parse(format, startDateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_between: failed to parse start date: %v", err))
	}

	for _, format := range formats {
		endDate, err = time.Parse(format, endDateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_between: failed to parse end date: %v", err))
	}

	// Check if date is between start and end (inclusive)
	isBetween := (date.Equal(startDate) || date.After(startDate)) &&
		(date.Equal(endDate) || date.Before(endDate))

	return Value(isBetween)
}

// dateCompareFunc compares two dates
// Usage: date_compare("2023-12-25", "2023-12-24") returns 1 (first is later)
// Returns: -1 (first < second), 0 (equal), 1 (first > second)
func dateCompareFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("date_compare requires 2 arguments (date1, date2)"))
	}

	date1Str, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("date_compare: first argument must be a date string"))
	}

	date2Str, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("date_compare: second argument must be a date string"))
	}

	// Parse both dates
	formats := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "2006-01-02 15:04:05"}

	var date1, date2 time.Time
	var err error

	for _, format := range formats {
		date1, err = time.Parse(format, date1Str)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_compare: failed to parse first date: %v", err))
	}

	for _, format := range formats {
		date2, err = time.Parse(format, date2Str)
		if err == nil {
			break
		}
	}
	if err != nil {
		return Value(fmt.Errorf("date_compare: failed to parse second date: %v", err))
	}

	if date1.Before(date2) {
		return Value(int64(-1))
	} else if date1.After(date2) {
		return Value(int64(1))
	} else {
		return Value(int64(0))
	}
}
