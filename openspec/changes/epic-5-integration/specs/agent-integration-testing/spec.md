## ADDED Requirements

### Requirement: Full Pipeline Integration Test
The test suite SHALL include an integration test that exercises the complete `init → simulate → export` pipeline with agent-driven simulation enabled.

#### Scenario: Happy path full pipeline
- **WHEN** the integration test runs `init` with a fixed seed, `simulate` with agent-driven phases, and `export` to a temporary directory
- **THEN** the simulation completes without errors
- **THEN** the export directory contains settlement, figure, faction, chronicle, and artifact pages
- **THEN** all wiki-links in exported pages resolve to existing files

### Requirement: Deterministic Integration Test
The test suite SHALL include a determinism test verifying that identical seeds produce byte-identical output across the full integrated pipeline.

#### Scenario: Same seed produces identical output
- **WHEN** the full pipeline is run twice with the same seed
- **THEN** all generated files in the export directory SHALL be byte-identical between the two runs
- **THEN** the timeline event stream SHALL be byte-identical between the two runs

### Requirement: Causal Chain Validation Test
The test suite SHALL include a test that validates causal chains are present in the generated timeline.

#### Scenario: War → migration → settlement chain
- **WHEN** the simulation runs with a seed that produces a war event
- **THEN** the timeline SHALL contain a migration event causally linked to the war event
- **THEN** the migration event SHALL reference the originating war event or settlement

#### Scenario: Raid → artifact transfer chain
- **WHEN** the simulation runs with a seed that produces a raid event
- **THEN** the timeline SHALL contain an artifact transfer event (if artifacts exist) causally linked to the raid

### Requirement: Artifact Persistence Validation Test
The test suite SHALL include a test that verifies artifacts persist and change hands correctly through the simulation timeline.

#### Scenario: Artifact creation and transfer
- **WHEN** the simulation runs and creates an artifact
- **THEN** the artifact SHALL appear in the world state at the end of simulation
- **THEN** the artifact's provenance history SHALL contain the creation event
- **THEN** if the artifact changes hands, the provenance history SHALL reflect the transfer

### Requirement: Faction Dynamics Validation Test
The test suite SHALL include a test that verifies faction membership changes are reflected in the exported output.

#### Scenario: Settlement faction switch
- **WHEN** the simulation runs and a settlement switches faction through conquest or diplomacy
- **THEN** the settlement's faction page in the export SHALL reflect the updated membership
- **THEN** the old faction's page SHALL no longer list the settlement as a member

### Requirement: Agent Decision Logic Test Coverage
The test suite SHALL include unit tests for agent decision logic covering all six settlement actions (expand, raid, conquer, fortify, ally, prosper).

#### Scenario: Each action type is tested
- **WHEN** the test suite runs
- **THEN** each of the six settlement action types SHALL have at least one dedicated test case
- **THEN** each test case SHALL verify the action's preconditions, execution, and event generation

### Requirement: Coverage Threshold Validation
The test suite SHALL validate that coverage thresholds are met after integration changes.

#### Scenario: Repository-wide coverage
- **WHEN** `go test ./... -coverprofile` is run after all integration changes
- **THEN** repository-wide statement coverage SHALL be ≥ 80%

#### Scenario: Domain and usecase coverage
- **WHEN** coverage is measured for `internal/domain` and `internal/usecase`
- **THEN** each SHALL have statement coverage ≥ 90%
