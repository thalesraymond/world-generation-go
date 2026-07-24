## ADDED Requirements

### Requirement: Suitability Scoring
The system SHALL evaluate each map tile and assign a suitability score for settlement based on geographic features.

#### Scenario: High suitability near water and flat terrain
- **WHEN** a tile is near a fresh water source, has low elevation variance, and is in a temperate biome
- **THEN** it receives a high suitability score (e.g., > 0.8)

#### Scenario: Low suitability in harsh environments
- **WHEN** a tile is in a desert or high mountain peak
- **THEN** it receives a low suitability score (e.g., < 0.2)
