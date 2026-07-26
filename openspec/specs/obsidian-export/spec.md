## ADDED Requirements

### Requirement: Generate Relational Directory Structure
The system SHALL export internal simulation structures into a relational directory hierarchy.

#### Scenario: Exporting diverse entities
- **WHEN** the export process is triggered
- **THEN** top-level directories are created based on entity types (e.g., `bases/`, `lore/`, `factions/`)
- **THEN** each entity is saved as a Markdown file within its respective type directory

### Requirement: Generate YAML Frontmatter
The system SHALL inject YAML frontmatter containing relevant metadata at the top of each generated Markdown file.

#### Scenario: Metadata inclusion for querying
- **WHEN** an entity file is generated
- **THEN** it includes a YAML frontmatter block `---` at the very beginning
- **THEN** the frontmatter includes key properties like `id`, `type`, `status`, and coordinates to enable Dataview queries

### Requirement: Generate Bi-directional Wiki-links
The system SHALL translate internal relationships between entities into Obsidian-compatible bi-directional wiki-links.

#### Scenario: Linking related entities
- **WHEN** an entity has a relationship to another entity (e.g., a character belongs to a base)
- **THEN** the Markdown output includes a `[[Entity Name]]` link
- **THEN** the linked names are properly sanitized to match the destination filename

### Requirement: Sanitize Filenames for Export
The system SHALL ensure that all generated filenames are safe for the filesystem and unique to prevent overwriting.

#### Scenario: Handling name collisions and illegal characters
- **WHEN** generating a filename for an entity
- **THEN** illegal filesystem characters are stripped or replaced
- **THEN** if a collision occurs (two entities with the same name), a unique identifier is appended to the filename
