## ADDED Requirements

### Requirement: CLI application structure and config management

The system SHALL provide a CLI entrypoint that parses commands, flags, and configuration using Cobra and Viper.

#### Scenario: Running the application without arguments

- **WHEN** the user runs the CLI binary without arguments
- **THEN** the system outputs help information and lists available commands

#### Scenario: Configuration precedence

- **WHEN** a setting is defined in a config file, an environment variable, and a CLI flag
- **THEN** the system uses the value from the CLI flag
