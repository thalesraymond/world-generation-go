## ADDED Requirements

### Requirement: Iterative Chronological Engine
The simulation engine SHALL advance time in discrete increments (e.g., years) and process updates for all world entities sequentially within that year.

#### Scenario: Advancing time
- **WHEN** the simulation loop ticks for a new year
- **THEN** all registered world entities process their temporal updates and emit relevant events.

#### Scenario: Deterministic execution
- **WHEN** the simulation runs multiple times with the same seed
- **THEN** the sequence of entity updates and emitted events SHALL remain identical across runs.
