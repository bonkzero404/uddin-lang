package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

func TestInterpreterBasicExecution(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "variable_assignment",
			program: `
			x = 5
			y = 10
			z = x + y
			print(z)
			`,
			expected: "15",
		},
		{
			name: "arithmetic_operations",
			program: `
			print(5 + 3)
			print(5 - 3)
			print(5 * 3)
			print(5 / 2)
			print(5 % 2)
			`,
			expected: "8\n2\n15\n2.5\n1",
		},
		{
			name: "comparison_operations",
			program: `
			print(5 == 5)
			print(5 != 3)
			print(5 > 3)
			print(5 >= 5)
			print(3 < 5)
			print(5 <= 5)
			`,
			expected: "true\ntrue\ntrue\ntrue\ntrue\ntrue",
		},
		{
			name: "logical_operations",
			program: `
			print(true and true)
			print(true and false)
			print(true or false)
			print(false or false)
			print(not true)
			print(not false)
			`,
			expected: "true\nfalse\ntrue\nfalse\nfalse\ntrue",
		},
		{
			name: "string_operations",
			program: `
			print("Hello" + " " + "World")
			print("Hello" * 3)
			print("Hello" == "Hello")
			print("Hello" != "World")
			`,
			expected: "Hello World\nHelloHelloHello\ntrue\ntrue",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Stdout: &buf,
			}

			prog, err := ParseProgram([]byte(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			_, err = Execute(prog, config)
			if err != nil {
				t.Fatalf("Failed to execute program: %v", err)
			}

			output := buf.String()
			// Normalize line endings
			output = strings.ReplaceAll(output, "\r\n", "\n")
			expected := strings.ReplaceAll(test.expected, "\r\n", "\n")

			// Check each expected line
			expectedLines := strings.Split(expected, "\n")
			for _, line := range expectedLines {
				if line != "" && !strings.Contains(output, line) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", line, output)
				}
			}
		})
	}
}

func TestInterpreterControlStructures(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "if_statement",
			program: `
			x = 10
			if (x > 5) then:
				print("x is greater than 5")
			else:
				print("x is not greater than 5")
			end
			`,
			expected: "x is greater than 5",
		},
		{
			name: "if_else_if_statement",
			program: `
			x = 3
			if (x > 5) then:
				print("x is greater than 5")
			else if (x > 0) then:
				print("x is positive but not greater than 5")
			else:
				print("x is not positive")
			end
			`,
			expected: "x is positive but not greater than 5",
		},
		{
			name: "while_loop",
			program: `
			i = 0
			sum = 0
			while (i < 5):
				sum = sum + i
				i = i + 1
			end
			print(sum)
			`,
			expected: "10",
		},
		{
			name: "for_loop",
			program: `
			sum = 0
			for (i in range(5)):
				sum = sum + i
			end
			print(sum)
			`,
			expected: "10",
		},
		{
			name: "nested_loops",
			program: `
			result = ""
			for (i in range(3)):
				for (j in range(3)):
					result = result + str(i) + str(j) + " "
				end
			end
			print(result)
			`,
			expected: "00 01 02 10 11 12 20 21 22",
		},
		{
			name: "range_with_two_args",
			program: `
			result = ""
			for (i in range(1, 4)):
				result = result + str(i) + " "
			end
			print(result)
			`,
			expected: "1 2 3",
		},
		{
			name: "range_edge_cases",
			program: `
			// Test range(0)
			print(len(range(0)))
			// Test range(3, 3) - should be empty
			print(len(range(3, 3)))
			// Test range(5, 2) - should be empty
			print(len(range(5, 2)))
			`,
			expected: "0\n0\n0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Stdout: &buf,
			}

			prog, err := ParseProgram([]byte(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			_, err = Execute(prog, config)
			if err != nil {
				t.Fatalf("Failed to execute program: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, test.expected) {
				t.Errorf("Expected output to contain '%s', but got:\n%s", test.expected, output)
			}
		})
	}
}

