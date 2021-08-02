package actionlint

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// TokenKind is kind of token.
type TokenKind int

const (
	// TokenKindUnknown is a default value of token as unknown token value.
	TokenKindUnknown TokenKind = iota
	// TokenKindEnd is a token for end of token sequence. Sequence without this
	// token means invalid.
	TokenKindEnd
	// TokenKindIdent is a token for identifier.
	TokenKindIdent
	// TokenKindString is a token for string literals.
	TokenKindString
	// TokenKindInt is a token for integers including hex integers.
	TokenKindInt
	// TokenKindFloat is a token for float numbers.
	TokenKindFloat
	// TokenKindLeftParen is a token for '('.
	TokenKindLeftParen
	// TokenKindRightParen is a token for ')'.
	TokenKindRightParen
	// TokenKindLeftBracket is a token for '['.
	TokenKindLeftBracket
	// TokenKindRightBracket is a token for ']'.
	TokenKindRightBracket
	// TokenKindDot is a token for '.'.
	TokenKindDot
	// TokenKindNot is a token for '!'.
	TokenKindNot
	// TokenKindLess is a token for '<'.
	TokenKindLess
	// TokenKindLessEq is a token for '<='.
	TokenKindLessEq
	// TokenKindGreater is a token for '>'.
	TokenKindGreater
	// TokenKindGreaterEq is a token for '>='.
	TokenKindGreaterEq
	// TokenKindEq is a token for '=='.
	TokenKindEq
	// TokenKindNotEq is a token for '!='.
	TokenKindNotEq
	// TokenKindAnd is a token for '&&'.
	TokenKindAnd
	// TokenKindOr is a token for '||'.
	TokenKindOr
	// TokenKindStar is a token for '*'.
	TokenKindStar
	// TokenKindComma is a token for ','.
	TokenKindComma
)

func (t TokenKind) String() string {
	switch t {
	case TokenKindUnknown:
		return "UNKNOWN"
	case TokenKindEnd:
		return "END"
	case TokenKindIdent:
		return "IDENT"
	case TokenKindString:
		return "STRING"
	case TokenKindInt:
		return "INTEGER"
	case TokenKindFloat:
		return "FLOAT"
	case TokenKindLeftParen:
		return "("
	case TokenKindRightParen:
		return ")"
	case TokenKindLeftBracket:
		return "["
	case TokenKindRightBracket:
		return "]"
	case TokenKindDot:
		return "."
	case TokenKindNot:
		return "!"
	case TokenKindLess:
		return "<"
	case TokenKindLessEq:
		return "<="
	case TokenKindGreater:
		return ">"
	case TokenKindGreaterEq:
		return ">="
	case TokenKindEq:
		return "=="
	case TokenKindNotEq:
		return "!="
	case TokenKindAnd:
		return "&&"
	case TokenKindOr:
		return "||"
	case TokenKindStar:
		return "*"
	case TokenKindComma:
		return ","
	default:
		panic("unreachable")
	}
}

// Token is a token lexed from expression syntax. For more details, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type Token struct {
	// Kind is kind of the token.
	Kind TokenKind
	// Value is string representation of the token.
	Value string
	// Offset is byte offset of token string starting.
	Offset int
	// Line is line number of start position of the token. Note that this value is 1-based.
	Line int
	// Column is column number of start position of the token. Note that this value is 1-based.
	Column int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s:%d:%d:%d", t.Kind.String(), t.Line, t.Column, t.Offset)
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t'
}

func isAlpha(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
}

func isNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func isHexNum(r rune) bool {
	return isNum(r) || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F'
}

func isAlnum(r rune) bool {
	return isAlpha(r) || isNum(r)
}

const expectedPunctChars = "''', '}', '(', ')', '[', ']', '.', '!', '<', '>', '=', '&', '|', '*', ',', ' '"
const expectedDigitChars = "'0'..'9'"
const expectedAlphaChars = "'a'..'z', 'A'..'Z'"
const expectedAllChars = expectedAlphaChars + ", " + expectedDigitChars + ", " + expectedPunctChars + ", '_'"

// ExprLexer is a struct to lex expression syntax. To know the syntax, see
// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions
type ExprLexer struct {
	src    string
	scan   scanner.Scanner
	lexErr *ExprError
	start  scanner.Position
}

// NewExprLexer makes new ExprLexer instance.
func NewExprLexer(src string) *ExprLexer {
	l := &ExprLexer{
		src: src,
		start: scanner.Position{
			Offset: 0,
			Line:   1,
			Column: 1,
		},
	}
	l.scan.Init(strings.NewReader(src))
	l.scan.Error = func(_ *scanner.Scanner, m string) {
		l.error(fmt.Sprintf("scan error while lexing expression: %s", m))
	}
	return l
}

func (lex *ExprLexer) error(msg string) {
	if lex.lexErr == nil {
		p := lex.scan.Pos()
		lex.lexErr = &ExprError{
			Message: msg,
			Offset:  p.Offset,
			Line:    p.Line,
			Column:  p.Column,
		}
	}
}

