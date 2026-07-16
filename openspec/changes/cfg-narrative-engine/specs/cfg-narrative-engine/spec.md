## ADDED Requirements

### Requirement: CFG Parser Initialization
The system SHALL be able to load and parse Context-Free Grammar files defining narrative rules.

#### Scenario: Successful grammar loading
- **WHEN** the narrative engine is initialized with a valid grammar file path
- **THEN** it successfully parses the file and populates its internal rule map without errors

#### Scenario: Invalid grammar format
- **WHEN** the narrative engine is initialized with a malformed grammar file
- **THEN** it returns an error detailing the parse failure

### Requirement: Event Translation
The system SHALL translate numerical events into text descriptions by resolving rules from the loaded grammar.

#### Scenario: Translating a simple event
- **WHEN** a numerical event with a predefined event type is passed to the engine
- **THEN** it selects the corresponding root rule and resolves it to a narrative string

#### Scenario: Variable injection during resolution
- **WHEN** a numerical event containing contextual data is translated
- **THEN** the resolved text correctly substitutes the contextual data into the generated narrative

### Requirement: Recursion Limit
The system SHALL prevent infinite recursion during grammar resolution.

#### Scenario: Deep recursion detected
- **WHEN** grammar resolution exceeds the predefined maximum depth
- **THEN** resolution halts and a fallback description or error message is returned
