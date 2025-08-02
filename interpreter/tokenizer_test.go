package interpreter

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
		values   []string
	}{
		{
			input:    "x = 5",
			expected: []Token{NAME, ASSIGN, INT, EOF},
			values:   []string{"x", "", "5", ""},
		},
		{
			input:    "if (x > 10) then: print(x) end",
			expected: []Token{IF, LPAREN, NAME, GT, INT, RPAREN, THEN, COLON, NAME, LPAREN, NAME, RPAREN, END, EOF},
			values:   []string{"", "", "x", "", "10", "", "", "", "print", "", "x", "", "", ""},
		},
		{
			input:    "// This is a comment\nx = 5",
			expected: []Token{NAME, ASSIGN, INT, EOF},
			values:   []string{"x", "", "5", ""},
		},
		{
			input:    "\"Hello, World!\"",
			expected: []Token{STR, EOF},
			values:   []string{"Hello, World!", ""},
		},
		{
			input:    "x == y and z != w",
			expected: []Token{NAME, EQUAL, NAME, AND, NAME, NOTEQUAL, NAME, EOF},
			values:   []string{"x", "", "y", "", "z", "", "w", ""},
		},
		{
			input:    "arr[0] = 42",
			expected: []Token{NAME, LBRACKET, INT, RBRACKET, ASSIGN, INT, EOF},
			values:   []string{"arr", "", "0", "", "", "42", ""},
		},
		{
			input:    "fun add(a, b): return a + b end",
			expected: []Token{FUN, NAME, LPAREN, NAME, COMMA, NAME, RPAREN, COLON, RETURN, NAME, PLUS, NAME, END, EOF},
			values:   []string{"", "add", "", "a", "", "b", "", "", "", "a", "", "b", "", ""},
		},
		{
			input:    "try: x = 1/0 catch (err): print(err) end",
			expected: []Token{TRY, COLON, NAME, ASSIGN, INT, DIVIDE, INT, CATCH, LPAREN, NAME, RPAREN, COLON, NAME, LPAREN, NAME, RPAREN, END, EOF},
			values:   []string{"", "", "x", "", "1", "", "0", "", "", "err", "", "", "print", "", "err", "", "", ""},
		},
		{
			input:    "power = a ** b",
			expected: []Token{NAME, ASSIGN, NAME, POWER, NAME, EOF},
			values:   []string{"power", "", "a", "", "b", ""},
		},
		{
			input:    "result = 2 ** 3 ** 2",
			expected: []Token{NAME, ASSIGN, INT, POWER, INT, POWER, INT, EOF},
			values:   []string{"result", "", "2", "", "3", "", "2", ""},
		},
		{
			input:    "/* This is a multiline comment */ x = 5",
			expected: []Token{NAME, ASSIGN, INT, EOF},
			values:   []string{"x", "", "5", ""},
		},
		{
			input:    "x = 5 /* inline comment */ + 3",
			expected: []Token{NAME, ASSIGN, INT, PLUS, INT, EOF},
			values:   []string{"x", "", "5", "", "3", ""},
		},
		{
			input:    "/* \n multiline \n comment \n */ x = 10",
			expected: []Token{NAME, ASSIGN, INT, EOF},
			values:   []string{"x", "", "10", ""},
		},
		{
			input:    "// single line\n/* multiline */ x = 1",
			expected: []Token{NAME, ASSIGN, INT, EOF},
			values:   []string{"x", "", "1", ""},
		},
	}

	for i, test := range tests {
		tokenizer := NewTokenizer([]byte(test.input))

		for j, expectedToken := range test.expected {
			pos, token, value := tokenizer.Next()

			if token != expectedToken {
				t.Errorf("Test %d, token %d: expected token %s, got %s", i, j, expectedToken, token)
			}

			if j < len(test.values) && value != test.values[j] {
				t.Errorf("Test %d, token %d: expected value '%s', got '%s'", i, j, test.values[j], value)
			}

			// Ensure position is valid
			if pos.Line < 1 || pos.Column < 1 {
				t.Errorf("Test %d, token %d: invalid position %v", i, j, pos)
			}
		}
	}
}

func TestTokenizerIllegalToken(t *testing.T) {
	// Test for illegal tokens
	tokenizer := NewTokenizer([]byte("@#$"))
	_, token, value := tokenizer.Next()

	if token != ILLEGAL {
		t.Errorf("Expected ILLEGAL token for '@#$', got %v", token)
	}

	if value == "" {
		t.Errorf("Expected error message for illegal token, got empty string")
	}
}

