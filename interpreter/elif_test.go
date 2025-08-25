package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

func TestElifFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		program  string
		expected string
	}{
		{
			name: "basic elif",
			program: `
				x = 5
				if (x > 10) then:
					print("greater than 10")
				elif (x > 0) then:
					print("greater than 0")
				else:
					print("zero or negative")
				end
			`,
			expected: "greater than 0",
		},
		{
			name: "multiple elif",
			program: `
				x = 15
				if (x < 10) then:
					print("less than 10")
				elif (x < 20) then:
					print("between 10 and 20")
				elif (x < 30) then:
					print("between 20 and 30")
				else:
					print("30 or greater")
				end
			`,
			expected: "between 10 and 20",
		},
		{
			name: "elif without else",
			program: `
				x = 0
				if (x > 0) then:
					print("positive")
				elif (x < 0) then:
					print("negative")
				end
				print("done")
			`,
			expected: "done",
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
			output = strings.ReplaceAll(output, "\r\n", "\n")
			expected := strings.ReplaceAll(test.expected, "\r\n", "\n")

			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain '%s', but got:\n%s", expected, output)
			}
		})
	}
}

func TestElifTokenizer(t *testing.T) {
	input := "elif (x > 0) then: print(x) end"
	tokenizer := NewTokenizer([]byte(input))
	
	expectedTokens := []Token{ELIF, LPAREN, NAME, GT, INT, RPAREN, THEN, COLON, NAME, LPAREN, NAME, RPAREN, END, EOF}
	
	for i, expected := range expectedTokens {
		_, tok, _ := tokenizer.Next()
		if tok != expected {
			t.Errorf("Token %d: expected %s, got %s", i, expected, tok)
		}
	}
}