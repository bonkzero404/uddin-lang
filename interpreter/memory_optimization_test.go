package interpreter

import (
	"strings"
	"testing"
)

// BenchmarkMemoryPoolUsage tests the effectiveness of memory pools
func BenchmarkMemoryPoolUsage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test small array pool
		arr := GetPooledArray(10)
		for j := 0; j < 10; j++ {
			arr = append(arr, Value(float64(j)))
		}
		PutPooledArray(arr)
		
		// Test map pool
		m := GetPooledMap()
		for j := 0; j < 5; j++ {
			m["key"+string(rune(j))] = Value(float64(j))
		}
		PutPooledMap(m)
		
		// Test string builder pool
		sb := GetPooledStringBuilder()
		for j := 0; j < 10; j++ {
			sb.WriteString("test")
		}
		PutPooledStringBuilder(sb)
	}
}

// BenchmarkDirectAllocation tests direct allocation without pools
func BenchmarkDirectAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Direct array allocation
		arr := make([]Value, 0, 16)
		for j := 0; j < 10; j++ {
			arr = append(arr, Value(float64(j)))
		}
		
		// Direct map allocation
		m := make(map[string]Value)
		for j := 0; j < 5; j++ {
			m["key"+string(rune(j))] = Value(float64(j))
		}
		
		// Direct string builder allocation
		var sb strings.Builder
		for j := 0; j < 10; j++ {
			sb.WriteString("test")
		}
	}
}

// BenchmarkOptimizedMakeSlice tests the optimized slice creation
func BenchmarkOptimizedMakeSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := OptimizedMakeSlice(5, 10)
		for j := 0; j < 15; j++ {
			arr = append(arr, Value(float64(j)))
		}
	}
}

// BenchmarkRegularMakeSlice tests regular slice creation
func BenchmarkRegularMakeSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := make([]Value, 5)
		for j := 0; j < 15; j++ {
			arr = append(arr, Value(float64(j)))
		}
	}
}

// BenchmarkOptimizedMakeMap tests the optimized map creation
func BenchmarkOptimizedMakeMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := OptimizedMakeMap(20)
		for j := 0; j < 20; j++ {
			m["key"+string(rune(j))] = Value(float64(j))
		}
	}
}

// BenchmarkRegularMakeMap tests regular map creation
func BenchmarkRegularMakeMap(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[string]Value)
		for j := 0; j < 20; j++ {
			m["key"+string(rune(j))] = Value(float64(j))
		}
	}
}