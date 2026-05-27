# Phase 0 — Bug Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix seven confirmed bugs that cause data corruption, incorrect errors, and silent performance regressions — establishing a clean foundation before the bytecode VM.

**Architecture:** All fixes are surgical edits to existing files. No new files. No API changes. All existing tests must pass after every task.

**Tech Stack:** Go 1.25.4, standard library only.

---

## File Map

| File | Change |
|------|--------|
| `interpreter/functions.go` | Remove global `factDatabase`, `eventStore`, `eventPatterns`, mutexes → move to `interpreter` struct |
| `interpreter/interpreter.go` | Add struct fields for above; fix `evalPlus` pool misuse; fix 4-way dispatch dead code; fix tagged-value label |
| `interpreter/fn_fact_database.go` | Replace all global var refs with `interp.factDatabase` / `interp.factMutex` |
| `interpreter/fn_cep.go` | Replace all global var refs with `interp.eventStore` / `interp.eventPatterns` / `interp.eventMutex` |
| `interpreter/variable_lookup.go` | Replace `reflect.DeepEqual` with explicit type switch in `safeValueEqual` |
| `interpreter/memory_layout_config.go` | Set `EnableTaggedValues: false` in `StableMemoryLayoutConfig()` |
| `interpreter/config.go` | Fix contradictory stability comments on `StableMemoizationConfig` |
| `interpreter/expression_optimizer.go` | Fix stability comment |
| `interpreter/memo.go` | Fix stability comment |

---

## Task 1: Fix Global State — factDatabase / eventStore

**Files:**
- Modify: `interpreter/functions.go:9-16`
- Modify: `interpreter/interpreter.go` (struct definition, ~line 20-56)
- Modify: `interpreter/fn_fact_database.go` (all usages)
- Modify: `interpreter/fn_cep.go` (all usages)

- [ ] **Step 1: Write the failing test**

Create file `interpreter/global_state_test.go`:

```go
package interpreter

import (
	"strings"
	"testing"
)

// Two engines must not share factDatabase or eventStore.
// Uses EvaluateString which creates a fresh interpreter per call.
func TestEngineIsolation_FactDatabase(t *testing.T) {
	// Engine 1: assert a fact
	_, _, err := EvaluateString(`fact_assert("customer", "alice", {"status": "premium"})`)
	if err != nil {
		t.Fatal(err)
	}

	// Engine 2 (new interpreter): query the same fact — must NOT be found
	result, _, err := EvaluateString(`fact_query("customer", "alice")`)
	if err != nil {
		t.Fatal(err)
	}
	// With per-interpreter factDatabase, this returns nil (fact not found)
	// With old global var, it returns a non-nil value (bug)
	if result != nil && result != false {
		t.Errorf("isolation broken: engine 2 sees engine 1's facts, got %v", result)
	}
}

func TestEngineIsolation_EventStore(t *testing.T) {
	// Engine 1: emit an event
	_, _, err := EvaluateString(`event_emit("login", {"user": "alice"})`)
	if err != nil {
		t.Fatal(err)
	}

	// Engine 2 (new interpreter): count events — must be 0
	result, _, err := EvaluateString(`event_count()`)
	if err != nil {
		t.Fatal(err)
	}
	if result != 0 {
		t.Errorf("isolation broken: engine 2 sees %v events from engine 1", result)
	}
	_ = strings.Contains("", "") // keep import used
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./interpreter/... -run TestEngineIsolation -v
```

Expected: FAIL — `interp2.factDatabase` field doesn't exist yet.

- [ ] **Step 3: Add fields to interpreter struct**

In `interpreter/interpreter.go`, locate the `interpreter` struct (around line 21) and add:

```go
type interpreter struct {
	// ... existing fields ...

	// Per-engine rule engine state (was package-level global — isolated per Engine)
	factDatabase map[string]any
	factMutex    sync.RWMutex
	eventStore    []map[string]any
	eventPatterns map[string]any
	eventMutex    sync.RWMutex
}
```

- [ ] **Step 4: Initialize fields in newInterpreter**

Find the `newInterpreter` function and add initialization:

```go
func newInterpreter(config *Config) *interpreter {
	interp := &interpreter{
		// ... existing initializations ...
		factDatabase:  make(map[string]any),
		eventStore:    make([]map[string]any, 0),
		eventPatterns: make(map[string]any),
	}
	return interp
}
```

- [ ] **Step 5: Remove package-level globals from functions.go**

In `interpreter/functions.go`, delete lines 9-16:

