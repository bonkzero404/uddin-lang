package interpreter

import (
	"testing"
	"time"
)

// Mock interpreter for testing
type mockInterpreter struct{}

// Mock builtin function for testing
func mockLenFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic("len() requires 1 argument")
	}
	switch v := args[0].(type) {
	case string:
		return len(v)
	case *[]Value:
		return len(*v)
	default:
		return 0
	}
}

func mockPrintFunc(interp *interpreter, pos Position, args []Value) Value {
	return nil
}

func TestBuiltinFunctionDispatcher_RegisterBuiltinFunction(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	
	// Test registering a function
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	// Check if function is registered
	if len(dispatcher.functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(dispatcher.functions))
	}
	
	if dispatcher.functions[0].Name != "len" {
		t.Errorf("Expected function name 'len', got '%s'", dispatcher.functions[0].Name)
	}
	
	if dispatcher.functions[0].ArgCount != 1 {
		t.Errorf("Expected arg count 1, got %d", dispatcher.functions[0].ArgCount)
	}
	
	if !dispatcher.functions[0].FastPath {
		t.Error("Expected fast path to be true")
	}
}

func TestBuiltinFunctionDispatcher_DispatchBuiltinFunction(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	// Test successful dispatch
	args := []Value{"hello"}
	result, exists := dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args)
	
	if !exists {
		t.Error("Expected function to exist")
	}
	
	if result != 5 {
		t.Errorf("Expected result 5, got %v", result)
	}
	
	// Test non-existent function
	_, exists = dispatcher.DispatchBuiltinFunction("nonexistent", nil, Position{}, args)
	if exists {
		t.Error("Expected function to not exist")
	}
}

func TestBuiltinFunctionDispatcher_ArgumentValidation(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	// Test with correct number of arguments
	args := []Value{"hello"}
	_, exists := dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args)
	if !exists {
		t.Error("Expected function to exist")
	}
	
	// Test with incorrect number of arguments
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for incorrect argument count")
		}
	}()
	
	wrongArgs := []Value{"hello", "world"}
	dispatcher.DispatchBuiltinFunction("len", nil, Position{}, wrongArgs)
}

func TestBuiltinFunctionDispatcher_CallCount(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	// Initial call count should be 0
	if count := dispatcher.GetCallCount("len"); count != 0 {
		t.Errorf("Expected initial call count 0, got %d", count)
	}
	
	// Call function multiple times
	args := []Value{"hello"}
	for i := 0; i < 5; i++ {
		dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args)
	}
	
	// Check call count
	if count := dispatcher.GetCallCount("len"); count != 5 {
		t.Errorf("Expected call count 5, got %d", count)
	}
}

func TestBuiltinFunctionDispatcher_MostCalledFunctions(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	dispatcher.RegisterBuiltinFunction("print", mockPrintFunc, -1, false)
	
	// Call functions different number of times
	args1 := []Value{"hello"}
	args2 := []Value{"world"}
	
	// Call len 3 times
	for i := 0; i < 3; i++ {
		dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args1)
	}
	
	// Call print 5 times
	for i := 0; i < 5; i++ {
		dispatcher.DispatchBuiltinFunction("print", nil, Position{}, args2)
	}
	
	// Get most called functions
	mostCalled := dispatcher.GetMostCalledFunctions(2)
	
	if len(mostCalled) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(mostCalled))
	}
	
	if mostCalled[0] != "print" {
		t.Errorf("Expected most called function to be 'print', got '%s'", mostCalled[0])
	}
	
	if mostCalled[1] != "len" {
		t.Errorf("Expected second most called function to be 'len', got '%s'", mostCalled[1])
	}
}

func TestArgumentValidationCache_ValidateArguments(t *testing.T) {
	cache := NewArgumentValidationCache(10)
	
	// Test validation (always returns true in mock implementation)
	result := cache.ValidateArguments("len", []string{"string"})
	if !result {
		t.Error("Expected validation to pass")
	}
	
	// Test caching - second call should use cache
	result2 := cache.ValidateArguments("len", []string{"string"})
	if !result2 {
		t.Error("Expected cached validation to pass")
	}
}

