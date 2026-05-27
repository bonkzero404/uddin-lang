package stdlib_regex

import (
	"fmt"

	"github.com/bonkzero404/uddin-lang/interpreter"
	"github.com/coregx/coregex"
)

type regexModule struct{}

func (m *regexModule) Name() string { return "regex" }

func (m *regexModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"is_match":  isMatchFunc,
		"match":     matchFunc,
		"find":      findFunc,
		"find_all":  findAllFunc,
		"replace":   replaceFunc,
		"split":     splitFunc,
	}
}

func init() {
	interpreter.RegisterModule(&regexModule{})
}

// isMatchFunc checks if a string matches a regular expression pattern.
// Args: pattern, str
func isMatchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "is_match() requires exactly 2 arguments, got %d", len(args)))
	}
	str, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "is_match() requires second argument to be a string"))
	}
	switch p := args[0].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return interpreter.Value(false)
		}
		return interpreter.Value(re.MatchString(str))
	case *coregex.Regexp:
		return interpreter.Value(p.MatchString(str))
	default:
		panic(interpreter.TypeErrorf(pos, "is_match() requires first argument to be a string or pre-compiled pattern"))
	}
}

// matchFunc tests if a string matches a regular expression pattern.
// Args: text, pattern
func matchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("regex.match() requires exactly 2 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.match() requires first argument to be a string, not %T", args[0]))
	}
	switch p := args[1].(type) {
	case string:
		re, err := coregex.Compile(p)
		if err != nil {
			return interpreter.Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
		return interpreter.Value(re.MatchString(text))
	case *coregex.Regexp:
		return interpreter.Value(p.MatchString(text))
	default:
		return interpreter.Value(fmt.Errorf("regex.match() requires second argument to be a string pattern, not %T", args[1]))
	}
}

// findFunc finds the first match of a regular expression in a string.
// Args: text, pattern
func findFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("regex.find() requires exactly 2 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.find() requires first argument to be a string, not %T", args[0]))
	}
	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return interpreter.Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return interpreter.Value(fmt.Errorf("regex.find() requires second argument to be a string pattern, not %T", args[1]))
	}
	match := re.FindString(text)
	if match == "" {
		return interpreter.Value(nil)
	}
	return interpreter.Value(match)
}

// findAllFunc finds all matches of a regular expression in a string.
// Args: text, pattern, limit? (optional)
func findAllFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		return interpreter.Value(fmt.Errorf("regex.find_all() requires 2 or 3 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.find_all() requires first argument to be a string, not %T", args[0]))
	}
	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return interpreter.Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return interpreter.Value(fmt.Errorf("regex.find_all() requires second argument to be a string pattern, not %T", args[1]))
	}
	limit := -1
	if len(args) == 3 {
		if l, ok := args[2].(int); ok {
			limit = l
		} else {
			return interpreter.Value(fmt.Errorf("regex.find_all() requires third argument to be an integer (limit), not %T", args[2]))
		}
	}
	matches := re.FindAllString(text, limit)
	result := make([]interpreter.Value, len(matches))
	for i, match := range matches {
		result[i] = interpreter.Value(match)
	}
	return interpreter.Value(&result)
}

// replaceFunc replaces matches of a regular expression with a replacement string.
// Args: text, pattern, replacement
func replaceFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 3 {
		return interpreter.Value(fmt.Errorf("regex.replace() requires exactly 3 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.replace() requires first argument to be a string, not %T", args[0]))
	}
	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return interpreter.Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return interpreter.Value(fmt.Errorf("regex.replace() requires second argument to be a string pattern, not %T", args[1]))
	}
	replacement, ok := args[2].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.replace() requires third argument to be a string (replacement), not %T", args[2]))
	}
	return interpreter.Value(re.ReplaceAllString(text, replacement))
}

// splitFunc splits a string using a regular expression pattern as delimiter.
// Args: text, pattern, limit? (optional)
func splitFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		return interpreter.Value(fmt.Errorf("regex.split() requires 2 or 3 arguments, got %d", len(args)))
	}
	text, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("regex.split() requires first argument to be a string, not %T", args[0]))
	}
	var re *coregex.Regexp
	switch p := args[1].(type) {
	case string:
		var err error
		re, err = coregex.Compile(p)
		if err != nil {
			return interpreter.Value(fmt.Errorf("invalid regex pattern: %v", err))
		}
	case *coregex.Regexp:
		re = p
	default:
		return interpreter.Value(fmt.Errorf("regex.split() requires second argument to be a string pattern, not %T", args[1]))
	}
	limit := -1
	if len(args) == 3 {
		if l, ok := args[2].(int); ok {
			limit = l
		} else {
			return interpreter.Value(fmt.Errorf("regex.split() requires third argument to be an integer (limit), not %T", args[2]))
		}
	}
	parts := re.Split(text, limit)
	result := make([]interpreter.Value, len(parts))
	for i, part := range parts {
		result[i] = interpreter.Value(part)
	}
	return interpreter.Value(&result)
}
