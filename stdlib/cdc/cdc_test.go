package stdlib_cdc_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/cdc"
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

func TestCDCEmitAndCount(t *testing.T) {
	out := runScript(t, `
import "cdc"
cdc.emit("user_login", {"user": "alice"})
cdc.emit("user_login", {"user": "bob"})
cdc.emit("page_view", {"page": "/home"})
total = cdc.count()
login_count = cdc.count("user_login")
print(total)
print(login_count)
`)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines of output, got: %q", out)
	}
	if !strings.Contains(lines[0], "3") {
		t.Fatalf("expected total count 3, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "2") {
		t.Fatalf("expected login count 2, got: %s", lines[1])
	}
}

func TestCDCClear(t *testing.T) {
	out := runScript(t, `
import "cdc"
cdc.emit("evt", {})
cdc.emit("evt", {})
cdc.clear()
n = cdc.count()
print(n)
`)
	if !strings.Contains(out, "0") {
		t.Fatalf("expected 0 after clear, got: %s", out)
	}
}

func TestCDCModuleImport(t *testing.T) {
	out := runScript(t, `
import "cdc"
print("cdc loaded")
`)
	if !strings.Contains(out, "cdc loaded") {
		t.Fatalf("expected cdc loaded, got: %s", out)
	}
}
