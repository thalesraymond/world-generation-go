## Purpose

Define the baseline CLI framework behavior for command routing and configuration precedence.

## Requirements

### Requirement: CLI application structure and config management

The system SHALL provide a CLI entrypoint that parses commands, flags, and configuration using Cobra and Viper.

#### Scenario: Running the application without arguments

- **WHEN** the user runs the CLI binary without arguments
- **THEN** the system outputs help information and lists available commands

#### Scenario: Registered command surface

- **WHEN** the user inspects root command help
- **THEN** the `init`, `simulate`, and `export` subcommands are listed.

### Requirement: Global Configuration Interface

The root command SHALL expose persistent `--config`, `--seed`, and `--output` flags. `--config` selects a YAML configuration file, `--seed` is the deterministic master-seed override, and `--output` selects the generated-artifact directory with a default of `./output`.

#### Scenario: Default output directory

- **WHEN** no output value is supplied through a flag, environment variable, or configuration file
- **THEN** the resolved output directory is `./output`.

### Requirement: Configuration File Discovery

The CLI SHALL read an explicitly selected configuration file when `--config` is supplied. Without that flag, it SHALL search for `worldgen.yaml` in the user home directory and the working directory. A missing implicitly searched configuration file SHALL not prevent command execution.

#### Scenario: Explicit configuration file

- **WHEN** the user invokes a command with `--config path/to/worldgen.yaml`
- **THEN** values from that YAML file are available to command configuration.

#### Scenario: Configuration precedence

- **WHEN** a setting is defined in a config file, an environment variable, and a CLI flag
- **THEN** the system uses the value from the CLI flag

#### Scenario: Environment variable precedence

- **WHEN** a setting is defined in a configuration file and its `WORLDGEN_` environment variable is set without a corresponding CLI flag
- **THEN** the system uses the environment-variable value.

### Requirement: Typed Configuration

The configuration boundary SHALL expose typed values for `config`, `seed`, `output`, `name`, `size`, `years`, `events`, and `format` so command handlers do not parse untyped configuration values directly.

#### Scenario: Command configuration loading

- **WHEN** a command begins execution
- **THEN** it obtains its resolved settings through the typed configuration boundary.
