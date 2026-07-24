## Purpose

Define the expected initialization behavior for scaffolding project configuration and structure.

## Requirements

### Requirement: Initialize project structure

The system SHALL provide an `init` subcommand that accepts a world name and a world-size preset for project initialization.

#### Scenario: Running init command

- **WHEN** the user runs `worldgen init --name "Ashtar" --size large`
- **THEN** the system acknowledges initialization
- **AND** the acknowledgement includes the resolved world name and size preset.

### Requirement: Initialization Options

The command SHALL accept `--name` and `--size` options. `--size` SHALL default to `medium` when no value is configured.

#### Scenario: Default size

- **WHEN** the user runs `worldgen init --name "Ashtar"` without a size value
- **THEN** the resolved size is `medium`.

### Requirement: Current Initialization Boundary

At the CLI-foundation stage, `init` SHALL validate and acknowledge the resolved initialization request without creating project files. Persistent project scaffolding is a future behavior and MUST be specified before being introduced.

#### Scenario: No filesystem scaffold at foundation stage

- **WHEN** the command completes successfully
- **THEN** it has not required a world-state file or an export directory to acknowledge the request.
