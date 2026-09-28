package interpreter

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// Sandbox restricts one execution to what an embedding host allows. The
// wafio WAF agent runs tenant-authored script rules through it: without it a
// script has the whole standard library (files, network, databases, process
// exit, sleep), and a runaway loop cannot be stopped.
//
// A Sandbox is read-only once built and may be shared by concurrent
// executions; Context is per execution, so build one Sandbox per run or copy
// it with WithContext.
type Sandbox struct {
	// Allowed names the builtins the program may use. Every other builtin is
	// absent from scope, the global builtin dispatcher is never consulted,
	// and import statements fail.
	Allowed map[string]bool
	// MaxOps stops execution once this many statements and expressions have
	// been evaluated (0 = unlimited).
	MaxOps int
	// MaxCallDepth bounds user-function nesting (0 = unlimited).
	MaxCallDepth int
	// MaxValueLen bounds the length (bytes for strings, elements for lists)
	// of a value built by +, *, range, join, replace, regex_replace, repeat
	// or str_pad (0 = unlimited). The check runs before the allocation.
	MaxValueLen int
	// Context aborts execution once it is done; checked every 256 operations
	// and on every loop iteration.
	Context context.Context
}

// NewSandbox returns a sandbox allowing exactly the named builtins.
func NewSandbox(allowed []string) *Sandbox {
	set := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		set[name] = true
	}
	return &Sandbox{Allowed: set}
}

// WithContext returns a copy of s bound to ctx.
func (s *Sandbox) WithContext(ctx context.Context) *Sandbox {
	c := *s
	c.Context = ctx
	return &c
}

// Limit reasons reported by LimitError.
const (
	LimitOps       = "op_limit"
	LimitCallDepth = "call_depth"
	LimitValueSize = "value_size"
	LimitValueTree = "value_depth" // nesting past maxValueDepth, or a value that contains itself
	LimitCanceled  = "canceled"
)

// LimitError reports that a sandboxed execution was stopped. try/catch cannot
// catch it: the script ends and Execute returns it.
type LimitError struct {
	Reason string
	pos    Position
}

func (e LimitError) Error() string {
	return fmt.Sprintf("%d:%d: execution stopped: %s", e.pos.Line, e.pos.Column, e.Reason)
}

// Position implements ErrorInterpreter.
func (e LimitError) Position() Position { return e.pos }

