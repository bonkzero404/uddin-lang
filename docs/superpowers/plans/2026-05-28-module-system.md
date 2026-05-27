# Module System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate 9 domain-specific builtin modules out of `interpreter/` into `stdlib/` packages, add `import "module"` / `import "module" as alias` syntax, and remove old flat-global function names.

**Architecture:** `ModuleContext` public interface (in `interpreter` package) wraps `*interpreter` so external `stdlib/` packages can call language-level operations without importing unexported types. Each stdlib package implements `Module` interface and registers itself via `init()`. `Engine`/`uddin.go` blank-imports all 9 stdlib packages.

**Tech Stack:** Go 1.22+, module path `github.com/bonkzero404/uddin-lang`, branch `phase-3-module-system` (cut from `main` after PR #9 merged).

---

## File Map

**New — interpreter layer:**
- `interpreter/module.go` — `ModuleContext`, `Module`, `ModuleFunc`, `CDCStore`, `FactDB` interfaces
- `interpreter/module_registry.go` — registry, `buildNamespaceObject`, `interpreterContext`

**New — stdlib packages:**
- `stdlib/regex/regex.go`
- `stdlib/datetime/datetime.go`
- `stdlib/json/json.go`
- `stdlib/fs/fs.go`
- `stdlib/http/http.go`
- `stdlib/waf/waf.go`
- `stdlib/cdc/cdc.go`
- `stdlib/fact/fact.go`
- `stdlib/database/database.go`

**Modified:**
- `interpreter/ast.go` — add `Alias string`, `IsModule bool` to `Import`
- `interpreter/tokenizer.go` — add `AS = 522`
- `interpreter/parser.go` — extend `import_()` for `as alias` + module detection
- `interpreter/interpreter.go` — add module dispatch in `executeImport`
- `interpreter/builtin_metadata.go` — remove 50+ entries for moved functions
- `interpreter/vm_builtins.go` — remove `directWafCidrMatch`, `directWafPathMatch` from `registerDirectBuiltins`
- `interpreter/compiler.go` — remove waf-specific `isRegexBuiltin` special cases if any; remove waf from direct table check
- `uddin.go` — add 9 blank imports

**Deleted after migration:**
`interpreter/fn_regex.go`, `interpreter/fn_datetime.go`, `interpreter/fn_serialization.go`,
`interpreter/fn_file_system.go`, `interpreter/fn_network.go`, `interpreter/fn_wafio.go`,
`interpreter/fn_cep.go`, `interpreter/fn_fact_database.go`, `interpreter/fn_database.go`

**New — docs:**
- `docs/docs/reference/modules.md`

---

## Task 1: Core interfaces + registry

**Files:**
- Create: `interpreter/module.go`
- Create: `interpreter/module_registry.go`
- Test: `interpreter/module_test.go`

- [ ] **Step 1: Write failing test**

```go
// interpreter/module_test.go
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
    // reset registry for test isolation
    globalModuleRegistry = map[string]Module{}
    RegisterModule(&mockModule{"testmod"})
    mod, ok := LookupModule("testmod")
    if !ok { t.Fatal("module not found") }
    if mod.Name() != "testmod" { t.Fatalf("got %s", mod.Name()) }
}

func TestLookupUnknownModule(t *testing.T) {
    globalModuleRegistry = map[string]Module{}
    _, ok := LookupModule("nonexistent")
    if ok { t.Fatal("should not find nonexistent module") }
}

func TestKnownModuleNames(t *testing.T) {
    globalModuleRegistry = map[string]Module{}
    RegisterModule(&mockModule{"beta"})
    RegisterModule(&mockModule{"alpha"})
    names := knownModuleNames()
    if names[0] != "alpha" { t.Fatalf("expected sorted: got %v", names) }
}

func TestBuildNamespaceObject(t *testing.T) {
    globalModuleRegistry = map[string]Module{}
    config := TestConfig()
    interp := newInterpreter(config)
    mod := &mockModule{"testmod"}
    ns := buildNamespaceObject(mod, interp)
    fn, ok := ns["hello"]
    if !ok { t.Fatal("hello not in namespace") }
    bf, ok := fn.(builtinFunction)
    if !ok { t.Fatal("not a builtinFunction") }
    result := bf.Function(interp, Position{}, nil)
    if result != Value("hello from testmod") { t.Fatalf("got %v", result) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./interpreter/... -run TestRegisterAndLookupModule -v
```
Expected: FAIL — `ModuleContext`, `ModuleFunc`, etc. not defined.

- [ ] **Step 3: Create `interpreter/module.go`**

```go
package interpreter

import (
    "io"
    "sync"
)

// ModuleFunc is the function signature for all stdlib module functions.
type ModuleFunc func(ctx ModuleContext, pos Position, args []Value) Value

// ModuleContext is the public API stdlib modules use to interact with
// the interpreter. Avoids exporting *interpreter.
type ModuleContext interface {
    LookupVar(name string) (Value, bool)
    SetVar(name string, val Value)
    Stdout() io.Writer
    // GetContext retrieves a host-injected map stored as "_<key>_ctx".
    // Used by waf module: GetContext("waf") → _waf_ctx map.
    GetContext(key string) (map[string]any, bool)
    // ReturnValue panics with returnResult — used by waf.return() to exit script.
    ReturnValue(val Value, pos Position)

    // CDCStore provides access to the interpreter's event store.
    CDCStore() CDCAccessor
    // FactDB provides access to the interpreter's fact database.
    FactDB() FactAccessor
}

// CDCAccessor wraps the interpreter's eventStore + eventPatterns + eventMutex.
type CDCAccessor interface {
    Lock()
    Unlock()
    RLock()
    RUnlock()
    Events() []map[string]any
    SetEvents(store []map[string]any)
    AppendEvent(event map[string]any)
    Patterns() map[string]any
    SetPattern(name string, pattern map[string]any)
}

// FactAccessor wraps the interpreter's factDatabase + factMutex.
type FactAccessor interface {
    Lock()
    Unlock()
    RLock()
    RUnlock()
    Get(key string) (any, bool)
    Set(key string, val any)
    Delete(key string)
    All() map[string]any
}

// Module is implemented by each stdlib package.
type Module interface {
    Name() string
    Functions() map[string]ModuleFunc
}

// TypeErrorf creates a type error suitable for panic in stdlib module functions.
func TypeErrorf(pos Position, format string, args ...any) error {
    return typeError(pos, format, args...)
}

// RuntimeErrorf creates a runtime error suitable for panic in stdlib module functions.
func RuntimeErrorf(pos Position, format string, args ...any) error {
    return runtimeError(pos, format, args...)
}
```

- [ ] **Step 4: Create `interpreter/module_registry.go`**

```go
package interpreter

import (
    "fmt"
    "io"
    "sort"
    "strings"
    "sync"
)

var globalModuleRegistry = map[string]Module{}

// RegisterModule adds m to the global registry. Called from stdlib init() functions.
func RegisterModule(m Module) {
    globalModuleRegistry[m.Name()] = m
}

// LookupModule returns the module for name, or (nil, false).
func LookupModule(name string) (Module, bool) {
    m, ok := globalModuleRegistry[name]
    return m, ok
}

// knownModuleNames returns sorted module names for error messages.
func knownModuleNames() []string {
    names := make([]string, 0, len(globalModuleRegistry))
    for name := range globalModuleRegistry {
        names = append(names, name)
    }
    sort.Strings(names)
    return names
}

// buildNamespaceObject converts a Module's Functions() into a map[string]Value
// where each value is a builtinFunction. Called when an import statement executes.
func buildNamespaceObject(mod Module, interp *interpreter) map[string]Value {
    ns := make(map[string]Value, len(mod.Functions()))
    for fname, fn := range mod.Functions() {
        fn, modName, fname := fn, mod.Name(), fname
        ns[fname] = builtinFunction{
            Function: func(interp *interpreter, pos Position, args []Value) Value {
                ctx := &interpreterContext{interp: interp}
                return fn(ctx, pos, args)
            },
            Name: modName + "." + fname,
        }
    }
    return ns
}

// interpreterContext wraps *interpreter to implement ModuleContext.
type interpreterContext struct{ interp *interpreter }

func (c *interpreterContext) LookupVar(name string) (Value, bool) {
    return c.interp.lookup(name)
}
func (c *interpreterContext) SetVar(name string, val Value) {
    c.interp.setVar(name, val)
}
func (c *interpreterContext) Stdout() io.Writer {
    return c.interp.stdout
}
func (c *interpreterContext) GetContext(key string) (map[string]any, bool) {
    val, ok := c.interp.lookup("_" + key + "_ctx")
    if !ok {
        return nil, false
    }
    m, ok := val.(map[string]any)
    return m, ok
}
func (c *interpreterContext) ReturnValue(val Value, pos Position) {
    panic(returnResult{value: val, pos: pos})
}
func (c *interpreterContext) CDCStore() CDCAccessor {
    return &cdcAccessor{interp: c.interp}
}
func (c *interpreterContext) FactDB() FactAccessor {
    return &factAccessor{interp: c.interp}
}

// cdcAccessor implements CDCAccessor using interpreter's event fields.
type cdcAccessor struct{ interp *interpreter }

func (a *cdcAccessor) Lock()    { a.interp.eventMutex.Lock() }
func (a *cdcAccessor) Unlock()  { a.interp.eventMutex.Unlock() }
func (a *cdcAccessor) RLock()   { a.interp.eventMutex.RLock() }
func (a *cdcAccessor) RUnlock() { a.interp.eventMutex.RUnlock() }
func (a *cdcAccessor) Events() []map[string]any { return a.interp.eventStore }
func (a *cdcAccessor) SetEvents(store []map[string]any) { a.interp.eventStore = store }
func (a *cdcAccessor) AppendEvent(event map[string]any) {
    a.interp.eventStore = append(a.interp.eventStore, event)
}
func (a *cdcAccessor) Patterns() map[string]any { return a.interp.eventPatterns }
func (a *cdcAccessor) SetPattern(name string, pattern map[string]any) {
    a.interp.eventPatterns[name] = pattern
}

// factAccessor implements FactAccessor using interpreter's fact fields.
type factAccessor struct{ interp *interpreter }

func (a *factAccessor) Lock()    { a.interp.factMutex.Lock() }
func (a *factAccessor) Unlock()  { a.interp.factMutex.Unlock() }
func (a *factAccessor) RLock()   { a.interp.factMutex.RLock() }
func (a *factAccessor) RUnlock() { a.interp.factMutex.RUnlock() }
func (a *factAccessor) Get(key string) (any, bool) {
    v, ok := a.interp.factDatabase[key]
    return v, ok
}
func (a *factAccessor) Set(key string, val any)  { a.interp.factDatabase[key] = val }
func (a *factAccessor) Delete(key string)        { delete(a.interp.factDatabase, key) }
func (a *factAccessor) All() map[string]any      { return a.interp.factDatabase }
```

- [ ] **Step 5: Run tests**

```bash
go test ./interpreter/... -run "TestRegisterAndLookupModule|TestLookupUnknownModule|TestKnownModuleNames|TestBuildNamespaceObject" -v
```
Expected: all 4 PASS.

- [ ] **Step 6: Commit**

```bash
git add interpreter/module.go interpreter/module_registry.go interpreter/module_test.go
git commit -m "feat(module): add ModuleContext interface + registry + namespace builder"
```

---

## Task 2: Parser — AS token + extended Import

**Files:**
- Modify: `interpreter/tokenizer.go`
- Modify: `interpreter/ast.go`
- Modify: `interpreter/parser.go`
- Test: `interpreter/parser_test.go` (add cases)

- [ ] **Step 1: Write failing tests**

Add to `interpreter/parser_test.go` (create if not exists):

```go
// interpreter/import_module_test.go
package interpreter

import "testing"

func TestParseImportModule_NoAlias(t *testing.T) {
    prog, err := ParseProgram([]byte(`import "database"`))
    if err != nil { t.Fatal(err) }
    if len(prog.Statements) != 1 { t.Fatalf("got %d statements", len(prog.Statements)) }
    imp, ok := prog.Statements[0].(*Import)
    if !ok { t.Fatal("not *Import") }
    if imp.Path != "database" { t.Fatalf("Path=%s", imp.Path) }
    if imp.Alias != "" { t.Fatalf("Alias=%s", imp.Alias) }
    if !imp.IsModule { t.Fatal("IsModule should be true") }
}

func TestParseImportModule_WithAlias(t *testing.T) {
    prog, err := ParseProgram([]byte(`import "database" as db`))
    if err != nil { t.Fatal(err) }
    imp := prog.Statements[0].(*Import)
    if imp.Path != "database" { t.Fatalf("Path=%s", imp.Path) }
    if imp.Alias != "db" { t.Fatalf("Alias=%s", imp.Alias) }
    if !imp.IsModule { t.Fatal("IsModule should be true") }
}

func TestParseImportFile_NoAlias(t *testing.T) {
    prog, err := ParseProgram([]byte(`import "utils.din"`))
    if err != nil { t.Fatal(err) }
    imp := prog.Statements[0].(*Import)
    if imp.IsModule { t.Fatal("IsModule should be false for .din file") }
    if imp.Path != "utils.din" { t.Fatalf("Path=%s", imp.Path) }
}

func TestParseImportFile_WithPath(t *testing.T) {
    prog, err := ParseProgram([]byte(`import "./libs/utils.din"`))
    if err != nil { t.Fatal(err) }
    imp := prog.Statements[0].(*Import)
    if imp.IsModule { t.Fatal("IsModule should be false for path with /") }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./interpreter/... -run "TestParseImportModule|TestParseImportFile" -v
```
Expected: FAIL — `Import.Path`, `Import.Alias`, `Import.IsModule` undefined.

- [ ] **Step 3: Add AS token to tokenizer**

In `interpreter/tokenizer.go`, in the keyword constants block after `ELIF = 521`:

```go
AS   = 522
```

In `keywordTokens` map:
```go
"as": AS,
```

In `tokenName` map (if it exists as a reverse map):
```go
AS: "as",
```

- [ ] **Step 4: Extend Import struct in ast.go**

Replace the existing `Import` struct at line ~964:

```go
// Import represents an import statement.
// IsModule=true: import "database" → stdlib module lookup.
// IsModule=false: import "file.din" → file execution (existing behavior).
type Import struct {
    pos      Position
    Path     string // module name ("database") or file path ("utils.din")
    Alias    string // "" = use Path as alias; "db" = import ... as db
    IsModule bool   // true when Path has no "/" and no ".din" suffix
}

func (s *Import) Position() Position { return s.pos }

func (s *Import) String() string {
    if s.Alias != "" {
        return fmt.Sprintf("import %q as %s", s.Path, s.Alias)
    }
    return fmt.Sprintf("import %q", s.Path)
}
```

Also remove the now-unused `Filename` field reference in `ast.go` (it was `Filename string`; replace with `Path string`).

- [ ] **Step 5: Extend parser import_() in parser.go**

Find the `import_()` function (around line 788) and replace:

```go
// import = IMPORT STR [AS IDENT]
func (p *parser) import_() Statement {
    pos := p.pos
    p.expect(IMPORT)
    if p.tok != STR {
        p.error("import requires a string path or module name, got %s", p.tok)
    }
    path := p.val
    p.next()

    alias := ""
    if p.tok == AS {
        p.next()
        if p.tok != IDENT {
            p.error("import ... as requires an identifier, got %s", p.tok)
        }
        alias = p.val
        p.next()
    }

    // A module name has no path separator and no .din suffix.
    // Anything else (e.g. "./utils.din", "libs/helper.din") is a file import.
    isModule := !strings.Contains(path, "/") &&
        !strings.Contains(path, "\\") &&
        !strings.HasSuffix(path, ".din")

    return &Import{pos: pos, Path: path, Alias: alias, IsModule: isModule}
}
```

- [ ] **Step 6: Fix references to old Import.Filename field**

Search for `s.Filename` or `imp.Filename` or `.Filename` in `interpreter/`:

```bash
grep -rn "\.Filename" interpreter/
```

In `interpreter/interpreter.go` `executeImport`, replace `s.Filename` with `s.Path`.
In `interpreter/ast.go` `String()` method, update if it referenced `Filename`.

- [ ] **Step 7: Run tests**

```bash
go test ./interpreter/... -run "TestParseImportModule|TestParseImportFile" -v
go test ./interpreter/... -v 2>&1 | tail -5
```
Expected: new tests PASS, existing tests PASS.

- [ ] **Step 8: Commit**

```bash
git add interpreter/tokenizer.go interpreter/ast.go interpreter/parser.go interpreter/import_module_test.go
git commit -m "feat(parser): add AS token + extend Import with Alias/IsModule fields"
```

---

## Task 3: Interpreter — module import dispatch

**Files:**
- Modify: `interpreter/interpreter.go` (executeImport)
- Test: `interpreter/import_module_test.go` (extend)

- [ ] **Step 1: Write failing tests**

Add to `interpreter/import_module_test.go`:

```go
func TestImportModule_InjectsNamespace(t *testing.T) {
    // mockModule already registered in TestRegisterAndLookupModule — reset registry
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
    if err != nil { t.Fatal(err) }
    _, err = Execute(prog, config)
    if err != nil { t.Fatal(err) }
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
    Execute(prog, config)
    if !strings.Contains(buf.String(), "hello from mymod") {
        t.Fatalf("got: %s", buf.String())
    }
}

func TestImportModule_UnknownModuleError(t *testing.T) {
    globalModuleRegistry = map[string]Module{}

    config := TestConfig()
    prog, _ := ParseProgram([]byte(`import "nonexistent"`))
    _, err := Execute(prog, config)
    if err == nil { t.Fatal("expected error for unknown module") }
    if !strings.Contains(err.Error(), "unknown module") {
        t.Fatalf("wrong error: %v", err)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./interpreter/... -run "TestImportModule_" -v
```
Expected: FAIL — executeImport still does file-only dispatch.

- [ ] **Step 3: Update executeImport in interpreter.go**

Find `executeImport` (around line 1664). Replace the function body:

```go
func (interp *interpreter) executeImport(s *Import) {
    if s.IsModule {
        mod, ok := LookupModule(s.Path)
        if !ok {
            known := knownModuleNames()
            panic(runtimeError(s.pos,
                "unknown module %q\n  available modules: %s",
                s.Path, strings.Join(known, ", ")))
        }
        alias := s.Alias
        if alias == "" {
            alias = s.Path
        }
        ns := buildNamespaceObject(mod, interp)
        interp.setVar(alias, ns)
        return
    }
    // File import: existing behavior unchanged.
    content, err := os.ReadFile(s.Path)
    if err != nil {
        panic(runtimeError(s.Position(), "failed to import file '%s': %s", s.Path, err))
    }
    prog, err := ParseProgram(content)
    if err != nil {
        panic(runtimeError(s.Position(), "failed to parse imported file '%s': %s", s.Path, err))
    }
    for _, statement := range prog.Statements {
        interp.executeStatement(statement)
    }
}
```

Also update any reference to `s.Filename` → `s.Path` in the `case *Import:` branch of `executeStatement`.

- [ ] **Step 4: Run tests**

```bash
go test ./interpreter/... -run "TestImportModule_" -v
go test ./interpreter/... -v 2>&1 | grep -E "FAIL|ok"
```
Expected: new tests PASS, no regressions.

- [ ] **Step 5: Commit**

```bash
git add interpreter/interpreter.go interpreter/import_module_test.go
git commit -m "feat(interpreter): dispatch import to module registry when IsModule=true"
```

---

## Task 4: stdlib/regex

**Files:**
- Create: `stdlib/regex/regex.go`
- Test: `stdlib/regex/regex_test.go`
- Note: Delete `interpreter/fn_regex.go` in Task 13.

- [ ] **Step 1: Write failing test**

```go
// stdlib/regex/regex_test.go
package stdlib_regex_test

import (
    "strings"
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/regex"
)

func runScript(t *testing.T, src string) string {
    t.Helper()
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    prog, err := interpreter.ParseProgram([]byte(src))
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestRegexMatch(t *testing.T) {
    out := runScript(t, `
import "regex"
print(regex.match("hello123", "\\d+"))
`)
    if !strings.Contains(out, "true") { t.Fatalf("got: %s", out) }
}

func TestRegexFind(t *testing.T) {
    out := runScript(t, `
import "regex"
print(regex.find("hello 123 world", "\\d+"))
`)
    if !strings.Contains(out, "123") { t.Fatalf("got: %s", out) }
}

func TestRegexIsMatch(t *testing.T) {
    out := runScript(t, `
import "regex"
print(regex.is_match("^\\d+$", "42"))
`)
    if !strings.Contains(out, "true") { t.Fatalf("got: %s", out) }
}

func TestRegexReplace(t *testing.T) {
    out := runScript(t, `
import "regex"
print(regex.replace("hello 123 world", "\\d+", "NUM"))
`)
    if !strings.Contains(out, "hello NUM world") { t.Fatalf("got: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/regex/... -v
```
Expected: FAIL — package does not exist.

- [ ] **Step 3: Create `stdlib/regex/regex.go`**

```go
package stdlib_regex

import (
    "fmt"

    "github.com/bonkzero404/uddin-lang/interpreter"
    "github.com/coregx/coregex"
)

type RegexModule struct{}

func (m *RegexModule) Name() string { return "regex" }

func (m *RegexModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "match":    matchFunc,
        "is_match": isMatchFunc,
        "find":     findFunc,
        "find_all": findAllFunc,
        "replace":  replaceFunc,
        "split":    splitFunc,
    }
}

func init() { interpreter.RegisterModule(&RegexModule{}) }

func compilePattern(pos interpreter.Position, arg interpreter.Value) (*coregex.Regexp, error) {
    switch p := arg.(type) {
    case string:
        return coregex.Compile(p)
    case *coregex.Regexp:
        return p, nil
    default:
        return nil, interpreter.TypeErrorf(pos, "regex: pattern must be string or compiled regexp, got %T", arg)
    }
}

func matchFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 2 {
        return interpreter.Value(fmt.Errorf("regex.match() requires 2 arguments, got %d", len(args)))
    }
    text, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.match() first argument must be string"))
    }
    re, err := compilePattern(pos, args[1])
    if err != nil {
        return interpreter.Value(fmt.Errorf("regex.match(): %v", err))
    }
    return interpreter.Value(re.MatchString(text))
}

func isMatchFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 2 {
        return interpreter.Value(fmt.Errorf("regex.is_match() requires 2 arguments, got %d", len(args)))
    }
    re, err := compilePattern(pos, args[0])
    if err != nil {
        return interpreter.Value(false)
    }
    text, ok := args[1].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "regex.is_match() second argument must be string"))
    }
    return interpreter.Value(re.MatchString(text))
}

func findFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 2 {
        return interpreter.Value(fmt.Errorf("regex.find() requires 2 arguments, got %d", len(args)))
    }
    text, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.find() first argument must be string"))
    }
    re, err := compilePattern(pos, args[1])
    if err != nil {
        return interpreter.Value(fmt.Errorf("regex.find(): %v", err))
    }
    match := re.FindString(text)
    if match == "" {
        return interpreter.Value(nil)
    }
    return interpreter.Value(match)
}

func findAllFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) < 2 || len(args) > 3 {
        return interpreter.Value(fmt.Errorf("regex.find_all() requires 2 or 3 arguments, got %d", len(args)))
    }
    text, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.find_all() first argument must be string"))
    }
    re, err := compilePattern(pos, args[1])
    if err != nil {
        return interpreter.Value(fmt.Errorf("regex.find_all(): %v", err))
    }
    limit := -1
    if len(args) == 3 {
        if l, ok := args[2].(int); ok {
            limit = l
        }
    }
    matches := re.FindAllString(text, limit)
    result := make([]interpreter.Value, len(matches))
    for i, m := range matches {
        result[i] = interpreter.Value(m)
    }
    return interpreter.Value(&result)
}

func replaceFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 3 {
        return interpreter.Value(fmt.Errorf("regex.replace() requires 3 arguments, got %d", len(args)))
    }
    text, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.replace() first argument must be string"))
    }
    re, err := compilePattern(pos, args[1])
    if err != nil {
        return interpreter.Value(fmt.Errorf("regex.replace(): %v", err))
    }
    replacement, ok := args[2].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.replace() third argument must be string"))
    }
    return interpreter.Value(re.ReplaceAllString(text, replacement))
}

func splitFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) < 2 || len(args) > 3 {
        return interpreter.Value(fmt.Errorf("regex.split() requires 2 or 3 arguments, got %d", len(args)))
    }
    text, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("regex.split() first argument must be string"))
    }
    re, err := compilePattern(pos, args[1])
    if err != nil {
        return interpreter.Value(fmt.Errorf("regex.split(): %v", err))
    }
    limit := -1
    if len(args) == 3 {
        if l, ok := args[2].(int); ok {
            limit = l
        }
    }
    parts := re.Split(text, limit)
    result := make([]interpreter.Value, len(parts))
    for i, p := range parts {
        result[i] = interpreter.Value(p)
    }
    return interpreter.Value(&result)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/regex/... -v
```
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/regex/regex.go stdlib/regex/regex_test.go
git commit -m "feat(stdlib/regex): add regex module with match/is_match/find/find_all/replace/split"
```

---

## Task 5: stdlib/datetime

**Files:**
- Create: `stdlib/datetime/datetime.go`
- Test: `stdlib/datetime/datetime_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/datetime/datetime_test.go
package stdlib_datetime_test

import (
    "strings"
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/datetime"
)

func runScript(t *testing.T, src string) string {
    t.Helper()
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    prog, err := interpreter.ParseProgram([]byte(src))
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestDatetimeNow(t *testing.T) {
    out := runScript(t, `
import "datetime"
now = datetime.now()
print(typeof(now))
`)
    if !strings.Contains(out, "string") && !strings.Contains(out, "object") {
        t.Fatalf("got: %s", out)
    }
}

func TestDatetimeSleep(t *testing.T) {
    // just verify it doesn't panic
    runScript(t, `
import "datetime"
datetime.sleep(1)
`)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/datetime/... -v
```
Expected: FAIL — package not found.

- [ ] **Step 3: Create `stdlib/datetime/datetime.go`**

Port all functions from `interpreter/fn_datetime.go`, changing signature from
`func xFunc(interp *interpreter, pos Position, args []Value) Value`
to `func xFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value`.

Key functions to port (full list from fn_datetime.go):
`date_now→now`, `time_now→time_now`, `sleep→sleep`, `date_format→format`,
`date_parse→parse`, `date_format_new→format_new`, `date_add→add`,
`date_subtract→subtract`, `date_diff→diff`, `date_between→between`,
`date_compare→compare`.

```go
package stdlib_datetime

import (
    "fmt"
    "time"

    "github.com/bonkzero404/uddin-lang/interpreter"
)

type DatetimeModule struct{}

func (m *DatetimeModule) Name() string { return "datetime" }

func (m *DatetimeModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "now":        nowFunc,
        "time_now":   timeNowFunc,
        "sleep":      sleepFunc,
        "format":     formatFunc,
        "parse":      parseFunc,
        "format_new": formatNewFunc,
        "add":        addFunc,
        "subtract":   subtractFunc,
        "diff":       diffFunc,
        "between":    betweenFunc,
        "compare":    compareFunc,
    }
}

func init() { interpreter.RegisterModule(&DatetimeModule{}) }

// nowFunc implements datetime.now() → current date/time string
func nowFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    return interpreter.Value(time.Now().Format(time.RFC3339))
}

// timeNowFunc implements datetime.time_now() → Unix timestamp in ms
func timeNowFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    return interpreter.Value(int(time.Now().UnixMilli()))
}

