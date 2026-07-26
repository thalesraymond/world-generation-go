# cfg-narrative-engine Specification

## Purpose

Define the Context-Free Grammar (CFG) narrative engine for translating numerical simulation events into rich, mythical text descriptions with variable injection, recursion protection, and figure-aware variables.

## Requirements

### Requirement: CFG Parser Initialization
The system SHALL be able to load and parse Context-Free Grammar files defining narrative rules.

#### Scenario: Successful grammar loading
- **WHEN** the narrative engine is initialized with a valid grammar file path
- **THEN** it successfully parses the file and populates its internal rule map without errors

#### Scenario: Invalid grammar format
- **WHEN** the narrative engine is initialized with a malformed grammar file
- **THEN** it returns an error detailing the parse failure

### Requirement: Event Translation
The system SHALL translate numerical events into text descriptions by resolving rules from the loaded grammar. When events include figure context variables, the engine SHALL substitute figure names, roles, and settlement names into the resolved text.

#### Scenario: Translating a simple event
- **WHEN** a numerical event with a predefined event type is passed to the engine
- **THEN** it selects the corresponding root rule and resolves it to a narrative string

#### Scenario: Variable injection during resolution
- **WHEN** a numerical event containing contextual data is translated
- **THEN** the resolved text correctly substitutes the contextual data into the generated narrative

#### Scenario: Figure variable injection
- **WHEN** a figure-generated event is passed to the narrative engine with figure context variables (`FigureName`, `FigureRole`, `SettlementName`)
- **THEN** the resolved narrative text SHALL include the figure's name in place of `<FigureName>` nonterminal references
- **THEN** the resolved narrative text SHALL include the figure's role in place of `<FigureRole>` nonterminal references
- **THEN** the resolved narrative text SHALL include the settlement name in place of `<SettlementName>` nonterminal references

### Requirement: Recursion Limit
The system SHALL prevent infinite recursion during grammar resolution.

#### Scenario: Deep recursion detected
- **WHEN** grammar resolution exceeds the predefined maximum depth
- **THEN** resolution halts and a fallback description or error message is returned

### Requirement: Grammar Support for Figure Variables
The default grammar SHALL include nonterminal symbols for figure names, roles, and settlement names that can be referenced in narrative rules.

#### Scenario: Figure variable nonterminals
- **WHEN** the default grammar is loaded
- **THEN** rules MAY reference `<FigureName>`, `<FigureRole>`, and `<SettlementName>` nonterminal symbols
- **THEN** when these variables are not provided (non-figure events), they SHALL resolve to a generic fallback description (e.g., "a figure", "the settlement")
