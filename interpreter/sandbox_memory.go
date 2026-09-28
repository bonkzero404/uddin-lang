package interpreter

import (
	"errors"
	"reflect"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// Memory accounting for Sandbox.MaxAllocBytes.
//
// MaxValueLen bounds each value, but not how many of them a run makes: a
// 100,000-op script could keep a thousand 1 MB strings, split a 1 MB string
// into a million elements (≈ 47 MB from one call) or compile a regex from a
// runtime-built pattern (≈ 330 MB from one call). The interpreter therefore
// charges every allocation a script causes against one per-run budget:
//
//   - + and * on strings, lists and maps: the result, before it is built;
//   - list and map literals, a new map key, the key list a for loop over a
//     map takes, and the argument list of a spread call;
//   - builtins: the fresh part of the result (a string or list the builtin
//     made, not one it passed through) plus the growth of any list, map,
//     set, stack or queue argument it changed;
//   - builtins that allocate far more than they return, or more than their
//     result before they can check it (range, split, regex_split,
//     regex_find_all, json_parse, repeat, str_pad, regex_replace and every
//     regex compile) reserve a worst-case estimate before they run and
//     settle to the real charge afterwards.
//
// Charges are cumulative — freed values are not given back — so the budget
// also bounds the run's peak. Over budget the run stops with LimitError
// "memory", which try/catch cannot catch. The byte figures are estimates of
// the Go heap cost on 64-bit platforms.
const (
	memWord   = 16 // one Value slot: an interface header in a list, or a boxed string header
	memScalar = 8  // a boxed int or float
	memList   = 24 // the slice header behind a list value
	memMap    = 48 // an empty map
	memMapEnt = 64 // one map entry, bucket share and growth slack included
	memObject = 64 // a set, stack or queue value

	// Worst-case bytes per element that the amplifying builtins allocate on
	// the way to their result (measured: split ≈ 47 B per part, range ≈ 40 B
	// per element, regex_find_all ≈ 150 B per match, json_parse ≈ 47 B per
	// input byte).
	memPerSplitPart  = 48
	memPerRangeElem  = 40
	memPerRegexMatch = 160
	memPerJSONByte   = 64

	// Compiling a regex costs about 1 KiB per program instruction plus a
	// fixed part (measured up to ≈ 180 KiB for a short case-folded pattern).
	memRegexBase = 256 << 10
	memRegexInst = 1 << 10
	// Parsing a pattern costs about 250 bytes per pattern byte; a pattern may
	// be at most regexBytesPerInst bytes per allowed instruction long.
	memRegexParseByte = 256
	regexBytesPerInst = 4
)

// LimitMemory reports a run that allocated more than Sandbox.MaxAllocBytes.
// LimitRegexSize reports a regex whose program is larger than
// Sandbox.MaxRegexProgram.
const (
	LimitMemory    = "memory"
	LimitRegexSize = "regex_size"
)

// charge adds n bytes to the run's allocation and stops the run once the
// total passes Sandbox.MaxAllocBytes. A negative n is an overflowed estimate.
func (interp *interpreter) charge(pos Position, n int) {
	max := interp.sandbox.MaxAllocBytes
	if max <= 0 {
		return
	}
	if n < 0 || n > max-interp.allocated {
		interp.allocated = max
		interp.stop(LimitMemory, pos)
	}
	interp.allocated += n
}

// maxErrorQuote bounds a script value quoted into an error message, and
// maxCaughtError the message a sandboxed catch binds: a caught
// "key not found" error used to carry a whole 1 MB key (quoted), and 200 of
// them kept in a list cost 1.36 GiB at 16 charged bytes each.
const (
	maxErrorQuote  = 128
	maxCaughtError = 1024
)

// errorQuote quotes s for an error message, cut to maxErrorQuote bytes.
func errorQuote(s string) string {
	if len(s) <= maxErrorQuote {
		return strconv.Quote(s)
	}
	cut := maxErrorQuote
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strconv.Quote(s[:cut]) + "... (" + strconv.Itoa(len(s)) + " bytes)"
}

// truncateMessage cuts msg to at most max bytes on a rune boundary, copying
// so the full message is not kept alive.
func truncateMessage(msg string, max int) string {
	if len(msg) <= max {
		return msg
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}
	return strings.Clone(msg[:cut]) + "..."
}

// bindCaughtError is the value a sandboxed catch binds: the message cut to
// maxCaughtError bytes and charged to the memory budget.
func (interp *interpreter) bindCaughtError(pos Position, v Value) Value {
	msg, ok := v.(string)
	if !ok {
		return v
	}
	msg = truncateMessage(msg, maxCaughtError)
	interp.charge(pos, memWord+len(msg))
	return msg
}

// sandboxErrorResult cuts the message of an error value a builtin returned
// (a date or pattern that does not parse is quoted in it) to
// maxCaughtError bytes.
func sandboxErrorResult(res Value) Value {
	if e, ok := res.(error); ok {
		if msg := e.Error(); len(msg) > maxCaughtError {
			return errors.New(truncateMessage(msg, maxCaughtError))
		}
	}
	return res
}

// sumEst adds a fixed part to an estimate from mulLen, keeping -1 (overflow).
func sumEst(base, product int) int {
	if product < 0 {
		return -1
	}
	return base + product
}

// settle replaces a reservation made with charge by the real charge.
func (interp *interpreter) settle(pos Position, reserved, used int) {
	interp.allocated -= reserved
	if interp.allocated < 0 {
		interp.allocated = 0
	}
	interp.charge(pos, used)
}

// concatBytes is what l + r allocates.
func concatBytes(l, r Value) int {
	switch l := l.(type) {
	case string:
		if r, ok := r.(string); ok {
			return memWord + len(l) + len(r)
		}
	case *[]Value:
		if r, ok := r.(*[]Value); ok && l != nil && r != nil {
			return memList + memWord*(len(*l)+len(*r))
		}
	case map[string]Value:
		if r, ok := r.(map[string]Value); ok {
			return memMap + memMapEnt*(len(l)+len(r))
		}
	}
	return 0
}

// repeatBytes is what l * r allocates for string and list repetition; -1
// when the size overflows.
func repeatBytes(l, r Value) int {
	n, ok := l.(int)
	other := r
	if !ok {
		n, ok = r.(int)
		other = l
	}
	if !ok || n < 0 {
		return 0 // numeric product, or an error the evaluator reports
	}
	switch v := other.(type) {
	case string:
		b := mulLen(len(v), n)
		if b < 0 {
			return -1
		}
		return memWord + b
	case *[]Value:
		if v == nil {
			return 0
		}
		b := mulLen(memWord*len(*v), n)
		if b < 0 {
			return -1
		}
		return memList + b
	}
	return 0
}

// chargeIteration charges what iterating over v allocates up front: the key
// list of a map (a string is iterated lazily, a list in place).
func (interp *interpreter) chargeIteration(pos Position, v Value) {
	if interp.sandbox == nil {
		return
	}
	if m, ok := v.(map[string]Value); ok {
		interp.charge(pos, memList+memWord*len(m))
	}
}

// chargeSpread charges the argument list a spread call (f(xs...)) builds
// from v, before it is built.
func (interp *interpreter) chargeSpread(pos Position, v Value) {
	if interp.sandbox == nil {
		return
	}
	n := 0
	switch v := v.(type) {
	case string:
		n = 2 * utf8.RuneCountInString(v) // one slot plus one short string per character
	case *[]Value:
		if v != nil {
			n = len(*v)
		}
	case map[string]Value:
		n = 2 * len(v) // the key list, then the arguments
	}
	interp.charge(pos, memWord*n)
}

// regexBuiltin describes a regex builtin: the index of its pattern argument,
// and whether it searches past the first match (find_all, replace, split),
// which compiles a second program (sandbox_regex.go).
type regexBuiltin struct {
	pattern int
	multi   bool
}

var regexPatternArg = map[string]regexBuiltin{
	"is_regex_match": {0, false},
	"regex_match":    {1, false},
	"regex_find":     {1, false},
	"regex_find_all": {1, true},
	"regex_replace":  {1, true},
	"regex_split":    {1, true},
}

// regexProgramSize estimates the compiled size (instructions) of pattern,
// parsed the way the regex builtins parse it (Perl syntax), stopping once it
// passes limit; ok is false for a pattern that does not parse, which the
// builtin then reports without compiling anything. The estimate walks the
// parse tree and multiplies repetitions instead of expanding them
// (Simplify and Compile would build the expanded program first).
func regexProgramSize(pattern string, limit int) (insts int, ok bool) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return 0, false
	}
	return regexProgramEstimate(re, limit), true
}

