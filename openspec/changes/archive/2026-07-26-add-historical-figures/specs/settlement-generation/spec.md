# settlement-generation Delta Specification

## MODIFIED Requirements

### Requirement: Settlement Identity and Faction

Generated settlements SHALL be named using deterministic combinatorial generation drawing from prefix and suffix tables. A settlement SHALL be assigned a type based on population thresholds. A settlement SHALL inherit the source tile's faction; a tile without faction influence SHALL produce an `independent` settlement. At creation time, 3–5 founding historical figures SHALL be generated for the settlement.

#### Scenario: Named settlement generation

- **WHEN** a settlement is founded
- **THEN** it receives a generated name composed from combinatorial name tables using the settlements RNG

#### Scenario: Typed settlement generation

- **WHEN** a settlement is founded with a population value
- **THEN** it receives a type classification appropriate to its population

#### Scenario: Unclaimed settlement

- **WHEN** an eligible source tile has no faction influence
- **THEN** the generated settlement faction is `independent`

#### Scenario: Founding figures generated

- **WHEN** a settlement is founded during world generation
- **THEN** between 3 and 5 historical figures SHALL be generated as founders
- **THEN** one founder SHALL be assigned the Leader role
- **THEN** the remaining founders SHALL be assigned Explorer or no role
- **THEN** founders SHALL have the settlement's founding year as their birth year (or an offset of up to 20 years prior for older founders)

## ADDED Requirements

### Requirement: Settlement Figure Generation RNG

Settlement generation SHALL derive and store a figure-specific RNG for each settlement.

#### Scenario: Figure RNG derivation

- **WHEN** a settlement is generated
- **THEN** a figure-specific `*randv2.Rand` SHALL be derived from the master seed using the settlement name as part of the component identifier
- **THEN** the RNG SHALL be stored or derivable for use during simulation figure lifecycle processing