## MODIFIED Requirements

### Requirement: Export Historical Figures
The export system SHALL create individual Markdown character files in the `characters/` directory from figure data. Figure pages SHALL additionally include stats, reputation, role history, and owned artifacts sections when the data is present.

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

#### Scenario: Figure stats in frontmatter and body
- **WHEN** a character file is generated for a figure with martial prowess and diplomatic skill stats
- **THEN** the frontmatter SHALL include `martialProwess` and `diplomaticSkill` fields
- **THEN** the body SHALL include a "Stats" section with the values

#### Scenario: Figure reputation in frontmatter and body
- **WHEN** a character file is generated for a figure with a reputation score
- **THEN** the frontmatter SHALL include a `reputation` field
- **THEN** the body SHALL include a "Reputation" section with the score

#### Scenario: Figure role history
- **WHEN** a character file is generated for a figure that has held multiple roles
- **THEN** the body SHALL include a "Role History" section listing each role and the years it was held

#### Scenario: Figure owned artifacts
- **WHEN** a character file is generated for a figure that owns artifacts
- **THEN** the body SHALL include an "Artifacts" section with wiki-links to each owned artifact's page

## ADDED Requirements

### Requirement: Agent Relationship Export on Settlement Pages
The export system SHALL include agent relationships (allies and rivals with sentiment scores) on settlement pages when relationship data is present.

#### Scenario: Settlement with allies
- **WHEN** a settlement with positive-sentiment relationships is exported
- **THEN** the settlement page SHALL include an "Allies" section with wiki-links and sentiment scores

#### Scenario: Settlement with rivals
- **WHEN** a settlement with negative-sentiment relationships is exported
- **THEN** the settlement page SHALL include a "Rivals" section with wiki-links and sentiment scores

### Requirement: Faction Dynamic Membership and Decisions Export
The export system SHALL show dynamic membership lists, strategic decisions, and treasury on faction pages.

#### Scenario: Faction membership section
- **WHEN** a faction with member settlements is exported
- **THEN** the faction page SHALL include a "Settlements" section with wiki-links to each member

#### Scenario: Faction strategic decisions section
- **WHEN** a faction with strategic decisions is exported
- **THEN** the faction page SHALL include a "Strategic Decisions" section with year and description for each decision

#### Scenario: Faction treasury in frontmatter and body
- **WHEN** a faction with a treasury value is exported
- **THEN** the frontmatter SHALL include a `treasury` field
- **THEN** the body SHALL include a "Treasury" section with the current value

### Requirement: Artifact Page Generation
The export system SHALL generate dedicated artifact pages in an `artifacts/` directory for each artifact in the world state.

#### Scenario: Artifact page creation
- **WHEN** the world state contains artifacts
- **THEN** an `artifacts/` directory SHALL be created
- **THEN** one Markdown file per artifact SHALL be created with YAML frontmatter and body

#### Scenario: Artifact page body content
- **WHEN** an artifact page is generated
- **THEN** the page SHALL include "Origin", "Provenance", and "Current Owner" sections
- **THEN** the provenance section SHALL list the ownership history chronologically with wiki-links to historical owners

### Requirement: Cross-Entity Wiki-link Consistency
The export system SHALL maintain consistent wiki-links across all entity types (settlements, figures, factions, artifacts).

#### Scenario: Bidirectional link integrity
- **WHEN** a settlement page links to a figure page
- **THEN** the figure page SHALL link back to the settlement page
- **THEN** both pages SHALL use the same sanitized filename for the link target

### Requirement: Dataview-Compatible Frontmatter Extension
The export system SHALL extend YAML frontmatter on all entity types to include fields compatible with Dataview queries, using consistent camelCase naming.

#### Scenario: Frontmatter field consistency
- **WHEN** any entity page is generated
- **THEN** the frontmatter SHALL include all fields required by the entity type's Dataview schema
- **THEN** all field names SHALL use camelCase convention
