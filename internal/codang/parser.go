package codang

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser is a recursive descent parser turning tokens into an AST.
type Parser struct {
	toks []Token
	pos  int
}

// Parse parses tokens into a Program AST.
func Parse(src string) (*Program, error) {
	lex := NewLexer(src)
	toks, err := lex.Tokenize()
	if err != nil {
		return nil, err
	}
	p := &Parser{toks: toks}
	return p.parseProgram()
}

func (p *Parser) parseProgram() (*Program, error) {
	prog := &Program{Metadata: map[string]string{}}

	for !p.atEnd() {
		// Skip newlines and comments
		for p.checkAny(TokNewline, TokComment) {
			t := p.advance()
			if t.Type == TokComment {
				// Check if it's a metadata comment: # @key value
				val := strings.TrimSpace(strings.TrimPrefix(t.Value, "#"))
				if strings.HasPrefix(val, "@") {
					key, value := parseMetadataLine(val)
					prog.Metadata[key] = value
				}
			}
		}

		if p.check(TokMetadata) {
			t := p.advance()
			key, value := parseMetadataLine(t.Value)
			prog.Metadata[key] = value
			continue
		}

		if p.atEnd() {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}

	return prog, nil
}

func parseMetadataLine(s string) (key, value string) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "@"))
	parts := strings.SplitN(s, " ", 2)
	key = parts[0]
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
		// Strip quotes
		value = strings.Trim(value, "\"'")
	}
	return
}

// --- Statement parsing ---

