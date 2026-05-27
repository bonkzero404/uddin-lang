package stdlib_json_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/json"
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

func TestJSONParse(t *testing.T) {
	out := runScript(t, `
import "json"
data = json.parse('{"name":"John","age":30}')
print(data["name"])
`)
	if !strings.Contains(out, "John") {
		t.Fatalf("expected John, got: %s", out)
	}
}

func TestJSONParseArray(t *testing.T) {
	out := runScript(t, `
import "json"
data = json.parse('[1, 2, 3]')
print(data[0])
`)
	if !strings.Contains(out, "1") {
		t.Fatalf("expected 1, got: %s", out)
	}
}

func TestJSONStringify(t *testing.T) {
	out := runScript(t, `
import "json"
obj = {"key": "value", "num": 42}
result = json.stringify(obj)
print(result)
`)
	if !strings.Contains(out, "key") || !strings.Contains(out, "value") {
		t.Fatalf("expected json with key/value, got: %s", out)
	}
}

func TestJSONRoundtrip(t *testing.T) {
	out := runScript(t, `
import "json"
original = '{"score":99,"name":"Alice"}'
data = json.parse(original)
result = json.stringify(data)
parsed2 = json.parse(result)
print(parsed2["name"])
`)
	if !strings.Contains(out, "Alice") {
		t.Fatalf("expected Alice, got: %s", out)
	}
}
