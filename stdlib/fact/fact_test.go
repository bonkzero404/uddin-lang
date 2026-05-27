package stdlib_fact_test

import (
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/fact"
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

func TestFactAssertAndExists(t *testing.T) {
	out := runScript(t, `
import "fact"
fact.assert("customer", "alice", {"age": 30})
exists = fact.exists("customer", "alice")
print(exists)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got: %s", out)
	}
}

func TestFactAssertSimple(t *testing.T) {
	out := runScript(t, `
import "fact"
fact.assert("flag", "feature_x")
e = fact.exists("flag", "feature_x")
print(e)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got: %s", out)
	}
}

func TestFactNotExists(t *testing.T) {
	out := runScript(t, `
import "fact"
e = fact.exists("nobody", "ghost")
print(e)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got: %s", out)
	}
}

func TestFactRetract(t *testing.T) {
	out := runScript(t, `
import "fact"
fact.assert("x", "k", 42)
fact.retract("x", "k")
e = fact.exists("x", "k")
print(e)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false after retract, got: %s", out)
	}
}

func TestFactCount(t *testing.T) {
	out := runScript(t, `
import "fact"
fact.clear("item")
fact.assert("item", "a", 1)
fact.assert("item", "b", 2)
n = fact.count("item")
print(n)
`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected 2, got: %s", out)
	}
}