func (p *Parser) parseStatement() (Node, error) {
	// Skip newlines
	for p.check(TokNewline) {
		p.advance()
	}
	if p.atEnd() || p.check(TokRBrace) {
		return nil, nil
	}

	// func definition
	if p.checkKeyword("func") {
		return p.parseFuncDef()
	}

	// if statement
	if p.checkKeyword("if") {
		return p.parseIf()
	}

	// return statement
	if p.checkKeyword("return") {
		p.advance()
		if p.check(TokNewline) || p.check(TokEOF) || p.check(TokRBrace) {
			return &ReturnStmt{Value: &NilLit{}}, nil
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.consumeNewline()
		return &ReturnStmt{Value: val}, nil
	}

	// assignment or expression
	ident := p.peek()
	if ident.Type == TokIdent && p.peekN(1).Type == TokAssign {
		return p.parseAssign()
	}

	// expression statement
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.consumeNewline()
	return &ExprStmt{Expr: expr}, nil
}

func (p *Parser) parseAssign() (Node, error) {
	name := p.advance().Value
	p.advance() // consume =
	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.consumeNewline()
	return &AssignStmt{Name: name, Value: value}, nil
}

func (p *Parser) parseFuncDef() (Node, error) {
	p.advance() // consume 'func'
	name := p.advance().Value
	// params
	var params []string
	if p.check(TokLParen) {
		p.advance()
		for !p.check(TokRParen) && !p.atEnd() {
			if p.check(TokIdent) {
				params = append(params, p.advance().Value)
			} else if p.check(TokComma) {
				p.advance()
			} else {
				p.advance()
			}
		}
		if p.check(TokRParen) {
			p.advance()
		}
	}
	// optional colon form: func name(args): body
	if p.check(TokColon) {
		p.advance()
		p.consumeNewline()
		// Single expression return
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.consumeNewline()
		return &FuncDef{Name: name, Params: params, Body: []Node{&ReturnStmt{Value: expr}}}, nil
	}
	// Block form: indent-based (we use braces or newlines)
	p.consumeNewline()
	var body []Node
	for !p.atEnd() && !p.check(TokRBrace) {
		// Skip blank lines
		for p.check(TokNewline) {
			p.advance()
		}
		if p.atEnd() || p.check(TokRBrace) {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	return &FuncDef{Name: name, Params: params, Body: body}, nil
}

func (p *Parser) parseIf() (Node, error) {
	p.advance() // consume 'if'
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	// Optional colon
	if p.check(TokColon) {
		p.advance()
	}
	p.consumeNewline()

	var thenBody []Node
	for !p.atEnd() && !p.checkKeyword("else") && !p.checkKeyword("elif") {
		for p.check(TokNewline) {
			p.advance()
		}
		if p.atEnd() || p.checkKeyword("else") || p.checkKeyword("elif") {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		}
	}

	var elseBody []Node
	if p.checkKeyword("else") {
		p.advance()
		if p.check(TokColon) {
			p.advance()
		}
		p.consumeNewline()
		for !p.atEnd() {
			for p.check(TokNewline) {
				p.advance()
			}
			if p.atEnd() {
				break
			}
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				elseBody = append(elseBody, stmt)
			}
		}
	}

	return &IfStmt{Cond: cond, Then: thenBody, ElseBody: elseBody}, nil
}

// --- Expression parsing (precedence climbing) ---

func (p *Parser) parseExpr() (Node, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.checkKeyword("or") {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: "or", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseCompare()
	if err != nil {
		return nil, err
	}
	for p.checkKeyword("and") {
		p.advance()
		right, err := p.parseCompare()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: "and", Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseCompare() (Node, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for p.check(TokOp) {
		op := p.peek().Value
		if op == "<" || op == ">" || op == "<=" || op == ">=" || op == "==" || op == "!=" {
			p.advance()
			right, err := p.parseAddSub()
			if err != nil {
				return nil, err
			}
			left = &BinaryOp{Op: op, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseAddSub() (Node, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for p.check(TokOp) {
		op := p.peek().Value
		if op == "+" || op == "-" {
			p.advance()
			right, err := p.parseMulDiv()
			if err != nil {
				return nil, err
			}
			left = &BinaryOp{Op: op, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseMulDiv() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.check(TokOp) {
		op := p.peek().Value
		if op == "*" || op == "/" || op == "%" {
			p.advance()
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			left = &BinaryOp{Op: op, Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.check(TokOp) && (p.peek().Value == "-" || p.peek().Value == "!") {
		op := p.advance().Value
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOp{Op: op, Operand: operand}, nil
	}
	if p.checkKeyword("not") {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOp{Op: "not", Operand: operand}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		if p.check(TokDot) {
			// method call: expr.method(args)
			p.advance()
			methodName := p.advance().Value
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			expr = &MethodCall{Target: expr, Method: methodName, Args: args}
		} else if p.check(TokLParen) {
			// function call: expr(args) — only valid if expr is an ident
			if ident, ok := expr.(*Ident); ok {
				args, err := p.parseArgs()
				if err != nil {
					return nil, err
				}
				expr = &CallExpr{Name: ident.Name, Args: args}
			} else {
				break
			}
		} else if p.check(TokLBrack) {
			// index: expr[index]
			p.advance()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if p.check(TokRBrack) {
				p.advance()
			}
			expr = &Index{Target: expr, Index: idx}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) parseArgs() ([]Node, error) {
	var args []Node
	if !p.check(TokLParen) {
		return args, nil
	}
	p.advance() // consume (
	for !p.check(TokRParen) && !p.atEnd() {
		if p.check(TokComma) {
			p.advance()
			continue
		}
		if p.check(TokKeyword) && p.peek().Value == "nil" {
			p.advance()
			args = append(args, nil)
			continue
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)
	}
	if p.check(TokRParen) {
		p.advance()
	}
	return args, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	t := p.peek()
	if t == nil || t.Type == TokEOF {
		return nil, fmt.Errorf("parser: unexpected end of input")
	}

	switch t.Type {
	case TokNumber:
		p.advance()
		f, _ := strconv.ParseFloat(t.Value, 64)
		return &NumberLit{Value: f}, nil

	case TokString:
		p.advance()
		return &StringLit{Value: t.Value}, nil

	case TokKeyword:
		switch t.Value {
		case "true":
			p.advance()
			return &BoolLit{Value: true}, nil
		case "false":
			p.advance()
			return &BoolLit{Value: false}, nil
		case "nil":
			p.advance()
			return &NilLit{}, nil
		}
		return nil, fmt.Errorf("parser: unexpected keyword '%s' at line %d", t.Value, t.Line)

	case TokIdent:
		p.advance()
		return &Ident{Name: t.Value}, nil

	case TokLParen:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.check(TokRParen) {
			p.advance()
		}
		return expr, nil

	case TokLBrack:
		return p.parseArray()

	case TokLBrace:
		// {a b, c d} inline polymeter — treat as array of strings
		return p.parseBraceArray()
	}

	return nil, fmt.Errorf("parser: unexpected token %s at line %d", t.Value, t.Line)
}

func (p *Parser) parseArray() (Node, error) {
	p.advance() // consume [
	var elements []Node
	for !p.check(TokRBrack) && !p.atEnd() {
		if p.check(TokComma) || p.check(TokNewline) {
			p.advance()
			continue
		}
		if p.check(TokNumber) {
			// [n, pattern] form → number + expr pair
			nTok := p.advance()
			n, _ := strconv.ParseFloat(nTok.Value, 64)
			if p.check(TokComma) {
				p.advance()
			}
			pat, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elements = append(elements, &NumberLit{Value: n})
			elements = append(elements, pat)
		} else {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)
		}
		if p.check(TokComma) || p.check(TokNewline) {
			p.advance()
		}
	}
	if p.check(TokRBrack) {
		p.advance()
	}
	return &ArrayLit{Elements: elements}, nil
}

func (p *Parser) parseBraceArray() (Node, error) {
	p.advance() // consume {
	var elements []Node
	for !p.check(TokRBrace) && !p.atEnd() {
		if p.check(TokComma) || p.check(TokNewline) {
			p.advance()
			continue
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}
	if p.check(TokRBrace) {
		p.advance()
	}
	// Check for %subdivision
	if p.check(TokOp) && p.peek().Value == "%" {
		p.advance()
		if p.check(TokNumber) {
			nTok := p.advance()
			n, _ := strconv.ParseFloat(nTok.Value, 64)
			elements = append(elements, &NumberLit{Value: n})
		}
	}
	return &ArrayLit{Elements: elements}, nil
}

// --- Token helpers ---

func (p *Parser) peek() *Token {
	if p.pos < len(p.toks) {
		return &p.toks[p.pos]
	}
	return nil
}

func (p *Parser) peekN(n int) *Token {
	if p.pos+n < len(p.toks) {
		return &p.toks[p.pos+n]
	}
	return nil
}

func (p *Parser) advance() Token {
	if p.pos < len(p.toks) {
		t := p.toks[p.pos]
		p.pos++
		return t
	}
	return Token{Type: TokEOF}
}

func (p *Parser) check(typ TokenType) bool {
	t := p.peek()
	return t != nil && t.Type == typ
}

func (p *Parser) checkAny(typs ...TokenType) bool {
	t := p.peek()
	if t == nil {
		return false
	}
	for _, typ := range typs {
		if t.Type == typ {
			return true
		}
	}
	return false
}

func (p *Parser) checkKeyword(kw string) bool {
	t := p.peek()
	return t != nil && t.Type == TokKeyword && t.Value == kw
}

func (p *Parser) atEnd() bool {
	t := p.peek()
	return t == nil || t.Type == TokEOF
}

func (p *Parser) consumeNewline() {
	for p.check(TokNewline) {
		p.advance()
	}
}
