package interpreter

import (
	"testing"
)

// ===== FastSlice Tests =====

// BenchmarkBuiltinAppend tests the performance of Go's built-in append
func BenchmarkBuiltinAppend(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var slice []any
		for j := 0; j < 1000; j++ {
			_ = append(slice, j)
		}
	}
}

// BenchmarkFastSliceAppend tests the performance of FastSlice
func BenchmarkFastSliceAppend(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := NewFastSlice(1000)
		for j := 0; j < 1000; j++ {
			fs.FastAppend(j)
		}
		_ = fs.ToSlice()
	}
}

// BenchmarkBuiltinAppendLarge tests built-in append with larger datasets
func BenchmarkBuiltinAppendLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var slice []any
		for j := 0; j < 10000; j++ {
			_ = append(slice, j)
		}
	}
}

// BenchmarkFastSliceAppendLarge tests FastSlice with larger datasets
func BenchmarkFastSliceAppendLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := NewFastSlice(10000)
		for j := 0; j < 10000; j++ {
			fs.FastAppend(j)
		}
		_ = fs.ToSlice()
	}
}

// BenchmarkBuiltinAppendBatch tests built-in append with batch operations
func BenchmarkBuiltinAppendBatch(b *testing.B) {
	batch := make([]any, 100)
	for i := range batch {
		batch[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var slice []any
		for j := 0; j < 100; j++ {
			_ = append(slice, batch...)
		}
	}
}

// BenchmarkFastSliceAppendBatch tests FastSlice with batch operations
func BenchmarkFastSliceAppendBatch(b *testing.B) {
	batch := make([]any, 100)
	for i := range batch {
		batch[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := NewFastSlice(10000)
		for j := 0; j < 100; j++ {
			fs.FastAppend(batch...)
		}
		_ = fs.ToSlice()
	}
}

// BenchmarkBatchAppend tests the BatchAppend function
func BenchmarkBatchAppend(b *testing.B) {
	slices := make([][]any, 10)
	for i := range slices {
		slices[i] = make([]any, 100)
		for j := range slices[i] {
			slices[i][j] = i*100 + j
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var base []any
		base = BatchAppend(base, slices...)
		_ = base
	}
}

// BenchmarkBuiltinAppendMultipleSlices tests built-in append with multiple slices
func BenchmarkBuiltinAppendMultipleSlices(b *testing.B) {
	slices := make([][]any, 10)
	for i := range slices {
		slices[i] = make([]any, 100)
		for j := range slices[i] {
			slices[i][j] = i*100 + j
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var base []any
		for _, s := range slices {
			base = append(base, s...)
		}
		_ = base
	}
}

// BenchmarkMemoryReallocation tests memory reallocation patterns
func BenchmarkMemoryReallocation(b *testing.B) {
	b.Run("BuiltinAppend", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var slice []any
			// Start small and grow to trigger multiple reallocations
			for j := 0; j < 2048; j++ {
				_ = append(slice, j)
			}
		}
	})

	b.Run("FastSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fs := NewFastSlice(16) // Start small
			for j := 0; j < 2048; j++ {
				fs.FastAppend(j)
			}
			_ = fs.ToSlice()
		}
	})

	b.Run("PreAllocated", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]any, 0, 2048) // Pre-allocate
			for j := 0; j < 2048; j++ {
				_ = append(slice, j)
			}
		}
	})
}

// TestFastAppendCorrectness tests correctness of FastSlice implementations
func TestFastAppendCorrectness(t *testing.T) {
	// Test FastSlice
	fs := NewFastSlice(4)
	fs.FastAppend(1, 2, 3)
	fs.FastAppend(4, 5)

	result := fs.ToSlice()
	expected := []any{1, 2, 3, 4, 5}

	if len(result) != len(expected) {
		t.Errorf("Length mismatch: got %d, want %d", len(result), len(expected))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Value mismatch at index %d: got %v, want %v", i, result[i], v)
		}
	}

	// Test BatchAppend
	slice1 := []any{1, 2}
	slice2 := []any{3, 4}
	slice3 := []any{5}

	batchResult := BatchAppend(nil, slice1, slice2, slice3)

	if len(batchResult) != len(expected) {
		t.Errorf("BatchAppend length mismatch: got %d, want %d", len(batchResult), len(expected))
	}

	for i, v := range expected {
		if batchResult[i] != v {
			t.Errorf("BatchAppend value mismatch at index %d: got %v, want %v", i, batchResult[i], v)
		}
	}
}

// BenchmarkAppendStrategiesComparison compares all append strategies
func BenchmarkAppendStrategiesComparison(b *testing.B) {
	b.Run("BuiltinAppend-1K", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var slice []any
			for j := 0; j < 1000; j++ {
				_ = append(slice, j)
			}
		}
	})

	b.Run("FastSlice-1K", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fs := NewFastSlice(1000)
			for j := 0; j < 1000; j++ {
				fs.FastAppend(j)
			}
			_ = fs.ToSlice()
		}
	})

	b.Run("HybridStrategy-1K", func(b *testing.B) {
		strategy := NewHybridAppendStrategy(1000)
		for i := 0; i < b.N; i++ {
			var slice []Value
			for j := 0; j < 1000; j++ {
				slice = strategy.AppendValues(slice, float64(j))
			}
		}
	})

	b.Run("SmartAppend-1K", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var slice []Value
			for j := 0; j < 1000; j++ {
				slice = SmartAppend(slice, float64(j))
			}
		}
	})
}