func TestTokenizerCommentHandling(t *testing.T) {
	input := `
	// This is a comment
	x = 5 // This is an end-of-line comment
	// Another comment
	y = 10
	`

	tokenizer := NewTokenizer([]byte(input))

	expected := []Token{NAME, ASSIGN, INT, NAME, ASSIGN, INT, EOF}
	values := []string{"x", "", "5", "y", "", "10", ""}

	for i, expectedToken := range expected {
		_, token, value := tokenizer.Next()

		if token != expectedToken {
			t.Errorf("Expected token %s, got %s", expectedToken, token)
		}

		if value != values[i] {
			t.Errorf("Expected value '%s', got '%s'", values[i], value)
		}
	}
}

func TestTokenizerStringLiterals(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{`"Simple string"`, "Simple string"},
		{`"String with \"escaped quotes\""`, `String with "escaped quotes"`},
		{`"String with \n newline"`, "String with \n newline"},
		{`"String with \t tab"`, "String with \t tab"},
		{`"String with \\ backslash"`, "String with \\ backslash"},
	}

	for i, test := range tests {
		tokenizer := NewTokenizer([]byte(test.input))
		_, token, value := tokenizer.Next()

		if token != STR {
			t.Errorf("Test %d: expected STR token, got %s", i, token)
		}

		if value != test.expect {
			t.Errorf("Test %d: expected value '%s', got '%s'", i, test.expect, value)
		}
	}
}

func TestTokenizerNumberLiterals(t *testing.T) {
	tests := []struct {
		input string
		token Token
		value string
	}{
		{"123", INT, "123"},
		{"0", INT, "0"},
		{"3.14", FLOAT, "3.14"},
		{"0.5", FLOAT, "0.5"},
		{"-42", MINUS, ""}, // This will be tokenized as MINUS followed by INT
		// Scientific notation tests
		{"1e5", FLOAT, "1e5"},
		{"2.5e-3", FLOAT, "2.5e-3"},
		{"9.461e15", FLOAT, "9.461e15"},
		{"1E10", FLOAT, "1E10"},
		{"3.14E+2", FLOAT, "3.14E+2"},
		{"6.022e-23", FLOAT, "6.022e-23"},
		// Digit separator tests
		{"1_000", INT, "1000"},
		{"1_000_000", INT, "1000000"},
		{"3_14.15_92", FLOAT, "314.1592"},
		{"1_000.5_00", FLOAT, "1000.500"},
		{"1_23e4_56", FLOAT, "123e456"},
		{"9_87.6_54e-3_21", FLOAT, "987.654e-321"},
	}

	for i, test := range tests {
		tokenizer := NewTokenizer([]byte(test.input))
		_, token, value := tokenizer.Next()

		if token != test.token {
			t.Errorf("Test %d: expected token %s, got %s", i, test.token, token)
		}

		if test.value != "" && value != test.value {
			t.Errorf("Test %d: expected value '%s', got '%s'", i, test.value, value)
		}
	}
}

// TestTokenizerDigitSeparatorErrors tests invalid digit separator usage
func TestTokenizerDigitSeparatorErrors(t *testing.T) {
	errorTests := []struct {
		input string
		expectedError string
	}{
		{"1__000", "consecutive digit separators '_' are not allowed"},
		{"1_000_", "number cannot end with digit separator '_'"},
		{"1_.5", "digit separator '_' cannot be followed by '.'"},
		{"1e2__3", "consecutive digit separators '_' are not allowed in exponent"},
		{"1e2_", "exponent cannot end with digit separator '_'"},
	}

	for i, test := range errorTests {
		tokenizer := NewTokenizer([]byte(test.input))
		_, token, value := tokenizer.Next()

		if token != ILLEGAL {
			t.Errorf("Test %d: expected ILLEGAL token, got %s", i, token)
		}

		if value != test.expectedError {
			t.Errorf("Test %d: expected error '%s', got '%s'", i, test.expectedError, value)
		}
	}
}

// TestUnterminatedMultilineComment tests that unterminated multiline comments produce errors
func TestUnterminatedMultilineComment(t *testing.T) {
	input := "x = 5 /* this comment never ends"
	tokenizer := NewTokenizer([]byte(input))

	// Skip the valid tokens first
	tokenizer.Next() // NAME
	tokenizer.Next() // ASSIGN
	tokenizer.Next() // INT

	// This should trigger the unterminated comment error
	_, token, value := tokenizer.Next()

	if token != ILLEGAL {
		t.Errorf("Expected ILLEGAL token for unterminated comment, got %s", token)
	}

	if value != "unterminated multiline comment" {
		t.Errorf("Expected error message 'unterminated multiline comment', got '%s'", value)
	}
}
