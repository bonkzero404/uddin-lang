package uddin

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestEngine_New(t *testing.T) {
	engine := New()
	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
	if engine.config == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestEngine_ExecuteString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name:     "simple_print",
			source:   `print("Hello, World!")`,
			expected: "Hello, World!",
		},
		{
			name:     "arithmetic",
			source:   `print(5 + 3)`,
			expected: "8",
		},
		{
			name: "variable_assignment",
			source: `x = 10
print(x * 2)`,
			expected: "20",
		},
		{
			name:     "string_concatenation",
			source:   `print("Hello" + " " + "World")`,
			expected: "Hello World",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			engine := New()
			engine.SetStdout(&buf)
			engine.SetUnitTestMode(true)

			stats, err := engine.ExecuteString(test.source)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if stats == nil {
				t.Fatal("Expected non-nil stats")
			}

			output := buf.String()
			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output to contain %q, got %q", test.expected, output)
			}
		})
	}
}

func TestEngine_EvaluateString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected any
	}{
		{
			name:     "integer",
			source:   "42",
			expected: 42,
		},
		{
			name:     "string",
			source:   `"hello"`,
			expected: "hello",
		},
		{
			name:     "boolean",
			source:   "true",
			expected: true,
		},
		{
			name:     "arithmetic",
			source:   "5 + 3 * 2",
			expected: 11,
		},
		{
			name:     "comparison",
			source:   "10 > 5",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine := New()
			engine.SetUnitTestMode(true)

			value, stats, err := engine.EvaluateString(test.source)
			if err != nil {
				t.Fatalf("Evaluation failed: %v", err)
			}

			if stats == nil {
				t.Fatal("Expected non-nil stats")
			}

			if value != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, value)
			}
		})
	}
}

func TestEngine_SetVariable(t *testing.T) {
	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	// Set a variable
	engine.SetVariable("x", 42)

	// Use the variable in code
	stats, err := engine.ExecuteString(`print(x)`)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	output := buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("Expected output to contain '42', got %q", output)
	}
}

