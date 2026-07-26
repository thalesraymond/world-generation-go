# simulation-loop Specification

## Purpose

Define the discrete-year simulation engine that advances world entities, processes agent decisions, and emits timeline events.

## MODIFIED Requirements

### Requirement: Agent Decision Loop in Entity Tick

The `settlementEntity.Tick()` method SHALL execute an agent decision loop in place of random event generation.

#### Scenario: Agent decision replaces random events

- **WHEN** a settlement entity ticks for a year
- **THEN** steps 1–4 (figure lifecycle: age, deaths, births, role events) SHALL remain unchanged
- **THEN** steps 5–6 (core settlement random events) SHALL be replaced by agent decision loop
- **THEN** the agent decision loop SHALL: (1) evaluate all six actions, (2) filter by preconditions, (3) score by goal alignment, (4) select via weighted random using agent RNG, (5) execute action, (6) emit event

#### Scenario: Agent RNG usage

- **WHEN** the agent decision loop executes
- **THEN** action selection SHALL use the settlement's agent RNG (`agentRNG` field)
- **THEN** relation shifts SHALL use the agent RNG for random magnitudes (e.g., Raid −0.3 to −0.5)
- **THEN** Expand target selection SHALL use the agent RNG for weighted random

#### Scenario: Sequential execution within year

- **WHEN** the simulation loop ticks for a year
- **THEN** all settlements SHALL execute their Tick() sequentially in settlement slice order
- **THEN** each settlement SHALL execute exactly one action per year
- **THEN** Expand actions that create new settlements SHALL add them to the slice (affecting subsequent years, not current iteration)

### Requirement: Event Emission from Agent Actions

The system SHALL emit events with new categories from agent actions.

#### Scenario: Agent event categories

- **WHEN** an agent action is executed
- **THEN** Expand SHALL emit event with Category "Expansion"
- **THEN** Raid SHALL emit event with Category "Raid"
- **THEN** Conquer SHALL emit event with Category "Conquest"
- **THEN** Fortify SHALL emit event with Category "Economy"
- **THEN** Ally SHALL emit event with Category "Diplomacy"
- **THEN** Prosper SHALL emit event with Category "Economy"

#### Scenario: Target settlement in events

- **WHEN** an agent action has a target (Raid, Conquer, Ally)
- **THEN** the event SHALL include `TargetSettlement string` field with the target's name
- **THEN** events without targets (Expand, Fortify, Prosper) SHALL omit the TargetSettlement field
