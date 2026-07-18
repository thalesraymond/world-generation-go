## ADDED Requirements

### Requirement: Export generated world data
The system SHALL provide an `export` subcommand that exports the simulation output to a specified format or destination.

#### Scenario: Running export command
- **WHEN** the user runs `app export`
- **THEN** the system writes the exported data to disk
