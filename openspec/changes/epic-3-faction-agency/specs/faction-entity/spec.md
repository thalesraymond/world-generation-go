## ADDED Requirements

### Requirement: Faction Identity
Factions SHALL have a unique identifier, a display name, and a cultural identity grouping.

#### Scenario: Creating a faction
- **WHEN** a faction is initialized with a name ("Ashfield Confederacy") and cultural identity ("Northern Kingdoms", "Martial")
- **THEN** it has a unique string ID, the provided name, and the provided identity attributes.

#### Scenario: Faction ID uniqueness
- **WHEN** multiple factions are created in the same world
- **THEN** each faction has a distinct ID regardless of name collisions.

### Requirement: Faction Leadership
Factions SHALL have a leader referenced by figure ID. The leader is the historical figure with the highest relevant skill among member settlements.

#### Scenario: Initial leader selection
- **WHEN** a faction is formed from founding settlements containing historical figures
- **THEN** the figure with the highest combined martial and diplomatic skill is selected as leader.

#### Scenario: Leader succession on death
- **WHEN** the current faction leader dies
- **THEN** the living figure with the highest relevant skill among remaining member settlements becomes the new leader.

#### Scenario: Faction with no living figures
- **WHEN** all figures in a faction's member settlements have died and no new figures exist
- **THEN** the faction has no leader and enters a "leaderless" state that reduces action success probability.

### Requirement: Faction Treasury
Factions SHALL maintain a shared treasury representing abstract wealth.

#### Scenario: Initial treasury
- **WHEN** a faction is formed
- **THEN** its treasury equals the sum of founding member settlements' wealth values.

#### Scenario: Treasury income from members
- **WHEN** the simulation advances one year
- **THEN** each member settlement contributes 10% of its current wealth to the faction treasury.

#### Scenario: War cost
- **WHEN** a faction is at war
- **THEN** it loses 100 treasury per year for the duration of the war.

#### Scenario: Alliance maintenance cost
- **WHEN** a faction has active alliances
- **THEN** it loses 20 treasury per year per active alliance.

#### Scenario: Treasury bankruptcy
- **WHEN** a faction's treasury reaches zero or below
- **THEN** it enters a "desperate" state: member settlements have elevated defection probability, and the faction cannot declare new wars.

### Requirement: Faction Members
Factions SHALL track which settlements belong to them.

#### Scenario: Founding members
- **WHEN** a faction is created from a set of settlements
- **THEN** those settlements are recorded as initial members.

#### Scenario: Adding a member
- **WHEN** a settlement joins a faction through conquest or defection
- **THEN** the settlement is added to the faction's member list and the joining event is recorded in faction history.

#### Scenario: Removing a member
- **WHEN** a settlement leaves a faction (conquered, defected, or absorbed)
- **THEN** the settlement is removed from the faction's member list and the departure is recorded in faction history.

#### Scenario: Faction collapse
- **WHEN** a faction has zero member settlements
- **THEN** the faction is dissolved: removed from the active faction registry, its historical record preserved.

### Requirement: Faction Relations
Factions SHALL maintain relationship scores with other factions ranging from −1.0 (hostile) to +1.0 (allied).

#### Scenario: Default relations
- **WHEN** a faction is created
- **THEN** its relations with all existing factions are initialized to 0.0 (neutral).

#### Scenario: War impact on relations
- **WHEN** a faction declares war on another faction
- **THEN** the declaring faction's relation with the target drops to −1.0, and the target's relation with the declarer also drops to −1.0.

#### Scenario: Alliance impact on relations
- **WHEN** two factions form an alliance
- **THEN** both factions set their mutual relation to +0.8.

#### Scenario: Relation time decay
- **WHEN** the simulation advances one year and a relation value has not been modified by an action
- **THEN** the relation drifts toward 0.0 by 0.01 (e.g., 0.5 becomes 0.49, −0.3 becomes −0.29).

#### Scenario: Trade adjacency bonus
- **WHEN** two factions have member settlements adjacent to each other and neither is at war
- **THEN** their mutual relation increases by 0.02 per year.

### Requirement: Faction Goals
Factions SHALL have strategic goals that influence their action selection.

#### Scenario: Goal-driven action selection
- **WHEN** a faction evaluates strategic options during its annual tick
- **THEN** it weights available actions based on how well they advance its current goals (e.g., "expand" goal favors war/conquest, "prosper" goal favors alliances/trade).

### Requirement: Faction Policy
Factions SHALL have a current policy that influences member settlement behavior.

#### Scenario: Policy types
- **WHEN** a faction sets its policy
- **THEN** the policy is one of: "expansion", "defense", or "diplomacy".

#### Scenario: Expansion policy effect
- **WHEN** a faction's policy is "expansion"
- **THEN** member settlements receive a weight bonus toward "expand" and "conquer" actions.

#### Scenario: Defense policy effect
- **WHEN** a faction's policy is "defense"
- **THEN** member settlements receive a weight bonus toward "fortify" actions.

#### Scenario: Diplomacy policy effect
- **WHEN** a faction's policy is "diplomacy"
- **THEN** member settlements receive a weight bonus toward "ally" and "prosper" actions.

### Requirement: Faction History
Factions SHALL record membership changes and strategic decisions with year, actors, and cause.

#### Scenario: Recording a membership change
- **WHEN** a settlement joins or leaves a faction
- **THEN** a history entry is recorded with the year, settlement name, source faction, destination faction, and cause ("conquest", "defection", "collapse").

#### Scenario: Recording a strategic decision
- **WHEN** a faction declares war, forms an alliance, or changes policy
- **THEN** a history entry is recorded with the year, action type, and involved parties.

### Requirement: Faction JSON Serialization
Factions SHALL be serializable to and from JSON for world state persistence.

#### Scenario: Round-trip serialization
- **WHEN** a faction is serialized to JSON and deserialized
- **THEN** all fields (ID, name, identity, leader, treasury, goals, members, relations, policy, history) are preserved exactly.
