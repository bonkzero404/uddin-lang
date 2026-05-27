# Module System Design Spec

## Goal

Introduce a stdlib module system for uddin-lang. Domain-specific builtin functions (database, WAF, HTTP, CDC, filesystem, JSON, regex, datetime, fact database) move from always-on global builtins into opt-in namespaced modules. Users import modules explicitly before use.

## Design Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Namespace style | `database.query()` | Consistent with Python/Lua, readable |
| Import syntax | `import "database"` / `import "database" as db` | Familiar, aliasable |
| Default alias | Module name (`"database"` → `database`) | Predictable |
| Backward compat | Breaking — flat globals removed | Language not in production; clean now |
| Directory | `stdlib/` as separate Go packages | Clean separation, testable per module |
| Module API | `ModuleContext` public interface | Avoids exporting `*interpreter` |
| Module registration | `init()` + blank import in `uddin.go` | Standard Go pattern |

## Scope

**In scope:**
- 9 stdlib modules: `database`, `waf`, `http`, `cdc`, `fs`, `json`, `fact`, `regex`, `datetime`
- `import "name"` and `import "name" as alias` syntax
- `ModuleContext` interface for module ↔ interpreter communication
- Remove all flat global functions that move to modules
- Update `vm_builtins.go` direct table (remove waf entries)
- Update `builtin_metadata.go` (remove moved functions)
- Documentation: module reference page, bytecode/VM internals page, update library reference
- GitHub Release with new version tag
- Deploy docs to GitHub Pages

**Not in scope:**
- File module imports (`import "./file.din"`) — deferred
- Dynamic plugin loading (`.so` files) — not needed
- Optional type annotations — deferred
- JIT compiler — deferred

## Architecture

```
.din source
  import "database"
  import "waf" as w
  database.query("SELECT ...")
  w.header("X-Forwarded-For")
         │
         ▼ parse
  ImportStatement{Path:"database", Alias:"", IsModule:true}
  ImportStatement{Path:"waf", Alias:"w", IsModule:true}
         │
         ▼ execute
  interpreter: lookup moduleRegistry["database"]
               inject scope: database = map[string]Value{...fns}
               inject scope: w = map[string]Value{...fns}
         │
         ▼
  stdlib/database/database.go  (package stdlib_database)
  stdlib/waf/waf.go            (package stdlib_waf)
  ...
```

## Section 1: Core Interfaces

**File: `interpreter/module.go`** (new file)

```go
package interpreter

import "io"

// ModuleContext is the public API that stdlib modules use to interact with
// the interpreter. Replaces direct *interpreter access from external packages.
type ModuleContext interface {
    LookupVar(name string) (Value, bool)
    SetVar(name string, val Value)
    Stdout() io.Writer
    // GetContext retrieves a host-injected context map (e.g. "_waf_ctx").
    GetContext(key string) (map[string]any, bool)
}

// ModuleFunc is the function signature for all stdlib module functions.
type ModuleFunc func(ctx ModuleContext, pos Position, args []Value) Value

// Module is implemented by each stdlib package.
type Module interface {
    Name() string                       // "database", "waf", etc.
    Functions() map[string]ModuleFunc   // function name → implementation
}

// RegisterModule adds m to the global module registry.
// Called from each stdlib package's init().
func RegisterModule(m Module)

// LookupModule returns the module registered under name, or false.
func LookupModule(name string) (Module, bool)
```

**File: `interpreter/module_registry.go`** (new file)

```go
package interpreter

var globalModuleRegistry = map[string]Module{}

func RegisterModule(m Module) {
    globalModuleRegistry[m.Name()] = m
}

func LookupModule(name string) (Module, bool) {
    m, ok := globalModuleRegistry[name]
    return m, ok
}

// buildNamespaceObject converts a Module into a map[string]Value
// where each value is a builtinFunction wrapping the ModuleFunc.
func buildNamespaceObject(mod Module, interp *interpreter) map[string]Value {
    ctx := &interpreterContext{interp}
    ns := make(map[string]Value, len(mod.Functions()))
    for fname, fn := range mod.Functions() {
        fn, modName, fname := fn, mod.Name(), fname
        ns[fname] = builtinFunction{
            Function: func(interp *interpreter, pos Position, args []Value) Value {
                ctx.interp = interp
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
```

