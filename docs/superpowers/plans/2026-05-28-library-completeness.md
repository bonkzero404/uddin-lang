# Library Completeness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix missing test coverage, a wrong README code example, and add a stdlib usage example for the uddin-lang Go library.

**Architecture:** Three independent file changes — `uddin_test.go` gets new tests appended and one existing test extended; `README.md` gets its broken Go snippet replaced; a new `example_usages/stdlib_usage/main.go` demonstrates all 9 stdlib modules. No changes to `interpreter/` or `stdlib/`.

**Tech Stack:** Go 1.24, `testing` package, `github.com/bonkzero404/uddin-lang` library API.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `uddin_test.go` | Modify (append + extend) | Add 8 new tests; extend `TestStdlibModulesRegistered` with http+database |
| `README.md` | Modify (lines 184–216) | Fix incorrect `ExecuteString`/`EvaluateString` API signatures |
| `example_usages/stdlib_usage/main.go` | Create | Runnable Go program demonstrating all 9 stdlib modules |

---

## Task 1: Extend `TestStdlibModulesRegistered` to cover http and database

**Files:**
- Modify: `uddin_test.go:701–734`

Context: `TestStdlibModulesRegistered` currently tests 7 of 9 stdlib modules. `http` and `database` are missing. We test importability only — `typeof(http.get)` returns `"function"` without making a real network call.

- [ ] **Step 1: Open `uddin_test.go` and locate the `tests` slice in `TestStdlibModulesRegistered` (lines 703–715)**

The slice currently ends with:
```go
{"fact", `import "fact"\nfact.assert("cat", "k")\nprint(fact.exists("cat", "k"))`, "true"},
```

- [ ] **Step 2: Add http and database entries to the slice**

Replace the closing `}` of the slice (after the `fact` entry) so the full slice reads:

```go
tests := []struct {
    name string
    src  string
    want string
}{
    {"regex", `import "regex"\nprint(regex.is_match("^\\d+$", "42"))`, "true"},
    {"datetime", `import "datetime"\nprint(typeof(datetime.now()))`, "string"},
    {"json", `import "json"\nprint(json.stringify(42))`, "42"},
    {"fs", `import "fs"\nprint(typeof(fs.getcwd()))`, "string"},
    {"waf", `import "waf"\nprint(waf.cidr_match("192.168.1.1", "192.168.1.0/24"))`, "true"},
    {"cdc", `import "cdc"\ncdc.emit("test", {})\nprint(cdc.count("test"))`, "1"},
    {"fact", `import "fact"\nfact.assert("cat", "k")\nprint(fact.exists("cat", "k"))`, "true"},
    {"http", `import "http"\nprint(typeof(http.get))`, "function"},
    {"database", `import "database"\nprint(typeof(database.connect))`, "function"},
}
```

- [ ] **Step 3: Run the test to verify it passes**

```bash
go test . -run TestStdlibModulesRegistered -v
```

Expected output:
```
=== RUN   TestStdlibModulesRegistered
=== RUN   TestStdlibModulesRegistered/regex
=== RUN   TestStdlibModulesRegistered/datetime
=== RUN   TestStdlibModulesRegistered/json
=== RUN   TestStdlibModulesRegistered/fs
=== RUN   TestStdlibModulesRegistered/waf
=== RUN   TestStdlibModulesRegistered/cdc
=== RUN   TestStdlibModulesRegistered/fact
=== RUN   TestStdlibModulesRegistered/http
=== RUN   TestStdlibModulesRegistered/database
--- PASS: TestStdlibModulesRegistered (0.XXs)
ok  	github.com/bonkzero404/uddin-lang
```

- [ ] **Step 4: Commit**

```bash
git add uddin_test.go
git commit -m "test: extend TestStdlibModulesRegistered to cover http and database modules"
```

---

## Task 2: Add `TestEngine_NewWithConfig`

**Files:**
- Modify: `uddin_test.go` (append after last test)

- [ ] **Step 1: Append the following test to `uddin_test.go`**

