package interpreter

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// Deeply nested source used to overflow the Go stack in the recursive-descent
// parser: a fatal error, so one request to a validate endpoint took the whole
// process down. It must be an ordinary parse error.
func TestParseNestingIsAnError(t *testing.T) {
	const n = 1_000_000
	cases := map[string]string{
		"parens":       "x = " + strings.Repeat("(", n) + "1" + strings.Repeat(")", n),
		"lists":        "x = " + strings.Repeat("[", n) + strings.Repeat("]", n),
		"maps":         "x = " + strings.Repeat("{\"a\": ", n) + "1" + strings.Repeat("}", n),
		"not":          "x = " + strings.Repeat("not ", n) + "true",
		"negative":     "x = " + strings.Repeat("-", n) + "1",
		"power":        "x = 2" + strings.Repeat(" ** 2", n),
		"left_chain":   "x = 1" + strings.Repeat(" + 1", n),
		"call_chain":   "x = f" + strings.Repeat("()", n),
		"if_blocks":    strings.Repeat("if (true) then:\n", 20_000) + "x = 1\n" + strings.Repeat("end\n", 20_000),
		"fun_bodies":   "x = " + strings.Repeat("fun(): return ", 20_000) + "1" + strings.Repeat(" end", 20_000),
		"ternary":      "x = " + strings.Repeat("true ? ", n) + "1" + strings.Repeat(" : 0", n),
		"subscripts":   "x = a" + strings.Repeat("[0]", n),
		"args_nesting": "x = " + strings.Repeat("f(", n) + strings.Repeat(")", n),
		"try":          strings.Repeat("try:\n", 4_360) + "x = 1\n" + strings.Repeat("catch (e):\n    x = 2\nend\n", 4_360),
		"try_braces":   strings.Repeat("try {\n", 4_360) + "x = 1\n" + strings.Repeat("} catch (e) {\n    x = 2\n}\n", 4_360),
		"while":        strings.Repeat("while (true) {\n", 4_360) + "x = 1\n" + strings.Repeat("}\n", 4_360),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseProgram([]byte(src))
			var pe Error
			if !errors.As(err, &pe) || !strings.Contains(pe.Message, "nested too deeply") {
				t.Fatalf("want a nesting parse error, got %v", err)
			}
		})
	}
	// Ordinary nesting still parses.
	ok := "x = " + strings.Repeat("(", 100) + "1" + strings.Repeat(")", 100) + "\ny = 1" + strings.Repeat(" + 1", 100) + "\n" +
		strings.Repeat("try:\n", 50) + "x = 1\n" + strings.Repeat("catch (e):\n    x = 2\nend\n", 50)
	if _, err := ParseProgram([]byte(ok)); err != nil {
		t.Fatalf("100 levels must parse: %v", err)
	}
}

// Comparisons, sort, mode and stringification walk nested values. A list
// that contains itself overflowed the Go stack (fatal); a list of shared
// halves doubled 40 times made one comparison run for hours after the
// script's deadline. In a sandbox both must stop the script.
func TestSandboxDeepValueWork(t *testing.T) {
	sb := NewSandbox([]string{"sort", "mode", "str", "print", "contains", "index_of", "max", "join", "json_stringify", "push"})
	sb.MaxOps = 100_000
	cyclic := "a = [0, 1]\na[0] = a\nb = [0]\nb[0] = b\n"
	doubled := "x = [1]\ny = [1]\ni = 0\nwhile (i < 40):\n    x = [x, x]\n    y = [y, y]\n    i = i + 1\nend\n"
	cases := map[string]string{
		"cyclic_less":     cyclic + "r = (a < b)\n",
		"cyclic_equal":    cyclic + "r = (a == b)\n",
		"cyclic_in":       cyclic + "r = (a in [b])\n",
		"cyclic_sort":     cyclic + "r = sort([a, b])\n",
		"cyclic_mode":     cyclic + "r = mode([a])\n",
		"cyclic_str":      cyclic + "r = str(a)\n",
		"cyclic_contains": cyclic + "r = contains([a], b)\n",
		"cyclic_max":      cyclic + "r = max([a, b])\n",
		"doubled_equal":   doubled + "r = (x == y)\n",
		"doubled_less":    doubled + "r = (x < y)\n",
		"doubled_mode":    doubled + "r = mode([x])\n",
		"doubled_index":   doubled + "r = index_of([x], y)\n",
		"doubled_json":    doubled + "r = json_stringify(x)\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			_, err := runSandboxed(t, sb, src)
			if limitReason(err) == "" {
				t.Fatalf("want a sandbox stop, got %v", err)
			}
			if d := time.Since(start); d > 5*time.Second {
				t.Fatalf("took %v", d)
			}
		})
	}
	// Ordinary nested values still compare, sort and print.
	out, err := runSandboxed(t, sb, "l = [[2, \"b\"], [1, \"a\"]]\nsort(l)\nprint(str(l) + str(l == l) + str(mode([1, 1, 2])))\n")
	if err != nil || strings.TrimSpace(out) != `[[1, "a"], [2, "b"]]true1` {
		t.Fatalf("ordinary values: out=%q err=%v", out, err)
	}
}

