# simulation-loop Specification

## Purpose

Define the discrete-year simulation engine that advances world entities, processes settlement figure lifecycles deterministically, and emits timeline events.

## Requirements

### Requirement: Iterative Chronological Engine

The simulation engine SHALL advance time in discrete increments (e.g., years) and process updates for all world entities sequentially within that year. Entities that contain settlements with historical figures SHALL process figure lifecycle updates (aging, births, deaths, role assignment, event generation) during their tick.

#### Scenario: Advancing time

- **WHEN** the simulation loop ticks for a new year
- **THEN** all registered world entities process their temporal updates and emit relevant events

#### Scenario: Deterministic execution

- **WHEN** the simulation runs multiple times with the same seed
- **THEN** the sequence of entity updates and emitted events SHALL remain identical across runs

### Requirement: Figure Lifecycle Processing in Entity Tick

Each settlement entity SHALL process its figure lifecycle as part of its `Tick()` method, in a deterministic order.

#### Scenario: Settlement tick with figures

- **WHEN** a settlement entity with historical figures ticks for a year
- **THEN** all living figures SHALL be aged by one year
- **THEN** death checks (age-based and event risk) SHALL be performed for each living figure in creation order
- **THEN** new figure births SHALL be evaluated based on settlement population and active figure count
- **THEN** role vacancy checks SHALL be performed (leaderless settlement triggers succession)
- **THEN** each figure with a role SHALL generate role-specific events

#### Scenario: Settlement tick processing order

- **WHEN** a settlement entity ticks for a year
- **THEN** figure processing SHALL occur after the core settlement events and before returning
- **THEN** the order SHALL be: age figures, check deaths, check births, assign roles, check marriages, generate role events, check role transitions, run settlement agent decision loop
