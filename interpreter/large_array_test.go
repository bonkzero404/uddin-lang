package interpreter

import (
	"testing"
)

// BenchmarkLargeArrayPoolUsage tests memory pool with large arrays
func BenchmarkLargeArrayPoolUsage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test large array pool
		arr := GetPooledArray(5000) // Large array
		for j := 0; j < 5000; j++ {
			arr = append(arr, Value(float64(j)))
		}
		PutPooledArray(arr)
	}
}

// BenchmarkLargeArrayDirectAllocation tests direct allocation for large arrays
func BenchmarkLargeArrayDirectAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Direct large array allocation
		arr := make([]Value, 0, 5000)
		for j := 0; j < 5000; j++ {
			_ = append(arr, Value(float64(j)))
		}
	}
}

// BenchmarkLargeArrayFilter tests filtering with large arrays
func BenchmarkLargeArrayFilter(b *testing.B) {
	// Create a large test array
	testArray := make([]Value, 10000)
	for i := 0; i < 10000; i++ {
		testArray[i] = Value(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter even numbers
		result := OptimizedArrayFilter(testArray, func(v Value) bool {
			if num, ok := v.(float64); ok {
				return int(num)%2 == 0
			}
			return false
		})
		_ = result // Use result to prevent optimization
	}
}

// BenchmarkLargeArrayFilterDirect tests filtering without pool optimization
func BenchmarkLargeArrayFilterDirect(b *testing.B) {
	// Create a large test array
	testArray := make([]Value, 10000)
	for i := 0; i < 10000; i++ {
		testArray[i] = Value(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter even numbers without pool
		result := make([]Value, 0, len(testArray)/2)
		for _, v := range testArray {
			if num, ok := v.(float64); ok {
				if int(num)%2 == 0 {
					result = append(result, v)
				}
			}
		}
		_ = result // Use result to prevent optimization
	}
}

// BenchmarkVeryLargeArrayOperations tests with extremely large arrays
func BenchmarkVeryLargeArrayOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test with very large array (50k elements)
		arr := GetPooledArray(50000)
		for j := 0; j < 50000; j++ {
			arr = append(arr, Value(float64(j)))
		}

		// Simulate some processing
		count := 0
		for _, v := range arr {
			if num, ok := v.(float64); ok && int(num)%100 == 0 {
				count++
			}
		}

		PutPooledArray(arr)
		_ = count // Use count to prevent optimization
	}
}

// BenchmarkMemoryGrowthPattern tests memory growth patterns
func BenchmarkMemoryGrowthPattern(b *testing.B) {
	sizes := []int{100, 1000, 5000, 10000, 50000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			arr := GetPooledArray(size)
			for j := 0; j < size; j++ {
				arr = append(arr, Value(float64(j)))
			}
			PutPooledArray(arr)
		}
	}
}

// BenchmarkChunkedArrayProcessing tests chunked processing for very large arrays
func BenchmarkChunkedArrayProcessing(b *testing.B) {
	// Create a very large test array (100k elements)
	testArray := make([]Value, 100000)
	for i := 0; i < 100000; i++ {
		testArray[i] = Value(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter using chunked processing
		result := OptimizedLargeArrayFilter(testArray, func(v Value) bool {
			if num, ok := v.(float64); ok {
				return int(num)%10 == 0
			}
			return false
		})
		_ = result // Use result to prevent optimization
	}
}

// BenchmarkRegularArrayProcessing tests regular processing for comparison
func BenchmarkRegularArrayProcessing(b *testing.B) {
	// Create a very large test array (100k elements)
	testArray := make([]Value, 100000)
	for i := 0; i < 100000; i++ {
		testArray[i] = Value(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter using regular processing
		result := OptimizedArrayFilter(testArray, func(v Value) bool {
			if num, ok := v.(float64); ok {
				return int(num)%10 == 0
			}
			return false
		})
		_ = result // Use result to prevent optimization
	}
}

// BenchmarkOptimalCapacity tests optimal capacity calculation
func BenchmarkOptimalCapacity(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			capacity := GetOptimalArrayCapacity(size)
			arr := make([]Value, 0, capacity)
			for j := 0; j < size; j++ {
				arr = append(arr, Value(float64(j)))
			}
			_ = arr // Use arr to prevent optimization
		}
	}
}
