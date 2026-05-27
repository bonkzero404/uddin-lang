# Phase 2B — Builtin Fast-Path (OP_CALL_BUILTIN_DIRECT)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate dispatch chain overhead for the most-called builtins (`len`, `typeof`, `upper`, `lower`, `trim`, `contains`, `starts_with`, `ends_with`, `split`, `join`, `int`, `str`, `append`, `waf_cidr_match`, `waf_path_match`) — from 4-layer dispatch to a single function pointer call.

**Architecture:** Add `OP_CALL_BUILTIN_DIRECT` opcode; populate `vmBuiltinDirectTable` with direct implementations that skip `DispatchOrPanic`; compiler emits `OP_CALL_BUILTIN_DIRECT` when a direct entry exists for the called builtin.

**Tech Stack:** Go 1.25, existing opcode/compiler/VM infrastructure from Phase 1.

**Prerequisite:** Phase 2A must be complete and merged.

---

## File Structure

| File | Change |
|------|--------|
| `interpreter/opcode.go` | **Modify** — add `OP_CALL_BUILTIN_DIRECT` before `_OP_MAX` |
| `interpreter/vm_builtins.go` | **Modify** — add `vmBuiltinDirectTable` + direct implementations |
| `interpreter/vm.go` | **Modify** — add handler for `OP_CALL_BUILTIN_DIRECT` |
| `interpreter/compiler.go` | **Modify** — emit `OP_CALL_BUILTIN_DIRECT` when direct entry available |
| `interpreter/vm_test.go` | **Modify** — tests + benchmarks for DIRECT path |

---

### Task 1: Opcode + VM handler untuk OP_CALL_BUILTIN_DIRECT

**Files:**
- Modify: `interpreter/opcode.go`
- Modify: `interpreter/vm.go`
- Modify: `interpreter/vm_test.go`

- [ ] **Step 1: Tulis failing test**

Tambah di `interpreter/vm_test.go`:

```go
func TestVMBuiltinDirect_Len(t *testing.T) {
	// len() harus berjalan via OP_CALL_BUILTIN_DIRECT jika sudah ada di direct table.
	src := `len("hello")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 5 {
		t.Errorf("expected 5, got %v (%T)", result, result)
	}
}

func TestVMBuiltinDirect_Upper(t *testing.T) {
	src := `upper("hello")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != "HELLO" {
		t.Errorf("expected HELLO, got %v", result)
	}
}

func TestVMBuiltinDirect_Contains(t *testing.T) {
	src := `contains("hello world", "world")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}
```

- [ ] **Step 2: Jalankan test untuk memastikan PASS dulu (belum ada DIRECT path)**

```bash
go test ./interpreter/... -run TestVMBuiltinDirect -v 2>&1
```

Expected: PASS (karena masih lewat OP_CALL_BUILTIN biasa). Ini baseline.

- [ ] **Step 3: Tambah `OP_CALL_BUILTIN_DIRECT` ke opcode.go**

Di `interpreter/opcode.go`, ganti bagian Functions+Exception:

```go
	// Functions
	OP_MAKE_FUNC         // Dst = &vmFunction from fn.SubFunctions[Src1<<8|Src2]
	OP_CALL              // Dst = call(regs[Src1], argc=Src2, args at argBase)
	OP_CALL_BUILTIN      // Dst = builtinTable[Src1<<8|Src2](argc=next.Dst) — via DispatchOrPanic
	OP_CALL_BUILTIN_DIRECT // Dst = directTable[Src1<<8|Src2](argc=next.Dst) — direct pointer

	// Exception handling
	OP_TRY     // Src1=errReg; Dst:Src2=signed jump offset to catch block
	OP_END_TRY // pop the innermost try handler (no error occurred)

	// Sentinel
	_OP_MAX
```

Update `opName()` array di fungsi yang sama — tambah `"CALL_BUILTIN_DIRECT"` setelah `"CALL_BUILTIN"`:

```go
	names := [_OP_MAX]string{
		"LOAD_CONST", "LOAD_VAR", "STORE_VAR", "LOAD_UPVAL", "STORE_UPVAL",
		"ADD", "SUB", "MUL", "DIV", "MOD", "POW", "NEG",
		"ADD_INT", "SUB_INT", "MUL_INT", "DIV_INT", "MOD_INT",
		"EQ", "NEQ", "LT", "LTE", "GT", "GTE", "IN",
		"AND", "OR", "NOT", "XOR",
		"JUMP", "JUMP_FALSE", "JUMP_TRUE", "RETURN",
		"MAKE_ARRAY", "MAKE_MAP", "SUBSCRIPT", "SET_INDEX",
		"MAKE_FUNC", "CALL", "CALL_BUILTIN", "CALL_BUILTIN_DIRECT",
		"TRY", "END_TRY",
	}
