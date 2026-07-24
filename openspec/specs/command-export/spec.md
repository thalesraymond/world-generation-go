## Purpose

Define the expected behavior of the export command for writing generated world data to an output destination.

## Requirements

### Requirement: Export generated world data

The system SHALL provide an `export` subcommand that accepts an export-format selection and resolves the output destination from global configuration.

#### Scenario: Running export command

- **WHEN** the user runs `worldgen export --format obsidian --output ./vault`
- **THEN** the system acknowledges the resolved format and output destination.

### Requirement: Export Format Option

The command SHALL accept a `--format` option whose default is `obsidian`.

#### Scenario: Default export format

- **WHEN** the user runs `worldgen export` without a format value
- **THEN** the resolved format is `obsidian`.

### Requirement: Current Export Boundary

At the CLI-foundation stage, `export` SHALL validate and acknowledge the request without writing files. Markdown serialization, vault layout, YAML frontmatter, and wiki-link generation are future behavior owned by the Obsidian-export capability.

#### Scenario: Foundation-stage export

- **WHEN** the command completes successfully before an exporter is installed
- **THEN** no exported-data file is required for the acknowledgement.
