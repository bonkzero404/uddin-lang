# Library Completeness Design

**Date:** 2026-05-28  
**Scope:** Tests + README fix + stdlib example (no API changes)

---

## Context

`uddin.go` exposes the `Engine` Go library API. All tests in `uddin_test.go` pass, but several methods have no test coverage and the README contains incorrect API signatures. No example exists for stdlib module usage via the library.

---

## What Is Broken / Missing

### 1. README API signature bugs (`README.md` lines ~197–214)

```go
// Current (wrong) — ExecuteString returns (*Stats, error), not (Value, error)
result, err := engine.ExecuteString(`return x + y`)

// Current (wrong) — EvaluateString returns (Value, *Stats, error), not (Value, error)
value, err := engine.EvaluateString("2 + 3 * 4")
```

Fix: update both snippets to match actual signatures.

### 2. Missing tests in `uddin_test.go`

| Test | Method under test | Gap |
|------|------------------|-----|
| `TestEngine_NewWithConfig` | `NewWithConfig()` | Not tested at all |
| `TestEngine_ExecuteFile` | `ExecuteFile()` | Not tested at all |
| `TestEngine_EnableMemoryOptimization` | `EnableMemoryOptimization()` | Not tested at all |
| `TestEngine_SetStdin` | `SetStdin()` | Not tested at all |
| `TestEngine_StdlibJSON` | `import "json"` via library | Not tested |
| `TestEngine_StdlibRegex` | `import "regex"` via library | Not tested |
| `TestEngine_TryCatch` | try-catch block via library | Not tested |
| Fix `TestStdlibModulesRegistered` | `http`, `database` modules | Only 7/9 modules covered |

### 3. Missing stdlib example (`example_usages/stdlib_usage/main.go`)

No example shows how to use `import "json"`, `import "regex"`, etc. via the Go library. All four existing examples use only core built-in functions.

---

## Design

### Section 1 — `uddin_test.go` additions

#### `TestEngine_NewWithConfig`
```go
config := DefaultConfig()
config.IsUnitTest = true
engine := NewWithConfig(config)
// assert engine != nil
// assert engine.config == config
```

#### `TestEngine_ExecuteFile`
- Write a temp `.din` file with `print("from file")`
- Call `engine.ExecuteFile(tempFile)`
- Assert output contains `"from file"`
- Defer remove temp file

#### `TestEngine_EnableMemoryOptimization`
- Call `engine.EnableMemoryOptimization()`
- Execute a simple program
- Assert no error (smoke test — verifies nothing breaks)

#### `TestEngine_SetStdin`
- Call `engine.SetStdin(strings.NewReader("hello"))`
- Assert config.Stdin is set (no execution needed — SetStdin just wires the reader)

#### `TestEngine_StdlibJSON`
```go
src := `import "json"
result = json.parse("{\"x\": 42}")
print(result["x"])`
// assert output contains "42"

src2 := `import "json"
print(json.stringify({"a": 1}))`
// assert output contains `"a"`
```

#### `TestEngine_StdlibRegex`
```go
src := `import "regex"
print(regex.is_match("^\\d+$", "12345"))`
// assert output contains "true"
```

#### `TestEngine_TryCatch`
```go
src := `
try:
    x = 1 / 0
catch (err):
    print("caught: " + str(err))
end`
// assert output contains "caught"
```

#### Fix `TestStdlibModulesRegistered`
Add two more cases:
```go
{"http",     `import "http"\nprint(typeof(http.get))`,      "function"},
{"database", `import "database"\nprint(typeof(database.connect))`, "function"},
```
These test importability without making real network/DB calls.

### Section 2 — README fix

In the "Using as Go Library" code block, replace:

```go
// Before
result, err := engine.ExecuteString(`
    x = 10
    y = 20
    return x + y
`)
fmt.Println("Result:", result)

value, err := engine.EvaluateString("2 + 3 * 4")
fmt.Println("Expression result:", value)
```

With:

```go
// After
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
fmt.Println("Output:", buf.String()) // Output: 30
fmt.Println("Ops:", stats.Ops)

value, stats, err := engine.EvaluateString("2 + 3 * 4")
if err != nil {
    panic(err)
}
fmt.Println("Expression result:", value) // Output: Expression result: 14
```

### Section 3 — `example_usages/stdlib_usage/main.go`

Single `main.go` demonstrating all 9 stdlib modules via library:

```
func main():
  // JSON
  import "json" → json.parse / json.stringify
  
  // Regex
  import "regex" → regex.is_match, regex.find_all
  
  // Datetime
  import "datetime" → datetime.now, datetime.format
  
  // Filesystem
  import "fs" → fs.getcwd, fs.file_exists
  
  // HTTP (show get/post, expect network may fail — wrapped in try-catch)
  import "http" → http.get with error handling
  
  // Database (show connect setup, wrapped in try-catch for no-DB env)
  import "database" → database.connect with error handling
  
  // Fact
  import "fact" → fact.assert, fact.query, fact.exists
  
  // CDC
  import "cdc" → cdc.emit, cdc.count
  
  // WAF
  import "waf" → waf.cidr_match, waf.path_match
```

Each section: brief `fmt.Println` header + engine execution + output printed to stdout.  
Network/DB sections wrapped in try-catch inside the `.din` script so the example runs without external services.

---

## Files Changed

| File | Change |
|------|--------|
| `uddin_test.go` | Add 8 tests, fix 1 test (http+database in stdlib check) |
| `README.md` | Fix 2 API signature examples in "Using as Go Library" section |
| `example_usages/stdlib_usage/main.go` | New file — stdlib demo |

## Out of Scope

- No new API methods (`GetVariable`, `Reset`, etc.)
- No restructuring of existing examples
- No changes to `interpreter/` package
