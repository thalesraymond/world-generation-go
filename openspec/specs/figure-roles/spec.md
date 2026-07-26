# figure-roles Specification

## Purpose

Define the modular `Role` interface, the `Leader` and `Explorer` implementations, role assignment logic (settlement needs + event catalyst), and the extensible design for future roles.

## Requirements

### Requirement: Role Interface Definition

The system SHALL define a `Role` interface with methods for name identification, event generation, and role transition validation.

#### Scenario: Role interface methods

- **WHEN** a new role type is implemented
- **THEN** the implementation SHALL provide a `Name() string` method returning the role's identifier
- **THEN** the implementation SHALL provide a `GenerateEvents(figure, settlement, rng) []Event` method producing role-specific events
- **THEN** the implementation SHALL provide a `CanTransitionTo(other Role) bool` method to validate allowed role transitions

#### Scenario: Adding a future role without modifying existing code

- **WHEN** a developer adds a new role type (e.g., `Artisan`) that implements the `Role` interface
- **THEN** no changes to `Leader` or `Explorer` code SHALL be required
- **THEN** the new role SHALL integrate with existing role assignment and event generation systems

### Requirement: Leader Role

The Leader role SHALL govern a settlement and generate political events. At most one active leader SHALL exist per settlement.

#### Scenario: Leader generates political events

- **WHEN** a figure with the Leader role generates events for a year
- **THEN** events SHALL include categories such as Politics, Settlement, and Conflict
- **THEN** event descriptions SHALL reference the leader's name and the settlement name

#### Scenario: Leadership succession on death

- **WHEN** a settlement's leader dies from age or event risk
- **THEN** the leader's first living child SHALL be the primary successor candidate
- **THEN** if no living child exists, a random eligible adult figure from the settlement SHALL be assigned the Leader role
- **THEN** a succession event SHALL be emitted to the timeline

#### Scenario: Settlement without leader

- **WHEN** a settlement has no living figure with the Leader role at the start of a simulation year
- **THEN** a living figure SHALL be selected (first child of last leader, or random adult) and assigned the Leader role during that year's tick

### Requirement: Explorer Role

The Explorer role SHALL venture beyond the settlement and generate discovery events tied to the pointcrawl graph.

#### Scenario: Explorer generates discovery events

- **WHEN** a figure with the Explorer role generates events for a year
- **THEN** events SHALL include categories such as Discovery
- **THEN** events MAY reference specific pointcrawl nodes (landmarks, wilderness, ruins) discovered by the explorer
- **THEN** event descriptions SHALL reference the explorer's name and the settlement name

#### Scenario: Explorer assignment

- **WHEN** settlement figures are generated or a role vacancy for Explorer occurs
- **THEN** a figure MAY be assigned the Explorer role if the settlement has pointcrawl nodes adjacent to its region
- **THEN** settlement needs SHALL drive Explorer assignment (settlements at region borders with many undiscovered nodes spawn more explorers)

### Requirement: Role Assignment — Settlement Needs

The system SHALL assign roles to figures based on settlement context (leader vacancies, unexplored frontier).

#### Scenario: Leader vacancy triggers assignment

- **WHEN** a settlement has no figure with the Leader role
- **THEN** the settlement SHALL assign the Leader role to the most suitable living figure (first child of previous leader, or eldest adult)

#### Scenario: Explorer assignment for frontier settlements

- **WHEN** a settlement has undiscovered pointcrawl nodes adjacent to its region and has fewer than the maximum explorer count
- **THEN** a roleless adult figure MAY be assigned the Explorer role during a settlement tick

### Requirement: Role Transition — Event Catalyzed

The system SHALL allow roles to change through event catalysis (leader dies in disaster, explorer founds a new settlement).

#### Scenario: Event-catalyzed leader transition

- **WHEN** a disaster or conflict event kills the settlement's leader
- **THEN** a transition event SHALL be emitted identifying the new leader
- **THEN** the Leadership Succession logic SHALL assign a successor

#### Scenario: Explorer becomes leader

- **WHEN** an explorer's actions lead to a significant settlement development (e.g., discovery leads to settlement expansion)
- **THEN** the explorer MAY transition to the Leader role if `CanTransitionTo` returns true
- **THEN** a role transition event SHALL be emitted
