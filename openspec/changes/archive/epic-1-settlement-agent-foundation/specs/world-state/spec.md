# world-state Specification

## Purpose

Define the world state data model including settlements, terrain layers, and agent state fields.

## MODIFIED Requirements

### Requirement: Settlement Struct Extension

The `Settlement` struct in `internal/domain/world/state.go` SHALL be extended with agent state fields.

#### Scenario: Agent fields added

- **WHEN** the `Settlement` struct is defined
- **THEN** it SHALL include `MilitaryStrength float64` with JSON tag `json:"militaryStrength"`
- **THEN** it SHALL include `Wealth float64` with JSON tag `json:"wealth"`
- **THEN** it SHALL include `Relations map[string]float64` with JSON tag `json:"relations"`
- **THEN** it SHALL include `Goals []string` with JSON tag `json:"goals"`
- **THEN** existing fields (Name, Type, X, Y, Faction, Population, Figures) SHALL remain unchanged

#### Scenario: Backward compatibility

- **WHEN** a `world_state.json` file without agent fields is deserialized
- **THEN** the system SHALL NOT require the new fields to be present
- **THEN** settlements SHALL have zero values for agent fields (0.0 for MilitaryStrength/Wealth, nil for Relations/Goals)
- **THEN** deserialization SHALL NOT produce an error

### Requirement: Relations Map Initialization

The system SHALL provide a function to initialize relations maps for new settlements.

#### Scenario: initRelations function

- **WHEN** `initRelations(self Settlement, allSettlements []Settlement)` is called
- **THEN** it SHALL return a `map[string]float64` with an entry for each settlement in `allSettlements` (excluding self)
- **THEN** entries for same-faction settlements (faction != "independent") SHALL have value +0.3
- **THEN** entries for different-faction settlements SHALL have value 0.0
