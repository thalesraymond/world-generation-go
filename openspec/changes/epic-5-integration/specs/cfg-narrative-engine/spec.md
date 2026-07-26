## MODIFIED Requirements

### Requirement: Event Translation
The system SHALL translate numerical events into text descriptions by resolving rules from the loaded grammar. When events include figure context variables, the engine SHALL substitute figure names, roles, and settlement names into the resolved text. The engine SHALL additionally support agent-decision event types (settlement actions), character-execution event types (figure-led actions with stats and reputation), artifact-transfer event types (creation, theft, loss), and faction-strategy event types (war declaration, alliance formation, policy shift).

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

#### Scenario: Agent-decision event translation
- **WHEN** a settlement agent decision event (e.g., raid, conquer, ally) is passed to the narrative engine
- **THEN** the engine SHALL select the corresponding agent-decision root rule
- **THEN** the resolved text SHALL include the settlement name, action type, and any figure executor context

#### Scenario: Artifact-transfer event translation
- **WHEN** an artifact transfer event (creation, theft, loss) is passed to the narrative engine
- **THEN** the engine SHALL resolve the artifact-transfer rule
- **THEN** the resolved text SHALL include the artifact name, type, and participants

#### Scenario: Faction-strategy event translation
- **WHEN** a faction strategy event (war declaration, alliance) is passed to the narrative engine
- **THEN** the engine SHALL resolve the faction-strategy rule
- **THEN** the resolved text SHALL include both faction names and the strategic action

### Requirement: Grammar Support for Figure Variables
The default grammar SHALL include nonterminal symbols for figure names, roles, and settlement names that can be referenced in narrative rules. The grammar SHALL additionally include nonterminal symbols for artifact names (`<ArtifactName>`, `<ArtifactType>`) and faction names (`<FactionName>`).

#### Scenario: Figure variable nonterminals
- **WHEN** the default grammar is loaded
- **THEN** rules MAY reference `<FigureName>`, `<FigureRole>`, and `<SettlementName>` nonterminal symbols
- **THEN** when these variables are not provided (non-figure events), they SHALL resolve to a generic fallback description (e.g., "a figure", "the settlement")

#### Scenario: Artifact and faction variable nonterminals
- **WHEN** the default grammar is loaded
- **THEN** rules MAY reference `<ArtifactName>`, `<ArtifactType>`, and `<FactionName>` nonterminal symbols
- **THEN** when these variables are not provided, they SHALL resolve to a generic fallback description

## ADDED Requirements

### Requirement: CFG Grammar Coverage for New Event Categories
The CFG grammar SHALL include production rules for all new agent event categories: settlement-decision, character-execution, artifact-transfer, and faction-strategy.

#### Scenario: Grammar rule existence
- **WHEN** the default grammar is loaded
- **THEN** production rules SHALL exist for each new event category
- **THEN** each rule SHALL reference appropriate nonterminal symbols for settlement names, figure names, artifact names, and faction names
