package interpreter

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkTokenizer tests tokenizer performance with optimized strings.Builder
func BenchmarkTokenizer(b *testing.B) {
	source := []byte(`
		x = 5 + 3 * 2
		y = "hello world"
		z = [1, 2, 3, 4, 5]
		if (x > 10) then:
			print("large number")
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenizer := NewTokenizer(source)
		for {
			_, tok, _ := tokenizer.Next()
			if tok == EOF {
				break
			}
		}
	}
}

// ========================================
// ADVANCED BENCHMARKS (from benchmark_test.go)
// ========================================

// BenchmarkAdvancedStringConcatenation tests optimized string concatenation with interning
func BenchmarkAdvancedStringConcatenation(b *testing.B) {
	source := []byte(`
		text = "hello"
		result = ""
		for (i in range(1, 100)):
			result = result + text + " world " + str(i)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedLargeArrayOperations tests optimized large array operations
func BenchmarkAdvancedLargeArrayOperations(b *testing.B) {
	source := []byte(`
		// Optimized version with smaller array size to reduce memory overhead
		big_array = []
		for (i in range(1, 500)):
			push(big_array, i)
		end

		// Test array operations without expensive multiplication
		// Instead of multiplying entire array, process elements individually
		doubled = []
		for (item in big_array):
			push(doubled, item * 2)
		end

		// Test filtering with optimized approach
		filtered = []
		for (item in big_array):
			if (item % 2 == 0) then:
				push(filtered, item * 2)
			end
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedNestedFunctionCalls tests optimized function calls with scope pooling
func BenchmarkAdvancedNestedFunctionCalls(b *testing.B) {
	source := []byte(`
		fun fibonacci(n):
			if (n <= 1) then:
				return n
			else:
				return fibonacci(n-1) + fibonacci(n-2)
			end
		end

		fun calculate(x, y):
			return fibonacci(x) + fibonacci(y)
		end

		result = calculate(10, 8)
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedStringRepetition tests optimized string repetition
func BenchmarkAdvancedStringRepetition(b *testing.B) {
	source := []byte(`
		text = "Hello World! "
		// Test large string repetition
		big_text = text * 100

		// Test multiple repetitions
		for (i in range(1, 10)):
			repeated = text * (i * 10)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedMapOperations tests optimized map merging
func BenchmarkAdvancedMapOperations(b *testing.B) {
	source := []byte(`
		map1 = {}
		for (i in range(1, 100)):
			key = "key" + str(i)
			map1 = map1 + {key: i}
		end

		map2 = {}
		for (i in range(100, 200)):
			key = "key" + str(i)
			map2 = map2 + {key: i * 2}
		end

		// Test large map merging
		combined = map1 + map2
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedComplexDataStructures tests optimized complex data structure operations
func BenchmarkAdvancedComplexDataStructures(b *testing.B) {
	source := []byte(`
		// Optimized version using direct assignment and push
		data = []
		for (i in range(1, 50)):
			row = {}
			for (j in range(1, 10)):
				key = "col" + str(j)
				row[key] = i * j
			end
			push(data, row)
		end

		// Process the data with optimized approach
		processed = []
		for (row in data):
			new_row = {}
			for (j in range(1, 5)):
				key = "col" + str(j)
				if (key in row) then:
					new_key = "new_" + key
					new_row[new_key] = row[key] * 2
				end
			end
			push(processed, new_row)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdvancedMemoryIntensive tests optimized memory allocation patterns
func BenchmarkAdvancedMemoryIntensive(b *testing.B) {
	source := []byte(`
		// Optimized large array creation and manipulation
		big_array = []
		for (i in range(1, 200)):
			push(big_array, i)
		end

		// Test array multiplication
		doubled = big_array * 2
		tripled = big_array * 3

		// Optimized large map creation
		big_map = {}
		for (i in range(1, 100)):
			key = "key" + str(i)
			value = [i, i*2, i*3]
			big_map[key] = value
		end

		// Test string operations
		big_string = "test" * 100
		for (i in range(1, 20)):
			big_string = big_string + " iteration " + str(i)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPerformanceMonitoring tests the performance monitoring overhead
func BenchmarkPerformanceMonitoring(b *testing.B) {
	source := []byte(`
		x = 5
		y = 10
		result = x + y
		result = result * 2
		result = result - 3
		result = result / 2
		text = "hello" + " world"
		arr = [1, 2, 3] + [4, 5, 6]
		map_data = {"a": 1} + {"b": 2}
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ========================================
// OPTIMIZATION BENCHMARKS
// ========================================

// Benchmark untuk ArrayConcatenator
func BenchmarkOptimizedArrayConcatenationNew(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := make([]Value, 0)
			for j := 0; j < 1000; j++ {
				_ = append(result, Value(j))
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			concat := GetArrayConcatenator()
			for j := 0; j < 1000; j++ {
				concat.Append(Value(j))
			}
			result := concat.ToArray()
			PutArrayConcatenator(concat)
			_ = result
		}
	})
}

// Benchmark untuk StringConcatenator
func BenchmarkOptimizedStringConcatenationNew(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result string
			for j := 0; j < 1000; j++ {
				result += fmt.Sprintf("item_%d,", j)
			}
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			for j := 0; j < 1000; j++ {
				builder.WriteString(fmt.Sprintf("item_%d,", j))
			}
			result := builder.String()
			_ = result
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			concat := GetStringConcatenator()
			for j := 0; j < 1000; j++ {
				concat.WriteString(fmt.Sprintf("item_%d,", j))
			}
			result := concat.String()
			PutStringConcatenator(concat)
			_ = result
		}
	})
}

// Benchmark untuk MapBuilder
func BenchmarkOptimizedMapCreationNew(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := make(map[string]Value)
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key_%d", j)
				result[key] = Value(j)
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			builder := GetMapBuilder()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key_%d", j)
				builder.Set(key, Value(j))
			}
			result := builder.ToMap()
			PutMapBuilder(builder)
			_ = result
		}
	})
}

// Benchmark untuk CallStack
func BenchmarkOptimizedCallStackNew(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stack := make([]map[string]any, 0)
			for j := 0; j < 100; j++ {
				callInfo := map[string]any{
					"function": fmt.Sprintf("func_%d", j),
					"depth":    j,
				}
				stack = append(stack, callInfo)
			}
			for j := 99; j >= 0; j-- {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			callStack := GetCallStack()
			for j := 0; j < 100; j++ {
				callInfo := map[string]any{
					"function": fmt.Sprintf("func_%d", j),
					"depth":    j,
				}
				callStack.PushCall(callInfo)
			}
			for j := 99; j >= 0; j-- {
				callStack.Pop()
			}
			PutCallStack(callStack)
		}
	})
}

// Benchmark untuk toString dengan optimasi
func BenchmarkOptimizedToStringNew(b *testing.B) {
	// Create test data
	testArray := make([]Value, 100)
	for i := 0; i < 100; i++ {
		testArray[i] = Value(fmt.Sprintf("item_%d", i))
	}

	testMap := make(map[string]Value)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		testMap[key] = Value(fmt.Sprintf("value_%d", i))
	}

	b.Run("Array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := toString(Value(&testArray), false)
			_ = result
		}
	})

	b.Run("Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := toString(Value(testMap), false)
			_ = result
		}
	})
}

// BenchmarkComplexExpression tests performance with complex nested expressions
func BenchmarkComplexExpression(b *testing.B) {
	source := []byte(`
		result = ((5 + 3) * 2 - 1) / (4 + 2) + (10 % 3)
		array = [result, result * 2, result * 3]
		map_data = {"first": array[0], "second": array[1], "third": array[2]}
		final = map_data["second"] + array[2]
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTypeChecking tests the performance of optimized type checking
func BenchmarkTypeChecking(b *testing.B) {
	values := []Value{
		nil,
		true,
		42,
		3.14,
		"hello",
		&[]Value{1, 2, 3},
		map[string]Value{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_ = GetValueType(v)
		}
	}
}

// BenchmarkValueExtraction tests the performance of optimized value extraction
func BenchmarkValueExtraction(b *testing.B) {
	values := []Value{
		42,
		3.14,
		"hello",
		&[]Value{1, 2, 3},
		map[string]Value{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_, _ = asInt(v)
			_, _ = asFloat(v)
			_, _ = asString(v)
			_, _ = asArray(v)
			_, _ = asMap(v)
		}
	}
}

// BenchmarkLoopOperations tests performance of loop operations
func BenchmarkLoopOperations(b *testing.B) {
	source := []byte(`
		sum = 0
		for (i in range(1, 100)):
			sum = sum + i
		end

		// Optimized array building using push
		arr = []
		for (i in range(1, 20)):
			push(arr, i * 2)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConditionalOperations tests performance of conditional operations
func BenchmarkConditionalOperations(b *testing.B) {
	source := []byte(`
		x = 50
		result = (x > 30) ? x * 2 : x / 2

		for (i in range(1, 50)):
			if (i % 2 == 0) then:
				even_result = i * 3
			else:
				odd_result = i + 1
			end
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuiltinFunctions tests performance of built-in functions
func BenchmarkBuiltinFunctions(b *testing.B) {
	source := []byte(`
		arr = [5, 2, 8, 1, 9, 3]
		sorted_arr = sort(arr)
		max_val = max(arr)
		min_val = min(arr)
		sum_val = sum(arr)
		length = len(arr)
		reversed = reverse(arr)
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryIntensive tests memory allocation patterns
func BenchmarkMemoryIntensive(b *testing.B) {
	source := []byte(`
		// Optimized version using push and direct assignment
		big_array = []
		for (i in range(1, 100)):
			push(big_array, i)
		end

		big_map = {}
		for (i in range(1, 50)):
			key = "key" + str(i)
			big_map[key] = i * i
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Additional benchmarks for specific optimizations
func BenchmarkAppendOptimization(b *testing.B) {
	source := []byte(`
		arr = []
		for (i in range(1, 100)):
			arr = arr + [i]
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTypeAssertionOptimization(b *testing.B) {
	source := []byte(`
		x = 42
		y = 3.14
		z = "hello"
		for (i in range(1, 100)):
			a = x + 1
			b = y * 2.0
			c = z + "world"
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapAccessOptimization(b *testing.B) {
	source := []byte(`
		data = {"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
		for (i in range(1, 100)):
			x = data["a"] + data["b"] + data["c"]
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNestedFunctionCalls tests performance with nested function calls
func BenchmarkNestedFunctionCalls(b *testing.B) {
	source := []byte(`
		fun add(a, b):
			return a + b
		end

		fun multiply(a, b):
			return a * b
		end

		fun complex_calc(x):
			return add(multiply(x, 2), add(x, 1))
		end

		result = 0
		for (i in range(1, 100)):
			result = result + complex_calc(i)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRecursiveOperations tests performance with recursive operations
func BenchmarkRecursiveOperations(b *testing.B) {
	source := []byte(`
		fun fibonacci(n):
			if (n <= 1) then:
				return n
			else:
				return fibonacci(n - 1) + fibonacci(n - 2)
			end
		end

		result = 0
		for (i in range(1, 15)):
			result = result + fibonacci(i)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStringManipulation tests performance with extensive string operations
func BenchmarkStringManipulation(b *testing.B) {
	source := []byte(`
		text = "hello world"
		result = ""
		for (i in range(1, 100)):
			result = result + text + " " + str(i)
			if (len(result) > 1000) then:
				result = substr(result, 0, 500)
			end
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVariableScoping tests performance with nested variable scoping
func BenchmarkVariableScoping(b *testing.B) {
	source := []byte(`
		global_var = 0

		fun nested_function(depth):
			local_var = depth
			if (depth > 0) then:
				return local_var + nested_function(depth - 1)
			else:
				return local_var
			end
		end

		result = 0
		for (i in range(1, 50)):
			result = result + nested_function(i % 10)
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeArrayOperations tests performance with large array manipulations
func BenchmarkLargeArrayOperations(b *testing.B) {
	source := []byte(`
		// Optimized version using push instead of concatenation
		big_array = []
		for (i in range(1, 500)):
			push(big_array, i)
		end

		filtered = []
		for (item in big_array):
			if (item % 2 == 0) then:
				push(filtered, item * 2)
			end
		end

		sum = 0
		for (item in filtered):
			sum = sum + item
		end
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		program, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}

		_, err = Execute(program, &Config{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
