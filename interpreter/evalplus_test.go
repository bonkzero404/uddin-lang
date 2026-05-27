package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

func TestEvalPlusLargeArrayNoDuplication(t *testing.T) {
	// Build two arrays of 600 elements each via the language (total > 1000 triggers pool path)
	src := `
a = []
i = 0
while (i < 600):
    a = a + [i]
    i = i + 1
end
b = []
j = 0
while (j < 600):
    b = b + [j * 2]
    j = j + 1
end
c = a + b
print(len(c))
`
	var buf bytes.Buffer
	config := TestConfig()
	config.Stdout = &buf

	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	_, err = Execute(prog, config)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "1200" {
		t.Errorf("expected 1200 elements, got %q", got)
	}
}

func TestEvalPlusArrayConcatSmall(t *testing.T) {
	src := `
a = [1, 2, 3]
b = [4, 5, 6]
c = a + b
print(len(c))
`
	var buf bytes.Buffer
	config := TestConfig()
	config.Stdout = &buf

	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	_, err = Execute(prog, config)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "6" {
		t.Errorf("expected 6 elements, got %q", got)
	}
}
