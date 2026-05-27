package stdlib_regex_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/regex"
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

func TestRegexIsMatch(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.is_match("^\\d+$", "12345")
print(result)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got: %s", out)
	}
}

func TestRegexMatch(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.match("hello@example.com", "^[a-z._%+-]+@[a-z.-]+\\.[a-zA-Z]{2,}$")
print(result)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got: %s", out)
	}
}

func TestRegexFind(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.find("hello world 123", "\\d+")
print(result)
`)
	if !strings.Contains(out, "123") {
		t.Fatalf("expected 123, got: %s", out)
	}
}

func TestRegexFindAll(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.find_all("hello 123 world 456", "\\d+")
print(result)
`)
	if !strings.Contains(out, "123") || !strings.Contains(out, "456") {
		t.Fatalf("expected both matches, got: %s", out)
	}
}

func TestRegexReplace(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.replace("hello 123 world", "\\d+", "XXX")
print(result)
`)
	if !strings.Contains(out, "XXX") {
		t.Fatalf("expected XXX, got: %s", out)
	}
}

func TestRegexSplit(t *testing.T) {
	out := runScript(t, `
import "regex"
result = regex.split("hello,world;test", "[,;]")
print(result)
`)
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Fatalf("expected split result, got: %s", out)
	}
}