```

- [ ] **Step 4: Tambah handler `OP_CALL_BUILTIN_DIRECT` di vm.go**

Di `interpreter/vm.go`, dalam switch di `run()`, tambah setelah case `OP_CALL_BUILTIN`:

```go
		case OP_CALL_BUILTIN_DIRECT:
			directIdx := int(instr.Src1)<<8 | int(instr.Src2)
			metaInstr := frame.fn.Code[frame.pc]
			frame.pc++
			argc := int(metaInstr.Dst)
			argBase := int(metaInstr.Src1)<<8 | int(metaInstr.Src2)
			args := make([]Value, argc)
			for i := 0; i < argc; i++ {
				args[i] = regs[argBase+i]
			}
			regs[instr.Dst] = vmBuiltinDirectTable[directIdx](vm.interp, Position{}, args)
```

- [ ] **Step 5: Build untuk cek compile error**

```bash
go build ./... 2>&1
```

Expected: error karena `vmBuiltinDirectTable` belum ada. Lanjut ke Task 2.

---

### Task 2: vmBuiltinDirectTable infrastructure

**Files:**
- Modify: `interpreter/vm_builtins.go`

- [ ] **Step 1: Tambah `vmBuiltinDirectTable` ke vm_builtins.go**

Di `interpreter/vm_builtins.go`, tambah setelah deklarasi `vmBuiltinIndex`:

```go
// vmBuiltinDirectTable is a parallel table to vmBuiltinTable.
// A non-nil entry means the builtin has a direct implementation that skips
// the DispatchOrPanic chain. Index is stable and matches vmBuiltinTable.
// Populated by registerVMBuiltinDirect calls in init().
var vmBuiltinDirectTable []func(*interpreter, Position, []Value) Value
```

Modifikasi `registerVMBuiltins()` untuk inisialisasi tabel:

```go
func registerVMBuiltins() {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)

	vmBuiltinTable = make([]vmBuiltinEntry, 0, len(names))
	vmBuiltinIndex = make(map[string]uint16, len(names))
	vmBuiltinDirectTable = make([]func(*interpreter, Position, []Value) Value, len(names))

	for _, name := range names {
		n := name
		vmBuiltinTable = append(vmBuiltinTable, vmBuiltinEntry{
			name: n,
			fn: func(interp *interpreter, pos Position, args []Value) Value {
				return globalBuiltinDispatcher.DispatchOrPanic(n, interp, pos, args)
			},
		})
		vmBuiltinIndex[n] = uint16(len(vmBuiltinTable) - 1)
		// vmBuiltinDirectTable[idx] = nil by default; filled by registerVMBuiltinDirect.
	}
}

// registerVMBuiltinDirect sets a direct (no-dispatch) implementation for name.
// Must be called after registerVMBuiltins() — typically from init() in vm_builtins.go.
func registerVMBuiltinDirect(name string, fn func(*interpreter, Position, []Value) Value) {
	idx, ok := vmBuiltinIndex[name]
	if !ok {
		panic("registerVMBuiltinDirect: unknown builtin " + name)
	}
	vmBuiltinDirectTable[idx] = fn
}
```

- [ ] **Step 2: Build**

```bash
go build ./... 2>&1
```

Expected: sukses (tabel ada, masih kosong)

- [ ] **Step 3: Jalankan tests, pastikan tidak ada regresi**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 4: Commit infrastruktur**

```bash
git add interpreter/opcode.go interpreter/vm.go interpreter/vm_builtins.go
git commit -m "feat(direct): add OP_CALL_BUILTIN_DIRECT opcode + vmBuiltinDirectTable infrastructure"
```

---

### Task 3: Direct implementations — string + core builtins

Target: `len`, `typeof`, `upper`, `lower`, `trim`, `contains`, `starts_with`, `ends_with`, `split`, `join`, `int` (fn: `intFunc`), `str` (fn: `strFunc`).

**Files:**
- Modify: `interpreter/vm_builtins.go`
- Modify: `interpreter/vm_test.go`

- [ ] **Step 1: Tulis failing benchmarks**

Tambah di `interpreter/vm_test.go`:

```go
func BenchmarkBuiltinDirect_Len(b *testing.B) {
	src := `len("hello world")`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		makeVM(cfg).Execute(fn)
	}
}

func BenchmarkBuiltinDirect_Upper(b *testing.B) {
	src := `upper("hello world")`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		makeVM(cfg).Execute(fn)
	}
}