func (lex *ExprLexer) token(kind TokenKind) *Token {
	p := lex.scan.Pos()
	s := lex.start
	t := &Token{
		Kind:   kind,
		Value:  lex.src[s.Offset:p.Offset],
		Offset: s.Offset,
		Line:   s.Line,
		Column: s.Column,
	}
	lex.start = p
	return t
}

func (lex *ExprLexer) eof() *Token {
	return &Token{
		Kind:   TokenKindEnd,
		Value:  "",
		Offset: lex.start.Offset,
		Line:   lex.start.Line,
		Column: lex.start.Column,
	}
}

func (lex *ExprLexer) skipWhite() {
	for {
		if r := lex.scan.Peek(); !isWhitespace(r) {
			return
		}
		lex.scan.Next()
		lex.start = lex.scan.Pos()
	}
}

func (lex *ExprLexer) unexpected(r rune, where string, expected string) *Token {
	var what string
	if r == scanner.EOF {
		what = "EOF"
	} else {
		what = "character " + strconv.QuoteRune(r)
	}

	var note string
	if r == '"' {
		note = ". do you mean string literals? only single quotes are available for string delimiter"
	}

	msg := fmt.Sprintf(
		"got unexpected %s while lexing %s, expecting %s%s",
		what,
		where,
		expected,
		note,
	)

	lex.error(msg)
	return lex.eof()
}

func (lex *ExprLexer) unexpectedEOF() *Token {
	lex.error("unexpected EOF while lexing expression")
	return lex.eof()
}

func (lex *ExprLexer) lexIdent() *Token {
	for {
		lex.scan.Next()
		// a-z, A-Z, 0-9, - or _
		// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
		if r := lex.scan.Peek(); !isAlnum(r) && r != '_' && r != '-' {
			return lex.token(TokenKindIdent)
		}
	}
}

func (lex *ExprLexer) lexNum() *Token {
	// The official document says number literals are 'Any number format supported by JSON' but actually
	// hex numbers starting with 0x are supported.

	r := lex.scan.Next() // precond: r is digit or '-'

	if r == '-' {
		r = lex.scan.Peek()
		if !isNum(r) {
			return lex.unexpected(r, "number after -", expectedDigitChars)
		}
		lex.scan.Next()
	}

	if r == '0' {
		r = lex.scan.Peek()
		if r == 'x' {
			lex.scan.Next()
			return lex.lexHexInt()
		}
		if isAlnum(r) && r != 'e' && r != 'E' {
			e := "'e', 'E', " + expectedPunctChars
			return lex.unexpected(r, "number after 0", e)
		}
	} else {
		// r is 1..9
		for {
			r = lex.scan.Peek()
			if !isNum(r) {
				break
			}
			lex.scan.Next()
		}
	}

	k := TokenKindInt

	if r == '.' {
		lex.scan.Next() // eat '.'
		r = lex.scan.Peek()
		if !isNum(r) {
			return lex.unexpected(r, "fraction part of float number", expectedDigitChars)
		}
		lex.scan.Next()

		for {
			r = lex.scan.Peek()
			if !isNum(r) {
				break
			}
			lex.scan.Next()
		}

		k = TokenKindFloat
	}

	if r == 'e' || r == 'E' {
		lex.scan.Next() // eat 'e' or 'E'
		r = lex.scan.Peek()
		if r == '-' {
			lex.scan.Next()
			r = lex.scan.Peek()
		}
		if !isNum(r) {
			return lex.unexpected(r, "exponent part of float number", expectedDigitChars)
		}
		lex.scan.Next()

		if r == '0' {
			r = lex.scan.Peek()
			if isNum(r) {
				return lex.unexpected(r, "number after 0 in exponent part", expectedPunctChars)
			}
		} else {
			for {
				r = lex.scan.Peek()
				if !isNum(r) {
					break
				}
				lex.scan.Next()
			}
		}

		k = TokenKindFloat
	}

	return lex.token(k)
}

func (lex *ExprLexer) lexHexInt() *Token {
	r := lex.scan.Peek()
	if !isHexNum(r) {
		e := expectedDigitChars + ", 'a'..'f', 'A'..'F'"
		return lex.unexpected(r, "hex integer", e)
	}
	lex.scan.Next()

	if r == '0' {
		r = lex.scan.Peek()
		if isHexNum(r) {
			return lex.unexpected(r, "number after 0x0", expectedPunctChars)
		}
	} else {
		for {
			r = lex.scan.Peek()
			if !isHexNum(r) {
				break
			}
			lex.scan.Next()
		}
	}

	// Note: GitHub Actions does not support exponent part like 0x1f2p-a8

	return lex.token(TokenKindInt)
}

