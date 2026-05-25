package interpreter

import (
	"bytes"
	"testing"
)

// makeTestInterp creates an interpreter with _waf_ctx set to the given context map.
func makeTestInterp(ctx map[string]any) *interpreter {
	var buf bytes.Buffer
	config := &Config{
		Vars:       map[string]Value{"_waf_ctx": ctx},
		Stdout:     &buf,
		IsUnitTest: true,
	}
	return newInterpreter(config)
}

// noPos is a zero Position used in direct built-in calls from tests.
var noPos = Position{}

// -- CIDRMatchHelper ----------------------------------------------------------

func TestCIDRMatchHelper_IPv4InRange(t *testing.T) {
	if !CIDRMatchHelper("192.168.1.10", "192.168.1.0/24") {
		t.Error("expected 192.168.1.10 to be inside 192.168.1.0/24")
	}
}

func TestCIDRMatchHelper_IPv4OutOfRange(t *testing.T) {
	if CIDRMatchHelper("10.0.0.1", "192.168.1.0/24") {
		t.Error("expected 10.0.0.1 NOT to be inside 192.168.1.0/24")
	}
}

func TestCIDRMatchHelper_InvalidIP(t *testing.T) {
	if CIDRMatchHelper("not-an-ip", "192.168.1.0/24") {
		t.Error("expected invalid IP to return false")
	}
}

func TestCIDRMatchHelper_InvalidCIDR(t *testing.T) {
	if CIDRMatchHelper("192.168.1.1", "not-a-cidr") {
		t.Error("expected invalid CIDR to return false")
	}
}

func TestCIDRMatchHelper_TorExitNode(t *testing.T) {
	if !CIDRMatchHelper("185.220.101.45", "185.220.0.0/16") {
		t.Error("expected 185.220.101.45 to be inside 185.220.0.0/16")
	}
}

// -- PathGlobMatch ------------------------------------------------------------

func TestPathGlobMatch_ExactMatch(t *testing.T) {
	if !PathGlobMatch("/api/users", "/api/users") {
		t.Error("exact path should match")
	}
}

func TestPathGlobMatch_WildcardSegment(t *testing.T) {
	if !PathGlobMatch("/api/*/users", "/api/v1/users") {
		t.Error("/api/*/users should match /api/v1/users")
	}
}

func TestPathGlobMatch_NonMatching(t *testing.T) {
	if PathGlobMatch("/admin/*", "/api/users") {
		t.Error("/admin/* should NOT match /api/users")
	}
}

func TestPathGlobMatch_DoubleWildcard(t *testing.T) {
	// filepath.Match does not support **, only single *
	// Single * matches within one segment
	matched := PathGlobMatch("/api/*", "/api/v1")
	if !matched {
		t.Error("/api/* should match /api/v1")
	}
}

func TestPathGlobMatch_InvalidPattern(t *testing.T) {
	// '[' without closing ']' is an invalid glob pattern
	if PathGlobMatch("[invalid", "/foo") {
		t.Error("invalid pattern should return false")
	}
}

// -- waf_header ---------------------------------------------------------------

func TestWAFHeader_Exists(t *testing.T) {
	ctx := map[string]any{
		"headers": map[string]string{
			"content-type": "application/json",
		},
	}
	interp := makeTestInterp(ctx)
	result := wafHeaderFunc(interp, noPos, []Value{"content-type"})
	if result != Value("application/json") {
		t.Errorf("expected 'application/json', got %v", result)
	}
}

func TestWAFHeader_CaseInsensitive(t *testing.T) {
	ctx := map[string]any{
		"headers": map[string]string{
			"x-forwarded-for": "1.2.3.4",
		},
	}
	interp := makeTestInterp(ctx)
	// Pass header name in uppercase — should still find it via ToLower
	result := wafHeaderFunc(interp, noPos, []Value{"X-Forwarded-For"})
	if result != Value("1.2.3.4") {
		t.Errorf("expected '1.2.3.4', got %v", result)
	}
}

func TestWAFHeader_Missing(t *testing.T) {
	ctx := map[string]any{
		"headers": map[string]string{},
	}
	interp := makeTestInterp(ctx)
	result := wafHeaderFunc(interp, noPos, []Value{"authorization"})
	if result != Value(nil) {
		t.Errorf("expected nil for missing header, got %v", result)
	}
}

func TestWAFHeader_NoCtx(t *testing.T) {
	interp := makeTestInterp(nil)
	result := wafHeaderFunc(interp, noPos, []Value{"content-type"})
	if result != Value(nil) {
		t.Errorf("expected nil when no ctx, got %v", result)
	}
}

// -- waf_query ----------------------------------------------------------------

