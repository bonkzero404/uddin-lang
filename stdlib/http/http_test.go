package stdlib_http_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/http"
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

func TestHTTPModuleImport(t *testing.T) {
	// Just test that the module can be imported and functions exist.
	out := runScript(t, `
import "http"
print("http module loaded")
`)
	if !strings.Contains(out, "http module loaded") {
		t.Fatalf("expected output, got: %s", out)
	}
}

func TestHTTPFunctionExists(t *testing.T) {
	out := runScript(t, `
import "http"
fn = http.get
print("ok")
`)
	if !strings.Contains(out, "ok") {
		t.Fatalf("expected ok, got: %s", out)
	}
}
