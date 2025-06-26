package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoopControlStatements(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "while_loop_with_break",
			program: `
				result = []
				i = 0
				while (i < 10):
					i = i + 1
					if (i == 5) then:
						break
					end
					result = result + [i]
				end
				print(result)
			`,
			expected: "[1, 2, 3, 4]",
		},
		{
			name: "while_loop_with_continue",
			program: `
				result = []
				i = 0
				while (i < 6):
					i = i + 1
					if (i % 2 == 0) then:
						continue
					end
					result = result + [i]
				end
				print(result)
			`,
			expected: "[1, 3, 5]",
		},
		{
			name: "for_loop_with_break",
			program: `
				result = []
				numbers = [1, 2, 3, 4, 5, 6, 7, 8]
				for (num in numbers):
					if (num == 5) then:
						break
					end
					result = result + [num]
				end
				print(result)
			`,
			expected: "[1, 2, 3, 4]",
		},
		{
			name: "for_loop_with_continue",
			program: `
				result = []
				numbers = [1, 2, 3, 4, 5, 6]
				for (num in numbers):
					if (num % 2 == 0) then:
						continue
					end
					result = result + [num]
				end
				print(result)
			`,
			expected: "[1, 3, 5]",
		},
		{
			name: "nested_loops_with_break",
			program: `
				result = []
				i = 1
				while (i <= 3):
					j = 1
					while (j <= 5):
						if (j == 3) then:
							break
						end
						result = result + [i * 10 + j]
						j = j + 1
					end
					i = i + 1
				end
				print(result)
			`,
			expected: "[11, 12, 21, 22, 31, 32]",
		},
		{
			name: "mixed_break_and_continue",
			program: `
				result = []
				i = 0
				while (i < 10):
					i = i + 1
					if (i % 2 == 0) then:
						continue
					end
					if (i > 7) then:
						break
					end
					result = result + [i]
				end
				print(result)
			`,
			expected: "[1, 3, 5, 7]",
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

			// Check if the expected output is contained in the actual output
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
			}
		})
	}
}

func TestLoopControlErrors(t *testing.T) {
	tests := []struct {
		name    string
		program string
		wantErr bool
	}{
		{
			name: "break_outside_loop",
			program: `
				break
			`,
			wantErr: true,
		},
		{
			name: "continue_outside_loop",
			program: `
				continue
			`,
			wantErr: true,
		},
		{
			name: "break_in_function_outside_loop",
			program: `
				fun test():
					break
				end
				test()
			`,
			wantErr: true,
		},
		{
			name: "continue_in_function_outside_loop",
			program: `
				fun test():
					continue
				end
				test()
			`,
			wantErr: true,
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
				if !test.wantErr {
					t.Fatalf("Unexpected parse error: %v", err)
				}
				return
			}

			_, err = Execute(prog, config)
			if test.wantErr && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("Unexpected execution error: %v", err)
			}
		})
	}
}