func TestWAFQuery_Exists(t *testing.T) {
	ctx := map[string]any{
		"query": "id=42&foo=bar",
	}
	interp := makeTestInterp(ctx)
	result := wafQueryFunc(interp, noPos, []Value{"id"})
	if result != Value("42") {
		t.Errorf("expected '42', got %v", result)
	}
}

func TestWAFQuery_SecondParam(t *testing.T) {
	ctx := map[string]any{
		"query": "id=1&foo=bar",
	}
	interp := makeTestInterp(ctx)
	result := wafQueryFunc(interp, noPos, []Value{"foo"})
	if result != Value("bar") {
		t.Errorf("expected 'bar', got %v", result)
	}
}

func TestWAFQuery_Missing(t *testing.T) {
	ctx := map[string]any{
		"query": "id=1",
	}
	interp := makeTestInterp(ctx)
	result := wafQueryFunc(interp, noPos, []Value{"missing"})
	if result != Value(nil) {
		t.Errorf("expected nil for missing param, got %v", result)
	}
}

func TestWAFQuery_EmptyQuery(t *testing.T) {
	ctx := map[string]any{
		"query": "",
	}
	interp := makeTestInterp(ctx)
	result := wafQueryFunc(interp, noPos, []Value{"id"})
	if result != Value(nil) {
		t.Errorf("expected nil for empty query, got %v", result)
	}
}

// -- waf_body_contains --------------------------------------------------------

func TestWAFBodyContains_Found(t *testing.T) {
	ctx := map[string]any{
		"body": `{"username":"admin","password":"' OR 1=1--"}`,
	}
	interp := makeTestInterp(ctx)
	result := wafBodyContainsFunc(interp, noPos, []Value{"OR 1=1"})
	if result != Value(true) {
		t.Error("expected body_contains to return true")
	}
}

func TestWAFBodyContains_NotFound(t *testing.T) {
	ctx := map[string]any{
		"body": `{"username":"alice"}`,
	}
	interp := makeTestInterp(ctx)
	result := wafBodyContainsFunc(interp, noPos, []Value{"<script>"})
	if result != Value(false) {
		t.Error("expected body_contains to return false")
	}
}

func TestWAFBodyContains_EmptyBody(t *testing.T) {
	ctx := map[string]any{
		"body": "",
	}
	interp := makeTestInterp(ctx)
	result := wafBodyContainsFunc(interp, noPos, []Value{"anything"})
	if result != Value(false) {
		t.Error("expected false for empty body")
	}
}

// -- waf_detected -------------------------------------------------------------

func TestWAFDetected_Found(t *testing.T) {
	ctx := map[string]any{
		"detected": []string{"sql", "xss"},
	}
	interp := makeTestInterp(ctx)
	result := wafDetectedFunc(interp, noPos, []Value{"sql"})
	if result != Value(true) {
		t.Error("expected detected('sql') to be true")
	}
}

func TestWAFDetected_NotFound(t *testing.T) {
	ctx := map[string]any{
		"detected": []string{"sql"},
	}
	interp := makeTestInterp(ctx)
	result := wafDetectedFunc(interp, noPos, []Value{"xss"})
	if result != Value(false) {
		t.Error("expected detected('xss') to be false")
	}
}

func TestWAFDetected_CaseInsensitive(t *testing.T) {
	ctx := map[string]any{
		"detected": []string{"SQL", "XSS"},
	}
	interp := makeTestInterp(ctx)
	result := wafDetectedFunc(interp, noPos, []Value{"sql"})
	if result != Value(true) {
		t.Error("expected case-insensitive detected() to return true")
	}
}

// -- waf_detected_any ---------------------------------------------------------

func TestWAFDetectedAny_Empty(t *testing.T) {
	ctx := map[string]any{
		"detected": []string{},
	}
	interp := makeTestInterp(ctx)
	result := wafDetectedAnyFunc(interp, noPos, []Value{})
	if result != Value(false) {
		t.Error("expected detected_any() == false for empty list")
	}
}

func TestWAFDetectedAny_NonEmpty(t *testing.T) {
	ctx := map[string]any{
		"detected": []string{"cmd"},
	}
	interp := makeTestInterp(ctx)
	result := wafDetectedAnyFunc(interp, noPos, []Value{})
	if result != Value(true) {
		t.Error("expected detected_any() == true when categories present")
	}
}

// -- waf_score ----------------------------------------------------------------

func TestWAFScore_ReturnsFloat(t *testing.T) {
	ctx := map[string]any{
		"score": float64(7.5),
	}
	interp := makeTestInterp(ctx)
	result := wafScoreFunc(interp, noPos, []Value{})
	if result != Value(float64(7.5)) {
		t.Errorf("expected 7.5, got %v", result)
	}
}