## Section 2: Parser Changes

**`interpreter/tokenizer.go`** — add `AS` keyword token:
```go
AS = 522
// in keywordTokens map:
"as": AS,
// in tokenName map:
AS: "as",
```

**`interpreter/ast.go`** — extend `Import` struct:
```go
type Import struct {
    Pos      Position
    Path     string  // "database" or "file.din"
    Alias    string  // "" = use module name as alias
    IsModule bool    // true = stdlib module, false = .din file import
}
```

**`interpreter/parser.go`** — extend `import_()`:
```go
func (p *parser) import_() Statement {
    pos := p.pos
    p.expect(IMPORT)
    if p.tok != STR {
        p.error("import requires a string, got %s", p.tok)
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

    // Module: no path separator, no .din suffix
    isModule := !strings.Contains(path, "/") &&
                !strings.Contains(path, "\\") &&
                !strings.HasSuffix(path, ".din")

    return &Import{pos, path, alias, isModule}
}
```

**`interpreter/interpreter.go`** — handle `IsModule` in `executeImport`:
```go
func (interp *interpreter) executeImport(stmt *Import) error {
    if stmt.IsModule {
        mod, ok := LookupModule(stmt.Path)
        if !ok {
            known := knownModuleNames() // sorted list from globalModuleRegistry
            return runtimeError(stmt.Pos,
                "unknown module %q\n  available: %s", stmt.Path, strings.Join(known, ", "))
        }
        alias := stmt.Alias
        if alias == "" {
            alias = stmt.Path
        }
        ns := buildNamespaceObject(mod, interp)
        interp.setVar(alias, ns)
        return nil
    }
    // existing file import logic unchanged
    return interp.executeFileImport(stmt.Path)
}
```

## Section 3: stdlib Directory Structure

```
stdlib/
  database/
    database.go    package stdlib_database
  waf/
    waf.go         package stdlib_waf
  http/
    http.go        package stdlib_http
  cdc/
    cdc.go         package stdlib_cdc
  fs/
    fs.go          package stdlib_fs
  json/
    json.go        package stdlib_json
  fact/
    fact.go        package stdlib_fact
  regex/
    regex.go       package stdlib_regex
  datetime/
    datetime.go    package stdlib_datetime
```

Each module file follows this pattern:
```go
package stdlib_waf

import "github.com/bonkzero404/uddin-lang/interpreter"

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
```

## Section 4: Function Name Mapping

All flat global functions below are **removed**. Equivalent in namespace:

### database
| Removed | Namespace |
|---------|-----------|
| `db_connect` | `database.connect` |
| `db_query` | `database.query` |
| `db_execute` | `database.execute` |
| `db_connect_with_pool` | `database.connect_pool` |
| `db_configure_pool` | `database.configure_pool` |
| `db_execute_batch` | `database.execute_batch` |
| `db_execute_async` | `database.execute_async` |
| `db_get_async_status` | `database.async_status` |
| `db_cancel_async` | `database.cancel_async` |
| `db_list_async_operations` | `database.list_async` |
| `db_cleanup_async_operations` | `database.cleanup_async` |
| `stream_tables` | `database.stream_tables` |
| `db_stop_stream` | `database.stop_stream` |
| `db_close` | `database.close` |

### waf
| Removed | Namespace |
|---------|-----------|
| `waf_header` | `waf.header` |
| `waf_query` | `waf.query` |
| `waf_body_contains` | `waf.body_contains` |
| `waf_cidr_match` | `waf.cidr_match` |
| `waf_path_match` | `waf.path_match` |
| `waf_score` | `waf.score` |
| `waf_action` | `waf.action` |
| `waf_detected` | `waf.detected` |
| `waf_detected_any` | `waf.detected_any` |
| `waf_detected_list` | `waf.detected_list` |
| `waf_return` | `waf.return` |