func TestInterpreterDataStructures(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "array_operations",
			program: `
			arr = [1, 2, 3, 4, 5]
			print(arr[0])
			print(arr[4])
			print(len(arr))
			arr[2] = 10
			print(arr[2])
			print(join(arr, ", "))
			`,
			expected: "1\n5\n5\n10\n1, 2, 10, 4, 5",
		},
		{
			name: "map_operations",
			program: `
			obj = {"name": "John", "age": 30}
			print(obj["name"])
			print(obj["age"])
			print(len(obj))
			obj["city"] = "New York"
			print(obj["city"])
			print(typeof(obj))
			`,
			expected: "John\n30\n2\nNew York\nobject",
		},
		{
			name: "array_methods",
			program: `
			arr = [1, 2, 3]
			arr = arr + [4]
			print(join(arr, ", "))
			print(len(arr))
			`,
			expected: "1, 2, 3, 4\n4",
		},
		{
			name: "enhanced_object_syntax",
			program: `
			// Test different object key syntaxes
			obj1 = {"name": "John", "age": 30}
			obj2 = {'name': "Alice", 'age': 25}
			obj3 = {name: "Bob", age: 35}
			obj4 = {"name": "Carol", 'age': 28, city: "Miami"}

			print(obj1["name"])
			print(obj2["name"])
			print(obj3["name"])
			print(obj4["name"])
			print(obj4["city"])
			`,
			expected: "John\nAlice\nBob\nCarol\nMiami",
		},
		{
			name: "negative_indexing",
			program: `
			// Test negative indexing for arrays
			arr = [1, 2, 3, 4, 5]
			print(arr[-1])  // Last element
			print(arr[-2])  // Second to last
			print(arr[-5])  // First element

			// Test negative indexing for strings
			text = "Hello"
			print(text[-1])  // Last char
			print(text[-2])  // Second to last char
			`,
			expected: "5\n4\n1\no\nl",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Stdout: &buf,
			}

			prog, err := ParseProgram([]byte(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			_, err = Execute(prog, config)
			if err != nil {
				t.Fatalf("Failed to execute program: %v", err)
			}

			output := buf.String()
			// Normalize line endings
			output = strings.ReplaceAll(output, "\r\n", "\n")
			expected := strings.ReplaceAll(test.expected, "\r\n", "\n")

			// Check each expected line
			expectedLines := strings.Split(expected, "\n")
			for _, line := range expectedLines {
				if line != "" && !strings.Contains(output, line) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", line, output)
				}
			}
		})
	}
}

func TestInterpreterErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		program      string
		expectError  bool
		errorMessage string
	}{
		{
			name: "division_by_zero",
			program: `
			try:
				x = 5 / 0
				print("This should not be printed")
			catch (err):
				print("Caught error: division by zero")
			end
			`,
			expectError:  false,
			errorMessage: "Caught error: division by zero",
		},
		{
			name: "array_index_out_of_bounds",
			program: `
			try:
				arr = [1, 2, 3]
				print(arr[5])
			catch (err):
				print("Caught error: index out of range")
			end
			`,
			expectError:  false,
			errorMessage: "Caught error: index out of range",
		},
		{
			name: "undefined_variable",
			program: `
			try:
				print(undefined_var)
			catch (err):
				print("Caught error: undefined variable")
			end
			`,
			expectError:  false,
			errorMessage: "Caught error: undefined variable",
		},
		{
			name: "type_error",
			program: `
			try:
				print("hello" - 5)
			catch (err):
				print("Caught error: invalid operation")
			end
			`,
			expectError:  false,
			errorMessage: "Caught error: invalid operation",
		},
		{
			name: "uncaught_error",
			program: `
			x = 5 / 0
			`,
			expectError:  true,
			errorMessage: "can't divide by zero",
		},
		{
			name: "nested_try_catch",
			program: `
			try:
				try:
					x = 5 / 0
				catch (err1):
					print("Inner catch: " + str(err1))
				end
			catch (err2):
				print("Outer catch: " + str(err2))
			end
			`,
			expectError:  false,
			errorMessage: "Inner catch:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Stdout: &buf,
			}

			prog, err := ParseProgram([]byte(test.program))
			if err != nil {
				t.Fatalf("Failed to parse program: %v", err)
			}

			_, err = Execute(prog, config)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				} else if !strings.Contains(err.Error(), test.errorMessage) {
					t.Errorf("Expected error message containing '%s', but got: %v", test.errorMessage, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					output := buf.String()
					if !strings.Contains(output, test.errorMessage) {
						t.Errorf("Expected output containing '%s', but got: %s", test.errorMessage, output)
					}
				}
			}
		})
	}
}

func TestInterpreterClosures(t *testing.T) {
	program := `
	fun test():
		return 42
	end

	result = test()
	print(result)
	`

	var buf bytes.Buffer
	config := &Config{
		Stdout: &buf,
	}

	prog, err := ParseProgram([]byte(program))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	_, err = Execute(prog, config)
	if err != nil {
		t.Fatalf("Failed to execute program: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("Expected output to contain '42', but got: %s", output)
	}
}

func TestInterpreterRecursion(t *testing.T) {
	program := `
	fun factorial(n):
		if (n <= 1) then:
			return 1
		else:
			return n * factorial(n - 1)
		end
	end

	print(factorial(5))
	`

	var buf bytes.Buffer
	config := &Config{
		Stdout: &buf,
	}

	prog, err := ParseProgram([]byte(program))
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	_, err = Execute(prog, config)
	if err != nil {
		t.Fatalf("Failed to execute program: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "120") {
		t.Errorf("Expected output to contain '120', but got: %s", output)
	}
}
