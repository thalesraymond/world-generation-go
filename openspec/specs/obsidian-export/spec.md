# obsidian-export Specification

## Purpose

Define how simulation data is exported to an Obsidian-compatible Markdown vault, including historical figures as character files with YAML frontmatter and bi-directional wiki-links.

## Requirements

### Requirement: Generate Relational Directory Structure
The system SHALL export internal simulation structures into a relational directory hierarchy. When figures are present, a `characters/` directory SHALL be created.

#### Scenario: Exporting diverse entities
- **WHEN** the export process is triggered
- **THEN** top-level directories are created based on entity types (e.g., `bases/`, `lore/`, `factions/`)
- **THEN** each entity is saved as a Markdown file within its respective type directory
- **THEN** a `characters/` directory SHALL be created if the world state contains historical figures

### Requirement: Generate YAML Frontmatter
The system SHALL inject YAML frontmatter containing relevant metadata at the top of each generated Markdown file.

#### Scenario: Metadata inclusion for querying
- **WHEN** an entity file is generated
- **THEN** it includes a YAML frontmatter block `---` at the very beginning
- **THEN** the frontmatter includes key properties like `id`, `type`, `status`, `subtype`, and coordinates to enable Dataview queries

#### Scenario: Settlement subtype in frontmatter
- **WHEN** a settlement entity file is generated
- **THEN** the frontmatter includes a `subtype` field with the settlement's classification (MajorCity, City, Village, or Abandoned)

### Requirement: Generate Bi-directional Wiki-links
The system SHALL translate internal relationships between entities into Obsidian-compatible bi-directional wiki-links. Settlement files SHALL link to their figures, and figure files SHALL link to their settlement, faction, parents, children, and spouse.

#### Scenario: Linking related entities
- **WHEN** an entity has a relationship to another entity (e.g., a character belongs to a base)
- **THEN** the Markdown output includes a `[[Entity Name]]` link
- **THEN** the linked names are properly sanitized to match the destination filename

#### Scenario: Settlement links to figures
- **WHEN** a settlement file is generated and the settlement has historical figures
- **THEN** the settlement body SHALL include wiki-links to each figure's character file
- **THEN** figures SHALL be grouped by role (Leader, Explorer, Others)

#### Scenario: Figure links to settlement and faction
- **WHEN** a character file is generated
- **THEN** the file SHALL include a wiki-link to the figure's settlement
- **THEN** the file SHALL include a wiki-link to the figure's faction

#### Scenario: Figure links to family members
- **WHEN** a character file is generated for a figure with relationships
- **THEN** the "Relationships" section SHALL include wiki-links to parent, child, and spouse character files

### Requirement: Sanitize Filenames for Export
The system SHALL ensure that all generated filenames are safe for the filesystem and unique to prevent overwriting.

#### Scenario: Handling name collisions and illegal characters
- **WHEN** generating a filename for an entity
- **THEN** illegal filesystem characters are stripped or replaced
- **THEN** if a collision occurs (two entities with the same name), a unique identifier is appended to the filename

### Requirement: Export Historical Figures
The export system SHALL create individual Markdown character files in the `characters/` directory from figure data.

#### Scenario: Character file export
- **WHEN** a world state with historical figures is exported
- **THEN** one Markdown file per figure SHALL be created in `characters/`
- **THEN** each file SHALL have YAML frontmatter containing figure metadata
- **THEN** each file SHALL have a Markdown body with role, settlement, faction, relationships, and chronicle sections

#### Scenario: Character file frontmatter completeness
- **WHEN** a character file is generated
- **THEN** the frontmatter SHALL include: `id`, `type: character`, `name`, `role`, `faction`, `birthYear`, `settlement`, `status`
- **THEN** `deathYear` SHALL be present when the figure is deceased and absent when alive
- **THEN** `parents`, `children`, `spouse` SHALL be present as YAML lists when relationships exist

### Requirement: Chronicle Event Integration with Figures
Chronicle event descriptions SHALL include a figure reference when events reference figures.

#### Scenario: Chronicle event with figure link
- **WHEN** a chronicle event references a figure (has a non-empty `figureId`)
- **THEN** the event description displayed in the chronicle SHALL contain an inline wiki-link reference to the figure
- **THEN** the reference SHALL use the event's `figureId` value
