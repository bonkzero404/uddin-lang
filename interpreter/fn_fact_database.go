package interpreter

import (
	"fmt"
	"strings"
)

// factAssertFunc adds a fact to the knowledge base
// Usage: fact_assert("customer", "john", {"age": 30, "status": "premium"})
func factAssertFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		panic(typeError(pos, "fact_assert() requires at least 2 arguments (category, key, [value])"))
	}

	category, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "fact_assert() requires first argument to be a category string"))
	}

	key, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "fact_assert() requires second argument to be a key string"))
	}

	interp.factMutex.Lock()
	defer interp.factMutex.Unlock()

	// Create category if it doesn't exist
	fullKey := category + ":" + key

	if len(args) == 2 {
		// Simple boolean fact
		interp.factDatabase[fullKey] = true
	} else {
		// Fact with value
		interp.factDatabase[fullKey] = args[2]
	}

	return Value(true)
}

// factRetractFunc removes a fact from the knowledge base
// Usage: fact_retract("customer", "john") or fact_retract("customer")
func factRetractFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		return Value(fmt.Errorf("fact_retract requires at least 1 argument (category, [key])"))
	}

	category, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("fact_retract: first argument must be a category string"))
	}

	interp.factMutex.Lock()
	defer interp.factMutex.Unlock()

	if len(args) == 1 {
		// Remove all facts in category
		prefix := category + ":"
		count := 0
		for key := range interp.factDatabase {
			if strings.HasPrefix(key, prefix) {
				delete(interp.factDatabase, key)
				count++
			}
		}
		return Value(int64(count))
	} else {
		// Remove specific fact
		key, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("fact_retract: second argument must be a key string"))
		}

		fullKey := category + ":" + key
		if _, exists := interp.factDatabase[fullKey]; exists {
			delete(interp.factDatabase, fullKey)
			return Value(true)
		}
		return Value(false)
	}
}

// factQueryFunc queries facts from the knowledge base
// Usage: fact_query("customer", "john") or fact_query("customer")
func factQueryFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		return Value(fmt.Errorf("fact_query requires at least 1 argument (category, [key])"))
	}

	category, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("fact_query: first argument must be a category string"))
	}

	interp.factMutex.RLock()
	defer interp.factMutex.RUnlock()

	if len(args) == 1 {
		// Return all facts in category
		prefix := category + ":"
		results := make(map[string]any)
		for key, value := range interp.factDatabase {
			if strings.HasPrefix(key, prefix) {
				actualKey := key[len(prefix):]
				results[actualKey] = value
			}
		}

		// Convert to Value
		resultMap := make(map[string]Value)
		for k, v := range results {
			resultMap[k] = Value(v)
		}
		return Value(&resultMap)
	} else {
		// Return specific fact
		key, ok := args[1].(string)
		if !ok {
			return Value(fmt.Errorf("fact_query: second argument must be a key string"))
		}

		fullKey := category + ":" + key
		if value, exists := interp.factDatabase[fullKey]; exists {
			return Value(value)
		}
		return Value(nil)
	}
}

// factExistsFunc checks if a fact exists in the knowledge base
// Usage: fact_exists("customer", "john")
func factExistsFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(fmt.Errorf("fact_exists requires 2 arguments (category, key)"))
	}

	category, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("fact_exists: first argument must be a category string"))
	}

	key, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("fact_exists: second argument must be a key string"))
	}

	interp.factMutex.RLock()
	defer interp.factMutex.RUnlock()

	fullKey := category + ":" + key
	_, exists := interp.factDatabase[fullKey]
	return Value(exists)
}

// factCountFunc returns the number of facts in a category or total
// Usage: fact_count("customer") or fact_count()
func factCountFunc(interp *interpreter, pos Position, args []Value) Value {
	interp.factMutex.RLock()
	defer interp.factMutex.RUnlock()

	if len(args) == 0 {
		// Return total count
		return Value(len(interp.factDatabase)) // int, not int64
	}

	category, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "fact_count() requires first argument to be a category string"))
	}

	// Count facts in category
	prefix := category + ":"
	count := 0
	for key := range interp.factDatabase {
		if strings.HasPrefix(key, prefix) {
			count++
		}
	}
	return Value(count)
}

// factClearFunc clears all facts or facts in a specific category
// Usage: fact_clear() or fact_clear("customer")
func factClearFunc(interp *interpreter, pos Position, args []Value) Value {
	interp.factMutex.Lock()
	defer interp.factMutex.Unlock()

	if len(args) == 0 {
		// Clear all facts
		count := len(interp.factDatabase)
		interp.factDatabase = make(map[string]any)
		return Value(int64(count))
	}

	category, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("fact_clear: first argument must be a category string"))
	}

	// Clear facts in category
	prefix := category + ":"
	count := 0
	for key := range interp.factDatabase {
		if strings.HasPrefix(key, prefix) {
			delete(interp.factDatabase, key)
			count++
		}
	}
	return Value(int64(count))
}

// factGetAllFunc returns all facts in the knowledge base
// Usage: fact_get_all()
func factGetAllFunc(interp *interpreter, pos Position, args []Value) Value {
	interp.factMutex.RLock()
	defer interp.factMutex.RUnlock()

	results := make(map[string]Value)
	for key, value := range interp.factDatabase {
		results[key] = Value(value)
	}

	return Value(&results)
}
