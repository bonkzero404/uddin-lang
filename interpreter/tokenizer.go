package interpreter

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Token represents the type of lexical token in the language
// Each token corresponds to a specific language element like operators,
// keywords, identifiers, etc.
type Token int

const (
	// Stop tokens
	ILLEGAL Token = iota
	EOF

	// Single-character tokens
	ASSIGN
	COLON
	COMMA
	DIVIDE
	DOT
	GT
	LBRACE
	LBRACKET
	LPAREN
	LT
	MINUS
	MODULO
	PLUS
	RBRACE
	RBRACKET
	RPAREN
	TIMES
	QUESTION

	// Alternative block tokens
	END

	// Two-character tokens
	EQUAL
	GTE
	LTE
	NOTEQUAL
	PLUSEQUAL
	MINUSEQUAL
	TIMESEQUAL
	DIVIDEEQUAL
	MODULOEQUAL

	// Three-character tokens
	ELLIPSIS

	// Keywords
	AND
	BREAK
	CATCH
	CONTINUE
	ELSE
	FALSE
	FOR
	FUN
	IF
	IMPORT
	IN
	NULL
	NOT
	OR
	RETURN
	THEN
	TRUE
	TRY
	WHILE
	XOR

	// Literals and identifiers
	INT
	FLOAT
	NAME
	STR
)

var keywordTokens = map[string]Token{
	"and":      AND,
	"break":    BREAK,
	"catch":    CATCH,
	"continue": CONTINUE,
	"else":     ELSE,
	"end":      END,
	"false":    FALSE,
	"for":      FOR,
	"fun":      FUN,
	"if":       IF,
	"import":   IMPORT,
	"in":       IN,
	"null":     NULL,
	"not":      NOT,
	"or":       OR,
	"return":   RETURN,
	"then":     THEN,
	"true":     TRUE,
	"try":      TRY,
	"while":    WHILE,
	"xor":      XOR,
}

var tokenNames = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	ASSIGN:   "=",
	COLON:    ":",
	COMMA:    ",",
	DIVIDE:   "/",
	DOT:      ".",
	GT:       ">",
	LBRACE:   "{",
	LBRACKET: "[",
	LPAREN:   "(",
	LT:       "<",
	MINUS:    "-",
	MODULO:   "%",
	PLUS:     "+",
	RBRACE:   "}",
	RBRACKET: "]",
	RPAREN:   ")",
	TIMES:    "*",
	QUESTION: "?",

	END: "end",

	EQUAL:       "==",
	GTE:         ">=",
	LTE:         "<=",
	NOTEQUAL:    "!=",
	PLUSEQUAL:   "+=",
	MINUSEQUAL:  "-=",
	TIMESEQUAL:  "*=",
	DIVIDEEQUAL: "/=",
	MODULOEQUAL: "%=",

	ELLIPSIS: "...",

	AND:      "and",
	BREAK:    "break",
	CATCH:    "catch",
	CONTINUE: "continue",
	ELSE:     "else",
	FALSE:    "false",
	FOR:      "for",
	FUN:      "fun",
	IF:       "if",
	IMPORT:   "import",
	IN:       "in",
	NULL:     "null",
	NOT:      "not",
	OR:       "or",
	RETURN:   "return",
	THEN:     "then",
	TRUE:     "true",
	TRY:      "try",
	WHILE:    "while",
	XOR:      "xor",

	INT:   "int",
	FLOAT: "float",
	NAME:  "name",
	STR:   "str",
}

func (t Token) String() string {
	return tokenNames[t]
}

// Position stores the line and column where a token starts in the source code
// This is used for error reporting and debugging purposes
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
}

// Tokenizer parses input source code to a stream of tokens.
// Use NewTokenizer() to create a tokenizer instance, and Next() to get the next
// token in the input stream.
type Tokenizer struct {
	input    []byte   // The source code as a byte array
	offset   int      // Current offset in the input
	ch       rune     // Current character being processed
	errorMsg string   // Error message if an error occurred
	pos      Position // Current position in the source
	nextPos  Position // Next position in the source
}

// NewTokenizer creates and initializes a new tokenizer for the given input.
// It sets up the initial position and reads the first character.
// Parameters:
//   - input: The source code as a byte array
//
// Returns a pointer to the initialized Tokenizer
func NewTokenizer(input []byte) *Tokenizer {
	t := new(Tokenizer)
	t.input = input
	t.nextPos.Line = 1   // Start at line 1
	t.nextPos.Column = 1 // Start at column 1
	t.next()             // Read the first character
	return t
}