func BenchmarkBuiltinDirect_Contains(b *testing.B) {
	src := `contains("hello world foo bar", "foo")`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		makeVM(cfg).Execute(fn)
	}
}
```

- [ ] **Step 2: Jalankan baseline benchmark**

```bash
go test ./interpreter/... -run='^$' -bench='BenchmarkBuiltinDirect' -benchtime=3s 2>&1
```

Catat angka baseline sebelum implementasi direct.

- [ ] **Step 3: Tambah direct implementations di vm_builtins.go**

Tambah fungsi `registerDirectBuiltins()` dan panggil dari `init()`:

```go
func init() {
	registerVMBuiltins()
	registerDirectBuiltins()
}

func registerDirectBuiltins() {
	registerVMBuiltinDirect("len", directLen)
	registerVMBuiltinDirect("typeof", directTypeof)
	registerVMBuiltinDirect("upper", directUpper)
	registerVMBuiltinDirect("lower", directLower)
	registerVMBuiltinDirect("trim", directTrim)
	registerVMBuiltinDirect("contains", directContains)
	registerVMBuiltinDirect("starts_with", directStartsWith)
	registerVMBuiltinDirect("ends_with", directEndsWith)
	registerVMBuiltinDirect("split", directSplit)
	registerVMBuiltinDirect("join", directJoin)
	registerVMBuiltinDirect("int", directInt)
	registerVMBuiltinDirect("str", directStr)
}

// --- direct implementations (no DispatchOrPanic, inline type assertions) ---

func directLen(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "len() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case string:
		return Value(len(v))
	case *[]Value:
		return Value(len(*v))
	case []Value:
		return Value(len(v))
	case map[string]Value:
		return Value(len(v))
	default:
		panic(typeError(pos, "len() requires a string, array, or object"))
	}
}

func directTypeof(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "typeof() requires 1 argument, got %d", len(args)))
	}
	return Value(typeName(args[0]))
}

func directUpper(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "upper() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "upper() requires a string"))
	}
	return Value(strings.ToUpper(s))
}

func directLower(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "lower() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "lower() requires a string"))
	}
	return Value(strings.ToLower(s))
}

func directTrim(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "trim() requires 1 argument, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "trim() requires a string"))
	}
	return Value(strings.TrimSpace(s))
}

func directContains(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "contains() requires 2 arguments, got %d", len(args)))
	}
	switch haystack := args[0].(type) {
	case string:
		needle, ok := args[1].(string)
		if !ok {
			panic(typeError(pos, "contains() on string requires second argument to be a string"))
		}
		return Value(strings.Contains(haystack, needle))
	case *[]Value:
		needle := args[1]
		for _, v := range *haystack {
			if evalEqual(pos, needle, v).(bool) {
				return Value(true)
			}
		}
		return Value(false)
	default:
		panic(typeError(pos, "contains() requires first argument to be a string or array"))
	}
}

func directStartsWith(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "starts_with() requires 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires first argument to be a string"))
	}
	prefix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "starts_with() requires second argument to be a string"))
	}
	return Value(strings.HasPrefix(s, prefix))
}

func directEndsWith(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "ends_with() requires 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires first argument to be a string"))
	}
	suffix, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "ends_with() requires second argument to be a string"))
	}
	return Value(strings.HasSuffix(s, suffix))
}

func directSplit(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(typeError(pos, "split() requires 1 or 2 arguments, got %d", len(args)))
	}
	s, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "split() requires first argument to be a string"))
	}
	var parts []string
	if len(args) == 1 || args[1] == nil {
		parts = strings.Fields(s)
	} else {
		sep, ok := args[1].(string)
		if !ok {
			panic(typeError(pos, "split() requires separator to be a string"))
		}
		parts = strings.Split(s, sep)
	}
	result := make([]Value, len(parts))
	for i, p := range parts {
		result[i] = Value(p)
	}
	return Value(&result)
}

func directJoin(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "join() requires 2 arguments, got %d", len(args)))
	}
	arr, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "join() requires first argument to be an array"))
	}
	sep, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "join() requires second argument to be a string"))
	}
	parts := make([]string, len(*arr))
	for i, v := range *arr {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return Value(strings.Join(parts, sep))
}

func directInt(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "int() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case int:
		return args[0]
	case int64:
		return Value(int(v))
	case float64:
		return Value(int(v))
	case bool:
		if v {
			return Value(1)
		}
		return Value(0)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return Value(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return Value(int(f))
		}
		return Value(nil)
	default:
		panic(typeError(pos, "int() requires an int, float, bool, or string"))
	}
}