// BuiltinNames returns the name of every registered builtin, sorted.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ForbiddenReferences returns, sorted and de-duplicated, every builtin name
// that prog references but allowed does not contain, plus "import" when the
// program has an import statement. Any identifier spelled like a forbidden
// builtin counts, even one the program assigns itself: at run time that name
// is simply absent from a sandboxed scope, so the check stays static and
// conservative. nil means the program only uses allowed builtins.
func ForbiddenReferences(prog *Program, allowed map[string]bool) []string {
	w := refWalker{allowed: allowed, found: map[string]bool{}}
	w.block(prog.Statements)
	if len(w.found) == 0 {
		return nil
	}
	out := make([]string, 0, len(w.found))
	for name := range w.found {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

type refWalker struct {
	allowed map[string]bool
	found   map[string]bool
}

func (w *refWalker) name(n string) {
	if _, builtin := builtins[n]; builtin && !w.allowed[n] {
		w.found[n] = true
	}
}

func (w *refWalker) block(b Block) {
	for _, s := range b {
		w.stmt(s)
	}
}

func (w *refWalker) stmt(s Statement) {
	switch s := s.(type) {
	case *Assign:
		w.expr(s.Target)
		w.expr(s.Value)
	case *OuterAssign:
		w.name(s.Name)
		w.expr(s.Value)
	case *If:
		w.expr(s.Condition)
		w.block(s.Body)
		w.block(s.Else)
	case *While:
		w.expr(s.Condition)
		w.block(s.Body)
	case *For:
		w.name(s.Name)
		w.expr(s.Iterable)
		w.block(s.Body)
	case *TryCatch:
		w.block(s.TryBlock)
		w.name(s.ErrVar)
		w.block(s.CatchBlock)
	case *Return:
		w.expr(s.Result)
	case *ExpressionStatement:
		w.expr(s.Expression)
	case *FunctionDefinition:
		w.name(s.Name)
		for _, p := range s.Parameters {
			w.name(p)
		}
		w.block(s.Body)
	case *Import:
		w.found["import"] = true
	case *Break, *Continue, nil:
	default:
		// A statement type this walker does not know may hide a reference;
		// fail closed.
		w.found[fmt.Sprintf("<%T>", s)] = true
	}
}

func (w *refWalker) expr(e Expression) {
	switch e := e.(type) {
	case nil, *Literal:
	case *Variable:
		w.name(e.Name)
	case *Binary:
		w.expr(e.Left)
		w.expr(e.Right)
	case *Unary:
		w.expr(e.Operand)
	case *Ternary:
		w.expr(e.Condition)
		w.expr(e.TrueExpr)
		w.expr(e.FalseExpr)
	case *Call:
		w.expr(e.Function)
		for _, a := range e.Arguments {
			w.expr(a)
		}
	case *List:
		for _, v := range e.Values {
			w.expr(v)
		}
	case *Map:
		for _, it := range e.Items {
			w.expr(it.Key)
			w.expr(it.Value)
		}
	case *Subscript:
		w.expr(e.Container)
		w.expr(e.Subscript)
	case *FunctionExpression:
		for _, p := range e.Parameters {
			w.name(p)
		}
		w.block(e.Body)
	default:
		w.found[fmt.Sprintf("<%T>", e)] = true
	}
}

// tick counts one operation and enforces the sandbox's op budget and context.
func (interp *interpreter) tick() {
	interp.stats.Ops++
	sb := interp.sandbox
	if sb == nil {
		return
	}
	if sb.MaxOps > 0 && interp.stats.Ops > sb.MaxOps {
		panic(LimitError{Reason: LimitOps, pos: interp.currentPos})
	}
	if interp.done != nil && interp.stats.Ops&255 == 0 {
		interp.checkDone()
	}
}

// checkDone aborts when the execution context is done.
func (interp *interpreter) checkDone() {
	if interp.done == nil {
		return
	}
	select {
	case <-interp.done:
		panic(LimitError{Reason: LimitCanceled, pos: interp.currentPos})
	default:
	}
}

// loopTick runs once per loop iteration, so an empty loop body still counts.
func (interp *interpreter) loopTick() {
	if interp.sandbox == nil {
		return
	}
	interp.tick()
	interp.checkDone()
}

// enterCall tracks user-function nesting in a sandbox; the returned func
// undoes it.
func (interp *interpreter) enterCall(pos Position) func() {
	interp.callDepth++
	if max := interp.sandbox.MaxCallDepth; max > 0 && interp.callDepth > max {
		interp.callDepth--
		panic(LimitError{Reason: LimitCallDepth, pos: pos})
	}
	return func() { interp.callDepth-- }
}

// sizeCheck panics when n exceeds the sandbox's value-length bound.
func (interp *interpreter) sizeCheck(pos Position, n int) {
	if max := interp.sandbox.MaxValueLen; max > 0 && (n < 0 || n > max) {
		panic(LimitError{Reason: LimitValueSize, pos: pos})
	}
}

// valueLen is the length a size-bounded value contributes: bytes of a
// string, elements of a list, 0 for anything else.
func valueLen(v Value) int {
	switch v := v.(type) {
	case string:
		return len(v)
	case *[]Value:
		if v == nil {
			return 0
		}
		return len(*v)
	}
	return 0
}

// mulLen returns size*count, or -1 when it overflows or count is negative.
func mulLen(size, count int) int {
	if count < 0 || size < 0 {
		return -1
	}
	if size == 0 || count == 0 {
		return 0
	}
	if count > (1<<62)/size {
		return -1
	}
	return size * count
}

// guardBinary checks + and * before they allocate.
func (interp *interpreter) guardBinary(pos Position, op Token, l, r Value) {
	if interp.sandbox.MaxValueLen <= 0 {
		return
	}
	switch op {
	case PLUS:
		interp.sizeCheck(pos, valueLen(l)+valueLen(r))
	case TIMES:
		if n, ok := l.(int); ok {
			interp.sizeCheck(pos, mulLen(valueLen(r), n))
		} else if n, ok := r.(int); ok {
			interp.sizeCheck(pos, mulLen(valueLen(l), n))
		}
	}
}

// evalTimesChecked is evalTimes with the sandbox's size check first; every
// * and *= goes through it.
func (interp *interpreter) evalTimesChecked(pos Position, l, r Value) Value {
	if interp.sandbox != nil {
		interp.guardBinary(pos, TIMES, l, r)
	}
	return evalTimes(pos, l, r)
}

// approxSize estimates the serialized size of v (string bytes plus small
// per-element overheads), stopping once it exceeds budget. The walk is
// iterative and bounded by budget, so a huge fan-out of shared references or
// a list that contains itself ends in a size error instead of an enormous
// allocation or unbounded recursion.
func approxSize(v Value, budget int) int {
	n := 0
	stack := []Value{v}
	for len(stack) > 0 && n <= budget {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch cur := cur.(type) {
		case string:
			n += len(cur) + 2
		case *[]Value:
			n += 2
			if cur != nil {
				for _, e := range *cur {
					n++
					if n > budget {
						break
					}
					stack = append(stack, e)
				}
			}
		case map[string]Value:
			n += 2
			for k, e := range cur {
				n += len(k) + 3
				if n > budget {
					break
				}
				stack = append(stack, e)
			}
		default:
			n += 8
		}
	}
	return n
}

// maxValueDepth bounds the nesting of a value that a sandboxed comparison,
// sort or stringification walks.
const maxValueDepth = 256

// deepBuiltins compare, sort, search or stringify nested values; their
// arguments go through guardDeep first. stringifyBuiltins also build a string
// from the whole value and get the output-size check.
var (
	stringifyBuiltins = map[string]bool{
		"str": true, "print": true, "println": true, "input": true,
		"json_stringify": true, "xml_stringify": true, "mode": true,
	}
	deepBuiltins = map[string]bool{
		"str": true, "print": true, "println": true, "input": true,
		"json_stringify": true, "xml_stringify": true, "mode": true,
		"join": true, "sort": true, "index_of": true, "last_index_of": true,
		"contains": true, "find": true, "set_remove": true, "max": true, "min": true,
	}
)

// guardDeep bounds an operation that walks nested values (==, <, in, sort,
// contains, str, mode, ...). It walks every node the operation could visit —
// a sub-list shared twice is visited twice, as the operation would — and each
// node costs one op, so the op budget and the context bound the work: a list
// of shared halves doubled 40 times (2^40 nodes) stops the script instead of
// burning a core long after its deadline. A value that contains itself, or
// nests deeper than maxValueDepth, stops it too, before the recursive
// operation could overflow the Go stack. The walk itself is iterative.
func (interp *interpreter) guardDeep(pos Position, vals ...Value) {
	type frame struct {
		v     Value
		depth int
		leave uintptr // non-zero: this container's subtree is done
	}
	var stack []frame
	for _, v := range vals {
		if _, _, ok := containerOf(v); ok {
			stack = append(stack, frame{v: v, depth: 1})
		}
	}
	if len(stack) == 0 {
		return
	}
	// Backstop when the sandbox has no op budget: the walk itself is bounded.
	nodeCap := interp.sandbox.MaxValueLen
	if nodeCap <= 0 {
		nodeCap = 1 << 20
	}
	onPath := map[uintptr]bool{}
	visited := 0
	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if f.leave != 0 {
			delete(onPath, f.leave)
			continue
		}
		interp.tick()
		if visited++; visited > nodeCap {
			panic(LimitError{Reason: LimitValueSize, pos: pos})
		}
		id, kids, ok := containerOf(f.v)
		if !ok {
			continue
		}
		if f.depth > maxValueDepth || (id != 0 && onPath[id]) {
			panic(LimitError{Reason: LimitValueTree, pos: pos})
		}
		if id != 0 {
			onPath[id] = true
			stack = append(stack, frame{leave: id})
		}
		for _, k := range kids {
			stack = append(stack, frame{v: k, depth: f.depth + 1})
		}
	}
}

// containerOf returns the identity (0 when it has none) and the elements of a
// list or map value; ok is false for scalars.
func containerOf(v Value) (id uintptr, kids []Value, ok bool) {
	switch c := v.(type) {
	case *[]Value:
		if c == nil {
			return 0, nil, true
		}
		return reflect.ValueOf(c).Pointer(), *c, true
	case []Value:
		return 0, c, true
	case map[string]Value:
		kids = make([]Value, 0, len(c))
		for _, e := range c {
			kids = append(kids, e)
		}
		return reflect.ValueOf(c).Pointer(), kids, true
	}
	return 0, nil, false
}

// isComparison reports the operators that walk nested operands.
func isComparison(op Token) bool {
	switch op {
	case EQUAL, NOTEQUAL, LT, LTE, GT, GTE, IN:
		return true
	}
	return false
}

// rangeArgBound keeps range() arithmetic far from int overflow.
const rangeArgBound = 1 << 40

// guardBuiltin checks the builtins whose result size is a function of their
// arguments before they allocate.
func (interp *interpreter) guardBuiltin(pos Position, name string, args []Value) {
	if interp.sandbox.MaxValueLen <= 0 {
		return
	}
	intArg := func(i int) (int, bool) {
		if i >= len(args) {
			return 0, false
		}
		n, ok := args[i].(int)
		return n, ok
	}
	strArg := func(i int) string {
		if i >= len(args) {
			return ""
		}
		s, _ := args[i].(string)
		return s
	}
	switch name {
	case "range":
		var ints []int
		for i := range args {
			n, ok := intArg(i)
			if !ok {
				return // the builtin reports the type error
			}
			if n > rangeArgBound || n < -rangeArgBound {
				interp.sizeCheck(pos, -1)
			}
			ints = append(ints, n)
		}
		n := 0
		switch len(ints) {
		case 1:
			n = ints[0]
		case 2:
			n = ints[1] - ints[0]
		case 3:
			if ints[2] != 0 {
				n = (ints[1] - ints[0]) / ints[2]
			}
		}
		if n > 0 {
			interp.sizeCheck(pos, n)
		}
	case "repeat":
		if n, ok := intArg(1); ok {
			interp.sizeCheck(pos, mulLen(len(strArg(0)), n))
		}
	case "str_pad":
		if n, ok := intArg(1); ok {
			interp.sizeCheck(pos, len(strArg(0))+mulLen(len(strArg(2)), n))
		}
	case "join":
		if len(args) < 1 {
			return
		}
		list, ok := args[0].(*[]Value)
		if !ok || list == nil {
			return
		}
		max := interp.sandbox.MaxValueLen
		interp.sizeCheck(pos, mulLen(len(strArg(1)), len(*list))+approxSize(list, max))
	default:
		if !stringifyBuiltins[name] {
			return
		}
		// Stringifying a list of shared references (or a list containing
		// itself) can be far larger than any value the size checks saw.
		max := interp.sandbox.MaxValueLen
		total := 0
		for _, a := range args {
			total += approxSize(a, max)
			if total > max {
				break
			}
		}
		interp.sizeCheck(pos, total)
	case "replace":
		s, old, repl := strArg(0), strArg(1), strArg(2)
		if len(repl) > len(old) {
			count := len(s) + 1
			if old != "" {
				count = strings.Count(s, old)
			}
			interp.sizeCheck(pos, len(s)+mulLen(len(repl)-len(old), count))
		}
	case "regex_replace":
		s, repl := strArg(0), strArg(2)
		// Worst case: a match at every position, each replaced by repl.
		interp.sizeCheck(pos, len(s)+mulLen(len(repl), len(s)+1))
	}
}

// callSandboxedBuiltin calls an allowed builtin directly: no global
// dispatcher (its per-call write lock and name-based fallback are not
// needed), same argument-count rule.
func (interp *interpreter) callSandboxedBuiltin(bf builtinFunction, pos Position, args []Value) Value {
	if !interp.sandbox.Allowed[bf.Name] {
		panic(nameError(pos, "builtin %q is not available", bf.Name))
	}
	meta := builtinMeta(bf.Name)
	if !meta.FastPath && meta.ArgCount >= 0 && len(args) != meta.ArgCount {
		plural := ""
		if meta.ArgCount != 1 {
			plural = "s"
		}
		panic(typeError(pos, "%s() requires %d arg%s, got %d", bf.Name, meta.ArgCount, plural, len(args)))
	}
	if deepBuiltins[bf.Name] {
		interp.guardDeep(pos, args...)
	}
	interp.guardBuiltin(pos, bf.Name, args)
	return bf.call(interp, pos, args)
}

var (
	builtinMetaOnce  sync.Once
	builtinMetaTable map[string]BuiltinMetadata
)

// builtinMeta is getBuiltinMetadata without rebuilding the whole metadata
// map on every call.
func builtinMeta(name string) BuiltinMetadata {
	builtinMetaOnce.Do(func() { builtinMetaTable = getBuiltinMetadataMap() })
	if meta, ok := builtinMetaTable[name]; ok {
		return meta
	}
	return BuiltinMetadata{ArgCount: -1, FastPath: false}
}