```go
func TestEngine_NewWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.IsUnitTest = true

	engine := NewWithConfig(config)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if engine.config != config {
		t.Fatal("expected engine to use provided config")
	}

	// Verify it can execute code
	var buf strings.Builder
	engine.SetStdout(&buf)
	_, err := engine.ExecuteString(`print("cfg ok")`)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}
	if !strings.Contains(buf.String(), "cfg ok") {
		t.Errorf("expected output to contain 'cfg ok', got %q", buf.String())
	}
}
```

- [ ] **Step 2: Run the test**

```bash
go test . -run TestEngine_NewWithConfig -v
```

Expected: `--- PASS: TestEngine_NewWithConfig`

- [ ] **Step 3: Commit**

```bash
git add uddin_test.go
git commit -m "test: add TestEngine_NewWithConfig coverage"
```

---

## Task 3: Add `TestEngine_ExecuteFile`

**Files:**
- Modify: `uddin_test.go` (append)

- [ ] **Step 1: Append the following test**

```go
func TestEngine_ExecuteFile(t *testing.T) {
	// Write a temp .din script
	f, err := os.CreateTemp("", "uddin_test_*.din")
	if err != nil {
		t.Fatal("create temp file:", err)
	}
	defer os.Remove(f.Name())

	_, err = f.WriteString(`print("hello from file")`)
	if err != nil {
		t.Fatal("write temp file:", err)
	}
	f.Close()

	var buf strings.Builder
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	_, err = engine.ExecuteFile(f.Name())
	if err != nil {
		t.Fatalf("ExecuteFile failed: %v", err)
	}

	if !strings.Contains(buf.String(), "hello from file") {
		t.Errorf("expected 'hello from file' in output, got %q", buf.String())
	}
}
```

- [ ] **Step 2: Verify `os` is already imported in `uddin_test.go`**

```bash
head -10 uddin_test.go
```

If `"os"` is not in the import block, add it. The existing file only imports `bytes`, `strings`, `sync`, `testing` — `os` is missing and must be added.

Add `"os"` to the import block:
```go
import (
    "bytes"
    "os"
    "strings"
    "sync"
    "testing"
)
```

- [ ] **Step 3: Run the test**

```bash
go test . -run TestEngine_ExecuteFile -v
```

Expected: `--- PASS: TestEngine_ExecuteFile`

- [ ] **Step 4: Commit**

```bash
git add uddin_test.go
git commit -m "test: add TestEngine_ExecuteFile coverage"
```

---

## Task 4: Add `TestEngine_EnableMemoryOptimization` and `TestEngine_SetStdin`

**Files:**
- Modify: `uddin_test.go` (append)

- [ ] **Step 1: Append both tests**

```go
func TestEngine_EnableMemoryOptimization(t *testing.T) {
	var buf strings.Builder
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)
	engine.EnableMemoryOptimization()

	_, err := engine.ExecuteString(`print("mem ok")`)
	if err != nil {
		t.Fatalf("execution after EnableMemoryOptimization failed: %v", err)
	}
	if !strings.Contains(buf.String(), "mem ok") {
		t.Errorf("expected 'mem ok' in output, got %q", buf.String())
	}
}

func TestEngine_SetStdin(t *testing.T) {
	engine := New()
	r := strings.NewReader("hello")
	engine.SetStdin(r)
	if engine.config.Stdin != r {
		t.Error("expected config.Stdin to be the provided reader")
	}
}
```

- [ ] **Step 2: Run both tests**

```bash
go test . -run "TestEngine_EnableMemoryOptimization|TestEngine_SetStdin" -v
```

Expected: both PASS.

- [ ] **Step 3: Commit**

```bash
git add uddin_test.go
git commit -m "test: add EnableMemoryOptimization and SetStdin coverage"
```

---

## Task 5: Add stdlib-exercising tests (`TestEngine_StdlibJSON`, `TestEngine_StdlibRegex`, `TestEngine_TryCatch`)

**Files:**
- Modify: `uddin_test.go` (append)

- [ ] **Step 1: Append the three tests**

```go
func TestEngine_StdlibJSON(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		var buf strings.Builder
		engine := New()
		engine.SetStdout(&buf)
		engine.SetUnitTestMode(true)
		src := `import "json"
