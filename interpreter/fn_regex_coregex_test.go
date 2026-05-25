package interpreter

import (
	"testing"

	"github.com/coregx/coregex"
)

// TestCoregexWAFPatterns verifies that coregex behaves identically to standard
// regexp for common WAF detection patterns.
func TestCoregexWAFPatterns(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{
			name:    "SQLi keyword match — positive",
			pattern: `(?i)(select|union|insert)\s`,
			input:   "SELECT * FROM users",
			want:    true,
		},
		{
			name:    "SQLi keyword match — negative",
			pattern: `(?i)(select|union|insert)\s`,
			input:   "normal text",
			want:    false,
		},
		{
			name:    "XSS special chars — positive",
			pattern: `[<>'"&]`,
			input:   "<script>",
			want:    true,
		},
		{
			name:    "XSS special chars — negative",
			pattern: `[<>'"&]`,
			input:   "hello world",
			want:    false,
		},
		{
			name:    "digits anywhere — positive",
			pattern: `\d+`,
			input:   "abc123def",
			want:    true,
		},
		{
			name:    "digits only anchored — negative",
			pattern: `^\d+$`,
			input:   "abc123def",
			want:    false,
		},
		{
			name:    "UNION SELECT injection",
			pattern: `(?i)union\s+select`,
			input:   "1 UNION SELECT password FROM users",
			want:    true,
		},
		{
			name:    "path traversal",
			pattern: `\.\.\/`,
			input:   "../../etc/passwd",
			want:    true,
		},
		{
			name:    "path traversal — clean path",
			pattern: `\.\.\/`,
			input:   "/api/v1/users",
			want:    false,
		},
		{
			name:    "script tag",
			pattern: `(?i)<script`,
			input:   "<SCRIPT>alert(1)</SCRIPT>",
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			re, err := coregex.Compile(tc.pattern)
			if err != nil {
				t.Fatalf("coregex.Compile(%q) error: %v", tc.pattern, err)
			}
			got := re.MatchString(tc.input)
			if got != tc.want {
				t.Errorf("pattern=%q input=%q: got %v, want %v",
					tc.pattern, tc.input, got, tc.want)
			}
		})
	}
}