// regexProgramEstimate is the instruction count re compiles to, roughly as
// regexp/syntax.Compile counts it, capped at limit+1.
func regexProgramEstimate(re *syntax.Regexp, limit int) int {
	capAt := func(n int) int {
		if n < 0 || n > limit {
			return limit + 1
		}
		return n
	}
	switch re.Op {
	case syntax.OpLiteral:
		return capAt(len(re.Rune))
	case syntax.OpCapture:
		return capAt(2 + regexProgramEstimate(re.Sub[0], limit))
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest:
		return capAt(1 + regexProgramEstimate(re.Sub[0], limit))
	case syntax.OpConcat, syntax.OpAlternate:
		n := len(re.Sub) // one split per alternative; concatenation adds none, the slack is harmless
		for _, sub := range re.Sub {
			n = capAt(n + regexProgramEstimate(sub, limit))
			if n > limit {
				return n
			}
		}
		return n
	case syntax.OpRepeat:
		count := re.Max
		if count < 0 {
			count = re.Min + 1 // x{n,} = n copies and a star
		}
		if count < 1 {
			count = 1
		}
		return capAt(mulLen(count, 1+regexProgramEstimate(re.Sub[0], limit)))
	default:
		return 1
	}
}

// guardRegex bounds a regex builtin's pattern before anything compiles it:
// its length (regexBytesPerInst per allowed instruction; parsing alone costs
// about 250 bytes per pattern byte), then its program size against
// Sandbox.MaxRegexProgram, charging the parse and the compile to the memory
// budget (twice for the multi-match builtins, which compile a second
// program). Every regex builtin compiles its pattern on each call, so the
// charge is transient: it returns the bytes charged, which the caller
// settles against what the call keeps. Compile time is charged to the
// regex work budget where the program is compiled (sandbox_regex.go).
func (interp *interpreter) guardRegex(pos Position, name string, args []Value) int {
	rb, isRegex := regexPatternArg[name]
	if !isRegex || rb.pattern >= len(args) {
		return 0
	}
	pattern, ok := args[rb.pattern].(string)
	if !ok {
		return 0
	}
	sb := interp.sandbox
	if sb.MaxRegexProgram <= 0 && sb.MaxAllocBytes <= 0 {
		return 0
	}
	if sb.MaxRegexProgram > 0 && len(pattern)/regexBytesPerInst > sb.MaxRegexProgram {
		interp.stop(LimitRegexSize, pos)
	}
	parse := mulLen(memRegexParseByte, len(pattern))
	interp.charge(pos, parse)
	limit := sb.MaxRegexProgram
	if limit <= 0 {
		limit = sb.MaxAllocBytes/memRegexInst + 1 // past it the compile charge stops the run anyway
	}
	insts, ok := regexProgramSize(pattern, limit)
	if !ok {
		return parse
	}
	if sb.MaxRegexProgram > 0 && insts > sb.MaxRegexProgram {
		interp.stop(LimitRegexSize, pos)
	}
	compile := sumEst(memRegexBase, mulLen(memRegexInst, insts))
	if rb.multi {
		compile = sumEst(compile, compile)
	}
	interp.charge(pos, compile)
	return parse + compile
}