func TestWAFScore_DefaultsToZero(t *testing.T) {
	ctx := map[string]any{}
	interp := makeTestInterp(ctx)
	result := wafScoreFunc(interp, noPos, []Value{})
	if result != Value(float64(0)) {
		t.Errorf("expected 0.0, got %v", result)
	}
}

// -- waf_action ---------------------------------------------------------------

func TestWAFAction_ReturnsString(t *testing.T) {
	ctx := map[string]any{
		"action": "BLOCK",
	}
	interp := makeTestInterp(ctx)
	result := wafActionFunc(interp, noPos, []Value{})
	if result != Value("BLOCK") {
		t.Errorf("expected 'BLOCK', got %v", result)
	}
}

func TestWAFAction_Allow(t *testing.T) {
	ctx := map[string]any{
		"action": "ALLOW",
	}
	interp := makeTestInterp(ctx)
	result := wafActionFunc(interp, noPos, []Value{})
	if result != Value("ALLOW") {
		t.Errorf("expected 'ALLOW', got %v", result)
	}
}

// -- waf_return ---------------------------------------------------------------

func TestWAFReturn_Block(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Vars:   map[string]Value{"_waf_ctx": map[string]any{}},
		Stdout: &buf,
	}
	interp := newInterpreter(config)

	var caught returnResult
	func() {
		defer func() {
			if r := recover(); r != nil {
				if rr, ok := r.(returnResult); ok {
					caught = rr
				}
			}
		}()
		wafReturnFunc(interp, noPos, []Value{"BLOCK"})
	}()

	if caught.value != Value("BLOCK") {
		t.Errorf("expected BLOCK, got %v", caught.value)
	}
	if buf.String() != "BLOCK" {
		t.Errorf("expected stdout 'BLOCK', got %q", buf.String())
	}
}

func TestWAFReturn_NormalizesLowercase(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Vars:   map[string]Value{"_waf_ctx": map[string]any{}},
		Stdout: &buf,
	}
	interp := newInterpreter(config)

	var caught returnResult
	func() {
		defer func() {
			if r := recover(); r != nil {
				if rr, ok := r.(returnResult); ok {
					caught = rr
				}
			}
		}()
		wafReturnFunc(interp, noPos, []Value{"allow"})
	}()

	if caught.value != Value("ALLOW") {
		t.Errorf("expected ALLOW after normalization, got %v", caught.value)
	}
}

func TestWAFReturn_InvalidAction(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Vars:   map[string]Value{"_waf_ctx": map[string]any{}},
		Stdout: &buf,
	}
	interp := newInterpreter(config)
	result := wafReturnFunc(interp, noPos, []Value{"DENY"})
	if _, ok := result.(error); !ok {
		t.Errorf("expected error return for invalid action DENY, got %T(%v)", result, result)
	}
}

// -- waf_cidr_match (via interpreter) -----------------------------------------

func TestWAFCIDRMatch_Via_Builtin(t *testing.T) {
	ctx := map[string]any{}
	interp := makeTestInterp(ctx)
	result := wafCIDRMatchFunc(interp, noPos, []Value{"10.0.0.5", "10.0.0.0/8"})
	if result != Value(true) {
		t.Error("expected cidr_match to return true for 10.0.0.5 in 10.0.0.0/8")
	}
}

// -- waf_path_match (via interpreter) -----------------------------------------

func TestWAFPathMatch_Via_Builtin(t *testing.T) {
	ctx := map[string]any{
		"path": "/admin/settings",
	}
	interp := makeTestInterp(ctx)
	result := wafPathMatchFunc(interp, noPos, []Value{"/admin/*"})
	if result != Value(true) {
		t.Error("expected path_match to return true for /admin/settings against /admin/*")
	}
}

func TestWAFPathMatch_NoMatch_Via_Builtin(t *testing.T) {
	ctx := map[string]any{
		"path": "/api/users",
	}
	interp := makeTestInterp(ctx)
	result := wafPathMatchFunc(interp, noPos, []Value{"/admin/*"})
	if result != Value(false) {
		t.Error("expected path_match to return false for /api/users against /admin/*")
	}
}

// -- Registration check -------------------------------------------------------

func TestWAFBuiltinsRegistered(t *testing.T) {
	names := []string{
		"waf_header", "waf_query", "waf_body_contains",
		"waf_cidr_match", "waf_path_match", "waf_score",
		"waf_action", "waf_detected", "waf_detected_any",
		"waf_detected_list", "waf_return",
	}
	for _, name := range names {
		if _, ok := builtins[name]; !ok {
			t.Errorf("builtin %q not registered", name)
		}
	}
}
