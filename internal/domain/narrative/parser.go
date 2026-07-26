package narrative

import "fmt"

type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.next()
	p.next()
	return p
}

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) Parse() (*Grammar, error) {
	rules := make(map[string]Rule)

	for p.cur.Type != TokenEOF {
		if p.cur.Type == TokenNewline {
			p.next()
			continue
		}
		rule, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		if _, exists := rules[rule.Name]; exists {
			return nil, fmt.Errorf("rule %q already defined at %s", rule.Name, p.cur.Pos)
		}
		rules[rule.Name] = *rule
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("empty grammar: no rules found")
	}

	return &Grammar{Rules: rules}, nil
}

func (p *Parser) parseRule() (*Rule, error) {
	if p.cur.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected rule name at %s, got %s (%q)", p.cur.Pos, p.cur.Type, p.cur.Value)
	}
	name := p.cur.Value
	p.next()

	if p.cur.Type != TokenDefinitionSep {
		return nil, fmt.Errorf("expected '::=' at %s, got %s (%q)", p.cur.Pos, p.cur.Type, p.cur.Value)
	}
	p.next()

	var alts []Alternative
	for {
		for p.cur.Type == TokenNewline {
			p.next()
		}

		if p.cur.Type == TokenEOF || p.cur.Type == TokenIdentifier {
			break
		}

		alt, err := p.parseAlternative()
		if err != nil {
			return nil, err
		}
		alts = append(alts, alt)

		for p.cur.Type == TokenNewline {
			p.next()
		}

		if p.cur.Type == TokenAlternative {
			p.next()
			continue
		}
		break
	}

	if len(alts) == 0 {
		return nil, fmt.Errorf("rule %q has no alternatives at %s", name, p.cur.Pos)
	}

	return &Rule{Name: name, Alternatives: alts}, nil
}

func (p *Parser) parseAlternative() (Alternative, error) {
	var alt Alternative
	for {
		switch p.cur.Type {
		case TokenTerminal:
			alt = append(alt, NewTerminal(p.cur.Value))
			p.next()
		case TokenNonTerminal:
			alt = append(alt, NewNonTerminal(p.cur.Value))
			p.next()
		case TokenVariable:
			alt = append(alt, NewVariable(p.cur.Value))
			p.next()
		default:
			if len(alt) == 0 {
				return nil, fmt.Errorf("expected symbol at %s, got %s (%q)", p.cur.Pos, p.cur.Type, p.cur.Value)
			}
			return alt, nil
		}
	}
}

func Parse(input string) (*Grammar, error) {
	return NewParser(NewLexer(input)).Parse()
}
