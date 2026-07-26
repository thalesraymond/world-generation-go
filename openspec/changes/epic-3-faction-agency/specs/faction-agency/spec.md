## ADDED Requirements

### Requirement: Strategic Decision Loop
Factions SHALL implement the `simulation.Entity` interface and evaluate their state each simulation year.

#### Scenario: Annual faction tick
- **WHEN** the simulation advances one year
- **THEN** each faction evaluates its health (member count, treasury, military aggregate, relations) and may choose one strategic action.

#### Scenario: No eligible actions
- **WHEN** a faction evaluates its state and no action preconditions are met
- **THEN** the faction takes no action and emits no event for that year.

#### Scenario: Deterministic tick order
- **WHEN** multiple factions exist in the world
- **THEN** factions are ticked in deterministic order (sorted by ID) to ensure reproducible simulation.

### Requirement: War Declaration
Factions SHALL be able to declare war on other factions when preconditions are met.

#### Scenario: Declaring war
- **WHEN** a faction's target faction exists, the mutual relation is below −0.3, and the declaring faction's military aggregate exceeds the target's
- **THEN** the declaring faction declares war: mutual relations drop to −1.0, member settlements enter defensive posture, and a war-drain of 100 treasury per year begins.

#### Scenario: War declaration event
- **WHEN** a faction declares war on another faction
- **THEN** an event is emitted with year, category "war", and description "Faction A declared war on Faction B".

#### Scenario: Insufficient military for war
- **WHEN** a faction's military aggregate is not greater than the target's
- **THEN** the faction cannot declare war on that target (precondition not met).

#### Scenario: Cannot declare war on ally
- **WHEN** a faction's relation with a target is above +0.5 (allied)
- **THEN** the faction cannot declare war on that target.

### Requirement: Alliance Formation
Factions SHALL be able to form alliances with other factions when preconditions are met.

#### Scenario: Forming an alliance
- **WHEN** a faction targets another faction with relation above +0.5, no active war between them, and the proposing faction's treasury exceeds 200
- **THEN** both factions set their mutual relation to +0.8, an alliance maintenance cost of 20 treasury per year begins, and member settlements gain a morale bonus.

#### Scenario: Alliance formation event
- **WHEN** two factions form an alliance
- **THEN** an event is emitted with year, category "alliance", and description "Faction A formed an alliance with Faction B".

#### Scenario: Alliance blocked by active war
- **WHEN** two factions are currently at war
- **THEN** neither can propose an alliance to the other.

#### Scenario: Alliance blocked by insufficient treasury
- **WHEN** a faction's treasury is at or below 200
- **THEN** the faction cannot propose new alliances.

### Requirement: Policy Setting
Factions SHALL be able to set a strategic policy that influences member settlement behavior.

#### Scenario: Setting policy
- **WHEN** a faction chooses to set a policy
- **THEN** its policy changes to one of "expansion", "defense", or "diplomacy", and an event is emitted.

#### Scenario: Policy influences member actions
- **WHEN** a member settlement evaluates its own action choices during its tick
- **THEN** the faction's current policy adjusts the weight of eligible actions (expansion boosts expand/conquer, defense boosts fortify, diplomacy boosts ally/prosper).

#### Scenario: Policy setting event
- **WHEN** a faction changes its policy
- **THEN** an event is emitted with year, category "policy" and description indicating the new policy.

### Requirement: Action Preconditions
Each strategic action SHALL have clearly defined preconditions that must be satisfied before execution.

#### Scenario: All preconditions met
- **WHEN** a faction selects an action and all preconditions are satisfied
- **THEN** the action executes, consequences are applied, and an event is emitted.

#### Scenario: Preconditions not met
- **WHEN** a faction selects an action but one or more preconditions are not satisfied
- **THEN** the action is skipped and the faction evaluates the next candidate action.

### Requirement: Faction RNG Isolation
Each faction SHALL receive a deterministic RNG derived from the master seed, independent of other factions.

#### Scenario: Deterministic faction decisions
- **WHEN** the simulation is run twice with the same master seed
- **THEN** each faction makes identical strategic decisions across both runs.

#### Scenario: Independent RNG streams
- **WHEN** multiple factions exist in the same simulation
- **THEN** consuming an RNG value from one faction does not affect the RNG state of any other faction.

### Requirement: Faction Event Emission
Factions SHALL emit events through the simulation event channel for all strategic actions.

#### Scenario: Event format
- **WHEN** a faction performs a strategic action
- **THEN** an event is emitted with the correct year, a category ("war", "alliance", "policy"), and a description including both faction names.

#### Scenario: No event on inaction
- **WHEN** a faction takes no strategic action in a year
- **THEN** no event is emitted for that faction in that year.
