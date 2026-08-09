package narrative

// DefaultGrammarProvider is the GrammarProvider implementation backed by the
// DefaultGrammar source. It satisfies the usecase-level GrammarProvider
// interface without importing the usecase package, keeping dependency
// direction inward.
type DefaultGrammarProvider struct{}

// Grammar returns the default grammar source.
func (DefaultGrammarProvider) Grammar() string { return DefaultGrammar }