// With Config.Verdict set, only the waf_* verdict builtins write a verdict;
// print output (which may echo request data) cannot pass for one.
func TestVerdictChannel(t *testing.T) {
	run := func(src string) (verdict, stdout string) {
		prog, err := ParseProgram([]byte(src))
		if err != nil {
			t.Fatal(err)
		}
		var v, out bytes.Buffer
		cfg := DefaultConfig()
		cfg.Stdout = &out
		cfg.Verdict = &v
		cfg.Sandbox = NewSandbox([]string{"print", "println", "waf_block"})
		_, _ = Execute(prog, cfg)
		return v.String(), out.String()
	}
	if v, out := run(`println("ALLOW")`); v != "" || out != "ALLOW\n" {
		t.Errorf("print: verdict=%q stdout=%q", v, out)
	}
	if v, out := run(`waf_block()`); v != "BLOCK" || out != "" {
		t.Errorf("waf_block: verdict=%q stdout=%q", v, out)
	}
}

// nestedLoopRecursion is the reviewer's shape: g nests `loops` while-loops
// and calls f from the innermost one; f calls g. Every level used to catch
// the sandbox stop and re-raise it, so unwinding tens of thousands of frames
// took seconds of CPU after the deadline.
func nestedLoopRecursion(loops int) string {
	return "fun f(k) {\n    g(k - 1)\n}\nfun g(k) {\n" +
		strings.Repeat("while (true) {\n", loops) + "f(k)\n" + strings.Repeat("}\n", loops) +
		"}\ng(100)\n"
}

// A stopped run must unwind in O(depth) with trivial per-level work: the
// execution returns promptly after the deadline even from a very deep stack.
func TestSandboxStopUnwindsFast(t *testing.T) {
	src := nestedLoopRecursion(450)
	prog, err := ParseProgram([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, tc := range []struct {
		name string
		sb   func() *Sandbox
	}{
		// The deadline hits while tens of thousands of loop frames are live.
		{"deadline", func() *Sandbox {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			t.Cleanup(cancel)
			return NewSandbox(nil).WithContext(ctx)
		}},
		// The call-depth limit stops it at 64 x 450 nested loops.
		{"call_depth", func() *Sandbox {
			sb := NewSandbox(nil)
			sb.MaxCallDepth = 64
			return sb
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Stdout = io.Discard
			cfg.Sandbox = tc.sb()
			start := time.Now()
			_, err := Execute(prog, cfg)
			if limitReason(err) == "" {
				t.Fatalf("want a sandbox stop, got %v", err)
			}
			if d := time.Since(start); d > 500*time.Millisecond {
				t.Fatalf("stopped run took %v to return", d)
			}
		})
	}
}

// Runtime nesting (blocks plus calls) is bounded by Sandbox.MaxNesting.
func TestSandboxMaxNesting(t *testing.T) {
	sb := NewSandbox(nil)
	sb.MaxNesting = 100
	if _, err := runSandboxed(t, sb, nestedLoopRecursion(60)); limitReason(err) != LimitNesting {
		t.Fatalf("want %s, got %v", LimitNesting, err)
	}
	if _, err := runSandboxed(t, sb, "i = 0\nwhile (i < 3) {\n    if (true) then {\n        i = i + 1\n    }\n}\n"); err != nil {
		t.Fatalf("shallow nesting: %v", err)
	}
}