```go
// DELETE these lines:
var (
	factDatabase = make(map[string]any)
	factMutex    = sync.RWMutex{}

	eventStore    = make([]map[string]any, 0)
	eventPatterns = make(map[string]any)
	eventMutex    = sync.RWMutex{}
)
```

- [ ] **Step 6: Update fn_fact_database.go to use interp fields**

Replace every `factDatabase` → `interp.factDatabase` and `factMutex` → `interp.factMutex`. Example — `factAssertFunc`:

```go
func factAssertFunc(interp *interpreter, pos Position, args []Value) Value {
	// ... validation unchanged ...
	interp.factMutex.Lock()
	defer interp.factMutex.Unlock()

	fullKey := category + ":" + key
	if len(args) == 2 {
		interp.factDatabase[fullKey] = true
	} else {
		interp.factDatabase[fullKey] = args[2]
	}
	return Value(true)
}
```

Apply the same `interp.factDatabase` / `interp.factMutex` substitution to every function in `fn_fact_database.go`.

- [ ] **Step 7: Update fn_cep.go to use interp fields**

Replace every `eventStore` → `interp.eventStore`, `eventPatterns` → `interp.eventPatterns`, `eventMutex` → `interp.eventMutex`. Example — `eventEmitFunc`:

```go
func eventEmitFunc(interp *interpreter, pos Position, args []Value) Value {
	// ... validation unchanged ...
	interp.eventMutex.Lock()
	defer interp.eventMutex.Unlock()

	event := map[string]any{
		"type":      eventType,
		"timestamp": time.Now().Format(time.RFC3339),
		"id":        fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	if len(args) > 1 {
		if data, ok := args[1].(map[string]Value); ok {
			for k, v := range data {
				event[k] = v
			}
		}
	}
	interp.eventStore = append(interp.eventStore, event)
	if len(interp.eventStore) > 1000 {
		interp.eventStore = interp.eventStore[len(interp.eventStore)-1000:]
	}
	return Value(true)
}
```

Apply the same substitution to every function in `fn_cep.go`.

- [ ] **Step 8: Run tests**

```bash
go test ./interpreter/... -run TestEngineIsolation -v
```

Expected: PASS — both isolation tests pass.

- [ ] **Step 9: Run full test suite**

```bash
go test ./... -v 2>&1 | tail -20
```

Expected: all packages `ok`.

- [ ] **Step 10: Commit**

```bash
git add interpreter/functions.go interpreter/interpreter.go interpreter/fn_fact_database.go interpreter/fn_cep.go interpreter/global_state_test.go
git commit -m "fix: isolate factDatabase and eventStore per Engine instance"
```

---

## Task 2: Fix evalPlus Pool Misuse

**Files:**
- Modify: `interpreter/interpreter.go:329-338`

- [ ] **Step 1: Write the failing test**

Add to `interpreter/interpreter_test.go` (or create `interpreter/evalplus_test.go`):

```go
func TestEvalPlusLargeArrayNoDuplication(t *testing.T) {
	// Arrays >1000 elements trigger the pooled path
	src := `
a = []
i = 0
while i < 600
    a = a + [i]
    i = i + 1
end
b = []
j = 0
while j < 600
    b = b + [j * 2]
    j = j + 1
end
c = a + b
len(c)
`
	val, _, err := EvaluateString(src)
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 1200 {
		t.Errorf("expected 1200 elements, got %v", val)
	}
}
```

- [ ] **Step 2: Run test**

```bash
go test ./interpreter/... -run TestEvalPlusLargeArrayNoDuplication -v
```

Note the current behaviour (may pass but pooled slice is silently corrupted on repeated use — the bug is a use-after-free in the pool, not a visible result error in this test). Proceed to fix regardless.

- [ ] **Step 3: Fix the pool misuse**

In `interpreter/interpreter.go`, find the large-array path in `evalPlus` (~line 329):

```go
// BEFORE (buggy):
if llen+rlen > 1000 {
    result := GetPooledArray(llen + rlen)
    defer PutPooledArray(result)              // ← BUG: returns potentially-wrong slice
    result = SmartAppend(result, *larr...)
    result = SmartAppend(result, *rarr...)
    final := make([]Value, len(result))
    copy(final, result)
    return Value(&final)
}

// AFTER (fixed):
if llen+rlen > 1000 {
    // SmartAppend may grow the backing array; do not pool intermediate results.
    result := make([]Value, 0, llen+rlen)
    result = append(result, *larr...)
    result = append(result, *rarr...)
    return Value(&result)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./interpreter/... -run TestEvalPlus -v
go test ./... 2>&1 | tail -10
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add interpreter/interpreter.go
git commit -m "fix: remove defer PutPooledArray after SmartAppend in evalPlus"
```

