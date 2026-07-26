## ADDED Requirements

### Requirement: Conquest-Driven Faction Change
Settlements SHALL change faction allegiance when conquered by a settlement from another faction.

#### Scenario: Successful conquest
- **WHEN** a settlement-agent executes the "conquer" action against a settlement belonging to another faction and succeeds
- **THEN** the conquered settlement's faction changes to the conqueror's faction, both factions' histories record the change with cause "conquest", and the attacker faction's relation with the victim faction decreases by 0.3.

#### Scenario: Conquest event
- **WHEN** a settlement is conquered and changes faction
- **THEN** an event is emitted with year, category "conquest", and description "Settlement X joined Faction A after being conquered from Faction B".

#### Scenario: Failed conquest
- **WHEN** a settlement-agent executes the "conquer" action but fails (preconditions not met or action check fails)
- **THEN** no faction change occurs and no membership event is recorded.

### Requirement: Diplomatic Defection
Settlements SHALL be able to voluntarily defect to a neighboring faction under favorable conditions.

#### Scenario: Defection to friendly faction
- **WHEN** a settlement's current faction has a treasury below 50 (desperate state), a neighboring faction has a positive relation (> 0.5) with the settlement's current faction, and the neighboring faction's policy is "diplomacy"
- **THEN** the settlement may defect: its faction changes to the neighboring faction, both factions record the change with cause "defection".

#### Scenario: Defection event
- **WHEN** a settlement defects to another faction
- **THEN** an event is emitted with year, category "defection", and description "Settlement X defected from Faction A to Faction B".

#### Scenario: No defection when stable
- **WHEN** a settlement's current faction has treasury above 50 and is not in a desperate state
- **THEN** the settlement cannot defect through diplomatic means.

### Requirement: Faction Dissolution
Factions SHALL be dissolved when they have no remaining member settlements.

#### Scenario: Faction collapse
- **WHEN** a faction loses its last member settlement (through conquest, defection, or destruction)
- **THEN** the faction is removed from the active registry, its history is preserved, and an event is emitted with category "dissolution".

#### Scenario: Dissolution event
- **WHEN** a faction is dissolved
- **THEN** an event is emitted with year, category "dissolution", and description "Faction A has dissolved — no settlements remain".

#### Scenario: Dissolved faction cannot act
- **WHEN** a faction has been dissolved
- **THEN** it is no longer ticked by the simulation and cannot be targeted by war declarations or alliance proposals.

### Requirement: New Faction Formation
New factions SHALL be formed when a group of settlements breaks away from an existing faction.

#### Scenario: Breakaway faction
- **WHEN** a settlement with a high-influence figure (martial or diplomatic skill > 80) belongs to a faction with a treasury below 30 and at least 3 member settlements, the settlement may break away and form a new faction with adjacent settlements that share high mutual relations (> 0.6)
- **THEN** a new faction is created, the breakaway settlements join it, and an event is emitted with category "formation".

#### Scenario: Formation event
- **WHEN** a new faction is formed through breakaway
- **THEN** an event is emitted with year, category "formation", and description "Faction A was formed by settlements breaking away from Faction B".

#### Scenario: Breakaway blocked by strong faction
- **WHEN** a settlement's current faction has a treasury above 100 and at least one active alliance
- **THEN** breakaway is not possible for member settlements.

### Requirement: Membership Change History
All faction membership changes SHALL be recorded with sufficient detail for narrative reconstruction.

#### Scenario: History entry completeness
- **WHEN** a membership change occurs (conquest, defection, collapse, or formation)
- **THEN** a history entry is recorded with: year, settlement name, source faction ID, destination faction ID, and cause.

#### Scenario: History ordering
- **WHEN** multiple membership changes occur across different years
- **THEN** history entries are maintained in chronological order within each faction.

#### Scenario: History persistence
- **WHEN** a faction is serialized to JSON
- **THEN** its complete membership change history is included in the serialization.

### Requirement: World State Faction Registry
The world state SHALL maintain a registry of active faction entities replacing the string-based influence grid.

#### Scenario: Faction registry initialization
- **WHEN** the world state is initialized for simulation
- **THEN** it contains a `Factions` map of faction ID to `*faction.Faction` entities, replacing the deprecated `FactionInfluence` string grid.

#### Scenario: Settlement faction reference
- **WHEN** a settlement belongs to a faction
- **THEN** its `Faction` field references the faction's ID string (not a display name).

#### Scenario: Unaffiliated settlements
- **WHEN** a settlement has no faction affiliation
- **THEN** its `Faction` field is an empty string, indicating independence.
