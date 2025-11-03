package interpreter

import (
	"testing"
	"time"
)

func TestNewTaggedValuePool(t *testing.T) {
	pool := NewTaggedValuePool()
	if pool == nil {
		t.Error("Expected pool to be created")
	}
}

func TestTaggedValuePool_GetPut(t *testing.T) {
	pool := NewTaggedValuePool()

	// Get a tagged value
	tv1 := pool.GetTaggedValue()
	if tv1 == nil {
		t.Error("Expected tagged value to be returned")
	}

	// Put it back
	pool.PutTaggedValue(tv1)

	// Get another one (should reuse the same instance)
	tv2 := pool.GetTaggedValue()
	if tv2 == nil {
		t.Error("Expected tagged value to be returned")
	}

	// Should be the same instance due to pooling
	if tv1 != tv2 {
		t.Error("Expected same instance to be reused")
	}
}

func TestCreateTaggedValue_Nil(t *testing.T) {
	tv := CreateTaggedValue(nil)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

	if tv.Type != TaggedNull {
		t.Errorf("Expected TaggedNull, got %v", tv.Type)
	}

	if tv.Data != 0 {
		t.Errorf("Expected data 0, got %d", tv.Data)
	}
}

func TestCreateTaggedValue_Bool(t *testing.T) {
	// Test true
	tv1 := CreateTaggedValue(true)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv1)

	if tv1.Type != TaggedBool {
		t.Errorf("Expected TaggedBool, got %v", tv1.Type)
	}

	if tv1.Data != 1 {
		t.Errorf("Expected data 1, got %d", tv1.Data)
	}

	// Test false
	tv2 := CreateTaggedValue(false)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv2)

	if tv2.Type != TaggedBool {
		t.Errorf("Expected TaggedBool, got %v", tv2.Type)
	}

	if tv2.Data != 0 {
		t.Errorf("Expected data 0, got %d", tv2.Data)
	}
}

func TestCreateTaggedValue_Int(t *testing.T) {
	value := 42
	tv := CreateTaggedValue(value)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

	if tv.Type != TaggedInt {
		t.Errorf("Expected TaggedInt, got %v", tv.Type)
	}

	if int(tv.Data) != value {
		t.Errorf("Expected data %d, got %d", value, tv.Data)
	}
}

func TestCreateTaggedValue_String(t *testing.T) {
	value := "hello world"
	tv := CreateTaggedValue(value)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

	if tv.Type != TaggedString {
		t.Errorf("Expected TaggedString, got %v", tv.Type)
	}

	if tv.Data == 0 {
		t.Error("Expected non-zero data pointer")
	}
}

