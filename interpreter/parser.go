package interpreter

import (
	"fmt"
	"strconv"
)

// Cache for commonly used error messages to reduce string formatting overhead
var commonErrorMessages = map[Token]string{
	LPAREN:   "expected (",
	RPAREN:   "expected )",
	LBRACE:   "expected {",
	RBRACE:   "expected }",
	LBRACKET: "expected [",
	RBRACKET: "expected ]",
	COMMA:    "expected ,",
	COLON:    "expected :",
	ASSIGN:   "expected =",
	EOF:      "unexpected end of file",
	THEN:     "expected then",
	END:      "expected end",
	ELSE:     "expected else",
	IN:       "expected in",
}

// Error is the error type returned by ParseExpression and ParseProgram when
// they encounter a syntax error. You can use this to get the location (line
// and column) of where the error occurred, as well as the error message.
type Error struct {
	Position Position
	Message  string
}

func (e Error) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
}

type parser struct {
	tokenizer *Tokenizer
	pos       Position
	tok       Token
	val       string
	nodePool  *ASTNodePool // Pool for AST nodes to reduce allocations
}

func (p *parser) next() {
	p.pos, p.tok, p.val = p.tokenizer.Next()
	if p.tok == ILLEGAL {
		p.error("%s", p.val)
	}
}

// Note: Node pool cleanup is handled automatically by Go's garbage collector
// No explicit cleanup needed as sync.Pool manages memory automatically

func (p *parser) error(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	panic(Error{p.pos, message})
}

func (p *parser) expect(tok Token) {
	if p.tok != tok {
		// Use cached error message if available, otherwise format dynamically
		if msg, exists := commonErrorMessages[tok]; exists {
			p.error("%s, but got %s", msg, p.tok)
		} else {
			p.error("expected %s, but got %s", tok, p.tok)
		}
	}
	p.next()
}

func (p *parser) matches(operators ...Token) bool {
	for _, operator := range operators {
		if p.tok == operator {
			return true
		}
	}
	return false
}

// program = statement*
func (p *parser) program() *Program {
	statements := p.statements(EOF)
	return &Program{statements}
}

func (p *parser) statements(end Token) Block {
	statements := Block{}
	for p.tok != end && p.tok != EOF {
		statements = SmartAppendBlock(statements, p.statement())
	}
	return statements
}

// statement = if | while | for | return | break | continue | import | fun | try | assign | expression
// assign    = NAME ASSIGN expression |
//
//	call subscript ASSIGN expression |
//	call dot ASSIGN expression
func (p *parser) statement() Statement {
	switch p.tok {
	case MEMO:
		return p.memoFun_()
	case IF, ELIF:
		return p.if_()
	case WHILE:
		return p.while()
	case FOR:
		return p.for_()
	case RETURN:
		return p.return_()
	case BREAK:
		return p.break_()
	case CONTINUE:
		return p.continue_()
	case IMPORT:
		return p.import_()
	case FUN:
		return p.fun_()
	case TRY:
		return p.tryCatch()
	}
	pos := p.pos
	expr := p.expression()
	if p.tok == ASSIGN || p.tok == PLUSEQUAL || p.tok == MINUSEQUAL ||
		p.tok == TIMESEQUAL || p.tok == DIVIDEEQUAL || p.tok == MODULOEQUAL {
		operator := p.tok
		pos = p.pos
		switch expr.(type) {
		case *Variable, *Subscript:
			p.next()
			value := p.expression()
			return &Assign{pos, expr, value, operator}
		default:
			p.error("invalid assignment target: only variables and array/object elements can be assigned to")
		}
	}
	return &ExpressionStatement{pos, expr}
}

