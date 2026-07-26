# figure-export Specification

## Purpose

Define Obsidian character file export: `characters/` directory with per-figure Markdown, YAML frontmatter containing all figure attributes, and wiki-links to settlements, chronicles, parents, children, and faction.

## Requirements

### Requirement: Characters Directory Creation

The export system SHALL create a `characters/` directory in the Obsidian vault output and populate it with one Markdown file per historical figure.

#### Scenario: Characters directory created

- **WHEN** the export process runs on a world state containing figures
- **THEN** a `characters/` directory SHALL be created under the output vault directory
- **THEN** each living or deceased figure SHALL generate a corresponding Markdown file
- **THEN** filenames SHALL be sanitized using the same rules as settlement and faction filenames

#### Scenario: No figures — no characters directory

- **WHEN** the export process runs on a world state with no figures or settlements containing no figures
- **THEN** the `characters/` directory SHALL NOT be created
- **THEN** the export SHALL succeed without error

### Requirement: Character File Content — Frontmatter

Each character file SHALL include YAML frontmatter with all figure identity and relationship metadata.

#### Scenario: Frontmatter fields

- **WHEN** a figure file is generated
- **THEN** the frontmatter SHALL include `id`, `type: character`, `name`, `role`, `faction`, `birthYear`, `deathYear` (or omit if alive), `settlement`
- **THEN** the frontmatter SHALL include `parents`, `children`, and `spouse` as YAML lists of wiki-linked names
- **THEN** the frontmatter SHALL include `status` ("alive" when `deathYear` is zero, "deceased" otherwise)

#### Scenario: Frontmatter for alive figure

- **WHEN** a figure is alive (deathYear is zero or unset)
- **THEN** the frontmatter SHALL NOT include a `deathYear` field
- **THEN** the `status` field SHALL be "alive"

### Requirement: Character File Content — Body

Each character file SHALL include a Markdown body with human-readable figure information and wiki-links.

#### Scenario: Character file body

- **WHEN** a figure file is generated
- **THEN** the body SHALL include sections for role, faction, settlement, lifespan, relationships, and chronicle
- **THEN** the faction SHALL be a wiki-link (`[[faction-name]]`)
- **THEN** the settlement SHALL be a wiki-link (`[[settlement-name]]`)
- **THEN** parents, children, and spouse SHALL be wiki-links to their respective character files

#### Scenario: Chronicle section with figure events

- **WHEN** a figure has associated timeline events (identified by `figureId` or `relatedFigures`)
- **THEN** the character file SHALL include a "Chronicle" section listing these events by year
- **THEN** each chronicle entry SHALL include the year and event description

### Requirement: Wiki-link Integration with Existing Exports

Settlement and faction files SHALL link to the figures that belong to them.

#### Scenario: Settlement file references figures

- **WHEN** a settlement file is generated and the settlement has figures
- **THEN** a "Characters" section SHALL be added to the settlement file body
- **THEN** figures SHALL be listed as wiki-links grouped by role (Leader section, Explorer section, Others)

#### Scenario: Faction file references figure leaders

- **WHEN** a faction file is generated and the faction has settlements with leaders
- **THEN** the faction file SHALL optionally reference the settlement leaders under each settlement listing

### Requirement: Chronicle Event Integration

Chronicle files SHALL include a figure reference when events reference figures.

#### Scenario: Chronicle event with figure

- **WHEN** a chronicle event is generated from a timeline event that has a `figureId`
- **THEN** the event description in the chronicle SHALL include an inline reference to the figure using wiki-link syntax
- **THEN** the reference SHALL use the event's `figureId` value
