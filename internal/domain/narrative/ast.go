package narrative

type Grammar struct {
	Rules map[string]Rule
}

type Rule struct {
	Name         string
	Alternatives []Alternative
}

type Alternative []Symbol

type Symbol interface {
	isSymbol()
}

type Terminal struct {
	Text string
}

func (Terminal) isSymbol() {}

type NonTerminal struct {
	Name string
}

func (NonTerminal) isSymbol() {}

type Variable struct {
	Name string
}

func (Variable) isSymbol() {}

func NewTerminal(text string) Terminal {
	return Terminal{Text: text}
}

func NewNonTerminal(name string) NonTerminal {
	return NonTerminal{Name: name}
}

func NewVariable(name string) Variable {
	return Variable{Name: name}
}