// block = (LBRACE statement* RBRACE) | (COLON statement* END)
func (p *parser) block() Block {
	switch p.tok {
	case LBRACE:
		p.expect(LBRACE)
		body := p.statements(RBRACE)
		p.expect(RBRACE)
		return body
	case COLON:
		p.expect(COLON)
		// We'll collect statements until we hit END, ELSE, ELIF, or CATCH
		statements := Block{}
		for p.tok != END && p.tok != ELSE && p.tok != ELIF && p.tok != CATCH && p.tok != EOF {
			statements = SmartAppendBlock(statements, p.statement())
		}

		// Only expect END if we're not at ELSE, ELIF, or CATCH
		if p.tok == END {
			p.expect(END)
		}

		return statements
	default:
		p.error("expected ':' to start block (after 'then', 'while(...)', 'for(...)', 'fun(...)', or 'try'), got %s", p.tok)
		return nil
	}
}

// if = IF LPAREN expression RPAREN THEN block |
//
//	IF LPAREN expression RPAREN THEN block ELSE block |
//	IF LPAREN expression RPAREN THEN block ELSE if
func (p *parser) if_() Statement {
	pos := p.pos
	if p.tok == ELIF {
		p.expect(ELIF)
	} else {
		p.expect(IF)
	}
	p.expect(LPAREN) // Require opening parenthesis
	condition := p.expression()
	p.expect(RPAREN) // Require closing parenthesis
	p.expect(THEN)   // Require 'then' keyword

	body := p.block()

	var elseBody Block
	if p.tok == ELSE {
		p.next()
		switch p.tok {
		case LBRACE, COLON:
			elseBody = p.block()
		case IF, ELIF:
			elseBody = Block{p.if_()}
		default:
			p.error("expected ':' or 'if' after 'else' keyword, got %s", p.tok)
		}
	} else if p.tok == ELIF {
		elseBody = Block{p.if_()}
	}

	return &If{pos, condition, body, elseBody}
}

// while = WHILE LPAREN expression RPAREN block
func (p *parser) while() Statement {
	pos := p.pos
	p.expect(WHILE)
	p.expect(LPAREN) // Require opening parenthesis
	condition := p.expression()
	p.expect(RPAREN) // Require closing parenthesis
	body := p.block()
	return &While{pos, condition, body}
}

// for = FOR LPAREN NAME IN expression RPAREN block
func (p *parser) for_() Statement {
	pos := p.pos
	p.expect(FOR)
	p.expect(LPAREN) // Require opening parenthesis
	name := p.val
	p.expect(NAME)
	p.expect(IN)
	iterable := p.expression()
	p.expect(RPAREN) // Require closing parenthesis
	body := p.block()
	return &For{pos, name, iterable, body}
}

// tryCatch = TRY block CATCH LPAREN NAME RPAREN block
func (p *parser) tryCatch() Statement {
	pos := p.pos
	p.expect(TRY)

	// For the try block, we don't expect END
	var tryBlock Block
	switch p.tok {
	case LBRACE:
		p.expect(LBRACE)
		tryBlock = p.statements(RBRACE)
		p.expect(RBRACE)
	case COLON:
		p.expect(COLON)
		tryBlock = p.statements(CATCH)
	default:
		p.error("expected ':' after 'try' keyword to start block, got %s", p.tok)
		return nil
	}

	p.expect(CATCH)
	p.expect(LPAREN) // Require opening parenthesis
	errVar := p.val
	p.expect(NAME)
	p.expect(RPAREN) // Require closing parenthesis
	catchBlock := p.block()

	return &TryCatch{pos, tryBlock, errVar, catchBlock}
}

// return = RETURN expression
func (p *parser) return_() Statement {
	pos := p.pos
	p.expect(RETURN)
	result := p.expression()
	return &Return{pos, result}
}

// fun = FUN NAME params block |
//
//	FUN params block
func (p *parser) fun_() Statement {
	pos := p.pos
	p.expect(FUN)
	if p.tok == NAME {
		name := p.val
		p.next()
		params, ellipsis := p.params()
		body := p.block()
		return &FunctionDefinition{pos, name, params, ellipsis, body, false}
	} else {
		params, ellipsis := p.params()
		body := p.block()
		expr := &FunctionExpression{pos, params, ellipsis, body}
		return &ExpressionStatement{pos, expr}
	}
}

