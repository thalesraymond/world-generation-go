## MODIFIED Requirements

### Requirement: Generate YAML Frontmatter

The system SHALL inject YAML frontmatter containing relevant metadata at the top of each generated Markdown file.

#### Scenario: Metadata inclusion for querying

- **WHEN** an entity file is generated
- **THEN** it includes a YAML frontmatter block `---` at the very beginning
- **THEN** the frontmatter includes key properties like `id`, `type`, `status`, `subtype`, and coordinates to enable Dataview queries

#### Scenario: Settlement subtype in frontmatter

- **WHEN** a settlement entity file is generated
- **THEN** the frontmatter includes a `subtype` field with the settlement's classification (MajorCity, City, Village, or Abandoned)