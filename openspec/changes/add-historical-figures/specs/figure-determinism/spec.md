# figure-determinism Specification

## Purpose

Define settlement-scoped RNG for all figure operations ensuring byte-identical figures, events, and exports across runs with the same seed.

## ADDED Requirements

### Requirement: Settlement-Scoped Figure RNG

Each settlement SHALL derive its figure-related PRNG from the master seed using a settlement-specific component identifier in the `state.Engine`.

#### Scenario: Deriving settlement figure RNG

- **WHEN** a settlement needs to generate figure data (birth, names, roles, lifespan)
- **THEN** a `*randv2.Rand` SHALL be derived from the master seed via `engine.GetPRNG("figures:" + settlementName)`
- **THEN** all figure operations for that settlement SHALL draw from this single RNG stream

#### Scenario: RNG isolation between settlements

- **WHEN** two settlements generate figures
- **THEN** the RNG state of one settlement's figures SHALL NOT affect the other settlement's figure attributes
- **THEN** adding or removing a settlement from the world SHALL NOT affect the figures of remaining settlements

### Requirement: Deterministic Name Generation

Figure names SHALL be generated deterministically from name tables using the settlement's figure RNG.

#### Scenario: Same seed produces same names

- **WHEN** two simulations run with the same master seed
- **THEN** every figure in every settlement SHALL have the same generated name
- **THEN** name generation SHALL use deterministic combinatorial selection (first-name table + surname/epithet table) via the settlement's figure RNG

#### Scenario: Different seed produces different names

- **WHEN** two simulations run with different master seeds
- **THEN** figures SHALL generally have different names (save for coincidental collisions in small name tables)

### Requirement: Deterministic Lifecycle

All figure lifecycle events (birth timing, role assignment, marriage, death) SHALL be deterministic given the same seed.

#### Scenario: Same seed, same births

- **WHEN** a simulation runs with seed S producing figure F in year Y at settlement A
- **THEN** running again with seed S SHALL produce the same figure F with the same birth year Y at settlement A

#### Scenario: Same seed, same lifespans

- **WHEN** a simulation runs with seed S and figure F dies in year D
- **THEN** running again with seed S SHALL produce figure F dying in the same year D

#### Scenario: Same seed, same roles

- **WHEN** a simulation runs with seed S and figure F is assigned role R in year Y
- **THEN** running again with seed S SHALL assign figure F the same role R in the same year Y

### Requirement: Deterministic Event Generation

Role-based event generation SHALL produce the same events across seed-identical runs.

#### Scenario: Same seed, same events

- **WHEN** two simulations run with the same master seed
- **THEN** the sequence of figure-generated events SHALL be byte-identical
- **THEN** event descriptions, categories, and figure references SHALL match exactly

#### Scenario: Deterministic JSON output

- **WHEN** a world state with figures is serialized to JSON across two seed-identical runs
- **THEN** the serialized JSON SHALL be byte-identical (`bytes.Equal`)

### Requirement: Deterministic Export

Obsidian character file export SHALL produce identical files across seed-identical runs.

#### Scenario: Same seed, same export

- **WHEN** two `worldgen export` runs process world states generated from the same seed
- **THEN** the `characters/` directory content SHALL be byte-identical across runs
- **THEN** character file frontmatter, body content, and wiki-links SHALL match exactly

### Requirement: RNG Order Consistency

All figure operations within a single settlement tick SHALL occur in a deterministic, stable order.

#### Scenario: Stable ordering

- **WHEN** a settlement ticks for a year
- **THEN** figures SHALL be processed in a fixed order (by creation order or ID)
- **THEN** aging, death checks, births, role assignment, and event generation SHALL occur in the same sequence each run