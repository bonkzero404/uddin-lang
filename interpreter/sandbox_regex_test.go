package interpreter

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The sandbox's regex engine rebuilds FindAll, ReplaceAll and Split from
// single searches; each must return exactly what Go regexp returns.
func TestBudgetRegexMatchesGoRegexp(t *testing.T) {
	interp := newInterpreter(&Config{Sandbox: NewSandbox(nil)})
	patterns := []string{
		`a`, `a*`, `a+`, `a|b`, `^a`, `(?m)^a`, `a$`, `(?m)a$`, `\ba\b`, `\B`, `\b`, `$`, `^`,
		`x*`, ``, `(a)(b)?`, `(?i)É`, `\d+`, `(?:a*b)|a`, `(?U)a+`, `\A`, `\z`, `.`, `(?s).`,
		`[^a]`, `(?P<w>\w+)`, `a|`, `|a`, `(?:)`, `\s*`, `(a*)*`, `(?i)(?:\w+\s*){2}x`, `é|e`,
		// Text that the follow-up program must not be built from by
		// concatenation: an open \Q, flag groups, lazy counted repeats.
		`\Qa`, `\Qa)b`, `\Q(`, `\Qa\Eb`, `a\Q`, `(?i)ab(?-i)c`, `(?U)a+b`, `a{2,3}?`, `(?m:^)x`,
		`[[:alpha:]]+`, `\pL+`, `(?P<x>a)|(?P<y>b)`, `x(?i:Y)z`, `(?s:.)*?a`,
	}
	inputs := []string{
		"", "a", "aaa", "banana", "a\nab\na", "héllo wörld", "\xff\xfea\xff", "ab ab", "x", "aab",
		"12 345 6", "É é e", "b\na", "  a  b  ", "a)b a)b", "((", "abABc aBc", "xYz xyz", "\nx\nx",
	}
	repls := []string{"-", "<$1>", "[${w}]", "$$", "${2}x", ""}
	for _, p := range patterns {
		want := regexp.MustCompile(p)
		got, err := newBudgetRegex(interp, Position{}, p, true)
		if err != nil {
			t.Fatalf("%q: %v", p, err)
		}
		for _, in := range inputs {
			if g, w := got.MatchString(in), want.MatchString(in); g != w {
				t.Errorf("MatchString(%q, %q) = %v, want %v", p, in, g, w)
			}
			if g, w := got.FindString(in), want.FindString(in); g != w {
				t.Errorf("FindString(%q, %q) = %q, want %q", p, in, g, w)
			}
			for _, n := range []int{-1, 0, 1, 2} {
				if g, w := got.FindAllString(in, n), want.FindAllString(in, n); !reflect.DeepEqual(g, w) {
					t.Errorf("FindAllString(%q, %q, %d) = %q, want %q", p, in, n, g, w)
				}
				if g, w := got.Split(in, n), want.Split(in, n); !reflect.DeepEqual(g, w) {
					t.Errorf("Split(%q, %q, %d) = %q, want %q", p, in, n, g, w)
				}
			}
			for _, r := range repls {
				if g, w := got.ReplaceAllString(in, r), want.ReplaceAllString(in, r); g != w {
					t.Errorf("ReplaceAllString(%q, %q, %q) = %q, want %q", p, in, r, g, w)
				}
			}
		}
	}
}

func regexSandbox() *Sandbox {
	sb := memorySandbox()
	sb.MaxRegexWork = 1 << 19 // wafio's engine.RegexBudgetPerRequest
	return sb
}

