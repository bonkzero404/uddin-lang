package interpreter

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func runSandboxed(t *testing.T, sb *Sandbox, src string) (string, error) {
	t.Helper()
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out bytes.Buffer
	cfg := DefaultConfig()
	cfg.Stdout = &out
	cfg.Sandbox = sb
	_, err = Execute(prog, cfg)
	return out.String(), err
}

func limitReason(err error) string {
	var le LimitError
	if errors.As(err, &le) {
		return le.Reason
	}
	return ""
}

func TestSandboxOnlyAllowedBuiltinsInScope(t *testing.T) {
	sb := NewSandbox([]string{"print", "upper"})
	out, err := runSandboxed(t, sb, `print(upper("ok"))`)
	if err != nil || out != "OK\n" {
		t.Fatalf("allowed builtins: out=%q err=%v", out, err)
	}
	// Not in scope and no dispatcher fallback: a NameError, nothing read.
	for _, src := range []string{`read_file("/etc/hostname")`, `x = sleep`, `exit(1)`} {
		if _, err := runSandboxed(t, sb, src); err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("%s: want name error, got %v", src, err)
		}
	}
}

func TestSandboxImportFails(t *testing.T) {
	_, err := runSandboxed(t, NewSandbox(nil), `import "/etc/passwd"`)
	if err == nil || !strings.Contains(err.Error(), "import is not available") {
		t.Fatalf("import in sandbox: %v", err)
	}
}

func TestSandboxOpLimitStopsLoops(t *testing.T) {
	sb := NewSandbox(nil)
	sb.MaxOps = 10_000
	cases := map[string]string{
		"while":       "x = 0\nwhile (true):\n    x = x + 1\nend\n",
		"try_swallow": "while (true):\n    try:\n        while (true):\n            x = 1\n        end\n    catch (e):\n        x = 2\n    end\nend\n",
		"recursion":   "fun f(n):\n    return f(n + 1)\nend\nf(0)\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := runSandboxed(t, sb, src); limitReason(err) == "" {
				t.Fatalf("want a sandbox stop, got %v", err)
			}
		})
	}
}

func TestSandboxEmptyForLoopCounts(t *testing.T) {
	sb := NewSandbox([]string{"range"})
	sb.MaxOps = 1_000
	if _, err := runSandboxed(t, sb, "for (i in range(100000)):\nend\n"); limitReason(err) != LimitOps {
		t.Fatalf("empty for body must still count ops: %v", err)
	}
}

func TestSandboxContextCancels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	sb := NewSandbox(nil).WithContext(ctx)
	start := time.Now()
	_, err := runSandboxed(t, sb, "x = 0\nwhile (true):\n    x = x + 1\nend\n")
	if limitReason(err) != LimitCanceled {
		t.Fatalf("want canceled, got %v", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("cancellation took %v", d)
	}
}

func TestSandboxCallDepth(t *testing.T) {
	sb := NewSandbox(nil)
	sb.MaxCallDepth = 50
	if _, err := runSandboxed(t, sb, "fun f(n):\n    return f(n + 1)\nend\nf(0)\n"); limitReason(err) != LimitCallDepth {
		t.Fatalf("want call_depth, got %v", err)
	}
	if _, err := runSandboxed(t, sb, "fun f(n):\n    if (n == 0) then:\n        return 0\n    end\n    return f(n - 1)\nend\nf(40)\n"); err != nil {
		t.Fatalf("depth 40 under a limit of 50: %v", err)
	}
}