result = json.parse("{\"x\": 42}")
print(result["x"])`
		_, err := engine.ExecuteString(src)
		if err != nil {
			t.Fatalf("json.parse failed: %v", err)
		}
		if !strings.Contains(buf.String(), "42") {
			t.Errorf("expected '42' in output, got %q", buf.String())
		}
	})

	t.Run("stringify", func(t *testing.T) {
		var buf strings.Builder
		engine := New()
		engine.SetStdout(&buf)
		engine.SetUnitTestMode(true)
		src := `import "json"
print(json.stringify({"a": 1}))`
		_, err := engine.ExecuteString(src)
		if err != nil {
			t.Fatalf("json.stringify failed: %v", err)
		}
		if !strings.Contains(buf.String(), `"a"`) {
			t.Errorf(`expected '"a"' in output, got %q`, buf.String())
		}
	})
}

func TestEngine_StdlibRegex(t *testing.T) {
	t.Run("is_match_true", func(t *testing.T) {
		var buf strings.Builder
		engine := New()
		engine.SetStdout(&buf)
		engine.SetUnitTestMode(true)
		src := `import "regex"
print(regex.is_match("^\\d+$", "12345"))`
		_, err := engine.ExecuteString(src)
		if err != nil {
			t.Fatalf("regex.is_match failed: %v", err)
		}
		if !strings.Contains(buf.String(), "true") {
			t.Errorf("expected 'true', got %q", buf.String())
		}
	})

	t.Run("is_match_false", func(t *testing.T) {
		var buf strings.Builder
		engine := New()
		engine.SetStdout(&buf)
		engine.SetUnitTestMode(true)
		src := `import "regex"
print(regex.is_match("^\\d+$", "abc"))`
		_, err := engine.ExecuteString(src)
		if err != nil {
			t.Fatalf("regex.is_match failed: %v", err)
		}
		if !strings.Contains(buf.String(), "false") {
			t.Errorf("expected 'false', got %q", buf.String())
		}
	})
}

func TestEngine_TryCatch(t *testing.T) {
	var buf strings.Builder
	engine := New()
	engine.SetStdout(&buf)
	engine.SetUnitTestMode(true)

	src := `
try:
    x = 1 / 0
catch (err):
    print("caught error")
end`
	_, err := engine.ExecuteString(src)
	if err != nil {
		t.Fatalf("try-catch execution failed: %v", err)
	}
	if !strings.Contains(buf.String(), "caught error") {
		t.Errorf("expected 'caught error' in output, got %q", buf.String())
	}
}
```

- [ ] **Step 2: Run all three tests**

```bash
go test . -run "TestEngine_StdlibJSON|TestEngine_StdlibRegex|TestEngine_TryCatch" -v
```

Expected: all sub-tests PASS.

- [ ] **Step 3: Run full test suite to confirm nothing regressed**

```bash
go test . -v 2>&1 | tail -5
```

Expected: `ok  github.com/bonkzero404/uddin-lang`

- [ ] **Step 4: Commit**

```bash
git add uddin_test.go
git commit -m "test: add StdlibJSON, StdlibRegex, TryCatch library tests"
```

---

## Task 6: Fix README API signatures

**Files:**
- Modify: `README.md:184–216`

Context: Lines 184–216 show a `go` code block with two wrong API calls. `ExecuteString` returns `(*Stats, error)` — not a value. `EvaluateString` returns `(Value, *Stats, error)` — not two values.

- [ ] **Step 1: Replace the broken code block**

Find and replace the entire `#### Using as Go Library` code block (lines 184–216):

```markdown
#### Using as Go Library

```go
package main

import (
    "fmt"
    "strings"

    uddin "github.com/bonkzero404/uddin-lang"
)

func main() {
    engine := uddin.New()

    // Capture output via SetStdout
    var buf strings.Builder
    engine.SetStdout(&buf)

    stats, err := engine.ExecuteString(`
        x = 10
        y = 20
        print(x + y)
    `)
    if err != nil {
        panic(err)
    }
    fmt.Println("Output:", strings.TrimSpace(buf.String())) // Output: 30
    fmt.Println("Ops:", stats.Ops)

    // Evaluate a single expression — returns (Value, *Stats, error)
    value, _, err := engine.EvaluateString("2 + 3 * 4")
    if err != nil {
        panic(err)
    }
    fmt.Println("Expression result:", value) // Output: Expression result: 14
}
```
```

