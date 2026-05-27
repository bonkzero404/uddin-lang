# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # build ./uddinlang binary
make test           # go test ./... -v
make install        # install to GOPATH/bin
make clean          # remove ./uddinlang binary
make mod-tidy       # go mod tidy

# Single test
go test ./interpreter/... -run TestName -v

# Run a script
./uddinlang <file.din>
./uddinlang --profile <file.din>      # with perf profiling output
./uddinlang --analyze <file.din>      # syntax check only, no execution
./uddinlang --to_json <file.din>      # emit JSON AST to stdout
./uddinlang --from_json <file.json>   # reconstruct .din source from JSON AST
./uddinlang --memory-optimize-stable <file.din>
./uddinlang --memory-optimize-experimental <file.din>  # NOT safe for concurrent code
```

Version metadata is injected at build time via `-ldflags` in the Makefile (`internal/version/`).

## Architecture

### Two public surfaces

| Path | Purpose |
|------|---------|
| `uddin.go` | Go library API — `Engine` struct + package-level convenience functions. Re-exports types from `interpreter/`. |
| `cmd/uddin-lang/main.go` | CLI binary (`uddinlang`). Thin wrapper over `interpreter.RunProgramWithOptions`. |

### `interpreter/` — all language implementation

**Pipeline**: source bytes → `Tokenizer` → `parser` (produces AST with pooled `ASTNodePool`) → `interpreter` (walks AST, calls `Evaluator` for expressions).

Key files:

| File | Responsibility |
|------|---------------|
| `tokenizer.go` | Lexer; `Token` enum starts at ASCII values for single-char, 300+ for multi-char, 500+ for keywords |
| `parser.go` | Recursive-descent parser; `Error` type carries `Position` |
| `ast.go` | AST node types / `Expression` interface |
| `interpreter.go` | `interpreter` struct, statement execution; `Value = any` |
| `evaluator.go` | Expression evaluation; `binaryEvalFuncs` dispatch table; type helpers (`asInt`, `asString`, …) |
| `program.go` | `Program` struct, error display with source pointer |
| `environment.go` | Variable scope stack; `CompactEnvironment` optimization |
| `config.go` | `Config` and `Stats` structs; config presets (see below) |
| `functions.go` | User-defined function calls |
| `builtin_dispatch.go` | Global `BuiltinFunctionDispatcher`; all built-ins register here |
| `expression_optimizer.go` | `ConstantFolder` — constant-folding pass |
| `pool_manager.go` | `UnifiedPoolManager` / `BoundedPool` — shared pools for arrays, maps, string builders |
| `memory_layout.go` | `SafeTaggedValue`, tagged-union memory layout |
| `concurrent_executor.go` | Worker pool for parallel language-level operations |
| `memo.go` | Memoization cache (`ProductionMemoCache` / `OptimizedMemoCache`) |

### Built-in function files

All call `globalBuiltinDispatcher.RegisterBuiltinFunction` at init time:

`fn_core`, `fn_string`, `fn_math`, `fn_array_collection`, `fn_datastructure`, `fn_type_convertion`, `fn_datetime`, `fn_serialization`, `fn_regex`, `fn_network`, `fn_database`, `fn_file_system`, `fn_memory`, `fn_cep`, `fn_fact_database`, `fn_wafio`

WAFio built-ins (`fn_wafio.go`) read request context from a `_waf_ctx map[string]any` variable injected into the interpreter scope.

Regex uses `coregex` (RE2/O(n)) — no ReDoS risk.

### Config presets

| Preset | When to use |
|--------|-------------|
| `DefaultConfig()` | Normal execution |
| `TestConfig()` | Tests — `IsUnitTest: true`, stdout discarded, exit is no-op |
| `StableConfig()` | Production + stable memory/memoization optimizations |
| `ExperimentalConfig()` | Benchmarking only — **not concurrent-safe** |

`IsUnitTest: true` prevents `main()` from auto-executing; required in all test harnesses.

### Return-statement mechanism

`return` is implemented via `panic(returnResult{value, pos})` caught by the function call handler in `functions.go`. Do not confuse with actual panics.

### Script file format

Source files use `.din` extension. Programs can define a `main()` function that auto-executes unless `IsUnitTest` is set.
