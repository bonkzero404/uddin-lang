package interpreter

import (
	"fmt"
	"reflect"
)

// ValueType represents the type of a runtime value
type ValueType string

const (
	TypeNull     ValueType = "null"
	TypeBool     ValueType = "bool"
	TypeInt      ValueType = "int"
	TypeFloat    ValueType = "float"
	TypeString   ValueType = "string"
	TypeArray    ValueType = "array"
	TypeObject   ValueType = "object"
	TypeFunction ValueType = "function"
)

// GetValueType returns the type of a runtime value
func GetValueType(value Value) ValueType {
	if value == nil {
		return TypeNull
	}

	switch value.(type) {
	case bool:
		return TypeBool
	case int:
		return TypeInt
	case float64:
		return TypeFloat
	case string:
		return TypeString
	case []Value:
		return TypeArray
	case map[string]Value:
		return TypeObject
	case *userFunction, builtinFunction:
		return TypeFunction
	default:
		return ValueType(fmt.Sprintf("unknown(%s)", reflect.TypeOf(value).String()))
	}
}

// IsNumeric checks if a value is numeric (int or float)
func IsNumeric(value Value) bool {
	switch value.(type) {
	case int, float64:
		return true
	default:
		return false
	}
}

// ToFloat converts a numeric value to float64
func ToFloat(value Value) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// ToInt converts a numeric value to int
func ToInt(value Value) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// ToString converts a value to its string representation
func ToString(value Value) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case string:
		return v
	case []Value:
		return fmt.Sprintf("%v", v)
	case map[string]Value:
		return fmt.Sprintf("%v", v)
	case *userFunction:
		return fmt.Sprintf("<function %s>", v.Name)
	case builtinFunction:
		return "<builtin function>"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// IsTruthy determines if a value is truthy in Uddin-Lang
func IsTruthy(value Value) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return len(v) > 0
	case []Value:
		return len(v) > 0
	case map[string]Value:
		return len(v) > 0
	default:
		return true
	}
}

// DeepCopy creates a deep copy of a value
func DeepCopy(value Value) Value {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool, int, float64, string:
		return v // These are immutable
	case []Value:
		copy := make([]Value, len(v))
		for i, item := range v {
			copy[i] = DeepCopy(item)
		}
		return copy
	case map[string]Value:
		copy := make(map[string]Value)
		for key, val := range v {
			copy[key] = DeepCopy(val)
		}
		return copy
	default:
		return v // Functions and other types are treated as references
	}
}
