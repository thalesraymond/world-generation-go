## ADDED Requirements

### Requirement: Settlement Relationship Export
The Obsidian export SHALL include agent relationships (allies and rivals with sentiment scores) on settlement pages.

#### Scenario: Settlement with allies
- **WHEN** a settlement with positive-sentiment relationships to other settlements is exported
- **THEN** the settlement page SHALL include an "Allies" section with wiki-links to allied settlements
- **THEN** the section SHALL include sentiment score values for each relationship

#### Scenario: Settlement with rivals
- **WHEN** a settlement with negative-sentiment relationships to other settlements is exported
- **THEN** the settlement page SHALL include a "Rivals" section with wiki-links to rival settlements

#### Scenario: Settlement with no relationships
- **WHEN** a settlement with no relationships is exported
- **THEN** the settlement page SHALL omit the relationships section entirely

### Requirement: Settlement Faction Membership Export
The Obsidian export SHALL show faction membership on settlement pages.

#### Scenario: Settlement belonging to a faction
- **WHEN** a settlement that belongs to a faction is exported
- **THEN** the settlement page frontmatter SHALL include a `faction` field with the faction name
- **THEN** the settlement page body SHALL include a wiki-link to the faction's page

### Requirement: Figure Stats and Reputation Export
The Obsidian export SHALL include figure stats, reputation, and role history on figure pages.

#### Scenario: Figure with stats
- **WHEN** a figure with martial prowess and diplomatic skill stats is exported
- **THEN** the figure page SHALL include a "Stats" section with the stat values
- **THEN** the frontmatter SHALL include `martialProwess` and `diplomaticSkill` fields

#### Scenario: Figure with reputation
- **WHEN** a figure with a reputation score is exported
- **THEN** the figure page SHALL include a "Reputation" section with the score value
- **THEN** the frontmatter SHALL include a `reputation` field

#### Scenario: Figure with role history
- **WHEN** a figure that has held multiple roles over time is exported
- **THEN** the figure page SHALL include a "Role History" section listing each role and the years it was held

### Requirement: Figure Owned Artifacts Export
The Obsidian export SHALL list artifacts currently owned by a figure on that figure's page.

#### Scenario: Figure owning artifacts
- **WHEN** a figure that owns artifacts is exported
- **THEN** the figure page SHALL include an "Artifacts" section with wiki-links to each owned artifact's page

#### Scenario: Figure with no artifacts
- **WHEN** a figure that owns no artifacts is exported
- **THEN** the figure page SHALL omit the artifacts section entirely

### Requirement: Faction Dynamic Membership Export
The Obsidian export SHALL show dynamic membership lists on faction pages.

#### Scenario: Faction with settlements
- **WHEN** a faction with member settlements is exported
- **THEN** the faction page SHALL include a "Settlements" section with wiki-links to each member settlement

#### Scenario: Faction with members that left
- **WHEN** a faction has historical membership changes (settlements joined or left)
- **THEN** the faction page SHALL include a "Membership History" section documenting changes

### Requirement: Faction Strategic Decisions Export
The Obsidian export SHALL show strategic decisions on faction pages.

#### Scenario: Faction with strategic decisions
- **WHEN** a faction with strategic decisions (wars, alliances, policy shifts) is exported
- **THEN** the faction page SHALL include a "Strategic Decisions" section listing each decision with year and description

### Requirement: Faction Treasury Export
The Obsidian export SHALL show faction treasury on faction pages.

#### Scenario: Faction with treasury
- **WHEN** a faction with a treasury value is exported
- **THEN** the faction page frontmatter SHALL include a `treasury` field
- **THEN** the faction page body SHALL include a "Treasury" section with the current value

### Requirement: Artifact Page Generation
The Obsidian export SHALL generate dedicated artifact pages for each artifact in the world state.

#### Scenario: Artifact page creation
- **WHEN** the world state contains artifacts
- **THEN** an `artifacts/` directory SHALL be created in the export
- **THEN** one Markdown file per artifact SHALL be created in `artifacts/`
- **THEN** each file SHALL have YAML frontmatter with `id`, `type: artifact`, `name`, `artifactType`, `currentOwner`

#### Scenario: Artifact page body content
- **WHEN** an artifact page is generated
- **THEN** the page SHALL include an "Origin" section with the creation event description
- **THEN** the page SHALL include a "Provenance" section listing the ownership history as a chronological list
- **THEN** the page SHALL include a wiki-link to the current owner's page (figure or settlement)

### Requirement: Artifact Provenance Wiki-links
The Obsidian export SHALL ensure artifact provenance entries include wiki-links to all historical owners.

#### Scenario: Artifact with multiple owners
- **WHEN** an artifact page is generated for an artifact that has changed hands multiple times
- **THEN** each provenance entry SHALL include a wiki-link to the owner at that point in history

### Requirement: Cross-Entity Wiki-link Consistency
The Obsidian export SHALL maintain consistent wiki-links across all entity types (settlements, figures, factions, artifacts).

#### Scenario: Bidirectional link integrity
- **WHEN** a settlement page links to a figure page
- **THEN** the figure page SHALL link back to the settlement page
- **THEN** both pages SHALL use the same sanitized filename for the link target

### Requirement: Dataview-Compatible Frontmatter Extension
The Obsidian export SHALL extend YAML frontmatter on all entity types to include fields compatible with Dataview queries.

#### Scenario: Dataview queryable frontmatter
- **WHEN** any entity page is generated
- **THEN** the frontmatter SHALL include all fields required by the respective entity type's Dataview schema
- **THEN** all fields SHALL use consistent naming conventions (camelCase)