func TestTaggedValue_ToValue(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected Value
	}{
		{"nil", nil, nil},
		{"bool true", true, true},
		{"bool false", false, false},
		{"int", 42, 42},
		{"string", "hello", "hello"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tv := CreateTaggedValue(test.input)
			defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

			result, err := tv.ToValue()
			if err != nil {
				t.Errorf("ToValue failed: %v", err)
				return
			}
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestTaggedValue_IsSmallValue(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"nil", nil, true},
		{"bool", true, true},
		{"int", 42, true},
		{"string", "hello", false},
		{"float", 3.14, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tv := CreateTaggedValue(test.input)
			defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

			result := tv.IsSmallValue()
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestNewCompactEnvironment(t *testing.T) {
	env := NewCompactEnvironment()
	if env == nil {
		t.Error("Expected environment to be created")
	}

	if len(env.vars) != 1 {
		t.Errorf("Expected 1 initial scope, got %d", len(env.vars))
	}
}

func TestCompactEnvironment_PushPopScope(t *testing.T) {
	env := NewCompactEnvironment()

	// Initial scope count
	initialCount := env.GetScopeCount()
	if initialCount != 1 {
		t.Errorf("Expected 1 initial scope, got %d", initialCount)
	}

	// Push a scope
	env.PushScope()
	if env.GetScopeCount() != 2 {
		t.Errorf("Expected 2 scopes after push, got %d", env.GetScopeCount())
	}

	// Pop a scope
	env.PopScope()
	if env.GetScopeCount() != 1 {
		t.Errorf("Expected 1 scope after pop, got %d", env.GetScopeCount())
	}

	// Try to pop the last scope (should not reduce below 1)
	env.PopScope()
	if env.GetScopeCount() != 1 {
		t.Errorf("Expected 1 scope after trying to pop last scope, got %d", env.GetScopeCount())
	}
}

func TestCompactEnvironment_AssignLookup(t *testing.T) {
	env := NewCompactEnvironment()

	// Assign a variable
	env.Assign("x", 42)

	// Look it up
	value, found := env.Lookup("x")
	if !found {
		t.Error("Expected variable to be found")
	}

	if value != 42 {
		t.Errorf("Expected value 42, got %v", value)
	}

	// Look up non-existent variable
	_, found = env.Lookup("y")
	if found {
		t.Error("Expected variable to not be found")
	}
}

func TestCompactEnvironment_ScopeIsolation(t *testing.T) {
	env := NewCompactEnvironment()

	// Assign in global scope
	env.Assign("global", "global_value")

	// Push new scope
	env.PushScope()

	// Assign in local scope
	env.Assign("local", "local_value")

	// Both should be accessible
	globalVal, found := env.Lookup("global")
	if !found || globalVal != "global_value" {
		t.Error("Expected global variable to be accessible")
	}

	localVal, found := env.Lookup("local")
	if !found || localVal != "local_value" {
		t.Error("Expected local variable to be accessible")
	}

	// Pop scope
	env.PopScope()

	// Global should still be accessible, local should not
	globalVal, found = env.Lookup("global")
	if !found || globalVal != "global_value" {
		t.Error("Expected global variable to still be accessible")
	}

	_, found = env.Lookup("local")
	if found {
		t.Error("Expected local variable to not be accessible after scope pop")
	}
}

func TestNewCacheFriendlyArray(t *testing.T) {
	arr := NewCacheFriendlyArray(10)
	if arr == nil {
		t.Error("Expected array to be created")
	}

	if arr.capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", arr.capacity)
	}

	if arr.size != 0 {
		t.Errorf("Expected size 0, got %d", arr.size)
	}
}

func TestCacheFriendlyArray_AppendGet(t *testing.T) {
	arr := NewCacheFriendlyArray(2)

	// Append values
	arr.Append(10)
	arr.Append(20)
	arr.Append(30) // This should trigger growth

	if arr.Size() != 3 {
		t.Errorf("Expected size 3, got %d", arr.Size())
	}

	// Check capacity growth
	if arr.capacity < 3 {
		t.Errorf("Expected capacity to grow, got %d", arr.capacity)
	}

	// Get values
	val1, ok := arr.Get(0)
	if !ok || val1 != 10 {
		t.Errorf("Expected value 10 at index 0, got %v", val1)
	}

	val2, ok := arr.Get(1)
	if !ok || val2 != 20 {
		t.Errorf("Expected value 20 at index 1, got %v", val2)
	}

	val3, ok := arr.Get(2)
	if !ok || val3 != 30 {
		t.Errorf("Expected value 30 at index 2, got %v", val3)
	}

	// Get out of bounds
	_, ok = arr.Get(3)
	if ok {
		t.Error("Expected out of bounds access to fail")
	}
}

func TestCacheFriendlyArray_Set(t *testing.T) {
	arr := NewCacheFriendlyArray(5)

	// Append some values
	arr.Append(10)
	arr.Append(20)

	// Set existing value
	ok := arr.Set(0, 100)
	if !ok {
		t.Error("Expected set to succeed")
	}

	val, _ := arr.Get(0)
	if val != 100 {
		t.Errorf("Expected value 100, got %v", val)
	}

	// Set out of bounds
	ok = arr.Set(5, 500)
	if ok {
		t.Error("Expected out of bounds set to fail")
	}
}

func TestCacheFriendlyArray_ToSlice(t *testing.T) {
	arr := NewCacheFriendlyArray(5)

	// Append values
	values := []Value{10, 20, 30}
	for _, v := range values {
		arr.Append(v)
	}

	// Convert to slice
	slice := arr.ToSlice()

	if len(slice) != len(values) {
		t.Errorf("Expected slice length %d, got %d", len(values), len(slice))
	}

	for i, expected := range values {
		if slice[i] != expected {
			t.Errorf("Expected value %v at index %d, got %v", expected, i, slice[i])
		}
	}
}

func TestNewCacheFriendlyMap(t *testing.T) {
	m := NewCacheFriendlyMap(10)
	if m == nil {
		t.Error("Expected map to be created")
	}

	if m.capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", m.capacity)
	}

	if m.size != 0 {
		t.Errorf("Expected size 0, got %d", m.size)
	}
}

func TestCacheFriendlyMap_SetGet(t *testing.T) {
	m := NewCacheFriendlyMap(2)

	// Set values
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3") // This should trigger growth

	if m.Size() != 3 {
		t.Errorf("Expected size 3, got %d", m.Size())
	}

	// Check capacity growth
	if m.capacity < 3 {
		t.Errorf("Expected capacity to grow, got %d", m.capacity)
	}

	// Get values
	val1, ok := m.Get("key1")
	if !ok || val1 != "value1" {
		t.Errorf("Expected value1, got %v", val1)
	}

	val2, ok := m.Get("key2")
	if !ok || val2 != "value2" {
		t.Errorf("Expected value2, got %v", val2)
	}

	val3, ok := m.Get("key3")
	if !ok || val3 != "value3" {
		t.Errorf("Expected value3, got %v", val3)
	}

	// Get non-existent key
	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("Expected non-existent key to not be found")
	}
}

