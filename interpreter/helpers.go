package interpreter

import (
	"fmt"
	"sort"
)

// toString converts any interpreter Value to its string representation
// This is used by the str() built-in function and for string concatenation
// Parameters:
//   - value: The Value to convert to a string
//   - quoteStr: Whether to quote string values (for display in arrays/objects)
//
// Returns a string representation of the value
func toString(value Value, quoteStr bool) string {
	var s string
	switch v := value.(type) {
	case nil:
		s = "null" // Null value
	case bool:
		if v {
			s = "true"
		} else {
			s = "false"
		}
	case int:
		s = fmt.Sprintf("%d", v) // Integer
	case int64:
		s = fmt.Sprintf("%d", v) // int64 support
	case float64:
		s = fmt.Sprintf("%g", v) // Float
	case string:
		if quoteStr {
			s = fmt.Sprintf("%q", v) // Quoted string for arrays/objects
		} else {
			s = v // Raw string for display
		}
	case []Value:
		// Convert array elements recursively (slice variant) - optimized
		concat := GetStringConcatenator()
		defer PutStringConcatenator(concat)
		concat.WriteString("[")
		for i, val := range v {
			if i > 0 {
				concat.WriteString(", ")
			}
			concat.WriteString(toString(val, true))
		}
		concat.WriteString("]")
		s = concat.String()
	case *[]Value:
		// Convert array elements recursively - optimized
		concat := GetStringConcatenator()
		defer PutStringConcatenator(concat)
		concat.WriteString("[")
		for i, val := range *v {
			if i > 0 {
				concat.WriteString(", ")
			}
			concat.WriteString(toString(val, true))
		}
		concat.WriteString("]")
		s = concat.String()
	case map[string]Value:
		// Convert object key-value pairs recursively - optimized
		strs := make([]string, 0, len(v))
		for k, val := range v {
			item := fmt.Sprintf("%q: %s", k, toString(val, true))
			strs = SmartAppendString(strs, item)
		}
		sort.Strings(strs) // Ensure str(output) is consistent
		concat := GetStringConcatenator()
		defer PutStringConcatenator(concat)
		concat.WriteString("{")
		for i, str := range strs {
			if i > 0 {
				concat.WriteString(", ")
			}
			concat.WriteString(str)
		}
		concat.WriteString("}")
		s = concat.String()
	case *map[string]Value:
		// Convert pointer to object key-value pairs recursively
		if v == nil {
			s = "null"
		} else {
			strs := make([]string, 0, len(*v))
			for k, val := range *v {
				item := fmt.Sprintf("%q: %s", k, toString(val, true))
				strs = SmartAppendString(strs, item)
			}
			sort.Strings(strs) // Ensure str(output) is consistent
			concat := GetStringConcatenator()
			defer PutStringConcatenator(concat)
			concat.WriteString("{")
			for i, str := range strs {
				if i > 0 {
					concat.WriteString(", ")
				}
				concat.WriteString(str)
			}
			concat.WriteString("}")
			s = concat.String()
		}
	case functionType:
		s = v.name() // Function representation
	case error:
		s = v.Error() // Error representation
	default:
		// Interpreter should never give us this
		return fmt.Sprintf("str() got unexpected type %T", v)
	}
	return s
}

// typeName returns the type name of a Value as a string
// This is used by the typeof() built-in function
// Parameters:
//   - v: The Value to get the type name of
//
// Returns a string representing the type name
func typeName(v Value) string {
	var t string
	switch v.(type) {
	case nil:
		t = "nullable" // Null value
	case bool:
		t = "boolean" // Boolean value
	case int:
		t = "integer" // Integer value
	case float64:
		t = "float" // Float value
	case string:
		t = "string" // String value
	case *[]Value:
		t = "array" // Array value
	case map[string]Value:
		t = "object" // Map/Object value
	case functionType:
		t = "function" // Function value
	default:
		// Interpreter should never give us this
		return "unknown"
	}
	return t
}
