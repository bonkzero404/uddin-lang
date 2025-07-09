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