// reserveBuiltin charges a worst-case estimate for a builtin that allocates
// more than its result, or allocates its result before the result can be
// measured, and returns the amount reserved; settle then replaces it with the
// real charge. Arguments of the wrong type reserve nothing: the builtin
// reports them.
func (interp *interpreter) reserveBuiltin(pos Position, name string, args []Value) int {
	strArg := func(i int) (string, bool) {
		if i >= len(args) {
			return "", false
		}
		s, ok := args[i].(string)
		return s, ok
	}
	intArg := func(i int) (int, bool) {
		if i >= len(args) {
			return 0, false
		}
		n, ok := args[i].(int)
		return n, ok
	}
	reserve := 0
	switch name {
	case "range":
		if n := rangeLen(args); n > 0 {
			reserve = sumEst(memList, mulLen(memPerRangeElem, n))
		}
	case "split":
		s, ok := strArg(0)
		if !ok {
			break
		}
		parts := len(s)/2 + 1 // whitespace split: fields are separated
		if sep, ok := strArg(1); ok {
			if sep == "" {
				parts = utf8.RuneCountInString(s)
			} else {
				parts = strings.Count(s, sep) + 1
			}
		}
		interp.sizeCheck(pos, parts)
		reserve = sumEst(memList, mulLen(memPerSplitPart, parts))
	case "regex_find_all", "regex_split":
		s, ok := strArg(0)
		if !ok {
			break
		}
		matches := len(s) + 1
		if limit, ok := intArg(2); ok && limit >= 0 && limit < matches {
			matches = limit
		}
		reserve = sumEst(memList, mulLen(memPerRegexMatch, matches))
	case "json_parse":
		if s, ok := strArg(0); ok {
			reserve = sumEst(memMap, mulLen(memPerJSONByte, len(s)))
		}
	case "repeat":
		if s, ok := strArg(0); ok {
			if n, ok := intArg(1); ok && n > 0 {
				reserve = sumEst(memWord, mulLen(len(s), n))
			}
		}
	case "str_pad":
		s, _ := strArg(0)
		pad, _ := strArg(2)
		if n, ok := intArg(1); ok && n > 0 {
			reserve = sumEst(memWord+len(s), mulLen(len(pad), n))
		}
	case "regex_replace":
		s, _ := strArg(0)
		repl, _ := strArg(2)
		// Worst case: a match at every position, each replaced by repl.
		reserve = sumEst(memWord+len(s), mulLen(len(repl), len(s)+1))
	}
	if reserve < 0 {
		reserve = -1 // an estimate overflowed: charge stops the run
	}
	interp.charge(pos, reserve)
	return reserve
}

