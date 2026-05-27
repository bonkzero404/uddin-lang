package interpreter

import (
	"strings"
	"testing"
)

type mockModule struct{ name string }

func (m *mockModule) Name() string { return m.name }
func (m *mockModule) Functions() map[string]ModuleFunc {
	return map[string]ModuleFunc{
		"hello": func(ctx ModuleContext, pos Position, args []Value) Value {
			return Value("hello from " + m.name)
		},
	}
}

func TestRegisterAndLookupModule(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	RegisterModule(&mockModule{"testmod"})
	mod, ok := LookupModule("testmod")
	if !ok {
		t.Fatal("module not found")
	}
	if mod.Name() != "testmod" {
		t.Fatalf("got %s", mod.Name())
	}
}

func TestLookupUnknownModule(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	_, ok := LookupModule("nonexistent")
	if ok {
		t.Fatal("should not find nonexistent module")
	}
}

func TestKnownModuleNames(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	RegisterModule(&mockModule{"beta"})
	RegisterModule(&mockModule{"alpha"})
	names := knownModuleNames()
	if names[0] != "alpha" {
		t.Fatalf("expected sorted: got %v", names)
	}
}

func TestBuildNamespaceObject(t *testing.T) {
	globalModuleRegistry = map[string]Module{}
	config := TestConfig()
	interp := newInterpreter(config)
	mod := &mockModule{"testmod"}
	ns := buildNamespaceObject(mod, interp)
	fn, ok := ns["hello"]
	if !ok {
		t.Fatal("hello not in namespace")
	}
	bf, ok := fn.(builtinFunction)
	if !ok {
		t.Fatal("not a builtinFunction")
	}
	result := bf.Function(interp, Position{}, nil)
	if result != Value("hello from testmod") {
		t.Fatalf("got %v", result)
	}
}

// Ensure strings import is used
var _ = strings.Join