func TestEngine_SetVariables(t *testing.T) {
	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	// Set multiple variables
	vars := map[string]any{
		"name":   "UDDIN",
		"age":    25,
		"active": true,
	}
	engine.SetVariables(vars)

	// Use the variables in code
	stats, err := engine.ExecuteString(`
		print(name)
		print(age)
		print(active)
	`)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	output := buf.String()
	expectedValues := []string{"UDDIN", "25", "true"}
	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

func TestEngine_ParseExpression(t *testing.T) {
	engine := New()

	expr, err := engine.ParseExpression([]byte("5 + 3"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if expr == nil {
		t.Fatal("Expected non-nil expression")
	}

	// Evaluate the parsed expression
	value, stats, err := engine.EvaluateExpression(expr)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if value != 8 {
		t.Errorf("Expected 8, got %v", value)
	}
}

func TestEngine_ParseProgram(t *testing.T) {
	engine := New()

	prog, err := engine.ParseProgram([]byte(`
		x = 5
		y = 10
		print(x + y)
	`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if prog == nil {
		t.Fatal("Expected non-nil program")
	}

	if len(prog.Statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(prog.Statements))
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test ParseExpression convenience function
	expr, err := ParseExpression([]byte("10 * 2"))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}
	if expr == nil {
		t.Fatal("Expected non-nil expression")
	}

	// Test ParseProgram convenience function
	prog, err := ParseProgram([]byte("x = 42"))
	if err != nil {
		t.Fatalf("ParseProgram failed: %v", err)
	}
	if prog == nil {
		t.Fatal("Expected non-nil program")
	}

	// Test EvaluateString convenience function
	value, stats, err := EvaluateString("7 * 6")
	if err != nil {
		t.Fatalf("EvaluateString failed: %v", err)
	}
	if value != 42 {
		t.Errorf("Expected 42, got %v", value)
	}
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Test ExecuteString convenience function
	stats, err = ExecuteString(`print("test")`)
	if err != nil {
		t.Fatalf("ExecuteString failed: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
}

func TestEngine_FunctionDefinitionAndCall(t *testing.T) {
	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	source := `
		fun add(a, b):
			return a + b
		end

		result = add(5, 3)
		print(result)
	`

	stats, err := engine.ExecuteString(source)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	output := buf.String()
	if !strings.Contains(output, "8") {
		t.Errorf("Expected output to contain '8', got %q", output)
	}
}

func TestEngine_ControlFlow(t *testing.T) {
	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	source := `
		x = 10
		if (x > 5) then:
			print("x is greater than 5")
		else:
			print("x is not greater than 5")
		end

		for (i in range(3)):
			print("Loop: " + str(i))
		end
	`

	stats, err := engine.ExecuteString(source)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	output := buf.String()
	expectedOutputs := []string{
		"x is greater than 5",
		"Loop: 0",
		"Loop: 1",
		"Loop: 2",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

func TestEngine_DataStructures(t *testing.T) {
	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	source := `
		// Array operations
		arr = [1, 2, 3, 4, 5]
		print("Array length: " + str(len(arr)))
		print("First element: " + str(arr[0]))

		// Map operations
		person = {"name": "John", "age": 30}
		print("Name: " + person["name"])
		print("Age: " + str(person["age"]))
	`

	stats, err := engine.ExecuteString(source)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	output := buf.String()
	expectedOutputs := []string{
		"Array length: 5",
		"First element: 1",
		"Name: John",
		"Age: 30",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

func TestEngine_ErrorHandling(t *testing.T) {
	engine := New()
	engine.SetUnitTestMode(true)

	// Test syntax error
	_, err := engine.ExecuteString("x = ")
	if err == nil {
		t.Error("Expected syntax error")
	}

	// Test runtime error
	_, err = engine.ExecuteString("print(undefined_variable)")
	if err == nil {
		t.Error("Expected runtime error")
	}
}

// Test AST conversion functions
func TestEngine_ASTConversion(t *testing.T) {
	engine := New()

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "simple_assignment",
			source: "x = 42",
		},
		{
			name:   "function_definition",
			source: "fun add(a, b): return a + b end",
		},
		{
			name:   "if_statement",
			source: "x = 5\nif (x > 0) then:\n    print(\"positive\")\nend",
		},
		{
			name:   "complex_expression",
			source: "result = (x + y) * 2 - z / 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to JSON
			jsonData, err := engine.ConvertStringToJSON(tt.source)
			if err != nil {
				t.Fatalf("ConvertStringToJSON failed: %v", err)
			}

			// Verify JSON is valid
			if len(jsonData) == 0 {
				t.Error("JSON data is empty")
			}

			// Convert back to source code
			convertedSource, err := engine.ConvertJSONToString(jsonData)
			if err != nil {
				t.Fatalf("ConvertJSONToString failed: %v", err)
			}

			// Verify converted source is not empty
			if len(convertedSource) == 0 {
				t.Error("Converted source is empty")
			}

			// Test that both versions can be parsed successfully
			_, err = engine.ParseProgram([]byte(tt.source))
			if err != nil {
				t.Fatalf("Original source parsing failed: %v", err)
			}

			_, err = engine.ParseProgram([]byte(convertedSource))
			if err != nil {
				t.Fatalf("Converted source parsing failed: %v", err)
			}
		})
	}
}

// Test convenience functions for AST conversion
func TestConvenienceASTFunctions(t *testing.T) {
	source := "x = 10\ny = 20\nresult = x + y"

	// Test ConvertStringToJSON convenience function
	jsonData, err := ConvertStringToJSON(source)
	if err != nil {
		t.Fatalf("ConvertStringToJSON failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data is empty")
	}

	// Test ConvertJSONToString convenience function
	convertedSource, err := ConvertJSONToString(jsonData)
	if err != nil {
		t.Fatalf("ConvertJSONToString failed: %v", err)
	}

	if len(convertedSource) == 0 {
		t.Error("Converted source is empty")
	}

	// Test byte array versions
	jsonData2, err := ConvertToJSON([]byte(source))
	if err != nil {
		t.Fatalf("ConvertToJSON failed: %v", err)
	}

	convertedSource2, err := ConvertFromJSON(jsonData2)
	if err != nil {
		t.Fatalf("ConvertFromJSON failed: %v", err)
	}

	// Both methods should produce similar results
	if len(convertedSource2) == 0 {
		t.Error("Converted source from byte array is empty")
	}
}

// Test AST conversion error handling
func TestEngine_ASTConversionErrors(t *testing.T) {
	engine := New()

	// Test invalid source code
	_, err := engine.ConvertStringToJSON("invalid @#$% syntax here")
	if err == nil {
		t.Error("Expected error for invalid syntax")
	}

	// Test invalid JSON
	invalidJSON := []byte("{invalid json}")
	_, err = engine.ConvertJSONToString(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// Test round-trip conversion with execution
func TestEngine_ASTRoundTripExecution(t *testing.T) {
	engine := New()

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "arithmetic",
			source: "x = 5\ny = 3\nresult = x * y + 2",
		},
		{
			name:   "function_call",
			source: "fun double(n): return n * 2 end\nresult = double(21)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute original
			originalStats, err := engine.ExecuteString(tt.source)
			if err != nil {
				t.Fatalf("Original execution failed: %v", err)
			}

			// Convert to JSON and back
			jsonData, err := engine.ConvertStringToJSON(tt.source)
			if err != nil {
				t.Fatalf("JSON conversion failed: %v", err)
			}

			convertedSource, err := engine.ConvertJSONToString(jsonData)
			if err != nil {
				t.Fatalf("Source conversion failed: %v", err)
			}

			// Execute converted
			convertedStats, err := engine.ExecuteString(convertedSource)
			if err != nil {
				t.Fatalf("Converted execution failed: %v", err)
			}

			// Both should execute successfully (we don't compare exact stats as they may differ slightly)
			if originalStats == nil || convertedStats == nil {
				t.Error("Execution stats should not be nil")
			}
		})
	}
}

func BenchmarkEngine_ExecuteString(b *testing.B) {
	engine := New()
	engine.SetUnitTestMode(true)

	source := `
		fun fibonacci(n):
			if (n <= 1) then:
				return n
			end
			return fibonacci(n - 1) + fibonacci(n - 2)
		end

		result = fibonacci(10)
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ExecuteString(source)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

func BenchmarkEngine_EvaluateString(b *testing.B) {
	engine := New()
	engine.SetUnitTestMode(true)

	source := "(5 + 3) * 2 - 1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := engine.EvaluateString(source)
		if err != nil {
			b.Fatalf("Evaluation failed: %v", err)
		}
	}
}

func TestEngineCache_SameResultTwice(t *testing.T) {
	engine := New()
	engine.SetUnitTestMode(true)

	src := `score + 1`
	prog, err := engine.ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}

	for i, wantScore := range []int{10, 20} {
		engine.SetVariable("score", wantScore)
		_, err := engine.ExecuteProgram(prog)
		if err != nil {
			t.Fatalf("exec %d: %v", i, err)
		}
	}
}

func TestEngineCache_ConcurrentSafe(t *testing.T) {
	src := `x = 1
x`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := New()
			e.SetUnitTestMode(true)
			_, err := e.ExecuteProgram(prog)
			if err != nil {
				t.Errorf("concurrent exec: %v", err)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkEngine_NoCacheVsCache(b *testing.B) {
	src := `
fun fib(n: int):
    if (n <= 1) then:
        return n
    end
    return fib(n - 1) + fib(n - 2)
end
fib(20)
`
	b.Run("ParseAndExecute", func(b *testing.B) {
		engine := New()
		engine.SetUnitTestMode(true)
		for b.Loop() {
			_, err := engine.ExecuteString(src)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CachedExecute", func(b *testing.B) {
		engine := New()
		engine.SetUnitTestMode(true)
		prog, _ := engine.ParseProgram([]byte(src))
		b.ResetTimer()
		for b.Loop() {
			_, err := engine.ExecuteProgram(prog)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestStdlibModulesRegistered(t *testing.T) {
	// Verify all 9 stdlib modules are registered and callable via import syntax.
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"regex", `import "regex"\nprint(regex.is_match("^\\d+$", "42"))`, "true"},
		{"datetime", `import "datetime"\nprint(typeof(datetime.now()))`, "string"},
		{"json", `import "json"\nprint(json.stringify(42))`, "42"},
		{"fs", `import "fs"\nprint(typeof(fs.getcwd()))`, "string"},
		{"waf", `import "waf"\nprint(waf.cidr_match("192.168.1.1", "192.168.1.0/24"))`, "true"},
		{"cdc", `import "cdc"\ncdc.emit("test", {})\nprint(cdc.count("test"))`, "1"},
		{"fact", `import "fact"\nfact.assert("cat", "k")\nprint(fact.exists("cat", "k"))`, "true"},
		{"http", `import "http"\nprint(typeof(http.get))`, "function"},
		{"database", `import "database"\nprint(typeof(database.connect))`, "function"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			engine.SetUnitTestMode(true)
			src := strings.ReplaceAll(tt.src, `\n`, "\n")
			var buf strings.Builder
			engine.SetStdout(&buf)
			_, err := engine.ExecuteString(src)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEngine_NewWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.IsUnitTest = true

	engine := NewWithConfig(config)
	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	// Verify it can execute code
	var buf bytes.Buffer
	engine.SetStdout(&buf)
	_, err := engine.ExecuteString(`print("cfg ok")`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !strings.Contains(buf.String(), "cfg ok") {
		t.Errorf("expected output to contain 'cfg ok', got %q", buf.String())
	}
}

func TestEngine_ExecuteFile(t *testing.T) {
	// Write a temp .din script
	f, err := os.CreateTemp("", "uddin_test_*.din")
	if err != nil {
		t.Fatal("create temp file:", err)
	}
	defer os.Remove(f.Name())

	_, err = f.WriteString(`print("hello from file")`)
	if err != nil {
		t.Fatal("write temp file:", err)
	}
	f.Close()

	var buf bytes.Buffer
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	_, err = engine.ExecuteFile(f.Name())
	if err != nil {
		t.Fatalf("ExecuteFile failed: %v", err)
	}

	if !strings.Contains(buf.String(), "hello from file") {
		t.Errorf("expected 'hello from file' in output, got %q", buf.String())
	}
}
