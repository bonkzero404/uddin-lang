package interpreter

import "testing"

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
