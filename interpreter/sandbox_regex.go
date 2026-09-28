package interpreter

import (
	"io"
	"regexp"
	"regexp/syntax"
	"strings"
	"unicode/utf8"

	"github.com/coregx/coregex"
)

// LimitRegexWork reports a run whose regex matching passed
// Sandbox.MaxRegexWork.
const LimitRegexWork = "regex_work"

// regexEngine is what the regex builtins need from a compiled pattern.
type regexEngine interface {
	MatchString(s string) bool
	FindString(s string) string
	FindAllString(s string, n int) []string
	ReplaceAllString(src, repl string) string
	Split(s string, n int) []string
}

// compileRegex compiles a regex builtin's pattern: coregex outside a
// sandbox; inside one, Go regexp (RE2) with every byte it reads counted
// against Sandbox.MaxRegexWork. coregex has no work bound: one match of
// (?i)(?:\w+\s*){50}x against 4 KiB took 267 ms. multi: the builtin searches
// past the first match (find_all, replace, split).
func (interp *interpreter) compileRegex(pos Position, pattern string, multi bool) (regexEngine, error) {
	if interp.sandbox == nil {
		re, err := coregex.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return re, nil
	}
	return newBudgetRegex(interp, pos, pattern, multi)
}

// Compiling costs regex work too: Go compiles at up to ~225 ns per program
// instruction, and a work unit is ~6.4 ns.
const (
	regexCompileUnitsBase    = 64
	regexCompileUnitsPerInst = 40
)

func compileUnits(insts int) int {
	return sumEst(regexCompileUnitsBase, mulLen(regexCompileUnitsPerInst, insts))
}

// budgetRegex runs Go regexp over a counting reader: every byte a search
// reads costs the program's instruction count (RE2 steps every live
// instruction once per byte), so the work of a call is measured, not
// estimated. The multi-match operations are rebuilt from single searches
// because Go's FindAll/ReplaceAll/Split can rescan the rest of the input for
// every match — (?:a*b)|a over 16 KiB of "a" took 2.2 s — and only a
// per-search reader sees that. Results equal Go regexp's.
type budgetRegex struct {
	interp *interpreter
	pos    Position
	re     *regexp.Regexp
	insts  int
	// after (multi-match builtins only) searches from a later position with
	// the rune before it as context: \A(?s:.)(?s:.)*?(pattern). The lazy
	// prefix makes it the leftmost match, group 1 is the pattern, and \A, ^
	// and \b inside the pattern see the real previous rune.
	after      *regexp.Regexp
	afterInsts int
}

func newBudgetRegex(interp *interpreter, pos Position, pattern string, multi bool) (*budgetRegex, error) {
	tree, err := syntax.Parse(pattern, syntax.Perl) // what regexp.Compile parses, same error
	if err != nil {
		return nil, err
	}
	b := &budgetRegex{interp: interp, pos: pos, insts: treeInsts(tree)}
	interp.chargeRegexWork(pos, compileUnits(b.insts))
	if b.re, err = regexp.Compile(pattern); err != nil {
		return nil, err
	}
	if !multi {
		return b, nil
	}
	// Built from the parsed tree, never by appending to the pattern text: a
	// pattern ending inside \Q... would take an appended ")" as a literal.
	wrap := &syntax.Regexp{Op: syntax.OpConcat, Sub: []*syntax.Regexp{
		{Op: syntax.OpBeginText},
		{Op: syntax.OpAnyChar},
		{Op: syntax.OpStar, Flags: syntax.NonGreedy, Sub: []*syntax.Regexp{{Op: syntax.OpAnyChar}}},
		{Op: syntax.OpCapture, Cap: 1, Sub: []*syntax.Regexp{tree}},
	}}
	b.afterInsts = treeInsts(wrap)
	interp.chargeRegexWork(pos, compileUnits(b.afterInsts))
	if b.after, err = regexp.Compile(wrap.String()); err != nil {
		return nil, err
	}
	return b, nil
}

// treeInsts is the compiled instruction count of a parsed pattern
// (guardRegex bounded its size first).
func treeInsts(re *syntax.Regexp) int {
	prog, err := syntax.Compile(re.Simplify())
	if err != nil || len(prog.Inst) == 0 {
		return 1
	}
	return len(prog.Inst)
}

// chargeRegexWork counts units (instructions × bytes) against
// Sandbox.MaxRegexWork and stops the run past it.
func (interp *interpreter) chargeRegexWork(pos Position, units int) {
	max := interp.sandbox.MaxRegexWork
	if max <= 0 {
		return
	}
	if units < 0 || units > max-interp.regexWork {
		interp.regexWork = max
		interp.stop(LimitRegexWork, pos)
	}
	interp.regexWork += units
}

// countingReader feeds s from i to a regexp and charges every rune read.
type countingReader struct {
	s      string
	i      int
	insts  int
	interp *interpreter
	pos    Position
}

