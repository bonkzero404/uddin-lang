package interpreter

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// K-53 review: a caught "key not found" error quoted the whole key, and the
// catch variable was never charged — 200 caught errors on a 1 MB key kept
// 1.36 GiB. The key is now quoted to 128 bytes and the bound message is
// charged; an error value a builtin returns (a date that does not parse is
// quoted in it) is cut and charged too.
func TestSandboxCaughtErrorsBounded(t *testing.T) {
	src := "k = \"a\" * 1000000\nm = {}\nl = []\ni = 0\nwhile (i < 200):\n" +
		"    try:\n        x = m[k]\n    catch (e):\n        push(l, e)\n    end\n    i = i + 1\nend\n" +
		"print(len(l), len(l[0]))\n"
	alloc, out, err := allocDuring(t, memorySandbox(), src)
	if err != nil {
		t.Fatalf("caught errors: %v", err)
	}
	var n, size int
	if _, scanErr := fmt.Sscan(strings.TrimSpace(out), &n, &size); scanErr != nil || n != 200 || size > 300 {
		t.Fatalf("output %q: want 200 errors of at most 300 bytes", out)
	}
	if alloc > 64<<20 {
		t.Errorf("allocated %d MiB", alloc>>20)
	}

	// An unparseable pattern comes back as an error value quoting it.
	src = "p = \"(\" + (\"a\" * 3000)\nl = []\ni = 0\nwhile (i < 200):\n    push(l, regex_match(\"x\", p))\n    i = i + 1\nend\n" +
		"print(len(str(l[0])))\n"
	_, out, err = allocDuring(t, memorySandbox(), src)
	if err != nil {
		t.Fatalf("returned errors: %v", err)
	}
	if _, scanErr := fmt.Sscan(strings.TrimSpace(out), &size); scanErr != nil || size > maxCaughtError+64 {
		t.Fatalf("returned error kept %q bytes", out)
	}
}