// memoFun_ parses a memoized function definition: memo fun name(params) body end
func (p *parser) memoFun_() Statement {
	pos := p.pos
	p.expect(MEMO)
	p.expect(FUN)
	if p.tok == NAME {
		name := p.val
		p.next()
		params, ellipsis := p.params()
		body := p.block()
		return &FunctionDefinition{pos, name, params, ellipsis, body, true}
	} else {
		p.error("expected function name after 'memo fun'")
		return nil
	}
}

// params = LPAREN RPAREN |
//
//	LPAREN NAME (COMMA NAME)* ELLIPSIS? COMMA? RPAREN |
func (p *parser) params() ([]string, bool) {
	p.expect(LPAREN)
	params := []string{}
	gotComma := true
	gotEllipsis := false
	for p.tok != RPAREN && p.tok != EOF && !gotEllipsis {
		if !gotComma {
			p.error("missing comma ',' between function parameters")
		}
		param := p.val
		p.expect(NAME)
		params = SmartAppendString(params, param)
		if p.tok == ELLIPSIS {
			gotEllipsis = true
			p.next()
		}
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	if p.tok != RPAREN && gotEllipsis {
		p.error("variadic parameter '...' must be the last parameter in function definition")
	}
	p.expect(RPAREN)
	return params, gotEllipsis
}

func (p *parser) binary(parseFunc func() Expression, operators ...Token) Expression {
	expr := parseFunc()
	for p.matches(operators...) {
		op := p.tok
		pos := p.pos
		p.next()
		right := parseFunc()
		// Use node pool for Binary allocation
		if binary := p.nodePool.GetBinary(); binary != nil {
			binary.pos = pos
			binary.Left = expr
			binary.Operator = op
			binary.Right = right
			expr = binary
		} else {
			expr = &Binary{pos, expr, op, right}
		}
	}
	return expr
}

// expression = xor (OR xor)*
func (p *parser) expression() Expression {
	return p.binary(p.xor, OR)
}

// xor = and (XOR and)*
func (p *parser) xor() Expression {
	return p.binary(p.and, XOR)
}

// and = not (AND not)*
func (p *parser) and() Expression {
	return p.binary(p.not, AND)
}

// not = NOT not | equality
func (p *parser) not() Expression {
	if p.tok == NOT {
		pos := p.pos
		p.next()
		operand := p.not()
		return &Unary{pos, NOT, operand}
	}
	return p.ternary()
}

// ternary = equality (QUESTION expression COLON expression)?
func (p *parser) ternary() Expression {
	expr := p.equality()

	if p.tok == QUESTION {
		pos := p.pos
		p.next() // consume QUESTION
		trueExpr := p.expression()

		if p.tok != COLON {
			p.error("missing ':' in ternary expression - expected format: condition ? true_value : false_value")
		}
		p.next() // consume COLON

		falseExpr := p.expression()
		expr = &Ternary{pos, expr, trueExpr, falseExpr}
	}

	return expr
}

// equality = comparison ((EQUAL | NOTEQUAL) comparison)*
func (p *parser) equality() Expression {
	return p.binary(p.comparison, EQUAL, NOTEQUAL)
}

// comparison = addition ((LT | LTE | GT | GTE | IN) addition)*
func (p *parser) comparison() Expression {
	return p.binary(p.addition, LT, LTE, GT, GTE, IN)
}

// addition = multiply ((PLUS | MINUS) multiply)*
func (p *parser) addition() Expression {
	return p.binary(p.multiply, PLUS, MINUS)
}

// multiply = power ((TIMES | DIVIDE | MODULO) power)*
func (p *parser) multiply() Expression {
	return p.binary(p.power, TIMES, DIVIDE, MODULO)
}

// power = negative (POWER power)? (right-associative)
func (p *parser) power() Expression {
	expr := p.negative()
	if p.tok == POWER {
		op := p.tok
		pos := p.pos
		p.next()
		right := p.power() // Right-associative: recurse to power, not negative
		expr = &Binary{pos, expr, op, right}
	}
	return expr
}

// negative = MINUS negative | call
func (p *parser) negative() Expression {
	if p.tok == MINUS {
		pos := p.pos
		p.next()
		operand := p.negative()
		return &Unary{pos, MINUS, operand}
	}
	return p.call()
}

// call      = primary (args | subscript | dot)*
// args      = LPAREN RPAREN |
//
//	LPAREN expression (COMMA expression)* ELLIPSIS? COMMA? RPAREN)
//
// subscript = LBRACKET expression RBRACKET
// dot       = DOT NAME
func (p *parser) call() Expression {
	expr := p.primary()
	for p.matches(LPAREN, LBRACKET, DOT) {
		switch p.tok {
		case LPAREN:
			pos := p.pos
			p.next()
			args := []Expression{}
			gotComma := true
			gotEllipsis := false
			for p.tok != RPAREN && p.tok != EOF && !gotEllipsis {
				if !gotComma {
					p.error("missing comma ',' between function arguments")
				}
				arg := p.expression()
				args = SmartAppendExpression(args, arg)
				if p.tok == ELLIPSIS {
					gotEllipsis = true
					p.next()
				}
				if p.tok == COMMA {
					gotComma = true
					p.next()
				} else {
					gotComma = false
				}
			}
			if p.tok != RPAREN && gotEllipsis {
				p.error("variadic argument '...' must be the last argument in function call")
			}
			p.expect(RPAREN)
			expr = &Call{pos, expr, args, gotEllipsis}
		case LBRACKET:
			pos := p.pos
			p.next()
			subscript := p.expression()
			p.expect(RBRACKET)
			expr = &Subscript{pos, expr, subscript}
		default:
			pos := p.pos
			p.next()
			subscript := &Literal{p.pos, p.val}
			p.expect(NAME)
			expr = &Subscript{pos, expr, subscript}
		}
	}
	return expr
}

// primary = NAME | INT | STR | TRUE | FALSE | NIL | list | map |
//
//	FUNC params block |
//	LPAREN expression RPAREN
func (p *parser) primary() Expression {
	switch p.tok {
	case NAME:
		name := p.val
		pos := p.pos
		p.next()
		// Use node pool for Variable allocation
		if variable := p.nodePool.GetVariable(); variable != nil {
			variable.pos = pos
			variable.Name = name
			return variable
		}
		return &Variable{pos, name}
	case INT:
		val := p.val
		pos := p.pos
		p.next()
		n, err := strconv.Atoi(val)
		if err != nil {
			// Tokenizer should never give us this
			panic(fmt.Sprintf("tokenizer gave INT token that isn't an int: %s", val))
		}
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = n
			return literal
		}
		return &Literal{pos, n}
	case FLOAT:
		val := p.val
		pos := p.pos
		p.next()
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			// Tokenizer should never give us this
			panic(fmt.Sprintf("tokenizer gave FLOAT token that isn't a float: %s", val))
		}
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = n
			return literal
		}
		return &Literal{pos, n}
	case STR:
		val := p.val
		pos := p.pos
		p.next()
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = val
			return literal
		}
		return &Literal{pos, val}
	case TRUE:
		pos := p.pos
		p.next()
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = true
			return literal
		}
		return &Literal{pos, true}
	case FALSE:
		pos := p.pos
		p.next()
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = false
			return literal
		}
		return &Literal{pos, false}
	case NULL:
		pos := p.pos
		p.next()
		// Use node pool for Literal allocation
		if literal := p.nodePool.GetLiteral(); literal != nil {
			literal.pos = pos
			literal.Value = nil
			return literal
		}
		return &Literal{pos, nil}
	case LBRACKET:
		return p.list()
	case LBRACE:
		return p.map_()
	case FUN:
		pos := p.pos
		p.next()
		args, ellipsis := p.params()
		body := p.block()
		return &FunctionExpression{pos, args, ellipsis, body}
	case LPAREN:
		p.next()
		expr := p.expression()
		p.expect(RPAREN)
		return expr
	default:
		p.error("unexpected token %s - expected a value (number, string, identifier, '(', '[', '{', 'fun', etc.)", p.tok)
		return nil
	}
}

