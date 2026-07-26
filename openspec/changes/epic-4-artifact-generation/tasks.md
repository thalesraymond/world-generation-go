# Tasks: Epic 4 — Artifact Generation

## Prerequisites

- [ ] Epic 1 (Settlement Agent Foundation) is complete — agent decision loop, core actions (Raid, Conquer, Prosper)
- [ ] Epic 2 (Character-Driven Execution) is complete — figure executors, Master Smith role, General role, stats
- [ ] Epic 3 (Faction-Level Agency) is complete — faction entities, strategic actions, dynamic membership

## Phase 1: Domain Model & RNG Foundation

### Task 1.1: Define Artifact domain types
- [ ] Create `internal/domain/artifact.go` with `ArtifactType` enum and `ProvenanceEvent` struct
- [ ] Create `Artifact` struct with all fields per design.md
- [ ] Ensure no imports beyond Go stdlib (Clean Architecture compliance)
- [ ] Write unit tests for `Artifact` construction and field access
- **Verification**: `go test ./internal/domain/... -run Artifact`

### Task 1.2: Implement artifact RNG derivation
- [ ] Create RNG derivation helpers for artifact operations (name, creation, transfer, decay)
- [ ] Each derivation consumes master seed + operation-scoped key
- [ ] Write determinism test: same master seed → same artifact names and creation decisions
- **Verification**: `go test ./internal/usecase/... -run ArtifactRNG`

### Task 1.3: Add artifacts to world state
- [ ] Add `Artifacts []Artifact` field to `world.State`
- [ ] Update JSON serialization round-trip to include artifacts
- [ ] Handle backward compatibility: deserialization of older state without `artifacts` field → empty slice
- [ ] Write round-trip test
- **Verification**: `go test ./internal/infra/... -run ArtifactSerialization`

---

## Phase 2: Artifact Creation

### Task 2.1: Implement CFG-based artifact name generation
- [ ] Define CFG grammar rules for artifact names (e.g., `<ArtifactName> ::= <Prefix> <BaseName>, the <Epithet>`)
- [ ] Integrate with existing CFG narrative engine (`internal/usecase/`)
- [ ] Support type-appropriate name patterns (weapons get martial names, tomes get scholarly names)
- [ ] Write unit tests for name generation determinism and type-appropriate patterns
- **Verification**: `go test ./internal/usecase/... -run ArtifactName`

### Task 2.2: Implement threshold-based creation (General victories)
- [ ] Implement `GeneralForgeWeapon` usecase: when General figure's victory counter reaches threshold, create weapon artifact
- [ ] Track victory count per General figure (requires Epic 2 figure stats infrastructure)
- [ ] Write unit test: General with 5 victories → artifact created; General with 4 victories → no artifact
- **Verification**: `go test ./internal/usecase/... -run GeneralForge`

### Task 2.3: Implement role-based creation (Master Smith)
- [ ] Implement `MasterSmithForge` usecase: when Master Smith executes Forge action with sufficient settlement wealth
- [ ] Artifact type randomized from available types with weighted probabilities
- [ ] Success probability ~60%, modified by settlement wealth level
- [ ] Write unit test: Master Smith with wealth threshold → artifact created; without → no artifact
- **Verification**: `go test ./internal/usecase/... -run MasterSmithForge`

### Task 2.4: Implement event-based creation (conquest capture, discovery)
- [ ] Conquest: conquered settlement's crown-type artifacts transfer to conqueror
- [ ] Prosper/discovery: low-probability artifact discovery for old settlements (age ≥ 200 years)
- [ ] Write unit test for conquest crown capture
- [ ] Write unit test for discovery probability thresholds
- **Verification**: `go test ./internal/usecase/... -run ArtifactDiscovery`

---

## Phase 3: Artifact Transfer & Provenance

### Task 3.1: Implement raid seizure logic
- [ ] Successful raid → raider may seize one random artifact from target settlement
- [ ] Seizure probability scales with raid success magnitude
- [ ] Append `ProvenanceEvent` of type "stolen" with from/to owner IDs
- [ ] Update `CurrentOwnerID` on seized artifact
- [ ] Write unit test: raid succeeds → artifact transferred; raid fails → artifact stays
- **Verification**: `go test ./internal/usecase/... -run ArtifactSeizure`

### Task 3.2: Implement conquest treasury capture
- [ ] Successful conquest → all target settlement artifacts transfer to conqueror
- [ ] Each artifact gets a "captured" provenance event
- [ ] Write unit test: conquest → all artifacts transferred
- **Verification**: `go test ./internal/usecase/... -run ConquestCapture`

### Task 3.3: Implement inheritance on figure death
- [ ] When a figure dies owning artifacts, transfer to heir (first living child) or settlement treasury
- [ ] Append "inherited" or "transferred" provenance event
- [ ] Write unit test: figure with artifacts dies → artifacts go to heir; no heir → settlement treasury
- **Verification**: `go test ./internal/usecase/... -run ArtifactInheritance`

### Task 3.4: Implement provenance tracking validation
- [ ] Every ownership change must record a provenance event
- [ ] Provenance must be chronologically ordered
- [ ] No provenance gaps allowed (an artifact cannot skip from owner A to owner C without recording the transfer)
- [ ] Write integration test: simulate 200 years, verify all artifacts have complete provenance chains
- **Verification**: `go test ./internal/usecase/... -run ProvenanceChain`

---

## Phase 4: Artifact-Driven Conflict

### Task 4.1: Implement narrative value model
- [ ] Newly created artifacts start at `NarrativeValue = 0.3`
- [ ] Each transfer increases narrative value by 0.15 (capped at 1.0)
- [ ] Each year without transfer decays narrative value by 0.01 (floor at 0.1)
- [ ] Write unit test for value increase on transfer and decay over time
- **Verification**: `go test ./internal/domain/... -run NarrativeValue`