// K-53 review: coregex ran one match of (?i)(?:\w+\s*){50}x over 4 KiB of
// words for 267-361 ms, far past a 5 ms deadline that was only checked
// between operations; Go regexp's FindAll/ReplaceAll/Split rescan the rest of
// the input per match ((?:a*b)|a over 16 KiB of "a": 2.2 s). Every byte a
// sandboxed regex reads is now charged, so each of these stops within a few
// milliseconds of work.
func TestSandboxRegexWorkBudget(t *testing.T) {
	words := "s = \"word1 word2 \" * 342\n"
	cases := map[string]string{
		"reviewer_4k":          words + "x = is_regex_match(\"(?i)(?:\\\\w+\\\\s*){50}x\", s)\n",
		"reviewer_10_calls":    words + "i = 0\nwhile (i < 10):\n    x = regex_match(s, \"(?i)(?:\\\\w+\\\\s*){50}x\")\n    i = i + 1\nend\n",
		"reviewer_64k":         words + "t = s * 16\nx = is_regex_match(\"(?i)(?:\\\\w+\\\\s*){50}x\", t)\n",
		"quadratic_find_all":   "s = \"a\" * 16384\nx = regex_find_all(s, \"(?:a*b)|a\")\n",
		"quadratic_replace":    "s = \"a\" * 16384\nx = regex_replace(s, \"(?:a*b)|a\", \"x\")\n",
		"quadratic_split":      "s = \"a\" * 16384\nx = regex_split(s, \"(?:a*b)|a\")\n",
		"many_cheap_calls_1mb": "s = \"a\" * 1000000\nx = regex_find(s, \"b\")\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			_, err := runSandboxed(t, regexSandbox(), src)
			d := time.Since(start)
			if limitReason(err) != LimitRegexWork {
				t.Fatalf("want LimitError %q, got %v", LimitRegexWork, err)
			}
			if d > 5*time.Second { // sanity bound only: the stop reason is the check (-race on a loaded host is slow)
				t.Errorf("took %v", d)
			}
		})
	}
	// Ordinary use fits: a 4 KiB body, typical patterns, several calls.
	src := "s = \"id=12&q=union select 1&x=34 \" * 128\n" +
		"a = len(regex_find_all(s, \"\\\\d+\"))\n" +
		"b = is_regex_match(\"(?i)union\\\\s+select\", s)\n" +
		"c = len(regex_split(s, \"&\"))\n" +
		"d = len(regex_replace(s, \"(?i)select\", \"S\"))\n" +
		"print(a, b, c)\n"
	out, err := runSandboxed(t, regexSandbox(), src)
	if err != nil || strings.TrimSpace(out) != "384 true 257" {
		t.Fatalf("ordinary regex use: out=%q err=%v", out, err)
	}
}

// K-53 verification: the follow-up program was built by appending ")" to
// the pattern text, so an unclosed \Q swallowed it — `\Qa` failed to compile
// and is_regex_match silently returned false. It is now built from the parse
// tree, only for the multi-match builtins, and every compile is charged.
func TestSandboxRegexQuoteAndCompileCharge(t *testing.T) {
	out, err := runSandboxed(t, regexSandbox(),
		"print(is_regex_match(\"\\\\Qa\", \"xa\"), regex_match(\"a)b\", \"\\\\Qa)b\"), len(regex_find_all(\"a a a\", \"\\\\Qa\")), regex_split(\"1a)b2\", \"\\\\Qa)b\"))\n")
	if err != nil || strings.TrimSpace(out) != `true true 3 ["1", "2"]` {
		t.Fatalf("\\Q patterns: out=%q err=%v", out, err)
	}

	interp := newInterpreter(&Config{Sandbox: NewSandbox(nil)})
	single, err := newBudgetRegex(interp, Position{}, `\Qa`, false)
	if err != nil || single.after != nil {
		t.Fatalf("match-only regex compiled a follow-up program: %v", err)
	}
	multi, err := newBudgetRegex(interp, Position{}, `\Qa`, true)
	if err != nil || multi.after == nil {
		t.Fatalf("multi-match regex without a follow-up program: %v", err)
	}

	// Compiling is charged: with a budget below one compile, even a match
	// on empty input stops.
	sb := regexSandbox()
	sb.MaxRegexWork = compileUnits(8) // "a" compiles to a handful of instructions
	if _, err := runSandboxed(t, sb, "x = is_regex_match(\"a\", \"\")\n"); err != nil {
		t.Fatalf("one small compile fits: %v", err)
	}
	if _, err := runSandboxed(t, sb, "x = regex_find_all(\"\", \"abcdefgh\")\n"); limitReason(err) != LimitRegexWork {
		t.Fatalf("two compiles over the budget: want regex_work, got %v", err)
	}
}

// The deadline is checked before every builtin call, not only every 256
// operations: a canceled run never starts another builtin.
func TestSandboxDeadlineBeforeBuiltin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sb := regexSandbox().WithContext(ctx)
	if _, err := runSandboxed(t, sb, "x = regex_match(\"a\", \"a\")\n"); limitReason(err) != LimitCanceled {
		t.Fatalf("want canceled before the builtin, got %v", err)
	}
}
