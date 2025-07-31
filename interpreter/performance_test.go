package interpreter

import (
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

// BenchmarkTokenizerLargeIdentifier tests performance with large identifiers
func BenchmarkTokenizerLargeIdentifier(b *testing.B) {
	// Create a large identifier to test strings.Builder efficiency
	longIdentifier := "very_long_identifier_name_that_tests_string_building_performance_with_many_characters"
	source := []byte(longIdentifier + " = 42")
	
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

// BenchmarkTokenizerStringLiterals tests performance with string literals
func BenchmarkTokenizerStringLiterals(b *testing.B) {
	source := []byte(`"This is a long string literal that tests the performance of string building in the tokenizer with escape sequences like \n and \t"`)
	
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

// BenchmarkParser tests parser performance with cached error messages
func BenchmarkParser(b *testing.B) {
	source := []byte(`
		x = 10
		if (x > 5) then:
			y = x * 2 + 1
			z = [y, y + 1, y + 2]
		end
		for (item in z):
			result = item * 2
		end
	`)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseProgram(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEvaluator tests evaluator performance with optimized type checking
func BenchmarkEvaluator(b *testing.B) {
	source := []byte(`
		x = [1, 2, 3, 4, 5]
		y = {"a": 1, "b": 2, "c": 3}
		z = x[2] + y["b"]
		w = "hello"[1]
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

// BenchmarkMathOperations tests performance of mathematical operations
func BenchmarkMathOperations(b *testing.B) {
	source := []byte(`
		x = 100
		y = 50
		result1 = x + y
		result2 = x - y
		result3 = x * y
		result4 = x / y
		result5 = x % y
		float_result = 3.14 * 2.71
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

// BenchmarkStringOperations tests performance of string operations
func BenchmarkStringOperations(b *testing.B) {
	source := []byte(`
		str1 = "Hello"
		str2 = "World"
		concatenated = str1 + " " + str2
		repeated = str1 * 5
		char_access = str1[0]
		length_check = len(concatenated)
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

// BenchmarkArrayOperations tests performance of array operations
func BenchmarkArrayOperations(b *testing.B) {
	source := []byte(`
		arr1 = [1, 2, 3, 4, 5]
		arr2 = [6, 7, 8, 9, 10]
		combined = arr1 + arr2
		repeated = arr1 * 3
		element = arr1[2]
		length = len(combined)
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

// BenchmarkMapOperations tests performance of map operations
func BenchmarkMapOperations(b *testing.B) {
	source := []byte(`
		map1 = {"a": 1, "b": 2, "c": 3}
		map2 = {"d": 4, "e": 5, "f": 6}
		combined = map1 + map2
		value = map1["b"]
		length = len(combined)
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

// BenchmarkFunctionCalls tests performance of function calls
func BenchmarkFunctionCalls(b *testing.B) {
	source := []byte(`
		fun square(x):
			return x * x
		end
		
		fun factorial(n):
			if (n <= 1) then:
				return 1
			else:
				return n * factorial(n - 1)
			end
		end
		
		result1 = square(10)
		result2 = factorial(5)
		result3 = max(result1, result2)
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

// BenchmarkLoopOperations tests performance of loop operations
func BenchmarkLoopOperations(b *testing.B) {
	source := []byte(`
		sum = 0
		for (i in range(1, 100)):
			sum = sum + i
		end
		
		arr = []
		for (i in range(1, 20)):
			arr = arr + [i * 2]
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
		big_array = []
		for (i in range(1, 100)):
			big_array = big_array + [i]
		end
		
		big_map = {}
		for (i in range(1, 50)):
			key = "key" + str(i)
			big_map = big_map + {key: i * i}
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

// BenchmarkNestedDataStructures tests performance with nested structures
func BenchmarkNestedDataStructures(b *testing.B) {
	source := []byte(`
		data = {
			"users": [
				{"id": 1, "name": "Alice", "scores": [85, 92, 78]},
				{"id": 2, "name": "Bob", "scores": [90, 88, 95]}
			],
			"metadata": {"version": "1.0", "count": 2}
		}
		result = data["users"][0]["scores"][1]
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

func BenchmarkStringConcatenation(b *testing.B) {
	source := []byte(`
		result = ""
		for (i in range(1, 50)):
			result = result + str(i) + ","
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

// BenchmarkComplexDataStructures tests performance with complex data structure manipulations
func BenchmarkComplexDataStructures(b *testing.B) {
	source := []byte(`
		data = []
		for (i in range(1, 50)):
			row = {}
			for (j in range(1, 10)):
				key = "col" + str(j)
				row[key] = i * j
			end
			data = data + [row]
		end
		
		sum = 0
		for (row in data):
			// Calculate sum by accessing known keys
			for (j in range(1, 10)):
				key = "col" + str(j)
				sum = sum + row[key]
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
		big_array = []
		for (i in range(1, 500)):
			big_array = big_array + [i]
		end
		
		filtered = []
		for (item in big_array):
			if (item % 2 == 0) then:
				filtered = filtered + [item * 2]
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