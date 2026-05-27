package stdlib_json

import (
	"fmt"

	"github.com/bonkzero404/uddin-lang/interpreter"
	gojson "github.com/goccy/go-json"
)

type jsonModule struct{}

func (m *jsonModule) Name() string { return "json" }

func (m *jsonModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"parse":     parseFunc,
		"stringify": stringifyFunc,
	}
}

func init() {
	interpreter.RegisterModule(&jsonModule{})
}

// parseFunc parses a JSON string and returns the corresponding interpreter value.
func parseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "json.parse() requires exactly 1 argument, got %d", len(args)))
	}
	jsonStr, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "json.parse() requires a string argument, not %T", args[0]))
	}
	var jsonData any
	if err := gojson.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		panic(interpreter.RuntimeErrorf(pos, "json.parse(): invalid JSON: %v", err))
	}
	return jsonToValue(jsonData)
}

// stringifyFunc converts an interpreter value to a JSON string.
func stringifyFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "json.stringify() requires exactly 1 argument, got %d", len(args)))
	}
	jsonBytes, err := valueToJSON(args[0])
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "json.stringify(): failed to convert to JSON: %v", err))
	}
	return interpreter.Value(string(jsonBytes))
}

// valueToJSON converts an interpreter.Value to JSON bytes.
func valueToJSON(v interpreter.Value) ([]byte, error) {
	switch val := v.(type) {
	case map[string]interpreter.Value:
		goMap := make(map[string]any)
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return gojson.Marshal(goMap)
	case []interpreter.Value:
		goSlice := make([]any, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return gojson.Marshal(goSlice)
	case *[]interpreter.Value:
		goSlice := make([]any, len(*val))
		for i, v := range *val {
			goSlice[i] = valueToInterface(v)
		}
		return gojson.Marshal(goSlice)
	default:
		return gojson.Marshal(valueToInterface(v))
	}
}

// valueToInterface converts an interpreter.Value to any for JSON marshaling.
func valueToInterface(v interpreter.Value) any {
	switch val := v.(type) {
	case map[string]interpreter.Value:
		goMap := make(map[string]any)
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return goMap
	case []interpreter.Value:
		goSlice := make([]any, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return goSlice
	case *[]interpreter.Value:
		goSlice := make([]any, len(*val))
		for i, v := range *val {
			goSlice[i] = valueToInterface(v)
		}
		return goSlice
	case nil:
		return nil
	default:
		return val
	}
}

// jsonToValue converts a JSON any to an interpreter.Value.
func jsonToValue(data any) interpreter.Value {
	switch val := data.(type) {
	case map[string]any:
		result := make(map[string]interpreter.Value)
		for k, v := range val {
			result[k] = jsonToValue(v)
		}
		return interpreter.Value(result)
	case []any:
		result := make([]interpreter.Value, len(val))
		for i, v := range val {
			result[i] = jsonToValue(v)
		}
		return interpreter.Value(&result)
	case float64:
		if val == float64(int(val)) {
			return interpreter.Value(int(val))
		}
		return interpreter.Value(val)
	case bool:
		return interpreter.Value(val)
	case string:
		return interpreter.Value(val)
	case nil:
		return interpreter.Value(nil)
	default:
		return interpreter.Value(fmt.Sprintf("%v", val))
	}
}
