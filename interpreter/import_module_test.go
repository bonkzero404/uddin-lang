package interpreter

import (
	"strings"
	"testing"
)

func TestParseImportModule_NoAlias(t *testing.T) {
	prog, err := ParseProgram([]byte(`import "database"`))
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("got %d statements", len(prog.Statements))
	}
	imp, ok := prog.Statements[0].(*Import)
	if !ok {
		t.Fatal("not *Import")
	}
	if imp.Path != "database" {
		t.Fatalf("Path=%s", imp.Path)
	}
	if imp.Alias != "" {
		t.Fatalf("Alias=%s", imp.Alias)
	}
	if !imp.IsModule {
		t.Fatal("IsModule should be true")
	}
}

func TestParseImportModule_WithAlias(t *testing.T) {
	prog, err := ParseProgram([]byte(`import "database" as db`))
	if err != nil {
		t.Fatal(err)
	}
	imp := prog.Statements[0].(*Import)
	if imp.Path != "database" {
		t.Fatalf("Path=%s", imp.Path)
	}
	if imp.Alias != "db" {
		t.Fatalf("Alias=%s", imp.Alias)
	}
	if !imp.IsModule {
		t.Fatal("IsModule should be true")
	}
}

func TestParseImportFile_NoAlias(t *testing.T) {
	prog, err := ParseProgram([]byte(`import "utils.din"`))
	if err != nil {
		t.Fatal(err)
	}
	imp := prog.Statements[0].(*Import)
	if imp.IsModule {
		t.Fatal("IsModule should be false for .din file")
	}
	if imp.Path != "utils.din" {
		t.Fatalf("Path=%s", imp.Path)
	}
}

func TestParseImportFile_WithPath(t *testing.T) {
	prog, err := ParseProgram([]byte(`import "./libs/utils.din"`))
	if err != nil {
		t.Fatal(err)
	}
	imp := prog.Statements[0].(*Import)
	if imp.IsModule {
		t.Fatal("IsModule should be false for path with /")
	}
}

func TestImportModule_InjectsNamespace(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	RegisterModule(&mockModule{"mymod"})

	config := TestConfig()
	var buf strings.Builder
	config.Stdout = &buf

	prog, err := ParseProgram([]byte(`
import "mymod"
result = mymod.hello()
print(result)
`))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Execute(prog, config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "hello from mymod") {
		t.Fatalf("got: %s", buf.String())
	}
}

func TestImportModule_WithAlias(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	RegisterModule(&mockModule{"mymod"})

	config := TestConfig()
	var buf strings.Builder
	config.Stdout = &buf

	prog, _ := ParseProgram([]byte(`
import "mymod" as m
result = m.hello()
print(result)
`))
	_, err := Execute(prog, config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "hello from mymod") {
		t.Fatalf("got: %s", buf.String())
	}
}

func TestImportModule_UnknownModuleError(t *testing.T) {
	globalModuleRegistry = map[string]Module{}

	config := TestConfig()
	prog, _ := ParseProgram([]byte(`import "nonexistent"`))
	_, err := Execute(prog, config)
	if err == nil {
		t.Fatal("expected error for unknown module")
	}
	if !strings.Contains(err.Error(), "unknown module") {
		t.Fatalf("wrong error: %v", err)
	}
}
