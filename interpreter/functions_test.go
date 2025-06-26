package interpreter

import (
	"strings"
	"testing"
)

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{"print", `print("test")`, "test"},
		{"len_string", `print(len("hello"))`, "5"},
		{"len_array", `print(len([1, 2, 3]))`, "3"},
		{"len_map", `print(len({"a": 1, "b": 2}))`, "2"},
		{"typeof_int", `print(typeof(42))`, "int"},
		{"typeof_float", `print(typeof(3.14))`, "float"},
		{"typeof_string", `print(typeof("hello"))`, "string"},
		{"typeof_bool", `print(typeof(true))`, "bool"},
		{"typeof_array", `print(typeof([1, 2, 3]))`, "array"},
		{"typeof_map", `print(typeof({"a": 1}))`, "object"},
		{"typeof_function", `fun test(): return 1 end
		print(typeof(test))`, "function"},
		{"str_int", `print(str(42))`, "42"},
		{"str_float", `print(str(3.14))`, "3.14"},
		{"str_bool", `print(str(true))`, "true"},
		{"int_string", `print(int("42"))`, "42"},
		{"range_single", `arr = range(3)
		print(join(arr, ", "))`, "0, 1, 2"},
		{"split", `arr = split("a,b,c", ",")
		print(join(arr, " "))`, `"a" "b" "c"`},
		{"join", `print(join([1, 2, 3], ", "))`, "1, 2, 3"},
		{"substr", `print(substr("hello", 1, 3))`, "el"},
		{"contains_string", `print(contains("hello", "ell"))`, "true"},
		{"contains_array", `print(contains([1, 2, 3], 2))`, "true"},
		{"str_pad", `print(str_pad("5", 2, "0"))`, "500"},
		{"is_regex_match", `print(is_regex_match("^\\d+$", "12345"))`, "true"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Jalankan program
			success, output := RunProgram(test.program)

			if !success {
				t.Fatalf("Program execution failed: %s", output)
			}

			// Periksa output
			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output to contain %q, but got: %s", test.expected, output)
			}
		})
	}
}

func TestUserDefinedFunctions(t *testing.T) {
	program := `
	fun add(a, b):
		return a + b
	end

	fun factorial(n):
		if (n <= 1) then:
			return 1
		else:
			return n * factorial(n - 1)
		end
	end

	fun fibonacci(n):
		if (n <= 1) then:
			return n
		else:
			return fibonacci(n - 1) + fibonacci(n - 2)
		end
	end

	print("add result: " + str(add(5, 3)))
	print("factorial result: " + str(factorial(5)))
	print("fibonacci result: " + str(fibonacci(7)))
	`

	success, output := RunProgram(program)
	if !success {
		t.Fatalf("Program execution failed: %s", output)
	}

	// Periksa hasil add
	if !strings.Contains(output, "add result: 8") {
		t.Errorf("add function failed, expected 8, got: %s", output)
	}

	// Periksa hasil factorial
	if !strings.Contains(output, "factorial result: 120") {
		t.Errorf("factorial function failed, expected 120, got: %s", output)
	}

	// Periksa hasil fibonacci
	if !strings.Contains(output, "fibonacci result: 13") {
		t.Errorf("fibonacci function failed, expected 13, got: %s", output)
	}
}

func TestFunctionErrors(t *testing.T) {
	tests := []struct {
		name        string
		program     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "too_few_arguments",
			program: `
			fun add(a, b):
				return a + b
			end
			add(1)
			`,
			expectError: true,
			errorMsg:    "requires 2 args, got 1",
		},
		{
			name: "too_many_arguments",
			program: `
			fun add(a, b):
				return a + b
			end
			add(1, 2, 3)
			`,
			expectError: true,
			errorMsg:    "requires 2 args, got 3",
		},
		{
			name: "undefined_function",
			program: `
			undefined_function()
			`,
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "recursion_limit",
			program: `
			fun factorial(n):
				if (n <= 1) then:
					return 1
				else:
					return n * factorial(n - 1)
				end
			end

			fun main():
				print(factorial(5))
			end
			`,
			expectError: false,
			errorMsg:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			success, output := RunProgram(test.program)

			if test.expectError {
				if success {
					t.Errorf("Expected error, but program succeeded")
				} else if !strings.Contains(output, test.errorMsg) {
					t.Errorf("Expected error message containing '%s', got: %s", test.errorMsg, output)
				}
			} else {
				if !success {
					t.Errorf("Program execution failed: %s", output)
				}
			}
		})
	}
}

func TestAutoRunMainFunction(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "auto_run_main",
			program: `
			fun main():
				print("Main function executed automatically")
			end
			`,
			expected: "Main function executed automatically",
		},
		{
			name: "auto_run_main_with_other_functions",
			program: `
			fun helper():
				return "helper called"
			end

			fun main():
				print("Main with helper: " + helper())
			end
			`,
			expected: "Main with helper: helper called",
		},
		{
			name: "auto_run_main_with_arguments",
			program: `
			fun greet(name):
				return "Hello, " + name
			end

			fun main():
				print(greet("Uddin-Lang"))
			end
			`,
			expected: "Hello, Uddin-Lang",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Jalankan program
			success, output := RunProgram(test.program)

			if !success {
				t.Fatalf("Program execution failed: %s", output)
			}

			// Periksa output
			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output to contain %q, but got: %s", test.expected, output)
			}
		})
	}
}