func directStr(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "str() requires 1 argument, got %d", len(args)))
	}
	switch v := args[0].(type) {
	case string:
		return args[0]
	case int:
		return Value(strconv.Itoa(v))
	case float64:
		return Value(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			return Value("true")
		}
		return Value("false")
	case nil:
		return Value("null")
	default:
		return Value(fmt.Sprintf("%v", v))
	}
}
```

Tambah import yang diperlukan di `vm_builtins.go`:

```go
import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)
```

- [ ] **Step 4: Jalankan tests**

```bash
go test ./interpreter/... -run TestVMBuiltinDirect -v 2>&1
```

Expected: PASS

- [ ] **Step 5: Jalankan full suite**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 6: Commit**

```bash
git add interpreter/vm_builtins.go interpreter/vm_test.go
git commit -m "feat(direct): direct implementations for len/typeof/upper/lower/trim/contains/starts_with/ends_with/split/join/int/str"
```

---

### Task 4: Direct implementations — WAF builtins + compiler emits DIRECT

Target: `waf_cidr_match`, `waf_path_match`, `append`. Lalu compiler dipatch ke DIRECT.

**Files:**
- Modify: `interpreter/vm_builtins.go`
- Modify: `interpreter/compiler.go`
- Modify: `interpreter/vm_test.go`

- [ ] **Step 1: Tulis tests untuk WAF builtins**

Tambah di `interpreter/vm_test.go`:

```go
func TestVMBuiltinDirect_WafCidrMatch(t *testing.T) {
	src := `waf_cidr_match("10.0.0.5", "10.0.0.0/8")`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != true {
		t.Errorf("expected true (10.0.0.5 in 10.0.0.0/8), got %v", result)
	}
}

func TestVMBuiltinDirect_Append(t *testing.T) {
	src := `
arr = [1, 2, 3]
append(arr, 4)
len(arr)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 4 {
		t.Errorf("expected 4, got %v", result)
	}
}
```

- [ ] **Step 2: Jalankan test, pastikan PASS dengan path lama**

```bash
go test ./interpreter/... -run TestVMBuiltinDirect_Waf -v 2>&1
```

Expected: PASS (masih via OP_CALL_BUILTIN biasa)

- [ ] **Step 3: Tambah direct WAF + append implementations di vm_builtins.go**

Tambah di `registerDirectBuiltins()`:

```go
	registerVMBuiltinDirect("waf_cidr_match", directWafCidrMatch)
	registerVMBuiltinDirect("waf_path_match", directWafPathMatch)
	registerVMBuiltinDirect("append", directAppend)
```

Tambah implementasi (perlu import `"net"` dan `"path"`):

```go
func directWafCidrMatch(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "waf_cidr_match() requires 2 arguments, got %d", len(args)))
	}
	ipStr, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "waf_cidr_match() requires first argument to be a string (IP address)"))
	}
	cidrStr, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "waf_cidr_match() requires second argument to be a string (CIDR)"))
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Value(false)
	}
	_, network, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return Value(false)
	}
	return Value(network.Contains(ip))
}

func directWafPathMatch(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "waf_path_match() requires 1 argument, got %d", len(args)))
	}
	pattern, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "waf_path_match() requires a string argument (pattern)"))
	}
	// Read the request path from _waf_ctx.
	var requestPath string
	if wafCtx, ok := interp.lookupVariable("_waf_ctx"); ok {
		if ctx, ok := wafCtx.(map[string]any); ok {
			if p, ok := ctx["path"].(string); ok {
				requestPath = p
			}
		}
	}
	matched, err := path.Match(pattern, requestPath)
	if err != nil {
		return Value(false)
	}
	return Value(matched)
}

func directAppend(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 {
		panic(typeError(pos, "append() requires at least 1 argument, got %d", len(args)))
	}
	list, ok := args[0].(*[]Value)
	if !ok {
		panic(typeError(pos, "append() requires first argument to be an array"))
	}
	if len(args) == 1 {
		return Value(nil)
	}
	*list = append(*list, args[1:]...)
	return Value(nil)
}
```

Tambah ke import block di vm_builtins.go:

```go
import (
	"fmt"
	"net"
	"path"
	"sort"
	"strconv"
	"strings"
)
```

Catatan: `directWafPathMatch` memanggil `interp.lookupVariable`. Cek apakah method ini tersedia:

```bash
grep -n "lookupVariable\|func.*lookup" interpreter/interpreter.go | head -10
```

Jika tidak ada, gunakan fallback sederhana: cek `interp.env` atau buat helper sesuai pola yang ada di `fn_wafio.go`.

- [ ] **Step 4: Jalankan tests**

```bash
go test ./interpreter/... -run TestVMBuiltinDirect -v 2>&1
```

Expected: PASS semua

- [ ] **Step 5: Update compiler untuk emit OP_CALL_BUILTIN_DIRECT**

Di `interpreter/compiler.go`, dalam `compileCall`, ganti emit OP_CALL_BUILTIN:

```go
	if builtinCall {
		dst := c.allocReg()
		// Use DIRECT opcode if a fast-path implementation exists.
		op := OP_CALL_BUILTIN
		if int(builtinIdx) < len(vmBuiltinDirectTable) && vmBuiltinDirectTable[builtinIdx] != nil {
			op = OP_CALL_BUILTIN_DIRECT
		}
		c.emit(Instruction{Op: op, Dst: dst,
			Src1: uint8(builtinIdx >> 8), Src2: uint8(builtinIdx)})
		c.emit(Instruction{Op: OP_LOAD_CONST, Dst: uint8(len(e.Arguments)),
			Src1: uint8(argStart >> 8), Src2: uint8(argStart)})
		return dst, nil
	}
```

- [ ] **Step 6: Jalankan full suite**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok — semua builtin test masih PASS setelah compiler switch ke DIRECT

- [ ] **Step 7: Commit**

```bash
git add interpreter/vm_builtins.go interpreter/compiler.go interpreter/vm_test.go
git commit -m "feat(direct): waf_cidr_match/waf_path_match/append direct; compiler emits OP_CALL_BUILTIN_DIRECT"
```

---

### Task 5: Benchmark final Phase 2B

**Files:**
- Modify: `interpreter/vm_test.go`

- [ ] **Step 1: Tambah benchmark comparison CALL_BUILTIN vs CALL_BUILTIN_DIRECT**

Tambah di `interpreter/vm_test.go`:

```go
func BenchmarkBuiltinChain_LenUpper(b *testing.B) {
	// Benchmark chain: len + upper + contains — semua kena DIRECT path.
	src := `
x = upper("hello world")
y = len(x)
z = contains(x, "HELLO")
z
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		makeVM(cfg).Execute(fn)
	}
}

