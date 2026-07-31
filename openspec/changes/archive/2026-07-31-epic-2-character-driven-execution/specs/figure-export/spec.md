## MODIFIED Requirements

### Requirement: Stats in Character Frontmatter

Character export frontmatter SHALL include figure stats when present.

#### Scenario: Frontmatter includes stats

- **WHEN** a figure has non-zero Martial, Diplomatic, or Infamy stats
- **THEN** the frontmatter SHALL include `martial`, `diplomatic`, and `infamy` fields
- **THEN** the values SHALL be integer strings

### Requirement: Reputation Summary in Frontmatter

Character export frontmatter SHALL include total reputation when non-zero.

#### Scenario: Frontmatter includes reputation

- **WHEN** a figure has reputation entries with a non-zero total delta
- **THEN** the frontmatter SHALL include a `reputation` field with the total value

### Requirement: Stats Section in Character Body

Character export body SHALL include a Stats section summarizing figure attributes.

#### Scenario: Body includes stats

- **WHEN** a character file is generated
- **THEN** the body SHALL include a `## Stats` section
- **THEN** the section SHALL list Martial, Diplomatic, Infamy, and Total Reputation values

### Requirement: Notable Deeds Section

Character export body SHALL list reputation entries as notable deeds.

#### Scenario: Body lists deeds

- **WHEN** a figure has reputation entries
- **THEN** the body SHALL include a `## Notable Deeds` section
- **THEN** each entry SHALL include the year, description, delta, and optional event name

#### Scenario: Empty deeds list

- **WHEN** a figure has no reputation entries
- **THEN** the Notable Deeds section SHALL display a placeholder indicating no deeds recorded

### Requirement: Role Transition History Section

Character export body SHALL list role transitions.

#### Scenario: Body lists transitions

- **WHEN** a figure has role transition entries
- **THEN** the body SHALL include a `## Role Transition History` section
- **THEN** each entry SHALL include the year, previous role, new role, and reason

#### Scenario: Empty transition history

- **WHEN** a figure has no role transitions
- **THEN** the Role Transition History section SHALL display a placeholder indicating no transitions recorded

### Requirement: Settlement Character Stats Summary

Settlement character listings SHALL include a compact stats summary per figure.

#### Scenario: Stats summary in settlement export

- **WHEN** a settlement file lists characters
- **THEN** each character entry SHALL include the role and a summary of Martial, Diplomatic, and Infamy stats (e.g., `M:15 D:10 I:3`)

### Requirement: Role Source Consistency

Exported role fields SHALL reflect the runtime role object when available.

#### Scenario: Role from RoleRole

- **WHEN** a figure has both `RoleRole` and `Role` string fields
- **THEN** the exported role SHALL be taken from `RoleRole.Name()` when `RoleRole` is non-nil