### Task 4.2: Implement artifact-motivated raid targeting
- [ ] Faction agent (Epic 3) evaluates artifacts: if faction once owned an artifact now held by another faction, and narrative value > threshold (0.5), increase raid probability against current holder
- [ ] Write unit test: faction with lost high-value artifact → higher raid probability against holder
- **Verification**: `go test ./internal/usecase/... -run ArtifactMotivatedRaid`

### Task 4.3: Implement artifact-motivated war declaration
- [ ] Faction agent (Epic 3) may declare war specifically to recover lost artifact
- [ ] War motivation text generated via CFG referencing artifact name
- [ ] Timeline event: "The Ashfield Alliance declared war on Coldcrest Confederacy to reclaim the Crown of Ashfield"
- [ ] Write unit test: faction with lost artifact + high narrative value → war declaration
- **Verification**: `go test ./internal/usecase/... -run ArtifactMotivatedWar`

---

## Phase 5: Obsidian Export

### Task 5.1: Implement artifact Markdown file generation
- [ ] Create `internal/infra/obsidian_artifact_writer.go`
- [ ] Generate one Markdown file per artifact in `notes/magic-items/`
- [ ] Sanitize filenames (replace illegal characters, handle collisions)
- [ ] Write unit test: artifact → Markdown file with correct filename
- **Verification**: `go test ./internal/infra/... -run ArtifactExport`

### Task 5.2: Implement YAML frontmatter for artifacts
- [ ] Frontmatter includes: `type: artifact`, `artifact_type`, `origin`, `creator`, `current_owner`, `owner_type`, `origin_year`, `status`, `provenance` (YAML list), `related_events`
- [ ] Optional fields omitted when empty (e.g., `creator` absent for discovered artifacts)
- [ ] Write unit test for frontmatter completeness and optional field handling
- **Verification**: `go test ./internal/infra/... -run ArtifactFrontmatter`

### Task 5.3: Implement wiki-link generation for artifact relationships
- [ ] Creator link: `[[Figure Name]]`
- [ ] Current owner link: `[[Figure Name]]` or `[[Settlement Name]]`
- [ ] Provenance events: wiki-links for figures, settlements, and timeline events
- [ ] "Related Figures" section: deduplicated list of all figures in provenance
- [ ] Write unit test for wiki-link accuracy and bi-directionality
- **Verification**: `go test ./internal/infra/... -run ArtifactWikiLinks`

### Task 5.4: Implement artifact provenance table in Markdown body
- [ ] Markdown table with columns: Year, Event Type, From, To, Description
- [ ] Sorted chronologically
- [ ] Write unit test for table structure and data accuracy
- **Verification**: `go test ./internal/infra/... -run ArtifactProvenanceTable`

---

## Phase 6: Integration & Determinism

### Task 6.1: Integrate artifact hooks into simulation pipeline
- [ ] After each settlement agent action, invoke artifact creation/transfer hooks
- [ ] After each figure death, invoke artifact inheritance hook
- [ ] Each simulation year: apply narrative value decay to all artifacts
- [ ] Write integration test: simulate 500-year world, verify artifacts created, transferred, and tracked
- **Verification**: `go test ./internal/usecase/... -run ArtifactSimulation`

### Task 6.2: Determinism end-to-end test
- [ ] Same seed, same configuration → identical artifact names, creation years, ownership history, and export output
- [ ] Test at multiple simulation lengths (100, 500, 1000 years)
- [ ] Verify byte-identical export output for same seed
- **Verification**: `go test ./internal/usecase/... -run DeterminismArtifact`

### Task 6.3: Coverage validation
- [ ] Run `go test ./... -coverprofile=coverage.out`
- [ ] Verify domain/usecase coverage >= 90%
- [ ] Verify repo-wide coverage >= 80%
- [ ] Add supplementary tests if thresholds not met
- **Verification**: `go tool cover -func=coverage.out | grep total`

### Task 6.4: Lint and vet
- [ ] Run `golangci-lint run ./...`
- [ ] Run `go vet ./...`
- [ ] Fix all warnings and errors
- **Verification**: CI green

---

## Task Dependency Graph

```
Phase 1 (Domain + RNG)
  ├── Phase 2 (Creation)
  │     ├── Task 2.1 (CFG names) — independent
  │     ├── Task 2.2 (General forge) — depends on Epic 2
  │     ├── Task 2.3 (Master Smith) — depends on Epic 2
  │     └── Task 2.4 (Conquest/Discovery) — depends on Epic 1
  ├── Phase 3 (Transfer)
  │     ├── Task 3.1 (Raid seizure) — depends on Epic 1 + Phase 1
  │     ├── Task 3.2 (Conquest capture) — depends on Epic 1 + Phase 1
  │     ├── Task 3.3 (Inheritance) — depends on Epic 2 + Phase 1
  │     └── Task 3.4 (Provenance validation) — depends on Tasks 3.1-3.3
  ├── Phase 4 (Conflict)
  │     ├── Task 4.1 (Narrative value) — depends on Phase 1
  │     ├── Task 4.2 (Raid motivation) — depends on Epic 3 + Phase 3
  │     └── Task 4.3 (War motivation) — depends on Epic 3 + Phase 3
  └── Phase 5 (Export)
        ├── Task 5.1 (File generation) — depends on Phase 1
        ├── Task 5.2 (Frontmatter) — depends on Task 5.1
        ├── Task 5.3 (Wiki-links) — depends on Task 5.1
        └── Task 5.4 (Provenance table) — depends on Phase 3 + Task 5.1

Phase 6 (Integration) depends on all Phases 1–5
```