// rangeLen is the element count range(args...) produces (0 when the
// arguments are not integers or the range is empty).
func rangeLen(args []Value) int {
	var ints [3]int
	if len(args) == 0 || len(args) > 3 {
		return 0
	}
	for i, a := range args {
		n, ok := a.(int)
		if !ok || n > rangeArgBound || n < -rangeArgBound {
			return 0
		}
		ints[i] = n
	}
	n := 0
	switch len(args) {
	case 1:
		n = ints[0]
	case 2:
		n = ints[1] - ints[0]
	case 3:
		if ints[2] != 0 {
			n = (ints[1] - ints[0]) / ints[2]
			if (ints[1]-ints[0])%ints[2] != 0 {
				n++
			}
		}
	}
	if n < 0 {
		return 0
	}
	return n
}

// containerState is what a builtin can grow: a list, map, set, stack or
// queue argument, recorded before the call.
type containerState struct {
	list  *[]Value
	data  uintptr // first element of the backing array (lists, stacks, queues)
	cap   int
	len   int
	m     map[string]Value
	set   *Set
	stack *Stack
	queue *Queue
}

func sliceState(s []Value) (uintptr, int, int) {
	return uintptr(unsafe.Pointer(unsafe.SliceData(s))), cap(s), len(s)
}

// snapshotContainers records the growable arguments of a builtin call.
func snapshotContainers(args []Value) []containerState {
	var out []containerState
	for _, a := range args {
		var st containerState
		switch c := a.(type) {
		case *[]Value:
			if c == nil {
				continue
			}
			st.list = c
			st.data, st.cap, st.len = sliceState(*c)
		case map[string]Value:
			st.m = c
			st.len = len(c)
		case *Set:
			if c == nil {
				continue
			}
			st.set = c
			st.len = len(c.data)
		case *Stack:
			if c == nil {
				continue
			}
			st.stack = c
			st.data, st.cap, st.len = sliceState(c.data)
		case *Queue:
			if c == nil {
				continue
			}
			st.queue = c
			st.data, st.cap, st.len = sliceState(c.data)
		default:
			continue
		}
		out = append(out, st)
	}
	return out
}

// sliceGrowth is what changing a backing array from (data, cap, len) to s
// allocated: the whole new array when it was reallocated (a pointer outside
// the old array), else the added elements.
func sliceGrowth(data uintptr, oldCap, oldLen int, s []Value) int {
	nd, nc, nl := sliceState(s)
	if nc > 0 && nd != data && (oldCap == 0 || nd < data || nd >= data+uintptr(oldCap)*memWord) {
		return memList + memWord*nc
	}
	if nl > oldLen {
		return memWord * (nl - oldLen)
	}
	return 0
}

