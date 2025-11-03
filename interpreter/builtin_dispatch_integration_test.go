package interpreter

import (
	"testing"
)

func TestBuiltinDispatchIntegration(t *testing.T) {
	// Test script yang menggunakan fungsi-fungsi yang didaftarkan di dispatcher
	code := `
		// Test fungsi-fungsi yang didaftarkan di dispatcher
		result1 = len("hello")
		result2 = typeof(42)
		result3 = str(3.14)
		result4 = upper("test")
		result5 = lower("TEST")
		result6 = abs(-10)
		result7 = contains("hello world", "world")
	`

	// Parse dan execute program
	program, err := ParseProgram([]byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	config := &Config{
		Vars:   make(map[string]Value),
		Args:   []string{},
		Stdin:  nil,
		Stdout: nil,
		Exit:   nil,
	}

	stats, err := Execute(program, config)
	if err != nil {
		t.Fatalf("Execution error: %v", err)
	}

	// Verifikasi bahwa ada builtin calls yang tercatat
	if stats.BuiltinCalls == 0 {
		t.Error("Expected builtin calls to be recorded, but got 0")
	}

	// Verifikasi dispatcher statistics
	dispatcher := GetGlobalBuiltinDispatcher()

	// Test beberapa fungsi yang seharusnya dipanggil
	functionTests := []struct {
		name     string
		expected int64
	}{
		{"len", 1},
		{"typeof", 1},
		{"str", 1},
		{"upper", 1},
		{"lower", 1},
		{"abs", 1},
		{"contains", 1},
	}

	for _, test := range functionTests {
		count := dispatcher.GetCallCount(test.name)
		if count < test.expected {
			t.Errorf("Expected at least %d calls to %s, got %d", test.expected, test.name, count)
		}
	}

	// Test most called functions
	mostCalled := dispatcher.GetMostCalledFunctions(5)
	if len(mostCalled) == 0 {
		t.Error("Expected some functions to be called, but got empty list")
	}

	t.Logf("Builtin calls recorded: %d", stats.BuiltinCalls)
	t.Logf("Most called functions: %v", mostCalled)
}

func TestBuiltinDispatchPerformance(t *testing.T) {
	// Test untuk memverifikasi bahwa dispatcher memberikan performa yang baik
	code := `
		// Test multiple calls untuk mengukur performa
		numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100]
		for (i in numbers):
			len_result = len("test string")
			type_result = typeof(42)
			str_result = str(3.14)
		end
	`

	program, err := ParseProgram([]byte(code))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	config := &Config{
		Vars:   make(map[string]Value),
		Args:   []string{},
		Stdin:  nil,
		Stdout: nil,
		Exit:   nil,
	}

	stats, err := Execute(program, config)
	if err != nil {
		t.Fatalf("Execution error: %v", err)
	}

	// Verifikasi bahwa banyak builtin calls tercatat
	if stats.BuiltinCalls < 300 { // 100 iterations * 3 functions per iteration
		t.Errorf("Expected at least 300 builtin calls, got %d", stats.BuiltinCalls)
	}

	dispatcher := GetGlobalBuiltinDispatcher()

	// Verifikasi call counts untuk fungsi yang sering dipanggil
	lenCalls := dispatcher.GetCallCount("len")
	typeofCalls := dispatcher.GetCallCount("typeof")
	strCalls := dispatcher.GetCallCount("str")
	rangeCalls := dispatcher.GetCallCount("range")

	if lenCalls < 100 {
		t.Errorf("Expected at least 100 len calls, got %d", lenCalls)
	}
	if typeofCalls < 100 {
		t.Errorf("Expected at least 100 typeof calls, got %d", typeofCalls)
	}
	if strCalls < 100 {
		t.Errorf("Expected at least 100 str calls, got %d", strCalls)
	}
	// Note: Not using range function anymore, using manual array instead

	t.Logf("Performance test completed - Total builtin calls: %d", stats.BuiltinCalls)
	t.Logf("len: %d, typeof: %d, str: %d, range: %d", lenCalls, typeofCalls, strCalls, rangeCalls)
}