### http
| Removed | Namespace |
|---------|-----------|
| `http_get` | `http.get` |
| `http_post` | `http.post` |
| `http_put` | `http.put` |
| `http_delete` | `http.delete` |
| `http_request` | `http.request` |
| `http_server_start` | `http.server_start` |
| `http_server_stop` | `http.server_stop` |
| `http_server_route` | `http.server_route` |
| `http_response` | `http.response` |
| `tcp_connect` | `http.tcp_connect` |
| `tcp_listen` | `http.tcp_listen` |
| `tcp_accept` | `http.tcp_accept` |
| `tcp_read` | `http.tcp_read` |
| `tcp_write` | `http.tcp_write` |
| `tcp_close` | `http.tcp_close` |
| `udp_connect` | `http.udp_connect` |
| `udp_listen` | `http.udp_listen` |
| `udp_read` | `http.udp_read` |
| `udp_write` | `http.udp_write` |
| `udp_close` | `http.udp_close` |
| `net_resolve` | `http.net_resolve` |
| `net_ping` | `http.net_ping` |

### cdc (Complex Event / CDC)
| Removed | Namespace |
|---------|-----------|
| `event_emit` | `cdc.emit` |
| `event_define_pattern` | `cdc.define_pattern` |
| `event_get_window` | `cdc.get_window` |
| `event_clear` | `cdc.clear` |
| `event_count` | `cdc.count` |

### fs
| Removed | Namespace |
|---------|-----------|
| `read_file` | `fs.read` |
| `write_file` | `fs.write` |
| `file_exists` | `fs.exists` |
| `file_size` | `fs.size` |
| `file_modified` | `fs.modified` |
| `file_permissions` | `fs.permissions` |
| `mkdir` | `fs.mkdir` |
| `rmdir` | `fs.rmdir` |
| `list_dir` | `fs.list` |
| `copy_file` | `fs.copy` |
| `move_file` | `fs.move` |
| `delete_file` | `fs.delete` |
| `path_join` | `fs.path_join` |
| `path_dirname` | `fs.dirname` |
| `path_basename` | `fs.basename` |
| `path_ext` | `fs.ext` |
| `getcwd` | `fs.cwd` |
| `chdir` | `fs.chdir` |

### json
| Removed | Namespace |
|---------|-----------|
| `json_parse` | `json.parse` |
| `json_stringify` | `json.stringify` |
| `xml_parse` | `json.xml_parse` |
| `xml_stringify` | `json.xml_stringify` |

### fact
| Removed | Namespace |
|---------|-----------|
| `fact_assert` | `fact.assert` |
| `fact_retract` | `fact.retract` |
| `fact_query` | `fact.query` |
| `fact_exists` | `fact.exists` |
| `fact_count` | `fact.count` |
| `fact_clear` | `fact.clear` |
| `fact_get_all` | `fact.get_all` |

### regex
| Removed | Namespace |
|---------|-----------|
| `is_regex_match` | `regex.is_match` |
| `regex_match` | `regex.match` |
| `regex_find` | `regex.find` |
| `regex_find_all` | `regex.find_all` |
| `regex_replace` | `regex.replace` |
| `regex_split` | `regex.split` |

### datetime
| Removed | Namespace |
|---------|-----------|
| `date_now` | `datetime.now` |
| `time_now` | `datetime.time_now` |
| `sleep` | `datetime.sleep` |
| `date_format` | `datetime.format` |
| `date_parse` | `datetime.parse` |
| `date_format_new` | `datetime.format_new` |
| `date_add` | `datetime.add` |
| `date_subtract` | `datetime.subtract` |
| `date_diff` | `datetime.diff` |
| `date_between` | `datetime.between` |
| `date_compare` | `datetime.compare` |

## Section 5: Files to Delete / Modify

**Deleted:**
- `interpreter/fn_database.go` → replaced by `stdlib/database/database.go`
- `interpreter/fn_wafio.go` → replaced by `stdlib/waf/waf.go`
- `interpreter/fn_network.go` → replaced by `stdlib/http/http.go`
- `interpreter/fn_cep.go` → replaced by `stdlib/cdc/cdc.go`
- `interpreter/fn_file_system.go` → replaced by `stdlib/fs/fs.go`
- `interpreter/fn_serialization.go` → replaced by `stdlib/json/json.go`
- `interpreter/fn_fact_database.go` → replaced by `stdlib/fact/fact.go`
- `interpreter/fn_regex.go` → replaced by `stdlib/regex/regex.go`
- `interpreter/fn_datetime.go` → replaced by `stdlib/datetime/datetime.go`