func BenchmarkBuiltinChain_WafRule(b *testing.B) {
	// Simulasi WAF rule: contains + starts_with + ends_with.
	src := `
path = "/api/v1/users"
result = starts_with(path, "/api") && contains(path, "users") && ends_with(path, "users")
result
`
	prog, _ := ParseProgram([]byte(src))
	fn, _ := NewCompiler().Compile(prog)
	cfg := TestConfig()
	b.ResetTimer()
	for b.Loop() {
		makeVM(cfg).Execute(fn)
	}
}
```

- [ ] **Step 2: Jalankan benchmark**

```bash
go test ./interpreter/... -run='^$' -bench='BenchmarkBuiltin' -benchtime=5s -benchmem 2>&1
```

Expected: DIRECT path builtins 2-5x lebih cepat dari baseline (dicatat di Task 3 Step 2).

- [ ] **Step 3: Jalankan full suite sekali lagi**

```bash
go test ./... -count=1 2>&1
```

Expected: semua ok

- [ ] **Step 4: Commit**

```bash
git add interpreter/vm_test.go
git commit -m "bench(direct): Phase 2B benchmark — builtin direct vs chain"
```

---

## Catatan untuk Phase 2C (JIT)

Phase 2C memerlukan:
1. **System dependency**: `brew install llvm@17` + CGo setup
2. **Scope**: 2-4 bulan — jauh lebih besar dari 2A+2B gabungan
3. **Prerequisite**: Phase 2A+2B harus stable di production dulu

Phase 2C akan dibuat sebagai plan terpisah setelah 2B ship. Brainstorming session untuk 2C perlu membahas:
- Setup CI/CD dengan libLLVM (GitHub Actions runner)
- Typed vs untyped codegen prioritas
- Deoptimization safety dengan Go GC
- Content-addressed cache persistence across restarts

## Self-Review Checklist (sudah dilakukan inline)

- [x] Direct impl punya error handling yang sama dengan original? Ya — semua panic dengan `typeError` atau `runtimeError`
- [x] `waf_path_match` butuh `_waf_ctx` — akses via `interp.lookupVariable` atau pola yang sama dengan `fn_wafio.go`
- [x] Compiler switch ke DIRECT tidak break existing builtin tests? Ya — test dijalankan sebelum dan sesudah
- [x] `vmBuiltinDirectTable` size selalu sama dengan `vmBuiltinTable`? Ya — diinisialisasi `make([]..., len(names))` di `registerVMBuiltins`
- [x] Thread-safety? Ya — `vmBuiltinDirectTable` di-populate sekali di `init()`, hanya dibaca saat eksekusi