func (r *countingReader) ReadRune() (rune, int, error) {
	if r.i >= len(r.s) {
		return 0, 0, io.EOF
	}
	c, w := utf8.DecodeRuneInString(r.s[r.i:])
	r.i += w
	r.interp.chargeRegexWork(r.pos, r.insts*w)
	return c, w, nil
}

// search returns the submatch indices of the leftmost match in s that
// starts at or after pos, or nil — what Go's regexp computes internally when
// FindAll, ReplaceAll and Split resume at pos.
func (b *budgetRegex) search(s string, pos int) []int {
	if pos == 0 {
		b.interp.chargeRegexWork(b.pos, b.insts) // setup, even for an empty input
		return b.re.FindReaderSubmatchIndex(&countingReader{s: s, insts: b.insts, interp: b.interp, pos: b.pos})
	}
	if b.after == nil {
		panic("uddin: multi-match search on a single-match regex")
	}
	_, w := utf8.DecodeLastRuneInString(s[:pos])
	start := pos - w
	b.interp.chargeRegexWork(b.pos, b.afterInsts)
	loc := b.after.FindReaderSubmatchIndex(&countingReader{s: s, i: start, insts: b.afterInsts, interp: b.interp, pos: b.pos})
	if loc == nil {
		return nil
	}
	out := loc[2:]
	for i := range out {
		if out[i] >= 0 {
			out[i] += start
		}
	}
	return out
}

func (b *budgetRegex) MatchString(s string) bool {
	b.interp.chargeRegexWork(b.pos, b.insts)
	return b.re.MatchReader(&countingReader{s: s, insts: b.insts, interp: b.interp, pos: b.pos})
}

func (b *budgetRegex) FindString(s string) string {
	loc := b.search(s, 0)
	if loc == nil {
		return ""
	}
	return s[loc[0]:loc[1]]
}

// allMatches is regexp's allMatches over search: an empty match right after
// the previous match is skipped, and an empty match advances one rune.
func (b *budgetRegex) allMatches(s string, n int, deliver func([]int)) {
	end := len(s)
	for pos, i, prevMatchEnd := 0, 0, -1; i < n && pos <= end; {
		matches := b.search(s, pos)
		if len(matches) == 0 {
			break
		}
		accept := true
		if matches[1] == pos {
			if matches[0] == prevMatchEnd {
				accept = false
			}
			width := 0
			if pos < end {
				_, width = utf8.DecodeRuneInString(s[pos:])
			}
			if width > 0 {
				pos += width
			} else {
				pos = end + 1
			}
		} else {
			pos = matches[1]
		}
		prevMatchEnd = matches[1]
		if accept {
			deliver(matches)
			i++
		}
	}
}

func (b *budgetRegex) FindAllString(s string, n int) []string {
	if n < 0 {
		n = len(s) + 1
	}
	var result []string
	b.allMatches(s, n, func(m []int) {
		result = append(result, s[m[0]:m[1]])
	})
	return result
}

// ReplaceAllString is regexp's replaceAll over search; repl expands $1,
// ${name} like regexp.
func (b *budgetRegex) ReplaceAllString(src, repl string) string {
	lastMatchEnd, searchPos := 0, 0
	var buf []byte
	expand := strings.Contains(repl, "$")
	for searchPos <= len(src) {
		a := b.search(src, searchPos)
		if len(a) == 0 {
			break
		}
		buf = append(buf, src[lastMatchEnd:a[0]]...)
		if a[1] > lastMatchEnd || a[0] == 0 {
			if expand {
				buf = b.re.ExpandString(buf, repl, src, a)
			} else {
				buf = append(buf, repl...)
			}
		}
		lastMatchEnd = a[1]
		_, width := utf8.DecodeRuneInString(src[searchPos:])
		if searchPos+width > a[1] {
			searchPos += width
		} else if searchPos+1 > a[1] {
			searchPos++
		} else {
			searchPos = a[1]
		}
	}
	buf = append(buf, src[lastMatchEnd:]...)
	return string(buf)
}

// Split is regexp's Split over allMatches.
func (b *budgetRegex) Split(s string, n int) []string {
	if n == 0 {
		return nil
	}
	if len(b.re.String()) > 0 && len(s) == 0 {
		return []string{""}
	}
	limit := n
	if limit < 0 {
		limit = len(s) + 1
	}
	var matches [][]int
	b.allMatches(s, limit, func(m []int) {
		matches = append(matches, []int{m[0], m[1]})
	})
	parts := make([]string, 0, len(matches))
	beg, end := 0, 0
	for _, match := range matches {
		if n > 0 && len(parts) >= n-1 {
			break
		}
		end = match[0]
		if match[1] != 0 {
			parts = append(parts, s[beg:end])
		}
		beg = match[1]
	}
	if end != len(s) {
		parts = append(parts, s[beg:])
	}
	return parts
}