// sleepFunc implements datetime.sleep(ms int)
func sleepFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 1 {
        panic(interpreter.TypeErrorf(pos, "datetime.sleep() requires 1 argument"))
    }
    ms, ok := args[0].(int)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "datetime.sleep() requires an integer (milliseconds)"))
    }
    time.Sleep(time.Duration(ms) * time.Millisecond)
    return interpreter.Value(nil)
}

// formatFunc implements datetime.format(dateStr, layout)
func formatFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 2 {
        return interpreter.Value(fmt.Errorf("datetime.format() requires 2 arguments"))
    }
    dateStr, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("datetime.format() first argument must be string"))
    }
    layout, ok := args[1].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("datetime.format() second argument must be string"))
    }
    t, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        t, err = time.Parse("2006-01-02", dateStr)
        if err != nil {
            return interpreter.Value(fmt.Errorf("datetime.format() cannot parse date: %v", err))
        }
    }
    return interpreter.Value(t.Format(layout))
}

// parseFunc, formatNewFunc, addFunc, subtractFunc, diffFunc, betweenFunc, compareFunc:
// Port directly from interpreter/fn_datetime.go following the same pattern as formatFunc above.
// Replace `interp *interpreter` → `_ interpreter.ModuleContext`
// Replace `Position` → `interpreter.Position`
// Replace `Value` → `interpreter.Value`
// Replace `typeError(pos, ...)` → `interpreter.TypeErrorf(pos, ...)`
```

Complete the remaining functions by porting them 1:1 from `interpreter/fn_datetime.go`.

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/datetime/... -v
go build ./stdlib/datetime/...
```
Expected: PASS, no build errors.