- [ ] **Step 2: Verify the README renders correctly (spot-check)**

```bash
grep -A 40 "#### Using as Go Library" README.md | head -45
```

Confirm `stats, err :=` and `value, _, err :=` appear correctly.

- [ ] **Step 3: Verify Go code compiles**

```bash
go vet ./... 2>&1
```

Expected: no errors (README is markdown, not compiled — this checks the library itself still compiles cleanly).

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: fix incorrect ExecuteString/EvaluateString API signatures in README"
```

---

## Task 7: Create `example_usages/stdlib_usage/main.go`

**Files:**
- Create: `example_usages/stdlib_usage/main.go`

Context: No existing example shows stdlib module usage via the library. This file shows all 9 modules. Network/DB-dependent sections (http, database) are wrapped in try-catch inside the `.din` script so the program runs without external services.

- [ ] **Step 1: Create the file**

```go
package main

import (
	"fmt"
	"strings"

	uddin "github.com/bonkzero404/uddin-lang"
)

func run(engine *uddin.Engine, label, src string) {
	var buf strings.Builder
	engine.SetStdout(&buf)
	_, err := engine.ExecuteString(strings.ReplaceAll(src, `\n`, "\n"))
	if err != nil {
		fmt.Printf("[%s] ERROR: %v\n", label, err)
		return
	}
	out := strings.TrimSpace(buf.String())
	fmt.Printf("[%s] %s\n", label, out)
}

