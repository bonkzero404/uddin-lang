package stdlib_database_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/database"
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

func TestDatabaseModuleImport(t *testing.T) {
	out := runScript(t, `
import "database"
print("database module loaded")
`)
	if !strings.Contains(out, "database module loaded") {
		t.Fatalf("expected output, got: %s", out)
	}
}

func TestDatabaseFunctionExists(t *testing.T) {
	out := runScript(t, `
import "database"
fn = database.connect
print("ok")
`)
	if !strings.Contains(out, "ok") {
		t.Fatalf("expected ok, got: %s", out)
	}
}
