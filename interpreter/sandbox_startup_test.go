package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestSandboxStartupAllowedNames(t *testing.T) {
	sb := NewSandbox([]string{"print", "upper", "missing_builtin"})
	sb.Allowed["upper"] = false
	for _, name := range []string{"upper", "missing_builtin"} {
		t.Run(name, func(t *testing.T) {
			if _, err := runSandboxed(t, sb, "value = "+name); err == nil || !strings.Contains(err.Error(), "not found") {
				t.Fatalf("%s must remain absent: %v", name, err)
			}
		})
	}
}

func TestSandboxStartupHostVariablesOverrideDefaults(t *testing.T) {
	prog, err := ParseProgram([]byte("print(len)\nprint(PI)\n"))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cfg := DefaultConfig()
	cfg.Sandbox = NewSandbox([]string{"print", "len"})
	cfg.Vars = map[string]Value{"len": "host value", "PI": 7}
	cfg.Stdout = &out
	if _, err := Execute(prog, cfg); err != nil || out.String() != "host value\n7\n" {
		t.Fatalf("host overrides: out=%q err=%v", out.String(), err)
	}
}

func TestSandboxStartupExecutionStateIsolated(t *testing.T) {
	prog, err := ParseProgram([]byte("print(marker)\nmarker = \"changed\"\nupper = marker\n"))
	if err != nil {
		t.Fatal(err)
	}
	sb := NewSandbox([]string{"print", "upper"})
	for i := range 16 {
		t.Run(fmt.Sprintf("run_%d", i), func(t *testing.T) {
			t.Parallel()
			marker := fmt.Sprintf("execution %d", i)
			var out bytes.Buffer
			cfg := DefaultConfig()
			cfg.Sandbox = sb
			cfg.Vars["marker"] = marker
			cfg.Stdout = &out
			for range 3 {
				out.Reset()
				if _, err := Execute(prog, cfg); err != nil || out.String() != marker+"\n" {
					t.Fatalf("isolated execution: out=%q err=%v", out.String(), err)
				}
				if cfg.Vars["marker"] != marker {
					t.Fatalf("execution mutated host variable: %v", cfg.Vars["marker"])
				}
			}
			check, err := ParseProgram([]byte(`print(upper("ok"))`))
			if err != nil {
				t.Fatal(err)
			}
			out.Reset()
			if _, err := Execute(check, cfg); err != nil || out.String() != "OK\n" {
				t.Fatalf("builtin shadow escaped execution: out=%q err=%v", out.String(), err)
			}
		})
	}
}

func TestSandboxStartupNamedClosureAndBuiltinAlias(t *testing.T) {
	sb := NewSandbox([]string{"print", "upper", "len"})
	out, err := runSandboxed(t, sb, `
alias = upper
message = "hello"
fun render():
    print(alias(message))
    print(len(message))
end
upper = "shadowed"
render()
`)
	if err != nil || out != "HELLO\n5\n" {
		t.Fatalf("closure and builtin alias: out=%q err=%v", out, err)
	}
}

func benchmarkSandboxConfig() *Config {
	// Match the size and shape of Wafio's pure-function allowlist.
	sb := NewSandbox(strings.Fields(`
print println typeof int float str char rune
append len range slice sort reverse push pop shift unshift
index_of last_index_of map filter reduce
set_new set_add set_remove set_has set_size set_to_array
stack_new stack_push stack_pop stack_peek stack_size stack_empty
queue_new queue_enqueue queue_dequeue queue_front queue_size queue_empty
join split lower upper contains substr find replace trim
starts_with ends_with reverse_str path_join path_dirname path_basename path_ext
abs max min pow sqrt cbrt round floor ceil trunc sin cos tan asin acos atan atan2
sinh cosh tanh log log10 log2 logb exp exp2 sum mean median mode std_dev variance
gcd lcm sign clamp lerp degrees radians is_nan is_infinite
is_regex_match regex_match regex_find regex_find_all regex_replace regex_split
json_parse json_stringify date_now time_now date_format date_format_new date_parse
date_add date_subtract date_diff date_between date_compare
waf_header waf_query waf_body_contains waf_cidr_match waf_path_match waf_score
waf_action waf_detected waf_detected_any waf_detected_list
waf_return waf_block waf_allow waf_log`))
	sb.MaxOps = 100_000
	sb.MaxCallDepth = 64
	sb.MaxNesting = 256
	sb.MaxValueLen = 1 << 20
	sb.MaxAllocBytes = 4 << 20
	sb.MaxRegexProgram = 1024
	sb.MaxRegexWork = 100_000
	cfg := DefaultConfig()
	cfg.Sandbox = sb
	cfg.Stdout = io.Discard
	cfg.Verdict = io.Discard
	cfg.Vars["_waf_ctx"] = map[string]any{
		"headers": map[string]string{"user-agent": "sqlmap/1.9", "content-type": "application/json"},
		"path":    "/api/users",
		"query":   "id=1",
		"body":    `{"id":1}`,
		"score":   float64(80),
	}
	return cfg
}

func TestSandboxStartupAllocationBudget(t *testing.T) {
	prog, err := ParseProgram([]byte("value = 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := benchmarkSandboxConfig()
	allocs := testing.AllocsPerRun(100, func() {
		if _, err := Execute(prog, cfg); err != nil {
			t.Fatal(err)
		}
	})
	// A trivial rule must not allocate once per allowed builtin or reserve
	// an input buffer. Leave headroom for map and interpreter bookkeeping.
	if allocs > 32 {
		t.Fatalf("trivial sandbox execution allocated %.0f times, want at most 32", allocs)
	}
}

var benchmarkSandboxInterpreter *interpreter

func BenchmarkSandboxStartup(b *testing.B) {
	cfg := benchmarkSandboxConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		benchmarkSandboxInterpreter = newInterpreter(cfg)
	}
}

func BenchmarkSandboxExecution(b *testing.B) {
	prog, err := ParseProgram([]byte(`
fun main():
    agent = waf_header("user-agent")
    if (contains(lower(agent), "sqlmap")) then:
        waf_log()
    end
end
`))
	if err != nil {
		b.Fatal(err)
	}
	cfg := benchmarkSandboxConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := Execute(prog, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
