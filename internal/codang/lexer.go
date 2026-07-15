package codang

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType identifies the kind of a token.
type TokenType int

const (
	TokEOF      TokenType = iota
	TokIdent              // name
	TokNumber             // 42, 3.14
	TokString             // "..."  (mini-notation lives here)
	TokOp                 // + - * / etc
	TokAssign             // =
	TokLParen             // (
	TokRParen             // )
	TokLBrack             // [
	TokRBrack             // ]
	TokLBrace             // {
	TokRBrace             // }
	TokComma              // ,
	TokColon              // :
	TokDot                // .
	TokArrow              // -> (used in return type hints, optional)
	TokNewline            // \n
	TokComment            // # ...
	TokMetadata           // @bpm 120
	TokKeyword            // func, return, if, else, true, false, nil
)

// Token is a single lexical unit from a .cdc file.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("{%s '%s' L%d:C%d}", tokTypeName(t.Type), t.Value, t.Line, t.Col)
}

func tokTypeName(t TokenType) string {
	switch t {
	case TokEOF:
		return "EOF"
	case TokIdent:
		return "Ident"
	case TokNumber:
		return "Number"
	case TokString:
		return "String"
	case TokOp:
		return "Op"
	case TokAssign:
		return "Assign"
	case TokLParen:
		return "LParen"
	case TokRParen:
		return "RParen"
	case TokLBrack:
		return "LBrack"
	case TokRBrack:
		return "RBrack"
	case TokLBrace:
		return "LBrace"
	case TokRBrace:
		return "RBrace"
	case TokComma:
		return "Comma"
	case TokColon:
		return "Colon"
	case TokDot:
		return "Dot"
	case TokArrow:
		return "Arrow"
	case TokNewline:
		return "Newline"
	case TokComment:
		return "Comment"
	case TokMetadata:
		return "Metadata"
	case TokKeyword:
		return "Keyword"
	}
	return "?"
}

// Lexer tokenizes a Codang source string.
type Lexer struct {
	src    []rune
	pos    int
	line   int
	col    int
	tokens []Token
}

// NewLexer creates a lexer from a source string.
func NewLexer(src string) *Lexer {
	return &Lexer{src: []rune(src), line: 1, col: 1}
}

// Tokenize returns all tokens from the source.
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.src) {
		c := l.src[l.pos]

		switch {
		case c == '\n':
			l.emit(TokNewline, "\n")
			l.line++
			l.col = 1
			l.pos++
		case c == '\r':
			l.pos++
		case unicode.IsSpace(c):
			l.pos++
			l.col++
		case c == '#':
			l.lexComment()
		case c == '@':
			l.lexMetadata()
		case c == '"':
			l.lexString()
		case c == '.' && l.isNumberStart():
			l.lexNumber()
		case l.isNumberStart():
			l.lexNumber()
		case c == '-' && l.peek() == '>':
			l.emit(TokArrow, "->")
			l.pos += 2
			l.col += 2
		case c == '=':
			l.emit(TokAssign, "=")
			l.pos++
			l.col++
		case c == '(':
			l.emit(TokLParen, "(")
			l.pos++
			l.col++
		case c == ')':
			l.emit(TokRParen, ")")
			l.pos++
			l.col++
		case c == '[':
			l.emit(TokLBrack, "[")
			l.pos++
			l.col++
		case c == ']':
			l.emit(TokRBrack, "]")
			l.pos++
			l.col++
		case c == '{':
			l.emit(TokLBrace, "{")
			l.pos++
			l.col++
		case c == '}':
			l.emit(TokRBrace, "}")
			l.pos++
			l.col++
		case c == ',':
			l.emit(TokComma, ",")
			l.pos++
			l.col++
		case c == ':':
			l.emit(TokColon, ":")
			l.pos++
			l.col++
		case c == '.':
			l.emit(TokDot, ".")
			l.pos++
			l.col++
		case isOp(c):
			l.lexOp()
		case isIdentStart(c):
			l.lexIdent()
		default:
			return nil, fmt.Errorf("lexer: unexpected character '%c' at line %d col %d", c, l.line, l.col)
		}
	}
	l.tokens = append(l.tokens, Token{Type: TokEOF, Line: l.line, Col: l.col})
	return l.tokens, nil
}

func (l *Lexer) emit(typ TokenType, val string) {
	l.tokens = append(l.tokens, Token{Type: typ, Value: val, Line: l.line, Col: l.col})
}

func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.src) {
		return l.src[l.pos+1]
	}
	return 0
}

func (l *Lexer) isNumberStart() bool {
	c := l.src[l.pos]
	if c >= '0' && c <= '9' {
		return true
	}
	if c == '.' && l.pos+1 < len(l.src) && l.src[l.pos+1] >= '0' && l.src[l.pos+1] <= '9' {
		return true
	}
	return false
}

func (l *Lexer) lexNumber() {
	start := l.pos
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if (c >= '0' && c <= '9') || c == '.' {
			l.pos++
			l.col++
		} else {
			break
		}
	}
	l.emit(TokNumber, string(l.src[start:l.pos]))
}

func (l *Lexer) lexString() {
	l.pos++ // skip opening quote
	l.col++
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '"' {
		if l.src[l.pos] == '\\' && l.pos+1 < len(l.src) {
			l.pos += 2
			l.col += 2
		} else {
			l.pos++
			l.col++
		}
	}
	val := string(l.src[start:l.pos])
	// Unescape simple sequences
	val = strings.ReplaceAll(val, "\\\"", "\"")
	val = strings.ReplaceAll(val, "\\n", "\n")
	val = strings.ReplaceAll(val, "\\\\", "\\")
	l.emit(TokString, val)
	if l.pos < len(l.src) {
		l.pos++ // skip closing quote
		l.col++
	}
}

func (l *Lexer) lexComment() {
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
		l.col++
	}
	l.emit(TokComment, string(l.src[start:l.pos]))
}

func (l *Lexer) lexMetadata() {
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
		l.col++
	}
	l.emit(TokMetadata, string(l.src[start:l.pos]))
}

func (l *Lexer) lexIdent() {
	start := l.pos
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.pos++
		l.col++
	}
	val := string(l.src[start:l.pos])
	// Check keywords
	switch val {
	case "func", "return", "if", "else", "elif", "true", "false", "nil", "for", "while", "and", "or", "not":
		l.emit(TokKeyword, val)
	default:
		l.emit(TokIdent, val)
	}
}

func (l *Lexer) lexOp() {
	start := l.pos
	for l.pos < len(l.src) && isOp(l.src[l.pos]) {
		l.pos++
		l.col++
	}
	l.emit(TokOp, string(l.src[start:l.pos]))
}

func isOp(c rune) bool {
	switch c {
	case '+', '-', '*', '/', '<', '>', '!', '|', '^', '%', '&', '=':
		return true
	}
	return false
}

func isIdentStart(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}

func isIdentPart(c rune) bool {
	return c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c)
}