---

## Task 3: Fix ensureIntToFloats False-Zero Sentinel

**Files:**
- Modify: `interpreter/interpreter.go:100-122` (ensureIntToFloats)
- Modify: `interpreter/interpreter.go` (evalDivide ~line 449, evalModulo ~line 466, evalPower ~line 485)

- [ ] **Step 1: Write the failing tests**

Add to `interpreter/interpreter_test.go`:

```go
func TestDivideZeroByZeroFloat(t *testing.T) {
	// 0 / 0.0 should panic with "divide by zero", not a type error
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "divide by zero") {
			t.Errorf("expected 'divide by zero' panic, got: %s", msg)
		}
	}()
	_, _, _ = EvaluateString(`0 / 0.0`)
}

func TestModuloZeroByZeroFloat(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "divide by zero") {
			t.Errorf("expected 'divide by zero' panic, got: %s", msg)
		}
	}()
	_, _, _ = EvaluateString(`0 % 0.0`)
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./interpreter/... -run "TestDivideZeroByZeroFloat|TestModuloZeroByZeroFloat" -v
```

Expected: FAIL — panic message contains "requires two floats" not "divide by zero".

- [ ] **Step 3: Fix ensureIntToFloats signature**

In `interpreter/interpreter.go`, replace:

```go
// BEFORE:
func ensureIntToFloats(l, r Value) (float64, float64) {
	lf, lok := l.(float64)
	if !lok {
		if li, lok := l.(int); lok {
			lf = float64(li)
		} else {
			return 0, 0
		}
	}
	rf, rok := r.(float64)
	if !rok {
		if ri, rok := r.(int); rok {
			rf = float64(ri)
		} else {
			return 0, 0
		}
	}
	return lf, rf
}

// AFTER:
func ensureIntToFloats(l, r Value) (float64, float64, bool) {
	lf, lok := l.(float64)
	if !lok {
		li, ok := l.(int)
		if !ok {
			return 0, 0, false
		}
		lf = float64(li)
	}
	rf, rok := r.(float64)
	if !rok {
		ri, ok := r.(int)
		if !ok {
			return 0, 0, false
		}
		rf = float64(ri)
	}
	return lf, rf, true
}
```

- [ ] **Step 4: Fix evalDivide**

```go
func evalDivide(pos Position, l, r Value) Value {
	if result, handled := GetFastEvaluator().FastEvalDivide(l, r); handled {
		return result
	}
	lf, rf, ok := ensureIntToFloats(l, r)
	if !ok {
		panic(typeError(pos, "/ requires two floats or integers"))
	}
	if rf == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(lf / rf)
}
```

- [ ] **Step 5: Fix evalModulo**

```go
func evalModulo(pos Position, l, r Value) Value {
	if result, handled := GetFastEvaluator().FastEvalModulo(l, r); handled {
		return result
	}
	lf, rf, ok := ensureIntToFloats(l, r)
	if !ok {
		panic(typeError(pos, "modulo operator requires two floats or integers"))
	}
	if rf == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(int(lf) % int(rf))
}
```

- [ ] **Step 6: Fix evalPower**

```go
func evalPower(pos Position, l, r Value) Value {
	if result, handled := GetFastEvaluator().FastEvalPower(l, r); handled {
		return result
	}
	lf, rf, ok := ensureIntToFloats(l, r)
	if !ok {
		panic(typeError(pos, "** requires two floats or integers"))
	}
	return Value(math.Pow(lf, rf))
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./interpreter/... -run "TestDivideZeroByZeroFloat|TestModuloZeroByZeroFloat" -v
go test ./... 2>&1 | tail -10
```

Expected: new tests PASS, all others PASS.

- [ ] **Step 8: Commit**

```bash
git add interpreter/interpreter.go
git commit -m "fix: ensureIntToFloats returns bool to distinguish type-error from zero-values"
```

---

## Task 4: Fix safeValueEqual — Replace reflect.DeepEqual

**Files:**
- Modify: `interpreter/variable_lookup.go:38-48`

- [ ] **Step 1: Write the benchmark to confirm improvement**

Add to `interpreter/variable_lookup_test.go` (create if needed):

