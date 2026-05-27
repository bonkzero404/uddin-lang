package main

import (
	"fmt"
	"strings"

	uddin "github.com/bonkzero404/uddin-lang"
)

func run(engine *uddin.Engine, label, src string) {
	var buf strings.Builder
	engine.SetStdout(&buf)
	_, err := engine.ExecuteString(strings.ReplaceAll(src, `\n`, "\n"))
	if err != nil {
		fmt.Printf("[%s] ERROR: %v\n", label, err)
		return
	}
	out := strings.TrimSpace(buf.String())
	fmt.Printf("[%s] %s\n", label, out)
}

func main() {
	fmt.Println("=== UDDIN-LANG stdlib module usage via Go library ===")

	// ── json ──────────────────────────────────────────────────────────────
	fmt.Println("\n── json ──")
	e := uddin.New()
	e.SetUnitTestMode(true)
	run(e, "parse", `import "json"\nresult = json.parse("{\"name\":\"Alice\",\"age\":30}")\nprint(result["name"])`)
	e2 := uddin.New()
	e2.SetUnitTestMode(true)
	run(e2, "stringify", `import "json"\nprint(json.stringify({"x": 1, "y": 2}))`)

	// ── regex ─────────────────────────────────────────────────────────────
	fmt.Println("\n── regex ──")
	e3 := uddin.New()
	e3.SetUnitTestMode(true)
	run(e3, "is_match", `import "regex"\nprint(regex.is_match("^\\d+$", "12345"))`)
	e4 := uddin.New()
	e4.SetUnitTestMode(true)
	run(e4, "find_all", `import "regex"\nmatches = regex.find_all("a1 b22 c333", "\\d+")\nprint(len(matches))`)

	// ── datetime ──────────────────────────────────────────────────────────
	fmt.Println("\n── datetime ──")
	e5 := uddin.New()
	e5.SetUnitTestMode(true)
	run(e5, "now type", `import "datetime"\nprint(typeof(datetime.now()))`)
	e6 := uddin.New()
	e6.SetUnitTestMode(true)
	run(e6, "format", `import "datetime"\nprint(typeof(datetime.format(datetime.now(), "2006-01-02")))`)

	// ── fs ────────────────────────────────────────────────────────────────
	fmt.Println("\n── fs ──")
	e7 := uddin.New()
	e7.SetUnitTestMode(true)
	run(e7, "getcwd", `import "fs"\nprint(typeof(fs.getcwd()))`)
	e8 := uddin.New()
	e8.SetUnitTestMode(true)
	run(e8, "exists", `import "fs"\nprint(fs.exists("/nonexistent/path"))`)

	// ── http ──────────────────────────────────────────────────────────────
	fmt.Println("\n── http ──")
	e9 := uddin.New()
	e9.SetUnitTestMode(true)
	run(e9, "get type", `import "http"\nprint(typeof(http.get))`)

	// ── database ─────────────────────────────────────────────────────────
	fmt.Println("\n── database ──")
	e10 := uddin.New()
	e10.SetUnitTestMode(true)
	run(e10, "connect type", `import "database"\nprint(typeof(database.connect))`)

	// ── fact ──────────────────────────────────────────────────────────────
	fmt.Println("\n── fact ──")
	e11 := uddin.New()
	e11.SetUnitTestMode(true)
	run(e11, "assert+query", `import "fact"\nfact.assert("user", "alice", {"role": "admin"})\nresult = fact.query("user", "alice")\nprint(result["role"])`)

	// ── cdc ───────────────────────────────────────────────────────────────
	fmt.Println("\n── cdc ──")
	e12 := uddin.New()
	e12.SetUnitTestMode(true)
	run(e12, "emit+count", `import "cdc"\ncdc.emit("login", {"user": "alice"})\ncdc.emit("login", {"user": "bob"})\nprint(cdc.count("login"))`)

	// ── waf ───────────────────────────────────────────────────────────────
	fmt.Println("\n── waf ──")
	e13 := uddin.New()
	e13.SetUnitTestMode(true)
	run(e13, "cidr_match", `import "waf"\nprint(waf.cidr_match("10.0.0.5", "10.0.0.0/8"))`)
	e14 := uddin.New()
	e14.SetUnitTestMode(true)
	run(e14, "path_match", `import "waf"\nprint(waf.path_match("/api/*", "/api/users"))`)

	fmt.Println("\n=== done ===")
}