func TestSandboxValueSize(t *testing.T) {
	sb := NewSandbox([]string{"range", "repeat", "join", "replace", "regex_replace", "str_pad", "str", "print", "json_stringify", "push"})
	sb.MaxValueLen = 1 << 20
	over := map[string]string{
		"string_doubling": "s = \"a\"\nwhile (true):\n    s = s + s\nend\n",
		"plus_equal":      "s = \"a\"\nwhile (true):\n    s += s\nend\n",
		"times_equal":     "s = \"ab\"\ns *= 100000000\n",
		"str_fanout":      "s = \"a\" * 1000000\nl = [s] * 1000\nx = str(l)\n",
		"print_fanout":    "s = \"a\" * 1000000\nl = [s] * 1000\nprint(l)\n",
		"json_fanout":     "s = \"a\" * 1000000\nl = [s] * 1000\nx = json_stringify(l)\n",
		"str_cycle":       "l = [1]\npush(l, l)\nx = str(l)\n",
		"string_times":    "s = \"a\" * 1000000000000\n",
		"times_overflow":  "s = \"ab\" * 4611686018427387904\n",
		"list_times":      "l = [1] * 100000000\n",
		"range":           "l = range(100000000)\n",
		"range_overflow":  "l = range(-4611686018427387904, 4611686018427387904)\n",
		"repeat":          "s = repeat(\"a\", 100000000)\n",
		"str_pad":         "s = str_pad(\"a\", 100000000, \"b\")\n",
		"join":            "s = \"a\" * 1000000\nl = [s] * 1000\nx = join(l, \"\")\n",
		"replace":         "s = \"a\" * 1000000\nx = replace(s, \"a\", \"bbbbbbbb\")\n",
		"regex_replace":   "s = \"a\" * 100000\nx = regex_replace(s, \"a\", \"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\")\n",
	}
	for name, src := range over {
		t.Run(name, func(t *testing.T) {
			if _, err := runSandboxed(t, sb, src); limitReason(err) != LimitValueSize {
				t.Fatalf("want value_size, got %v", err)
			}
		})
	}
	if _, err := runSandboxed(t, sb, "s = \"ab\" * 1000\nl = range(1000)\nx = replace(s, \"a\", \"cc\")\n"); err != nil {
		t.Fatalf("small values must pass: %v", err)
	}
}

// Comparing a list that contains itself used to recurse until the Go stack
// overflowed (fatal, not recoverable); it is now a runtime error.
func TestCyclicComparisonIsAnError(t *testing.T) {
	sb := NewSandbox([]string{"push", "contains"})
	for _, src := range []string{
		"l = [1]\npush(l, l)\nm = [1]\npush(m, m)\nx = (l == m)\n",
		"l = [1]\npush(l, l)\nm = [1]\npush(m, m)\nx = contains([l], m)\n",
	} {
		if _, err := runSandboxed(t, sb, src); err == nil || !strings.Contains(err.Error(), "nested too deeply") {
			t.Errorf("%q: got %v", src, err)
		}
	}
}

func TestSandboxSequentialCallbacks(t *testing.T) {
	// map over >= 100 elements normally fans out to worker interpreters that
	// would escape the op budget; in a sandbox it must stay bounded.
	sb := NewSandbox([]string{"map", "range"})
	sb.MaxOps = 5_000
	src := "fun spin(x):\n    while (true):\n        x = x + 1\n    end\nend\nl = map(range(200), spin)\n"
	if _, err := runSandboxed(t, sb, src); limitReason(err) != LimitOps {
		t.Fatalf("map callback escaped the op budget: %v", err)
	}
}

func TestForbiddenReferences(t *testing.T) {
	allowed := map[string]bool{"print": true, "waf_block": true}
	src := `
fun check(input):
    x = read_file("/etc/wafio/certs/waf/agent.key")
    f = http_post
    y = [sleep, {"k": exit}]
    try:
        z = input
    catch (e):
        print(e)
    end
end
import "lib.din"
waf_block()
`
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	got := ForbiddenReferences(prog, allowed)
	want := []string{"exit", "http_post", "import", "input", "read_file", "sleep"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ForbiddenReferences = %v, want %v", got, want)
	}
	clean, _ := ParseProgram([]byte("fun f(a):\n    print(a)\nend\nf(1)\nwaf_block()\n"))
	if got := ForbiddenReferences(clean, allowed); got != nil {
		t.Fatalf("clean program reported %v", got)
	}
}

func TestBuiltinNamesSortedAndComplete(t *testing.T) {
	names := BuiltinNames()
	if len(names) != len(builtins) {
		t.Fatalf("BuiltinNames has %d names, builtins has %d", len(names), len(builtins))
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Fatalf("not sorted at %d: %q >= %q", i, names[i-1], names[i])
		}
	}
}

// Concurrent sandboxed executions share no mutable interpreter state
// (meaningful under -race).
func TestSandboxConcurrentExecutions(t *testing.T) {
	sb := NewSandbox([]string{"print", "upper", "len", "range", "join", "str"})
	sb.MaxOps = 100_000
	prog, err := ParseProgram([]byte("l = []\nfor (i in range(20)):\n    l = l + [str(i)]\nend\nprint(upper(join(l, \",\")))\n"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				var out bytes.Buffer
				cfg := DefaultConfig()
				cfg.Stdout = &out
				cfg.Sandbox = sb
				if _, err := Execute(prog, cfg); err != nil {
					errs <- err
					return
				}
				if !strings.Contains(out.String(), `"19"`) {
					errs <- errors.New("unexpected output " + out.String())
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