// containerGrowth is what a builtin added to its container arguments.
func containerGrowth(before []containerState, args []Value) int {
	total := 0
	for _, st := range before {
		switch {
		case st.list != nil:
			total += sliceGrowth(st.data, st.cap, st.len, *st.list)
		case st.m != nil:
			if n := len(st.m) - st.len; n > 0 {
				total += memMapEnt * n
			}
		case st.set != nil:
			if n := len(st.set.data) - st.len; n > 0 {
				// entry, insertion-order slot and the "s:"-prefixed key copy
				total += (memMapEnt + memWord) * n
				for _, a := range args {
					if s, ok := a.(string); ok {
						total += (len(s) + 2) * n
					}
				}
			}
		case st.stack != nil:
			total += sliceGrowth(st.data, st.cap, st.len, st.stack.data)
		case st.queue != nil:
			total += sliceGrowth(st.data, st.cap, st.len, st.queue.data)
		}
	}
	return total
}

// withinString reports whether s lies inside the bytes of arg (a builtin
// returned arg or a substring of it, which allocates nothing).
func withinString(s, arg string) bool {
	if len(arg) == 0 {
		return false
	}
	p := uintptr(unsafe.Pointer(unsafe.StringData(s)))
	a := uintptr(unsafe.Pointer(unsafe.StringData(arg)))
	return p >= a && p+uintptr(len(s)) <= a+uintptr(len(arg))
}

// stringBytes is what a string produced by a builtin costs: its boxed header,
// plus its bytes unless it shares them with a string argument.
func stringBytes(s string, args []Value) int {
	for _, a := range args {
		if as, ok := a.(string); ok && withinString(s, as) {
			return memWord
		}
	}
	return memWord + len(s)
}

// elementFreshBuiltins return lists whose elements they made themselves
// (substrings of the input, or numbers); their elements are charged too.
var elementFreshBuiltins = map[string]bool{
	"range": true, "split": true, "regex_split": true, "regex_find_all": true,
}

// resultBytes is the fresh part of a builtin's result. A value the builtin
// passed through (one of its arguments, or a substring of one) costs nothing
// beyond its header; any other string, list or map is charged as new — for
// an element a builtin returns (pop, max, ...) that over-counts, which is the
// safe side. json_parse builds a whole tree and is walked.
func resultBytes(name string, res Value, args []Value, budget int) int {
	switch r := res.(type) {
	case nil, bool:
		return 0
	case int, float64:
		return memScalar
	case string:
		return stringBytes(r, args)
	case *[]Value:
		if r == nil {
			return 0
		}
		for _, a := range args {
			if a, ok := a.(*[]Value); ok && a == r {
				return 0
			}
		}
		if name == "json_parse" {
			return treeBytes(r, budget)
		}
		n := memList + memWord*cap(*r)
		if elementFreshBuiltins[name] {
			for _, e := range *r {
				switch e := e.(type) {
				case string:
					n += stringBytes(e, args)
				case int, float64:
					n += memScalar
				}
			}
		}
		return n
	case map[string]Value:
		for _, a := range args {
			if a, ok := a.(map[string]Value); ok && sameMap(a, r) {
				return 0
			}
		}
		if name == "json_parse" {
			return treeBytes(r, budget)
		}
		return memMap + memMapEnt*len(r)
	case *Set, *Stack, *Queue:
		for _, a := range args {
			if a == res {
				return 0
			}
		}
		return memObject
	case error:
		// Builtins report bad input as an error value, whose message can
		// quote that input (a date that does not parse, a regex).
		return memWord + len(r.Error())
	}
	return 0
}

func sameMap(a, b map[string]Value) bool {
	return reflect.ValueOf(a).UnsafePointer() == reflect.ValueOf(b).UnsafePointer()
}

// treeBytes estimates the heap cost of a freshly built value tree, stopping
// once it passes budget (the caller only needs to know it is over).
func treeBytes(v Value, budget int) int {
	n := 0
	stack := []Value{v}
	for len(stack) > 0 && n <= budget {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch c := cur.(type) {
		case string:
			n += memWord + len(c)
		case int, float64:
			n += memScalar
		case *[]Value:
			if c == nil {
				continue
			}
			n += memList + memWord*cap(*c)
			stack = append(stack, (*c)...)
		case map[string]Value:
			n += memMap
			for k, e := range c {
				n += memMapEnt + len(k)
				stack = append(stack, e)
			}
		}
	}
	return n
}
