package narrative

import (
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	l := NewLexer(`a ::= "x" | <b> $c`)

	want := []Token{
		{Type: TokenIdentifier, Value: "a"},
		{Type: TokenDefinitionSep, Value: "::="},
		{Type: TokenTerminal, Value: "x"},
		{Type: TokenAlternative, Value: "|"},
		{Type: TokenNonTerminal, Value: "b"},
		{Type: TokenVariable, Value: "c"},
		{Type: TokenEOF, Value: ""},
	}

	for i, tt := range want {
		tok := l.NextToken()
		if tok.Type != tt.Type {
			t.Fatalf("token %d: expected type %v, got %v", i, tt.Type, tok.Type)
		}
		if tok.Value != tt.Value {
			t.Fatalf("token %d: expected value %q, got %q", i, tt.Value, tok.Value)
		}
	}
}

func TestLexer_NewlinesAndComments(t *testing.T) {
	l := NewLexer(`# leading
rule ::= "one"
# trailing`)

	want := []TokenType{
		TokenNewline,
		TokenIdentifier,
		TokenDefinitionSep,
		TokenTerminal,
		TokenNewline,
		TokenEOF,
	}

	for i, tt := range want {
		tok := l.NextToken()
		if tok.Type != tt {
			t.Fatalf("token %d: expected type %v, got %v (%q)", i, tt, tok.Type, tok.Value)
		}
	}
}

func TestLexer_EscapeSequences(t *testing.T) {
	l := NewLexer(`a ::= "\n\t\\\""`)

	tok := l.NextToken()
	if tok.Type != TokenIdentifier || tok.Value != "a" {
		t.Fatalf("expected identifier a, got %v %q", tok.Type, tok.Value)
	}

	tok = l.NextToken()
	if tok.Type != TokenDefinitionSep {
		t.Fatalf("expected def sep, got %v", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != TokenTerminal {
		t.Fatalf("expected terminal, got %v", tok.Type)
	}
	want := "\n\t\\\""
	if tok.Value != want {
		t.Fatalf("expected %q, got %q", want, tok.Value)
	}
}

func TestLexer_TokenTypeString(t *testing.T) {
	cases := []struct {
		tok  TokenType
		want string
	}{
		{TokenIllegal, "ILLEGAL"},
		{TokenEOF, "EOF"},
		{TokenType(-1), "TOKEN(-1)"},
	}

	for _, tc := range cases {
		if got := tc.tok.String(); got != tc.want {
			t.Fatalf("TokenType(%d).String() = %q, want %q", tc.tok, got, tc.want)
		}
	}
}

func TestLexer_Position_String(t *testing.T) {
	p := Position{Line: 3, Col: 5}
	if got := p.String(); got != "3:5" {
		t.Fatalf("expected %q, got %q", "3:5", got)
	}
}

func TestLexer_Errors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  TokenType
	}{
		{
			name:  "illegal character",
			input: `a ::= @`,
			want:  TokenIllegal,
		},
		{
			name:  "unclosed non-terminal",
			input: `a ::= <foo`,
			want:  TokenIllegal,
		},
		{
			name:  "unclosed terminal",
			input: `a ::= "foo`,
			want:  TokenIllegal,
		},
		{
			name:  "terminal newline before close",
			input: "a ::= \"foo\nbar\"",
			want:  TokenIllegal,
		},
		{
			name:  "lone colon",
			input: `a : "x"`,
			want:  TokenIllegal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLexer(tc.input)
			for {
				tok := l.NextToken()
				if tok.Type == TokenEOF {
					t.Fatalf("expected %v token before EOF", tc.want)
				}
				if tok.Type == tc.want {
					return
				}
			}
		})
	}
}