// next reads the next rune (Unicode character) from the input and updates the position.
// It handles UTF-8 decoding, end-of-file detection, and position tracking.
func (t *Tokenizer) next() {
	// Save current position before moving to next character
	t.pos = t.nextPos

	// Decode the next UTF-8 character
	ch, size := utf8.DecodeRune(t.input[t.offset:])

	// Handle end of input
	if size == 0 {
		t.ch = -1 // Use -1 to indicate EOF
		return
	}

	// Handle invalid UTF-8 sequences
	if ch == utf8.RuneError {
		t.ch = -1
		t.errorMsg = fmt.Sprintf("invalid UTF-8 byte 0x%02x", t.input[t.offset])
		return
	}

	// Update position tracking
	if ch == '\n' {
		t.nextPos.Line++     // Move to next line
		t.nextPos.Column = 1 // Reset column to beginning of line
	} else {
		t.nextPos.Column++ // Move to next column
	}

	// Store the character and advance the offset
	t.ch = ch
	t.offset += size
}

// skipWhitespaceAndComments skips over any whitespace characters and comments in the input.
// This ensures that tokens returned by Next() are meaningful language elements,
// ignoring spaces, tabs, newlines, and comments.
func (t *Tokenizer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace characters (space, tab, carriage return, newline)
		for t.ch == ' ' || t.ch == '\t' || t.ch == '\r' || t.ch == '\n' {
			t.next()
		}

		// Check for comments starting with '/'
		if t.ch == '/' && t.offset < len(t.input) {
			nextChar := rune(t.input[t.offset])

			// Single-line comment (//)
			if nextChar == '/' {
				t.next() // Skip first '/'
				t.next() // Skip second '/'

				// Skip everything until end of line or end of input
				for t.ch != '\n' && t.ch >= 0 {
					t.next()
				}
				t.next() // Skip the newline character
				continue
			}

			// Multi-line comment (/* */)
			if nextChar == '*' {
				t.next() // Skip '/'
				t.next() // Skip '*'

				// Skip everything until we find */
				for {
					if t.ch < 0 {
						// Reached end of input without closing comment
						t.errorMsg = "unterminated multiline comment"
						return
					}

					if t.ch == '*' && t.offset < len(t.input) && rune(t.input[t.offset]) == '/' {
						t.next() // Skip '*'
						t.next() // Skip '/'
						break
					}

					t.next()
				}
				continue
			}
		}

		// No more comments or whitespace to skip
		break
	}
}

// isNameStart determines if the given rune can be the start of a name (identifier)
// Valid name starts are underscore or any alphabetic character (a-z, A-Z)
// Parameters:
//   - ch: The character to check
//
// Returns true if the character can start a name, false otherwise
func isNameStart(ch rune) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// Language constants and version information
const (
	// Version information
	LanguageName    = "Uddin-Lang"
	LanguageVersion = "1.0.0"

	// Default values
	DefaultFileExtension = ".din"
	DefaultExamplesDir   = "./examples"

	// Error message constants
	ErrorTypePrefix    = "Type Error"
	ErrorValuePrefix   = "Value Error"
	ErrorNamePrefix    = "Name Error"
	ErrorRuntimePrefix = "Runtime Error"
	ErrorSyntaxPrefix  = "Syntax Error"
)

// GetVersionInfo returns version information about the interpreter
func GetVersionInfo() string {
	return fmt.Sprintf("%s v%s", LanguageName, LanguageVersion)
}

// IsValidFileExtension checks if a filename has the correct extension
func IsValidFileExtension(filename string) bool {
	return strings.HasSuffix(filename, DefaultFileExtension)
}

