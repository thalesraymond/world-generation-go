# agent-actions Specification

## Purpose

Define the six core agent actions (Expand, Raid, Conquer, Fortify, Ally, Prosper) with preconditions, execution logic, consequences, and event generation.

## ADDED Requirements

### Requirement: Expand Action

The system SHALL implement an Expand action that founds new settlements in unclaimed suitable tiles.

#### Scenario: Expand precondition check

- **WHEN** a settlement considers the Expand action
- **THEN** the action SHALL pass precondition if settlement population > 50.0
- **THEN** the action SHALL pass precondition if settlement wealth > 200.0 (expansion cost)
- **THEN** the action SHALL pass precondition if at least one unclaimed suitable tile exists within range (via pointcrawl query)
- **THEN** the action SHALL fail precondition if any condition is false

#### Scenario: Expand execution

- **WHEN** the Expand action is executed
- **THEN** a new settlement SHALL be created at the target tile with: Name (generated), X/Y (target coordinates), Faction (parent's faction), Population (20% of parent), MilitaryStrength (20% of parent), Wealth (30% of parent's wealth), Relations (initialized via `initRelations()`), Goals (randomized)
- **THEN** the parent settlement's wealth SHALL decrease by 200.0 (expansion cost)
- **THEN** the new settlement SHALL be appended to the world's settlement slice
- **THEN** an Expansion event SHALL be emitted with Description: "{parent} founded {newSettlement}"

#### Scenario: Expand with no targets

- **WHEN** the Expand action is executed but no suitable targets exist
- **THEN** no new settlement SHALL be created
- **THEN** no wealth SHALL be spent
- **THEN** a failure event SHALL be emitted with Description: "{settlement} expansion failed: no suitable targets"

### Requirement: Raid Action

The system SHALL implement a Raid action that steals wealth from hostile neighbors.

#### Scenario: Raid precondition check

- **WHEN** a settlement considers the Raid action
- **THEN** the action SHALL pass precondition if settlement military > target military × 0.8
- **THEN** the action SHALL pass precondition if settlement relations to target < −0.5 (hostile)
- **THEN** the action SHALL pass precondition if target is within range (Euclidean distance ≤ 10 tiles)
- **THEN** the action SHALL fail precondition if any condition is false

#### Scenario: Raid execution success

- **WHEN** the Raid action is executed with 70% success probability (weighted random via agent RNG)
- **THEN** on success: target wealth SHALL decrease by 50.0, raider wealth SHALL increase by 50.0, raider relations to target SHALL shift −0.4, target relations to raider SHALL shift −0.3
- **THEN** a Raid event SHALL be emitted with Category "Raid", TargetSettlement (target name), Outcome "success", Amount "50"

#### Scenario: Raid execution failure

- **WHEN** the Raid action fails (30% probability)
- **THEN** no wealth SHALL be transferred
- **THEN** raider relations to target SHALL shift −0.2 (penalty for failed raid)
- **THEN** a Raid event SHALL be emitted with Category "Raid", TargetSettlement (target name), Outcome "failure"

### Requirement: Conquer Action

The system SHALL implement a Conquer action that militarily absorbs weaker neighbors.

#### Scenario: Conquer precondition check

- **WHEN** a settlement considers the Conquer action
- **THEN** the action SHALL pass precondition if settlement military > target military × 1.5
- **THEN** the action SHALL pass precondition if settlement relations to target < −0.7 (very hostile)
- **THEN** the action SHALL pass precondition if target is within range (Euclidean distance ≤ 10 tiles)
- **THEN** the action SHALL fail precondition if any condition is false

#### Scenario: Conquer execution

- **WHEN** the Conquer action is executed
- **THEN** the target settlement's faction SHALL change to the attacker's faction
- **THEN** the attacker's military SHALL decrease by 20% (war cost)
- **THEN** attacker relations to target SHALL shift −0.8
- **THEN** target relations to attacker SHALL shift −0.8
- **THEN** a Conquest event SHALL be emitted with Category "Conquest", TargetSettlement (target name), Description: "{attacker} conquered {target}"

### Requirement: Fortify Action

The system SHALL implement a Fortify action that invests wealth into military strength.

#### Scenario: Fortify precondition check

- **WHEN** a settlement considers the Fortify action
- **THEN** the action SHALL pass precondition if settlement wealth > 100.0
- **THEN** the action SHALL fail precondition if wealth ≤ 100.0

#### Scenario: Fortify execution

- **WHEN** the Fortify action is executed
- **THEN** the settlement SHALL convert 100.0 wealth into military strength (military += 100.0, wealth -= 100.0)
- **THEN** an Economy event SHALL be emitted with Category "Economy", Description: "{settlement} invests in fortifications"

### Requirement: Ally Action

The system SHALL implement an Ally action that proposes alliances with friendly settlements.

#### Scenario: Ally precondition check

- **WHEN** a settlement considers the Ally action
- **THEN** the action SHALL pass precondition if settlement relations to target > 0.5 (friendly)
- **THEN** the action SHALL pass precondition if no existing alliance flag is set between settlements
- **THEN** the action SHALL fail precondition if any condition is false

#### Scenario: Ally execution

- **WHEN** the Ally action is executed
- **THEN** an alliance flag SHALL be set between the two settlements (implementation: relations capped at minimum +0.6)
- **THEN** both settlements' relations to each other SHALL shift +0.4
- **THEN** a Diplomacy event SHALL be emitted with Category "Diplomacy", TargetSettlement (target name), Description: "{settlement} forms alliance with {target}"

### Requirement: Prosper Action

The system SHALL implement a Prosper action for passive growth of population and wealth.

#### Scenario: Prosper precondition check

- **WHEN** a settlement considers the Prosper action
- **THEN** the action SHALL always pass precondition (default fallback)

#### Scenario: Prosper execution

- **WHEN** the Prosper action is executed
- **THEN** the settlement's population SHALL increase by 2.0 × suitability score
- **THEN** the settlement's wealth SHALL increase by 5.0 × suitability score
- **THEN** the settlement's relations to all other settlements SHALL shift +0.05 (gradual warming)
- **THEN** an Economy event SHALL be emitted with Category "Economy", Description: "{settlement} prospers"
