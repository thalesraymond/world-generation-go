# historical-figures Specification

## Purpose

Define the HistoricalFigure data model, lifecycle (birth, aging, death), generation mechanics, settlement embedding, and serialization in world state.

## ADDED Requirements

### Requirement: Historical Figure Data Model

The system SHALL define a `HistoricalFigure` struct with identity, life, role, and relationships. Each figure SHALL store a unique identifier, a generated name, birth year, death year (zero while alive), faction, role, settlement reference, and family relationships.

#### Scenario: Figure creation with identity

- **WHEN** a historical figure is created for a settlement during world generation
- **THEN** the figure SHALL be assigned a unique ID derived from the settlement name and figure index
- **THEN** the figure SHALL be assigned a generated name using deterministic name tables
- **THEN** the figure SHALL be assigned the settlement's faction
- **THEN** the figure SHALL have a birth year set to the settlement founding year or a relative offset

#### Scenario: Figure fields in JSON serialization

- **WHEN** a world state containing figures is serialized to JSON
- **THEN** the `figures` array on each settlement SHALL include all figure fields (id, name, birthYear, deathYear, faction, role, relationships)
- **THEN** round-trip deserialization SHALL produce equivalent figure data

### Requirement: Figure Lifecycle — Birth

The system SHALL generate founding figures at settlement creation and spawn new figures over time during simulation.

#### Scenario: Founding figures at settlement creation

- **WHEN** a settlement is founded during world generation
- **THEN** between 3 and 5 founding historical figures SHALL be created for the settlement
- **THEN** at least one founder SHALL hold the Leader role
- **THEN** the remaining founders MAY hold Explorer or no role

#### Scenario: Population-scaled births over time

- **WHEN** the simulation advances a year and a settlement's population exceeds the birth threshold
- **THEN** a new figure MAY be born to the settlement with probability proportional to settlement population
- **THEN** birth probability SHALL decrease as the number of active (alive) figures approaches 10–15, preventing unbounded growth
- **THEN** new figures SHALL be born roleless or assigned roles based on settlement needs

### Requirement: Figure Lifecycle — Aging

The system SHALL age all living figures by one year each simulation tick.

#### Scenario: Annual aging

- **WHEN** the simulation engine ticks a settlement for a new year
- **THEN** every living figure (death year is zero) in that settlement SHALL have its age incremented by one year

### Requirement: Figure Lifecycle — Death

The system SHALL determine figure death through age-based lifespan limits combined with event-based mortality risk.

#### Scenario: Age-based death

- **WHEN** a figure's age reaches or exceeds its maximum lifespan (determined at birth from the range 70–90 years)
- **THEN** the figure SHALL die, and its death year SHALL be set to the current simulation year

#### Scenario: Event risk death

- **WHEN** the simulation processes a figure's annual tick and the figure's age is between 30 and its maximum lifespan
- **THEN** there SHALL be a low probability (approximately 1–2%) that the figure dies from a random event (illness, accident, conflict)
- **THEN** if the figure dies, its death year SHALL be set to the current simulation year
- **THEN** a death event SHALL be emitted to the timeline

#### Scenario: Death outside normal range

- **WHEN** a figure's age falls below 30
- **THEN** the figure SHALL NOT die from event risk (only through explicit narrative events like war or disaster)

### Requirement: Figure Population Management

The system SHALL cap the maximum number of active (living) figures per settlement to prevent combinatorial explosion.

#### Scenario: Figure cap enforcement

- **WHEN** a settlement has 10–15 or more living figures
- **THEN** new births SHALL NOT occur until the living figure count falls below the cap

#### Scenario: Figure cap configurable

- **WHEN** a world is generated with default configuration
- **THEN** the figure cap SHALL be between 10 and 15 active figures per settlement

### Requirement: Serialization Compatibility

The `Settlement` struct in `world.State` SHALL include a `Figures` field that serializes as a JSON array. Existing settlement fields SHALL remain unchanged.

#### Scenario: JSON round-trip with figures

- **WHEN** a state with settlements containing figures is serialized and deserialized
- **THEN** all settlement fields including figures SHALL be equivalent
- **THEN** deserialization of a state without figures (older format) SHALL produce an empty `Figures` slice, not an error

#### Scenario: Backward compatibility

- **WHEN** a world_state.json generated without historical figures is loaded by a version that supports figures
- **THEN** the system SHALL NOT require the `Figures` field to be present
- **THEN** settlements SHALL have an empty or nil `Figures` slice