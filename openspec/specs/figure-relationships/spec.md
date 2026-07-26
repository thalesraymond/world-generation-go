# figure-relationships Specification

## Purpose

Define family tree relationships (parent/child, spouse), marriage formation, inheritance/succession on leader death, and relationship persistence and serialization.

## Requirements

### Requirement: Relationship Data Model

The system SHALL store family relationships on each `HistoricalFigure` as a `Relationships` struct containing parent IDs, child IDs, and spouse IDs.

#### Scenario: Relationship fields on figure

- **WHEN** a historical figure is created
- **THEN** the figure SHALL have a `Relationships` field with `Parents []string`, `Children []string`, and `Spouse []string` ID arrays
- **THEN** each relationship ID SHALL reference another figure's unique identifier (settlement-qualified to avoid cross-settlement collisions)

#### Scenario: Relationship serialization

- **WHEN** a figure with relationships is serialized to JSON
- **THEN** the `relationships` field SHALL include parent, child, and spouse IDs
- **THEN** round-trip deserialization SHALL preserve all relationship references

### Requirement: Parent–Child Relationships

The system SHALL establish parent–child relationships at birth and track lineage for succession.

#### Scenario: Newborn figure has parents assigned

- **WHEN** a new figure is born to a settlement
- **THEN** the newborn SHALL be assigned parent figures from the settlement (typically 1–2 parents)
- **THEN** the assigned parent(s) SHALL have their `Children` list updated to include the newborn's ID

#### Scenario: Founding figure parents

- **WHEN** founding figures are created at settlement creation
- **THEN** founders SHALL NOT have parents assigned (they are the first generation)
- **THEN** founders MAY form parent–child relationships with each other through later marriage and births

### Requirement: Spousal Relationships

The system SHALL establish marriage relationships between figures, enabling dynastic and political connections.

#### Scenario: Marriage formation at settlement creation

- **WHEN** a settlement is founded with multiple founding figures
- **THEN** pairs of founders MAY be assigned spousal relationships through deterministic RNG decisions
- **THEN** each spouse SHALL have the other's ID in their `Spouse` list

#### Scenario: Marriage over time

- **WHEN** new figures reach adulthood (age ~18–25) during simulation
- **THEN** they MAY be married to figures in other settlements (same faction) or within their own settlement
- **THEN** marriages between figures of allied factions SHALL be possible and generate alliance events

### Requirement: Inheritance and Succession

The system SHALL use parent–child relationships to determine leadership succession when a leader dies.

#### Scenario: Firstborn inherits leadership

- **WHEN** a settlement's leader dies and has living children
- **THEN** the eldest living child SHALL be the primary succession candidate
- **THEN** if the eldest child is already dead, the next eldest living child SHALL be selected

#### Scenario: No children — random successor

- **WHEN** a settlement's leader dies and has no living children
- **THEN** a random adult figure from the settlement SHALL be assigned the Leader role
- **THEN** a succession event SHALL note that the new leader is not a direct heir

#### Scenario: Succession ties to spouse

- **WHEN** a leader dies and has a living spouse
- **THEN** the spouse MAY be in the pool of successor candidates (as an alternative to children or random figures)
- **THEN** if no children exist and spouse is selected, a "Leadership passed to spouse" event SHALL be emitted
