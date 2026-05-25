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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.MatchString(input)
	}
}

func BenchmarkCIDRMatchHelper(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CIDRMatchHelper("185.220.101.45", "185.220.0.0/16")
	}
}

func BenchmarkPathGlobMatch(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PathGlobMatch("/api/*/users", "/api/v1/users")
	}
}

func BenchmarkWAFHeaderLookup(b *testing.B) {
	ctx := map[string]interface{}{
		"headers": map[string]string{
			"content-type":    "application/json",
			"x-forwarded-for": "185.220.101.45",
			"user-agent":      "Mozilla/5.0 (compatible; attack-bot/1.0)",
			"authorization":   "Bearer eyJhbGciOiJIUzI1NiJ9.test",
			"accept":          "text/html,application/xhtml+xml",
		},
	}
	interp := makeTestInterp(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wafHeaderFunc(interp, noPos, []Value{"x-forwarded-for"})
	}
}

func BenchmarkWAFCIDRMatch_Builtin(b *testing.B) {
	ctx := map[string]interface{}{}
	interp := makeTestInterp(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wafCIDRMatchFunc(interp, noPos, []Value{"185.220.101.45", "185.220.0.0/16"})
	}
}

func BenchmarkWAFDetected(b *testing.B) {
	ctx := map[string]interface{}{
		"detected": []string{"sql", "xss", "cmd"},
	}
	interp := makeTestInterp(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wafDetectedFunc(interp, noPos, []Value{"xss"})
	}
}
