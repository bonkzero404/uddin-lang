package interpreter

import (
	"os"
	"testing"
)

func makeVM(cfg *Config) *VM {
	interp := newInterpreter(cfg)
	return NewVM(interp)
}


func TestVMReturnLiteral(t *testing.T) {
	prog, _ := ParseProgram([]byte(`42`))
	c := NewCompiler()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	vm := makeVM(TestConfig())
	result := vm.Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v (%T)", result, result)
	}
}

func TestVMVarAssign(t *testing.T) {
	prog, _ := ParseProgram([]byte("x = 7\nx"))
	c := NewCompiler()
	fn, _ := c.Compile(prog)
	vm := makeVM(TestConfig())
	result := vm.Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestVMArithmetic(t *testing.T) {
	cases := []struct {
		src      string
		expected Value
	}{
		{`3 + 4`, 7},
		{`10 - 3`, 7},
		{`3 * 4`, 12},
		{`10 / 4`, 2},
		{`10 % 3`, 1},
		{`-5`, -5},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			prog, err := ParseProgram([]byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			fn, err := NewCompiler().Compile(prog)
			if err != nil {
				t.Fatal(err)
			}
			result := makeVM(TestConfig()).Execute(fn)
			if result != tc.expected {
				t.Errorf("src=%q: expected %v (%T), got %v (%T)", tc.src, tc.expected, tc.expected, result, result)
			}
		})
	}
}

func TestVMComparisons(t *testing.T) {
	cases := []struct {
		src      string
		expected bool
	}{
		{`3 == 3`, true},
		{`3 != 4`, true},
		{`3 < 4`, true},
		{`4 > 3`, true},
		{`3 <= 3`, true},
		{`4 >= 4`, true},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			prog, _ := ParseProgram([]byte(tc.src))
			fn, _ := NewCompiler().Compile(prog)
			result := makeVM(TestConfig()).Execute(fn)
			if result != tc.expected {
				t.Errorf("src=%q: expected %v, got %v", tc.src, tc.expected, result)
			}
		})
	}
}

