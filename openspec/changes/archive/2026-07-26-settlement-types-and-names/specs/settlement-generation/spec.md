## MODIFIED Requirements

### Requirement: Settlement Identity and Faction

Generated settlements SHALL be named using deterministic combinatorial generation drawing from prefix and suffix tables. A settlement SHALL be assigned a type based on population thresholds. A settlement SHALL inherit the source tile's faction; a tile without faction influence SHALL produce an `independent` settlement.

#### Scenario: Named settlement generation

- **WHEN** a settlement is founded
- **THEN** it receives a generated name composed from combinatorial name tables using the settlements RNG

#### Scenario: Typed settlement generation

- **WHEN** a settlement is founded with a population value
- **THEN** it receives a type classification appropriate to its population

#### Scenario: Unclaimed settlement

- **WHEN** an eligible source tile has no faction influence
- **THEN** the generated settlement faction is `independent`