func (lex *ExprLexer) lexString() *Token {
	lex.scan.Next() // eat '
	for {
		switch lex.scan.Peek() {
		case '\'':
			lex.scan.Next()
			if lex.scan.Peek() != '\'' { // when not escaped single quote ''
				return lex.token(TokenKindString)
			}
		case scanner.EOF:
			return lex.unexpected(scanner.EOF, "end of string literal", "'''")
		}
		lex.scan.Next()
	}
}

func (lex *ExprLexer) lexEnd() *Token {
	lex.scan.Next() // eat '}'
	if r := lex.scan.Peek(); r != '}' {
		return lex.unexpected(r, "end marker }}", "'}'")
	}
	lex.scan.Next()
	// }} is an end marker of interpolation
	return lex.token(TokenKindEnd)
}

func (lex *ExprLexer) lexLess() *Token {
	lex.scan.Next() // eat '<'
	k := TokenKindLess
	if lex.scan.Peek() == '=' {
		k = TokenKindLessEq
		lex.scan.Next()
	}
	return lex.token(k)
}

func (lex *ExprLexer) lexGreater() *Token {
	lex.scan.Next() // eat '>'
	k := TokenKindGreater
	if lex.scan.Peek() == '=' {
		k = TokenKindGreaterEq
		lex.scan.Next()
	}
	return lex.token(k)
}

func (lex *ExprLexer) lexEq() *Token {
	lex.scan.Next() // eat '='
	if r := lex.scan.Peek(); r != '=' {
		return lex.unexpected(r, "== operator", "'='")
	}
	lex.scan.Next()
	return lex.token(TokenKindEq)
}

func (lex *ExprLexer) lexBang() *Token {
	lex.scan.Next() // eat '!'
	k := TokenKindNot
	if lex.scan.Peek() == '=' {
		lex.scan.Next() // eat '='
		k = TokenKindNotEq
	}
	return lex.token(k)
}

func (lex *ExprLexer) lexAnd() *Token {
	lex.scan.Next() // eat '&'
	if r := lex.scan.Peek(); r != '&' {
		return lex.unexpected(r, "&& operator", "'&'")
	}
	lex.scan.Next()
	return lex.token(TokenKindAnd)
}

func (lex *ExprLexer) lexOr() *Token {
	lex.scan.Next() // eat '|'
	if r := lex.scan.Peek(); r != '|' {
		return lex.unexpected(r, "|| operator", "'|'")
	}
	lex.scan.Next()
	return lex.token(TokenKindOr)
}

func (lex *ExprLexer) lexChar(k TokenKind) *Token {
	lex.scan.Next()
	return lex.token(k)
}

// Next lexes next token to lex input incrementally. Lexer must be initialized with Init() method
// before the first call of this method. This method is stateful. Lexer advances offset by lexing
// token. To get the offset, use Offset() method.
func (lex *ExprLexer) Next() *Token {
	lex.skipWhite()

	r := lex.scan.Peek()
	if r == scanner.EOF {
		return lex.unexpectedEOF()
	}

	// Ident starts with a-z or A-Z or _
	// https://docs.github.com/en/actions/reference/context-and-expression-syntax-for-github-actions#contexts
	if isAlpha(r) || r == '_' {
		return lex.lexIdent()
	}

	if isNum(r) || r == '-' {
		return lex.lexNum()
	}

	switch r {
	case '\'':
		return lex.lexString()
	case '}':
		return lex.lexEnd()
	case '!':
		return lex.lexBang()
	case '<':
		return lex.lexLess()
	case '>':
		return lex.lexGreater()
	case '=':
		return lex.lexEq()
	case '&':
		return lex.lexAnd()
	case '|':
		return lex.lexOr()
	case '(':
		return lex.lexChar(TokenKindLeftParen)
	case ')':
		return lex.lexChar(TokenKindRightParen)
	case '[':
		return lex.lexChar(TokenKindLeftBracket)
	case ']':
		return lex.lexChar(TokenKindRightBracket)
	case '.':
		return lex.lexChar(TokenKindDot)
	case '*':
		return lex.lexChar(TokenKindStar)
	case ',':
		return lex.lexChar(TokenKindComma)
	default:
		return lex.unexpected(r, "expression", expectedAllChars)
	}
}

// Offset returns the current offset (scanning position).
func (lex *ExprLexer) Offset() int {
	return lex.scan.Pos().Offset
}

// Err returns an error while lexing. When multiple errors occur, the first one is returned.
func (lex *ExprLexer) Err() *ExprError {
	return lex.lexErr
}

// LexExpression lexes the given string as expression syntax. The parameter must contain '}}' which
// represents end of expression. Otherwise this function will report an error that it encountered
// unexpected EOF.
func LexExpression(src string) ([]*Token, int, *ExprError) {
	l := NewExprLexer(src)
	ts := []*Token{}
	for {
		t := l.Next()
		if l.lexErr != nil {
			return nil, l.scan.Pos().Offset, l.lexErr
		}
		ts = append(ts, t)
		if t.Kind == TokenKindEnd {
			return ts, l.scan.Pos().Offset, nil
		}
	}
}
