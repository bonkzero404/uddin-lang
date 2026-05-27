package stdlib_waf

import (
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

type wafModule struct{}

func (m *wafModule) Name() string { return "waf" }

func (m *wafModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"header":        wafHeaderFunc,
		"query":         wafQueryFunc,
		"body_contains": wafBodyContainsFunc,
		"cidr_match":    wafCIDRMatchFunc,
		"path_match":    wafPathMatchFunc,
		"score":         wafScoreFunc,
		"action":        wafActionFunc,
		"detected":      wafDetectedFunc,
		"detected_any":  wafDetectedAnyFunc,
		"detected_list": wafDetectedListFunc,
		"return":        wafReturnFunc,
	}
}

func init() {
	interpreter.RegisterModule(&wafModule{})
}

// cidrMatch checks whether ip falls within the CIDR block.
func cidrMatch(ip, cidr string) bool {
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

// pathGlobMatch reports whether path matches the glob pattern.
func pathGlobMatch(pattern, path string) bool {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

// getWafCtx retrieves the _waf_ctx map from the module context.
func getWafCtx(ctx interpreter.ModuleContext) map[string]any {
	wafMap, ok := ctx.GetContext("waf")
	if !ok {
		return nil
	}
	return wafMap
}

func wafHeaderFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.header() requires exactly 1 argument, got %d", len(args)))
	}
	name, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.header() requires a string argument, not %T", args[0]))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(nil)
	}
	headers, ok := wafCtx["headers"].(map[string]string)
	if !ok {
		return interpreter.Value(nil)
	}
	lowerName := strings.ToLower(name)
	for k, v := range headers {
		if strings.ToLower(k) == lowerName {
			return interpreter.Value(v)
		}
	}
	return interpreter.Value(nil)
}

func wafQueryFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.query() requires exactly 1 argument, got %d", len(args)))
	}
	name, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.query() requires a string argument, not %T", args[0]))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(nil)
	}
	query, _ := wafCtx["query"].(string)
	if query == "" {
		return interpreter.Value(nil)
	}
	for _, pair := range strings.Split(query, "&") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key, _ := url.QueryUnescape(parts[0])
			if key == name {
				val, _ := url.QueryUnescape(parts[1])
				return interpreter.Value(val)
			}
		}
	}
	return interpreter.Value(nil)
}

func wafBodyContainsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.body_contains() requires exactly 1 argument, got %d", len(args)))
	}
	needle, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.body_contains() requires a string argument, not %T", args[0]))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(false)
	}
	body, _ := wafCtx["body"].(string)
	return interpreter.Value(strings.Contains(body, needle))
}

func wafCIDRMatchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("waf.cidr_match() requires exactly 2 arguments, got %d", len(args)))
	}
	ip, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.cidr_match() requires string arguments, not %T", args[0]))
	}
	cidr, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.cidr_match() requires string arguments, not %T", args[1]))
	}
	return interpreter.Value(cidrMatch(ip, cidr))
}

func wafPathMatchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.path_match() requires exactly 1 argument, got %d", len(args)))
	}
	pattern, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.path_match() requires a string argument, not %T", args[0]))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(false)
	}
	path, _ := wafCtx["path"].(string)
	return interpreter.Value(pathGlobMatch(pattern, path))
}

func wafScoreFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 0 {
		return interpreter.Value(fmt.Errorf("waf.score() takes no arguments, got %d", len(args)))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(float64(0))
	}
	score, _ := wafCtx["score"].(float64)
	return interpreter.Value(score)
}

func wafActionFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 0 {
		return interpreter.Value(fmt.Errorf("waf.action() takes no arguments, got %d", len(args)))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value("")
	}
	action, _ := wafCtx["action"].(string)
	return interpreter.Value(action)
}

func wafDetectedFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.detected() requires exactly 1 argument, got %d", len(args)))
	}
	category, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.detected() requires a string argument, not %T", args[0]))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(false)
	}
	detected, _ := wafCtx["detected"].([]string)
	target := strings.ToLower(category)
	for _, cat := range detected {
		if strings.ToLower(cat) == target {
			return interpreter.Value(true)
		}
	}
	return interpreter.Value(false)
}

func wafDetectedAnyFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 0 {
		return interpreter.Value(fmt.Errorf("waf.detected_any() takes no arguments, got %d", len(args)))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		return interpreter.Value(false)
	}
	detected, _ := wafCtx["detected"].([]string)
	return interpreter.Value(len(detected) > 0)
}

func wafDetectedListFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 0 {
		return interpreter.Value(fmt.Errorf("waf.detected_list() takes no arguments, got %d", len(args)))
	}
	wafCtx := getWafCtx(ctx)
	if wafCtx == nil {
		empty := []interpreter.Value{}
		return interpreter.Value(&empty)
	}
	detected, _ := wafCtx["detected"].([]string)
	result := make([]interpreter.Value, len(detected))
	for i, cat := range detected {
		result[i] = interpreter.Value(cat)
	}
	return interpreter.Value(&result)
}

func wafReturnFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("waf.return() requires exactly 1 argument, got %d", len(args)))
	}
	action, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("waf.return() requires a string argument, not %T", args[0]))
	}
	normalized := strings.ToUpper(strings.TrimSpace(action))
	switch normalized {
	case "ALLOW", "LOG", "BLOCK":
		// valid
	default:
		return interpreter.Value(fmt.Errorf("waf.return() invalid action %q: must be ALLOW, LOG, or BLOCK", action))
	}
	fmt.Fprintf(ctx.Stdout(), "%s", normalized)
	ctx.ReturnValue(interpreter.Value(normalized), pos)
	return interpreter.Value(nil) // unreachable
}
