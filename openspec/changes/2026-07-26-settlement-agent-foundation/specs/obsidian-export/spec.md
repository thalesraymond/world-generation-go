# obsidian-export Specification

## Purpose

Define the Obsidian-compatible Markdown export format including agent state sections for settlements.

## MODIFIED Requirements

### Requirement: Agent State in Settlement Export

Settlement Markdown files SHALL include agent state sections.

#### Scenario: Military Strength section

- **WHEN** a settlement Markdown file is exported
- **THEN** it SHALL include a "## Military Strength" section
- **THEN** the section SHALL include the numeric value: `**Value:** {MilitaryStrength}`
- **THEN** the section SHALL include a tier description: Weak (<50), Moderate (50–150), Strong (150–300), Mighty (>300)

#### Scenario: Wealth section

- **WHEN** a settlement Markdown file is exported
- **THEN** it SHALL include a "## Wealth" section
- **THEN** the section SHALL include the numeric value: `**Value:** {Wealth}`
- **THEN** the section SHALL include a tier description: Poor (<500), Comfortable (500–1500), Prosperous (1500–3000), Rich (>3000)

#### Scenario: Relations section

- **WHEN** a settlement Markdown file is exported
- **THEN** it SHALL include a "## Relations" section
- **THEN** the section SHALL include a "### Allies" subsection listing top 5 settlements with highest positive relations
- **THEN** each ally entry SHALL be formatted as a wiki-link: `- [[{settlementName}]] (+{value})`
- **THEN** the section SHALL include a "### Rivals" subsection listing top 5 settlements with most negative relations
- **THEN** each rival entry SHALL be formatted as a wiki-link: `- [[{settlementName}]] ({value})`

#### Scenario: Goals section

- **WHEN** a settlement Markdown file is exported
- **THEN** it SHALL include a "## Goals" section
- **THEN** the section SHALL list all goals as bullet points: `- {goal}`

### Requirement: Existing Export Compatibility

The agent state sections SHALL be additive to existing export sections.

#### Scenario: Existing sections preserved

- **WHEN** a settlement Markdown file is exported
- **THEN** existing sections (Figures, Chronicles, etc.) SHALL remain unchanged
- **THEN** agent state sections SHALL be added after existing sections
