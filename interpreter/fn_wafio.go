package interpreter

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
)

// wafCtx extracts the _waf_ctx map from the interpreter variables.
// Returns nil if not set or not the correct type.
func wafCtx(interp *interpreter) map[string]any {
	ctxVal, ok := interp.lookup("_waf_ctx")
	if !ok {
		return nil
	}
	ctx, ok := ctxVal.(map[string]any)
	if !ok {
		return nil
	}
	return ctx
}

// CIDRMatchHelper checks whether ip falls within the CIDR block.
// Exported for direct use in tests and by WAFio integration code.
func CIDRMatchHelper(ip, cidr string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(parsedIP)
}

// PathGlobMatch reports whether path matches the glob pattern.
// Uses filepath.Match semantics. Exported for testability.
func PathGlobMatch(pattern, path string) bool {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

// wafHeaderFunc implements waf_header(name) → string|null
// Returns the value of the named request header (case-insensitive).
func wafHeaderFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_header", args, 1); err != Value(nil) {
		return err
	}
	name, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_header() requires a string argument, not %s", typeName(args[0])))
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(nil)
	}
	headers, ok := ctx["headers"].(map[string]string)
	if !ok {
		return Value(nil)
	}
	lowerName := strings.ToLower(name)
	for k, v := range headers {
		if strings.ToLower(k) == lowerName {
			return Value(v)
		}
	}
	return Value(nil)
}

// wafQueryFunc implements waf_query(name) → string|null
// Returns the named query parameter from the raw query string.
func wafQueryFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_query", args, 1); err != Value(nil) {
		return err
	}
	name, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_query() requires a string argument, not %s", typeName(args[0])))
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(nil)
	}
	query, _ := ctx["query"].(string)
	if query == "" {
		return Value(nil)
	}
	// Parse and URL-decode query params
	for pair := range strings.SplitSeq(query, "&") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key, _ := url.QueryUnescape(parts[0])
			if key == name {
				val, _ := url.QueryUnescape(parts[1])
				return Value(val)
			}
		}
	}
	return Value(nil)
}

// wafBodyContainsFunc implements waf_body_contains(needle) → bool
// Returns true if the request body contains the given substring.
func wafBodyContainsFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_body_contains", args, 1); err != Value(nil) {
		return err
	}
	needle, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_body_contains() requires a string argument, not %s", typeName(args[0])))
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(false)
	}
	body, _ := ctx["body"].(string)
	return Value(strings.Contains(body, needle))
}

// wafCIDRMatchFunc implements waf_cidr_match(ip, cidr) → bool
// Checks whether the given IP address falls within the CIDR range.
func wafCIDRMatchFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_cidr_match", args, 2); err != Value(nil) {
		return err
	}
	ip, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_cidr_match() requires string arguments, not %s", typeName(args[0])))
	}
	cidr, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("waf_cidr_match() requires string arguments, not %s", typeName(args[1])))
	}
	return Value(CIDRMatchHelper(ip, cidr))
}

// wafPathMatchFunc implements waf_path_match(pattern) → bool
// Glob-matches the request path against pattern.
func wafPathMatchFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_path_match", args, 1); err != Value(nil) {
		return err
	}
	pattern, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_path_match() requires a string argument, not %s", typeName(args[0])))
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(false)
	}
	path, _ := ctx["path"].(string)
	return Value(PathGlobMatch(pattern, path))
}

// wafScoreFunc implements waf_score() → float64
// Returns the current WAF threat score from the request context.
func wafScoreFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_score", args, 0); err != Value(nil) {
		return err
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(float64(0))
	}
	score, _ := ctx["score"].(float64)
	return Value(score)
}

// wafActionFunc implements waf_action() → string
// Returns the current WAF action ("ALLOW", "LOG", or "BLOCK").
func wafActionFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_action", args, 0); err != Value(nil) {
		return err
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value("")
	}
	action, _ := ctx["action"].(string)
	return Value(action)
}

// wafDetectedFunc implements waf_detected(category) → bool
// Returns true if the given category appears in ctx["detected"] (case-insensitive).
func wafDetectedFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_detected", args, 1); err != Value(nil) {
		return err
	}
	category, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_detected() requires a string argument, not %s", typeName(args[0])))
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(false)
	}
	detected, _ := ctx["detected"].([]string)
	target := strings.ToLower(category)
	for _, cat := range detected {
		if strings.ToLower(cat) == target {
			return Value(true)
		}
	}
	return Value(false)
}

// wafDetectedAnyFunc implements waf_detected_any() → bool
// Returns true if any categories were detected.
func wafDetectedAnyFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_detected_any", args, 0); err != Value(nil) {
		return err
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(false)
	}
	detected, _ := ctx["detected"].([]string)
	return Value(len(detected) > 0)
}

// wafDetectedListFunc implements waf_detected_list() → []Value
// Returns the list of detected categories as an array.
func wafDetectedListFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_detected_list", args, 0); err != Value(nil) {
		return err
	}
	ctx := wafCtx(interp)
	if ctx == nil {
		return Value(&[]Value{})
	}
	detected, _ := ctx["detected"].([]string)
	result := make([]Value, len(detected))
	for i, cat := range detected {
		result[i] = Value(cat)
	}
	return Value(&result)
}

// wafReturnFunc implements waf_return(action) → (noreturn)
// Writes the action to interp.stdout and exits the current function via panic.
// Valid actions: "ALLOW", "LOG", "BLOCK" (normalized to uppercase).
func wafReturnFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "waf_return", args, 1); err != Value(nil) {
		return err
	}
	action, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("waf_return() requires a string argument, not %s", typeName(args[0])))
	}
	normalized := strings.ToUpper(strings.TrimSpace(action))
	switch normalized {
	case "ALLOW", "LOG", "BLOCK":
		// valid
	default:
		return Value(fmt.Errorf("waf_return() invalid action %q: must be ALLOW, LOG, or BLOCK", action))
	}
	fmt.Fprintf(interp.stdout, "%s", normalized)
	panic(returnResult{pos: pos, value: Value(normalized)})
}

func init() {
	builtins["waf_header"] = builtinFunction{wafHeaderFunc, "waf_header"}
	builtins["waf_query"] = builtinFunction{wafQueryFunc, "waf_query"}
	builtins["waf_body_contains"] = builtinFunction{wafBodyContainsFunc, "waf_body_contains"}
	builtins["waf_cidr_match"] = builtinFunction{wafCIDRMatchFunc, "waf_cidr_match"}
	builtins["waf_path_match"] = builtinFunction{wafPathMatchFunc, "waf_path_match"}
	builtins["waf_score"] = builtinFunction{wafScoreFunc, "waf_score"}
	builtins["waf_action"] = builtinFunction{wafActionFunc, "waf_action"}
	builtins["waf_detected"] = builtinFunction{wafDetectedFunc, "waf_detected"}
	builtins["waf_detected_any"] = builtinFunction{wafDetectedAnyFunc, "waf_detected_any"}
	builtins["waf_detected_list"] = builtinFunction{wafDetectedListFunc, "waf_detected_list"}
	builtins["waf_return"] = builtinFunction{wafReturnFunc, "waf_return"}
}
