package interpreter

import (
	"fmt"
	"strconv"
)

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
}

func (p *parser) next() {
	p.pos, p.tok, p.val = p.tokenizer.Next()
	if p.tok == ILLEGAL {
		p.error("%s", p.val)
	}
}

func (p *parser) error(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	panic(Error{p.pos, message})
}

func (p *parser) expect(tok Token) {
	if p.tok != tok {
		p.error("expected %s and not %s", tok, p.tok)
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
		statements = append(statements, p.statement())
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
	case IF:
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
	if p.tok == ASSIGN {
		pos = p.pos
		switch expr.(type) {
		case *Variable, *Subscript:
			p.next()
			value := p.expression()
			return &Assign{pos, expr, value}
		default:
			p.error("expected name, subscript, or dot expression on left side of =")
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
            // We'll collect statements until we hit END, ELSE, or CATCH
            statements := Block{}
            for p.tok != END && p.tok != ELSE && p.tok != CATCH && p.tok != EOF {
                statements = append(statements, p.statement())
            }

            // Only expect END if we're not at ELSE or CATCH
            if p.tok == END {
                p.expect(END)
            }

            return statements
        default:
            p.error("expected { or :, not %s", p.tok)
            return nil
	}
}

// if = IF LPAREN expression RPAREN THEN block |
//
//	IF LPAREN expression RPAREN THEN block ELSE block |
//	IF LPAREN expression RPAREN THEN block ELSE if
func (p *parser) if_() Statement {
	pos := p.pos
	p.expect(IF)
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
            case IF:
                elseBody = Block{p.if_()}
            default:
                p.error("expected { or : or if after else, not %s", p.tok)
            }
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
            p.error("expected { or :, not %s", p.tok)
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
		return &FunctionDefinition{pos, name, params, ellipsis, body}
	} else {
		params, ellipsis := p.params()
		body := p.block()
		expr := &FunctionExpression{pos, params, ellipsis, body}
		return &ExpressionStatement{pos, expr}
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
			p.error("expected , between parameters")
		}
		param := p.val
		p.expect(NAME)
		params = append(params, param)
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
		p.error("can only have ... after last parameter")
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
		expr = &Binary{pos, expr, op, right}
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
			p.error("expected : in ternary expression")
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

// multiply = negative ((TIMES | DIVIDE | MODULO) negative)*
func (p *parser) multiply() Expression {
	return p.binary(p.negative, TIMES, DIVIDE, MODULO)
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
                        p.error("expected , between arguments")
                    }
                    arg := p.expression()
                    args = append(args, arg)
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
                    p.error("can only have ... after last argument")
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
            return &Literal{pos, n}
        case STR:
            val := p.val
            pos := p.pos
            p.next()
            return &Literal{pos, val}
        case TRUE:
            pos := p.pos
            p.next()
            return &Literal{pos, true}
        case FALSE:
            pos := p.pos
            p.next()
            return &Literal{pos, false}
        case NULL:
            pos := p.pos
            p.next()
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
            p.error("expected expression, not %s", p.tok)
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
			p.error("expected , between list elements")
		}
		value := p.expression()
		values = append(values, value)
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
			p.error("expected , between object items")
		}
		key := p.mapKey()
		p.expect(COLON)
		value := p.expression()
		items = append(items, MapItem{key, value})
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
	p := parser{tokenizer: t}
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
	p := parser{tokenizer: t}
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
		p.error("expected string filename after import, got %s", p.tok)
	}
	filename := p.val
	p.next()
	return &Import{pos, filename}
}
