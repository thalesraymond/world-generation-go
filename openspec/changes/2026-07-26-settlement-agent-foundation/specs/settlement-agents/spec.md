# settlement-agents Specification

## Purpose

Define the settlement agent state vector, decision loop logic, and deterministic RNG isolation for agent-driven simulation.

## ADDED Requirements

### Requirement: Agent State Vector

The system SHALL define an agent state vector for each settlement comprising MilitaryStrength, Wealth, Relations, and Goals.

#### Scenario: State vector fields

- **WHEN** a settlement is created during world generation
- **THEN** it SHALL have `MilitaryStrength float64` initialized to population × 0.1
- **THEN** it SHALL have `Wealth float64` initialized to 100.0 (or config value)
- **THEN** it SHALL have `Relations map[string]float64` initialized via `initRelations()` with same-faction baseline +0.3
- **THEN** it SHALL have `Goals []string` initialized to 2–3 unique values from ["grow", "defend", "expand"]

#### Scenario: State vector serialization

- **WHEN** a world state is serialized to JSON
- **THEN** all agent state fields SHALL be included in settlement JSON
- **THEN** round-trip deserialization SHALL produce equivalent state vector values

### Requirement: Agent Decision Loop

The system SHALL execute a decision loop for each settlement each simulation year, selecting and executing one action.

#### Scenario: Annual decision cycle

- **WHEN** the simulation ticks a settlement for a new year
- **THEN** the settlement SHALL evaluate all six actions (Expand, Raid, Conquer, Fortify, Ally, Prosper)
- **THEN** actions failing preconditions SHALL be excluded from consideration
- **THEN** remaining actions SHALL be scored based on goal alignment
- **THEN** one action SHALL be selected via weighted random using the settlement's agent RNG
- **THEN** the selected action SHALL be executed, modifying state and emitting an event

#### Scenario: Goal-based scoring

- **WHEN** an action is scored for a settlement
- **THEN** Expand SHALL score high if "expand" is in the settlement's Goals
- **THEN** Fortify SHALL score high if "defend" is in the settlement's Goals
- **THEN** Prosper and Fortify SHALL score high if "grow" is in the settlement's Goals
- **THEN** actions SHALL score low (baseline 1.0) if no goal alignment

#### Scenario: Weighted random selection

- **WHEN** multiple actions pass preconditions
- **THEN** the settlement SHALL select one action via weighted random, where weight = goal score
- **THEN** the selection SHALL use the settlement's agent RNG for determinism
- **THEN** same seed SHALL produce identical action selection across runs

### Requirement: Agent RNG Isolation

The system SHALL derive a settlement-scoped agent RNG separate from the figure RNG.

#### Scenario: Agent RNG derivation

- **WHEN** a settlement is initialized for simulation
- **THEN** an agent RNG SHALL be derived via `engine.GetPRNG("agent:" + settlement.Name)`
- **THEN** the agent RNG SHALL be separate from the figure RNG (`"figures:" + settlement.Name`)
- **THEN** all agent operations (action selection, relation shifts, expand target selection) SHALL use the agent RNG

#### Scenario: Agent RNG determinism

- **WHEN** the same seed is used across multiple runs
- **THEN** the agent RNG SHALL produce identical sequences
- **THEN** agent decisions SHALL be identical across runs

### Requirement: Default Fallback Action

The system SHALL default to Prosper action when no other actions pass preconditions.

#### Scenario: No valid actions

- **WHEN** a settlement evaluates all six actions and none pass preconditions
- **THEN** Prosper SHALL be selected as the default action
- **THEN** Prosper SHALL execute, increasing population and wealth based on suitability
