## MODIFIED Requirements

### Requirement: Figure-Aware Grammar Rules

The default grammar SHALL include figure-aware production rules for Conflict, Politics, and Discovery events that consume injected variables.

#### Scenario: Figure-aware Conflict rule

- **WHEN** a Conflict event is translated with `FigureName`, `FigureRole`, `SettlementName`, and `TargetSettlement` variables
- **THEN** the engine SHALL prefer the `Conflict.figure` rule
- **THEN** the resolved text SHALL include the figure's name, role, and settlement, and the target settlement

#### Scenario: Figure-aware Politics rule

- **WHEN** a Politics event is translated with figure variables
- **THEN** the engine SHALL prefer the `Politics.figure` rule
- **THEN** the resolved text SHALL describe a political act brokered by the named figure

#### Scenario: Figure-aware Discovery rule

- **WHEN** a Discovery event is translated with figure variables
- **THEN** the engine SHALL prefer the `Discovery.figure` rule
- **THEN** the resolved text SHALL describe the named figure discovering or charting new lands

### Requirement: Grammar Rules for New Event Categories

The default grammar SHALL include rules for the Marriage, RoleTransition, Succession, and ReputationChange event categories.

#### Scenario: Marriage event narration

- **WHEN** a Marriage event is translated
- **THEN** the engine SHALL resolve the `Marriage` rule
- **THEN** the resolved text SHALL include the figure's name, settlement, and year

#### Scenario: Role transition event narration

- **WHEN** a RoleTransition event is translated
- **THEN** the engine SHALL resolve the `RoleTransition` rule
- **THEN** the resolved text SHALL describe the figure's old and new role

#### Scenario: Succession event narration

- **WHEN** a Succession event is translated
- **THEN** the engine SHALL resolve the `Succession` rule
- **THEN** the resolved text SHALL describe the new leader rising to power

#### Scenario: Reputation change event narration

- **WHEN** a ReputationChange event is translated
- **THEN** the engine SHALL resolve the `ReputationChange` rule
- **THEN** the resolved text SHALL describe the figure's growing reputation

### Requirement: Rule-Name Dot Notation

The grammar lexer and parser SHALL accept rule names containing a dot separator.

#### Scenario: Dotted rule names

- **WHEN** a grammar contains a rule named `Conflict.figure`
- **THEN** the parser SHALL accept it as a valid rule name
- **THEN** the engine SHALL resolve it by the full dotted name

### Requirement: Rule Selection Fallback

The engine SHALL support explicit rule selection while falling back to the event category when the explicit rule is absent.

#### Scenario: Explicit rule fallback

- **WHEN** `NarrateWithRule` is called with a rule name that does not exist in the grammar
- **THEN** the engine SHALL return the event description as a fallback

#### Scenario: Event category fallback

- **WHEN** a figure-aware rule is not available or not applicable
- **THEN** the engine SHALL resolve the base category rule (e.g., `Conflict`) instead