- [ ] **Step 5: Commit**

```bash
git add stdlib/datetime/
git commit -m "feat(stdlib/datetime): add datetime module"
```

---

## Task 6: stdlib/json

**Files:**
- Create: `stdlib/json/json.go`
- Test: `stdlib/json/json_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/json/json_test.go
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
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestJsonParse(t *testing.T) {
    out := runScript(t, `
import "json"
obj = json.parse("{\"name\": \"alice\", \"age\": 30}")
print(obj["name"])
`)
    if !strings.Contains(out, "alice") { t.Fatalf("got: %s", out) }
}

func TestJsonStringify(t *testing.T) {
    out := runScript(t, `
import "json"
obj = {"name": "alice", "age": 30}
print(json.stringify(obj))
`)
    if !strings.Contains(out, "alice") { t.Fatalf("got: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/json/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/json/json.go`**

Port from `interpreter/fn_serialization.go`. Functions: `json_parse→parse`, `json_stringify→stringify`, `xml_parse→xml_parse`, `xml_stringify→xml_stringify`.

```go
package stdlib_json

import (
    "github.com/bonkzero404/uddin-lang/interpreter"
)

type JSONModule struct{}

func (m *JSONModule) Name() string { return "json" }

func (m *JSONModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "parse":         parseFunc,
        "stringify":     stringifyFunc,
        "xml_parse":     xmlParseFunc,
        "xml_stringify": xmlStringifyFunc,
    }
}

func init() { interpreter.RegisterModule(&JSONModule{}) }

// parseFunc, stringifyFunc, xmlParseFunc, xmlStringifyFunc:
// Port 1:1 from interpreter/fn_serialization.go.
// Change signature: (interp *interpreter, pos Position, args []Value)
//               to: (_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value)
// Prefix all types: Value → interpreter.Value, typeError → interpreter.TypeErrorf
```

