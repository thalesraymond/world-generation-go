## 1. Parser Implementation

- [ ] 1.1 Create AST structs for the Context-Free Grammar nodes (e.g., Rules, Alternatives, Terminals, Non-Terminals)
- [ ] 1.2 Implement a lexer to tokenize the BNF-like grammar files
- [ ] 1.3 Implement the parser to construct the grammar AST from tokens
- [ ] 1.4 Write unit tests for parsing valid and invalid grammar definitions

## 2. Engine Core

- [ ] 2.1 Implement the Narrative Engine struct to hold loaded grammar rules
- [ ] 2.2 Create a rule resolution mechanism that handles variable injection from an event context
- [ ] 2.3 Add recursion depth tracking and limits to the resolution process
- [ ] 2.4 Integrate the engine with numerical event streams to produce narrative strings

## 3. Example Grammars & Tests

- [ ] 3.1 Create sample BNF grammar files modeling mythical events
- [ ] 3.2 Write end-to-end tests feeding sample numerical events into the narrative engine and validating textual output
