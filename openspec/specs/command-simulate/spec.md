## Purpose

Define the expected behavior of the simulate command to trigger deterministic world generation.

## Requirements

### Requirement: Trigger world generation

The system SHALL provide a `simulate` subcommand that accepts the requested number of years and an event-density preset for a simulation request.

#### Scenario: Running simulate command

- **WHEN** the user runs `worldgen simulate --years 500 --events dense`
- **THEN** the system reports the resolved year count and event-density preset
- **AND** the status reports that the request is queued.

### Requirement: Simulation Options

The command SHALL accept `--years` and `--events` options. The default year count SHALL be `100`, and the default event density SHALL be `normal`.

#### Scenario: Default simulation options

- **WHEN** the user runs `worldgen simulate` with no command-specific flags
- **THEN** the resolved values are `100` years and `normal` event density.

### Requirement: Current Simulation Boundary

At the CLI-foundation stage, the command SHALL not claim that world generation has completed. Chronological execution, persistent state loading, and live timeline streaming require the separate simulation-loop capability before they are exposed as command behavior.

#### Scenario: Foundation-stage status

- **WHEN** the command succeeds before the simulation-loop capability is available
- **THEN** it emits a queued status rather than simulated events or a completed-world claim.