// json_stringify escapes a control byte to six bytes (\u0001); the size
// check used to count one.
func TestSandboxJSONStringifyEscapes(t *testing.T) {
	src := "s = char(1) * 200000\nx = json_stringify([s])\n"
	if _, err := runSandboxed(t, memorySandbox(), src); limitReason(err) != LimitValueSize {
		t.Fatalf("200 KB of control bytes stringifies to 1.2 MB: want value_size, got %v", err)
	}
	bs := `\`
	want := `["a` + bs + `u003cb","` + bs + `u0001"]`
	if out, err := runSandboxed(t, memorySandbox(), "print(json_stringify([\"a<b\", char(1)]))\n"); err != nil ||
		strings.TrimSpace(out) != want {
		t.Fatalf("small values: out=%q err=%v", out, err)
	}
}

// memorySandbox has the limits the wafio WAF agent runs script rules with.
func memorySandbox() *Sandbox {
	sb := NewSandbox([]string{
		"append", "push", "unshift", "len", "range", "str", "join", "split", "lower", "trim",
		"map", "filter", "sort", "json_parse", "is_regex_match", "regex_match", "regex_find_all",
		"regex_replace", "regex_split", "regex_find", "set_new", "set_add", "print",
		"date_parse", "char", "json_stringify", "push",
	})
	sb.MaxOps = 100_000
	sb.MaxCallDepth = 64
	sb.MaxNesting = 256
	sb.MaxValueLen = 1 << 20
	sb.MaxAllocBytes = 4 << 20
	sb.MaxRegexProgram = 1024
	return sb
}

// allocDuring runs src and reports the bytes the process allocated meanwhile.
func allocDuring(t *testing.T, sb *Sandbox, src string) (uint64, string, error) {
	t.Helper()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	out, err := runSandboxed(t, sb, src)
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc, out, err
}

// K-53 remainder: MaxValueLen bounds one value, not how many a run makes or
// what one builtin call allocates on the way. Each of these used to allocate
// from ~50 MB (one split) to tens of GB (a loop within the op budget); under
// the memory budget each stops with a LimitError while the process has
// allocated at most a few times the budget.
func TestSandboxMemoryBudgetStopsAmplifiers(t *testing.T) {
	const allocCeiling = 16 << 20 // 4x the budget: reservations cover the transient part
	cases := []struct {
		name, src, want string
	}{
		{"retain_concat", "s = \"a\" * 1000000\nl = []\ni = 0\nwhile (i < 20000):\n    push(l, s + str(i))\n    i = i + 1\nend\n", LimitMemory},
		{"split_one_call", "s = \"a\" * 1000000\nl = split(s, \"\")\n", LimitMemory},
		{"split_loop", "s = \"a,\" * 32768\nl = []\ni = 0\nwhile (i < 20000):\n    push(l, split(s, \",\"))\n    i = i + 1\nend\n", LimitMemory},
		{"range_loop", "l = []\ni = 0\nwhile (i < 20000):\n    push(l, range(100000))\n    i = i + 1\nend\n", LimitMemory},
		{"list_repeat_loop", "l = []\ni = 0\nwhile (i < 20000):\n    push(l, [i] * 100000)\n    i = i + 1\nend\n", LimitMemory},
		{"map_merge_loop", "m = {}\ni = 0\nwhile (i < 20000):\n    n = {}\n    n[str(i)] = i\n    m = m + n\n    i = i + 1\nend\n", LimitMemory},
		{"set_add_loop", "s = \"a\" * 100000\nst = set_new()\ni = 0\nwhile (i < 20000):\n    set_add(st, s + str(i))\n    i = i + 1\nend\n", LimitMemory},
		{"unshift_loop", "l = []\ni = 0\nwhile (i < 20000):\n    unshift(l, i)\n    i = i + 1\nend\n", LimitMemory},
		{"map_key_iteration", "m = {}\nfor (i in range(10000)):\n    m[str(i)] = i\nend\nwhile (true):\n    for (k in m):\n        break\n    end\nend\n", LimitMemory},
		{"json_parse", "s = \"[\" + (\"1,\" * 500000) + \"1]\"\nl = json_parse(s)\n", LimitMemory},
		{"regex_find_all", "s = \"a\" * 1000000\nl = regex_find_all(s, \"a\")\n", LimitMemory},
		{"regex_split", "s = \"a\" * 1000000\nl = regex_split(s, \"a\")\n", LimitMemory},
		{"regex_runtime_pattern", "p = \"(a|b)\" * 100000\nx = is_regex_match(p, \"ab\")\n", LimitRegexSize},
		{"regex_repeat_pattern", "p = \"(?:a{100}){10}\" * 5\nx = regex_match(\"ab\", p)\n", LimitRegexSize},
		{"try_cannot_swallow", "s = \"a\" * 1000000\ntry:\n    l = split(s, \"\")\ncatch (e):\n    x = 1\nend\nprint(\"after\")\n", LimitMemory},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			alloc, out, err := allocDuring(t, memorySandbox(), tc.src)
			if got := limitReason(err); got != tc.want {
				t.Fatalf("want LimitError %q, got %v", tc.want, err)
			}
			if alloc > allocCeiling {
				t.Errorf("allocated %d MiB before stopping, want < %d MiB", alloc>>20, allocCeiling>>20)
			}
			if strings.Contains(out, "after") {
				t.Errorf("try/catch swallowed the memory stop")
			}
		})
	}
}

// Ordinary rule scripts stay far below the budget.
func TestSandboxMemoryOrdinaryScripts(t *testing.T) {
	src := `
cookie = "a=1; session=abc; " * 200
n = 0
for (p in split(cookie, ";")):
    kv = split(trim(p), "=")
    if (len(kv) == 2 and lower(kv[0]) == "session") then:
        n = n + 1
    end
end
body = json_parse("{\"user\":\"x\",\"items\":[" + join(map(range(200), str), ",") + "]}")
l = []
for (i in range(2000)):
    append(l, i)
end
m = {}
for (i in range(500)):
    m[str(i)] = i
end
sort(l)
hit = is_regex_match("(?i)(union\\s+select|sleep\\(|benchmark\\()", "id=1 UNION SELECT 1")
parts = regex_split("a1b22c333", "[0-9]+")
clean = regex_replace("<script>x</script>", "(?i)</?script>", "")
print(n, len(body["items"]), len(l), len(m), l[0], hit, len(parts), clean)
`
	out, err := runSandboxed(t, memorySandbox(), src)
	if err != nil {
		t.Fatalf("ordinary script stopped: %v", err)
	}
	if strings.TrimSpace(out) != "200 200 2000 500 0 true 4 x" {
		t.Fatalf("output %q", out)
	}
}

// A zero budget keeps the old behavior: nothing is charged or refused.
func TestSandboxMemoryUnlimitedWhenZero(t *testing.T) {
	sb := memorySandbox()
	sb.MaxAllocBytes = 0
	sb.MaxRegexProgram = 0
	if _, err := runSandboxed(t, sb, "s = \"a\" * 100000\nl = split(s, \"\")\nx = is_regex_match(\"(a|b)\" * 2000, \"ab\")\n"); err != nil {
		t.Fatalf("zero limits must not stop the run: %v", err)
	}
}

// append grew a list to its exact new length on every call, copying the
// whole list each time: 5,000 appends allocated ~386 MB. It now grows like Go
// append.
func TestAppendAmortized(t *testing.T) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	prog, err := ParseProgram([]byte("l = []\ni = 0\nwhile (i < 20000):\n    append(l, i)\n    push(l, i)\n    i = i + 1\nend\nprint(len(l))\n"))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cfg := DefaultConfig()
	cfg.Stdout = &out
	if _, err := Execute(prog, cfg); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	if strings.TrimSpace(out.String()) != "40000" {
		t.Fatalf("output %q", out.String())
	}
	if alloc := after.TotalAlloc - before.TotalAlloc; alloc > 32<<20 {
		t.Errorf("40,000 appends allocated %d MiB", alloc>>20)
	}
}

// Iterating over a string built the whole character list first (~100 bytes
// per character, 104 MB for 1 MB); characters are now produced one at a time,
// with the same values.
func TestStringIterationLazy(t *testing.T) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	_, err := runSandboxed(t, NewSandbox([]string{"print"}), "s = \"a\" * 1000000\nfor (c in s):\n    break\nend\n")
	runtime.ReadMemStats(&after)
	if err != nil {
		t.Fatal(err)
	}
	if alloc := after.TotalAlloc - before.TotalAlloc; alloc > 4<<20 {
		t.Errorf("breaking out of a loop over 1 MB allocated %d MiB", alloc>>20)
	}
	out, err := runSandboxed(t, NewSandbox([]string{"print"}), "for (c in \"héx\"):\n    print(c)\nend\n")
	if err != nil || out != "h\né\nx\n" {
		t.Fatalf("characters: out=%q err=%v", out, err)
	}
}