func TestCacheFriendlyMap_Update(t *testing.T) {
	m := NewCacheFriendlyMap(5)

	// Set initial value
	m.Set("key1", "value1")

	// Update the same key
	m.Set("key1", "updated_value1")

	// Size should remain 1
	if m.Size() != 1 {
		t.Errorf("Expected size 1, got %d", m.Size())
	}

	// Value should be updated
	val, ok := m.Get("key1")
	if !ok || val != "updated_value1" {
		t.Errorf("Expected updated_value1, got %v", val)
	}
}

func TestCacheFriendlyMap_Delete(t *testing.T) {
	m := NewCacheFriendlyMap(5)

	// Set values
	m.Set("key1", "value1")
	m.Set("key2", "value2")

	// Delete existing key
	ok := m.Delete("key1")
	if !ok {
		t.Error("Expected delete to succeed")
	}

	if m.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", m.Size())
	}

	// Key should not be found
	_, found := m.Get("key1")
	if found {
		t.Error("Expected deleted key to not be found")
	}

	// Other key should still exist
	val, found := m.Get("key2")
	if !found || val != "value2" {
		t.Error("Expected other key to still exist")
	}

	// Delete non-existent key
	ok = m.Delete("nonexistent")
	if ok {
		t.Error("Expected delete of non-existent key to fail")
	}
}

