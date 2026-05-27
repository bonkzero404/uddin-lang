package stdlib_waf_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/waf"
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

func TestWAFCIDRMatch(t *testing.T) {
	out := runScript(t, `
import "waf"
result = waf.cidr_match("192.168.1.1", "192.168.1.0/24")
print(result)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got: %s", out)
	}
}

func TestWAFCIDRMatchFalse(t *testing.T) {
	out := runScript(t, `
import "waf"
result = waf.cidr_match("10.0.0.1", "192.168.1.0/24")
print(result)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got: %s", out)
	}
}

func TestWAFModuleImport(t *testing.T) {
	out := runScript(t, `
import "waf"
print("waf loaded")
`)
	if !strings.Contains(out, "waf loaded") {
		t.Fatalf("expected waf loaded, got: %s", out)
	}
}
