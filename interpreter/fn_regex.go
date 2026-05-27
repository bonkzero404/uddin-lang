package interpreter

import (
	"fmt"

	"github.com/coregx/coregex"
)

// isregexFunc implements the is_regex_match() built-in function
// Checks if a string matches a regular expression pattern
// Parameters:
//   - pattern: Regular expression pattern
//   - str: String to check against the pattern
//
// Returns true if the string matches the pattern, false otherwise
// Example: is_regex_match("^-?\\d+\\.\\d+$", "3.14") -> true
func isregexFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "is_regex_match", args, 2); err != Value(nil) {
		return err
	}
	str, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "is_regex_match() requires second argument to be a string"))
	}
	switch p := args[0].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return false
		}
		return re.MatchString(str)
	case *coregex.Regexp:
		return p.MatchString(str)
	default:
		panic(typeError(pos, "is_regex_match() requires first argument to be a string or pre-compiled pattern"))
	}
}

// regexMatchFunc tests if a string matches a regular expression pattern
// Parameters:
//   - text: The string to test (first argument)
//   - pattern: The regular expression pattern (second argument)
//
// Returns true if the text matches the pattern, false otherwise
// Example: regex_match("hello@example.com", "^[\\w._%+-]+@[\\w.-]+\\.[a-zA-Z]{2,}$") -> true
func regexMatchFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		return Value(fmt.Errorf("regex_match() requires exactly 2 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_match() requires first argument to be a string, not %s", typeName(args[0])))
	}
	switch p := args[1].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
		return Value(re.MatchString(text))
	case *coregex.Regexp:
		return Value(p.MatchString(text))
	default:
		return Value(fmt.Errorf("regex_match() requires second argument to be a string pattern, not %s", typeName(args[1])))
	}
}

// regexFindFunc finds the first match of a regular expression in a string
// Parameters:
//   - text: The string to search in (first argument)
//   - pattern: The regular expression pattern (second argument)
//
// Returns the first match as a string, or null if no match found
// Example: regex_find("hello world 123", "\\d+") -> "123"
func regexFindFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		return Value(fmt.Errorf("regex_find() requires exactly 2 arguments, got %d", len(args)))
	}

	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_find() requires first argument to be a string, not %s", typeName(args[0])))
	}

	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return Value(fmt.Errorf("regex_find() requires second argument to be a string pattern, not %s", typeName(args[1])))
	}

	match := re.FindString(text)
	if match == "" {
		return Value(nil)
	}

	return Value(match)
}

// regexFindAllFunc finds all matches of a regular expression in a string
// Parameters:
//   - text: The string to search in (first argument)
//   - pattern: The regular expression pattern (second argument)
//   - limit: Optional maximum number of matches (third argument, -1 for all)
//
// Returns an array of all matches
// Example: regex_find_all("hello 123 world 456", "\\d+") -> ["123", "456"]
func regexFindAllFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return Value(fmt.Errorf("regex_find_all() requires 2 or 3 arguments, got %d", len(args)))
	}

	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_find_all() requires first argument to be a string, not %s", typeName(args[0])))
	}

	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return Value(fmt.Errorf("regex_find_all() requires second argument to be a string pattern, not %s", typeName(args[1])))
	}

	limit := -1 // Default: find all matches
	if len(args) == 3 {
		if l, ok := args[2].(int); ok {
			limit = l
		} else {
			return Value(fmt.Errorf("regex_find_all() requires third argument to be an integer (limit), not %s", typeName(args[2])))
		}
	}

	matches := re.FindAllString(text, limit)
	result := make([]Value, len(matches))
	for i, match := range matches {
		result[i] = Value(match)
	}

	return Value(&result)
}

// regexReplaceFunc replaces matches of a regular expression with a replacement string
// Parameters:
//   - text: The string to search and replace in (first argument)
//   - pattern: The regular expression pattern (second argument)
//   - replacement: The replacement string (third argument)
//
// Returns the modified string with replacements
// Example: regex_replace("hello 123 world", "\\d+", "XXX") -> "hello XXX world"
func regexReplaceFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 3 {
		return Value(fmt.Errorf("regex_replace() requires exactly 3 arguments, got %d", len(args)))
	}

	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_replace() requires first argument to be a string, not %s", typeName(args[0])))
	}

	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return Value(fmt.Errorf("regex_replace() requires second argument to be a string pattern, not %s", typeName(args[1])))
	}

	replacement, ok := args[2].(string)
	if !ok {
		return Value(fmt.Errorf("regex_replace() requires third argument to be a string (replacement), not %s", typeName(args[2])))
	}

	result := re.ReplaceAllString(text, replacement)
	return Value(result)
}

// regexSplitFunc splits a string using a regular expression pattern as delimiter
// Parameters:
//   - text: The string to split (first argument)
//   - pattern: The regular expression pattern delimiter (second argument)
//   - limit: Optional maximum number of splits (third argument, -1 for all)
//
// Returns an array of string parts
// Example: regex_split("hello,world;test", "[,;]") -> ["hello", "world", "test"]
func regexSplitFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		return Value(fmt.Errorf("regex_split() requires 2 or 3 arguments, got %d", len(args)))
	}

	text, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("regex_split() requires first argument to be a string, not %s", typeName(args[0])))
	}

	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return Value(fmt.Errorf("regex_split() requires second argument to be a string pattern, not %s", typeName(args[1])))
	}

	limit := -1 // Default: split all
	if len(args) == 3 {
		if l, ok := args[2].(int); ok {
			limit = l
		} else {
			return Value(fmt.Errorf("regex_split() requires third argument to be an integer (limit), not %s", typeName(args[2])))
		}
	}

	parts := re.Split(text, limit)
	result := make([]Value, len(parts))
	for i, part := range parts {
		result[i] = Value(part)
	}

	return Value(&result)
}