Complete all 4 functions from fn_serialization.go.

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/json/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/json/
git commit -m "feat(stdlib/json): add json module with parse/stringify/xml_parse/xml_stringify"
```

---

## Task 7: stdlib/fs

**Files:**
- Create: `stdlib/fs/fs.go`
- Test: `stdlib/fs/fs_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/fs/fs_test.go
package stdlib_fs_test

import (
    "os"
    "strings"
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/fs"
)

func runScript(t *testing.T, src string) string {
    t.Helper()
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    prog, err := interpreter.ParseProgram([]byte(src))
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestFsWriteRead(t *testing.T) {
    defer os.Remove("test_fs_tmp.txt")
    out := runScript(t, `
import "fs"
fs.write("test_fs_tmp.txt", "hello fs")
content = fs.read("test_fs_tmp.txt")
print(content)
`)
    if !strings.Contains(out, "hello fs") { t.Fatalf("got: %s", out) }
}

func TestFsExists(t *testing.T) {
    out := runScript(t, `
import "fs"
print(fs.exists("/nonexistent_path_xyz"))
`)
    if !strings.Contains(out, "false") { t.Fatalf("got: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/fs/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/fs/fs.go`**

Port from `interpreter/fn_file_system.go`. All 18 functions:
`read_file→read`, `write_file→write`, `file_exists→exists`, `file_size→size`,
`file_modified→modified`, `file_permissions→permissions`, `mkdir→mkdir`,
`rmdir→rmdir`, `list_dir→list`, `copy_file→copy`, `move_file→move`,
`delete_file→delete`, `path_join→path_join`, `path_dirname→dirname`,
`path_basename→basename`, `path_ext→ext`, `getcwd→cwd`, `chdir→chdir`.

```go
package stdlib_fs

import (
    "github.com/bonkzero404/uddin-lang/interpreter"
)

type FSModule struct{}

func (m *FSModule) Name() string { return "fs" }

func (m *FSModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "read":        readFunc,
        "write":       writeFunc,
        "exists":      existsFunc,
        "size":        sizeFunc,
        "modified":    modifiedFunc,
        "permissions": permissionsFunc,
        "mkdir":       mkdirFunc,
        "rmdir":       rmdirFunc,
        "list":        listFunc,
        "copy":        copyFunc,
        "move":        moveFunc,
        "delete":      deleteFunc,
        "path_join":   pathJoinFunc,
        "dirname":     dirnameFunc,
        "basename":    basenameFunc,
        "ext":         extFunc,
        "cwd":         cwdFunc,
        "chdir":       chdirFunc,
    }
}

func init() { interpreter.RegisterModule(&FSModule{}) }

// Port all 18 functions 1:1 from interpreter/fn_file_system.go.
// Signature change: (interp *interpreter, pos Position, args []Value)
//               to: (_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value)
```

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/fs/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/fs/
git commit -m "feat(stdlib/fs): add fs module with read/write/exists/mkdir/list/copy/move/delete/path ops"
```

---

## Task 8: stdlib/http

**Files:**
- Create: `stdlib/http/http.go`
- Test: `stdlib/http/http_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/http/http_test.go
package stdlib_http_test

import (
    "strings"
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/http"
)

func TestHttpModuleRegistered(t *testing.T) {
    mod, ok := interpreter.LookupModule("http")
    if !ok { t.Fatal("http module not registered") }
    fns := mod.Functions()
    for _, name := range []string{"get", "post", "put", "delete", "request"} {
        if _, ok := fns[name]; !ok {
            t.Errorf("missing function: %s", name)
        }
    }
}

func TestHttpGet_BadURL(t *testing.T) {
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    prog, _ := interpreter.ParseProgram([]byte(`
import "http"
result = http.get("http://localhost:1")
print(typeof(result))
`))
    interpreter.Execute(prog, config)
    // Should return error object, not panic
    out := buf.String()
    if strings.Contains(out, "panic") { t.Fatalf("unexpected panic: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/http/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/http/http.go`**

Port from `interpreter/fn_network.go`. Functions:
HTTP client: `http_get→get`, `http_post→post`, `http_put→put`, `http_delete→delete`, `http_request→request`
HTTP server: `http_server_start→server_start`, `http_server_stop→server_stop`, `http_server_route→server_route`, `http_response→response`
TCP: `tcp_connect→tcp_connect`, `tcp_listen→tcp_listen`, `tcp_accept→tcp_accept`, `tcp_read→tcp_read`, `tcp_write→tcp_write`, `tcp_close→tcp_close`
UDP: `udp_connect→udp_connect`, `udp_listen→udp_listen`, `udp_read→udp_read`, `udp_write→udp_write`, `udp_close→udp_close`
Net: `net_resolve→net_resolve`, `net_ping→net_ping`

```go
package stdlib_http

import (
    "github.com/bonkzero404/uddin-lang/interpreter"
)

type HTTPModule struct{}

func (m *HTTPModule) Name() string { return "http" }

func (m *HTTPModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "get":          getFunc,
        "post":         postFunc,
        "put":          putFunc,
        "delete":       deleteFunc,
        "request":      requestFunc,
        "server_start": serverStartFunc,
        "server_stop":  serverStopFunc,
        "server_route": serverRouteFunc,
        "response":     responseFunc,
        "tcp_connect":  tcpConnectFunc,
        "tcp_listen":   tcpListenFunc,
        "tcp_accept":   tcpAcceptFunc,
        "tcp_read":     tcpReadFunc,
        "tcp_write":    tcpWriteFunc,
        "tcp_close":    tcpCloseFunc,
        "udp_connect":  udpConnectFunc,
        "udp_listen":   udpListenFunc,
        "udp_read":     udpReadFunc,
        "udp_write":    udpWriteFunc,
        "udp_close":    udpCloseFunc,
        "net_resolve":  netResolveFunc,
        "net_ping":     netPingFunc,
    }
}

func init() { interpreter.RegisterModule(&HTTPModule{}) }

// Port all functions 1:1 from interpreter/fn_network.go.
// Signature: (interp *interpreter, pos Position, args []Value)
//        to: (_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value)
```

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/http/... -v
go build ./stdlib/http/...
```
Expected: PASS / builds clean.

- [ ] **Step 5: Commit**

```bash
git add stdlib/http/
git commit -m "feat(stdlib/http): add http module with HTTP client/server, TCP, UDP, net utils"
```

---

## Task 9: stdlib/waf

**Files:**
- Create: `stdlib/waf/waf.go`
- Test: `stdlib/waf/waf_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/waf/waf_test.go
package stdlib_waf_test

import (
    "strings"
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/waf"
)

func TestWafCidrMatch(t *testing.T) {
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    prog, _ := interpreter.ParseProgram([]byte(`
import "waf"
print(waf.cidr_match("192.168.1.5", "192.168.1.0/24"))
`))
    interpreter.Execute(prog, config)
    if !strings.Contains(buf.String(), "true") { t.Fatalf("got: %s", buf.String()) }
}

func TestWafPathMatch(t *testing.T) {
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Vars = map[string]interpreter.Value{
        "_waf_ctx": map[string]any{"path": "/api/users"},
    }
    config.Stdout = &buf
    prog, _ := interpreter.ParseProgram([]byte(`
import "waf"
print(waf.path_match("/api/*"))
`))
    interpreter.Execute(prog, config)
    if !strings.Contains(buf.String(), "true") { t.Fatalf("got: %s", buf.String()) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/waf/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/waf/waf.go`**

Port from `interpreter/fn_wafio.go`. Uses `ctx.GetContext("waf")` instead of `wafCtx(interp)`.
Uses `ctx.Stdout()` instead of `interp.stdout`.
Uses `ctx.ReturnValue(val, pos)` instead of `panic(returnResult{...})`.

```go
package stdlib_waf

import (
    "fmt"
    "net"
    "net/url"
    "path/filepath"
    "strings"

    "github.com/bonkzero404/uddin-lang/interpreter"
)

type WafModule struct{}

func (m *WafModule) Name() string { return "waf" }

func (m *WafModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "header":        headerFunc,
        "query":         queryFunc,
        "body_contains": bodyContainsFunc,
        "cidr_match":    cidrMatchFunc,
        "path_match":    pathMatchFunc,
        "score":         scoreFunc,
        "action":        actionFunc,
        "detected":      detectedFunc,
        "detected_any":  detectedAnyFunc,
        "detected_list": detectedListFunc,
        "return":        returnFunc,
    }
}

func init() { interpreter.RegisterModule(&WafModule{}) }

func getWafCtx(ctx interpreter.ModuleContext) map[string]any {
    m, _ := ctx.GetContext("waf")
    return m
}

func headerFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 1 {
        return interpreter.Value(fmt.Errorf("waf.header() requires 1 argument"))
    }
    name, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("waf.header() requires a string argument"))
    }
    wctx := getWafCtx(ctx)
    if wctx == nil { return interpreter.Value(nil) }
    headers, ok := wctx["headers"].(map[string]string)
    if !ok { return interpreter.Value(nil) }
    lower := strings.ToLower(name)
    for k, v := range headers {
        if strings.ToLower(k) == lower {
            return interpreter.Value(v)
        }
    }
    return interpreter.Value(nil)
}

func cidrMatchFunc(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 2 {
        return interpreter.Value(fmt.Errorf("waf.cidr_match() requires 2 arguments"))
    }
    ipStr, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("waf.cidr_match() first argument must be string"))
    }
    cidrStr, ok := args[1].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("waf.cidr_match() second argument must be string"))
    }
    ip := net.ParseIP(ipStr)
    if ip == nil { return interpreter.Value(false) }
    _, network, err := net.ParseCIDR(cidrStr)
    if err != nil { return interpreter.Value(false) }
    return interpreter.Value(network.Contains(ip))
}

func pathMatchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 1 {
        return interpreter.Value(fmt.Errorf("waf.path_match() requires 1 argument"))
    }
    pattern, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("waf.path_match() requires a string argument"))
    }
    wctx := getWafCtx(ctx)
    if wctx == nil { return interpreter.Value(false) }
    path, _ := wctx["path"].(string)
    matched, err := filepath.Match(pattern, path)
    if err != nil { return interpreter.Value(false) }
    return interpreter.Value(matched)
}

func returnFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) != 1 {
        return interpreter.Value(fmt.Errorf("waf.return() requires 1 argument"))
    }
    action, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("waf.return() requires a string argument"))
    }
    normalized := strings.ToUpper(strings.TrimSpace(action))
    switch normalized {
    case "ALLOW", "LOG", "BLOCK":
    default:
        return interpreter.Value(fmt.Errorf("waf.return() invalid action %q: must be ALLOW, LOG, or BLOCK", action))
    }
    fmt.Fprintf(ctx.Stdout(), "%s", normalized)
    ctx.ReturnValue(interpreter.Value(normalized), pos)
    return interpreter.Value(nil) // unreachable
}

// Port remaining functions from interpreter/fn_wafio.go:
// queryFunc, bodyContainsFunc, scoreFunc, actionFunc,
// detectedFunc, detectedAnyFunc, detectedListFunc
// Same pattern: use ctx.GetContext("waf") instead of wafCtx(interp)
```

Complete the remaining 7 functions by porting from `interpreter/fn_wafio.go`.

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/waf/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/waf/
git commit -m "feat(stdlib/waf): add waf module using ModuleContext.GetContext + ReturnValue"
```

---

## Task 10: stdlib/cdc

**Files:**
- Create: `stdlib/cdc/cdc.go`
- Test: `stdlib/cdc/cdc_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/cdc/cdc_test.go
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
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestCDCEmitAndCount(t *testing.T) {
    out := runScript(t, `
import "cdc"
cdc.emit("login", {"user": "alice"})
cdc.emit("login", {"user": "bob"})
print(cdc.count("login"))
`)
    if !strings.Contains(out, "2") { t.Fatalf("got: %s", out) }
}

func TestCDCClear(t *testing.T) {
    out := runScript(t, `
import "cdc"
cdc.emit("page_view", {})
cdc.clear("page_view")
print(cdc.count("page_view"))
`)
    if !strings.Contains(out, "0") { t.Fatalf("got: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/cdc/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/cdc/cdc.go`**

Port from `interpreter/fn_cep.go`. Use `ctx.CDCStore()` instead of direct `interp.eventStore`.

```go
package stdlib_cdc

import (
    "fmt"
    "time"

    "github.com/bonkzero404/uddin-lang/interpreter"
)

type CDCModule struct{}

func (m *CDCModule) Name() string { return "cdc" }

func (m *CDCModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "emit":           emitFunc,
        "define_pattern": definePatternFunc,
        "get_window":     getWindowFunc,
        "clear":          clearFunc,
        "count":          countFunc,
    }
}

func init() { interpreter.RegisterModule(&CDCModule{}) }

func emitFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) < 1 {
        return interpreter.Value(fmt.Errorf("cdc.emit() requires at least 1 argument"))
    }
    eventType, ok := args[0].(string)
    if !ok {
        return interpreter.Value(fmt.Errorf("cdc.emit() first argument must be a string (event type)"))
    }

    store := ctx.CDCStore()
    store.Lock()
    defer store.Unlock()

    event := map[string]any{
        "type":      eventType,
        "timestamp": time.Now().Format(time.RFC3339),
        "id":        fmt.Sprintf("%d", time.Now().UnixNano()),
    }
    if len(args) > 1 {
        event["data"] = args[1]
    }

    store.AppendEvent(event)
    // Keep only last 1000 events
    events := store.Events()
    if len(events) > 1000 {
        store.SetEvents(events[len(events)-1000:])
    }
    return interpreter.Value(true)
}

func countFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    store := ctx.CDCStore()
    store.RLock()
    defer store.RUnlock()

    if len(args) == 0 {
        return interpreter.Value(len(store.Events()))
    }
    eventType, ok := args[0].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "cdc.count() argument must be a string"))
    }
    count := 0
    for _, ev := range store.Events() {
        if et, ok := ev["type"].(string); ok && et == eventType {
            count++
        }
    }
    return interpreter.Value(count)
}

func clearFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    store := ctx.CDCStore()
    store.Lock()
    defer store.Unlock()

    if len(args) == 0 {
        count := len(store.Events())
        store.SetEvents(make([]map[string]any, 0))
        return interpreter.Value(count)
    }
    eventType, ok := args[0].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "cdc.clear() argument must be a string"))
    }
    var remaining []map[string]any
    count := 0
    for _, ev := range store.Events() {
        if et, ok := ev["type"].(string); ok && et == eventType {
            count++
        } else {
            remaining = append(remaining, ev)
        }
    }
    store.SetEvents(remaining)
    return interpreter.Value(count)
}

// Port definePatternFunc and getWindowFunc from interpreter/fn_cep.go
// using ctx.CDCStore() instead of interp.eventStore / interp.eventPatterns
```

Complete `definePatternFunc` and `getWindowFunc` from `interpreter/fn_cep.go`.

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/cdc/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/cdc/
git commit -m "feat(stdlib/cdc): add cdc module using CDCAccessor for event store"
```

---

## Task 11: stdlib/fact

**Files:**
- Create: `stdlib/fact/fact.go`
- Test: `stdlib/fact/fact_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/fact/fact_test.go
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
    if err != nil { t.Fatal(err) }
    _, err = interpreter.Execute(prog, config)
    if err != nil { t.Fatal(err) }
    return buf.String()
}

