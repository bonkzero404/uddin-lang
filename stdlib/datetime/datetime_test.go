package stdlib_datetime_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/datetime"
)

func runScript(t *testing.T, src string) string {
	t.Helper()
	var buf strings.Builder
	config := interpreter.TestConfig()
	config.Stdout = &buf
	prog, err := interpreter.ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_, err = interpreter.Execute(prog, config)
	if err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestDatetimeNow(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.now()
print(result)
`)
	// RFC3339 format contains "T" and "Z" or "+00:00"
	if !strings.Contains(out, "T") {
		t.Fatalf("expected RFC3339 date, got: %s", out)
	}
}

func TestDatetimeTimeNow(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.time_now()
print(result)
`)
	out = strings.TrimSpace(out)
	if len(out) < 10 {
		t.Fatalf("expected unix ms timestamp, got: %s", out)
	}
}

func TestDatetimeFormat(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.format("2020-01-02T15:04:05Z", "YYYY-MM-DD")
print(result)
`)
	if !strings.Contains(out, "2020-01-02") {
		t.Fatalf("expected 2020-01-02, got: %s", out)
	}
}

func TestDatetimeParse(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.parse("2023-12-25", "2006-01-02")
print(result)
`)
	if !strings.Contains(out, "2023-12-25") {
		t.Fatalf("expected 2023-12-25 in output, got: %s", out)
	}
}

func TestDatetimeDiff(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.diff("2023-12-26T10:30:00Z", "2023-12-25T10:30:00Z", "hours")
print(result)
`)
	if !strings.Contains(out, "24") {
		t.Fatalf("expected 24, got: %s", out)
	}
}

func TestDatetimeCompare(t *testing.T) {
	out := runScript(t, `
import "datetime"
result = datetime.compare("2023-12-26", "2023-12-25")
print(result)
`)
	if !strings.Contains(out, "1") {
		t.Fatalf("expected 1, got: %s", out)
	}
}