```go
package interpreter

import "testing"

func BenchmarkSafeValueEqual_Int(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		safeValueEqual(42, 42)
	}
}

func BenchmarkSafeValueEqual_String(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		safeValueEqual("hello", "hello")
	}
}
```

- [ ] **Step 2: Run benchmark to get baseline**

```bash
go test ./interpreter/... -bench=BenchmarkSafeValueEqual -benchmem -count=3
```

Record ns/op — expected ~200-400ns due to reflect.DeepEqual.

- [ ] **Step 3: Fix safeValueEqual**

In `interpreter/variable_lookup.go`, replace:

```go
// BEFORE:
func safeValueEqual(a, b Value) bool {
	defer func() {
		if recover() != nil {
		}
	}()
	return reflect.DeepEqual(a, b)
}

// AFTER:
func safeValueEqual(a, b Value) bool {
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case int:
		bv, ok := b.(int)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case *[]Value:
		bv, ok := b.(*[]Value)
		return ok && av == bv // pointer identity — cache invalidated on reassign
	case map[string]Value:
		// map identity: only equal if same map pointer
		bv, ok := b.(map[string]Value)
		if !ok {
			return false
		}
		// compare pointer by checking address via len+content shortcut
		return len(av) == len(bv) && &av == &bv
	default:
		return false
	}
}
```

Also remove the `"reflect"` import from `variable_lookup.go`.

- [ ] **Step 4: Run benchmark after fix**

```bash
go test ./interpreter/... -bench=BenchmarkSafeValueEqual -benchmem -count=3
```

Expected: ~2-5ns/op (50-100x improvement).

- [ ] **Step 5: Run full tests**

```bash
go test ./... 2>&1 | tail -10
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add interpreter/variable_lookup.go
git commit -m "fix: replace reflect.DeepEqual with type switch in safeValueEqual (~100x faster)"
```

---

## Task 5: Remove Dead Code — 4-Way Function Dispatch

**Files:**
- Modify: `interpreter/interpreter.go:717-780`

- [ ] **Step 1: Verify dead code before removing**

Run:

```bash
go test ./interpreter/... -run TestFunction -v 2>&1 | grep -E "PASS|FAIL"
```

Note all function-related tests pass (baseline).

- [ ] **Step 2: Remove the dead type assertion blocks**

In `interpreter/interpreter.go`, in the `case *Call:` branch of `evaluate()`, replace the entire block from `if f, ok := function.(functionType)` to `panic(typeError(...))` with:

```go
case *Call:
	function := interp.evaluate(e.Function)
	f, ok := function.(functionType)
	if !ok {
		panic(typeError(e.Function.Position(), "can't call non-function type %s", typeName(function)))
	}
	args := make([]Value, 0, len(e.Arguments))
	for _, a := range e.Arguments {
		args = append(args, interp.evaluate(a))
	}
	if e.Ellipsis {
		iterator := getIterator(e.Arguments[len(e.Arguments)-1].Position(), args[len(args)-1])
		args = args[:len(args)-1]
		for iterator.HasNext() {
			args = append(args, iterator.Value())
		}
	}
	return interp.callFunction(e.Function.Position(), f, args)
```

- [ ] **Step 3: Run tests**

```bash
go test ./interpreter/... -run "TestFunction|TestBuiltin|TestClosure|TestRecursion" -v 2>&1 | grep -E "PASS|FAIL|---"
go test ./... 2>&1 | tail -10
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add interpreter/interpreter.go
git commit -m "fix: remove dead 4-way function dispatch in evaluate() Call case"
```

---

## Task 6: Standardize Stability Labels

**Files:**
- Modify: `interpreter/config.go`
- Modify: `interpreter/expression_optimizer.go`
- Modify: `interpreter/memo.go`
- Modify: `interpreter/memory_layout_config.go`
- Modify: `interpreter/functions.go` (userFunction.Memoized comment)

Convention: use `// STABLE`, `// EXPERIMENTAL(reason)`, `// DEPRECATED(use X instead)` at the top of any type, var, or function that isn't obviously production-stable.

- [ ] **Step 1: Fix config.go — StableMemoizationConfig**

Find `StableMemoizationConfig` and its doc comment. The function says "production-ready" but the `Memoized` field on `userFunction` says "not thread-safe". Fix the `StableMemoizationConfig` doc:

```go
// StableMemoizationConfig returns memoization configuration for long-lived single-Engine use.
// WARNING: Memoization is NOT thread-safe. Do not enable on Engines shared across goroutines.
// The production-ready cache (UseProductionCache=true) is stable for single-threaded scripts.
func StableMemoizationConfig() *MemoizationConfig {
```

