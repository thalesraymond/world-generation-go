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
- **THEN** the resolved narrative text SHALL include the figure's name in place of `$FigureName` variable references
- **THEN** the resolved narrative text SHALL include the figure's role in place of `$FigureRole` variable references
- **THEN** the resolved narrative text SHALL include the settlement name in place of `$SettlementName` variable references

### Requirement: Recursion Limit
The system SHALL prevent infinite recursion during grammar resolution.

#### Scenario: Deep recursion detected
- **WHEN** grammar resolution exceeds the predefined maximum depth
- **THEN** resolution halts and a fallback description or error message is returned

### Requirement: Grammar Support for Figure Variables
The default grammar SHALL include variable symbols for figure names, roles, and settlement names that can be referenced in narrative rules.

#### Scenario: Figure variable symbols
- **WHEN** the default grammar is loaded
- **THEN** rules MAY reference `$FigureName`, `$FigureRole`, `$SettlementName`, `$TargetSettlement`, `$year`, `$category`, and `$description` variable symbols
- **THEN** when these variables are not provided, the missing variable token SHALL be emitted literally in the output

### Requirement: Figure-Aware Rule Overrides
The default grammar SHALL include dotted rule names that provide figure-aware alternatives for generic categories.

#### Scenario: Figure-aware rules
- **WHEN** the default grammar is loaded
- **THEN** rules named `Conflict.figure`, `Politics.figure`, and `Discovery.figure` SHALL exist
- **THEN** these rules SHALL reference `$FigureName`, `$FigureRole`, `$SettlementName`, and `$TargetSettlement` variables

### Requirement: Event Category Grammar Rules
The default grammar SHALL include rules for all event categories emitted by the simulation, including Marriage, RoleTransition, Succession, and ReputationChange.

#### Scenario: Lifecycle event rules
- **WHEN** the default grammar is loaded
- **THEN** rules named `Marriage`, `RoleTransition`, `Succession`, and `ReputationChange` SHALL exist
- **THEN** these rules SHALL reference `$FigureName`, `$FigureRole`, `$SettlementName`, and `$year` variables

### Requirement: Explicit Rule Resolution
The engine SHALL support resolving a specific rule by name while retaining the standard event context variables.

#### Scenario: NarrateWithRule
- **WHEN** `NarrateWithRule` is called with an event, context, RNG, and explicit rule name
- **THEN** the engine SHALL resolve the named rule using the same context as `Narrate`
- **THEN** if the rule is missing, the engine SHALL fall back to the event description
