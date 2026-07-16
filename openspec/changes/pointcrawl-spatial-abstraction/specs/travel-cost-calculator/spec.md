## ADDED Requirements

### Requirement: Calculate travel cost in watches
The system SHALL compute the cost of traversing a path between two nodes, measured in "watches", based on the environmental friction and distance.

#### Scenario: Calculate safe journey cost
- **WHEN** a route covers a short distance over a standard plain trail
- **THEN** the system outputs a total cost of 1 Watch (Safe Journey).

#### Scenario: Calculate high risk journey cost
- **WHEN** a route covers a medium distance over a dangerous orogenic slope
- **THEN** the system outputs a total cost of 4 Watches (High Risk).

#### Scenario: Calculate long hostile journey cost
- **WHEN** a route covers a long distance through a hostile mythic forest
- **THEN** the system outputs a total cost of 5 Watches (Need for Rest).