func TestCacheFriendlyMap_ToMap(t *testing.T) {
	m := NewCacheFriendlyMap(5)

	// Set values
	expected := map[string]Value{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range expected {
		m.Set(k, v)
	}

	// Convert to regular map
	result := m.ToMap()

	if len(result) != len(expected) {
		t.Errorf("Expected map length %d, got %d", len(expected), len(result))
	}

	for k, expectedVal := range expected {
		if actualVal, exists := result[k]; !exists || actualVal != expectedVal {
			t.Errorf("Expected %s=%v, got %v", k, expectedVal, actualVal)
		}
	}
}

func TestSimpleHash(t *testing.T) {
	// Test that same strings produce same hash
	hash1 := simpleHash("hello")
	hash2 := simpleHash("hello")

	if hash1 != hash2 {
		t.Error("Expected same strings to produce same hash")
	}

	// Test that different strings produce different hashes (usually)
	hash3 := simpleHash("world")
	if hash1 == hash3 {
		t.Log("Warning: Different strings produced same hash (collision)")
	}
}

func TestNewMemoryLayoutOptimizer(t *testing.T) {
	optimizer := NewMemoryLayoutOptimizer()
	if optimizer == nil {
		t.Error("Expected optimizer to be created")
	}

	stats := optimizer.GetMemoryStats()
	if stats.TaggedValuesCreated != 0 {
		t.Error("Expected initial stats to be zero")
	}
}

func TestMemoryLayoutOptimizer_Stats(t *testing.T) {
	optimizer := NewMemoryLayoutOptimizer()

	// Increment counters
	optimizer.IncrementTaggedValuesCreated()
	optimizer.IncrementTaggedValuesReused()
	optimizer.IncrementCacheHits()
	optimizer.IncrementCacheMisses()

	stats := optimizer.GetMemoryStats()

	if stats.TaggedValuesCreated != 1 {
		t.Errorf("Expected 1 tagged value created, got %d", stats.TaggedValuesCreated)
	}

	if stats.TaggedValuesReused != 1 {
		t.Errorf("Expected 1 tagged value reused, got %d", stats.TaggedValuesReused)
	}

	if stats.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats.CacheHits)
	}

	if stats.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}

	// Reset stats
	optimizer.ResetStats()
	stats = optimizer.GetMemoryStats()

	if stats.TaggedValuesCreated != 0 {
		t.Error("Expected stats to be reset")
	}
}

// Benchmark tests
func BenchmarkCreateTaggedValue(b *testing.B) {
	values := []Value{nil, true, 42, 3.14, "hello world"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := values[i%len(values)]
		tv := CreateTaggedValue(value)
		GetGlobalTaggedValuePool().PutTaggedValue(tv)
	}
}

func BenchmarkTaggedValue_ToValue(b *testing.B) {
	tv := CreateTaggedValue(42)
	defer GetGlobalTaggedValuePool().PutTaggedValue(tv)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tv.ToValue()
	}
}

func BenchmarkCacheFriendlyArray_Append(b *testing.B) {
	arr := NewCacheFriendlyArray(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Append(i)
	}
}

func BenchmarkCacheFriendlyArray_Get(b *testing.B) {
	arr := NewCacheFriendlyArray(1000)
	for i := 0; i < 1000; i++ {
		arr.Append(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Get(int32(i % 1000))
	}
}

func BenchmarkCacheFriendlyMap_Set(b *testing.B) {
	m := NewCacheFriendlyMap(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set("key"+string(rune(i)), i)
	}
}

func BenchmarkCacheFriendlyMap_Get(b *testing.B) {
	m := NewCacheFriendlyMap(1000)
	for i := 0; i < 1000; i++ {
		m.Set("key"+string(rune(i)), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get("key" + string(rune(i%1000)))
	}
}

func BenchmarkCompactEnvironment_Assign(b *testing.B) {
	env := NewCompactEnvironment()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Assign("var"+string(rune(i)), i)
	}
}

func BenchmarkCompactEnvironment_Lookup(b *testing.B) {
	env := NewCompactEnvironment()
	for i := 0; i < 1000; i++ {
		env.Assign("var"+string(rune(i)), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Lookup("var" + string(rune(i%1000)))
	}
}

// Test concurrent access
func TestTaggedValuePool_ConcurrentAccess(t *testing.T) {
	pool := NewTaggedValuePool()

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				tv := pool.GetTaggedValue()
				pool.PutTaggedValue(tv)
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
}

func TestMemoryLayoutOptimizer_ConcurrentAccess(t *testing.T) {
	optimizer := NewMemoryLayoutOptimizer()

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				optimizer.IncrementTaggedValuesCreated()
				optimizer.IncrementCacheHits()
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

	// Check final stats
	stats := optimizer.GetMemoryStats()
	expectedCreated := int64(1000)
	expectedHits := int64(1000)

	if stats.TaggedValuesCreated != expectedCreated {
		t.Errorf("Expected %d tagged values created, got %d", expectedCreated, stats.TaggedValuesCreated)
	}

	if stats.CacheHits != expectedHits {
		t.Errorf("Expected %d cache hits, got %d", expectedHits, stats.CacheHits)
	}
}