**Modified:**
- `interpreter/ast.go` — extend `Import` struct
- `interpreter/tokenizer.go` — add `AS = 522`
- `interpreter/parser.go` — extend `import_()`
- `interpreter/interpreter.go` — add module dispatch in import execution
- `interpreter/builtin_metadata.go` — remove all moved function entries
- `interpreter/vm_builtins.go` — remove `directWafCidrMatch`, `directWafPathMatch`
- `interpreter/compiler.go` — remove waf-specific builtin optimizations
- `interpreter/module.go` — new file (interfaces)
- `interpreter/module_registry.go` — new file (registry + buildNamespaceObject)
- `uddin.go` — blank import all 9 stdlib packages

**New:**
- `stdlib/database/database.go`
- `stdlib/waf/waf.go`
- `stdlib/http/http.go`
- `stdlib/cdc/cdc.go`
- `stdlib/fs/fs.go`
- `stdlib/json/json.go`
- `stdlib/fact/fact.go`
- `stdlib/regex/regex.go`
- `stdlib/datetime/datetime.go`

## Section 6: Error Handling

```
// Module not registered:
RuntimeError at line N: unknown module "xyz"
  available modules: cdc, database, datetime, fact, fs, http, json, regex, waf

// Old flat function called without import:
RuntimeError at line N: undefined variable "db_query"
  hint: use 'import "database"' then call database.query(...)
```

The hint for undefined flat function names is implemented by checking
if the name matches a known moved function in the error handler.

## Section 7: Docs + Release Plan

### Docs changes (5 items from requirements)

1. **Module reference** — new `docs/docs/reference/modules.md`
   - Import syntax, alias syntax
   - Table of all 9 modules with function lists
   - Usage examples per module

2. **Bytecode/VM internals** — new `docs/docs/reference/internals/bytecode.md`
   - VM architecture (register-based, frame stack)
   - Opcode table with descriptions
   - Phase 2A (bytecode cache, Engine pool) and Phase 2B (OP_CALL_BUILTIN_DIRECT) changes

3. **Library reference update** — `docs/docs/reference/library.md`
   - Remove all flat function entries that moved to modules
   - Add import examples for each module group

4. **GitHub Pages deploy** — push to `gh-pages` branch
   - Site at `https://bonkzero404.github.io/uddin-lang`

5. **Version release** — new GitHub Release
   - Bump version in `internal/version/`
   - Tag `v0.x.0` with changelog

### Release branch

```
main
  └── phase-1-bytecode-vm   (current, pushed)
        └── phase-3-module-system  (new branch for this work)
```

After all tasks pass: merge to main → tag release → deploy docs.

## Usage Examples

```din
// Database
import "database"
conn = database.connect("mysql", "user:pass@tcp(localhost:3306)/mydb", 10, 5, "10m", "5m")
rows = database.query(conn, "SELECT id, name FROM users WHERE active = ?", true)
database.close(conn)

// WAF (script runs in WAF engine context)
import "waf"
ip = waf.header("X-Forwarded-For")
if waf.cidr_match(ip, "10.0.0.0/8"):
    waf.return("BLOCK")
end

// HTTP
import "http"
resp = http.get("https://api.example.com/data")

// CDC / Complex Events
import "cdc"
cdc.emit("user_login", {"user": "john"})
count = cdc.count("user_login")

// Regex
import "regex"
matched = regex.match("hello@example.com", "^[\\w.]+@[\\w.]+$")
parts = regex.split("a,b,c", ",")

// Datetime
import "datetime"
now = datetime.now()
fmt = datetime.format(now, "2006-01-02")

// Alias
import "database" as db
import "http" as net
rows = db.query(conn, "SELECT ...")
resp = net.get("https://...")
```