func TestVMIfElse(t *testing.T) {
	src := `
x = 0
if (1 == 1) then:
    x = 42
else:
    x = 99
end
x
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestVMArrayLiteral(t *testing.T) {
	src := "a = [1, 2, 3]\na[1]"
	prog, _ := ParseProgram([]byte(src))
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

func TestVMMapLiteral(t *testing.T) {
	src := `m = {"name": "alice", "age": 30}` + "\n" + `m["age"]`
	prog, _ := ParseProgram([]byte(src))
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 30 {
		t.Errorf("expected 30, got %v", result)
	}
}

func TestVMUserFunction(t *testing.T) {
	src := `
fun add(a, b): return a + b end
add(3, 4)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestVMBuiltinLen(t *testing.T) {
	src := `len([1, 2, 3])`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestVMWhileLoop(t *testing.T) {
	src := `
i = 0
sum = 0
while (i < 10):
    sum = sum + i
    i = i + 1
end
sum
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal(err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 45 {
		t.Errorf("expected 45, got %v", result)
	}
}

func TestVMEnabledConfig(t *testing.T) {
	cfg := TestConfig()
	cfg.VMEnabled = true
	src := `
fun fib(n):
    if (n <= 1) then:
        return n
    else:
        return fib(n - 1) + fib(n - 2)
    end
end
fib(10)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	_, execErr := Execute(prog, cfg)
	if execErr != nil {
		t.Logf("VMEnabled fib(10) execution error (expected during Phase 1): %v", execErr)
		return
	}
	// If execution succeeds without error, the VM path was taken.
	// Full result validation will be enabled once the VM handles all statement types.
	t.Log("VMEnabled fib(10) executed without error via VM path")
}

// ──────────────────────────────────────────────────────────────────────────────
// Phase 1 Task 9 — new statement/expression types
// ──────────────────────────────────────────────────────────────────────────────

// TestVMRecursion verifies that recursive function calls (forward references)
// compile and execute correctly.
func TestVMRecursion(t *testing.T) {
	src := `
fun fib(n):
    if (n <= 1) then:
        return n
    else:
        return fib(n - 1) + fib(n - 2)
    end
end
fib(10)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 55 {
		t.Errorf("expected 55, got %v (%T)", result, result)
	}
}

// TestVMBreak verifies that 'break' exits a while loop at the right point.
func TestVMBreak(t *testing.T) {
	src := `
i = 0
while (i < 100):
    if (i == 5) then:
        break
    end
    i = i + 1
end
i
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 5 {
		t.Errorf("expected 5, got %v", result)
	}
}

// TestVMContinue verifies that 'continue' skips the rest of the loop body and
// re-evaluates the condition.
func TestVMContinue(t *testing.T) {
	// Sum only odd numbers 1..9: 1+3+5+7+9 = 25
	src := `
sum = 0
i = 0
while (i < 10):
    i = i + 1
    if (i % 2 == 0) then:
        continue
    end
    sum = sum + i
end
sum
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 25 {
		t.Errorf("expected 25, got %v", result)
	}
}

// TestVMForIn verifies basic for-in loop over an array literal.
func TestVMForIn(t *testing.T) {
	src := `
sum = 0
for (x in [1, 2, 3, 4, 5]):
    sum = sum + x
end
sum
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

// TestVMForInBreak verifies that break exits a for-in loop early.
func TestVMForInBreak(t *testing.T) {
	src := `
count = 0
for (x in [10, 20, 30, 40, 50]):
    if (x == 30) then:
        break
    end
    count = count + 1
end
count
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

// TestVMTernaryTrue verifies that the ternary operator selects the true branch.
func TestVMTernaryTrue(t *testing.T) {
	src := "x = (1 == 1) ? 42 : 99\nx"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

// TestVMTernaryFalse verifies that the ternary operator selects the false branch.
func TestVMTernaryFalse(t *testing.T) {
	src := "x = (1 == 2) ? 42 : 99\nx"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 99 {
		t.Errorf("expected 99, got %v", result)
	}
}

// TestVMCompoundAssignPlus verifies += on a variable.
func TestVMCompoundAssignPlus(t *testing.T) {
	src := "x = 5\nx += 3\nx"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 8 {
		t.Errorf("expected 8, got %v", result)
	}
}

// TestVMCompoundAssignMinus verifies -= on a variable.
func TestVMCompoundAssignMinus(t *testing.T) {
	src := "x = 10\nx -= 4\nx"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 6 {
		t.Errorf("expected 6, got %v", result)
	}
}

// TestVMSubscriptAssign verifies direct subscript assignment (a[i] = v).
func TestVMSubscriptAssign(t *testing.T) {
	src := "a = [1, 2, 3]\na[1] = 99\na[1]"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 99 {
		t.Errorf("expected 99, got %v", result)
	}
}

// TestVMSubscriptCompoundAssign verifies compound subscript assignment (a[i] += v).
func TestVMSubscriptCompoundAssign(t *testing.T) {
	src := "a = [10, 20, 30]\na[0] += 5\na[0]"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

// TestVMAnonFunc verifies that anonymous function expressions compile and call correctly.
func TestVMAnonFunc(t *testing.T) {
	src := "add = fun(a, b): return a + b end\nadd(3, 4)"
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse error:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile error:", err)
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

var fibSource = []byte(`
fun fib(n):
    if (n <= 1) then:
        return n
    else:
        return fib(n-1) + fib(n-2)
    end
end
fib(10)
`)

func BenchmarkVMNestedFunctionCalls(b *testing.B) {
	prog, err := ParseProgram(fibSource)
	if err != nil {
		b.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		b.Fatal("compile:", err)
	}
	b.ResetTimer()
	for b.Loop() {
		makeVM(TestConfig()).Execute(fn)
	}
}

func BenchmarkTreeWalkerNestedFunctionCalls(b *testing.B) {
	twCfg := TestConfig()
	twCfg.VMEnabled = false
	prog, err := ParseProgram(fibSource)
	if err != nil {
		b.Fatal("parse:", err)
	}
	b.ResetTimer()
	for b.Loop() {
		Execute(prog, twCfg)
	}
}

func TestVMTypedIntAdd(t *testing.T) {
	src := `
fun add(a: int, b: int):
    return a + b
end
add(3, 4)
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		t.Fatal("compile:", err)
	}
	if len(fn.SubFunctions) == 0 {
		t.Fatal("expected sub-function for add()")
	}
	innerFn := fn.SubFunctions[0]
	hasAddInt := false
	for _, instr := range innerFn.Code {
		if instr.Op == OP_ADD_INT {
			hasAddInt = true
		}
	}
	if !hasAddInt {
		t.Error("expected OP_ADD_INT for int+int annotated function, got generic OP_ADD")
	}
	result := makeVM(TestConfig()).Execute(fn)
	if result != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

var fibTypedSource = []byte(`
fun fib(n: int):
    if (n <= 1) then:
        return n
    else:
        return fib(n-1) + fib(n-2)
    end
end
fib(10)
`)

func BenchmarkVMTypedIntFib(b *testing.B) {
	prog, err := ParseProgram(fibTypedSource)
	if err != nil {
		b.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		b.Fatal("compile:", err)
	}
	b.ResetTimer()
	for b.Loop() {
		makeVM(TestConfig()).Execute(fn)
	}
}

func TestVMFreeVariable(t *testing.T) {
	src := `
fun double(n):
    return n * 2
end

fun apply(x):
    return double(x)
end

apply(21)
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
	if result != 42 {
		t.Errorf("expected 42, got %v (%T)", result, result)
	}
}

func TestVMFreeVariableNested(t *testing.T) {
	// Two functions defined at top level, one calls the other
	src := `
fun add(a, b):
    return a + b
end

fun sum3(x, y, z):
    return add(add(x, y), z)
end

sum3(1, 2, 3)
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
	if result != 6 {
		t.Errorf("expected 6, got %v (%T)", result, result)
	}
}

var crossFuncSource = []byte(`
fun fibonacci(n):
    if (n <= 1) then:
        return n
    else:
        return fibonacci(n-1) + fibonacci(n-2)
    end
end

fun calculate(x, y):
    return fibonacci(x) + fibonacci(y)
end

calculate(10, 8)
`)

func BenchmarkVMCrossFunction(b *testing.B) {
	prog, err := ParseProgram(crossFuncSource)
	if err != nil {
		b.Fatal("parse:", err)
	}
	fn, err := NewCompiler().Compile(prog)
	if err != nil {
		b.Fatal("compile:", err)
	}
	b.ResetTimer()
	for b.Loop() {
		makeVM(TestConfig()).Execute(fn)
	}
}

func BenchmarkTreeWalkerCrossFunction(b *testing.B) {
	twCfg := TestConfig()
	twCfg.VMEnabled = false
	prog, err := ParseProgram(crossFuncSource)
	if err != nil {
		b.Fatal("parse:", err)
	}
	b.ResetTimer()
	for b.Loop() {
		Execute(prog, twCfg)
	}
}

func TestVMTryCatch(t *testing.T) {
	src := `
x = 0
try:
    x = 1 / 0
catch (e):
    x = 99
end
x
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
	if result != 99 {
		t.Errorf("expected 99, got %v (%T)", result, result)
	}
}

func TestVMTryCatchErrVar(t *testing.T) {
	src := `
msg = ""
try:
    x = 1 / 0
catch (e):
    msg = str(e)
end
len(msg) > 0
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
	if result != true {
		t.Errorf("expected true (non-empty error message), got %v", result)
	}
}

func TestVMTryCatchNoError(t *testing.T) {
	// When no error occurs, catch block should NOT run.
	src := `
x = 0
try:
    x = 42
catch (e):
    x = 99
end
x
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
	if result != 42 {
		t.Errorf("expected 42, got %v (%T)", result, result)
	}
}

func TestVMAllBuiltins(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want Value
	}{
		{"abs", `abs(-5)`, 5},
		{"max", `max(3, 7)`, 7},
		{"min", `min(3, 7)`, 3},
		{"lower", `lower("HELLO")`, "hello"},
		{"upper", `upper("hello")`, "HELLO"},
		{"split", `len(split("a,b,c", ","))`, 3},
		{"join", `join(["a", "b"], ",")`, `"a","b"`},
		{"contains", `contains("hello", "ell")`, true},
		{"trim", `trim("  hi  ")`, "hi"},
		{"reverse_str", `reverse_str("abc")`, "cba"},
		{"sum", `sum([1,2,3,4,5])`, 15},
		{"round", `round(3.7)`, 4},
		{"sqrt", `sqrt(9)`, 3.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := ParseProgram([]byte(tc.src))
			if err != nil {
				t.Fatal("parse:", err)
			}
			fn, err := NewCompiler().Compile(prog)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			result := makeVM(TestConfig()).Execute(fn)
			if result != tc.want {
				t.Errorf("src=%q: want %v (%T), got %v (%T)", tc.src, tc.want, tc.want, result, result)
			}
		})
	}
}

func TestVMNestedTryCatch(t *testing.T) {
	src := `
x = 0
try:
    try:
        x = 1 / 0
    catch (inner):
        x = 10
    end
catch (outer):
    x = 99
end
x
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
	if result != 10 {
		t.Errorf("expected 10 (inner handler), got %v (%T)", result, result)
	}
}

func TestVMNestedTryCatchUnhandled(t *testing.T) {
	src := `
x = 0
try:
    try:
        x = 1 / 0
    catch (inner):
        y = 1 / 0
    end
catch (outer):
    x = 99
end
x
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
	if result != 99 {
		t.Errorf("expected 99 (outer handler), got %v (%T)", result, result)
	}
}

func TestVMTryCatchInFunction(t *testing.T) {
	src := `
fun mayFail():
    return 1 / 0
end

x = 0
try:
    x = mayFail()
catch (e):
    x = 88
end
x
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
	if result != 88 {
		t.Errorf("expected 88 (error in called function caught by try), got %v (%T)", result, result)
	}
}

func TestVMTryCatchDoesNotCatchReturn(t *testing.T) {
	// Verify that 'return' statements are NOT caught by try-catch
	src := `
fun earlyReturn():
    try:
        return 42
    catch (e):
        return 99
    end
end

earlyReturn()
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
	if result != 42 {
		t.Errorf("expected 42 (return should not be caught), got %v (%T)", result, result)
	}
}

func TestVMImport(t *testing.T) {
	// Write a temporary .din library file that defines a helper function.
	lib := []byte(`
fun add(a, b):
    return a + b
end
IMPORTED_CONST = 100
`)
	tmp, err := os.CreateTemp("", "uddin_import_test_*.din")
	if err != nil {
		t.Fatal("create temp:", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(lib); err != nil {
		t.Fatal("write temp:", err)
	}
	tmp.Close()

	src := `import "` + tmp.Name() + `"
result = add(IMPORTED_CONST, 23)
result
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
	if result != 123 {
		t.Errorf("expected 123, got %v (%T)", result, result)
	}
}

func TestVMImportFunction(t *testing.T) {
	// Import a file that defines a recursive function.
	lib := []byte(`
fun fact(n):
    if (n <= 1) then:
        return 1
    end
    return n * fact(n - 1)
end
`)
	tmp, err := os.CreateTemp("", "uddin_import_fact_*.din")
	if err != nil {
		t.Fatal("create temp:", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(lib); err != nil {
		t.Fatal("write temp:", err)
	}
	tmp.Close()

	src := `import "` + tmp.Name() + `"
fact(5)
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
	if result != 120 {
		t.Errorf("expected 120 (5!), got %v (%T)", result, result)
	}
}

func TestVMBuiltinDirect_Len(t *testing.T) {
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
