package narrative

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenIllegal TokenType = iota
	TokenEOF
	TokenNewline
	TokenDefinitionSep
	TokenAlternative
	TokenNonTerminal
	TokenTerminal
	TokenVariable
	TokenIdentifier
)

func (t TokenType) String() string {
	switch t {
	case TokenIllegal:
		return "ILLEGAL"
	case TokenEOF:
		return "EOF"
	case TokenNewline:
		return "NEWLINE"
	case TokenDefinitionSep:
		return "DEF_SEP"
	case TokenAlternative:
		return "ALTERNATIVE"
	case TokenNonTerminal:
		return "NON_TERMINAL"
	case TokenTerminal:
		return "TERMINAL"
	case TokenVariable:
		return "VARIABLE"
	case TokenIdentifier:
		return "IDENTIFIER"
	default:
		return fmt.Sprintf("TOKEN(%d)", int(t))
	}
}

type Position struct {
	Line int
	Col  int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

type Token struct {
	Type  TokenType
	Value string
	Pos   Position
}

type Lexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return l.makeToken(TokenEOF, "")
	}

	ch := l.input[l.pos]

	switch {
	case ch == '\n':
		return l.readNewline()
	case ch == '#':
		l.skipComment()
		if l.pos >= len(l.input) {
			return l.makeToken(TokenEOF, "")
		}
		return l.NextToken()
	case ch == ':':
		return l.readDefinitionSep()
	case ch == '|':
		return l.readSingle(TokenAlternative, "|")
	case ch == '<':
		return l.readNonTerminal()
	case ch == '"':
		return l.readTerminal()
	case ch == '$':
		return l.readVariable()
	case unicode.IsLetter(ch) || ch == '_':
		return l.readIdentifier()
	default:
		return l.readIllegal()
	}
}

func (l *Lexer) makeToken(t TokenType, value string) Token {
	pos := Position{Line: l.line, Col: l.col}
	tok := Token{Type: t, Value: value, Pos: pos}
	if t != TokenEOF {
		l.advance()
	}
	return tok
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		ch := l.input[l.pos]
		l.pos++
		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\n' {
			return
		}
		l.advance()
	}
}

func (l *Lexer) readNewline() Token {
	pos := Position{Line: l.line, Col: l.col}
	l.advance()
	return Token{
		Type:  TokenNewline,
		Value: "\n",
		Pos:   pos,
	}
}

func (l *Lexer) readDefinitionSep() Token {
	startCol := l.col
	if l.pos+2 < len(l.input) && l.input[l.pos] == ':' && l.input[l.pos+1] == ':' && l.input[l.pos+2] == '=' {
		l.advance()
		l.advance()
		l.advance()
		return Token{
			Type:  TokenDefinitionSep,
			Value: "::=",
			Pos:   Position{Line: l.line, Col: startCol},
		}
	}
	l.advance()
	return Token{
		Type:  TokenIllegal,
		Value: string(l.input[l.pos-1]),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}

func (l *Lexer) readSingle(t TokenType, value string) Token {
	tok := l.makeToken(t, value)
	return tok
}

func (l *Lexer) readNonTerminal() Token {
	startCol := l.col
	l.advance()
	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '>' {
			l.advance()
			return Token{
				Type:  TokenNonTerminal,
				Value: buf.String(),
				Pos:   Position{Line: l.line, Col: startCol},
			}
		}
		buf.WriteRune(ch)
		l.advance()
	}
	return Token{
		Type:  TokenIllegal,
		Value: "<" + buf.String(),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}

func (l *Lexer) readTerminal() Token {
	startCol := l.col
	l.advance()
	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			l.advance()
			return Token{
				Type:  TokenTerminal,
				Value: buf.String(),
				Pos:   Position{Line: l.line, Col: startCol},
			}
		}
		if ch == '\\' {
			l.advance()
			if l.pos >= len(l.input) {
				return Token{
					Type:  TokenIllegal,
					Value: buf.String(),
					Pos:   Position{Line: l.line, Col: startCol},
				}
			}
			next := l.input[l.pos]
			switch next {
			case 'n':
				buf.WriteRune('\n')
			case 't':
				buf.WriteRune('\t')
			case '\\':
				buf.WriteRune('\\')
			case '"':
				buf.WriteRune('"')
			default:
				buf.WriteRune('\\')
				buf.WriteRune(next)
			}
			l.advance()
			continue
		}
		if ch == '\n' {
			return Token{
				Type:  TokenIllegal,
				Value: buf.String(),
				Pos:   Position{Line: l.line, Col: startCol},
			}
		}
		buf.WriteRune(ch)
		l.advance()
	}
	return Token{
		Type:  TokenIllegal,
		Value: buf.String(),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}

func (l *Lexer) readVariable() Token {
	startCol := l.col
	l.advance()
	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if !unicode.IsLetter(ch) && ch != '_' && !unicode.IsDigit(ch) {
			break
		}
		buf.WriteRune(ch)
		l.advance()
	}
	return Token{
		Type:  TokenVariable,
		Value: buf.String(),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}

func (l *Lexer) readIdentifier() Token {
	startCol := l.col
	var buf strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || ch == '_' || ch == '-' || unicode.IsDigit(ch) {
			buf.WriteRune(ch)
			l.advance()
		} else {
			break
		}
	}
	return Token{
		Type:  TokenIdentifier,
		Value: buf.String(),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}

func (l *Lexer) readIllegal() Token {
	startCol := l.col
	ch := l.input[l.pos]
	l.advance()
	return Token{
		Type:  TokenIllegal,
		Value: string(ch),
		Pos:   Position{Line: l.line, Col: startCol},
	}
}