func TestArgumentValidationCache_CacheEviction(t *testing.T) {
	cache := NewArgumentValidationCache(2)
	
	// Fill cache beyond capacity
	cache.ValidateArguments("func1", []string{"string"})
	cache.ValidateArguments("func2", []string{"int"})
	cache.ValidateArguments("func3", []string{"float"})
	
	// Cache should have been partially cleared
	if len(cache.cache) > 2 {
		t.Errorf("Expected cache size <= 2, got %d", len(cache.cache))
	}
}

func TestSpecializedBuiltinFunctions_FastLen(t *testing.T) {
	sbf := &SpecializedBuiltinFunctions{}
	
	// Test string length
	length, ok := sbf.FastLen("hello")
	if !ok || length != 5 {
		t.Errorf("Expected length 5 for string, got %d, ok=%v", length, ok)
	}
	
	// Test array length
	arr := &[]Value{1, 2, 3}
	length, ok = sbf.FastLen(arr)
	if !ok || length != 3 {
		t.Errorf("Expected length 3 for array, got %d, ok=%v", length, ok)
	}
	
	// Test map length
	m := map[string]Value{"a": 1, "b": 2}
	length, ok = sbf.FastLen(m)
	if !ok || length != 2 {
		t.Errorf("Expected length 2 for map, got %d, ok=%v", length, ok)
	}
	
	// Test unsupported type
	_, ok = sbf.FastLen(123)
	if ok {
		t.Error("Expected FastLen to fail for unsupported type")
	}
}

func TestSpecializedBuiltinFunctions_FastTypeof(t *testing.T) {
	sbf := &SpecializedBuiltinFunctions{}
	
	tests := []struct {
		value    Value
		expected string
	}{
		{nil, "null"},
		{true, "bool"},
		{42, "int"},
		{3.14, "float"},
		{"hello", "string"},
		{&[]Value{}, "array"},
		{map[string]Value{}, "object"},
	}
	
	for _, test := range tests {
		result := sbf.FastTypeof(test.value)
		if result != test.expected {
			t.Errorf("Expected type '%s' for value %v, got '%s'", test.expected, test.value, result)
		}
	}
}

func TestSpecializedBuiltinFunctions_FastStr(t *testing.T) {
	sbf := &SpecializedBuiltinFunctions{}
	
	tests := []struct {
		value    Value
		expected string
	}{
		{"hello", "hello"},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{nil, "null"},
	}
	
	for _, test := range tests {
		result := sbf.FastStr(test.value)
		if result != test.expected {
			t.Errorf("Expected string '%s' for value %v, got '%s'", test.expected, test.value, result)
		}
	}
}

// Benchmark tests
func BenchmarkBuiltinFunctionDispatcher_Dispatch(b *testing.B) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	args := []Value{"hello world"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args)
	}
}

func BenchmarkSpecializedBuiltinFunctions_FastLen(b *testing.B) {
	sbf := &SpecializedBuiltinFunctions{}
	value := "hello world this is a test string"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sbf.FastLen(value)
	}
}

func BenchmarkSpecializedBuiltinFunctions_FastTypeof(b *testing.B) {
	sbf := &SpecializedBuiltinFunctions{}
	value := "hello world"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sbf.FastTypeof(value)
	}
}

func BenchmarkArgumentValidationCache(b *testing.B) {
	cache := NewArgumentValidationCache(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ValidateArguments("len", []string{"string"})
	}
}

// Test concurrent access
func TestBuiltinFunctionDispatcher_ConcurrentAccess(t *testing.T) {
	dispatcher := NewBuiltinFunctionDispatcher()
	dispatcher.RegisterBuiltinFunction("len", mockLenFunc, 1, true)
	
	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			args := []Value{"hello"}
			for j := 0; j < 100; j++ {
				dispatcher.DispatchBuiltinFunction("len", nil, Position{}, args)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
	
	// Check final call count
	expectedCount := int64(1000)
	actualCount := dispatcher.GetCallCount("len")
	if actualCount != expectedCount {
		t.Errorf("Expected call count %d, got %d", expectedCount, actualCount)
	}
}

func TestArgumentValidationCache_ConcurrentAccess(t *testing.T) {
	cache := NewArgumentValidationCache(100)
	
	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				cache.ValidateArguments("func", []string{"string"})
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}