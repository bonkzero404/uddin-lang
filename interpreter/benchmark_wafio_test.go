package interpreter

import (
	"testing"

	"github.com/coregx/coregex"
)

func BenchmarkCoregexWAFPattern(b *testing.B) {
	re, err := coregex.Compile(`(?i)(select|union|insert|drop|delete|update)\s`)
	if err != nil {
		b.Fatalf("coregex.Compile error: %v", err)
	}
	input := "GET /api/users?id=1 UNION SELECT * FROM users--"
	for b.Loop() {
		re.MatchString(input)
	}
}

func BenchmarkCIDRMatchHelper(b *testing.B) {
	for b.Loop() {
		CIDRMatchHelper("185.220.101.45", "185.220.0.0/16")
	}
}

func BenchmarkPathGlobMatch(b *testing.B) {
	for b.Loop() {
		PathGlobMatch("/api/*/users", "/api/v1/users")
	}
}

func BenchmarkWAFHeaderLookup(b *testing.B) {
	ctx := map[string]any{
		"headers": map[string]string{
			"content-type":    "application/json",
			"x-forwarded-for": "185.220.101.45",
			"user-agent":      "Mozilla/5.0 (compatible; attack-bot/1.0)",
			"authorization":   "Bearer eyJhbGciOiJIUzI1NiJ9.test",
			"accept":          "text/html,application/xhtml+xml",
		},
	}
	interp := makeTestInterp(ctx)
	for b.Loop() {
		wafHeaderFunc(interp, noPos, []Value{"x-forwarded-for"})
	}
}

func BenchmarkWAFCIDRMatch_Builtin(b *testing.B) {
	ctx := map[string]any{}
	interp := makeTestInterp(ctx)
	for b.Loop() {
		wafCIDRMatchFunc(interp, noPos, []Value{"185.220.101.45", "185.220.0.0/16"})
	}
}

func BenchmarkWAFDetected(b *testing.B) {
	ctx := map[string]any{
		"detected": []string{"sql", "xss", "cmd"},
	}
	interp := makeTestInterp(ctx)
	for b.Loop() {
		wafDetectedFunc(interp, noPos, []Value{"xss"})
	}
}