// list = LBRACKET RBRACKET |
//
//	LBRACKET expression (COMMA expression)* COMMA? RBRACKET
func (p *parser) list() Expression {
	pos := p.pos
	p.expect(LBRACKET)
	values := []Expression{}
	gotComma := true
	for p.tok != RBRACKET && p.tok != EOF {
		if !gotComma {
			p.error("missing comma ',' between array elements")
		}
		value := p.expression()
		values = SmartAppendExpression(values, value)
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	p.expect(RBRACKET)
	return &List{pos, values}
}

// map = LBRACE RBRACE |
//
//	LBRACE expression COLON expression
//	       (COMMA expression COLON expression)* COMMA? RBRACE
func (p *parser) map_() Expression {
	pos := p.pos
	p.expect(LBRACE)
	items := []MapItem{}
	gotComma := true
	for p.tok != RBRACE && p.tok != EOF {
		if !gotComma {
			p.error("missing comma ',' between object properties")
		}
		key := p.mapKey()
		p.expect(COLON)
		value := p.expression()
		items = SmartAppendMapItem(items, MapItem{key, value})
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	p.expect(RBRACE)
	return &Map{pos, items}
}

// mapKey parses a map key, which can be:
// - A string literal: "key" or 'key'
// - An identifier: key (converted to string literal)
// - Any other expression
func (p *parser) mapKey() Expression {
	switch p.tok {
	case NAME:
		// Convert bare identifier to string literal
		name := p.val
		pos := p.pos
		p.next()
		return &Literal{pos, name}
	default:
		// For quoted strings and other expressions
		return p.expression()
	}
}

// ParseExpression parses a single expression into an Expression interface
// (can be one of many expression types). If the expression parses correctly,
// return an Expression and nil. If there's a syntax error, return nil and
// a parser.Error value.
func ParseExpression(input []byte) (e Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert to parser.Error or re-panic
			err = r.(Error)
		}
	}()
	t := NewTokenizer(input)
	p := parser{
		tokenizer: t,
		nodePool:  GetGlobalASTNodePool(),
	}
	p.next()
	return p.expression(), nil
}

// ParseProgram parses an entire program and returns a *Program (which is
// basically a list of statements). If the program parses correctly, return
// a *Program and nil. If there's a syntax error, return nil and a
// parser.Error value.
func ParseProgram(input []byte) (prog *Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert to parser.Error or re-panic
			err = r.(Error)
		}
	}()
	t := NewTokenizer(input)
	p := parser{
		tokenizer: t,
		nodePool:  GetGlobalASTNodePool(),
	}
	p.next()
	return p.program(), nil
}

// break = BREAK
func (p *parser) break_() Statement {
	pos := p.pos
	p.expect(BREAK)
	return &Break{pos}
}

// continue = CONTINUE
func (p *parser) continue_() Statement {
	pos := p.pos
	p.expect(CONTINUE)
	return &Continue{pos}
}

// import = IMPORT STR
func (p *parser) import_() Statement {
	pos := p.pos
	p.expect(IMPORT)
	if p.tok != STR {
		p.error("import statement requires a string filename, got %s - example: import \"filename.din\"", p.tok)
	}
	filename := p.val
	p.next()
	return &Import{pos, filename}
}