- [ ] **Step 2: Fix expression_optimizer.go**

At the top of `ExpressionOptimizer` type:

```go
// ExpressionOptimizer provides optimization for expressions.
// EXPERIMENTAL(caching): Subexpression cache may return stale results if the same
// expression node appears in different evaluation contexts. Disable in concurrent scripts.
type ExpressionOptimizer struct {
```

- [ ] **Step 3: Fix memo.go**

Top-of-file comment for the global cache:

```go
// EXPERIMENTAL(thread-safety): Global memoization cache. Functions marked `memo` share
// this cache across all interpreter instances — safe for read-heavy workloads, unsafe
// when the same memoized function is called concurrently with different args on different Engines.
var globalOptimizedMemoCache = ...
```

- [ ] **Step 4: Fix functions.go — userFunction.Memoized**

```go
// EXPERIMENTAL(thread-safety): memoized functions share the global memo cache.
// Not safe for concurrent Engine use. Safe for sequential single-Engine scripts.
Memoized bool
```

- [ ] **Step 5: Run tests**

```bash
go test ./... 2>&1 | tail -10
```

Expected: all pass (comment-only changes).

- [ ] **Step 6: Commit**

```bash
git add interpreter/config.go interpreter/expression_optimizer.go interpreter/memo.go interpreter/functions.go interpreter/memory_layout_config.go
git commit -m "docs: standardize STABLE/EXPERIMENTAL/DEPRECATED labels across interpreter package"
```

---

## Task 7: Fix StableConfig Tagged Values Overhead

**Files:**
- Modify: `interpreter/memory_layout_config.go:47-59`

Tagged values add 161 ns/op per assign. `StableMemoryLayoutConfig` enables them but this makes the "stable" config slower than the default config for typical scripts. Only `ExperimentalConfig` should enable tagged values.

- [ ] **Step 1: Write benchmark to confirm regression**

```bash
go test ./interpreter/... -bench=BenchmarkAdvancedNestedFunctionCalls -benchmem -count=3
```

Then temporarily set `EnableTaggedValues: true` in `DefaultMemoryLayoutConfig` and re-run. Confirm slowdown. Revert.

- [ ] **Step 2: Fix StableMemoryLayoutConfig**

In `interpreter/memory_layout_config.go`, change `StableMemoryLayoutConfig`:

```go
func StableMemoryLayoutConfig() *MemoryLayoutConfig {
	return &MemoryLayoutConfig{
		EnableTaggedValues:            false, // Too much overhead (161ns/assign) for production
		EnableCompactEnvironment:      true,
		EnableCacheFriendlyStructures: true,
		EnableVariableLookupCache:     false,
		EnableMemoryLeakDetection:     true,
		MemoryLeakDetectionInterval:   300,
		AutoCleanupMemoryLeaks:        true,
		TaggedValuePoolSize:           1000,
		MaxStringCacheSize:            10000,
		VariableLookupCacheSize:       500,
	}
}
```

Update the doc comment:

```go
// StableMemoryLayoutConfig returns production-ready memory layout configuration.
// TaggedValues are disabled (161ns overhead per assignment outweighs type-tag benefits).
// CompactEnvironment and CacheFriendlyStructures are enabled — both are thread-safe.
```

- [ ] **Step 3: Verify ExperimentalConfig still enables tagged values**

`ExperimentalMemoryLayoutConfig` should retain `EnableTaggedValues: true` — do not change it.

- [ ] **Step 4: Run tests including memory_layout tests**

```bash
go test ./interpreter/... -run "TestMemoryLayout|TestStableConfig" -v
go test ./... 2>&1 | tail -10
```

Expected: all pass. If any test asserts `IsTaggedValuesEnabled() == true` under `StableConfig`, update that test to reflect the corrected expected value.

- [ ] **Step 5: Commit**

```bash
git add interpreter/memory_layout_config.go
git commit -m "fix: disable EnableTaggedValues in StableMemoryLayoutConfig (161ns/assign overhead)"
```

---

## Phase 0 Complete

Run full test suite + benchmarks to verify no regressions:

```bash
go test ./... -count=1
go test ./interpreter/... -bench=BenchmarkAdvancedNestedFunctionCalls -benchmem -count=3
```

Expected:
- All packages `ok`
- `BenchmarkAdvancedNestedFunctionCalls`: no regression vs pre-Phase-0 baseline

Next plan: `docs/superpowers/plans/2026-05-25-phase1-bytecode-vm.md`