// Next returns the position, token type, and token value of the next token in the source.
// For ordinary tokens (like operators and keywords), the token value is empty.
// For INT, FLOAT, NAME, and STR tokens, it contains the actual value.
// For an ILLEGAL token, it contains the error message.
//
// Returns:
//   - Position: The position (line and column) where the token starts
//   - Token: The type of the token
//   - string: The token value (if applicable)
func (t *Tokenizer) Next() (Position, Token, string) {
	t.skipWhitespaceAndComments()
	// Handle end-of-file or errors
	if t.ch < 0 {
		if t.errorMsg != "" {
			return t.pos, ILLEGAL, t.errorMsg
		}
		return t.pos, EOF, ""
	}

	// Save current position and initialize token variables
	pos := t.pos
	token := ILLEGAL
	value := ""

	// Get current character and advance to next
	ch := t.ch
	t.next()

	// Process names (identifiers) and keywords
	if isNameStart(ch) {
		// Collect all characters that can be part of a name
		runes := []rune{ch}
		for isNameStart(t.ch) || (t.ch >= '0' && t.ch <= '9') {
			runes = append(runes, t.ch)
			t.next()
		}
		name := string(runes)

		// Check if it's a keyword or a regular identifier
		token, isKeyword := keywordTokens[name]
		if !isKeyword {
			token = NAME
			value = name
		}
		return pos, token, value
	}

	switch ch {
	case ':':
		token = COLON
	case ',':
		token = COMMA
	case '/':
		if t.ch == '=' {
			t.next()
			token = DIVIDEEQUAL
		} else {
			token = DIVIDE
		}
	case '{':
		token = LBRACE
	case '[':
		token = LBRACKET
	case '(':
		token = LPAREN
	case '-':
		if t.ch == '=' {
			t.next()
			token = MINUSEQUAL
		} else {
			token = MINUS
		}
	case '%':
		if t.ch == '=' {
			t.next()
			token = MODULOEQUAL
		} else {
			token = MODULO
		}
	case '+':
		if t.ch == '=' {
			t.next()
			token = PLUSEQUAL
		} else {
			token = PLUS
		}
	case '}':
		token = RBRACE
	case ']':
		token = RBRACKET
	case ')':
		token = RPAREN
	case '*':
		if t.ch == '=' {
			t.next()
			token = TIMESEQUAL
		} else {
			token = TIMES
		}
	case '?':
		token = QUESTION

	case '=':
		if t.ch == '=' {
			t.next()
			token = EQUAL
		} else {
			token = ASSIGN
		}
	case '!':
		if t.ch == '=' {
			t.next()
			token = NOTEQUAL
		} else {
			token = ILLEGAL
			value = fmt.Sprintf("expected != instead of !%c", t.ch)
		}
	case '<':
		if t.ch == '=' {
			t.next()
			token = LTE
		} else {
			token = LT
		}
	case '>':
		if t.ch == '=' {
			t.next()
			token = GTE
		} else {
			token = GT
		}

	case '.':
		if t.ch == '.' {
			t.next()
			if t.ch != '.' {
				return pos, ILLEGAL, "unexpected .."
			}
			t.next()
			token = ELLIPSIS
		} else {
			token = DOT
		}
	// Process numeric literals (integers and floats)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		runes := []rune{ch}
		isFloat := false

		// Collect all digits and at most one decimal point
		for t.ch >= '0' && t.ch <= '9' || t.ch == '.' {
			if t.ch == '.' {
				if isFloat {
					return pos, ILLEGAL, "unexpected second '.' in number"
				}
				isFloat = true
			}
			runes = append(runes, t.ch)
			t.next()
		}

		// Determine if it's an integer or float based on presence of decimal point
		if isFloat {
			token = FLOAT
		} else {
			token = INT
		}
		value = string(runes)

	// Process string literals enclosed in double quotes
	case '"':
		runes := []rune{}

		// Collect all characters until closing quote
		for t.ch != '"' {
			c := t.ch

			// Check for unterminated string
			if c < 0 {
				return pos, ILLEGAL, "didn't find end quote in string"
			}

			// Strings cannot contain raw newlines
			if c == '\r' || c == '\n' {
				return pos, ILLEGAL, "can't have newline in string"
			}

			// Handle escape sequences
			if c == '\\' {
				t.next()
				switch t.ch {
				case '"', '\\':
					c = t.ch // Quote or backslash
				case 't':
					c = '\t' // Tab
				case 'r':
					c = '\r' // Carriage return
				case 'n':
					c = '\n' // Newline
				default:
					return pos, ILLEGAL, fmt.Sprintf("invalid string escape \\%c", t.ch)
				}
			}

			runes = append(runes, c)
			t.next()
		}

		// Skip the closing quote
		t.next()
		token = STR
		value = string(runes)

	// Process string literals enclosed in single quotes
	case '\'':
		runes := []rune{}

		// Collect all characters until closing quote
		for t.ch != '\'' {
			c := t.ch

			// Check for unterminated string
			if c < 0 {
				return pos, ILLEGAL, "didn't find end quote in string"
			}

			// Strings cannot contain raw newlines
			if c == '\r' || c == '\n' {
				return pos, ILLEGAL, "can't have newline in string"
			}

			// Handle escape sequences
			if c == '\\' {
				t.next()
				switch t.ch {
				case '\'', '\\':
					c = t.ch // Quote or backslash
				case 't':
					c = '\t' // Tab
				case 'r':
					c = '\r' // Carriage return
				case 'n':
					c = '\n' // Newline
				default:
					return pos, ILLEGAL, fmt.Sprintf("invalid string escape \\%c", t.ch)
				}
			}

			runes = append(runes, c)
			t.next()
		}

		// Skip the closing quote
		t.next()
		token = STR
		value = string(runes)

	default:
		token = ILLEGAL
		value = fmt.Sprintf("unexpected %c", ch)
	}
	return pos, token, value
}
