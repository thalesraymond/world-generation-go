# timeline-streaming Specification

## Purpose

Define the timeline event structure, categories, and streaming behavior during simulation.

## MODIFIED Requirements

### Requirement: Event Struct Extension

The `simulation.Event` struct SHALL be extended with an optional target settlement field.

#### Scenario: TargetSettlement field added

- **WHEN** the `Event` struct is defined
- **THEN** it SHALL include `TargetSettlement string` with JSON tag `json:"targetSettlement,omitempty"`
- **THEN** existing fields (Year, Category, Description, FigureID, RelatedFigures, SettlementName) SHALL remain unchanged

#### Scenario: Target settlement in agent events

- **WHEN** an agent action with a target is executed (Raid, Conquer, Ally)
- **THEN** the event SHALL include the target settlement's name in the TargetSettlement field
- **THEN** events without targets (Expand, Fortify, Prosper) SHALL omit the TargetSettlement field (empty string)

### Requirement: New Event Categories

The system SHALL support new event categories for agent actions.

#### Scenario: Agent event categories

- **WHEN** agent actions emit events
- **THEN** the following categories SHALL be supported: "Expansion", "Raid", "Conquest", "Diplomacy", "Economy"
- **THEN** existing categories SHALL remain supported: "Conflict", "Disaster", "Politics", "Discovery", "Settlement", "Birth", "Death"

#### Scenario: Event formatting with target

- **WHEN** an event with TargetSettlement is formatted via `FormatEvent()`
- **THEN** the output SHALL include the target: `"[Year] (Category) SettlementName → TargetSettlement: Description"`
- **THEN** events without targets SHALL format as before: `"[Year] (Category) SettlementName: Description"`