func TestFactAssertAndExists(t *testing.T) {
    out := runScript(t, `
import "fact"
fact.assert("customer", "alice", {"tier": "premium"})
print(fact.exists("customer", "alice"))
`)
    if !strings.Contains(out, "true") { t.Fatalf("got: %s", out) }
}

func TestFactRetract(t *testing.T) {
    out := runScript(t, `
import "fact"
fact.assert("item", "x")
fact.retract("item", "x")
print(fact.exists("item", "x"))
`)
    if !strings.Contains(out, "false") { t.Fatalf("got: %s", out) }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/fact/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/fact/fact.go`**

Port from `interpreter/fn_fact_database.go`. Use `ctx.FactDB()` instead of `interp.factDatabase`.

```go
package stdlib_fact

import (
    "fmt"
    "strings"

    "github.com/bonkzero404/uddin-lang/interpreter"
)

type FactModule struct{}

func (m *FactModule) Name() string { return "fact" }

func (m *FactModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "assert":  assertFunc,
        "retract": retractFunc,
        "query":   queryFunc,
        "exists":  existsFunc,
        "count":   countFunc,
        "clear":   clearFunc,
        "get_all": getAllFunc,
    }
}

func init() { interpreter.RegisterModule(&FactModule{}) }

func assertFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) < 2 {
        panic(interpreter.TypeErrorf(pos, "fact.assert() requires at least 2 arguments (category, key, [value])"))
    }
    category, ok := args[0].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "fact.assert() first argument must be a string (category)"))
    }
    key, ok := args[1].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "fact.assert() second argument must be a string (key)"))
    }

    db := ctx.FactDB()
    db.Lock()
    defer db.Unlock()

    fullKey := category + ":" + key
    if len(args) >= 3 {
        db.Set(fullKey, args[2])
    } else {
        db.Set(fullKey, true)
    }
    return interpreter.Value(true)
}

func existsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
    if len(args) < 2 {
        panic(interpreter.TypeErrorf(pos, "fact.exists() requires 2 arguments (category, key)"))
    }
    category, ok := args[0].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "fact.exists() first argument must be string"))
    }
    key, ok := args[1].(string)
    if !ok {
        panic(interpreter.TypeErrorf(pos, "fact.exists() second argument must be string"))
    }

    db := ctx.FactDB()
    db.RLock()
    defer db.RUnlock()

    _, exists := db.Get(category + ":" + key)
    return interpreter.Value(exists)
}

// Port retractFunc, queryFunc, countFunc, clearFunc, getAllFunc
// from interpreter/fn_fact_database.go using ctx.FactDB() instead of interp.factDatabase
```

Complete the remaining 5 functions from `interpreter/fn_fact_database.go`.

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/fact/... -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add stdlib/fact/
git commit -m "feat(stdlib/fact): add fact module using FactAccessor for knowledge base"
```

---

## Task 12: stdlib/database

**Files:**
- Create: `stdlib/database/database.go`
- Test: `stdlib/database/database_test.go`

- [ ] **Step 1: Write failing test**

```go
// stdlib/database/database_test.go
package stdlib_database_test

import (
    "testing"

    "github.com/bonkzero404/uddin-lang/interpreter"
    _ "github.com/bonkzero404/uddin-lang/stdlib/database"
)

func TestDatabaseModuleRegistered(t *testing.T) {
    mod, ok := interpreter.LookupModule("database")
    if !ok { t.Fatal("database module not registered") }
    fns := mod.Functions()
    for _, name := range []string{"connect", "query", "execute", "close"} {
        if _, ok := fns[name]; !ok {
            t.Errorf("missing function: %s", name)
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./stdlib/database/... -v
```
Expected: FAIL.

- [ ] **Step 3: Create `stdlib/database/database.go`**

Move all content from `interpreter/fn_database.go` into this new package.

- Package declaration: `package stdlib_database`
- Add import for `github.com/bonkzero404/uddin-lang/interpreter`
- Keep package-level vars (`databaseConnections`, `databaseMutex`, `ConnectionPool`, etc.) unchanged — they stay as package-level state, no interpreter state needed
- Change all function signatures from `(interp *interpreter, pos Position, args []Value)` to `(_ interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value)`
- Replace `Value(...)` with `interpreter.Value(...)`, `Position` with `interpreter.Position`
- Replace `typeError(pos, ...)` with `interpreter.TypeErrorf(pos, ...)`

Module struct:

```go
type DatabaseModule struct{}

func (m *DatabaseModule) Name() string { return "database" }

func (m *DatabaseModule) Functions() map[string]interpreter.ModuleFunc {
    return map[string]interpreter.ModuleFunc{
        "connect":        connectFunc,
        "query":          queryFunc,
        "execute":        executeFunc,
        "connect_pool":   connectWithPoolFunc,
        "configure_pool": configurePoolFunc,
        "execute_batch":  executeBatchFunc,
        "execute_async":  executeAsyncFunc,
        "async_status":   getAsyncStatusFunc,
        "cancel_async":   cancelAsyncFunc,
        "list_async":     listAsyncOperationsFunc,
        "cleanup_async":  cleanupAsyncOperationsFunc,
        "stream_tables":  streamTablesFunc,
        "stop_stream":    stopStreamFunc,
        "close":          closeFunc,
    }
}

func init() { interpreter.RegisterModule(&DatabaseModule{}) }
```

- [ ] **Step 4: Run tests**

```bash
go test ./stdlib/database/... -v
go build ./stdlib/database/...
```
Expected: PASS / builds clean.

- [ ] **Step 5: Commit**

```bash
git add stdlib/database/
git commit -m "feat(stdlib/database): add database module with connect/query/execute/stream/async ops"
```

---

## Task 13: Cleanup — remove fn_ files + update metadata

**Files:**
- Delete: 9 `interpreter/fn_*.go` files
- Modify: `interpreter/builtin_metadata.go`
- Modify: `interpreter/vm_builtins.go`
- Modify: `interpreter/compiler.go` (if waf-specific refs remain)

- [ ] **Step 1: Remove waf direct entries from vm_builtins.go**

In `interpreter/vm_builtins.go`, in `registerDirectBuiltins()`, remove:
```go
registerVMBuiltinDirect("waf_cidr_match", directWafCidrMatch)
registerVMBuiltinDirect("waf_path_match", directWafPathMatch)
```

Also delete `directWafCidrMatch` and `directWafPathMatch` function implementations from the file.

- [ ] **Step 2: Remove moved functions from builtin_metadata.go**

In `interpreter/builtin_metadata.go`, remove ALL entries for functions now in stdlib.
Functions to remove (complete list):

```
// database: db_connect, db_query, db_execute, db_connect_with_pool,
//   db_configure_pool, db_execute_batch, db_execute_async, db_get_async_status,
//   db_cancel_async, db_list_async_operations, db_cleanup_async_operations,
//   stream_tables, db_stop_stream, db_close

// waf: waf_header, waf_query, waf_body_contains, waf_cidr_match,
//   waf_path_match, waf_score, waf_action, waf_detected, waf_detected_any,
//   waf_detected_list, waf_return

// http: http_get, http_post, http_put, http_delete, http_request,
//   http_server_start, http_server_stop, http_server_route, http_response,
//   tcp_connect, tcp_listen, tcp_accept, tcp_read, tcp_write, tcp_close,
//   udp_connect, udp_listen, udp_read, udp_write, udp_close,
//   net_resolve, net_ping

// cdc: event_emit, event_define_pattern, event_get_window, event_clear, event_count

// fact: fact_assert, fact_retract, fact_query, fact_exists, fact_count,
//   fact_clear, fact_get_all

// regex: is_regex_match, regex_match, regex_find, regex_find_all,
//   regex_replace, regex_split

// datetime: date_now, time_now, sleep, date_format, date_parse,
//   date_format_new, date_add, date_subtract, date_diff, date_between, date_compare

// json: json_parse, json_stringify, xml_parse, xml_stringify

// fs: read_file, write_file, file_exists, file_size, file_modified,
//   file_permissions, mkdir, rmdir, list_dir, copy_file, move_file, delete_file,
//   path_join, path_dirname, path_basename, path_ext, getcwd, chdir
```

- [ ] **Step 3: Delete the fn_ files**

```bash
git rm interpreter/fn_database.go \
       interpreter/fn_wafio.go \
       interpreter/fn_network.go \
       interpreter/fn_cep.go \
       interpreter/fn_file_system.go \
       interpreter/fn_serialization.go \
       interpreter/fn_fact_database.go \
       interpreter/fn_regex.go \
       interpreter/fn_datetime.go
```