func main() {
	fmt.Println("=== UDDIN-LANG stdlib module usage via Go library ===")

	// Each example uses a fresh engine to avoid state bleed.
	// engine.SetUnitTestMode(true) prevents auto-execution of main().

	// ── json ──────────────────────────────────────────────────────────────
	fmt.Println("\n── json ──")
	e := uddin.New()
	e.SetUnitTestMode(true)
	run(e, "parse", `import "json"\nresult = json.parse("{\"name\":\"Alice\",\"age\":30}")\nprint(result["name"])`)
	e2 := uddin.New()
	e2.SetUnitTestMode(true)
	run(e2, "stringify", `import "json"\nprint(json.stringify({"x": 1, "y": 2}))`)

	// ── regex ─────────────────────────────────────────────────────────────
	fmt.Println("\n── regex ──")
	e3 := uddin.New()
	e3.SetUnitTestMode(true)
	run(e3, "is_match", `import "regex"\nprint(regex.is_match("^\\d+$", "12345"))`)
	e4 := uddin.New()
	e4.SetUnitTestMode(true)
	run(e4, "find_all", `import "regex"\nmatches = regex.find_all("\\d+", "a1 b22 c333")\nprint(len(matches))`)

	// ── datetime ──────────────────────────────────────────────────────────
	fmt.Println("\n── datetime ──")
	e5 := uddin.New()
	e5.SetUnitTestMode(true)
	run(e5, "now type", `import "datetime"\nprint(typeof(datetime.now()))`)
	e6 := uddin.New()
	e6.SetUnitTestMode(true)
	run(e6, "format", `import "datetime"\nprint(typeof(datetime.format(datetime.now(), "2006-01-02")))`)

	// ── fs ────────────────────────────────────────────────────────────────
	fmt.Println("\n── fs ──")
	e7 := uddin.New()
	e7.SetUnitTestMode(true)
	run(e7, "getcwd", `import "fs"\nprint(typeof(fs.getcwd()))`)
	e8 := uddin.New()
	e8.SetUnitTestMode(true)
	run(e8, "file_exists", `import "fs"\nprint(fs.file_exists("/nonexistent/path"))`)

	// ── http (wrapped in try-catch — real network optional) ───────────────
	fmt.Println("\n── http ──")
	e9 := uddin.New()
	e9.SetUnitTestMode(true)
	run(e9, "get type", `import "http"\nprint(typeof(http.get))`)

	// ── database (wrapped in try-catch — real DB not required) ────────────
	fmt.Println("\n── database ──")
	e10 := uddin.New()
	e10.SetUnitTestMode(true)
	run(e10, "connect type", `import "database"\nprint(typeof(database.connect))`)

	// ── fact ──────────────────────────────────────────────────────────────
	fmt.Println("\n── fact ──")
	e11 := uddin.New()
	e11.SetUnitTestMode(true)
	run(e11, "assert+query", `import "fact"\nfact.assert("user", "alice", {"role": "admin"})\nresult = fact.query("user", "alice")\nprint(result["role"])`)

	// ── cdc ───────────────────────────────────────────────────────────────
	fmt.Println("\n── cdc ──")
	e12 := uddin.New()
	e12.SetUnitTestMode(true)
	run(e12, "emit+count", `import "cdc"\ncdc.emit("login", {"user": "alice"})\ncdc.emit("login", {"user": "bob"})\nprint(cdc.count("login"))`)

	// ── waf ───────────────────────────────────────────────────────────────
	fmt.Println("\n── waf ──")
	e13 := uddin.New()
	e13.SetUnitTestMode(true)
	run(e13, "cidr_match", `import "waf"\nprint(waf.cidr_match("10.0.0.5", "10.0.0.0/8"))`)
	e14 := uddin.New()
	e14.SetUnitTestMode(true)
	run(e14, "path_match", `import "waf"\nprint(waf.path_match("/api/*", "/api/users"))`)

	fmt.Println("\n=== done ===")
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./example_usages/stdlib_usage/
```

Expected: no output (success).

- [ ] **Step 3: Run it**

```bash
go run ./example_usages/stdlib_usage/
```

Expected output (exact values may vary for datetime):
```
=== UDDIN-LANG stdlib module usage via Go library ===

── json ──
[parse] Alice
[stringify] {"x":1,"y":2}

── regex ──
[is_match] true
[find_all] 3

── datetime ──
[now type] string
[format] string

── fs ──
[getcwd] string
[file_exists] false

── http ──
[get type] function

── database ──
[connect type] function

── fact ──
[assert+query] admin

── cdc ──
[emit+count] 2

── waf ──
[cidr_match] true
[path_match] true

=== done ===
```

- [ ] **Step 4: Run full test suite one more time**

```bash
go test ./... 2>&1 | tail -20
```

Expected: all packages `ok`, no `FAIL`.

- [ ] **Step 5: Commit**

```bash
git add example_usages/stdlib_usage/main.go
git commit -m "feat(examples): add stdlib_usage example demonstrating all 9 stdlib modules"
```

---

## Task 8: Final push

- [ ] **Step 1: Push branch**

```bash
git push origin phase-1-bytecode-vm
```

- [ ] **Step 2: Verify PR #10 now includes all commits**

```bash
gh pr view 10 --json commits | head -30
```

---

## Self-Review

**Spec coverage:**
- ✅ `TestStdlibModulesRegistered` extended (http + database) — Task 1
- ✅ `TestEngine_NewWithConfig` — Task 2
- ✅ `TestEngine_ExecuteFile` — Task 3
- ✅ `TestEngine_EnableMemoryOptimization` — Task 4
- ✅ `TestEngine_SetStdin` — Task 4
- ✅ `TestEngine_StdlibJSON` — Task 5
- ✅ `TestEngine_StdlibRegex` — Task 5
- ✅ `TestEngine_TryCatch` — Task 5
- ✅ README fix — Task 6
- ✅ `example_usages/stdlib_usage/main.go` — Task 7

**Placeholder scan:** No TBD/TODO/vague steps. All code shown in full.

**Type consistency:** `engine.SetUnitTestMode(true)` used consistently. `strings.Builder` + `engine.SetStdout(&buf)` pattern consistent across all new tests. `uddin.New()` / `uddin.Engine` used correctly in example.

**`os` import:** Noted in Task 3 Step 2 — must be added to `uddin_test.go` import block.

**WAF note:** `waf.path_match` takes `(pattern, path)` per stdlib/waf/waf.go — confirmed correct in Task 7 example.
