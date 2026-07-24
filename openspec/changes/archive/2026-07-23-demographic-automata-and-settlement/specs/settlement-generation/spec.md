## ADDED Requirements

### Requirement: Settlement Instantiation
The system SHALL instantiate discrete settlements at locations meeting population and suitability thresholds.

#### Scenario: Founding a new settlement
- **WHEN** a tile's population density reaches a threshold AND its suitability score is above a minimum value AND it is far enough from existing settlements
- **THEN** a new settlement is instantiated at that location and assigned to the local faction