- [ ] **Step 4: Delete waf-related tests from interpreter/**

```bash
grep -l "waf_cidr_match\|waf_path_match\|directWaf" interpreter/*_test.go
```

Remove `TestVMBuiltinDirect_WafCidrMatch`, `BenchmarkBuiltinChain_WafRule` from `interpreter/vm_test.go`.
Remove `interpreter/wafio_test.go` if it exists.

- [ ] **Step 5: Run full test suite**

```bash
go test ./... -v 2>&1 | grep -E "FAIL|ok|error"
```
Expected: all packages green. If any `undefined: xxx` errors, trace to missing removal/update.

- [ ] **Step 6: Commit**

```bash
git add interpreter/vm_builtins.go interpreter/builtin_metadata.go interpreter/compiler.go
git add interpreter/vm_test.go
git commit -m "chore: remove fn_ files replaced by stdlib modules; clean waf direct table entries"
```

---

## Task 14: Wire up uddin.go + end-to-end test

**Files:**
- Modify: `uddin.go`
- Test: `uddin_test.go` (extend)

- [ ] **Step 1: Write failing end-to-end test**

Add to `uddin_test.go`:

```go
func TestEngine_DatabaseModuleImport(t *testing.T) {
    // Just test that import "database" resolves without error.
    // Actual DB connection not tested (no DB in CI).
    prog, err := interpreter.ParseProgram([]byte(`import "database"`))
    if err != nil { t.Fatal(err) }
    config := interpreter.TestConfig()
    e := New(config)
    _, err = e.ExecuteProgram(prog)
    if err != nil { t.Fatal(err) }
}

func TestEngine_RegexModuleImport(t *testing.T) {
    prog, err := interpreter.ParseProgram([]byte(`
import "regex"
result = regex.match("hello123", "\\d+")
`))
    if err != nil { t.Fatal(err) }
    var buf strings.Builder
    config := interpreter.TestConfig()
    config.Stdout = &buf
    e := New(config)
    _, err = e.ExecuteProgram(prog)
    if err != nil { t.Fatal(err) }
}

func TestEngine_AllModulesImportable(t *testing.T) {
    modules := []string{"regex", "datetime", "json", "fs", "http", "waf", "cdc", "fact", "database"}
    for _, mod := range modules {
        prog, err := interpreter.ParseProgram([]byte(`import "` + mod + `"`))
        if err != nil { t.Fatalf("parse %s: %v", mod, err) }
        config := interpreter.TestConfig()
        e := New(config)
        _, err = e.ExecuteProgram(prog)
        if err != nil { t.Fatalf("import %s: %v", mod, err) }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test . -run "TestEngine_.*ModuleImport|TestEngine_AllModules" -v
```
Expected: FAIL — modules not registered (no blank imports yet).

- [ ] **Step 3: Add blank imports to uddin.go**

In `uddin.go`, add to the import block:

```go
import (
    // ... existing imports ...

    // stdlib modules — blank imports trigger init() registration
    _ "github.com/bonkzero404/uddin-lang/stdlib/cdc"
    _ "github.com/bonkzero404/uddin-lang/stdlib/database"
    _ "github.com/bonkzero404/uddin-lang/stdlib/datetime"
    _ "github.com/bonkzero404/uddin-lang/stdlib/fact"
    _ "github.com/bonkzero404/uddin-lang/stdlib/fs"
    _ "github.com/bonkzero404/uddin-lang/stdlib/http"
    _ "github.com/bonkzero404/uddin-lang/stdlib/json"
    _ "github.com/bonkzero404/uddin-lang/stdlib/regex"
    _ "github.com/bonkzero404/uddin-lang/stdlib/waf"
)
```

Also add blank imports to `cmd/uddin-lang/main.go` if it doesn't import `uddin.go` transitively:

```bash
grep -n "import" cmd/uddin-lang/main.go | head -5
```

If main.go imports `uddin` package, the blank imports in `uddin.go` are sufficient. Otherwise add them to main.go as well.

- [ ] **Step 4: Run full test suite**

```bash
go test ./... -v 2>&1 | grep -E "FAIL|ok"
go build ./...
```
Expected: all PASS, binary builds.

- [ ] **Step 5: Verify CLI works**

```bash
make build
echo 'import "regex"
result = regex.match("hello123", "\\d+")
print(result)' > /tmp/test_modules.din
./uddinlang /tmp/test_modules.din
```
Expected output: `true`

- [ ] **Step 6: Commit**

```bash
git add uddin.go cmd/uddin-lang/main.go uddin_test.go
git commit -m "feat: wire stdlib modules via blank imports; all 9 modules importable"
```

---

## Task 15: Write docs/docs/reference/modules.md

**Files:**
- Create: `docs/docs/reference/modules.md`
- Modify: `docs/sidebars.js`
- Modify: `docs/docs/reference/index.md`
- Modify: `docs/docs/reference/builtin-functions.md` (add note about moved functions)

- [ ] **Step 1: Create modules.md**

```markdown
---
id: modules
title: Standard Library Modules
sidebar_label: Modules
---

# Standard Library Modules

Uddin-Lang provides domain-specific functionality as opt-in modules.
Import a module before using its functions:

```din
import "database"
import "database" as db   // custom alias
```

Without an import, module functions are unavailable.

## Import Syntax

| Syntax | Effect |
|--------|--------|
| `import "database"` | Available as `database.connect()`, `database.query()` |
| `import "database" as db` | Available as `db.connect()`, `db.query()` |

---

## database

Relational database connectivity (MySQL, PostgreSQL).

```din
import "database"
conn = database.connect("mysql", "user:pass@tcp(host:3306)/db", 10, 5, "10m", "5m")
rows = database.query(conn, "SELECT id, name FROM users WHERE active = ?", true)
database.close(conn)
```

| Function | Description |
|----------|-------------|
| `database.connect(driver, dsn, maxOpen, maxIdle, lifetime, idleTime)` | Open connection pool |
| `database.query(conn, sql, args...)` | Execute SELECT, return rows |
| `database.execute(conn, sql, args...)` | Execute INSERT/UPDATE/DELETE |
| `database.connect_pool(driver, dsn, ...)` | Connect with explicit pool config |
| `database.configure_pool(conn, maxOpen, maxIdle, lifetime)` | Reconfigure pool |
| `database.execute_batch(conn, operations)` | Execute multiple statements |
| `database.execute_async(conn, sql, args...)` | Non-blocking execution |
| `database.async_status(opId)` | Check async operation status |
| `database.cancel_async(opId)` | Cancel async operation |
| `database.list_async()` | List pending async operations |
| `database.cleanup_async()` | Remove completed async operations |
| `database.stream_tables(conn, tables, handler)` | CDC streaming via binlog |
| `database.stop_stream(streamId)` | Stop a running stream |
| `database.close(conn)` | Close connection pool |

---

## waf

Web Application Firewall request inspection. Requires `_waf_ctx` to be injected by the host.

```din
import "waf"
ip = waf.header("X-Forwarded-For")
if waf.cidr_match(ip, "10.0.0.0/8"):
    waf.return("BLOCK")
end
```

| Function | Description |
|----------|-------------|
| `waf.header(name)` | Get HTTP request header value |
| `waf.query(name)` | Get query parameter value |
| `waf.body_contains(needle)` | Check if request body contains string |
| `waf.cidr_match(ip, cidr)` | Check if IP is within CIDR range |
| `waf.path_match(pattern)` | Glob-match request path |
| `waf.score()` | Get current threat score |
| `waf.action()` | Get current action (ALLOW/LOG/BLOCK) |
| `waf.detected(category)` | Check if threat category detected |
| `waf.detected_any()` | Check if any threat detected |
| `waf.detected_list()` | Get list of detected categories |
| `waf.return(action)` | Set action and exit script (ALLOW/LOG/BLOCK) |

---

## http

HTTP client, HTTP server, TCP and UDP networking.

```din
import "http"
resp = http.get("https://api.example.com/data")
```

| Function | Description |
|----------|-------------|
| `http.get(url)` | HTTP GET request |
| `http.post(url, body, contentType)` | HTTP POST request |
| `http.put(url, body, contentType)` | HTTP PUT request |
| `http.delete(url)` | HTTP DELETE request |
| `http.request(method, url, body, headers)` | Custom HTTP request |
| `http.server_start(host, port)` | Start HTTP server |
| `http.server_stop(serverId)` | Stop HTTP server |
| `http.server_route(serverId, method, path, handler)` | Register route handler |
| `http.response(status, body, headers)` | Build HTTP response object |
| `http.tcp_connect(host, port)` | Open TCP connection |
| `http.tcp_listen(host, port)` | Listen for TCP connections |
| `http.tcp_accept(listener)` | Accept incoming TCP connection |
| `http.tcp_read(conn, bufSize)` | Read from TCP connection |
| `http.tcp_write(conn, data)` | Write to TCP connection |
| `http.tcp_close(conn)` | Close TCP connection |
| `http.udp_connect(host, port)` | Open UDP connection |
| `http.udp_listen(host, port)` | Listen for UDP datagrams |
| `http.udp_read(conn, bufSize)` | Read UDP datagram |
| `http.udp_write(conn, data)` | Write UDP datagram |
| `http.udp_close(conn)` | Close UDP connection |
| `http.net_resolve(host)` | Resolve hostname to IPs |
| `http.net_ping(host, port, timeoutMs)` | TCP ping check |

---

## cdc

Complex Event Processing / Change Data Capture event streaming.

```din
import "cdc"
cdc.emit("user_login", {"user": "alice"})
count = cdc.count("user_login")
recent = cdc.get_window("5m", "user_login")
```

| Function | Description |
|----------|-------------|
| `cdc.emit(eventType, data)` | Emit an event to the store |
| `cdc.define_pattern(name, sequence, timeWindow?)` | Define event sequence pattern |
| `cdc.get_window(duration, eventType?)` | Get events within time window |
| `cdc.clear(eventType?)` | Clear events (all or by type) |
| `cdc.count(eventType?)` | Count events (all or by type) |

---

## fs

Filesystem operations.

```din
import "fs"
content = fs.read("/etc/hosts")
fs.write("/tmp/output.txt", "hello")
files = fs.list("/tmp")
```

| Function | Description |
|----------|-------------|
| `fs.read(path)` | Read file contents as string |
| `fs.write(path, content)` | Write string to file |
| `fs.exists(path)` | Check if path exists |
| `fs.size(path)` | Get file size in bytes |
| `fs.modified(path)` | Get last modified timestamp |
| `fs.permissions(path)` | Get file permissions string |
| `fs.mkdir(path, recursive?)` | Create directory |
| `fs.rmdir(path)` | Remove directory |
| `fs.list(path)` | List directory contents |
| `fs.copy(src, dst)` | Copy file |
| `fs.move(src, dst)` | Move/rename file |
| `fs.delete(path)` | Delete file |
| `fs.path_join(parts...)` | Join path components |
| `fs.dirname(path)` | Get parent directory |
| `fs.basename(path)` | Get filename component |
| `fs.ext(path)` | Get file extension |
| `fs.cwd()` | Get current working directory |
| `fs.chdir(path)` | Change current directory |

---

## json

JSON and XML serialization.

```din
import "json"
obj = json.parse("{\"name\": \"alice\"}")
text = json.stringify(obj)
```

| Function | Description |
|----------|-------------|
| `json.parse(jsonString)` | Parse JSON string to object |
| `json.stringify(value)` | Serialize value to JSON string |
| `json.xml_parse(xmlString)` | Parse XML string to object |
| `json.xml_stringify(value)` | Serialize value to XML string |

---

## fact

In-memory knowledge base / fact database.

```din
import "fact"
fact.assert("customer", "alice", {"tier": "premium"})
exists = fact.exists("customer", "alice")
records = fact.query("customer")
```

| Function | Description |
|----------|-------------|
| `fact.assert(category, key, value?)` | Add fact to knowledge base |
| `fact.retract(category, key)` | Remove fact |
| `fact.query(category, key?)` | Query facts by category or key |
| `fact.exists(category, key)` | Check if fact exists |
| `fact.count()` | Count total facts |
| `fact.clear()` | Clear all facts |
| `fact.get_all()` | Get all facts as object |

---

## regex

Regular expression operations using RE2 engine (no ReDoS risk).

```din
import "regex"
matched = regex.match("hello123", "\\d+")
result = regex.replace("hello world", "world", "din")
parts = regex.split("a,b,c", ",")
```

| Function | Description |
|----------|-------------|
| `regex.match(text, pattern)` | Test if text matches pattern |
| `regex.is_match(pattern, text)` | Test if text matches pattern (args reversed) |
| `regex.find(text, pattern)` | Find first match |
| `regex.find_all(text, pattern, limit?)` | Find all matches |
| `regex.replace(text, pattern, replacement)` | Replace matches |
| `regex.split(text, pattern, limit?)` | Split by pattern |

---

## datetime

Date and time operations.

```din
import "datetime"
now = datetime.now()
formatted = datetime.format(now, "2006-01-02")
future = datetime.add(now, "24h")
```

| Function | Description |
|----------|-------------|
| `datetime.now()` | Current datetime as RFC3339 string |
| `datetime.time_now()` | Current Unix timestamp (milliseconds) |
| `datetime.sleep(ms)` | Sleep for N milliseconds |
| `datetime.format(dateStr, layout)` | Format datetime string |
| `datetime.parse(dateStr, layout)` | Parse string to datetime |
| `datetime.format_new(layout)` | Format current time with layout |
| `datetime.add(dateStr, duration)` | Add duration to datetime |
| `datetime.subtract(dateStr, duration)` | Subtract duration from datetime |
| `datetime.diff(dateStr1, dateStr2)` | Get difference between datetimes |
| `datetime.between(dateStr, start, end)` | Check if date is in range |
| `datetime.compare(dateStr1, dateStr2)` | Compare two datetimes (-1/0/1) |
```

- [ ] **Step 2: Add to sidebar in `docs/sidebars.js`**

Add `'reference/modules'` to `referenceSidebar` after `reference/builtin-functions`:

```js
referenceSidebar: [
    'reference/index',
    'reference/syntax',
    'reference/builtin-functions',
    'reference/modules',           // ← add this
    'reference/database-streaming',
    'reference/memory-optimization',
    'reference/bytecode-vm',
    'reference/library',
],
```

- [ ] **Step 3: Add link to reference/index.md**

Add after the Built-in Functions section:

```markdown
### [Standard Library Modules](./modules)
Domain-specific functionality available via explicit import including:
- `database` — MySQL/PostgreSQL connectivity
- `waf` — Web Application Firewall request inspection
- `http` — HTTP client/server, TCP, UDP
- `cdc` — Complex event processing
- `fs` — Filesystem operations
- `json` — JSON/XML serialization
- `fact` — In-memory knowledge base
- `regex` — Regular expression operations
- `datetime` — Date and time utilities
```

- [ ] **Step 4: Update reference/builtin-functions.md**

Add a note at the top of the "Built-in Functions" page:

```markdown
> **Note:** The following function groups have moved to stdlib modules and require
> `import "module"` before use: database, WAF, HTTP, CDC, filesystem, JSON,
> fact database, regex, and datetime. See [Standard Library Modules](./modules).
```

- [ ] **Step 5: Commit**

```bash
git add docs/docs/reference/modules.md docs/sidebars.js \
        docs/docs/reference/index.md docs/docs/reference/builtin-functions.md
git commit -m "docs: add modules reference page; update sidebar and builtin-functions note"
```

---

## Task 16: Version bump + release

**Files:**
- `internal/version/version.go` (no change needed — version set at build time via ldflags)
- Makefile `tag-release` target

- [ ] **Step 1: Run full test suite one last time**

```bash
go test ./... -count=1
```
Expected: all PASS.

- [ ] **Step 2: Build binary**

```bash
make build
./uddinlang --version
```
Expected: binary builds, shows version.

- [ ] **Step 3: Create PR from phase-3-module-system to main**

```bash
gh pr create --title "feat: stdlib module system — 9 modules, import syntax (v1.2.0)" \
  --body "..."
```

- [ ] **Step 4: After PR merged to main — tag v1.2.0**

```bash
git checkout main && git pull
make tag-release VERSION=v1.2.0
```

- [ ] **Step 5: Deploy docs to gh-pages**

```bash
cd docs && npm run build
# Deploy via gh-pages action or:
# git subtree push --prefix docs/build origin gh-pages
```

Check: `https://bonkzero404.github.io/uddin-lang` shows updated docs.

- [ ] **Step 6: Create GitHub Release**

```bash
gh release create v1.2.0 \
  --title "v1.2.0 — stdlib module system" \
  --notes "$(cat <<'EOF'
## v1.2.0 — stdlib module system

### Breaking changes

All domain-specific builtins now require explicit `import`:

- `import "database"` → `database.query()`, `database.connect()`, ...
- `import "waf"` → `waf.header()`, `waf.cidr_match()`, ...
- `import "http"` → `http.get()`, `http.post()`, ...
- `import "cdc"` → `cdc.emit()`, `cdc.count()`, ...
- `import "fs"` → `fs.read()`, `fs.write()`, ...
- `import "json"` → `json.parse()`, `json.stringify()`, ...
- `import "fact"` → `fact.assert()`, `fact.query()`, ...
- `import "regex"` → `regex.match()`, `regex.find()`, ...
- `import "datetime"` → `datetime.now()`, `datetime.format()`, ...

Flat global names (`db_query`, `waf_header`, `event_emit`, etc.) are removed.

### New features

- `import "module"` syntax
- `import "module" as alias` syntax
- Module reference documentation at https://bonkzero404.github.io/uddin-lang/reference/modules

### Also in this release

- Bytecode VM (Phase 1) — register-based VM, 5–15× speedup vs tree-walker
- Engine program cache + VM pool (Phase 2A)
- OP_CALL_BUILTIN_DIRECT fast-path for 13 core builtins (Phase 2B)
EOF
)"
```

---

## Self-Review

**Spec coverage check:**
- ✅ 9 stdlib modules: Tasks 4–12
- ✅ `import "name"` syntax: Task 2
- ✅ `import "name" as alias` syntax: Task 2
- ✅ `ModuleContext` interface: Task 1
- ✅ Remove flat global functions: Task 13
- ✅ Update `vm_builtins.go` (remove waf direct): Task 13
- ✅ Update `builtin_metadata.go`: Task 13
- ✅ Module reference docs: Task 15
- ✅ Library reference update: Task 15
- ✅ GitHub Pages deploy: Task 16
- ✅ Version release: Task 16
- ✅ Error messages (unknown module): Task 3

**Type consistency check:**
- `ModuleContext` interface defined in Task 1, used in Tasks 4–12 ✅
- `CDCAccessor` interface defined in Task 1, used in Task 10 ✅
- `FactAccessor` interface defined in Task 1, used in Task 11 ✅
- `interpreter.TypeErrorf` defined in Task 1 (`module.go`), used in Tasks 4–12 ✅
- `interpreter.RuntimeErrorf` defined in Task 1, available for Tasks 4–12 ✅
- `Import.Path` (not `Filename`) used consistently in Tasks 2–3 ✅
