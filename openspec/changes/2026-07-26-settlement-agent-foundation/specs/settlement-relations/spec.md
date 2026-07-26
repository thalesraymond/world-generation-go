# settlement-relations Specification

## Purpose

Define inter-settlement relations tracking, initialization, shift calculations, and usage in action preconditions.

## ADDED Requirements

### Requirement: Relations Map Storage

The system SHALL store relations as a map from settlement name to float value for each settlement.

#### Scenario: Relations map structure

- **WHEN** a settlement is created
- **THEN** it SHALL have a `Relations map[string]float64` field
- **THEN** the map SHALL be keyed by settlement name (string)
- **THEN** values SHALL be float64 in range −1.0 to +1.0

#### Scenario: Relations serialization

- **WHEN** a world state is serialized to JSON
- **THEN** the Relations map SHALL be included as a JSON object
- **THEN** round-trip deserialization SHALL produce equivalent relations

### Requirement: Relations Initialization

The system SHALL initialize relations with baseline values based on faction alignment.

#### Scenario: Same-faction baseline

- **WHEN** a settlement is created and another settlement shares the same faction (not "independent")
- **THEN** the initial relations value SHALL be +0.3

#### Scenario: Different-faction baseline

- **WHEN** a settlement is created and another settlement has a different faction
- **THEN** the initial relations value SHALL be 0.0

#### Scenario: Independent faction baseline

- **WHEN** a settlement is created with faction "independent"
- **THEN** its relations to all other settlements SHALL be 0.0 (no faction bonus)

### Requirement: Relation Shifts Per Action

The system SHALL shift relations by specific values based on action type.

#### Scenario: Raid relation shift

- **WHEN** a settlement executes a Raid action against a target
- **THEN** on success: raider relations to target SHALL shift −0.4, target relations to raider SHALL shift −0.3
- **THEN** on failure: raider relations to target SHALL shift −0.2

#### Scenario: Conquer relation shift

- **WHEN** a settlement executes a Conquer action against a target
- **THEN** both settlements' relations to each other SHALL shift −0.8

#### Scenario: Ally relation shift

- **WHEN** a settlement executes an Ally action with a target
- **THEN** both settlements' relations to each other SHALL shift +0.4

#### Scenario: Prosper relation shift

- **WHEN** a settlement executes a Prosper action
- **THEN** its relations to all other settlements SHALL shift +0.05

### Requirement: Relation Clamping

The system SHALL clamp relation values to the range −1.0 to +1.0 after each shift.

#### Scenario: Clamp at lower bound

- **WHEN** a relation shift would result in a value < −1.0
- **THEN** the value SHALL be clamped to −1.0

#### Scenario: Clamp at upper bound

- **WHEN** a relation shift would result in a value > +1.0
- **THEN** the value SHALL be clamped to +1.0

### Requirement: Relations in Action Preconditions

The system SHALL use relations values in action precondition checks.

#### Scenario: Raid requires hostile relations

- **WHEN** a settlement evaluates Raid action precondition
- **THEN** the action SHALL fail if relations to target ≥ −0.5

#### Scenario: Ally requires friendly relations

- **WHEN** a settlement evaluates Ally action precondition
- **THEN** the action SHALL fail if relations to target ≤ 0.5

#### Scenario: Conquer requires very hostile relations

- **WHEN** a settlement evaluates Conquer action precondition
- **THEN** the action SHALL fail if relations to target ≥ −0.7
